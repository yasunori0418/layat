# layat

*Read this in [Japanese (日本語)](README.ja.md).*

> **layat** — lays contents at root-relative targets, as the manifest says.

*Formerly **nput** — see [Migrating from nput](#migrating-from-nput).*

layat is a Nix library and module set that **places the contents of an already-fetched
Nix store path at a `root`-relative target** — as a symlink or a copy. It does **not**
generate configuration. `root` is chosen explicitly with the `projectRoot` / `homeRoot` /
`systemRoot` markers (there is **no implicit default**).

> **Status: MVP / implementation phase.** The implemented scope is the standalone CLI +
> **project mode** as the core, with **home mode** also supported. NixOS / nix-darwin
> modules and **system mode** are future work. See [MVP status](#mvp-status) for the full
> matrix. APIs may still change.

---

## Why layat

layat separates **fetching** (`src` is a store path) from **placement** (a fixed runtime
engine), and keeps placement behavior in a **single core** that you drive explicitly:

- **No configuration generation.** It places what the repository already contains.
- **Independent units.** Each config (`layat.<name>`) is its own Nix profile; one update
  never ripples to another.
- **No home-manager dependency.** `lib/` depends only on nixpkgs; module integrations only
  *kick* the engine.
- **A self-recorded manifest, not readlink pattern-matching.** Stale removal uses the
  previous generation's manifest. See [`docs/concept.md`](docs/concept.md#home-manager-homefile-との配置意味論の差).

---

## How it works

```
[layat CLI]  packages.layat — on PATH, the primary UX
  · discovers an entrypoint (flake.nix / shell.nix / default.nix)
  · obtains a named manifest through a generator (`nix` by default)
  · drives the engine to place, prune stale links, and swap the profile
   ↓ manifest.json
[engine]    a Go library (stdlib-only)
  · takes manifest.json as input; invokes only `nix` (profile) and `git` (toplevel)
  · native filesystem operations for place / replace / remove + conservative stale removal
```

- The **engine** owns placement and stale removal. It reads `manifest.json` — the stable
  Nix↔Go contract — and performs native filesystem operations.
- `lib.mkManifest` is a **pure function** that produces a link-farm derivation
  (`manifest.json` + a symlink farm). It has no side effects.
- An **entrypoint** is the Nix file the CLI reads (`flake.nix`, `shell.nix`, or
  `default.nix`); it exposes a named manifest under `layat.<name>`.
- A **generator** is what the CLI obtains the manifest from. `nix` is the default (see
  [Choosing the generator](#choosing-the-generator)); `apply --manifest` uses the
  *prebuilt* generator, which takes a built link-farm as it is.

---

## Requirements

- **Nix** with experimental features enabled in your environment:
  `experimental-features = nix-command` (and `flakes` for flake entrypoints).
  layat does **not** silently inject `--extra-experimental-features`; if a feature is not
  enabled it stops with a clear message explaining the prerequisite and how to enable it.
- **git** on `PATH` (used to resolve the project root in project mode).

---

## Install

### Standalone (home mode)

Install the CLI globally so `layat` is on `PATH`:

```bash
nix profile install github:yasunori0418/layat
```

To avoid `schemaVersion` skew, align the CLI and the `layat` your flake pins to the **same
input** (the global CLI and your flake's `layat.lib` are otherwise separate inputs that can
drift; the engine rejects a `schemaVersion` newer than its own).

### Project mode (canonical: pin in the devShell)

For project mode, the canonical form is to **bundle a pinned `layat`** in the project's
devShell so the CLI and `layat.lib` come from the same flake input (locked by `flake.lock`):

```nix
devShells.${system}.default = pkgs.mkShell {
  packages  = [ layat.packages.${system}.layat ];   # pinned layat on PATH
  shellHook = "layat apply <name> --no-wait";        # placed on `nix develop` / direnv entry
};
```

### Scaffold a new project

`layat init` is a transparent wrapper over `nix flake init -t`; layat itself generates
nothing, and existing files are never overwritten:

```bash
layat init standalone   # homeRoot example
layat init project      # projectRoot example + devShell wiring + .gitignore guide
```

---

## Quickstart

### Project mode (the central use)

In **project mode** the root is the project root (the git toplevel). Placements are
**ephemeral** — regenerated on each clone, never committed — so activation never touches
git state. This is the central way to use layat: embed it in a repo and place store paths
at arbitrary in-repo paths, kicked from a devShell.

```nix
# flake.nix — entering the repo places .claude/skills/nix from a fetched store path
{
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  inputs.layat.url    = "github:yasunori0418/layat";

  inputs.claude-skills.url   = "github:someone/claude-skills";
  inputs.claude-skills.flake = false;

  outputs = { self, nixpkgs, layat, ... }@inputs:
    let
      system = "x86_64-linux";
      pkgs   = nixpkgs.legacyPackages.${system};
    in
    {
      layat.${system}.skills = layat.lib.mkManifest {
        inherit pkgs;
        root = layat.lib.projectRoot;        # resolves to the git toplevel at runtime
        entries = {
          ".claude/skills/nix" = { src = inputs.claude-skills; subpath = "skills/nix"; };
        };
      };

      devShells.${system}.default = pkgs.mkShell {
        packages  = [ layat.packages.${system}.layat ];   # pinned layat (canonical)
        shellHook = "layat apply skills --no-wait";       # place on shell entry
      };
    };
}
```

```bash
# Enter the shell — the placement runs automatically via the shellHook
nix develop          # or: direnv allow

# List the targets the project owner should add to .gitignore (stdout only; no writes)
layat gitignore skills >> .gitignore
```

- Generations are an internal mechanism here; `rollback` / `list-generations` are **not**
  exposed in project mode (rollback is meaningless for ephemeral placements).
- In a devShell, use a **named apply** (`layat apply skills`) or
  `layat apply --all --project-root`. A bare `--all` would also place any home-mode configs
  into `$HOME` — a footgun in mixed entrypoints.

### Home mode (standalone, roles as separate profiles)

In **home mode** the root is `$HOME`, selected with `homeRoot`. Each config is its own
profile, committed every apply, with user-facing `rollback`.

```nix
# flake.nix — each role is a named manifest = an independent profile
outputs.layat.${system} = {
  vim-plugins = layat.lib.mkManifest {
    inherit pkgs;
    root = layat.lib.homeRoot;
    entries = {
      ".local/share/nvim/site/pack/foo/start/foo" = { src = inputs.vim-foo; };
      ".local/share/nvim/site/pack/bar/start/bar" = { src = inputs.vim-bar; };
    };
  };

  zsh-plugins = layat.lib.mkManifest {
    inherit pkgs;
    root = layat.lib.homeRoot;
    entries = {
      ".zsh/plugins/autosuggestions"     = { src = inputs.zsh-autosuggestions; };
      ".zsh/plugins/syntax-highlighting" = { src = inputs.zsh-syntax-highlighting; };
    };
  };
};
```

```bash
# Update / apply / roll back each role independently (separate profiles)
layat apply vim-plugins
layat rollback vim-plugins          # home mode only
layat apply zsh-plugins
layat list-generations vim-plugins
```

After updating a `src` (a flake input update, `npins update`, …), re-apply only the
affected config to keep the change from touching any other tool.

### home-manager module

The module pins `root = homeRoot` (you do not re-specify `root`). It kicks the engine from
`home.activation` via `layat apply --manifest <link-farm>` — it never delegates to
`home.file`. layat keeps its own profile as an **internal mechanism**; user-facing rollback
is unified on the host (`home-manager --rollback`).

```nix
imports = [ inputs.layat.homeManagerModules.default ];

layat = {
  enable = true;
  entries = {
    # external repo (store link)
    ".claude/skills/nix" = { src = inputs.skills-repo; subpath = "skills/nix"; };
    # theme as a copy (place-once, then user-managed)
    ".local/share/themes/dark" = { src = inputs.themes; subpath = "dark"; method = "copy"; };
    # live editing of local dotfiles via out-of-store symlink
    ".config/nvim" = { src = layat.lib.mkOutOfStoreSymlink "/home/me/dotfiles"; subpath = "home/.config/nvim"; };
  };
};
```

> The home-manager module is a **single manifest = one profile** (fixed name `default`) in
> the MVP — it has no `<name>` dimension, so **role separation is not available through the
> module**. Use the standalone CLI path (`layat.<name>` entrypoints) for multiple
> independent profiles.

### Adding layat to an existing flake

`layat init` is for new projects. To retrofit an existing `flake.nix`, do these four steps
by hand (layat never auto-merges your flake — "do not generate configuration"):

1. **Add the input**: `inputs.layat.url = "github:yasunori0418/layat";`
2. **Expose a manifest**:
   `outputs.layat.<system>.<name> = layat.lib.mkManifest { root = layat.lib.projectRoot; entries = { ... }; };`
3. **Bundle pinned layat in the devShell**: `packages = [ layat.packages.${system}.layat ];`
4. **Wire a named apply**: `shellHook = "layat apply <name> --no-wait";`

If the repo uses **flake-parts**, write step 2 via the flake module instead (keeps `pkgs`
consistent with `perSystem`):

```nix
imports = [ inputs.layat.flakeModules.default ];
perSystem = { pkgs, ... }: {
  layat.<name> = inputs.layat.lib.mkManifest {
    inherit pkgs;
    root = inputs.layat.lib.projectRoot;
    entries = { ... };
  };
};
# flake-parts transposes this to flake.layat.<system>.<name> — the CLI addressing is unchanged.
```

> `nix flake check` reports `warning: unknown flake output 'layat'` but exits 0 (harmless,
> expected — the `layat` namespace keeps manifests out of `packages`). Verify artifacts with
> `nix build .#layat.<system>.<name>`.

---

## `entries` schema

`entries` is an **attrset keyed by `target`** — the attribute key is the identifier and
supplies the default `target`. Uniqueness is guaranteed natively by Nix attrset keys; there
is no manual `name` field.

```nix
entries = {
  "<target>" = {
    src     = ...;          # required
    subpath = ".";          # optional, default "." (the whole repository)
    target  = "<target>";   # optional, default = the attribute key
    method  = "symlink";    # optional, default "symlink" | "copy"
  };
  # ...
};
```

The entry submodule is **strict** (unknown keys are rejected); typos and old names
(`name` / `source` / `dir` / `mode`) are evaluation-time errors.

### `src` — required

The placement source. The default is a **store link** (a symlink into the Nix store, which
guarantees reproducibility). Out-of-store is opt-in via an explicit marker.

| `src` value | Symlink points to | Use |
|---|---|---|
| `path` (e.g. `inputs.myrepo`) | Nix store (immutable) | version-pinned external repo |
| `builtins.path { path = /home/...; name = "..."; }` | Nix store (local copied in) | a local tree via the store |
| `set` (e.g. `pkgs.fetchFromGitHub { ... }`) | Nix store (immutable) | version-pinned external repo |
| `marker` (`layat.lib.mkOutOfStoreSymlink "/abs/path"`) | local FS (live) | local dotfiles under development |

```nix
src = inputs.myrepo;                                    # store link
src = pkgs.fetchFromGitHub { owner = "..."; repo = "..."; rev = "..."; hash = "..."; };
src = builtins.path { path = /path/to/dotfiles; name = "dotfiles"; };
src = layat.lib.mkOutOfStoreSymlink "/path/to/dotfiles"; # out-of-store (live), explicit

# Removed: passing a bare string for implicit out-of-store is not supported.
# src = "/path/to/dotfiles";   # error
```

### `subpath` — default `"."`

The relative path selecting which path **inside `src`** to take. Works for files and
directories. Omitting it (or `"."`) selects the whole repository.

```nix
subpath = ".";                  # whole repository (explicit form)
subpath = "skills/nix";         # a subdirectory
subpath = "themes/dark.json";   # a single file
```

`src` and `subpath` are orthogonal: `src` = *which thing* (store path / repository),
`subpath` = *which path inside it*.

### `target` — default = the attribute key

The destination, **relative to `root`**. The attribute key is the default `target` and is
also the entry's identity (the diff key for stale removal, and its uniqueness key). You can
make the key a logical label and override `target` explicitly (like `home.file`).

### `method` — default `"symlink"`

Selects the placement kind:

| `method` | `src` kind | Behavior | Generations |
|---|---|---|---|
| `"symlink"` | path / set | symlink into the Nix store (read-only) | yes (profile) |
| `"symlink"` | marker | out-of-store symlink to a local path (live) | yes (link target only) |
| `"copy"` | path / set | **place-once** copy (writable, user-managed) | **no** |
| `"copy"` | marker | evaluation-time error (contradictory) | — |

**copy is place-once and user-managed.** Once materialized, layat does not touch the target.
The store's read-only mode (`0444` / `0555`) is preserved but owner-write is added so the
copy is editable. To follow an upstream `src` update, use `layat apply --recopy` (overwrites
all copy targets unconditionally) or `layat reset` then re-apply. Copies are not generation-
managed and are never rolled back. If a foreign real file already occupies a copy target,
layat skips it but emits a **warning** (it never overwrites your file).

### Generating entries dynamically

Interpolate names into the target key. Derive the target from a variable **you** control —
not from `baseNameOf src` (a store path resolves to `/nix/store/<hash>-source`, so
`baseNameOf` yields `<hash>-source`):

```nix
let plugins = [ "telescope" "treesitter" "cmp" ]; in
layat.lib.mkManifest {
  inherit pkgs;
  root = layat.lib.homeRoot;
  entries = builtins.listToAttrs (map (n: {
    name  = ".local/share/nvim/site/pack/plugins/start/${n}";  # key = target
    value = { src = inputs.${n}; };
  }) plugins);
}
```

To enumerate subdirectories, `builtins.readDir` an **already-realised store path /
`flake = false` input** (a raw `fetchFromGitHub` derivation would trigger import-from-
derivation and break pure flake eval):

```nix
let
  skills = builtins.readDir "${inputs.claude-skills}/skills";
  names  = builtins.attrNames (nixpkgs.lib.filterAttrs (_: t: t == "directory") skills);
in
layat.lib.mkManifest {
  inherit pkgs;
  root = layat.lib.homeRoot;
  entries = builtins.listToAttrs (map (n: {
    name  = ".claude/skills/${n}";
    value = { src = inputs.claude-skills; subpath = "skills/${n}"; };
  }) names);
}
```

---

## Command reference

The CLI discovers the entrypoint in the CWD (`flake.nix` → `shell.nix` → `default.nix`),
overridable with `-f`. Each `layat.<name>` is an independent profile; `<name> = default` is
resolved by `layat apply` when the name is omitted.

```bash
layat apply [<name>]            # apply layat.<name> (omitted = layat.default); builds, commits a new generation, places
layat apply <name> --dryrun     # read-only plan: place/replace/remove/conflict/no-op, zero side effects
layat apply <name> --recopy     # also overwrite every copy target from its src unconditionally
layat apply --manifest <farm>   # apply a pre-built link-farm directly (no entrypoint discovery / eval / build)
layat apply --all               # apply every layat.* in lexicographic order; continues past failures
layat apply --all --project-root # apply only projectRoot configs (also --home-root / --system-root)
layat reset <name> [target...]  # tear down placements (no profile change); target omitted = all entries
layat reset <name> --dryrun     # show what would be removed; zero side effects
layat rollback <name>           # roll back to the previous generation (home mode only; name required)
layat list-generations <name>   # list generations (home mode only)
layat list-generations --all    # list generations for all home-mode configs
layat gitignore <name>          # print placement targets for .gitignore to stdout (no writes; project mode only)
layat gitignore --all           # sorted + deduped targets for all projectRoot configs
layat prune                     # delete the orphan profile series whose recorded root is gone (no name; lists the roots and confirms)
layat prune --dryrun            # show the series that would be deleted; zero side effects
layat init <template>           # wrapper over `nix flake init -t github:yasunori0418/layat#<template>`
```

### Global flags

```text
-f, --file <path>   # specify the entrypoint explicitly (overrides auto-discovery)
--root <path>       # override the resolved root in any mode
--no-wait           # on lock contention, skip instead of waiting (for shellHook; explicit apply blocks by default)
-v, --verbose       # print the placement report (summary + per-target lines); default is silent on success
--version           # print the embedded version and exit (cobra default format `layat version X.Y.Z`; no short flag, since -v is --verbose)
--debug             # reveal the internal nix commands on stderr (for troubleshooting)
--project-root      # --all qualifier: only projectRoot configs (also --home-root / --system-root)
--recopy            # apply qualifier: overwrite every copy target from src
--manifest <path>   # apply only: apply a pre-built link-farm directly
-y, --yes           # skip the confirmation prompt of a destructive command (reset / prune; for scripts / CI)
--generator <name>  # manifest generator (only `nix` today; see "Choosing the generator")
```

### Choosing the generator

The generator is chosen explicitly, never guessed from the files present; the first of these wins:

1. `--generator <name>`
2. the `LAYAT_GENERATOR` environment variable
3. `layat.toml` in the `-f` directory (the file's directory if `-f` is a file) or the CWD; parents are not searched
4. `$XDG_CONFIG_HOME/layat/config.toml` (default `~/.config/layat/config.toml`)
5. the default `nix` (the only generator today)

```toml
# layat.toml / config.toml — `generator` is the only key
generator = "nix"
```

An empty value or a settings file without `generator` counts as unset. An unknown key, a TOML
syntax error, or an unknown generator name exits 1 with one line on stderr (`E_INPUT` under
`--json`). `apply --manifest` reads none of these and rejects `--generator`, `-f`, and `--all`;
`prune` and `init` ignore the mechanism.

### Output and exit codes

- **Silent on success by default.** The placement report, try-lock skip notices, and the
  `apply --all` summary need `-v`. `--debug` shows the internal nix commands. No `--quiet`.
- **`--json`** writes a [outturn](https://github.com/yasunori0418/outturn)-conformant JSON
  envelope (a single document) to stdout at command completion — the machine-readable
  second contract, orthogonal to `-v` (every subcommand carries its payload; `--all` lists one
  `SubjectResult` per config. `reset --dryrun` is the one path still on the minimal shape).
- **Stream discipline**: stdout is reserved for machine-readable output (`gitignore`
  listings, `apply --dryrun` plans) — printed even at the default verbosity, so
  `layat gitignore <name> >> .gitignore` and `layat apply <name> --dryrun | ...` pipe safely.
  **Warnings and errors always go to stderr** and are never silenced.

| Exit code | Meaning |
|---|---|
| `0` | success / no-op / `--no-wait` try-lock skip |
| `1` | general error (eval error, engine runtime error, `apply --all` partial failure) |
| `2` | `apply --dryrun` detected a conflict (usable as a CI pre-gate) |

### Behavior notes

- **Idempotent.** Only stale links recorded as placed and still matching the record are removed.
  A foreign symlink is replaced with a warning; a real file or directory at the target is an error.
- **Generations** ride on layat's own Nix profile; use `nix-env` / `nix-collect-garbage` on it.
  project mode commits no new generation when the link-farm is unchanged, but repairs drift.
- **`apply --all`** applies each config atomically on its own profile, continues past
  failures, and exits non-zero if any failed. It is **not** atomic as a whole.
- **`reset`** removes layat-managed symlinks and copy targets (the only way to remove a copy).
  It requires a name and confirmation or `-y`, and leaves the profile/generations untouched.

---

## Comparison with other tools

The axis is not "feature presence" but **"hide placement behind a module abstraction, or
expose it as a pure function you control."**

| Tool | Role | Approach | Difference from layat |
|---|---|---|---|
| npins / niv | source version pinning | — | does not place files (orthogonal — compose with layat) |
| home-manager `home.file` | file placement + generations | module (generate / declare) | requires HM; whole-environment model; the file module cannot be extracted standalone |
| `mkOutOfStoreSymlink` (HM) | out-of-store symlink | helper inside a module | HM-only; layat provides an equivalent as a dependency-free explicit function |
| nixpkgs `linkFarm` / `symlinkJoin` | store-internal symlink trees | pure function | output stays *inside* the store; never placed at an arbitrary out-of-store path (layat uses it internally) |
| `nix profile` | generation management | — | placement target is fixed at `~/.nix-profile`; no arbitrary-path placement (layat rides on it) |
| `systemd.tmpfiles` (`L`) | declarative symlink to an arbitrary path | module (NixOS) | low-level, NixOS-only, no copy / generations / fetch abstraction |
| numtide/system-manager | non-NixOS `/etc` + systemd + packages | module (`lib.evalModules`) | overlapping domain but the **opposite** approach; no arbitrary-path placement, HOME dotfiles, or subdirectory extraction |
| `git clone` (shell) | clone and place | imperative | no reproducibility or Nix integration |
| **layat** | independent placement of fetched sources + generations + explicit out-of-store | **pure function, user-managed** | — |

---

## MVP status

| Area | Status |
|---|---|
| Standalone CLI (`apply` / `reset` / `rollback` / `list-generations` / `gitignore` / `prune` / `init`) | implemented (core) |
| project mode (`projectRoot`) | implemented (core) |
| home mode (`homeRoot`) | implemented |
| home-manager module | implemented — single profile (fixed name `default`); no role separation |
| Generations / rollback (home mode) | implemented |
| copy (place-once) / out-of-store symlink | implemented |
| flake-parts module | implemented |
| `manifest.json` schema | v1 only; no migration / backward-compat machinery yet |
| `--json` machine-readable output | implemented — outturn-conformant envelope on every subcommand, with per-command payloads (items / changes / info) and one `SubjectResult` per config under `--all`; `reset --dryrun` still emits the minimal shape |
| NixOS / nix-darwin modules | future |
| system mode (`systemRoot` = `/`) | future (seam only; evaluation-time error if selected today) |

**Known limitations / honest caveats**

- Boot / init / filesystem / partition layers are not layat's domain.
- Removing a clone leaves an orphan profile directory under
  `<state>/nix/profiles/layat/` (the store is freed by `nix-collect-garbage`, but the
  profile directory remains). `layat prune` deletes the series whose recorded root no longer
  exists; it lists the root paths and asks before deleting anything.
- The home-manager module cannot separate roles into multiple profiles in the MVP — use the
  standalone CLI for that.

---

## Migrating from nput

nput has been renamed to **layat**. The rename is a breaking change
and **no compatibility shim is provided**. Why the name changed, the rejected candidates and the
policy behind the notice period are recorded in
[`docs/adr/0054-rename-nput-to-layat.md`](docs/adr/0054-rename-nput-to-layat.md).

### Staying on the old name

If you would rather not migrate yet, pin the last nput-named state:

```nix
inputs.nput.url = "github:yasunori0418/nput/legacy-nput";
```

`legacy-nput` is an annotated git tag on the merge commit of the notice PR — a tag, not a GitHub
Release. GitHub
keeps redirecting the old repository URL after the rename, so this pin keeps resolving.

### What you have to change

| Before | After |
|---|---|
| `github:yasunori0418/nput` | `github:yasunori0418/layat` |
| `nput.*` module options (`nput.enable`, `nput.entries`, `nput.backup.*`) | `layat.*` |
| `home.activation.nput` | `home.activation.layat` |
| `perSystem.nput.<name>` / `#nput.<system>.<name>` | `perSystem.layat.<name>` / `#layat.<system>.<name>` |
| `packages.nput` / the `nput` binary | `packages.layat` / the `layat` binary |
| `--json`: `E_NPUT_*` / `W_NPUT_*`, `tool.name = "nput"` | `E_LAYAT_*` / `W_LAYAT_*`, `tool.name = "layat"` |
| `<target>.nput-backup` | `<target>.layat-backup` |

The `--json` codes are the one change consumers cannot ignore: outturn requires the
`E_<TOOL>_<NAME>` shape, so the prefix moves with the tool name.

### Old generations: migrate or drop

nput's generations live in `<state>/nix/profiles/nput/`; layat reads
`<state>/nix/profiles/layat/`. Nothing is migrated for you: layat neither moves that directory
nor reads it. Decide which of the two you want **before the first `layat apply`** — the
migration below moves the whole directory into place and assumes
`<state>/nix/profiles/layat/` does not exist yet. Check that it doesn't, and move it aside if it
does: `mv` into an existing directory does not fail — it silently nests `nput/` inside it and
exits 0. Whichever you pick, any `<target>.nput-backup` files are left behind; remove them by
hand once you are satisfied with the new placement.

#### Migrate the generations

Moving the directory alone is not enough. The generation links are protected from the Nix
garbage collector by *indirect roots* under `/nix/var/nix/gcroots/auto/` that point at the
**old absolute paths**; after a plain `mv` every one of them dangles and the next GC removes
them, taking the moved link farms — including the previous generation's `manifest.json` —
with it. Re-register a root per generation link in the same step as the `mv`, so that no GC runs
in between. The re-registration recreates every link and resets its mtime; if the original
generation dates matter to you, copy the old directory aside first (`cp -a` preserves symlink
timestamps) and restore the moved links from that copy with `touch -h -r` afterwards.

```sh
old="${XDG_STATE_HOME:-$HOME/.local/state}/nix/profiles/nput"
new="${XDG_STATE_HOME:-$HOME/.local/state}/nix/profiles/layat"
mv "$old" "$new" && find "$new" -name 'profile-*-link' -type l \
  -exec sh -c 'nix-store --add-root "$1" --indirect -r "$(readlink "$1")" >/dev/null' _ {} \;
```

The old roots — the ones still pointing into `<state>/nix/profiles/nput/` — are left alone;
they disappear at the next GC.

Check the result with `layat list-generations <name>` — home mode only, as with `nput`
(`default` under home-manager; if you applied the config with `--root`, pass the same `--root`).
The old generations should be listed.

The previous generation's manifest came along, so the first `layat apply` after a migration is
an ordinary apply — no foreign warnings, and stale removal keeps working.

#### Drop the generations

If you do not care about the history, remove the old state directory:

```sh
rm -rf "${XDG_STATE_HOME:-$HOME/.local/state}/nix/profiles/nput"
```

**Delete it rather than leaving it behind.** As long as the generation links exist, their
indirect roots keep the old store paths alive and the garbage collector will never reclaim
them. Once the links are gone the `gcroots/auto/` entries are pruned at the next GC and the
store paths become collectable.

With no previous generation to read, the first `layat apply` starts at generation 1 and behaves
as if every target had been placed by a stranger:

- **symlink entries are overwritten**, last-write-wins, with a `W_LAYAT_FOREIGN_SYMLINK`
  warning. The run does not fail.
- **copy entries are skipped**, because a real file already occupies the target. Pass
  `--backup` to move the existing file aside and place the copy.
- **targets that existed only in an old nput generation are left in place.** Stale removal
  needs the previous generation's manifest, and layat has none.

---

## Documentation

`docs/` is a three-layer structure: **README → overview document → item**. Items hold the
normative content (one claim per Markdown file, with YAML frontmatter); the overview documents
(maintained in Japanese) give the big picture and an index into the items.

- `docs/concept.md` — concept (index into solution / use_case items and ADRs), philosophy, north-star, comparison
- `docs/design.md` — design (index into design items)
- `docs/spec.md` — specification (index into requirement items)
- `docs/glossary.md` — canonical English terminology

| Directory | Type | Prefix |
|---|---|---|
| `docs/solution/` | solution | `SOL` |
| `docs/use-cases/` | use_case | `UC` |
| `docs/requirements/` | requirement | `REQ` |
| `docs/design/` | design | `DSG` |
| `docs/infrastructure/` | infrastructure | `INF` |
| `docs/adr/` | adr (kept out of the tree) | `ADR` |

Items form a graph (`docs/model.yaml` defines the types and relations), traversed with `sara`
from the devShell:

```bash
nix develop ./dev --command sara check    # validate the whole graph
```

For querying the graph (`sara query`, which takes a full ID) see `docs/agents/domain.md`.

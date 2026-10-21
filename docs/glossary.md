# Glossary

Canonical English terms for layat. This is the authoritative spelling reference for README text, code comments, and command output. When a concept appears in prose or output, use the **canonical term** defined here and avoid the listed alternatives so wording stays consistent across the project.

Each entry fixes one canonical spelling. Definitions are intentionally short; for full rationale see the requirement items under `docs/requirements/` (reachable through the index in `docs/spec.md`) and `docs/adr/` (design decisions). The Japanese-language counterpart lives in `docs/glossary.ja.md`.

## Core placement abstraction

### placement primitive
The core of layat: a pure function that places a Nix store path at a `root`-relative `target`. It is not hidden behind a module abstraction — users compose it directly.
- **Avoid**: "placement framework", "configuration management" (layat does not generate configuration).

### engine
The placement core that owns both placement (native filesystem operations) and stale removal. It takes `manifest.json` as input, invokes only `git` (toplevel) as an external command, and is implemented as a Go library driven by the **layat CLI**. It leaves recording and switching generations to the **generation backend**. It does not generate a bash script per config. "Library" here means internal in-binary layering under `internal/`, not a publicly importable, reusable module — the stable surface is the `manifest.json` contract. The engine is stdlib-only.
- **Avoid**: "per-config generated bash script", "per-layer placement logic", "a single flat implementation fused with the CLI", "importing the engine as a public Go module".

### layat
The name of the tool itself. A compressed coinage of "**lay** \<src\> **at** \<target\>": it places src at target, as the manifest says. It follows the traditional UNIX way of compressing a verb phrase into a command name, as `chroot` (change root) and `getopt` (get options) do. The tool was formerly named **nput**, whose "n" stood for *nix*; that prefix stopped describing the tool once the engine's only contract became `manifest.json`, regardless of who generates it (→ [ADR-0054](adr/0054-rename-nput-to-layat.md)). Spell it lowercase, including at the start of a sentence.
- **Avoid**: "nput" for anything other than the old name in migration context; "Layat" / "LayAt" / "layAt"; reading the name as nix-specific.

### layat CLI
The primary user-facing UX; the `packages.layat` binary on `PATH`. It discovers an **entrypoint**, obtains a named manifest through a **generator** (`nix` by default), and has the engine place it. Subcommands include `apply [<name>]`, `apply --all`, `reset`, `rollback`, `list-generations`, `gitignore`, `prune`, and `init`.
- **Avoid**: describing a per-config `nix run .#x` wrapper as the primary UX; describing `apply` as "always builds the entrypoint" (a built link-farm can be applied with `--manifest`).

### generator
What the **layat CLI** obtains a manifest from. Its contract has four operations under `internal/generator`: Discover (find the entrypoint), Roots (return the root kind and root without building), Build (place a link-farm that can be committed as a generation — a store path — at the pending path), and DryBuild (return the store path without a gcroot). The nix implementation (nixgen, `internal/generator/nixgen`) is the default. A generator is chosen only explicitly: `--generator`, then `LAYAT_GENERATOR`, then `layat.toml`, then `$XDG_CONFIG_HOME/layat/config.toml`, then the default `nix`. The engine does not care which generator produced the manifest (→ [ADR-0055](adr/0055-manifest-generator-contract.md), [ADR-0056](adr/0056-generator-selection-and-config.md)).
- **Avoid**: saying the CLI runs `nix build` / `nix eval` itself (the nix generator does); saying the generator is guessed from the files present.

### prebuilt
The generator behind `apply --manifest <link-farm>`: Discover returns the given path, Roots reads that link-farm's `manifest.json`, and Build / DryBuild return the path. It is not one of the `--generator` choices (→ [ADR-0055](adr/0055-manifest-generator-contract.md)).
- **Avoid**: listing it as a `--generator` value; describing `--manifest` as a route separate from the generators.

### generation backend
What the **engine** leaves recording, listing, and switching generations to. Its contract has five operations under `internal/generation`, with paths and integers as the only inputs and outputs: Commit (record a link-farm as a new generation and point the profile at it), List (list the generations, marking the current one), Switch (point the profile at an existing generation), UserBase (the base the user-state profileDirs are keyed under), and SystemBase (the base the system-mode profileDirs are keyed under). The nix implementation (nixprofile, `internal/generation/nixprofile`) is the default and only backend, and the only place that runs `nix-env`; its bases are `<state>/nix/profiles/layat` and `/nix/var/nix/profiles/layat`. The CLI always injects nixprofile, and the engine errors when given no backend. The engine does not care who manages generations; the profileDir layout (`profile` / `profile-<N>-link` / `.pending`) is the fixed part of the contract that the engine reads directly. There is no flag, environment variable, or `layat.toml` key that chooses a backend, and no setting that overrides its bases (→ [ADR-0057](adr/0057-generation-backend-contract.md)).
- **Avoid**: saying the engine runs `nix-env`; describing a way to choose the backend or a setting for its bases.

### entrypoint
The Nix config file the layat CLI reads: one of `flake.nix`, `shell.nix`, or `default.nix`. It exposes a named manifest under `layat.<name>`. The config is still written in Nix and evaluated by `nix build`.
- **Avoid**: saying "layat discovers the config contents from the CWD" — it discovers the entrypoint *file*; the config is fixed by Nix evaluation.

### module
An integration layer such as standalone, home-manager, a future NixOS module, or a devShell `shellHook`. It is purely the wiring that kicks the engine; it never places files itself and never translates to `home.file` or `systemd.tmpfiles`.
- **Avoid**: "the module places files", "the module translates to native mechanisms".

## Placement input and output

### entry / entries
A placement definition. `entries` is an attrset keyed by `target`, where each value is an entry (`{ src; subpath?; target?; method?; }`). The attribute key is the identifier and supplies the default `target`. There is no manual `name` field; key uniqueness is guaranteed natively by Nix attrset keys.
- **Avoid**: "file entry" (the source may be a directory); giving an entry a `name` field; calling `entries` a list of `{ name; … }` (the old form).

### src
The placement source of an **entry** (which store path or repository). The default is a **store link** into the Nix store; out-of-store is opt-in via an explicit marker. It is orthogonal to **subpath**; do not read `src` as an abbreviation of `source`.
- **Avoid**: confusing it with **subpath**; calling it `source`.

### subpath
The relative path selecting which path inside **src** to take, for an **entry**. Works for both files and directories; default `"."`. Omitting `subpath` is the canonical way to select the whole repository (`subpath = "."` is the explicit form).
- **Avoid**: the old names `source` / `dir`; confusing it with **src**; assuming a dedicated token/marker is needed for "whole repo" (omission expresses it).

### target
The placement destination of an **entry**, specified as a path relative to **root**. The `entries` attribute key is the default `target` and is also the entry's identity (the diff key for stale removal, and its uniqueness key).

### root
The base path for placement. Chosen explicitly via the `root` argument of the public API (there is **no implicit default**). Its type is a union of `string` (absolute path, fixed at evaluation time) and `marker` (resolved at runtime). The markers are **projectRoot**, **homeRoot**, and **systemRoot**. Each marker also accepts the applied form `homeRoot { subdir = "..."; }`. The applied form does not move root; it is evaluation-time sugar equivalent to prefixing every target with `<subdir>/`.
- **Avoid**: "`$HOME` is fixed", "defaults to home mode", "the default when `root` is omitted"; describing a marker as "sugar that returns a path string" (it is a runtime-resolution kind, not expanded to a path at evaluation time); describing the applied form as "root actually resolves to the subdir" (`$HOME/.config` becoming root). Root and `rootKind` stay the same as the bare marker; only the targets change.

## Placement modes

### home mode
The placement mode where **root** = `$HOME`, selected explicitly with the `homeRoot` marker. Used by both standalone and modules such as home-manager. It commits a generation each time and exposes `--rollback` to the user.
- **Avoid**: "standalone-only", "the default when `root` is omitted".

### project mode
The placement mode where **root** = the project root, selected with the `projectRoot` marker. Its placements are **ephemeral placement** (not committed); generations and rollback are not exposed to the user. The root resolves to the git toplevel.
- **Avoid**: "relative to CWD", "relative to the config file".

### system mode
The placement mode where **root** = `/`, selected with the `systemRoot` marker. Used for the distro concept (root = `/`).

### projectRoot
The root marker selecting **project mode**. At runtime it resolves the git toplevel as **root** (overridable with `--root`). One of the root markers alongside `homeRoot` / `systemRoot`, following the same "pass a marker to opt into behavior" pattern as `mkOutOfStoreSymlink`. Also accepts the applied form `projectRoot { subdir = "..."; }` (see **root**).
- **Avoid**: interpreting it as pointing to the config file location.

### homeRoot
The root marker selecting **home mode**. At runtime it resolves `$HOME` as **root**. It promotes the previously implicit `$HOME` default into an explicit marker. The applied form `homeRoot { subdir = ".config"; }` keeps root at `$HOME` and places every target under `.config/`.
- **Avoid**: saying that the applied form resolves root to `$HOME/.config`.

### systemRoot
The root marker selecting **system mode** (root = `/`). Promotes what ADR-0004 called a "future absolute-path string seam" into a first-class marker. Also accepts the applied form `systemRoot { subdir = "..."; }` (see **root**).

### ephemeral placement
The nature of placements in **project mode**: regenerated on each clone and never committed to the project, so activation never touches git state.
- **Avoid**: confusing it with "vendoring" or "placement that commits artifacts".

## Placement kinds

### store link
The core, default placement: a symlink whose destination is a Nix store path. The default path that guarantees reproducibility. "Unification" means making the store the default/core and demoting out-of-store to an explicit exception.
- **Avoid**: confusing it with out-of-store symlink; calling it a "copy".

### out-of-store symlink
A live symlink to a local path, opted into only via an explicit marker. An explicit escape hatch, not a first-class feature. There are two markers. `layat.lib.mkOutOfStoreSymlink "/abs/path"` points to a local absolute path (for live editing of dotfiles under development).
For in-repo trees, the root-relative marker `layat.lib.mkRootRelativeSymlink "rel/path"` takes a path relative to the placement root; the engine resolves it against the root at runtime and creates a relative link from the target's parent.
- **Avoid**: treating it as default behavior; producing it via implicit branching on the type of `src`.

## State management

### generation
The unit of rollback. The default **generation backend**, nixprofile, manages it on layat's own Nix profile (the `nix-env --profile <dir>` style). nixprofile commits (`--set`) and switches (`--switch-generation`, which rollback also uses) generations with `nix-env --profile <dir>`, and lists them by scanning the generation links (`profile-<N>-link`) in the profileDir read-only, without `nix-env`; pruning (`--delete-generations`) and store GC (`nix-collect-garbage`) are left to the user's nix commands.
- **Avoid**: narrating it under a "stateless script" premise (reversed since the initial direction); managing it with the new `nix profile` CLI (which requires a profile-manifest and does not work on `nix-env --set` profiles).

### store manifest
The generation-derived record of "what layat placed". Concretely the `manifest.json` (carrying a `schemaVersion`) inside the link-farm derivation, generated by Nix (`lib.mkManifest`) and read by the Go engine — the Nix↔Go contract. It underpins the engine's conservative stale removal (it deletes only layat-managed symlinks that point where the record says, and never touches the user's real files).

### last-applied record
layat's own record of what it last placed under a root — a filesystem fact, not a generation. It lives at `<state>/layat/<key>/last-applied.json`, keyed like the profileDir (`<name>` for home without `--root`, otherwise `<roothash>/<name>` with the backref `.root`). It holds `recordVersion` (1), `linkFarm`, `appliedAt`, and `manifest` (the placed `manifest.json` embedded verbatim), so it does not depend on store GC. apply (including generation skip) and rollback rewrite it after the filesystem operations succeed and before the generation operation; reset rewrites it after removal and deletes it once every entry is gone. The baseline for stale removal, drift repair, reset's removal set, and rollback's departing generation is taken from **the record, else the profile's current generation, else empty**; when the record and the profile disagree on `linkFarm`, the record wins with a one-line warning. On the first apply (neither exists) the baseline is empty and a one-line "no last-applied record" warning is printed. A broken record is an error, not treated as absent (→ [ADR-0058](adr/0058-last-applied-record.md)).
- **Avoid**: calling the record a generation (the rollback target is read from the generation link); describing it as the flock key (that is the profileDir).

### method
The entry field that selects the placement kind (renamed from the old `mode`). Renamed to avoid misreading it as a unix file mode.
- **Avoid**: the old name `mode`.

## Flagged ambiguities

- **"symlink" alone is ambiguous.** Always disambiguate to either **store link** or **out-of-store symlink**. The default store link is itself realized as a symlink, so the bare word does not identify the kind.
- **"unification" is not "removal".** Store-link unification does not delete out-of-store; it demotes it from the default and isolates it behind an explicit function.
- **`src` and `subpath` are distinct.** `src` = which thing (store path / repository); `subpath` = which path inside it. Similar names, orthogonal concepts. Do not use the old name `source` (= today's `subpath`).
- **"standalone" is a launch form, not a placement mode.** Standalone (invoking the CLI directly) is orthogonal to placement mode (home / project / system); the mode is decided by the `root` marker.

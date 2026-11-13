---
id: "REQ-496b1a07-5b74-416b-9e5f-3952b4c03737"
type: requirement
name: "entrypoint は `layat.<name>` に named manifest を公開し CLI は形ごとの attr path で build する"
section: cli-entrypoint
status: stable
derives_from:
  - "UC-1c280dce-7c72-44c0-95ea-d06344f62a47"
specification: |
  An entrypoint SHALL expose a named manifest (the result of `mkManifest`) at
  `layat.<name>`, so that a single project can hold multiple configurations, each with its
  own independent profile. For a flake entrypoint the manifest SHALL be exposed at
  `outputs.layat.<system>.<name>` (written directly) or at `perSystem.layat.<name>` (via the
  flake-parts module), and the CLI SHALL build it with `nix build <ep>#layat.<system>.<name>`,
  interpolating the current `<system>`. For a `default.nix` / `shell.nix` entrypoint the
  manifest SHALL be exposed at `passthru.layat.<name>` of a `mkShell` derivation or at a
  top-level attrset `{ layat.<name> = ...; }`, and the CLI SHALL build it with
  `nix build -f <ep> layat.<name>`; a legacy entrypoint has no dimension corresponding to
  the flake `<system>`, so its addressing is the flat `layat.<name>`.
specification_ja: |
  entrypoint は `layat.<name>` に named manifest（`mkManifest` の結果）を公開しなければ
  ならない。これにより 1 プロジェクトに複数 config を置ける（役割ごと独立 profile）。
  flake entrypoint では `outputs.layat.<system>.<name>`（直書き）または
  `perSystem.layat.<name>`（flakeModule 経由）に公開しなければならず、CLI は現行 `<system>` を
  差し込んで `nix build <ep>#layat.<system>.<name>` で build しなければならない。`default.nix` /
  `shell.nix` entrypoint では `mkShell` derivation の `passthru.layat.<name>` または
  トップレベル attrset `{ layat.<name> = ...; }` に公開しなければならず、CLI は
  `nix build -f <ep> layat.<name>` で build しなければならない。legacy entrypoint には flake の
  `<system>` に相当する次元が無く、addressing はフラットな `layat.<name>` になる。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:29:58+09:00"
---
# entrypoint は `layat.<name>` に named manifest を公開し CLI は形ごとの attr path で build する

## 仕様

entrypoint は `layat.<name>` に named manifest（`mkManifest` の結果）を公開する。
1 プロジェクトに複数 config を置ける（役割ごと独立 profile・→ concept.md）。

| entrypoint | 公開形 | CLI が叩く build |
|---|---|---|
| `flake.nix`（直書き）| `outputs.layat.<system>.<name> = mkManifest { ... }` | `nix build <ep>#layat.<system>.<name>`（CLI が現行 `<system>` を差し込む）|
| `flake.nix`（flake-parts）| `perSystem.layat.<name> = mkManifest { inherit pkgs; ... }`（flakeModule 経由）| `nix build <ep>#layat.<system>.<name>`（同上）|
| `default.nix` / `shell.nix`（passthru・canonical）| `pkgs.mkShell { ...; passthru.layat.<name> = mkManifest { ... }; }` | `nix build -f <ep> layat.<name>` |
| `default.nix` / `shell.nix`（トップレベル attrset）| `{ layat.<name> = mkManifest { ... }; }` | `nix build -f <ep> layat.<name>` |

legacy entrypoint には flake の `<system>` に相当する次元が無く、addressing は
フラットな `layat.<name>` になる。

> **上の表は原文の写しで、規範は frontmatter が正**。表が併記する次の点は本 item の
> 規範ではない。
>
> - flake-parts の使い方が満たすべき条件（transpose・`pkgs` の一貫性）→ [flake-parts での使い方は直書きと同一の derivation を生み CLI のアドレッシングを変えない](flake-parts-module-path.md)
> - passthru 形を canonical とする理由と `mkShell` との共存条件 → [legacy entrypoint は mkShell passthru 形を canonical とし CLI の attr path を分岐させない](legacy-entrypoint-passthru-canonical.md)

## 出典

`docs/spec.md`「CLI 仕様」→「アドレッシング（`layat.<name>`）」の本文と表。

決定の実体は [named manifest を専用の `layat.<name>` 名前空間に公開する](../adr/layat-name-namespace.md)で、
legacy entrypoint の addressing を `nix build -f <ep> layat.<name>` に統一したのは
[legacy entrypoint の eval と build は新 CLI の -f 形で行う](../adr/legacy-eval-via-nix-file-flag.md)（旧決定の `nix-build -A` 記述を改訂）。

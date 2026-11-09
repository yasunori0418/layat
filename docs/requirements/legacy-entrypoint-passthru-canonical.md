---
id: "REQ-c890ce4a-6528-4ab3-ac86-23d7aebff7da"
type: requirement
name: "legacy entrypoint は mkShell passthru 形を canonical とし CLI の attr path を分岐させない"
status: stable
derives_from:
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  For `shell.nix` / `default.nix`, the canonical form SHALL be a passthru form that
  coexists with the nixpkgs `mkShell`: `shell.nix` SHALL return a plain `mkShell`
  derivation and carry the named manifest at `passthru.layat.<name>`, so that it is
  compatible with bundling the CLI into the `project` template devShell. Because
  `passthru` is merged into the attrset of the host derivation, the attr path the CLI
  invokes (`layat.<name>`) SHALL be identical to that of the top-level attrset form, and
  the CLI SHALL NOT branch its implementation between the two. A bare `nix-shell` (without
  `-A`) SHALL keep working for both forms. Handling multiple manifests at once SHALL be
  served by the existing batched eval of `apply --all`, and no helper that composes
  manifests via `inputsFrom` of `mkShell` SHALL be provided, because it would conflict
  with the "one profile = one config" atomicity.
specification_ja: |
  `shell.nix` / `default.nix` は nixpkgs `mkShell` と共存できる passthru 形を canonical と
  しなければならない。`shell.nix` は素の `mkShell` derivation を返し、
  `passthru.layat.<name>` に named manifest を載せなければならない（`project` template の
  devShell 同梱と両立する）。`passthru` はホスト derivation の attrset にマージされるため、
  CLI が評価する attr path（`layat.<name>`）はトップレベル attrset 形と同一でなければならず、
  CLI は実装分岐を持ってはならない。素の `nix-shell`（`-A` 不要）はどちらの形でも使えなくなっては
  ならない。複数 manifest の一括処理は既存の `apply --all` の一括 eval でそのまま充足させ
  なければならず、`mkShell` の `inputsFrom` で manifest を合成するヘルパは持ってはならない
  （1 profile = 1 config の atomic 性と衝突するため）。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-05T00:29:04+09:00"
---
# REQ-c890ce4a-6528-4ab3-ac86-23d7aebff7da: legacy entrypoint は mkShell passthru 形を canonical とし CLI の attr path を分岐させない

## 仕様

`shell.nix` / `default.nix` は **nixpkgs `mkShell` と共存できる passthru 形**を canonical と
する。`shell.nix` は素の `mkShell` derivation を返し、`passthru.layat.<name>` に named
manifest を載せる（`project` template の devShell 同梱と両立する）。`passthru` はホスト
derivation の attrset にマージされるため、**CLI が評価する attr path（`layat.<name>`）は
トップレベル attrset 形と同一**で実装分岐を持たない。素の `nix-shell`（`-A` 不要）は
どちらの形でも壊れない。

```nix
# shell.nix（canonical: passthru 形）
{ pkgs ? import <nixpkgs> {}, layat ? import (fetchTarball "https://github.com/yasunori0418/layat/archive/....tar.gz") }:
pkgs.mkShell {
  packages = [ ];
  shellHook = "layat apply skills --no-wait";
  passthru.layat = {
    skills = layat.lib.mkManifest { inherit pkgs; root = layat.lib.projectRoot; entries = { }; };
  };
}
```

複数 manifest の一括処理は既存の `apply --all` の一括 eval
（`nix eval -f <ep> layat --apply … --json`）でそのまま充足し、`mkShell` の `inputsFrom` で
manifest を合成するヘルパは持たない（1 profile = 1 config の atomic 性と衝突するため）。

> **上は原文の写しで、規範は frontmatter が正**。原文が併記する次の点は本 item の
> 規範ではない。
>
> - legacy entrypoint の addressing がフラットな `layat.<name>` になること → [entrypoint は `layat.<name>` に named manifest を公開し CLI は形ごとの attr path で build する](named-manifest-addressing.md)
> - `apply --all` の一括 eval そのもの → [apply --all は rootKind を 1 回の一括 eval で取る](apply-all-batched-eval.md)
> - 1 profile = 1 config の atomic 性（→ [1 config = 1 profile を全モードで自前に持つ](../adr/one-config-one-profile-all-modes.md)）→ [世代は link farm derivation を layat 自前 profile へコミットして積み、前世代 manifest から古くなった配置を除去する](generation-mechanism-link-farm-manifest.md)

## 出典

`docs/spec.md`「CLI 仕様」→「アドレッシング」の blockquote「legacy entrypoint の
canonical 形」。

決定の実体は [legacy entrypoint は mkShell の passthru 形を正準にする](../adr/legacy-entrypoint-mkshell-passthru-canonical.md)・[CLI が評価する attr path を legacy entrypoint の書き方で分岐させない](../adr/attr-path-independent-of-entrypoint-shape.md)・[複数の manifest を inputsFrom で合成するヘルパを持たない](../adr/no-inputsfrom-manifest-composition-helper.md)。

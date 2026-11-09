---
id: "REQ-14f0aec9-abae-4621-82f3-40536a1ad904"
type: requirement
name: "layat CLI は PATH 常駐の一次 UX で、project mode は devShell 同梱を canonical とする"
status: stable
derives_from:
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  The `layat` CLI (`packages.layat`) SHALL be the primary UX, installed on PATH. Two
  installation forms SHALL be supported. In the standalone (home mode) form the CLI is
  installed globally (e.g. `nix profile install github:yasunori0418/layat`); the global CLI
  and the `layat.lib` pinned by the user flake are separate inputs and their `schemaVersion`
  MAY skew, so aligning the CLI and the flake pin to the same input SHOULD be recommended,
  and bundling the CLI into a devShell SHALL NOT be required for this form. In the project
  mode (canonical) form the pinned `layat` SHALL be bundled into the devShell of
  `templates/project` (`packages = [ layat.packages.${system}.layat ]`), so that the CLI and
  `layat.lib` come from the same flake input and their `schemaVersion` agrees. How the
  engine reacts to a skewed `schemaVersion` is stated by the manifest schema
  specification and is not restated here.
specification_ja: |
  `layat` CLI（`packages.layat`）は PATH に常駐する一次 UX でなければならない。導入形は
  2 つを支援しなければならない。CLI 単体（home mode）ではグローバルに導入し（例:
  `nix profile install github:yasunori0418/layat`）、グローバル CLI とユーザー flake が
  pin する `layat.lib` は別入力になるため `schemaVersion` が skew し得る。よって CLI と
  flake pin の `layat` を同一 input から揃えることを推奨すべきであり、この形では
  devShell 同梱を強制してはならない。project mode（canonical）では `templates/project` の
  devShell に pin 版
  `layat` を同梱し（`packages = [ layat.packages.${system}.layat ]`）、CLI と `layat.lib`
  が同一 flake 入力から来て `schemaVersion` が一致するようにしなければならない。
  skew した `schemaVersion` に engine がどう反応するかは manifest スキーマ仕様の担当で、
  本 item では規定しない。
risk_exempt: "配布形態の宣言であり、テストで確かめる振る舞いが無い。"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T23:05:37+09:00"
---
# layat CLI は PATH 常駐の一次 UX で、project mode は devShell 同梱を canonical とする

## 仕様

`layat` CLI（`packages.layat`）は PATH に常駐する一次 UX。

- **standalone（home mode）**: `nix profile install github:yasunori0418/layat` 等で
  グローバルに導入する。グローバル CLI とユーザー flake が pin する `layat.lib`
  （manifest 生成側）は別入力になり `schemaVersion` が skew し得る。skew を避けるには
  **CLI と flake pin の `layat` を同一 input から揃える**ことを推奨する。project mode の
  ような devShell 同梱は強制しない（PATH 常駐の利便を優先）。
- **project mode（canonical）**: `templates/project` の devShell に **pin 版 `layat` を同梱**する
  （`packages = [ layat.packages.${system}.layat ]`）。`nix develop` / direnv 入室時に
  flake.lock で固定した `layat` が PATH に載り、CLI と `layat.lib`（manifest
  `schemaVersion`）が同一 flake 入力から来て一致する。グローバル install 依存だと
  flake.lock の pin を CLI で破り version skew を招くため、project mode では devShell
  同梱を canonical とする。

> **上は原文の写しで、規範は frontmatter が正**。原文が併記する次の 2 点は本 item の
> 規範ではない。
>
> - engine が自身の対応版より新しい `schemaVersion` を拒否すること（→ [manifest.json を schemaVersion 1 の唯一の契約として固定する](../adr/manifest-json-contract-schemaversion-v1.md)）は
>   [manifest.json が Nix と engine の唯一の安定契約であり schemaVersion は 1 に固定する](manifest-single-stable-contract.md)、MVP が v1 のみであることは [MVP は manifest v1 のみを発行・受理しマイグレーション機構を持たない](manifest-v1-only-no-migration.md) の担当
> - devShell 同梱そのものの配線（`templates/project` の内容）→
>   「`layat init`」→「テンプレートの内容」節の担当（[template は動く example を 1 config だけ置きバリエーションはコメントで示す](init-template-contents.md)）

## 出典

`docs/spec.md`「CLI 仕様（一次 UX）」の導入 2 項目。

決定の実体は [CLI が一次 UX で devShell に pin 版を同梱する](../adr/cli-primary-ux-devshell-pinned.md)（PATH 常駐・
`packages.layat`・project mode では devShell 同梱が canonical）。

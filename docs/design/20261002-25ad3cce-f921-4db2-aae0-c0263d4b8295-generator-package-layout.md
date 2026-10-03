---
id: "DSG-25ad3cce-f921-4db2-aae0-c0263d4b8295"
type: design
name: "internal/generator は契約・prebuilt・Fake を親パッケージに、nix 実装を nixgen サブパッケージに置き、cmd 層は注入点だけを持つ"
satisfies:
  - "REQ-194e4209-d804-4a4b-a2b8-3d39c6c33729"
  - "REQ-f4d7d4ab-fbdb-48c6-b29f-08dd88e72645"
---
# DSG-25ad3cce-f921-4db2-aae0-c0263d4b8295: internal/generator は契約・prebuilt・Fake を親パッケージに、nix 実装を nixgen サブパッケージに置き、cmd 層は注入点だけを持つ

## 設計

```
cmd/layat                       ← 生成器の選択と注入だけ（newGenerator()）
  │ import
  ├─ internal/generator         ← 契約（Generator インターフェイス・generator.Error）
  │                               + prebuilt 実装（--manifest）
  │                               + テストダブル Fake（fake.go・非 _test ファイル）
  ├─ internal/generator/nixgen  ← nix 実装（entrypoint 発見・system・nix eval / build・
  │                               nix 固有の文字列判定と案内文）
  └─ internal/engine            ← 変更しない。Build / DryBuild を既存の BuildFunc seam へ渡す
```

- **契約と nix 実装をパッケージで分ける**。`nixgen` は `generator` を import する側に立ち、
  逆向きの依存を持たない。nix 固有の知識（`layat.<system>.<name>` の attr path・`currentSystem`・
  experimental-features の判定と案内文）は `nixgen` の外へ出ない。
- **prebuilt と Fake は親パッケージに置く**。どちらも nix を呼ばず契約だけで書けるため、
  親パッケージに置いても nix 依存は入らない。`Fake` を非 `_test` ファイルに置くのは、
  `cmd/layat` のテストからも import して使うため（`_test.go` のシンボルは他パッケージから見えない）。
- **cmd 層は `newGenerator()` の注入点だけを持つ**。選択の解決（REQ-badc7e10-0ba4-40d4-b334-6e40193119db）の結果で
  nixgen / prebuilt を組み立て、各サブコマンドへ渡す。サブコマンドは `Generator` インターフェイス
  だけに依存し、nix を直接呼ばない。
- 2 層構成の境界（CLI と engine の間は `manifest.json`）は変えない。生成器は CLI 層の内側の
  構成で、engine から見た入力は従来どおり link-farm の store path と `manifest.json`。

## 出典

ADR-0055「manifest 生成器の契約」§1 と、Issue #395 の計画（2026-09-26）の裁定「テストダブル
`Fake` は `internal/generator/fake.go`（非 `_test` ファイル）に置く」。

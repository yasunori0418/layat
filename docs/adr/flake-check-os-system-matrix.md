---
id: "ADR-f2e72cfa-f46e-46b7-9d50-a98d60d8980f"
type: adr
name: "flake check を 3 環境の os×system matrix で回す"
section: ci-release-docs
status: stable
origin: "参照: PR #40、https://github.com/yasunori0418/cryoflow の CI 構成。ADR-0027 §1 を書き直した"
justifies:
  - "INF-d1230e1f-8ba8-49d8-8386-409bfbb7dd27"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T17:26:22+09:00"
---
# flake check を 3 環境の os×system matrix で回す

- ステータス: stable
- 日付: 2026-11-03
- 起点: 参照: PR #40、https://github.com/yasunori0418/cryoflow の CI 構成。ADR-0027 §1 を書き直した

## 背景

lib のテストと engine の Go の静的解析は `nix flake check` に集約してある。flake の `perSystem` は
4 つの system を定義しているが、CI のランナーで用意できる環境は限られる。どの system の上で
flake check を実行するかを決める必要がある。

## 決定

`nix flake check -L` を、x86_64-linux・aarch64-linux・aarch64-darwin の 3 環境の os×system
matrix で実行する。x86_64-darwin は CI の対象から外す。`perSystem` の 4 system の定義は残す。

ランナーと system の対応は [CI パイプライン（flake check の os×system マトリクスと E2E ジョブ）](../infrastructure/ci-pipeline.md) に置く。

## 根拠

「非 NixOS でも nix さえあれば動く」という主張は、Linux と macOS、x86 と arm の差で崩れうる。
実際にサポートする 3 環境で flake check を回すのが、評価やビルドの差による回帰を防ぐ最小の
保険になる。x86_64-darwin は GitHub ホストの標準ランナーが乏しく、安定して回せない。

## 棄却した代替案

### `ubuntu-latest` の 1 環境だけで回す

aarch64 や darwin に固有の評価・ビルドの回帰を検出できない。

### x86_64-darwin を含む 4 system 全部で回す

GitHub ホストの標準 x86_64 macOS ランナーが乏しく、安定して運用しにくい。

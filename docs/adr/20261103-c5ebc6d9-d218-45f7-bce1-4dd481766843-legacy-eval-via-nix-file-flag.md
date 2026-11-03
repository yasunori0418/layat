---
id: "ADR-c5ebc6d9-d218-45f7-bce1-4dd481766843"
type: adr
name: "legacy entrypoint の eval と build は新 CLI の -f 形で行う"
status: 採用
issues:
  - "#112"
  - "#107"
origin: "ADR-0032 §4 を書き直した"
justifies:
  - "REQ-496b1a07-5b74-416b-9e5f-3952b4c03737"
  - "DSG-92f54490-872a-42ac-bbd7-d06e9ee381c6"
---
# ADR-c5ebc6d9-d218-45f7-bce1-4dd481766843: legacy entrypoint の eval と build は新 CLI の -f 形で行う

- ステータス: 採用
- 日付: 2026-11-03
- 関連: GitHub Issue #112, #107
- 起点: ADR-0032 §4 を書き直した

## 背景

flake entrypoint の eval と build は、`nix-command` を前提にした新しい nix CLI（`nix eval` /
`nix build`）で行う。legacy entrypoint には当初、旧 CLI の `nix-build -A layat.<name>` を使う
想定があった。

## 決定

legacy entrypoint の eval と build は、旧 CLI ではなく新 CLI の `-f` 形（`nix eval -f <ep>` /
`nix build -f <ep>`）で行う。`--impure` は付けない。

用途ごとのコマンド形は REQ-496b1a07-5b74-416b-9e5f-3952b4c03737 に、nix を呼び出す処理を flake と
共通で使う設計は DSG-92f54490-872a-42ac-bbd7-d06e9ee381c6 に置く。

## 根拠

flake の側はすでに新 CLI を使っている。legacy だけ旧 CLI を混ぜると、experimental-features が
有効でないときの案内などのエラー処理を新旧の両方に持つことになる。新 CLI に揃えれば、nix を
呼び出す処理とエラー処理をそのまま共有でき、実装もテストも単純になる。

`-f` による評価は pure-eval を強制しない（強制するのは flake の評価だけである）。そのため
`<nixpkgs>` などの NIX_PATH の参照は `--impure` なしで解決できる。

## 棄却した代替案

### 旧 CLI の `nix-build -A` を使う

新 CLI を前提にしている flake の側と食い違い、エラー処理を二重に持つことになる。

### legacy の eval と build に `--impure` を自動で付ける

`-f` の評価では不要である。利用者が設定した pure-eval を黙って上書きすることにもなる。

---
id: "ADR-c5ebc6d9-d218-45f7-bce1-4dd481766843"
type: adr
name: "legacy entrypoint の eval と build は新 CLI の -f 形で行う"
section: cli-commands
status: stable
sources:
  - id: "issue-112"
    resource: "https://github.com/yasunori0418/layat/issues/112"
    title: "#112"
  - id: "issue-107"
    resource: "https://github.com/yasunori0418/layat/issues/107"
    title: "#107"
origin: "ADR-0032 §4 を書き直した"
justifies:
  - "REQ-496b1a07-5b74-416b-9e5f-3952b4c03737"
  - "DSG-92f54490-872a-42ac-bbd7-d06e9ee381c6"
references:
  - "ADR-28605860-770e-41df-95dd-7dc05b20497a"
  - "ADR-6b7becf0-f029-41e3-895e-b672dce94d3d"
  - "ADR-c7ffc21a-4be1-4468-b63c-2f47eedd9ab3"
  - "ADR-e683ac05-c0e0-41c1-8b6a-38b605972fe1"
  - "ADR-f40c3d37-638d-420a-a0fe-3908cbf0aec0"
  - "ADR-fe035038-4415-4a1e-aa5c-509fdfce9955"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:28:06+09:00"
---
# legacy entrypoint の eval と build は新 CLI の -f 形で行う

- ステータス: stable
- 日付: 2026-11-03
- 関連: [apply --all は config ごとの情報を 1 回の eval でまとめて取る](apply-all-single-eval.md), [CLI は cobra で組む](cobra-cli-framework.md), [root の必須性を別の確認を書かずに標準のエラーで担保する](root-required.md), [named manifest を専用の `layat.<name>` 名前空間に公開する](layat-name-namespace.md), [home mode の CLI と lib の版ずれを manifest の確認に任せて許容する](home-mode-skew-guarded-by-schemaversion.md), [1 config = 1 profile を全モードで自前に持つ](one-config-one-profile-all-modes.md), GitHub Issue #112, #107
- 起点: ADR-0032 §4 を書き直した

## 背景

flake entrypoint の eval と build は、`nix-command` を前提にした新しい nix CLI（`nix eval` /
`nix build`）で行う。legacy entrypoint には当初、旧 CLI の `nix-build -A layat.<name>` を使う
想定があった。

## 決定

legacy entrypoint の eval と build は、旧 CLI ではなく新 CLI の `-f` 形（`nix eval -f <ep>` /
`nix build -f <ep>`）で行う。`--impure` は付けない。

用途ごとのコマンド形は [entrypoint は `layat.<name>` に named manifest を公開し CLI は形ごとの attr path で build する](../requirements/named-manifest-addressing.md) に、nix を呼び出す処理を flake と
共通で使う設計は [legacy entrypoint の分岐は attr path 組み立てに閉じ、nix 呼び出しヘルパを共通で再利用する](../design/legacy-nix-invocation-reuse.md) に置く。

## 根拠

flake の側はすでに新 CLI を使っている。legacy だけ旧 CLI を混ぜると、experimental-features が
有効でないときの案内などのエラー処理を新旧の両方に持つことになる。新 CLI に揃えれば、nix を
呼び出す処理とエラー処理をそのまま共有でき、実装もテストも単純になる。

pure-eval を強制するのは flake の評価だけで、`-f` による評価は強制しない。そのため
`<nixpkgs>` などの NIX_PATH の参照は `--impure` なしで解決できる。

## 棄却した代替案

### 旧 CLI の `nix-build -A` を使う

新 CLI を前提にしている flake の側と食い違い、エラー処理を二重に持つことになる。

### legacy の eval と build に `--impure` を自動で付ける

`-f` の評価では不要である。利用者が設定した pure-eval を知らせずに上書きすることにもなる。

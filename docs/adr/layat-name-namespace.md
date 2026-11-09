---
id: "ADR-e683ac05-c0e0-41c1-8b6a-38b605972fe1"
type: adr
name: "named manifest を専用の `layat.<name>` 名前空間に公開する"
status: stable
sources:
  - id: "issue-4"
    resource: "https://github.com/yasunori0418/layat/issues/4"
    title: "#4"
origin: "Issue #4（flake.nix 以外の entrypoint）。ADR-0007 §4 を書き直した"
justifies:
  - "REQ-496b1a07-5b74-416b-9e5f-3952b4c03737"
  - "REQ-205d744d-5a53-4511-bc09-892ba01d4e6f"
references:
  - "ADR-28605860-770e-41df-95dd-7dc05b20497a"
  - "ADR-955c2601-88fb-4be8-85de-ec7d0a8c727d"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:28:06+09:00"
---
# named manifest を専用の `layat.<name>` 名前空間に公開する

- ステータス: stable
- 日付: 2026-11-01
- 関連: [apply --all は config ごとの情報を 1 回の eval でまとめて取る](apply-all-single-eval.md), [shell.nix / default.nix の再現性は利用者の責任にする](shell-default-nix-best-effort.md), GitHub Issue #4
- 起点: Issue #4（flake.nix 以外の entrypoint）。ADR-0007 §4 を書き直した

## 背景

1 つのプロジェクトに役割ごとの config を複数置き、それぞれを独立した profile で扱いたい。CLI は
entrypoint から、適用する manifest を名前で引ける必要がある。

flake の標準 output である `packages.<system>.<name>` に置けば、`nix build .#<name>` で system が
自動で補われる。一方で manifest が通常のパッケージとして `nix flake show` や `nix build` の対象に並ぶ。

## 決定

entrypoint は named manifest を専用の `layat` 名前空間の `layat.<name>` に公開する。flake では
`layat.<system>.<name>` とし、system は CLI が補う。shell.nix / default.nix での公開形は [legacy entrypoint は mkShell の passthru 形を正準にする](legacy-entrypoint-mkshell-passthru-canonical.md) が
決める。

entrypoint の形ごとの公開形と build の attr path は [entrypoint は `layat.<name>` に named manifest を公開し CLI は形ごとの attr path で build する](../requirements/named-manifest-addressing.md) に、
名前空間を分けて `packages` に manifest を含めないことは [config 名 default を慣例の解決先とし専用 layat 名前空間で packages を汚さない](../requirements/default-config-name-and-namespace.md) に置く。

## 根拠

名前空間を分ければ、manifest が通常のパッケージとして誤って build されず、`packages` を利用者自身の
パッケージのために残せる。名前で区別するので 1 つの entrypoint に複数の config を置け、役割ごとに
独立した profile と世代を持てる。system の次元は CLI が補うので、利用者は名前だけを指定すればよい。

## 棄却した代替案

### `packages.<system>.<name>` を流用する

system は自動で補われるが、manifest が通常のパッケージとして一覧や build の対象に並び、誤って build
されうる。

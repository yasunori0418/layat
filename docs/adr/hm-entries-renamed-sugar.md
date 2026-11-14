---
id: "ADR-3447db4a-cdec-4573-bbad-05bbffc34a65"
type: adr
name: "layat.entries を configs.default.entries への rename 糖衣として残す"
section: modules
status: stable
sources:
  - id: "issue-585"
    resource: "https://github.com/yasunori0418/layat/issues/585"
    title: "#585"
origin: "ADR-0035 §2 を書き直した"
justifies:
  - "REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:27:41+09:00"
---
# layat.entries を configs.default.entries への rename 糖衣として残す

- ステータス: stable
- 日付: 2026-11-03
- 関連: GitHub Issue #585, `docs/spec.md`
- 起点: ADR-0035 §2 を書き直した

## 背景

HM モジュールは `layat.configs.<name>` で config ごとに entries を受け取る。それ以前は
`layat.entries` で受け取り、profile は `default` に固定していた。`configs` を正とした後も、
`layat.entries` をどう扱うかを決める必要がある。

## 決定

`layat.entries` を `lib.mkRenamedOptionModule` で `configs.default.entries` への rename 糖衣として
残し、deprecated とする。docs は `configs.<name>` を正として書き、削除の時期は定めない。規範は
[全モジュールは共通オプションの同一集合を公開し、entries は configs 経由・root はモジュールが pin する](../requirements/module-common-options.md) に置く。

## 根拠

リポジトリは公開済みで、利用者がいないとは言い切れない。profile を 1 つしか使わない利用者には
`layat.entries` が今も最短の書き方である。`mkRenamedOptionModule` の 1 行で既存の設定を変えずに使い続けられ、
home-manager 標準の rename 警告がそのまま deprecated の告知になる。新しい利用者はこの警告で
`configs` へ誘導される。

## 棄却した代替案

### `layat.entries` を廃止して `configs.<name>` に一本化する

公開済みの API の互換を失う。糖衣 1 行で避けられるコストに見合わない。

### 警告を出さずに両方を正式な書き方として並べる

同じ設定に 2 通りの書き方が残り続け、docs でどちらを正とするかが定まらない。

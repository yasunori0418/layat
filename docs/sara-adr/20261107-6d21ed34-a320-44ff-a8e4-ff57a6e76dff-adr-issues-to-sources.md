---
id: "SDR-6d21ed34-a320-44ff-a8e4-ff57a6e76dff"
type: sara_adr
name: "ADR の issues を sources に移す"
description: "ADR の外部参照の Issue を OKF の sources リストへ移し、origin は拡張キーとして残す"
status: stable
issues:
  - "#662"
origin: "epic #662 の grilling（2026-11-07）で確定"
justifies:
  - "QA-030f926b-5ae7-4543-b8cb-e66aae5e2b5e"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-07T22:51:53+09:00"
---
# SDR-6d21ed34-a320-44ff-a8e4-ff57a6e76dff: ADR の issues を sources に移す

- ステータス: stable
- 日付: 2026-11-07
- 関連: GitHub Issue #662, `docs/adr/README.md`
- 起点: epic #662 の grilling（2026-11-07）で確定

## 背景

ADR / SDR は、item にできない外部参照を `issues`（`#N` の文字列のリスト）と `origin`（散文）に
持つ。`#N` はこのリポジトリの Issue だと分かっている人にしか読めず、他のリポジトリへ切り出すと
指す先が変わる。

OKF には、item の出典を `id` / `resource` / `title` の組のリストで持つ `sources` キーがある。
SDR-bab72f31-7cb1-47a3-bb2d-7876757d86fb で OKF と互換のキーを足すと決めた。

## 決定

ADR / SDR の `issues` を `sources` のリストへ移す。`origin` は散文のまま、layat の拡張キーとして残す。

- `resource` には Issue の URL を書く。`#N` は `https://github.com/yasunori0418/layat/issues/N`、
  `owner/repo#N` は `https://github.com/owner/repo/issues/N` にする
- `id` は `issue-N`（他のリポジトリなら `issue-<repo>-N`）の形にし、1 つの item の中で一意にする
- `title` には Issue のタイトルを書く
- `sources` の形（`resource` が空でないこと・`id` の一意性）は `dev/tests/okf-frontmatter.sh` が
  確かめる

## 根拠

URL なら、リポジトリの外から読んでも指す先が一意に決まる。`id` を持たせると、本文から特定の出典を
名指しできる。`title` があれば、URL を開かなくても出典の中身の見当がつく。

`origin` は「どのセッションで何を書き直したか」という散文で、`id` / `resource` / `title` の組に
分けられない。無理に `sources` へ入れると `resource` の無い要素ができるので、拡張キーとして残す。

## 棄却した代替案

### `issues` を残して `sources` を併記する

同じ Issue を 2 つのキーに書くことになり、片方だけを直す食い違いが起きる。

### `origin` も `sources` に入れる

`origin` には URL の無いもの（grilling のセッション・旧 ADR の番号）が多く、`resource` が空の
要素になる。

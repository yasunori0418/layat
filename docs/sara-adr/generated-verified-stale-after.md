---
id: "SDR-9a646576-31bd-4cea-996d-54993ff2b8fa"
type: sara_adr
name: "item に生成者・確認の記録・期限を持たせる"
description: "生成者と時刻の generated を必須にし、確認の記録 verified と期限 stale_after を任意で持たせ、入れ子キーの形を契約テストで確かめる"
status: stable
sources:
  - id: "issue-662"
    resource: "https://github.com/yasunori0418/layat/issues/662"
    title: "#662"
origin: "epic #662 の grilling（2026-11-07）で確定"
justifies:
  - "INF-659b139d-0cf8-4c65-b30d-93c5ee2dfc71"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-07T23:10:46+09:00"
---
# SDR-9a646576-31bd-4cea-996d-54993ff2b8fa: item に生成者・確認の記録・期限を持たせる

- ステータス: stable
- 日付: 2026-11-07
- 関連: GitHub Issue #662, `docs/model.md`, `dev/tests/okf-frontmatter.sh`
- 起点: epic #662 の grilling（2026-11-07）で確定

## 背景

item を誰がいつ書いたか、人が内容を確かめたか、いつまで有効な事実かは、frontmatter から読めない。
git の履歴には書いた時刻が残るが、frontmatter だけを直したコミットや改名のコミットも同じ履歴に並ぶので、
本文を最後に変えた時刻を取り出すには手間がかかる。

OKF には、生成者と時刻の `generated`、確認の記録の `verified`、期限の `stale_after` がある。
SDR-bab72f31-7cb1-47a3-bb2d-7876757d86fb で OKF と互換のキーを足すと決めた。

## 決定

`generated` を全ての item で必須にし、`verified` と `stale_after` は任意で持たせる。

- `generated` は `by`（actor）と `at`（時刻）の組にする。actor は人なら `human:<GitHub ID>`、
  エージェントなら `claude-code/<モデル ID>` の形で書く。`at` は UTC オフセット付きの ISO 8601 で書く
- `generated.at` は、本文を最後に意味のある形で変えた時刻にする。git の履歴と比べられる値に
  するため、frontmatter だけの変更や改名では動かさない。本文を変えたら `generated` を更新する
- 起票では `sara-new` が `generated` を付ける。actor は環境変数 `SARA_NEW_GENERATED_BY` から取る
- `verified` は `generated` と同じ形の要素のリストにする。自動では付けず、人が内容を確かめたときに
  だけ手で足す
- `stale_after` は UTC オフセット付きの ISO 8601 の時刻にする。期限のある事実（外部ツールの版や
  公開状況など）を含む item にだけ、書き手が付ける
- 入れ子のキー（`generated` / `verified` / `sources`）の形は `dev/tests/okf-frontmatter.sh` が
  確かめる。平たいキーは `docs/model.yaml` のフィールドにして sara に確かめさせる
- `generated.at` と git の履歴が合っているかは、この決定では確かめない

## 根拠

`generated` が全ての item にあれば、どの item が古いかを frontmatter だけで比べられる。actor を
人とエージェントで書き分けると、エージェントが書いたまま人が確かめていない item を絞り込める。
`at` を本文の変更に結び付けるのは、frontmatter の機械的な書き換えで時刻が一斉に動くと、鮮度の
比較に使えなくなるためである。

`verified` を自動で付けると、確かめていない item にも確認の記録が付き、記録の意味が無くなる。
`stale_after` を全ての item に求めると、期限の無い規範に架空の期限を書くことになる。

sara は入れ子のキーに型を持たないので、形を確かめる処理は契約テストに置くしかない。本文の変更と
`generated` の更新が揃っているかは、規約で守る。

## 棄却した代替案

### `generated` を git の履歴から毎回導く

frontmatter だけを読む OKF の読み手には届かない。履歴から導く処理は、埋め戻しと、後続の epic で履歴と比べる
仕組みに使う。

### `verified` をレビューの承認から自動で付ける

レビューは差分を見るもので、item の内容全体を確かめたことにはならない。

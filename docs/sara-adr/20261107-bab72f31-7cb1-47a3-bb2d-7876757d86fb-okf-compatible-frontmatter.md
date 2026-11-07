---
id: "SDR-bab72f31-7cb1-47a3-bb2d-7876757d86fb"
type: sara_adr
name: "OKF v0.2 と互換の frontmatter キーを item に足す"
description: "OKF v0.2 と互換の frontmatter キーを sara の item に足す方針と、バンドル単位の適合を追わない範囲を決める"
status: 採用
issues:
  - "#662"
origin: "epic #662 の grilling（2026-11-07）で確定"
justifies:
  - "QA-67715bb3-1162-4ccf-8441-2f44257a57da"
  - "INF-659b139d-0cf8-4c65-b30d-93c5ee2dfc71"
---
# SDR-bab72f31-7cb1-47a3-bb2d-7876757d86fb: OKF v0.2 と互換の frontmatter キーを item に足す

- ステータス: 採用
- 日付: 2026-11-07
- 関連: GitHub Issue #662, `docs/model.md`
- 起点: epic #662 の grilling（2026-11-07）で確定

## 背景

item のメタ情報は sara の型とフィールドだけで表してきた。生成者・確認の記録・鮮度・来歴の情報は
item 自体に無く、git の履歴や Issue をたどらないと分からない。読み手は第一に item を書く人と
エージェント自身で、第二に sara を持たない別のエージェントやリポジトリになる（#500 で切り出す
基盤を使う側）。後者は sara のモデルを読まないので、広く知られた形式のキーで書かれているほうが読める。

Open Knowledge Format（OKF）は、Markdown の知識ベースの frontmatter キーと、バンドル単位の構成を
定める形式である。v0.2 は本家で退役が進んでおり、v0.3 は本文がまだ公開されていない。

## 決定

OKF v0.2 と互換の frontmatter キーを、sara の item に足す。対応はキー単位で取り、バンドル単位の
適合は追わない。

- OKF の `title` には sara の組み込みの `name` を当てる。`title` は足さない
- OKF の必須キー `type` は sara の組み込みフィールドで満たす。欠落は `sara check` が拒否するので、
  確かめる仕組みを新たに足さない
- `description` は新しく起こす item で必須にする。既存の item には手で埋め戻さない
- `tags` は使わない
- バンドルの宣言（`okf_version`）・`log.md`・frontmatter の無い文書への frontmatter 追加は行わない
- v0.3 の規則は、本文が公開されてから採否を決め直す
- 個々のキーの意味と書き方は `docs/model.md` の「OKF 互換の frontmatter」節に置く

## 根拠

sara は未知のキーを無視し、OKF が常に求めるキーは `type` だけである。キー単位なら、sara の検証を
保ったまま OKF の読み手にも読める frontmatter にできる。

バンドル単位の適合には、frontmatter の無い概要文書や運用文書まで OKF の様式に合わせる作業が要る。
退役が進む版にそこまで合わせても、v0.3 で規則が変われば作業が無駄になる。キー単位の互換なら、
v0.3 への追随もキーの足し引きで済む。

`name` を `title` として扱えば、同じ値を 2 つのキーに書かずに済む。`tags` は分類の語彙を別に
管理する必要があり、分類はすでに型と relation が担っている。既存の item の `description` は、
index.md の生成器が本文の最初の段落で代用できる。

## 棄却した代替案

### OKF のバンドルとして完全に適合させる

`okf_version` と `log.md` を置き、全ての Markdown に frontmatter を付けることになる。退役が進む
v0.2 のために frontmatter の無い 12 文書まで書き換える費用が、得られるものに見合わない。

### sara 独自のキー名で同じ情報を持つ

sara のモデルを読まない読み手にキーの意味が伝わらない。OKF と同じ名前にしておけば、読み手は
OKF の定義で解釈できる。

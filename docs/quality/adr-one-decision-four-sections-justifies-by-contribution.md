---
id: "QA-6de13779-f0bc-4e47-8ffd-3d58de265762"
type: quality
name: "ADR は 1 本に 1 つの決定を 4 節で書き、justifies は規範の形成に寄与した item へ張る"
description: "ADR は単独で覆せる決定ごとに 1 本を 4 節で書き、justifies は規範の形成に寄与した item へ張る"
status: stable
derives_from:
  - "SOL-9fcd1d6e-6204-42e6-92bb-1faf966f0b3e"
specification: |
  ADRs SHALL be split so that each one holds a single decision that can be overturned on
  its own. The body of an ADR SHALL consist of four sections: background, decision,
  rationale and rejected alternatives. When ADRs are merged or split, no decision that is
  still in effect, other than a completed one-off process decision, SHALL be omitted, and
  each such decision SHALL be raised as one ADR.
  The `justifies` of an ADR SHALL point to one or more items, each of whose norm the ADR
  substantively contributed to forming.
specification_ja: |
  ADR は、単独で覆せる決定ごとに分けなければならない。ADR の本文は、背景・決定・根拠・
  棄却した代替案の 4 節としなければならない。ADR の統合と分割では、完了した一回限りの
  プロセス決定を除き、有効な決定を省かず 1 決定 1 本で起こさなければならない。ADR の
  `justifies` は、その ADR が規範の形成に実質的に寄与した item を 1 本以上指さなければ
  ならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-10T23:55:19+09:00"
---
# ADR は 1 本に 1 つの決定を 4 節で書き、justifies は規範の形成に寄与した item へ張る

## 仕様

### 粒度: 1 本に 1 つの決定

1 本の ADR には**単独で覆せる 1 つの決定**だけを書く。

判定の問いは「この決定だけを覆したとき、同じ ADR の他の決定はそのまま残るか」。残るなら別の
決定なので、ADR を分ける。逆に、片方を覆すともう片方が意味を失う決定どうしは 1 つの決定として
1 本にまとめる。

- タイトルは 1 つの動詞句にする。「／」「と」「・」で主題を並べたタイトルは、複数の決定を
  抱えている兆候になる
- `justifies` の本数が多いこと（目安 8 本以上）は分割を考えるきっかけにする。規則ではないので、
  1 つの決定が多くの item を決めているなら分けない

### 本文の構成

本文は背景・決定・根拠・棄却した代替案の 4 節を持つ。規範の細目（入力・出力・エラー・境界
条件など）は ADR に書き下さず、要求や設計の item に書いて `justifies` で指す。ADR は「なぜその
規範にしたか」を残す場所で、規範そのものを定める場所ではない。

検討して棄却した代替案は、その ADR の「棄却した代替案」節に残す。棄却だけを独立した ADR に
するのは、棄却に独自の起点（提案の Issue や議論）があるときに限る。

### 統合と分割の手順

複数の決定を抱えた ADR や、改訂を重ねた ADR を書き直すときは、次の手順をとる。

1. 古い ADR の決定を棚卸しする。決定ごとに、いまも有効か・後の ADR が改訂したか・撤回されたかを
   分類する
2. 撤回された決定には新しい ADR を起こさない。完了した一回限りのプロセス決定も同じ。それ以外の
   有効な決定は、価値が低く見えても省略せず 1 決定 1 本で起こす
3. 改訂された決定は改訂後の内容を現時点の決定として書く。改訂の経緯は書かず、背景と根拠だけを書く
4. 新しい ADR の `justifies` は古い ADR の `justifies` から当該決定に対応する item だけを引き継ぐ
5. 古い ADR の全決定に後継が揃ったら、`supersedes` を張って古い ADR を `deprecated` にする

### `justifies` の選定基準

`justifies` は必須で、ADR が決めた item を 1 本以上張る。空の `justifies` を埋め忘れとして
機械判定できる状態を保つため、空のまま出さない。張る先の item がまだ無ければ、ADR と同じ変更の中で
item を起こしてから張る。決定を item に落とせない ADR は、決定が定まっていないか粒度が大きすぎる
ので、先にそちらを解く。

どの item を張るかの基準は「その ADR が、その item の規範の形成に実質的に寄与したか」である。決定の
実体である ADR 1 本だけが張るのではなく、寄与した ADR が複数あれば全てが張ってよい。

次の 2 つは寄与が成立しないので張らない。

1. 接続先の item 自身が `justifies` を受けない理由を本文に明記している。ただし「その規範を
   この item には含めない」という宣言と「その ADR に依拠していない」ことは別なので、文面を読んで
   判定する
2. ADR がその規範を決めていない。背景で既存の前提として言及するだけの ADR は張らない

### 網羅性の監査

張り先の候補は 2 方向から挙げる。

- ADR の決定節を上から順に読み、規範ごとに張り先を探す
- item の出典節が「〜はこの ADR が定めている」と名指ししている ADR を挙げる。名指しの無い item も
  あるので、こちらだけでは挙げ切れない

挙げた候補は**ADR の決定節に、その item の規範そのものを決めた決定文があるか**で絞る。出典節の
名指しは周辺規範の淵源も含むので、名指しがあることは寄与の証明にならない。次の 2 類型は原則として
張らない。

- item の本文がその規範を別の item へ委譲しているもの。委譲先に同じ ADR が張られていれば寄与は
  回収済み。ただし item の規範文そのものが、その ADR の決めた規範を対称化・拡張の対象として直に
  述べているときは張る。他の item を経由した参照はこれに当たらない。判定材料は本文の注記ではなく
  規範文に置く
- 「淵源」「姿勢」といった態度の供給元としてだけ挙がるもの。語で判定せず、決定節に対応する
  決定文があるかで見る

網羅性はグラフの検証では判定できない。検証が見るのは「1 本以上張ってあるか」だけで、「決めた規範を
全て張っているか」は見ない。新しい ADR を書いたら、決定節を上から順に見て規範ごとに張り先を確かめる。

## 出典

[ADR は単独で覆せる 1 つの決定だけを書く](../sara-adr/adr-one-decision-granularity.md) が
Issue #499 で定めた粒度に、本文の構成・統合と分割の手順・`justifies` の選定基準を合わせて
ADR の規則として運用してきたものを、Issue #649 で quality item として立てた。

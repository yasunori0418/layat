---
id: "QA-d9cf5fcb-6363-49eb-9511-48b233c0c978"
type: quality
name: "risk はまず requirement に張り、設計を差し替えると消える懸念だけを design に張る"
description: "risk はまず requirement に張り、設計を差し替えると消える懸念だけを design に張る"
status: stable
derives_from:
  - "SOL-9fcd1d6e-6204-42e6-92bb-1faf966f0b3e"
specification: |
  A `risk` SHALL attach to a `requirement` by default. It SHALL attach to a `design` only
  when it would disappear if that design were replaced by an alternative satisfying the
  same requirement, and a concern that the requirement itself might not be met SHALL
  always attach to the `requirement`. This judgement SHALL be made once per `threatens`
  edge, not once per risk. A `requirement` that holds no requirement-level risk SHALL
  state why in `risk_exempt`, in one sentence.
specification_ja: |
  `risk` は既定で `requirement` に張らなければならない。`design` に張ってよいのは、同じ
  `requirement` を満たす別案へその設計を置き換えたら消える risk である場合に限らなければならず、
  `requirement` そのものが満たされないかもしれないという懸念は常に `requirement` に張らなければ
  ならない。この判定は risk ごとではなく `threatens` の 1 本ごとに行わなければならない。
  `requirement` 側の risk を 1 つも持たない `requirement` は、その理由を `risk_exempt` に 1 文で
  書かなければならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-11T00:57:16+09:00"
---
# risk はまず requirement に張り、設計を差し替えると消える懸念だけを design に張る

## 仕様

`risk` は `threatens` で `requirement` と `design` のどちらも指せる。モデルはその両方を許し、
本数も組み合わせも縛らないので、どちらに張るかは判断であり、次の規則で決める。

> risk は既定で `requirement` に張る。`design` に張るのは、**その設計を別案に置き換えたら
> 消える** risk、つまりその設計を選んだこと自体に伴う risk に限る。`requirement` そのものが
> 満たされないかもしれないという懸念は、常に `requirement` に張る。

判定の問い: 同じ `requirement` を満たす別の設計へ差し替えたと想像する。risk が一緒に消えるなら
design に属する。差し替えても残るなら `requirement` に属する。

この問いは risk ごとに 1 回ではなく、`threatens` の 1 本ごとに当てる。1 つの脅威の複数の面を
挙げる risk は、両方の種類の辺を同時に持ってよい。差し替えても残る面は `requirement` へ、設計に
固有の面は `design` へ張る。下の design 側の例がまさにこの形である。規則が禁じるのは、問いが
支持しない種類の辺であり、複数の辺を得た risk ではない。

例は次のとおり。

| risk | 張り先 | 理由 |
|---|---|---|
| [entrypoint の形ごとに root 解決と配置の意味論が食い違う](../risks/entrypoint-root-resolution.md)（→ [legacy entrypoint の分岐は attr path 組み立てに閉じ、nix 呼び出しヘルパを共通で再利用する](../design/legacy-nix-invocation-reuse.md)） | `design` | legacy の分岐を attr path の組み立てに閉じるのは、「entrypoint の形をまたいで実装を 1 つにする」ための設計の 1 つである。これを置き換えれば（たとえば entrypoint ごとに nix の呼び出しを別々に書けば）、この食い違いは消える。 |
| `apply` が失敗したときにファイルシステムへ痕跡を残す | `requirement` | `requirement` そのものが破られることである。どの設計もこれに答えなければならず、設計を変えても懸念は消えない。 |

設計に固有の risk を `requirement` に張ると、それが特定の選択に伴う代償だという情報が失われる。
`requirement` 側の risk を 1 つの design に張ると、同じ懸念を抱える他の設計からそれが見えなくなる。

`requirement` 側の risk を 1 つも持たない `requirement` は、その理由を `risk_exempt` に 1 文で書く。

### 張り先が重要な理由

test_condition は risk からぶら下がる（`test_condition --mitigates--> risk`）。risk を置き違えると、
それを守るテストも置き違える。`requirement` 側の risk を 1 つの design の下に置くと、同じ懸念を
共有する他の設計は、その懸念についてテストされていないように見える。

## 出典

epic #203 が定めた使い分けを規約として運用してきたものに、[requirement に risk_exempt を設け、理由を書いた requirement を未カバーから外す](../sara-adr/requirement-risk-exempt.md) が置いた
`risk_exempt` の書き方を合わせ、Issue #649 で quality item として立てた。
この規約を定めた決定は [risk の張り先は requirement を既定にし、threatens の 1 本ごとに判定する](../sara-adr/risk-attaches-to-requirement-by-default.md) に記録してある。

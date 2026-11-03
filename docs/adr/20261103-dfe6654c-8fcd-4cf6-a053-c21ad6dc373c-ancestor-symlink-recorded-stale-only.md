---
id: "ADR-dfe6654c-8fcd-4cf6-a053-c21ad6dc373c"
type: adr
name: "祖先 symlink は自分が記録して古くなったものだけを移行する"
status: 採用
issues:
  - "#170"
origin: "dotfiles 側での実遭遇（#170）と、次期マイルストーンの grilling セッション（2026-07-11）で設計確定。ADR-0046 §1 を書き直した"
justifies:
  - "REQ-c9ab91c1-f778-4f87-a2ea-c66d6b3c2575"
supersedes:
  - "ADR-0046"
references:
  - "ADR-88231ffc-4836-4252-be34-894944fd28c3"
  - "ADR-095220c9-5a3d-4fd4-94bd-8fa1a02dd36d"
  - "ADR-9755ba76-4c02-4b4e-a8c0-031a24a3a5b9"
  - "ADR-62540f8f-a8ba-4441-97b0-36b50d7862d8"
  - "ADR-1cbacca1-1930-46ca-bf82-30310f82216e"
  - "ADR-4c8d459a-c371-4f0d-b009-16e2ca096e49"
  - "ADR-eb1aa25f-a0ca-4efe-bcd9-4cbe11368764"
  - "ADR-ffe1afcb-0b51-4b87-b698-0a6dea307bdd"
  - "ADR-8b8c9109-a479-4571-8b10-5a6079127441"
  - "ADR-1908099c-87f7-4df1-bda5-7731df7d9ed7"
  - "ADR-efa7685c-6fa4-45df-b7a3-8f3f58e4118c"
  - "ADR-764ffff1-6b5a-4ed7-89d0-5700eb3ee883"
  - "ADR-9723dbb3-6659-4a69-ac4f-411c98bfa818"
  - "ADR-4f08aa18-6a38-49fd-a784-450677d80e1a"
  - "ADR-d97154cd-a2e5-4853-a682-800ff9435c89"
  - "ADR-3e283fe2-0177-4caf-8b60-577c5ee3b10a"
  - "ADR-b06dd8ab-317e-41ee-9c80-dd70dcd7f2fd"
  - "ADR-40dd7945-5792-4b80-bdfa-43f95f482554"
  - "ADR-9f2cc068-8edb-4e3d-958b-0dff8f6ed90c"
---
# ADR-dfe6654c-8fcd-4cf6-a053-c21ad6dc373c: 祖先 symlink は自分が記録して古くなったものだけを移行する

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-88231ffc-4836-4252-be34-894944fd28c3, ADR-095220c9-5a3d-4fd4-94bd-8fa1a02dd36d, ADR-9755ba76-4c02-4b4e-a8c0-031a24a3a5b9, ADR-62540f8f-a8ba-4441-97b0-36b50d7862d8, ADR-1cbacca1-1930-46ca-bf82-30310f82216e, ADR-4c8d459a-c371-4f0d-b009-16e2ca096e49, ADR-eb1aa25f-a0ca-4efe-bcd9-4cbe11368764, ADR-ffe1afcb-0b51-4b87-b698-0a6dea307bdd, ADR-8b8c9109-a479-4571-8b10-5a6079127441, ADR-1908099c-87f7-4df1-bda5-7731df7d9ed7, ADR-efa7685c-6fa4-45df-b7a3-8f3f58e4118c, ADR-764ffff1-6b5a-4ed7-89d0-5700eb3ee883, ADR-9723dbb3-6659-4a69-ac4f-411c98bfa818, ADR-4f08aa18-6a38-49fd-a784-450677d80e1a, ADR-d97154cd-a2e5-4853-a682-800ff9435c89, ADR-3e283fe2-0177-4caf-8b60-577c5ee3b10a, ADR-b06dd8ab-317e-41ee-9c80-dd70dcd7f2fd, ADR-40dd7945-5792-4b80-bdfa-43f95f482554, ADR-9f2cc068-8edb-4e3d-958b-0dff8f6ed90c, GitHub Issue #170
- 起点: dotfiles 側での実遭遇（#170）と、次期マイルストーンの grilling セッション（2026-07-11）で設計確定。ADR-0046 §1 を書き直した

## 背景

`.claude/skills` のようにディレクトリ全体を 1 つの symlink で置いていた entry を、配下に子の
entry を並べる形（`.claude/skills/foo` など）へ移そうとすると、apply が丸ごと失敗していた。
target の祖先が symlink なら、それが自分の前世代の置いた symlink でも、一律に conflict に
していたからである。conflict で止まるので旧世代の symlink も除去されず、新旧どちらの状態にも
収束しない。移行のたびに、apply の前に手で `rm` する必要があった。

祖先の symlink で止めるのは、store を指す symlink の配下へ子を置くと read-only の store に
書こうとしたり dangling を作ったりする汚染を防ぐためである（ADR-0015）。この安全策は、
素性の分からない祖先に対しては保ちたい。

## 決定

祖先の symlink の conflict を外して移行へ回すのは、その祖先が記録済みで、しかも古くなって
いるときに限る。記録済みとは、前世代の manifest がその祖先を entry として記録し、on-disk の
symlink が記録の dest と一致すること。古くなっているとは、その祖先が次世代の manifest に
無いこと。foreign な祖先、
前世代が無いとき、次世代にも祖先の entry が残る自己矛盾の manifest は、これまでどおり error で
止める。

判定の規範は REQ-c9ab91c1-f778-4f87-a2ea-c66d6b3c2575 に置く。

## 根拠

記録済みと古くなっていることの両方を求めれば、自分で置いて自分で消すという既知で安全な場合だけを最小の
条件で切り出せる。記録済みの判定は、古くなった配置の除去が記録どおりの symlink だけを消すときの不変条件
（ADR-9f2cc068-8edb-4e3d-958b-0dff8f6ed90c）と同じである。

記録済みだけを条件に緩めると、次世代にも祖先が残る自己矛盾の manifest まで通してしまう。親を消せない
のに子を置くことになる。古くなっていることだけを条件に緩めると foreign な祖先まで通し、store の汚染が戻る。
既知のものだけ移行し、未知のものでは安全に止まる、という非対称を保つ。

前世代の manifest は state ディレクトリに永続しているので、この判定は mode によらず成り立つ。
前世代が無いとき（初回の apply、config の改名など）は foreign 扱いになり、安全側で止まる。

## 棄却した代替案

### apply の前に手で `rm` する運用を続ける

移行のたびに手作業が要り、宣言的に管理するという layat の考え方に合わない。

### 祖先の symlink を見つけたら warning を出して置き換える

foreign な symlink の配下への誤ったネスト（store の汚染）まで通してしまう。

### 記録済みだけを条件にする

次世代にも祖先が残る自己矛盾の manifest まで緩め、親を消せないのに子を置く矛盾を生む。

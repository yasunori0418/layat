---
id: "ADR-dfe6654c-8fcd-4cf6-a053-c21ad6dc373c"
type: adr
name: "祖先 symlink は自己記録の stale だけを移行する"
status: 採用
issues:
  - "#170"
origin: "dotfiles 側での実遭遇（#170）と、次期マイルストーンの grilling セッション（2026-07-11）で設計確定。ADR-0046 §1 を書き直した"
justifies:
  - "REQ-c9ab91c1-f778-4f87-a2ea-c66d6b3c2575"
references:
  - "ADR-0015"
  - "ADR-9f2cc068-8edb-4e3d-958b-0dff8f6ed90c"
---
# ADR-dfe6654c-8fcd-4cf6-a053-c21ad6dc373c: 祖先 symlink は自己記録の stale だけを移行する

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-0015, ADR-9f2cc068-8edb-4e3d-958b-0dff8f6ed90c, GitHub Issue #170
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

祖先の symlink の conflict を外して移行へ回すのは、その祖先が recorded かつ stale のときに
限る。recorded は、前世代の manifest がその祖先を entry として記録し、on-disk の symlink が
記録の dest と一致すること。stale は、その祖先が次世代の manifest に無いこと。foreign な祖先、
前世代が無いとき、次世代にも祖先の entry が残る自己矛盾の manifest は、これまでどおり error で
止める。

判定の規範は REQ-c9ab91c1-f778-4f87-a2ea-c66d6b3c2575 に置く。

## 根拠

recorded と stale の両方を求めれば、自分で置いて自分で消すという既知で安全な場合だけを最小の
条件で切り出せる。recorded の判定は、stale 除去が記録どおりの symlink だけを消すときの不変条件
（ADR-9f2cc068-8edb-4e3d-958b-0dff8f6ed90c）と同じである。

recorded だけで緩めると、次世代にも祖先が残る自己矛盾の manifest まで通してしまう。親を消せない
のに子を置くことになる。stale だけで緩めると foreign な祖先まで通し、store の汚染が戻る。
既知のものだけ移行し、未知のものでは安全に止まる、という非対称を保つ。

前世代の manifest は state ディレクトリに永続しているので、この判定は mode によらず成り立つ。
前世代が無いとき（初回の apply、config の改名など）は foreign 扱いになり、安全側で止まる。

## 棄却した代替案

### apply の前に手で `rm` する運用を続ける

移行のたびに手作業が要り、宣言的に管理するという layat の考え方に合わない。

### 祖先の symlink を見つけたら warning を出して置き換える

foreign な symlink の配下への誤ったネスト（store の汚染）まで通してしまう。

### recorded だけを条件にする

次世代にも祖先が残る自己矛盾の manifest まで緩め、親を消せないのに子を置く矛盾を生む。

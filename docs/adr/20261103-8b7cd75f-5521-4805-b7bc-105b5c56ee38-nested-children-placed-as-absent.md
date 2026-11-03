---
id: "ADR-8b7cd75f-5521-4805-b7bc-105b5c56ee38"
type: adr
name: "移行する祖先の配下の entry は lstat を経ず不在として配置する"
status: 採用
issues:
  - "#170"
origin: "dotfiles 側での実遭遇（#170）と、次期マイルストーンの grilling セッション（2026-07-11）で設計確定。ADR-0046 §2 を書き直した"
justifies:
  - "REQ-c9ab91c1-f778-4f87-a2ea-c66d6b3c2575"
references:
  - "ADR-0015"
  - "ADR-dfe6654c-8fcd-4cf6-a053-c21ad6dc373c"
---
# ADR-8b7cd75f-5521-4805-b7bc-105b5c56ee38: 移行する祖先の配下の entry は lstat を経ず不在として配置する

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-0015, ADR-dfe6654c-8fcd-4cf6-a053-c21ad6dc373c, GitHub Issue #170
- 起点: dotfiles 側での実遭遇（#170）と、次期マイルストーンの grilling セッション（2026-07-11）で設計確定。ADR-0046 §2 を書き直した

## 背景

祖先の symlink が自己記録の stale なら、配置の前に除去して配下へ子を置く
（ADR-dfe6654c-8fcd-4cf6-a053-c21ad6dc373c）。plan を計算する時点では、祖先の symlink はまだ
FS に残っている。この時点で配下の子をどう分類するかを決める必要があった。

## 決定

祖先が移行の対象のとき、その配下の entry は lstat の結果を使わず、target が不在であるものと
して新しく配置する（symlink は新規配置、copy は place-once の新規コピー）。

分類の規範は REQ-c9ab91c1-f778-4f87-a2ea-c66d6b3c2575 に置く。

## 根拠

OS の lstat は途中の component をたどる。移行する祖先の symlink を経由すると、子の target は
前世代の farm の中身（store の中）を指し、存在するように見える。通常どおり分類すると、
foreign の張替えや構造の不一致と誤って判定し、store の中へ書き込もうとする。これは祖先の
symlink で止めていた理由（ADR-0015）の汚染そのものである。

祖先は配置の前に除去するので、子は不在として新しく置いてよい。祖先を除去する前の lstat は
信用できない、という事実に対する対処で、汚染を防ぐ安全策を保ったまま移行を成り立たせる。

## 棄却した代替案

### 子も通常どおり lstat の結果で分類する

祖先の symlink をたどった結果で分類し、store の中を既存の実体と誤認する。張替えや conflict と
判定して store へ書こうとするか、移行が止まる。

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
  - "ADR-3b8175b9-5ebf-4123-8855-f698dc843d6b"
  - "ADR-2d4fdda4-b2e1-4161-8b71-3cff5b265f2f"
  - "ADR-8357b519-cb23-4984-ba56-254046779f47"
  - "ADR-4f08aa18-6a38-49fd-a784-450677d80e1a"
  - "ADR-d97154cd-a2e5-4853-a682-800ff9435c89"
  - "ADR-3e283fe2-0177-4caf-8b60-577c5ee3b10a"
  - "ADR-b06dd8ab-317e-41ee-9c80-dd70dcd7f2fd"
  - "ADR-40dd7945-5792-4b80-bdfa-43f95f482554"
  - "ADR-dfe6654c-8fcd-4cf6-a053-c21ad6dc373c"
---
# ADR-8b7cd75f-5521-4805-b7bc-105b5c56ee38: 移行する祖先の配下の entry は lstat を経ず不在として配置する

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-88231ffc-4836-4252-be34-894944fd28c3, ADR-095220c9-5a3d-4fd4-94bd-8fa1a02dd36d, ADR-9755ba76-4c02-4b4e-a8c0-031a24a3a5b9, ADR-62540f8f-a8ba-4441-97b0-36b50d7862d8, ADR-1cbacca1-1930-46ca-bf82-30310f82216e, ADR-4c8d459a-c371-4f0d-b009-16e2ca096e49, ADR-eb1aa25f-a0ca-4efe-bcd9-4cbe11368764, ADR-ffe1afcb-0b51-4b87-b698-0a6dea307bdd, ADR-8b8c9109-a479-4571-8b10-5a6079127441, ADR-1908099c-87f7-4df1-bda5-7731df7d9ed7, ADR-efa7685c-6fa4-45df-b7a3-8f3f58e4118c, ADR-764ffff1-6b5a-4ed7-89d0-5700eb3ee883, ADR-9723dbb3-6659-4a69-ac4f-411c98bfa818, ADR-3b8175b9-5ebf-4123-8855-f698dc843d6b, ADR-2d4fdda4-b2e1-4161-8b71-3cff5b265f2f, ADR-8357b519-cb23-4984-ba56-254046779f47, ADR-4f08aa18-6a38-49fd-a784-450677d80e1a, ADR-d97154cd-a2e5-4853-a682-800ff9435c89, ADR-3e283fe2-0177-4caf-8b60-577c5ee3b10a, ADR-b06dd8ab-317e-41ee-9c80-dd70dcd7f2fd, ADR-40dd7945-5792-4b80-bdfa-43f95f482554, ADR-dfe6654c-8fcd-4cf6-a053-c21ad6dc373c, GitHub Issue #170
- 起点: dotfiles 側での実遭遇（#170）と、次期マイルストーンの grilling セッション（2026-07-11）で設計確定。ADR-0046 §2 を書き直した

## 背景

祖先の symlink が、自己記録の古くなったものなら、配置の前に除去して配下へ子を置く
（ADR-dfe6654c-8fcd-4cf6-a053-c21ad6dc373c）。plan を計算する時点では祖先の symlink がまだ
FS に残っているので、この時点で配下の子をどう分類するかを決める必要があった。

## 決定

祖先が移行の対象のとき、その配下の entry は lstat の結果を使わず、target が不在であるものと
して新しく配置する（symlink は新規配置、copy は place-once の新規コピー）。

分類の規範は REQ-c9ab91c1-f778-4f87-a2ea-c66d6b3c2575 に置く。

## 根拠

OS の lstat は途中の component をたどるので、移行する祖先の symlink を経由した子の target は、
前世代の farm の中身（store の中）を指して存在するように見える。通常どおり分類すると、
foreign の張替えや構造の不一致と誤って判定し、store の中へ書き込もうとする。これは祖先の
symlink で止めていた理由（ADR-dfe6654c-8fcd-4cf6-a053-c21ad6dc373c）の汚染そのものである。

祖先は配置の前に除去するので、子は不在として新しく置いてよい。祖先を除去する前の lstat は
信用できない。この事実に合わせて分類を変えれば、汚染を防ぐ安全策を保ったまま移行を成り立たせ
られる。

## 棄却した代替案

### 子も通常どおり lstat の結果で分類する

祖先の symlink をたどった結果で分類し、store の中を既存の実体と誤認する。張替えや conflict と
判定して store へ書こうとするか、移行が止まる。

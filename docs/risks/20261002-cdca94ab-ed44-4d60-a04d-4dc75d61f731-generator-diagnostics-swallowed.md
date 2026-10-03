---
id: "RISK-cdca94ab-ed44-4d60-a04d-4dc75d61f731"
type: risk
name: "生成器の診断が握り潰され、前提条件の案内も出ない"
likelihood: medium
impact: medium
level: medium
threatens:
  - "REQ-7a2f1ecf-4675-45aa-80c0-a8fc58db9edd"
  - "REQ-f9920c87-8551-4aa3-bf03-26fdf4191ed6"
---
# RISK-cdca94ab-ed44-4d60-a04d-4dc75d61f731: 生成器の診断が握り潰され、前提条件の案内も出ない

## リスク

生成器は診断を CLI が渡す writer へ素通しし、失敗を `generator.Error` で構造化して返す
（→ REQ-7a2f1ecf-4675-45aa-80c0-a8fc58db9edd）。nix 生成器は experimental-features が未有効の
とき、自動付与せず有効化の手段を案内して止まる（→ REQ-f9920c87-8551-4aa3-bf03-26fdf4191ed6）。
これが崩れる形は 3 つある。

- **成功時の握り潰し** — eval の経路が nix の stderr をキャプチャしたまま捨て、成功時の警告
  （評価の非推奨・dirty tree 等）が利用者に届かない。生成器化の前はこの非対称があった
- **分類と案内の欠落** — experimental-features 未有効や config 名の不在が `Failed` に落ち、
  `Guidance` が空になる。nix 固有の文字列判定が nixgen の外へ漏れ、別経路で判定が食い違う
- **出力面の取り違え** — 人間向けに生の診断を再掲して二重に出す、あるいは `--json` の
  `errors[].message` からキャプチャが抜けて、stdout しか読まない消費側が原因を失う。
  `apply --all` の並列 build で診断が行の途中で混ざり、どの config の行か読めなくなる

## 影響

利用者が失敗の原因や解決手段を得られず、手で nix を叩き直して調べることになる。impact を
medium とするのは、配置や世代には影響せず、再実行と手動調査で回復できるため。

likelihood を medium とするのは、診断の経路（eval / build の 2 経路・人間向け / `--json` /
`--debug` の 3 形）が多く、文字列判定が nix のメッセージ変更に左右されるため。

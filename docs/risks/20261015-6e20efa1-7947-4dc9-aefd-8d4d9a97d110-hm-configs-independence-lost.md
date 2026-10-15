---
id: "RISK-6e20efa1-7947-4dc9-aefd-8d4d9a97d110"
type: risk
name: "home-manager モジュールで名前つき config ごとの独立性が失われる"
threatens:
  - "REQ-c847d1af-a437-46bb-bd64-42083810d034"
  - "REQ-c6891aeb-13c0-4ae7-9ad1-5c343735266a"
  - "REQ-5923ac79-4a2d-43cd-b56c-2f1000c01b44"
likelihood: medium
impact: medium
level: medium
---
# RISK-6e20efa1-7947-4dc9-aefd-8d4d9a97d110: home-manager モジュールで名前つき config ごとの独立性が失われる

## リスク

home-manager モジュールは `layat.configs.<name>` ごとに独立した配置物と profile を持ち、
activation は全 config を 1 回の `apply --all` で engine へ渡す。この独立性が崩れると、
standalone 経路と同じ役割分離が HM 経路では成り立たなくなる。

崩れ方は次の形をとる。

- **配置物の混線** — ある config の entry が別の config の配置物へ混ざる、または config の
  一部が起動の引数から漏れる
- **profile の共有** — 複数の config が同じ profile と世代を使い、一方の apply が他方の
  entry を stale とみなして撤去する
- **衝突の素通り** — 2 つの config が同じ target を持っても前段で止まらず、後勝ちの置換で
  apply のたびに配置が入れ替わる
- **失敗の巻き込み** — 1 つの config の配置失敗で他の config まで配置されない、または
  失敗が activation の終了コードに現れない

配線の退化（engine を呼ばず native 機構へ翻訳する等）は RISK-d734e24e-9af9-49f1-a22d-ade92f6554eb の
射程で、本 item は config 間の独立性だけを持つ。

## 影響

崩れても失われるのは layat が管理する symlink の配置と世代で、copy target は stale 除去の
対象外のため消えない。配線を直して再 activation すれば各 config の配置と世代は戻るので、
impact は medium とする。いずれの崩れ方も activation 自体は成功しうるため、利用者が気づき
にくい点は沈黙性として残る。

likelihood を medium とするのは、config の集合を起動引数へ展開する配線と `apply --all` との
合成が新しく、オプションや CLI 契約の変更のたびに退行が入りうる一方、評価アサートと実
activate の検証層がその多くを捕まえるため。

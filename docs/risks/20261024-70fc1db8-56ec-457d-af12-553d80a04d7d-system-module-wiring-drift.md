---
id: "RISK-70fc1db8-56ec-457d-af12-553d80a04d7d"
type: risk
name: "NixOS / nix-darwin モジュールが system mode の engine 起動の配線から外れる"
likelihood: medium
impact: high
level: high
threatens:
  - "REQ-c1b3ca5f-d2f7-443c-bc4b-b18413ca97b9"
  - "REQ-a220bb28-89de-42b0-a854-e7468755ecb1"
  - "REQ-d37d8433-a0c0-4387-aee9-9526f5cef70c"
---
# RISK-70fc1db8-56ec-457d-af12-553d80a04d7d: NixOS / nix-darwin モジュールが system mode の engine 起動の配線から外れる

## リスク

NixOS / nix-darwin のモジュールは、host の activation から system root を pin した engine を
起動するだけの配線である。home-manager モジュールと同じ起動を組むが、hook の置き場所と
起動環境は host ごとに異なり、そこが崩れても評価は通ってしまう。

崩れ方は次の形をとる。

- **hook の取り違え** — nix-darwin で独自名の activation script に起動を置くと、定義はできても
  実行されない。nix-darwin では利用者が `postActivation` に追記した内容を上書きする形でも起きる
- **root の取り違え** — 配置先が system root に固定されず、subdir が target に前置されない、
  または別の config に前置される
- **起動の形の分岐** — `--manifest` / `--jobs` / `--recopy` / `--backup` の規則が
  home-manager モジュールと食い違う。`layat.recopy` が false でも `--recopy` が付くか、
  true でも付かない（→ REQ-d37d8433-a0c0-4387-aee9-9526f5cef70c）。config が無いのに engine を起動する
- **dry-activate の取りこぼし** — NixOS の dry-activate 用スクリプトに起動が含まれない、
  または含まれても `--dryrun` が付かず、FS を変えない前提の操作で実配置が実行される
- **起動環境の欠落** — activation の PATH に `nix` が無いまま起動し、世代コミットが失敗する。
  `etc` / `users` より前に実行される

いずれも、activation から全 config の link-farm を 1 回の起動に渡す契約
（→ REQ-a220bb28-89de-42b0-a854-e7468755ecb1）からの逸脱として現れる。

## 影響

system mode は root 権限で `/` 配下を書き換える。dry-activate で実配置が実行される・root が
固定されず意図しない場所へ配置される、という重い側の崩れは、利用者が確認のつもりで
実行した操作で system のファイルを変える。規約の最重方向則に従い impact を high とする。
hook の取り違えは、配置が起きないことを利用者に知らせないという沈黙性を持ち、host の rebuild が成功して
見えるため気づきにくい。

likelihood を medium とするのは、起動の組み立ては `modules/kick.nix` に集約されて
home-manager と共有される一方、hook の置き場所・PATH・dry-activate の分岐は host ごとの
モジュールに残り、host 側の module system の変更でも崩れうるため。これらは評価の時点で
決まるので、評価アサートを持つ検証層が退行の多くを捕まえる。

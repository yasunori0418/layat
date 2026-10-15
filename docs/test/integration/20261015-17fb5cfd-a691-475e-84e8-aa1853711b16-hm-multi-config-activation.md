---
id: "TC-17fb5cfd-a691-475e-84e8-aa1853711b16"
type: test_condition
name: "HM の複数 config が独立した配置物と profile を持ち、衝突と部分失敗が config 単位で扱われる"
mitigates:
  - "RISK-6e20efa1-7947-4dc9-aefd-8d4d9a97d110"
---
# TC-17fb5cfd-a691-475e-84e8-aa1853711b16: HM の複数 config が独立した配置物と profile を持ち、衝突と部分失敗が config 単位で扱われる

## 条件

home-manager モジュールに名前つき config を複数与え、評価と実 activate の 2 層で確かめる。

評価の層（activate しない）:

- **配置物の分離** — config ごとの配置物が別の store パスになり、それぞれが自分の entry
  だけを持つこと
- **起動への展開** — 全 config が名前の辞書順に `<name>=<配置物>` の組として 1 回の
  `apply --all` へ渡ること。config が 1 つも無いときは engine を起動しないこと

実 activate の層:

- **profile の独立** — 各 config が状態ディレクトリ配下に独立した profile と世代を持つこと
- **衝突の前段停止** — 2 つの config が同じ target を持つとき、何も配置せずに非ゼロで
  終わり、出力に target と両 config 名が載ること
- **部分失敗の局所化** — 1 つの config の配置が失敗しても他の config は配置されて世代を
  コミットし、activate は非ゼロで終わること

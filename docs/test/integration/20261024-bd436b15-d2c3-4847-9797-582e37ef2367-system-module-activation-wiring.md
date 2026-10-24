---
id: "TC-bd436b15-d2c3-4847-9797-582e37ef2367"
type: test_condition
name: "NixOS / nix-darwin の activation が評価時点で system root を pin した engine 起動へ配線される"
mitigates:
  - "RISK-70fc1db8-56ec-457d-af12-553d80a04d7d"
---
# TC-bd436b15-d2c3-4847-9797-582e37ef2367: NixOS / nix-darwin の activation が評価時点で system root を pin した engine 起動へ配線される

## 条件

NixOS / nix-darwin のモジュールを host の構成として評価し、生成された activation の内容に
対してアサートする。実際に activate はしない。ここで確かめる性質は評価時に決まるが、
activate 後にしか観測できないため、評価だけで先に捕まえる。

- **hook の置き場所** — NixOS では `etc` と `users` の後に走る独自の activation script として、
  nix-darwin では利用者の追記と共存する `postActivation` への追記として起動が入ること
- **起動の形** — 名前つき config ごとのビルド済み配置物を名前の辞書順に並べ、engine を
  `apply --all` の 1 回の起動で渡すこと。並列度・退避・再コピーのオプションが
  home-manager モジュールと同じ規則でフラグとして乗ること。config が無ければ起動しないこと
- **root の pin** — 文書が root の種別として system を固定し、subdir を与えた config の
  target にだけ subdir が前置されること
- **NixOS の起動環境** — dry-activate 用スクリプトに含まれ、dry-activate のときだけ
  `--dryrun` を付けること。PATH に nix の実行ファイルを前置すること

実 activate を要するもの（`/etc` への実配置・profile の世代コミット・再 activate で世代が
増えないこと）はこの条件の外に置く。

配置元は固定の store パスを持つ test double で与え、評価結果をマシンによらず安定させる。

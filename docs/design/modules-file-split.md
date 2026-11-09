---
id: "DSG-aeb5e219-4784-4950-845c-35f9bab9179c"
type: design
name: "modules は common.nix に共通オプションを集約し、統合層ごとに 1 ファイルへ分ける"
status: stable
satisfies:
  - "REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10"
  - "REQ-a220bb28-89de-42b0-a854-e7468755ecb1"
  - "REQ-c1b3ca5f-d2f7-443c-bc4b-b18413ca97b9"
  - "REQ-8085f194-c903-4ecb-abd8-c719fe7b3292"
  - "REQ-c50df875-2cb0-4e72-8a21-858359a11cae"
  - "REQ-068f2fb3-5459-4325-9359-c931b5049e0d"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:22:37+09:00"
---
# modules は common.nix に共通オプションを集約し、統合層ごとに 1 ファイルへ分ける

## 設計

```
modules/
├── common.nix         # options 定義のみ（全モジュールが import）
├── kick.nix           # 純粋関数。config → link-farm と engine の起動文字列を組む（全モジュールが使う）
├── flake-parts.nix    # flake-parts module（perSystem.layat を flake.layat.<system> へ transpose）
├── home-manager.nix   # home.activation から layat エンジンを起動（root = homeRoot を pin）
├── nixos.nix          # system.activationScripts.layat から layat エンジンを起動（root = systemRoot を pin）
└── nix-darwin.nix     # system.activationScripts.postActivation から layat エンジンを起動（root = systemRoot を pin）
```

**`common.nix` を options 定義のみに絞り、全モジュールが import する**のが分割の要。
[全モジュールは共通オプションの同一集合を公開し、entries は configs 経由・root はモジュールが pin する](../requirements/module-common-options.md) は「モジュールごとに集合が
分岐しない」ことを求めるが、これは各モジュールが自前で options を定義していると
守れない（片方にだけ足す・デフォルトがずれる、といった分岐が構造的に起きる）。
定義を 1 ファイルに置いて import させれば、集合の同一性がファイル参照の一意性として
担保される。

`common.nix` が持つのは [全モジュールは共通オプションの同一集合を公開し、entries は configs 経由・root はモジュールが pin する](../requirements/module-common-options.md) が定める `enable` / `configs`（`entries` と `subdir` を含む）/
`backup` と、素の `entries` の deprecated 糖衣まで。root オプションは [全モジュールは共通オプションの同一集合を公開し、entries は configs 経由・root はモジュールが pin する](../requirements/module-common-options.md) の通り
どのモジュールも公開せず、各モジュールが自分の性質で pin する（home-manager → `homeRoot`、
NixOS / nix-darwin → `systemRoot`）。
home-manager は config ごとに `subdir` が非 null なら `homeRoot { subdir = <subdir>; }` を、null なら素の
`homeRoot` を root に組む（[HM モジュールは `configs.<name>.subdir` が非 null の config を homeRoot の subdir 適用形で組み、全 target を `$HOME/<subdir>/` 配下に置く](../requirements/hm-config-subdir-wiring.md)）。subdir の確認は `common.nix` にも
`home-manager.nix` にも置かず、lib の manifest 評価に委ねる。

**engine 起動の組み立ては `kick.nix` へ切り出し、`common.nix` には置かない**。`kick.nix` は
module ではなく、`config.layat` と root マーカーを組む関数を受けて link-farm 群と起動文字列を返す
純粋関数である。`--manifest` / `--jobs` / `--recopy` / `--backup` の規則を各モジュールが別々に
組むと分岐するため 1 か所に置く。NixOS / nix-darwin に固有のオプションは無い
（[NixOS / nix-darwin モジュールは root を systemRoot に固定する](../adr/system-modules-pin-systemroot.md)・[全モジュールは共通オプションの同一集合を公開し、entries は configs 経由・root はモジュールが pin する](../requirements/module-common-options.md)）。

統合層を 1 層 1 ファイルにするのは、各ファイルが持つのが配線
（[全モジュールと devShell は engine をキックするだけの配線とし、ネイティブ機構へ翻訳しない](../requirements/modules-are-engine-wiring.md)）であって、層ごとに異なるのは activation の
フック名と root の供給元だけだからである。

- HM の kick 契約: [home-manager モジュールの engine kick は activation から全 config のビルド済み link-farm を 1 回の起動で渡し、失敗で switch を止める](../requirements/hm-activation-contract.md)
- NixOS / nix-darwin の kick 契約: [NixOS / nix-darwin モジュールは activation から全 config のビルド済み link-farm を systemRoot で 1 回の起動に渡す](../requirements/system-module-activation-contract.md)
- 配線の内容: [engine の起動を entrypoint 駆動とビルド済み manifest の 2 クラスに束ね、統合層の差を link-farm の取得方法だけに閉じる](integration-layer-kick-classes.md) / [NixOS / nix-darwin モジュールは modules/kick.nix を介した system mode の配線とする](module-stubs.md)

`flake-parts.nix` は他と性質が違い、
[flake-parts での使い方は直書きと同一の derivation を生み CLI のアドレッシングを変えない](../requirements/flake-parts-module-path.md) を満たすための
transpose を担う。

## 出典

分割時点の原文（縮退前）の `docs/design.md`「プロジェクト構成」の `modules/` 内訳と
「モジュール統合設計」→「共通オプション」節。

なお同原文の当該節は `options.layat.entries` を直下に置く旧形で書かれているが、
本 item は現行の [全モジュールは共通オプションの同一集合を公開し、entries は configs 経由・root はモジュールが pin する](../requirements/module-common-options.md) /
[HM モジュール経由でも名前つき config ごとに役割分離した独立 profile を取れる](../requirements/hm-named-configs-profiles.md) に合わせている。
`layat.configs.<name>` の導入元は [HM モジュールは `layat.configs.<name>` ごとに独立した profile を持つ](../adr/hm-configs-independent-profiles.md)。

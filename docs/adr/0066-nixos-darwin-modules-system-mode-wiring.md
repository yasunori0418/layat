---
id: "ADR-0066"
type: adr
name: "NixOS / nix-darwin モジュールを system mode の配線として実装する"
status: 採用
issues:
  - "#519"
  - "#493"
origin: "Issue #493（epic: NixOS / nix-darwin モジュールの実装）の grilling（2026-10-24）"
justifies:
  - "REQ-a220bb28-89de-42b0-a854-e7468755ecb1"
  - "REQ-c1b3ca5f-d2f7-443c-bc4b-b18413ca97b9"
  - "REQ-8d965ca2-f8fd-44a4-87f3-94e850e9f85b"
  - "REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10"
  - "DSG-0e186e89-c8c6-4bca-9daa-03855e8d5cda"
revises:
  - "ADR-0003"
  - "ADR-0036"
references:
  - "ADR-0026"
  - "ADR-0062"
  - "ADR-0064"
---
# ADR-0066: NixOS / nix-darwin モジュールを system mode の配線として実装する

- ステータス: 採用
- 日付: 2026-10-24
- 関連: ADR-0026, ADR-0062, ADR-0064, GitHub Issue #519, #493
- 改訂対象: ADR-0003「決定」の NixOS / nix-darwin の root 供給元（`config.users.users.<user>.home`）を `systemRoot` に改める（モジュールが配線に徹することは不変）。ADR-0036 の「NixOS / nix-darwin モジュールは将来・別マイルストーン」を実装済みへ改める
- 起点: Issue #493（epic: NixOS / nix-darwin モジュールの実装）の grilling（2026-10-24）

## 背景

`nixosModules.default` / `darwinModules.default` は `common.nix` を import するだけのスタブで、host の構成から配置を宣言できない。system mode（`root = systemRoot`）は engine / CLI 側だけが実装済みで（→ ADR-0036）、`layat apply --manifest` はビルド済み manifest を渡す口として公開済み（→ ADR-0026）。残っているのは、host の activation から engine を起動する配線である。

当初の想定（ADR-0003）では、NixOS / nix-darwin モジュールは `layat.user` で配置先ユーザーを受け取り、その home を root に供給するとしていた。

## 決定

### 1. root は system mode 一本とし、`layat.user` を持たない

- 両モジュールは root を `systemRoot` に pin する。config ごとの `subdir` が非 null なら `systemRoot { subdir = <subdir>; }` を組む（HM の `homeRoot` と同じ形・→ ADR-0064）。
- `layat.user` オプションは設けない。ユーザー home への配置は home-manager モジュール（NixOS / nix-darwin の中で使う形を含む）が担う。

### 2. activation の hook

- NixOS は `system.activationScripts.layat` を `deps = [ "etc" "users" ]`・`supportsDryActivation = true` で定義する。
- nix-darwin は `system.activationScripts.postActivation.text` へ起動を追記する。nix-darwin は固定名のスクリプトだけを連結して実行するため、独自名のスクリプトは実行されない。

### 3. 起動の形は HM と同じにし、組み立てを `modules/kick.nix` へ集約する

- config が 1 つ以上あるとき、`layat apply --all --manifest <name>=<link-farm> … --jobs <n>[ --recopy][ --backup=<suffix>]` を 1 回だけ起動する。config が無ければ起動しない。
- config → link-farm の対応と起動文字列の組み立ては、純粋関数 `modules/kick.nix` に置き、home-manager モジュールもこれを使う。`modules/common.nix` は options 定義だけを持つ。
- configs 間の target 衝突は `apply --all` の前段検査に委ね、モジュールの評価時に assertions を書かない。

### 4. NixOS の起動環境

- `NIXOS_ACTION` が `dry-activate` のときだけ `--dryrun` を付ける。未設定（boot 時）を含む他の値では通常どおり起動する。
- activation の PATH に `config.nix.package` の `bin` を前置する。engine は世代コミットで `nix-env` を呼ぶが、NixOS の activation の PATH には `nix` が無い。
- nix-darwin には `--dryrun` の分岐を持たせない。

### 5. rollback と検証

- ユーザー向けの rollback は host の世代（`nixos-rebuild` / `darwin-rebuild`）に一本化し、layat の profile は前進だけで追従する（HM と同じ）。boot のたびに実行される再 activation で世代が増えないようにする扱いは ADR-0067 が決める。
- モジュールの評価結果は eval アサートの flake check（`checks.nixos-module` / `checks.darwin-module`）で、NixOS の実 activate は `runNixOSTest` の VM テスト（`checks.nixos-vm`。x86_64-linux のみ）で検証する。どちらも `tests/e2e/` のハーネスとは別系統に置く。nix-darwin の flake input は check 専用とし、`lib/` とモジュール本体へ依存を持ち込まない。

## 根拠

- **system mode 一本にする理由**: host の activation は root 権限で実行される。ここから user home へ配置すると、所有者の扱い・ユーザーの解決・profile 基底の選択をモジュール側で抱えることになる。home への配置は HM モジュールが既に担っており、同じ home へ配置する経路を 2 本にしない。
- **`postActivation` へ追記する理由**: nix-darwin では独自名の activation script が実行されない。`text` は `types.lines` なので、利用者の追記とも共存できる。
- **`kick.nix` へ集約する理由**: `--manifest` / `--jobs` / `--recopy` / `--backup` の規則を 3 モジュールが別々に組むと、片方だけが変わって分岐する。
- **dry-activate で `--dryrun` を付ける理由**: dry-activate は FS を変えないことが前提の操作であり、layat もその前提に合わせる。`supportsDryActivation = true` が無いと、スクリプトが dry-activate 用の activate に含まれず、分岐自体が実行されない。

## 既知の注意点

- **macOS には `/var/lib` が無く、engine が root で `/var/lib/layat` を作る**。system mode の記録基底（→ ADR-0062）は darwin でも変えず、`/var/db` へ切り替える分岐は足さない。
- nix-darwin は boot 時に activation のうち checks / etc / keyboard しか実行しない。再起動後の配置は次の `darwin-rebuild switch` まで検査されない。

## 影響

- **`modules/`**: `nixos.nix` / `nix-darwin.nix` を実装し、`kick.nix` を新設して `home-manager.nix` をそれへ寄せる。
- **`flake.nix`**: `nixosModules.default` / `darwinModules.default` が HM と同じ薄い wrapper で `layatPackage` を注入する。`layat.package` オプションは公開しない。
- **要求**: `layat.user` の要求を削除し、home mode の root 解決（REQ-8d965ca2-f8fd-44a4-87f3-94e850e9f85b）から NixOS / nix-darwin の句を落とす。hook 名（REQ-c1b3ca5f-d2f7-443c-bc4b-b18413ca97b9）と root pin（REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10）を改め、activation の起動契約（REQ-a220bb28-89de-42b0-a854-e7468755ecb1）を起票する。
- engine・CLI・manifest.json スキーマは本 ADR では変えない（世代スキップの範囲は ADR-0067）。

## 棄却した代替案

- **`layat.user` を取り、user home へ home mode で配置する（当初案）**: 根拠の 1 点目のとおり、HM モジュールと重複する経路を root 権限で持つことになる。
- **nix-darwin でも `system.activationScripts.layat` を使う**: 定義はできても実行されない。
- **起動文字列をモジュールごとに組む**: 規則が 3 か所に分かれ、HM との同一性をテストで守り続けることになる。
- **module の評価時に configs 間の target 衝突を assertions で止める**: `apply --all` の前段検査と規則を二重に持つことになる。

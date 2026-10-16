---
id: "ADR-0060"
type: adr
name: "HM activation を prebuilt manifest 群への `apply --all` 1 回起動にし、configs 間の衝突検出を CLI 前段へ委譲する"
status: 採用
issues:
  - "#128"
  - "#471"
origin: "Issue #128（HM 複数 profile・ADR-0035 の実装）の着手前 grilling（2026-10-15）"
justifies:
  - "REQ-c847d1af-a437-46bb-bd64-42083810d034"
  - "REQ-5923ac79-4a2d-43cd-b56c-2f1000c01b44"
  - "REQ-8085f194-c903-4ecb-abd8-c719fe7b3292"
  - "REQ-dec58330-6dad-47f7-8f56-2402764a89c7"
  - "REQ-38506419-c23d-4fce-b167-30715ce6a69d"
  - "REQ-194e4209-d804-4a4b-a2b8-3d39c6c33729"
  - "REQ-e1e1114b-ba07-4d57-8e04-6e30e39a5da3"
  - "REQ-5c6b07da-3d06-414d-8770-4f438234b322"
  - "REQ-fc1118b1-b0e8-4ddf-80f6-c70956651693"
revises:
  - "ADR-0026"
  - "ADR-0035"
  - "ADR-0038"
  - "ADR-0039"
  - "ADR-0055"
references:
  - "ADR-0045"
---
# ADR-0060: HM activation を prebuilt manifest 群への `apply --all` 1 回起動にし、configs 間の衝突検出を CLI 前段へ委譲する

- ステータス: 採用
- 日付: 2026-10-15
- 関連: ADR-0045, GitHub Issue #128, #471
- 改訂対象: ADR-0035 §3（profile ごとに独立した engine 起動・bash 側の部分失敗集約）と §4（configs 間 target 衝突の eval 時 assertion）を、`apply --all` 1 回起動と CLI 前段検査へ改める。ADR-0026 §2（`--manifest` と `--all` の排他）と §3（activation の invocation 形）を、`<name>=<link-farm>` の繰り返し指定で `--all` と合成できる形へ改める。ADR-0038 §2 末尾（HM は nix 側で検出）を CLI 前段検査へ寄せ、§1 の検査集合に prebuilt 集合を加える。ADR-0039 §3（複数 manifest を 1 コマンドに渡さない・HM activation の並列化はスコープ外）を反転する。ADR-0055 §2 / §5（prebuilt 生成器は単一 link-farm・`--all` との排他は不変）を、名前付き link-farm 集合を扱う形へ改める
- 起点: Issue #128（HM 複数 profile・ADR-0035 の実装）の着手前 grilling（2026-10-15）

## 背景

ADR-0035 は HM の `layat.configs.<name>` を決め、activation を「profile ごとに `layat apply --manifest <link-farm> <name>` を 1 回ずつ起動し、辞書順に走らせ、部分失敗を bash 側で集約する」形にした。configs 間の target 衝突は module eval の assertion で止めるとした。

その後、manifest の取得は生成器の契約越しになり（→ ADR-0055）、`apply --all` は前段の衝突検査（→ ADR-0038）・build と配置の並列化（→ ADR-0039）・部分失敗の集約（REQ-4cbd9a0d-9f94-4747-8881-56020dc6d5af）を CLI に持った。ADR-0035 の形で実装すると、これらと同じ機能（configs のループ・部分失敗の集約・衝突検出）を nix と bash で二重に持つことになる。layat は manifest を解釈して配置するアプリケーションであり、集合の扱いは CLI に寄せる方が層として素直である。

## 決定

### 1. activation は prebuilt manifest 群への `apply --all` 1 回起動にする

- `home.activation.layat` は全 config の link-farm を 1 回の起動に渡す。形は `layat apply --all --manifest <name>=<link-farm> … --jobs <n>` で、`<name>=<link-farm>` は config ごとに 1 つ、`<name>` の辞書順に並べる。`--backup` / `--recopy` などの修飾は同じ起動に付ける。
- `--jobs` は値が 0 でも常に明示で付ける。値は共通オプション `layat.jobs` から取る（オプションの定義は ADR-0061）。
- 各 config は従来どおり 1 config = 1 profile = 1 manifest で、profileDir 単位の flock・baseline との diff・stale 除去・世代コミットは config ごとに独立して走る（`apply --all` の既存の性質）。
- 部分失敗の続行と集約、非ゼロ終了は `apply --all` のもの（REQ-4cbd9a0d-9f94-4747-8881-56020dc6d5af）をそのまま使う。activation は `run` で非ゼロ終了を返して switch を止めるだけで、bash 側にループも集約も書かない。
- 報告（人間向けの集約表示・`--json` の `results[]`）は辞書順、配置の並列度は `--jobs` に従う。
- config が 1 つも無いときは `layat` を起動しない（`--manifest` の無い `apply --all` は entrypoint 経路になるため）。

### 2. `--manifest` は `--all` と合成するとき `<name>=<link-farm>` を繰り返す

- `apply --manifest <link-farm> [name]`（単一形）の挙動とエラー文は変えない。
- `--all` と合成するときは `--manifest <name>=<link-farm>` を 1 回以上繰り返す。最初の `=` で分割し、`<name>` が空なら入力エラーにする。
- 次は入力エラー（`--json` では `E_INPUT`）: `--all` なしで `--manifest` を 2 回以上指定する、`--all` なしで `<name>=` 形を使う、`--all` と `<link-farm>` 形を合成する、`--all` と合成した組で同じ `<name>` を 2 回以上指定する。`--manifest` と `-f` / `--generator` の併用は従来どおり入力エラー。
- 実装は prebuilt 生成器の集合化とする。一括の Roots は name → `manifest.Root`（`Targets` は各 link-farm の `manifest.json` の entries から導出）を返し、Build / DryBuild は name で link-farm を選ぶ。manifest.json スキーマ v1 は変えない。
- root フィルタ（`--home-root` 等）は特別扱いせず、`manifest.json` の rootKind による一般の選択規則に従う。

### 3. configs 間の target 衝突は activation の `apply --all` 前段検査で止める

- 同一 HM configuration の configs が正規化後 target を重複させた場合、activation の `apply --all` の前段検査（→ ADR-0038 §1）が build と配置の前に停止させる。どの config も配置されず、activation は非ゼロで終わり switch が止まる。
- module eval では検出しない（nix の `assertions` を持たない）。前段検査は prebuilt 集合にもそのまま適用され、全 config が home の 1 バケットに入る。

## 根拠

- **世代単位は曖昧にならない**: ADR-0035 と ADR-0039 §3 が「複数 manifest を 1 コマンドに渡す」CLI 拡張を棄却した理由は、世代コミットの単位が曖昧になることだった。`apply --all` は既に config ごとに独立した profile へ独立に世代をコミットし、全体を atomic にしない（REQ-4cbd9a0d-9f94-4747-8881-56020dc6d5af）。名前付き link-farm の集合を渡しても `<name>=<link-farm>` の 1 組が 1 profile の世代になり、atomic 性の単位は profile のまま変わらない。部分失敗時の状態も `apply --all` と同じ言葉で説明できる。
- **衝突検出を一本化する**: standalone の `--all` と HM activation が同じ前段検査を通るので、「同一の起動に載る衝突は build 前に止まる」が両経路で同じ実装になる。nix 側に第 2 の検出器を持つと、正規化規則（属性キー既定値・`target` 上書き・`subdir`）を 2 か所で揃え続ける負担が残る。
- **HM activation も並列化の対象にする**: ADR-0039 が「モジュール配線側の将来論点」とした HM の並列化が、`--jobs` を渡すだけで手に入る。
- **eval で止めなくなる代償は小さい**: 衝突は eval ではなく activation で止まるが、配置は 1 つも行われず switch は失敗する。誤設定を通さない点は eval 停止と変わらない。

## 影響

- **CLI（`cmd/layat`・`internal/generator`）**: `--manifest` の繰り返し指定と `--all` 合成、入力エラーの判定、prebuilt 生成器の集合化。
- **`modules/home-manager.nix`**: config ごとの `mkManifest` と、activation の `apply --all` 1 行起動。configs ループと部分失敗集約の bash は書かない。
- **`modules/common.nix`**: configs 間衝突の assertion を足さない。
- **要求**: REQ-c847d1af-a437-46bb-bd64-42083810d034 / REQ-5923ac79-4a2d-43cd-b56c-2f1000c01b44 / REQ-8085f194-c903-4ecb-abd8-c719fe7b3292 / REQ-dec58330-6dad-47f7-8f56-2402764a89c7 / REQ-38506419-c23d-4fce-b167-30715ce6a69d / REQ-194e4209-d804-4a4b-a2b8-3d39c6c33729 を本 ADR に合わせて改訂する。`--backup` の配線先（REQ-e1e1114b-ba07-4d57-8e04-6e30e39a5da3）と、configs 間衝突の委譲記述（REQ-5c6b07da-3d06-414d-8770-4f438234b322 / REQ-fc1118b1-b0e8-4ddf-80f6-c70956651693 / REQ-c6891aeb-13c0-4ae7-9ad1-5c343735266a）も同じ経路へ揃える。
- **ADR-0026 / ADR-0035 / ADR-0038 / ADR-0039 / ADR-0055**: 改訂注記を同じ変更で追記する。

## 棄却した代替案

- **ADR-0035 §3 のまま profile ごとに起動し bash で集約する**: `apply --all` が持つループ・集約・並列化を bash で再実装することになる。終了コードの優先度（error > conflict）も bash 側で写し直す必要がある。
- **module eval の assertion と CLI 前段検査を両方持つ**: 同じ規則の検出器が 2 つになり、正規化規則の乖離が衝突の見逃しや誤検出を生む。
- **`--manifest` を `<link-farm>` の繰り返しにし、name を manifest.json に持たせる**: manifest.json スキーマ v1 の拡張が要る。name は profile 選択で manifest の内容ではないため、起動側で与える方が既存の「位置引数 `name` は `--manifest` と直交」に沿う。
- **`--all` を付けずに `--manifest` の複数指定だけで集合とみなす**: `apply` の単一 config と集合の区別がフラグから読めなくなる。集合は `--all` が担うという既存の区分を保つ。

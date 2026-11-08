---
id: "ADR-87a2cf34-9592-4c81-8980-246bca579ca0"
type: adr
name: "manifest の取得をプロセス内の Go インターフェイスの生成器越しにする"
status: stable
sources:
  - id: "issue-395"
    resource: "https://github.com/yasunori0418/layat/issues/395"
    title: "#395"
origin: "Issue #395（epic: manifest 生成器のインターフェイス化）の grilling（2026-09-09）。ADR-0055 §1 を書き直した"
justifies:
  - "REQ-194e4209-d804-4a4b-a2b8-3d39c6c33729"
  - "DSG-25ad3cce-f921-4db2-aae0-c0263d4b8295"
  - "REQ-f4d7d4ab-fbdb-48c6-b29f-08dd88e72645"
references:
  - "ADR-28605860-770e-41df-95dd-7dc05b20497a"
  - "ADR-2f472fc4-23ed-4169-aa49-8bded9e336d9"
  - "ADR-51863e94-6a9d-4029-95e7-a7ad2bfa93c1"
  - "ADR-5caf1c14-a054-4c8d-95a2-c7f89e0c5246"
  - "ADR-63039f93-cd78-4356-8cd3-82c2f49a1aaa"
  - "ADR-6b7becf0-f029-41e3-895e-b672dce94d3d"
  - "ADR-c7ffc21a-4be1-4468-b63c-2f47eedd9ab3"
  - "ADR-f40c3d37-638d-420a-a0fe-3908cbf0aec0"
  - "ADR-fe035038-4415-4a1e-aa5c-509fdfce9955"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# ADR-87a2cf34-9592-4c81-8980-246bca579ca0: manifest の取得をプロセス内の Go インターフェイスの生成器越しにする

- ステータス: stable
- 日付: 2026-11-03
- 関連: [apply --all は config ごとの情報を 1 回の eval でまとめて取る](apply-all-single-eval.md), [eval を先に行い、flock を取ってから build する](eval-first-then-flock-then-build.md), [store path は out-link で受け取る](store-path-via-out-link.md), [内部コマンドの開示で透明性を保つ](internal-command-disclosure.md), [nix の experimental-features は利用者が有効にしておく前提にする](experimental-features-user-prerequisite.md), [CLI は cobra で組む](cobra-cli-framework.md), [root の必須性を別の確認を書かずに標準のエラーで担保する](root-required.md), [home mode の CLI と lib の版ずれを manifest の確認に任せて許容する](home-mode-skew-guarded-by-schemaversion.md), [1 config = 1 profile を全モードで自前に持つ](one-config-one-profile-all-modes.md), GitHub Issue #395
- 起点: Issue #395（epic: manifest 生成器のインターフェイス化）の grilling（2026-09-09）。ADR-0055 §1 を書き直した

## 背景

engine が受け取る契約は `manifest.json` 1 本で、engine は誰が manifest を生成したかに関知しない。
一方 `cmd/layat` は nix を前提に組まれていた。entrypoint の発見、`layat.<system>.<name>` の attr path、
rootKind の先読み eval、`nix build` による link-farm の取得、nix 固有のエラー文字列による失敗の分類が
CLI 層に散らばり、「manifest の生成は nix でなくてもよい」という位置付けを阻んでいた。

## 決定

manifest の取得は、プロセス内の Go インターフェイスとして定義した manifest 生成器越しに行う。
cmd 層に残すのは生成器の選択と engine への注入だけにする。engine は既存の注入用の関数（`BuildFunc`
など）をそのまま使い、変更しない。

生成器を経由するのは manifest を得るサブコマンドの全て（apply（`--dryrun` を含む）・gitignore・
list-generations・reset・rollback）とする。prune は nix を使わないので経由せず、`nix flake init` を
呼ぶ init は生成器化の範囲外にする。

生成器の契約の規範は [CLI は manifest 生成器の契約（Discover / Roots / Build / DryBuild）越しに manifest を得る](../requirements/manifest-generator-contract.md)、
パッケージ構成は [internal/generator は契約・prebuilt・Fake を親パッケージに、nix 実装を nixgen サブパッケージに置き、cmd 層は注入点だけを持つ](../design/generator-package-layout.md)、engine と CLI の 2 層構成は
[layat は CLI とエンジンの 2 層で構成する](../requirements/two-layer-architecture.md) に置く。

## 根拠

実際に使う生成器は nix 1 つしかない。Go インターフェイスなら、nix 実装を切り出し、`--manifest` の
処理とテストダブルを同じ契約の実装にまとめるだけで、cmd 層から nix への依存を外せる。プロセス間の
規約を決めるのは、2 つ目の実生成器が現れてからで足りる。

## 棄却した代替案

### 外部プログラム規約（`layat-manifest-<lang>` を exec する）

2 つ目の実生成器が無い段階で、wire 形式・発見規則・バージョニングを決めることになり、使われない
仕様が固まる。

### 生成器の抽象を設けず、cmd 層の nix 呼び出しを保つ

生成器を足すたびに cmd 層全体へ分岐が入る。`--manifest` の処理とテストダブルも、nix の処理とは
別の実装のまま残る。

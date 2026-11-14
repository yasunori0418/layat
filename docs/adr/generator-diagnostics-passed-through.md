---
id: "ADR-6687c0dc-957d-4913-99c6-168b0fe552ad"
type: adr
name: "生成器の診断を成功時も stderr へ素通しする"
section: manifest-generator
status: stable
sources:
  - id: "issue-396"
    resource: "https://github.com/yasunori0418/layat/issues/396"
    title: "#396"
origin: "Issue #396（評価器の診断の扱い）の結論。ADR-0055 §6 の診断の出力先を書き直した"
justifies:
  - "REQ-7a2f1ecf-4675-45aa-80c0-a8fc58db9edd"
references:
  - "ADR-36bfbfbd-f1dd-4581-a75a-943d80f6ccf5"
  - "ADR-3dc21349-5504-48a4-ab88-812b92c3bdf9"
  - "ADR-a19bf484-368d-4d01-b539-52de32510fa5"
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
  at: "2026-11-03T23:00:56+09:00"
---
# 生成器の診断を成功時も stderr へ素通しする

- ステータス: stable
- 日付: 2026-11-03
- 関連: [指定が無いときは成功時に何も出力しない](silent-on-success.md), [-v で配置レポートを opt-in で出す](verbose-opts-into-placement-report.md), [内部コマンドの開示を --debug に分ける](debug-discloses-internal-commands.md), [apply --all は config ごとの情報を 1 回の eval でまとめて取る](apply-all-single-eval.md), [eval を先に行い、flock を取ってから build する](eval-first-then-flock-then-build.md), [store path は out-link で受け取る](store-path-via-out-link.md), [内部コマンドの開示で透明性を保つ](internal-command-disclosure.md), [nix の experimental-features は利用者が有効にしておく前提にする](experimental-features-user-prerequisite.md), [CLI は cobra で組む](cobra-cli-framework.md), [root の必須性を別の確認を書かずに標準のエラーで担保する](root-required.md), [home mode の CLI と lib の版ずれを manifest の確認に任せて許容する](home-mode-skew-guarded-by-schemaversion.md), [1 config = 1 profile を全モードで自前に持つ](one-config-one-profile-all-modes.md), GitHub Issue #396
- 起点: Issue #396（評価器の診断の扱い）の結論。ADR-0055 §6 の診断の出力先を書き直した

## 背景

nix の eval は成功時に子プロセスの stderr を捨てていたため、`mkManifest` の `lib.warn` のような
評価器の警告が利用者に届かなかった。build は stderr を素通ししていて、eval と build で挙動が
食い違っていた。一方 CLI は成功時に何も出さない規律を持つ（→ [指定が無いときは成功時に何も出力しない](silent-on-success.md)）。

## 決定

生成器は診断を CLI から渡された writer へ書き、stdout には書かない。成功時も eval と build の両方で
子プロセスの stderr を素通しする。生成器の stderr は成功時に沈黙する規律の対象外にする。

writer と stdout の規範は [生成器は診断を CLI が渡す writer へ素通しし、失敗を構造化して返し、表示は CLI が決める](../requirements/generator-diagnostics-contract.md) に置く。

## 根拠

評価器の警告や deprecation の通知は、成功した実行でこそ利用者が受け取るべき情報である。沈黙の規律は
layat 自身の進捗やレポートを抑えるためのもので、生成器の警告を握り潰す根拠にはしない。stdout に
書かないのは、`--json` のときに stdout をエンベロープだけにするためである。

## 棄却した代替案

### 生成器の stderr も成功時は沈黙させる

評価器の警告が利用者に届かない欠陥（Issue #396）をそのまま固定する。

### stderr をキャプチャし、CLI が要約して出す

要約の基準が生成器ごとに異なり、CLI が生成器の出力形式を知る必要が生じる。

---
id: "ADR-9755ba76-4c02-4b4e-a8c0-031a24a3a5b9"
type: adr
name: "世代バックエンドを nix 実装 1 つに留める"
section: generations
status: stable
sources:
  - id: "issue-406"
    resource: "https://github.com/yasunori0418/layat/issues/406"
    title: "#406"
origin: "Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0057 §1・§6a を書き直した"
justifies:
  - "REQ-833205ca-adfc-4c86-bf66-ae4d8fe4885b"
references:
  - "ADR-2119433e-de69-4447-85a5-575d58246007"
  - "ADR-521765d7-7e31-42f4-b548-979185a7586a"
  - "ADR-86a77232-dc99-4c04-bd6d-2127410e8828"
  - "ADR-b06dd8ab-317e-41ee-9c80-dd70dcd7f2fd"
  - "ADR-62540f8f-a8ba-4441-97b0-36b50d7862d8"
  - "ADR-1cbacca1-1930-46ca-bf82-30310f82216e"
  - "ADR-4c8d459a-c371-4f0d-b009-16e2ca096e49"
  - "ADR-eb1aa25f-a0ca-4efe-bcd9-4cbe11368764"
  - "ADR-ffe1afcb-0b51-4b87-b698-0a6dea307bdd"
  - "ADR-8b8c9109-a479-4571-8b10-5a6079127441"
  - "ADR-1908099c-87f7-4df1-bda5-7731df7d9ed7"
  - "ADR-efa7685c-6fa4-45df-b7a3-8f3f58e4118c"
  - "ADR-764ffff1-6b5a-4ed7-89d0-5700eb3ee883"
  - "ADR-9723dbb3-6659-4a69-ac4f-411c98bfa818"
  - "ADR-45380022-d959-4b22-aa05-5993ac1f3c87"
  - "ADR-a51897c5-f709-4b7a-8db2-61e6ebb0040c"
  - "ADR-f607b605-c69e-4c7e-ac89-75b7bce5bee1"
  - "ADR-88231ffc-4836-4252-be34-894944fd28c3"
  - "ADR-d15a216f-b5d6-4df9-9825-b23a341bec1d"
  - "ADR-2f472fc4-23ed-4169-aa49-8bded9e336d9"
  - "ADR-dfdcc588-2e23-4d3b-ae17-b508a8d85ff0"
  - "ADR-553eea60-6ab4-4da7-93a9-3acf44b2fae3"
  - "ADR-36cb1e78-90c0-47d2-bbb1-4bfec5db2964"
  - "ADR-9acd3e01-58f1-48ca-b84f-f33a5bc1aaa5"
  - "ADR-5908eb28-40a5-49d5-a09e-15b35082b970"
  - "ADR-e88f074e-95f4-4e04-97fc-eae2d0c695a3"
  - "ADR-aec6bd1d-f776-4d7c-9337-1ed78fee6180"
  - "ADR-c4a3d9be-19ea-461d-ab0f-2037adc494e1"
  - "ADR-3878ea60-4f4f-48a3-a389-3b7fbb672897"
  - "ADR-30f105c7-85a8-4fab-9dea-3277d932b524"
  - "ADR-419914ba-9896-4dd1-853b-fb514fcb9f38"
  - "ADR-51863e94-6a9d-4029-95e7-a7ad2bfa93c1"
  - "ADR-58fdbad9-5ed7-469d-b6b2-6462b2173888"
  - "ADR-5fc58c8f-dc0e-4c1d-8c1e-17d7ac7f0385"
  - "ADR-74423bb8-1453-4689-a25e-30bf46536afc"
  - "ADR-b1e1dd5e-61a4-4d04-8389-137ef2347c56"
  - "ADR-d97154cd-a2e5-4853-a682-800ff9435c89"
  - "ADR-f90509f2-d75a-4162-baa5-a5cfdb8a7c30"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# 世代バックエンドを nix 実装 1 つに留める

- ステータス: stable
- 日付: 2026-11-03
- 関連: [engine を単一の固定 Go バイナリにする](engine-fixed-go-binary.md), [lib は純データ生成に徹する](lib-produces-data-only.md), [mkActivationScript で config ごとのラッパー derivation を提供する](mkactivationscript-wrapper.md), [manifest.json を schemaVersion 1 の唯一の契約として固定する](manifest-json-contract-schemaversion-v1.md), [世代操作はパスと整数だけを受け渡す 5 操作の契約越しに行う](generation-backend-contract-paths-and-ints.md), [世代の一覧を nix-env を経由せず世代リンクの走査で得る](generation-list-scans-links.md), [世代バックエンドのテストダブルを FS ベースの Fake に統一する](fs-based-generation-fake.md), [世代バックエンドの注入を必須にし nix-env への暗黙のフォールバックを持たない](generation-backend-required-no-fallback.md), [profileDir の内部レイアウトを世代バックエンド契約の固定部分にする](profiledir-layout-fixed-in-contract.md), [前世代の読み取りは契約に載せず engine が直接読む](generation-reads-outside-contract.md), [engine が直接叩く外部コマンドを git に限る](engine-external-command-git-only.md), [Build の事後条件を世代バックエンドがコミットできる link-farm のパスに置く](build-postcondition-committable-link-farm.md), [state の基底を設定で上書きさせない](state-bases-not-configurable.md), [.pending の後片付けを gcroot を知らない engine の FS 操作として行う](pending-cleanup-without-gcroot.md), [flock は明示実行で待ち、hook 実行でスキップする](flock-mode-explicit-wait-hook-skip.md), [独立した操作はサブコマンドにし、操作の修飾はフラグにする](subcommand-set.md), [終了コードは 0 / 1 / 2 の 3 値にする](exit-code-scheme.md), [CLI が一次 UX で devShell に pin 版を同梱する](cli-primary-ux-devshell-pinned.md), [templates と layat init は固定 ref で導入する](templates-and-init-fixed-ref.md), [eval を先に行い、flock を取ってから build する](eval-first-then-flock-then-build.md), [テストを sandbox 可否で flake check と E2E に分ける](test-strategy-layers.md), [engine は cmd/ と internal/ に置きライブラリ化しない](engine-internal-cmd-layout.md), [生成器は明示の指定だけで選ぶ](generator-chosen-explicitly-only.md), [生成器の選択値を nix だけにする](generator-values-nix-only.md), [設定ファイルは -f のディレクトリか cwd だけで探す](config-file-lookup-no-upward-search.md), [設定ファイルを strict な TOML で書く](config-file-strict-toml.md), [CLI の依存に go-toml/v2 を許可する](cli-depends-on-go-toml.md), [入力の不正を E_INPUT に分類する](input-errors-classified-e-input.md), [--generator を prune と init では無視する](generator-flag-ignored-by-prune-init.md), [symlink farm を GC アンカー専用にする](farm-as-gc-anchor-only.md), [profile は専用ディレクトリレイアウトに置く](profile-directory-layout.md), [store path は out-link で受け取る](store-path-via-out-link.md), [.pending の残留を許容する](pending-leftover-tolerated.md), [空の entries を正当な全クリアとして許可する](empty-entries-allowed.md), [新規配置と張替えを先に、世代のコミットを最後に行う](placement-order-commit-last.md), [apply --all は失敗しても残りを続け、最後に集約して非 0 で終える](apply-all-continue-and-aggregate.md), [rollback は FS を収束させてからポインタを動かす](rollback-baseline-and-pointer-last.md), [世代は link farm と nix profile に乗せる](generations-on-link-farm-and-nix-profile.md)
- 起点: Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0057 §1・§6a を書き直した

## 背景

engine の世代操作は、世代バックエンドの契約越しに行う。
契約の実装は nix profile を使うものしかない。
世代管理を独立した CLI へ切り出す方針はあるが、それは別の issue（#447）が扱う。

契約を設けるにあたり、非 nix の実装と、実装を選ぶ仕組みまで作るかが問題になった。

## 決定

世代バックエンドの実装は nix profile を使う nixprofile 1 つに留める。
非 nix の実装は作らない。
実装を選ぶフラグ・環境変数・`layat.toml` のキーも足さない。

2 つ目の実装が現れたら、manifest 生成器の選択と同じ優先順位（フラグ → 環境変数 → プロジェクト設定 → ユーザー設定 → 既定）で選ぶ。
項目名と値はそのときに決める。

nixprofile が唯一の実装である間の規範は [engine は世代操作を世代バックエンドの契約（Commit / List / Switch / UserBase / SystemBase）越しに行う](../requirements/generation-backend-contract.md) に置く。

## 根拠

実装が 1 つの段階で選択の仕組みを作っても、選ぶ対象が無いので検証できず、使われない項目が設定インターフェイスに固まるだけになる。

非 nix の実装は、世代が参照する link-farm を GC から守る仕組みを自前で背負うので、ADR-0006 が engine による世代 symlink の直接管理を退けた理由が、そのまま残る。

選び方の優先順位を生成器と揃えておけば、2 つ目の実装が来たときに設定の解決規則を新しく設計せずに済む。

## 棄却した代替案

### 非 nix の世代バックエンドを同時に実装する

自前の symlink・世代の採番・GC アンカーの代替を持つ案。
GC への安全性を自前で背負う。
世代管理は別のアプリケーションとして検討する。

### 実装が 1 つのうちに選択の設定を用意しておく

選べる値が nixprofile しかなく、使われない項目が設定インターフェイスに残る。

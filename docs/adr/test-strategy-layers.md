---
id: "ADR-dfdcc588-2e23-4d3b-ae17-b508a8d85ff0"
type: adr
name: "テストを sandbox 可否で flake check と E2E に分ける"
status: stable
origin: "ADR-0006 §テスト戦略・ADR-0012 §3 を書き直した"
justifies:
  - "TP-e7c25263-6d2d-4a37-8275-26906889d912"
  - "TP-36e90d5d-4524-4294-bc72-ee263bb02782"
  - "TP-b7f1dc79-0222-4b6e-9e91-0545046e34f2"
  - "TP-229b69c0-cf5e-4fb6-a353-27e5064d93e9"
  - "DSG-fb49e36c-eb20-4efa-8cac-88ef0873db8e"
  - "DSG-836aa5cb-0389-4adb-990b-144fe5aeffe3"
  - "DSG-901351ea-f01b-4bb0-a470-b890d92c980b"
  - "DSG-2947b4a5-ba5c-4218-8478-624cba156116"
  - "INF-d1230e1f-8ba8-49d8-8386-409bfbb7dd27"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:26:59+09:00"
---
# テストを sandbox 可否で flake check と E2E に分ける

- ステータス: stable
- 日付: 2026-11-01
- 起点: ADR-0006 §テスト戦略・ADR-0012 §3 を書き直した

## 背景

layat は Nix の lib、Go の engine と CLI、実際の nix と FS を使う配置の 3 層からなる。層ごとに
検証の手段と、検証を実行する場所を決める必要がある。nix の sandbox で成立するテストと、
nix profile や実 FS を書き換えるために sandbox では成立しないテストがある。

## 決定

lib は nix-unit で不変条件を、namaka で manifest 全体のスナップショットを検証する。engine と CLI は
nix を使わない実 FS の Go テストで検証し、tmpdir での統合テストを主にして、保守的削除の
不変条件は table-driven で厚く検証する。非 NixOS でも nix があれば動くという主張は、実際の nix を
使う E2E で検証する。

sandbox で成立する lib と Go のテストは `nix flake check` に集めて PR ごとに実行する。E2E は
`nix flake check` の外の別ジョブで実行する。NixOS モジュールの実際の activate は runNixOSTest の
check で検証し、E2E のハーネスとは別系統に置く。

層ごとの検証範囲、手段、CI のジョブ構成は、次の item に置く。

- 層ごとの検証範囲
  - [評価テストは nix-unit で不変条件を、namaka で manifest 全体のスナップショットを見る](../test-plan/nix-unit-namaka-split.md)
  - [エンジンとコマンド層の Go テストは nix を介さない実 FS 統合テストを主戦力とする](../test-plan/go-test-layering.md)
  - [非 NixOS で動く主張を実 nix の一気通貫 E2E で検証する](../test-plan/e2e-harness-scope.md)
  - [NixOS / nix-darwin モジュールで使うときの実 activate は E2E ハーネスの対象外とする](../test-plan/nixos-vm-test-future.md)
- 手段
  - [lib は nix-unit の評価テストと namaka のスナップショットの 2 手段で検証する](../design/lib-test-nix-unit-namaka.md)
  - [engine は nix を使わない tmpdir 統合テストで検証し、古くなった配置の除去の安全不変条件を table-driven に重点配分する](../design/engine-test-go-unit-tmpdir.md)
  - [E2E は tests/e2e の bash ハーネスに置き、シナリオの詳細は同ディレクトリの README へ委譲する](../design/e2e-harness-shape.md)
  - [NixOS モジュールの実 activate は runNixOSTest の checks.nixos-vm で検証し、E2E ハーネスとは別系統に置く](../design/nixos-vm-test-position.md)
- CI のジョブ構成
  - [CI パイプライン（flake check の os×system マトリクスと E2E ジョブ）](../infrastructure/ci-pipeline.md)

## 根拠

lib と Go のテストは純粋な評価か偽の source で動き、sandbox と相性がよい。`nix flake check` に
集めると、CI と手元で同じ判定が得られる。E2E は nix profile を実際に書き換えて rollback まで
確かめるので、ネットワークと特権が制限された sandbox では成立しない。flake check の外に出すのは
技術的な必然で、検証したい環境（非 NixOS と nix）とジョブの環境が一致する。

古くなった配置の除去は利用者のファイルを消しうる、最も安全性が問われるロジックなので、Go の
ユニットテストで条件の組み合わせを網羅する。

## 棄却した代替案

### 全ての層を CI 用 devShell の 1 ジョブで実行する

純粋な評価で済む lib と Go のテストまで `nix flake check` の外に出て、PR ごとの常時の確認と
手元での同じ判定を失う。

### 全ての層を `nix flake check` に詰める

E2E の profile の書き込みと特権が sandbox と衝突し、成立しない。

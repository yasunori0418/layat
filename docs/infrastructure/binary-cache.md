---
id: "INF-af33c5a1-f846-4f09-8e43-c6ea042d6e06"
type: infrastructure
name: "バイナリキャッシュ（main push 時の cachix 投入）"
status: stable
depends_on:
  - "INF-d1230e1f-8ba8-49d8-8386-409bfbb7dd27"
satisfies:
  - "QA-58522afb-31d5-4a1f-a7df-0858efa9e44b"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:22:37+09:00"
---
# バイナリキャッシュ（main push 時の cachix 投入）

## 構成

cachix のキャッシュ `yasunori0418` へ `packages.<system>.layat` のビルド生成パスを投入する。
consumer の CI / ローカルが「最新 main の `layat` を実ビルドせずキャッシュから取得する」ための基盤。

- 構成は [CI パイプライン（flake check の os×system マトリクスと E2E ジョブ）](ci-pipeline.md) と同じ os×system の 3 環境マトリクス
  （`ubuntu-latest`=x86_64-linux / `ubuntu-24.04-arm`=aarch64-linux /
  `macos-latest`=aarch64-darwin）で `nix build .#packages.<system>.layat` をネイティブ実行する
- 投入は `.github/actions/setup-nix` が内包する `cachix-action` が authToken 指定で生成パスを
  自動 push する処理に任せる。明示の push ステップは置かない
- 単一パッケージのため、未キャッシュのものだけを選別する plan / collect の多段構成は採らない

## トリガ

`push: main` + `paths` + `workflow_dispatch`。layat バイナリの内容を決める入力
（`**.nix` / `**.go` / `go.mod` / `go.sum` / `flake.lock` / `dev/flake.lock` / `VERSION`）が変わった
main push でのみ投入し、docs だけの変更では実行しない。`flake.nix` は `VERSION` を読んで
`packages.layat` の version に使うため、bump PR のマージでも投入が起動する。

投入契機を tag push にしない理由は、tag リリースに紐づけるとリリース間の main 変更がキャッシュ
されず、layat が flake input として消費される形と噛み合わないため。ここは [リリース自動化（VERSION ファイル起点の bump PR・自動タグ・自動リリースノート）](release-automation.md)
とは独立に動く — キャッシュは main の最新、リリースは `VERSION` の変更が
それぞれ駆動する。

なお required status check を持つ [CI パイプライン（flake check の os×system マトリクスと E2E ジョブ）](ci-pipeline.md) と異なり、この workflow は
main への push 後に実行されるため、トリガ段の `paths` フィルタをそのまま使える。

## CI パイプラインへの依存

投入の実行契機は独立（PR 対 main push）だが、**実行基盤は [CI パイプライン（flake check の os×system マトリクスと E2E ジョブ）](ci-pipeline.md) に
乗っている**。この workflow は自前の nix セットアップを持たず、`.github/actions/setup-nix`
（`install-nix-action` + cachix 認証込みの `cachix-action`）へ丸ごと委ねる。投入の処理そのものが
その composite action の中にあるため、`setup-nix` の構成が変われば投入も直接影響を受ける。

os×system マトリクスも CI パイプラインと同一の内容を踏襲するが、こちらは共有の仕組みではない。
`setup-nix` と違って reusable workflow / composite action のような共有機構は挟んでおらず、
同じ 3 組をこの workflow の `strategy.matrix` へ独立にハードコードした写しになっている。
したがって **マトリクスを変えるときは 2 箇所を揃えて直す**。CI パイプライン側だけで環境を
増減しても、この workflow は利用者に知らせずに旧構成のままビルドし続け、キャッシュに載る system が CI の
検証対象とずれる。[マージの条件（main の ruleset と required status checks）](merge-gate.md) が ruleset 側について挙げる追従点と同じ種類のものが、
マトリクスの定義にもある。

## 出典

[キャッシュへの投入を main への push で起動する](../adr/cachix-push-on-main.md)、[キャッシュは 3 system の matrix ビルドで投入する](../adr/ci-cache-via-matrix-build.md)、
[キャッシュへの投入を cachix-action の自動 push に任せる](../adr/cachix-push-via-cachix-action.md)。

# 基盤（`infrastructure`）

* [バイナリキャッシュ（main push 時の cachix 投入）](./binary-cache.md) - cachix のキャッシュ `yasunori0418` へ `packages.<system>.layat` のビルド生成パスを投入する。consumer の CI / ローカルが「最新 main の `layat` を実ビルドせずキャッシュから取得する」ための基盤。
* [CI パイプライン（flake check の os×system マトリクスと E2E ジョブ）](./ci-pipeline.md) - GitHub Actions 上の `test.yml` のうち、**flake check と E2E の系統**と、テストを実行しないが同じ workflow が抱える対応表生成（`test-doc-matrix`）。同居する `sara` ジョブだけはトレーサビリティ検証基盤（INF-659b139d-0cf8-4c65-b30d-93c5ee2dfc71）の担当で、本 item には入らない。
* [ドキュメントサイト（Astro Starlight + Cloudflare Pages・ビルド時リファレンス生成）](./docs-site.md) - Astro（Starlight）のサイトを **同一リポジトリの `site/`** に置く。別リポジトリへは分離しない。リファレンスの抽出元（`lib/` の doc-comment・Go ソース）とサイトが同じリポジトリにあることで、API 変更とドキュメント更新が同一 PR で原子的に進む。
* [マージの条件（main の ruleset と required status checks）](./merge-gate.md) - `main`（`~DEFAULT_BRANCH`）に GitHub の ruleset を適用し、テスト成功を **技術的に**マージ条件へ課す。CLAUDE.md の「main 直接コミット禁止・PR 経由マージ」は運用ルールに過ぎず、テストが失敗していてもマージできる状態を塞ぐ。
* [リリース自動化（VERSION ファイル起点の bump PR・自動タグ・自動リリースノート）](./release-automation.md) - バージョンの一次情報をリポジトリ直下の `VERSION`（プレーンテキスト 1 行・semver）に置き、**その変更が main へマージされることがリリースを駆動する**。手動のタグ打ちは要らない。
* [トレーサビリティ検証基盤（sara によるドキュメントグラフの機械検証）](./traceability-verification.md) - `docs/` を [sara](https://github.com/cledouarec/sara)（Rust 製 CLI。個人 NUR 経由で dev flakeへ入れる）のナレッジグラフとして扱い、要求・設計・リスク・テストの紐付けが一貫しているかを機械検証する。要求とリスクの紐付けの一貫性管理が主目的で、「仕様とテストの全体像把握」を支える。

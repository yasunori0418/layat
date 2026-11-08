---
id: "INF-8b97573f-d4d6-4abf-85e2-d859afbd96c6"
type: infrastructure
name: "マージの条件（main の ruleset と required status checks）"
status: stable
dod:
  - "DOD-01"
  - "DOD-02"
depends_on:
  - "INF-d1230e1f-8ba8-49d8-8386-409bfbb7dd27"
satisfies:
  - "QA-a5f7f088-a459-4bb2-9674-82b1a4a52053"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:22:37+09:00"
---
# INF-8b97573f-d4d6-4abf-85e2-d859afbd96c6: マージの条件（main の ruleset と required status checks）

## 構成

`main`（`~DEFAULT_BRANCH`）に GitHub の ruleset を適用し、テスト成功を **技術的に**
マージ条件へ課す。CLAUDE.md の「main 直接コミット禁止・PR 経由マージ」は運用ルールに過ぎず、
テストが失敗していてもマージできる状態を塞ぐ。

required status check に載せるのは、[CI パイプライン（flake check の os×system マトリクスと E2E ジョブ）](ci-pipeline.md) が PR で実際に出す 4 つの
check 名:

- `e2e`
- `flake-check (ubuntu-latest, x86_64-linux)`
- `flake-check (ubuntu-24.04-arm, aarch64-linux)`
- `flake-check (macos-latest, aarch64-darwin)`

matrix ジョブの check 名は `(os, system)` 付きになるため、ジョブ名 / matrix を変えると check 名も
変わる。ruleset 側の追従が要る。

ruleset `main branch protection` は 4 ルールで構成する。

| ルール | 効果 |
|---|---|
| `required_status_checks` | 上の 4 つの check の成功を必須にする |
| `pull_request` | `main` への直接 push を禁止する |
| `deletion` | `main` の削除を禁止する |
| `non_fast_forward` | `main` への force push を禁止する |

`deletion` / `non_fast_forward` は [main へのマージにテストの成功を必須にする](../adr/tests-required-for-merge.md) / [main への直接 push を ruleset で禁じる](../adr/no-direct-push-to-main.md) が挙げていない。ruleset の適用時に追加した。

縛りの度合い:

- bypass を許さない（`bypass_actors` 未指定）。owner でも PR + テストを通さないとマージできない
- up-to-date は強制しない（`strict_required_status_checks_policy: false`）。個人運用で PR が
  多くないうちは再走コストを避ける
- レビュー承認は不要（`required_approving_review_count: 0`）。テスト成功のみをマージ条件にする

## public 化

ruleset / branch protection は個人アカウント所有の private + free プランでは API・UI ともに
ロックされる（`403: Upgrade to GitHub Pro or make this repository public`）。GitHub Pro 課金は
採らず public 化で機能を解放する方針を採り、**この前提ステップは完了済み**。リポジトリは public
で、ruleset は `enforcement: active` で適用されている。

## docs-only PR との両立

required status check は **実行されない check を `Expected / 待機中` として扱いマージをブロックし
続ける**。一方で **ジョブが生成されたうえで `if:` によりスキップされた check は "Skipped" =
成功扱い**になる。この非対称性が、[CI パイプライン（flake check の os×system マトリクスと E2E ジョブ）](ci-pipeline.md) がトリガ段の `paths` フィルタを使わず変更検出
action + `if:` スキップを採る理由である。

ただし `if:` を**ジョブ段**に置けるのは matrix を持たないジョブに限る。`(os, system)` 付きの
3 context が required な `flake-check` はステップ段に条件を置く（理由と実装は
[CI パイプライン（flake check の os×system マトリクスと E2E ジョブ）](ci-pipeline.md)「トリガと変更検出」）。

## 出典

[main へのマージにテストの成功を必須にする](../adr/tests-required-for-merge.md)、[main への直接 push を ruleset で禁じる](../adr/no-direct-push-to-main.md)、
[変更検出で nix を回さない PR の必須の確認を skip で通す](../adr/change-detection-gates-ci-jobs.md)。`docs/dev/definition-of-done.md` の
DOD-01 / DOD-02 は、この ruleset の required check と同じ対象を指す。

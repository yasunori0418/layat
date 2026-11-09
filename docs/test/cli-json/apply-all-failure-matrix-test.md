---
id: "CASE-4a8860b1-5a88-4225-9702-f946be9c984e"
type: test_case
name: "apply_all_failure_matrix_test.go — build 失敗 × 配置失敗 × 成功 × skip（× conflict）の部分失敗マトリクス"
status: stable
resource: "cmd/layat/apply_all_failure_matrix_test.go"
covers:
  - "TC-e42a5438-039b-41b3-9a39-0e8dc1f25cb3"
  - "TC-9ccb2ffe-7bf6-41ea-a9ce-df06cdda0509"
  - "TC-ddee6cc4-bc10-4107-bc7e-288a5fb62f1f"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:24:34+09:00"
---
# apply_all_failure_matrix_test.go — build 失敗 × 配置失敗 × 成功 × skip（× conflict）の部分失敗マトリクス

## 対象

`cmd/layat/apply_all_failure_matrix_test.go`

## 検証内容

config ごとの結果を種別名の config として注入し、非空の組合せを全て `runApplyAll` と同じ
合成（stage 1 の並列 build → 失敗の短絡 → stage 2 の集約）で `--jobs 1` と全並列の両方に
流す。`go test -race` で集約のデータ競合も検出する。合成と非 dryrun の終了コード判定は
テスト側で `runApplyAll` と同じ形に組み立てており、`runApplyAll` 自体の配線は検証範囲外
（stage 1 の配線は [apply_all_build_test.go — 並列 build の並列度上限・失敗集約・--jobs の値域・--debug prefix・集約の同一性](apply-all-build-test.md)、コマンド全体は
[e2e 09-apply-all — apply --all の辞書順集約・並列度非依存・前段衝突 error・--no-wait 併用](../integration/e2e-09-apply-all.md) の e2e が担う）。

- **非 dryrun の組合せ** — stage 1 build 失敗・stage 2 の item 起因失敗（部分 Result 付き）・
  stage 2 の subject 起因失敗（Result 無し）・成功・try-lock skip の 31 通りで、
  applied / skipped / failures が種別の数どおりで、失敗があるときだけ exit 1（skip のみは 0）、
  失敗 config ごとに stderr の `layat: apply <name> failed:` が出ること。build 失敗の config
  で stage 2 が呼ばれないこと
- **dryrun の組合せ** — stage 1 build 失敗・read-only apply 失敗・conflict・成功の 15 通りで、
  終了コードが error(1) > conflict(2) > 0 の優先度に従うこと（conflict が error を隠さない）
- **エンベロープの層** — 各組合せのエンベロープが適合の確認を通り、`results[]` が辞書順、
  集約 status は失敗があるときだけ error、トップレベル `errors[]` は無いこと。build 失敗・
  subject 起因の失敗は subject の `errors[]` に `E_LAYAT_FAILED` が 1 件、item 起因の失敗は
  `errors[]` 無しで failed item（`E_LAYAT_FAILED`）、conflict は failed item
  （`E_LAYAT_COLLISION`）、成功と skip は success で `errors[]` 無しであること

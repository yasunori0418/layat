---
id: "DSG-836aa5cb-0389-4adb-990b-144fe5aeffe3"
type: design
name: "engine は nix を使わない tmpdir 統合テストで検証し、古くなった配置の除去の安全不変条件を table-driven に重点配分する"
status: stable
satisfies:
  - "REQ-16aef46b-7bb8-4ca1-b962-e9f3ed1fd1d2"
  - "REQ-b74a118a-1272-44eb-944c-7725163211c6"
  - "REQ-5e75aabc-0e8f-4a6c-92bd-a712dc68a940"
  - "REQ-d2277c7a-7992-49af-a9dc-4cc73843a6f9"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:22:37+09:00"
---
# engine は nix を使わない tmpdir 統合テストで検証し、古くなった配置の除去の安全不変条件を table-driven に重点配分する

## 設計

engine（`internal/`）は Go ユニットテスト + tmpdir 統合テストで検証する。
統合テストは実 FS を使うが、source は偽物・nix は使わない。

重点は **古くなった配置を保守的に除去するときの安全不変条件（誤削除防止）を table-driven で網羅する**ことに置く。

この形を選ぶ理由。

- **実 FS を使うが nix を使わない**という中間の位置取りが、engine の責務範囲と
  一致する。[engine は第三者依存ゼロの stdlib-only とし内部層に閉じる](../requirements/engine-stdlib-only.md) の通り engine は stdlib-only で、その仕事は
  `manifest.json` を入力に FS を操作することに尽きる。nix を実際に動かさなくても
  engine の全処理を実行できる（実 nix の一気通貫は [E2E は tests/e2e の bash ハーネスに置き、シナリオの詳細は同ディレクトリの README へ委譲する](e2e-harness-shape.md) の E2E が担う）。
  FS をモックしないのは FS の意味論を実行する層（`internal/engine/` の symlink 生成・
  unlink・rename・copy）に限った方針で、symlink・パーミッション・親ディレクトリと
  いった検証対象がモックでは再現しきれないため。分類判定に閉じる層
  （`internal/planner/`）はこの対象外で、fake FS を差し込んだ table-driven で覆う。
  planner が判定するのは「前世代の記録 × 現状の実体」の分類であって FS の意味論
  そのものではないため、プローブ（Lstat / Readlink / ReadDir）を差し替えても
  検証対象は失われない（→ [配置ロジックを plan を計算する planner と FS 意味論を実行する engine に分け、dryrun と apply が同一の plan を共有する](planner-engine-split.md) の層分割・[エンジンとコマンド層の Go テストは nix を介さない実 FS 統合テストを主戦力とする](../test-plan/go-test-layering.md) が定める「FS に触れずに
  判断する層」のユニットレベル被覆）
- **古くなった配置の除去に重点を置く**のは、そこに不具合が出たときの被害が非対称だからである。
  [古くなった配置の除去は前世代の記録通りを指す symlink のみに限り、copy は消さず孤児を警告する](../requirements/stale-removal-invariants.md) という不変条件が破れるとユーザーの実ファイルが
  消える。配置側の失敗は「置かれない」で済むが、除去側の失敗は復旧できない
- **table-driven にする**のは、この不変条件が「前世代の記録 × 現状の実体」の
  組み合わせに対する判定表そのものだからである。記録あり/なし、実体が
  symlink/通常ファイル/ディレクトリ/不在、リンク先が記録通り/別物 — といった
  組み合わせを網羅する形式として、ケース表がそのままテストになる

同じ tmpdir 統合テストが、[途中失敗した apply / rollback はインメモリ undo ジャーナルで全 FS 変更を巻き戻す](../requirements/undo-journal-rollback.md) と
[copy は target 不在のときだけマテリアライズする place-once で世代管理の対象外とする](../requirements/copy-place-once.md) という 2 つの要求の検証も担う。
どちらも「途中で失敗させる」「target を先に作っておく」といった FS 状態の作り込みが
必要で、実 FS 上でないと組めない。

## 出典

`docs/design.md`「テスト戦略」のテーブル 2 行目。

---
id: "TC-539ce047-3ec3-4fb2-a4d1-9f83a4f75101"
type: test_condition
name: "アグリゲータの名前衝突の確認が衝突を落とし、衝突元を全件示す"
status: stable
mitigates:
  - "RISK-3de9753f-3fd3-4364-b1ab-64c68c15ec77"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:53:44+09:00"
---
# TC-539ce047-3ec3-4fb2-a4d1-9f83a4f75101: アグリゲータの名前衝突の確認が衝突を落とし衝突元を全件示す

## テスト条件

nix-unit のスイートを組み立てるマージ（`tests/nix-unit-lib.nix` の `mergeTests`）が、
[評価テストは nix-unit で不変条件を、namaka で manifest 全体のスナップショットを見る](../../test-plan/nix-unit-namaka-split.md) の定めるテスト名の一意性を機械的に強制していることをアサートする。合成した
`{ file, tests }` のリストを渡し、readDir が返す実ファイル構成に依存せず throw 側の処理を
実行する。

- 衝突が無ければマージ結果に全ファイルのテスト名が揃うこと。名前だけでなく値も写ること
  （名前を集めてアサートの中身を落とす実装にならない担保）
- 入力が 0 件・1 件なら、1 ファイルがテストを何件持っていても衝突扱いしないこと
  （同名を定義しうるのは 2 ファイル以上からで、ここが衝突判定の境界の下側になる）
- 2 ファイルが同名を定義したら throw すること。報告に衝突したテスト名と両方のファイル名が
  出ること（後勝ちで消えた側を追えなければ、落ちても直し先が分からない）
- 3 ファイル以上が同名なら owners が 3 件とも列挙されること（2 件で打ち切らない）
- 衝突するテスト名が複数組あれば報告が組ごとの行になること（最初の 1 組で止めない）
- 衝突していないテスト名は報告に出ないこと（無関係なファイルを巻き込んで報告しない）

**この条件が見るのは検出器そのもの**であり、`tests/nix-unit/` の実ファイルが現に衝突して
いないことは見ない。後者はスイート全体の評価が担う（衝突があれば `checks.nix-unit` が
評価時に失敗する）。実ファイルに衝突が無い限り throw 側の処理は一度も実行されないため、
検出器が常に「衝突なし」を返す退行はダミー入力でしか検知できない（→ Issue #308）。

射程は `mergeTests` の入出力に限る。`tests/nix-unit.nix` が `builtins.readDir` で
`tests/nix-unit/*.nix` を列挙して `{ file, tests }` を組み立てる側は見ない。列挙が leaf を
取りこぼしても本条件は緑のままで、この残余は [評価スイート自身の盲点でアサートが空虚に緑になる](../../risks/eval-suite-blind-spot.md) が引き続き持つ。

報告文字列の比較は nix-unit の `expectedError.msg`（正規表現マッチ）で行うため、
[nix-unit: gates.nix — 評価時に throw する拒否条件群](nix-unit-gates.md) と同じく「メッセージ全体を固定していない」形の弱さを残す。owners の列挙と
行数は行そのものをパターンに書いて見るので、件数の打ち切りは検知する。

## 覆う CASE

- [nix-unit: aggregator-merge.nix — 衝突の確認をダミー入力で検証する](nix-unit-aggregator-merge.md)（`tests/nix-unit/aggregator-merge.nix`）

## 出典

検出器自身を検証するテストが無いという検出（→ Issue #308。検出元は Issue #287 のレビュー
収束）。検出器が緩和する脅威は [評価スイート自身の盲点でアサートが空虚に緑になる](../../risks/eval-suite-blind-spot.md)、一意性を要求する規約は [評価テストは nix-unit で不変条件を、namaka で manifest 全体のスナップショットを見る](../../test-plan/nix-unit-namaka-split.md) が持つ。

---
id: "RISK-3de9753f-3fd3-4364-b1ab-64c68c15ec77"
type: risk
name: "評価スイート自身の盲点でアサートが空虚に緑になる"
status: stable
likelihood: low
impact: high
level: medium
threatens:
  - "DSG-fb49e36c-eb20-4efa-8cac-88ef0873db8e"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:28:26+09:00"
---
# RISK-3de9753f-3fd3-4364-b1ab-64c68c15ec77: 評価スイート自身の盲点でアサートが空虚に緑になる

## リスク

この設計（[lib は nix-unit の評価テストと namaka のスナップショットの 2 手段で検証する](../design/lib-test-nix-unit-namaka.md)）には、nix-unit（点の不変条件）と namaka（面のスナップショット）の組み合わせに固有の盲点が 2 つある。どちらも「スイートが緑なのに
守られていない」という形をとり、他のどのアサートの信頼性も同時に下げる。

**テスト名の後勝ち衝突**: nix-unit のスイートはテストディレクトリの列挙で組み立てられ、
`//` マージが衝突を後勝ちでエラーも警告も出さずに解決する。テスト名がファイル横断で衝突すると、落ちるはずの
アサートが実行されないまま緑になる。落ちないだけでなく、実行されていないこと自体が出力に
現れない。

**スナップショットへの委ね過ぎ**: namaka のスナップショットは変更時に丸ごと再承認される。
規範である性質をスナップショットだけに委ねると、「変わったこと」は差分として見えても
「変わってよかったのか」は問われないまま承認が通る。

**この破れが design に張る理由**: どちらも 2 手段併用という検証設計の選択に由来し、別の設計
（例: 生成した文書を Go 側の契約テストで直接検証する）へ置き換えれば消える。
`docs/agents/sara-graph.md` の判別規約に従い design 側へ張る。

## 評価

- likelihood: low — 判定を支配するテスト名の衝突側が機械的な確認で覆われた。leaf ファイルを
  追加するたびに起こりうる衝突を、アグリゲータのマージ前の重複の確認が評価時に必ずエラーで止める。
  スナップショットへの委ね過ぎは [manifest 文書全体をスナップショットに固定し、どの不変条件も名指ししない変更を差分として出す](../test/manifest-eval/manifest-document-regression.md) が覆うが、こちらは名前付きアサートとスナップショットの分担を決める規約であって
  機械的な確認ではない（→「対処」節）。当初は衝突側に機械的な手立てが無く `high` と
  判定していたが、確認の追加で自動検証の見逃し幅が消えた
- impact: high — 自動検証の形骸化なので、規約の継承則に従い「その自動検証が検出するはずの
  regression の impact」を継ぐ。テスト名の後勝ち衝突は nix-unit スイート全体に影響するため、
  継承元は 2 本ある。評価層が出す manifest の形（[manifest.json の形が engine の読む契約から気付かれずにずれる](manifest-contract-drift.md)・medium）と、`gates.nix` と
  `escapes-base.nix` が担う拒否の条件が守るパス安全性の破れ（[評価時に止めるべき入力が止まらずに engine 実行時まで到達する](eval-gate-bypass.md)・high）である。
  継承元が複数あるときは最も重いものを継ぐという規約に従い high を採る

  スイートが開発内に閉じることを理由に low へは落とさない。失われるのは自動検証が与えていた保護で、
  その値打ちは守っていた対象と同じである。空虚に緑になったアサートは成功として現れるので、
  気づく契機を持たないことは変わらない

## 対処

スナップショットへの委ね過ぎは [manifest 文書全体をスナップショットに固定し、どの不変条件も名指ししない変更を差分として出す](../test/manifest-eval/manifest-document-regression.md) が緩和する（同 TC が、規範である性質は名前付き
アサート側で守りスナップショットへ委ねない旨を条件として持つ）。

**テスト名の後勝ち衝突はアグリゲータ自身が緩和する**（→ Issue #287）。アグリゲータ
（確認を含むマージの実体は `tests/nix-unit-lib.nix` の `mergeTests`）は `//` マージの前に
ファイル横断の重複キーを確認し、衝突があれば衝突したテスト名と定義元ファイル名を示して
throw する。一意性は [評価テストは nix-unit で不変条件を、namaka で manifest 全体のスナップショットを見る](../test-plan/nix-unit-namaka-split.md) が定めるスイート組み立ての規約だが、
その遵守はレビューに任せず、評価時に機械的に強制される。各 leaf ファイルがファイル固有の
接頭辞（`testEscapesBase*` / `testAnchorName*` / `testResolveMarker*` / `testFarmEntries*` など）を
付ける慣行はそのまま衝突を避ける第一手として残り、`escapes-base.nix` のヘッダコメントが
その意図を明記している。

**その確認自身も [アグリゲータの名前衝突の確認が衝突を落とし、衝突元を全件示す](../test/manifest-eval/aggregator-collision-detection.md) が緩和する**（→ Issue #308）。実ファイルに衝突が無い限り確認の
throw 側は一度も実行されず、常に「衝突なし」を返す退行が本 risk と同じ形（緑なのに守られて
いない）で再現するため、確認の処理を `tests/nix-unit-lib.nix` の純関数へ切り出してダミー入力で
アサートする。列挙（`builtins.readDir` で `{ file, tests }` を組み立てる側）は同 TC の射程外で、
列挙が leaf を取りこぼす形の破れは残余として本 risk が引き続き持つ。

同種の「テスト自身の検出力の限界」は他にも残る。[nix-unit: gates.nix — 評価時に throw する拒否条件群](../test/manifest-eval/nix-unit-gates.md) の `expectedError.msg` は部分
一致で条件の取り違えを検出できず、本 risk と同じ「緑なのに守られていない」形の残余である。

[GC アンカー対象の抽出条件と farm の組み立て結果をアサートする](../test/manifest-eval/farm-anchor-derivation.md) の組み立て検証も、現状は `lib/manifest.nix` の生成式を複製した独立コードなので
本体の変更に追随しない。ただしこちらは Issue #289 が `anchorLines` の生成式を `__internal`
から参照できる形に切り出し、CASE 対象のテストを共有式参照へ再構成する予定で、完了すれば追随する
ようになる（同 issue の完了時に本節を更新する）。

## 出典

`tests/nix-unit.nix` のアグリゲータ実装（当時は `lib.foldl' (acc: m: acc // m)` の後勝ちマージのみ
で、重複の確認は Issue #287 で追加した）と `tests/namaka/` の運用からの逆算（→ Issue #273
「L1〜L4」節）。2 手段併用の設計判断は [lib は nix-unit の評価テストと namaka のスナップショットの 2 手段で検証する](../design/lib-test-nix-unit-namaka.md)、その規約は [評価テストは nix-unit で不変条件を、namaka で manifest 全体のスナップショットを見る](../test-plan/nix-unit-namaka-split.md) が持つ。

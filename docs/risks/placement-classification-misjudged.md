---
id: "RISK-e8449214-7794-4d0d-b584-a3a152e2c1f9"
type: risk
name: "配置分類を誤り、利用者の実体を本人に知らせずに破損させる"
status: stable
threatens:
  - "REQ-622787dc-4512-4ce9-9c7d-7b32bbb70557"
  - "REQ-053cfed2-265a-4997-a27b-97b0bed10d8a"
  - "REQ-95e97d01-5c34-44b3-bc85-9ca53472bc3d"
  - "REQ-9dc7dac7-4e9e-494a-a17b-73853c119653"
likelihood: medium
impact: high
level: high
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T13:36:37+09:00"
---
# RISK-e8449214-7794-4d0d-b584-a3a152e2c1f9: 配置分類を誤り、利用者の実体を本人に知らせずに破損させる

engine は target の現況を「不在 / 自プロファイルが前世代で記録した symlink / 記録に無い
foreign symlink / 通常ファイル・実ディレクトリ」に分類し、それぞれ新規配置・無警告の張替え・
警告付き後勝ち・conflict 停止へ振り分ける。この分類を誤ると、ユーザーが自分で置いたファイルを
警告なく上書きする（回復不能）、あるいは自分が前世代で張ったリンクにまで foreign 警告を出して
毎回のノイズにする、という両方向の失敗が起きる。

conflict 検出時に「最初の 1 件で止めて残りを報告しない」ことも同じ脅威の一部として扱う。
ユーザーは apply → 1 件修正 → 再 apply → 次の 1 件、という往復を強いられ、実質的に
「上書きせず停止する」保証が使いものにならなくなる。

## 想定する失敗

- 通常ファイルが占有する target を symlink で上書きし、ユーザーのファイルを消す
- 自己記録の symlink を foreign と誤判定し、正常な再 apply で毎回警告を出す
- foreign symlink を自己記録と誤判定し、他ツールの配置物を無警告で置き換える
- conflict が複数あるとき 1 件しか報告せず、対処ガイダンスも欠く
- out-of-store のリンク先が不在のまま配置し、dangling symlink を残す（また、確認自体が
  失敗しただけの状況をリンク先不在と誤って報告する）
- 配置元の実在を判定できる層で確認せず、store path の `src` の不在が評価を通るか、out-of-store
  のリンク先や `src` 内の `subpath` の不在が engine を止めない（→ [配置元の実在は判定できる層で確認し、いずれの層でも停止する](../requirements/src-subpath-existence-checks.md)）

## 評価

- likelihood: medium — 分類は [target の現況に応じた配置分類（新規 / 自己記録の張替え / foreign 後勝ち）が仕様どおりであること](../test/engine-core/placement-kinds.md) / [実体が占有する target と祖先 symlink を上書きせず停止すること](../test/engine-core/no-overwrite-stop.md) / [out-of-store の配置前の確認がリンク先不在と確認の失敗を区別すること](../test/engine-core/out-of-store-preflight.md) / [conflict が全件・種類に応じたガイダンス付きで列挙され、集約エラー 1 本で返ること](../test/engine-core/conflict-full-report.md) の
  4 本が対象・停止・確認・報告の各面を覆っており密度は高い。ただし分類そのものが engine の
  中心にあり、配置の形（method・祖先 symlink・out-of-store）を足すたびに枝が増える
- impact: high — 通常ファイルを symlink で上書きする方向の誤りは、ユーザーのファイルを消す。
  失われるのは layat の記録に無い実体なので再 apply でも `reset` でも戻らない。foreign を
  自己記録と誤判定する方向も他ツールの配置物を置き換える形で同じ結果になる。どちらも警告なく
  起きるため、上書きされたこと自体に気づく契機も無い

## 張り先の判断

4 本とも `requirement` へ張る。いずれも「target を上書きしない」「foreign は警告して後勝ち」
「conflict は全件列挙する」「配置元の不在で停止する」というプロダクトの観測可能な振る舞い
そのものへの懸念であり、配置アルゴリズムを別設計へ差し替えても懸念は残る（→ `docs/agents/sara-graph.md`）。

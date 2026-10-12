---
id: "ADR-0059"
type: adr
name: "root 相対 out-of-store マーカー"
status: 採用
issues:
  - "#460"
  - "#362"
origin: "Issue #362（epic: src 側のプロジェクトルート実行時解決）の grilling（2026-10-12）。PR #361 で dev/skills を `.claude/skills/` へ可変 symlink で置く需要が出たこと"
justifies:
  - "REQ-a77181dd-a59c-40bb-855e-3430ad970eb6"
  - "DSG-fd39b9ee-a5d2-46c2-bdcd-fc399a1e0296"
  - "REQ-81249072-56b8-42f5-807b-ea623c1efe99"
  - "REQ-0b0cd1e3-bfeb-45c1-978d-e2e11c568336"
  - "REQ-1dcc9a33-b0f2-43e0-8310-fc4b19e68fe7"
  - "REQ-a8a923ad-07fb-4582-b90a-07a6e0c41baa"
  - "REQ-99ca5381-6c53-426c-b145-7b4297c53868"
revises:
  - "ADR-0001"
  - "ADR-0013"
references:
  - "ADR-0003"
  - "ADR-0015"
  - "ADR-0046"
  - "ADR-0058"
---
# ADR-0059: root 相対 out-of-store マーカー

- ステータス: 採用
- 日付: 2026-10-12
- 関連: ADR-0003（link 生成は engine が担う）, ADR-0015（祖先 symlink error・devShell 同梱で CLI と lib の版ずれが起きない）, ADR-0046（自己記録の祖先 symlink は配置前に除去）, ADR-0058（last-applied 記録は manifest をそのまま保持）, #460, #362
- 改訂対象: ADR-0001 の `src` の型 `path | set | marker` のうち marker を、`mkOutOfStoreSymlink`（絶対パス）と `mkRootRelativeSymlink`（配置 root 相対）の 2 構築子へ拡張する（out-of-store を明示関数でのみ opt-in する決定は不変）。ADR-0013 §1 の `srcKind` 列挙に `"rootRelative"` を加え（`schemaVersion = 1` は不変）、§4 の copy + out-of-store marker の eval 時エラーを root 相対マーカーにも適用する
- 起点: Issue #362（epic: src 側のプロジェクトルート実行時解決）の grilling（2026-10-12）。PR #361 で dev/skills を `.claude/skills/` へ可変 symlink で置く需要が出たこと

## 背景

PR #361 で、リポジトリ内の実体（`dev/skills/<名前>/`）を同じリポジトリの `.claude/skills/<名前>` へ可変 symlink で置く需要が出た。`mkOutOfStoreSymlink` の引数は Nix 評価時に確定する絶対パスで、pure eval の flake は自身のチェックアウトの絶対パスを得られない。そのため devShell の shellHook に素の `ln -sfn` ループを書いて代替している。

engine 側の事実として、planner の前世代照合・stale 除去・drift 修復・`--json` の `old` / `new` は、どれもディスク上の readlink 結果と `planner.LinkDest(entry)` の文字列完全一致で判定する。`LinkDest` は Entry のフィールドだけから計算する純関数で、root を受け取らない。manifest の `validate()` は `schemaVersion` と root だけを見て、`srcKind` / `method` / `rootKind` の値は検査していない。

## 決定

### 1. リンクは target 起点の相対パスで張る

engine が張る symlink の中身は `filepath.Rel(dir(target), src/subpath)` の相対パスにする。target `.claude/skills/sara-docs`・src `dev/skills/sara-docs` なら `../../dev/skills/sara-docs` になる。

### 2. 基準点はその manifest の配置 root

src は engine が実行時に解決した配置 root（project / home / fixed。`--root` 上書きを含む）に対する相対パスとする。target と同じ root なので、src 側で root の種別を選ぶ口は作らない。全 root 種別で使える（system mode は未実装のまま）。

### 3. 祖先 symlink の不変条件に依存する

target の親が symlink だと、相対リンクは symlink の先を起点に解決されて別の場所を指す。layat は foreign な祖先 symlink を配置前に conflict として止め（ADR-0015 §4）、自己記録の stale な祖先 symlink は配置前に除去する（ADR-0046）。配置の時点で target の祖先は常に実ディレクトリなので、相対リンクは意図した場所を指す。この不変条件を緩める変更は root 相対マーカーの正しさを壊す。

### 4. API は別関数 `lib.mkRootRelativeSymlink`

`lib.mkRootRelativeSymlink "dev/skills/sara-docs"` は `_layatMarker = "rootRelative"` のマーカーを返す純粋関数とする。`mkOutOfStoreSymlink` は絶対パス専用のまま据え置く。

### 5. manifest は `srcKind: "rootRelative"` を足し、`schemaVersion` は 1 のまま

`srcKind` に `"rootRelative"` を加え、`src` には root 相対文字列をそのまま書く。`schemaVersion` は 1 に据え置く。代わりに engine の `validate()` が entry ごとに `srcKind` / `method` / `rootKind` の値を検査し、未知の値を明示エラーにする。既存の manifest は全て既知値なので挙動は変わらない。

### 6. 評価時ゲートは 3 つ

`normalizeManifest` が次を拒否する。

1. 絶対パスと `..` による root 外への脱出（target / subpath と同じ `pathChecks.isUnsafe`）
2. `method = "copy"` との併用（outOfStore と同じ理由。ライブ symlink と place-once の意図が矛盾する）
3. `src/subpath` と `target` の一致、または一方が他方の祖先であること（symlink ループの防止）

link farm の GC アンカーは `srcKind == "store"` の entry だけを対象にするので、root 相対 entry は自然に対象外になる。

### 7. engine で root を要るのは preflight だけ

`LinkDest` は root を受け取らない純関数のままにする。preflight の out-of-store 実在検査に、`<root>/<src>/<subpath>` を Lstat して不在なら dangling を作らず停止する分岐を足す。copy 経路は決定 6 で併用を拒否するので触らない。

### 8. 要求と設計の置き場

要求は REQ-a77181dd-a59c-40bb-855e-3430ad970eb6、相対 `LinkDest` の計算と root の扱いは DSG-fd39b9ee-a5d2-46c2-bdcd-fc399a1e0296 に置く。src の時点（REQ-81249072-56b8-42f5-807b-ea623c1efe99）・manifest の `srcKind` 列挙（REQ-0b0cd1e3-bfeb-45c1-978d-e2e11c568336）・マーカーの判別（REQ-1dcc9a33-b0f2-43e0-8310-fc4b19e68fe7）・リンク先と版管理（REQ-a8a923ad-07fb-4582-b90a-07a6e0c41baa）・src の種別（REQ-99ca5381-6c53-426c-b145-7b4297c53868）を root 相対の形を含むよう改める。

## 根拠

- **相対リンクなら照合経路に手を入れずに済む**。`LinkDest` が root 不要の純関数のままなので、planner・stale 除去・drift 修復・`--json` は相対の文字列をそのまま比べる。last-applied 記録は manifest をそのまま保持する（ADR-0058）ので、前世代の root 相対 entry からも同じ文字列を再計算できる。
- **worktree の移動・改名に耐える**。リンクの中身が root の絶対パスを含まないので、リポジトリを動かしてもリンクは有効なまま。home mode では `~/dotfiles/...` を HOME 非依存に書けるので、絶対パスの OS 別ホーム慣例の差を考えずに済む。
- **基準点を target と同じ root に固定すれば、src 側の root 選択は構造的に要らない**。相対リンクは target からの位置関係しか持たないので、別の root を基準にした src は表現できない。
- **`schemaVersion` を上げなくても版ずれは起きない**。CLI と lib は devShell に同梱され（ADR-0015）、古い engine が新しい manifest を読む経路は構造的に無い。値検査を足せば、仮に経路ができても未知の `srcKind` は黙って誤配置せずに明示エラーで止まる。

## 影響

- lib（マーカー・正規化・評価時ゲート 3 つ・nix-unit）は Issue #461、engine（`srcKind` 定数と値の検査・相対 `LinkDest`・preflight・Go テスト）は Issue #462、dogfood（`dev/layat.nix` の宣言化・shellHook の整理・e2e・利用者向け文書）は Issue #463。
- ADR-0001 / ADR-0013 に改訂注記を足す。
- 要求: REQ-a77181dd-a59c-40bb-855e-3430ad970eb6 を起票し、REQ-81249072-56b8-42f5-807b-ea623c1efe99 / REQ-0b0cd1e3-bfeb-45c1-978d-e2e11c568336 / REQ-1dcc9a33-b0f2-43e0-8310-fc4b19e68fe7 / REQ-a8a923ad-07fb-4582-b90a-07a6e0c41baa / REQ-99ca5381-6c53-426c-b145-7b4297c53868 を改訂する。REQ-eb363122-385a-499c-a074-c95efb949d07（`mkOutOfStoreSymlink` は絶対パス）は据え置く。設計は DSG-fd39b9ee-a5d2-46c2-bdcd-fc399a1e0296。

## 棄却した代替案

- **絶対パスで張る**（engine が `<root>/<src>/<subpath>` を絶対パスにして張る）: `LinkDest` が root を要るようになり、planner・stale 除去・drift 修復・`--json` の照合経路を全て root 付きに書き換えることになる。worktree を動かすとリンクが切れ、home mode では HOME 依存の文字列が残る。
- **`mkOutOfStoreSymlink` の overload**（相対文字列を受理したら root 相対とみなす）: 絶対パスの書き忘れ（`"dotfiles/nvim"`）が黙って root 相対へ化ける。型ベースの暗黙分岐を廃した ADR-0001 と同じ問題を引数の形で持ち込む。
- **src 側で root 種別を選ぶ形**（`{ root = projectRoot; path = "..."; }`）: target と別の root を基準にした src は相対リンクで表現できず、絶対パスで張る案に戻る。基準点を target の root に固定すれば選択肢自体が要らない。
- **`schemaVersion` を 2 に上げる**: CLI と lib が同梱で版ずれが起きない以上、全利用者の manifest を版上げする代償に見合う保護が無い。未知値の明示エラーは `validate()` の値検査で得られる。

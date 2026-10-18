---
id: "ADR-0062"
type: adr
name: "system mode の last-applied 記録を `/var/lib/layat` に置く"
status: 採用
issues:
  - "#140"
  - "#129"
origin: "Issue #140（epic #129: system mode の実装）の計画（2026-10-18）。ADR-0036 の起票時には last-applied 記録（ADR-0058）が無く、system mode の記録の置き場が未決だった"
justifies:
  - "REQ-0d1f557b-7ae7-41ea-b17b-cdb5c80935d4"
revises:
  - "ADR-0036"
  - "ADR-0058"
references:
  - "ADR-0025"
  - "ADR-0057"
---
# ADR-0062: system mode の last-applied 記録を `/var/lib/layat` に置く

- ステータス: 採用
- 日付: 2026-10-18
- 関連: ADR-0025, ADR-0057, GitHub Issue #140, #129
- 改訂対象: ADR-0036 §3 の system mode の状態の置き場に、last-applied 記録の基底 `/var/lib/layat` を足す（profile の基底 `/nix/var/nix/profiles/layat` と鍵は不変）。ADR-0058 §1 の記録の置き場 `<state>/layat/<鍵>` を、system mode に限って `/var/lib/layat/<鍵>` へ切り替える（鍵の規則・記録の内容は不変）
- 起点: Issue #140（epic #129: system mode の実装）の計画（2026-10-18）。ADR-0036 の起票時には last-applied 記録（ADR-0058）が無く、system mode の記録の置き場が未決だった

## 背景

ADR-0036 §3 は system mode の profile 状態を `/nix/var/nix/profiles/layat` に置くと決めた。その後
ADR-0058 が、stale 除去の baseline として layat 自前の last-applied 記録 `<state>/layat/<鍵>` を
足した。`<state>` は実行ユーザーの state（`$XDG_STATE_HOME`、既定 `$HOME/.local/state`）なので、
このままでは root で実行した system mode の記録が `/root/.local/state/layat` に置かれる。

ADR-0036 は profile について「システムの状態が root の個人領域に住む」形を棄却している。記録も
同じ理由で個人領域に置けないが、記録は nix profile ではないため、profile と同じ基底に置く理由も無い。

## 決定

### 1. system mode の記録基底は `/var/lib/layat`

- `rootKind = "system"` のとき、last-applied 記録を `/var/lib/layat/<鍵>/last-applied.json` に置く。
- 鍵は profileDir と同型にする。`--root` なしは `<name>` 直キー、`--root` ありは `<roothash>/<name>` と
  `<roothash>` 階層の backref `.root`（ADR-0058 §1 の規則を system mode にそのまま当てる）。
- home / project / fixed の記録の置き場（`<state>/layat/<鍵>`）は変えない。記録の内容・書き込みの
  手順・baseline の優先順位（ADR-0058 §2〜§4）も system mode で変わらない。

### 2. 基底の上書きは engine の入力とテスト用の環境変数に限る

- engine は system mode の 2 つの基底を入力で上書きできる。profile 基底は `SystemDir`（空なら
  世代バックエンドの `SystemBase`）、記録基底の `<state>` は `SystemStateDir`（空なら `/var/lib`）。
- CLI はこの 2 つを環境変数 `LAYAT_SYSTEM_PROFILE_BASE` / `LAYAT_SYSTEM_STATE_DIR` から渡す口を、
  全コマンド共通のテスト用として持つ（配線は CLI 側の変更で行う）。どちらもサポート対象外で、
  設定ファイルのキーにはしない。空文字は上書きと見なさない。
- 用途は、root 権限を user namespace で得る e2e が、実システムの基底を汚さずに配置と世代を検証する
  ことに限る。

## 根拠

- **`/var/lib` を選ぶ理由**: 記録は layat が自分で読み書きする可変の状態で、FHS の
  `/var/lib/<パッケージ>` がその置き場にあたる。nix profile ではなく GC root でもないので、
  nix の慣習に合わせる必要が無い。
- **profile の基底（`/nix/var/nix/profiles/layat`）に同居させない理由**: その基底は世代バックエンドの
  所有物（ADR-0057）で、記録は世代と独立した layat の所有物（ADR-0058）。user mode でも記録
  （`<state>/layat`）と profile（`<state>/nix/profiles/layat`）は別の基底にある。同居させると
  `layat prune` の走査対象に記録のディレクトリが入る。
- **基底を設定で上書きさせない理由**: 基底を変えると既存の世代と記録を見失う。ADR-0057 は profile の
  基底を設定ファイルで上書きする機構を足さないと決めており、記録の基底もそれに揃える。テストの
  隔離に要るのは環境変数の口だけ。

## 影響

- **`internal/engine` / `internal/paths`**: 記録の解決が system mode のとき `SystemStateDir`
  （既定 `/var/lib`）を基底にする。`Options` / `ResetOptions` / `RollbackOptions` に `SystemDir` /
  `SystemStateDir` を足す。
- **`cmd/layat`**: 環境変数 2 つの読み出しと、全コマンドから engine への受け渡し（別の変更で行う）。
- **ADR-0036 / ADR-0058**: 改訂注記を同じ変更で足す。

## 棄却した代替案

- **実行ユーザーの state（root なら `/root/.local/state/layat`）のまま**: 実装差分は無いが、システムの
  状態が root の個人領域に置かれる。ADR-0036 が profile について棄却した形と同じで、将来のモジュール
  導入時に記録を移す痛みが残る。
- **`/nix/var/nix/profiles/layat` の下に同居させる**: 置き場の数は減るが、世代バックエンドの所有する
  基底へ layat 自前の状態を持ち込み、prune の走査と衝突する。
- **基底を設定ファイルのキーで上書きできるようにする**: 利用者の需要が無く、上書きのたびに
  既存の世代と記録を見失うおそれが生じる。

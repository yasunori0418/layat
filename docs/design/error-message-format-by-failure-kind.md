---
id: "DSG-a45f92c5-89de-487f-8c93-60d3b509bad0"
type: design
name: "エラーの文字列は失敗の種類ごとの書式で書き、括弧の対象は原因が対象を持たないときだけ添える"
description: "エラーの文字列は I/O と parse・子プロセス・検証・前提の不成立・入力の誤り・sentinel の種類ごとの書式で書き、対象を括弧で添えるのは原因のエラーが対象を持たないときと原因に無い情報を足すときだけにする"
section: core-logic
status: stable
satisfies:
  - "REQ-9341fa5d-836e-4023-af53-cc7d273438d1"
  - "REQ-7a2f1ecf-4675-45aa-80c0-a8fc58db9edd"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-14T22:59:25+09:00"
---

# エラーの文字列は失敗の種類ごとの書式で書き、括弧の対象は原因が対象を持たないときだけ添える

## 設計

layat の Go のコードがエラーの文字列を作るときは、失敗の種類ごとに下の書式を使う。どこで包むかは
[エラーは op と対象が分かるいちばん下の関数で 1 回だけ wrap し、伝搬では素通しする](error-wrap-once-at-origin.md) が決める。文字列は
`layat:` を持たない。`layat:` は stderr へ書く場所で付ける
（→ [stderr へ書く行は cmd の 1 箇所で `layat:` を付け、単発の失敗・`--all` の失敗行・status の行・警告が同じ整形を通る](stderr-lines-prefixed-in-one-place.md)）。

### 失敗の種類ごとの書式

| 種類 | 書式 | 例 |
|---|---|---|
| I/O と parse の失敗 | `cannot <op>[ (<target>)]: %w` | `cannot read manifest: open /state/abc/manifest.json: no such file or directory` |
| 子プロセスの失敗 | `cannot run <コマンド> (<target>): %w` | `cannot run nix-env --set (/nix/var/nix/profiles/per-user/u/layat-x): exit status 1` |
| 検証の失敗 | `invalid <検証したもの> (<target>): <詳細>` | `invalid manifest.json (/state/abc/manifest.json): <詳細>` |
| 前提が成り立たない（元エラーなし） | 成り立たない事柄を書き、対象を括弧で添える | `no previous generation (profile: <profile>)` |
| 入力の誤り | 1 行で書く。書式は決めない | — |
| sentinel | 対象も案内文も持たない短い句 | `locked by another process`、`outside a git repository` |

- **I/O と parse の失敗**: 括弧の対象を書くのは、原因のエラーが対象を持たないとき（子プロセスの終了・
  flock・parse・検証）と、原因に無い情報を足すとき（copy の src と dst の組を
  `cannot copy (<src> -> <dst>): %w` と書くとき）だけである。原因が同じパスを出す `os.*` の失敗では書かない。
  どの呼び出しが原因にパスを出すかは、下の「パスを持つエラーを返す呼び出し」の一覧で決める
- **`failed to` の形は使わない**: 失敗は `cannot <op>` で書く。`failed to <op>` や `<op> failed` の形は使わない
- **sentinel**: 対象は包む側が付け、案内文は cmd の表示が足す。sentinel 自身は `%w` で透過させ、
  `errors.Is` で判定する

### パスを持つエラーを返す呼び出し

layat の Go のコードが使う標準ライブラリの呼び出しを、原因のエラーにパスが出るかで分ける。
`*fs.PathError` と `*os.LinkError` は文字列に op とパスを持つ。

| 原因のエラー | 呼び出し | 括弧の対象 |
|---|---|---|
| `*fs.PathError`（渡したパスを出す） | `os.Open`・`os.OpenFile`・`os.Stat`・`os.Lstat`・`os.ReadFile`・`os.WriteFile`・`os.ReadDir`・`os.Mkdir`・`os.Remove`・`os.Chmod`・`os.Readlink`・`os.CreateTemp`、`*os.File` のメソッド（`Read`・`Write`・`Close` など。`io.Copy` が返すものを含む）、`filepath.WalkDir` のコールバックに渡るエラー | 書かない |
| `*fs.PathError`（祖先や子孫のパスを出すことがある） | `os.MkdirAll`・`os.RemoveAll` | 書かない（失敗した場所が原因に出る） |
| `*os.LinkError`（2 つのパスを出す） | `os.Rename`・`os.Symlink` | 書かない |
| 渡したパスと違うパスを出す | `filepath.EvalSymlinks`（リンク先や途中の要素の `Lstat` の失敗を返し、リンクの段数の上限ではパスを持たない） | 書く |
| パスを持たない | `*exec.ExitError`（子プロセスの終了）、`syscall.Flock`、`encoding/json` の decode と `json.Marshal`、`filepath.Abs`、`filepath.Rel`（文字列に 2 つのパスが入るが、どちらが対象かを示さない） | 書く |

`planner` の `FS` のように、`os` の同名の関数をそのまま呼ぶ型を通した呼び出しは、`os` の関数と
同じに扱う。

### 警告の文

警告に原因のエラーを埋め込むときは、警告の頭で原因の op や対象を繰り返さず、何を飛ばして続行するかだけを
書く（例: `skipping stale removal: %v`）。原因の op と対象は埋め込んだエラーが出す。

## 出典

[エラーの文字列は `cannot <op>` を保ち、原因が対象を持たないときだけ対象を括弧で添える](../adr/error-messages-formatted-by-failure-kind.md) の決定を、
失敗の種類ごとの書式として書き下した。

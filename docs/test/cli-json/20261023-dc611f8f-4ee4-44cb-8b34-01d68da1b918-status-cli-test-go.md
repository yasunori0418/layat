---
id: "CASE-dc611f8f-4ee4-44cb-8b34-01d68da1b918"
type: test_case
name: "status_test.go — status の subject 名・フィルタ・一致 0 件・ストリームの排他・--json の形・error 時の続行"
target: "cmd/layat/status_test.go"
covers:
  - "TC-1bc3e39a-116a-4f26-96b4-1d15e6f837bd"
---
# CASE-dc611f8f-4ee4-44cb-8b34-01d68da1b918: status_test.go — status の subject 名・フィルタ・一致 0 件・ストリームの排他・--json の形・error 時の続行

## 対象

`cmd/layat/status_test.go`

FS も nix も使わない CLI 層のテスト。engine を呼ぶ変数 `statusFn` を差し替え、直キー・
`<roothash>` 系列・backref が読めない系列・系列全体で 1 件として返る系列を並べた結果を
`runStatus` に渡す。エンベロープは `outturn_test.go` の `newStatusTestRun` で組み、stdout /
stderr は `captureOutErr` で取る。

## 検証内容

**subject 名**

- 直キーは `<mode>:<name>`、`<roothash>` 系列は `<mode>:<name>@<root>`、backref が読めない系列は
  `<mode>:<name>@<roothash>`、系列全体で 1 件として返る系列は `<mode>:@<roothash>`

**system 基底の受け渡し**

- `LAYAT_SYSTEM_STATE_DIR` の値がそのまま `StatusOptions.SystemStateDir` に渡り、未設定なら空で
  渡る。`StateDir` は空のまま engine に任せる

**テキスト出力と error 時の続行**

- subject ごとに `# <subject> (applied <appliedAt>)` と `<target>\t<method>\t<dest>` の行を出し、
  subject の間を空行 1 行で区切る。dest は `planner.LinkDest` の値（store の src と subpath の連結、
  root 相対の src は target からの相対パス）
- error を持つ subject は stdout に出さず、subject 名付きで stderr に出す。他の subject は
  stdout に出たまま、exit code 1 の error が返る

**フィルタ**

- `[name]` は `<name>` 部分の完全一致で mode・root を問わずに絞る。前方一致にも、subject 名
  全体（`home:docs` など）との一致にもならない
- 系列全体で 1 件として返る系列の error subject は `[name]` を指定しても残る。一致する config が
  無くてもこの subject が出て、一致 0 件の通知は出ず、exit 1 になる
- 別の `<name>` の error subject は絞り込みで落ち、残った subject が全て読めれば exit 0

**一致 0 件**

- テキストでも `--json` でも stdout は空で、stderr に通知を 1 行出して exit 0。通知は `[name]`
  指定時が `layat: status: no applied config named <name>`、記録が 1 件も無いときが
  `layat: status: no applied config`。`--json` のエンベロープは `status: success`・`results: []`
  で conformance を通る

**engine の失敗**

- engine の `Status` 自体が失敗する（user の記録基底が読めない）と、subject を作る前に error を
  返す。`--json` ではトップレベルの `errors[]` に 1 件載り、エンベロープの status は error

**`--json` の形とストリームの排他**

- `--json` のとき `runStatus` は stdout に何も書かない
- `results[i].result.info` が `{mode, root, appliedAt, linkFarm, entries}` で、entries は
  `{target, method, srcKind, src, subpath}`。entry の無い manifest でも `entries: []`
- `items: []`、`generation` スロット無し、`dryRun: false`
- error を持つ subject は `status: error` で `errors[]` に 1 件を持ち、エンベロープの status は
  error。エンベロープは conformance を通る

全コマンドの run が `outturnReport` に載ることは、`outturn_test.go` の
`TestBeginRunPublishesEveryCommand` が status も含めて確かめる。

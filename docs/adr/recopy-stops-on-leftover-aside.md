---
id: "ADR-e99a7cc3-b7e8-4028-a299-f81fcef74997"
type: adr
name: "recopy の退避先が既にあれば上書きせず error で止める"
description: "apply --recopy は既存の copy target を退避する前に退避先を確かめ、前の run が残した退避物があれば上書きせず entry の失敗で止める"
section: removal-migration
status: stable
sources:
  - id: "issue-677"
    resource: "https://github.com/yasunori0418/layat/issues/677"
    title: "#677 Go のエラー処理を規約に揃える"
origin: "Issue #677 の grilling（2026-11-14）。commit の失敗の後に残った recopy の退避物を、次の --recopy が上書きするか毎回失敗する欠陥（#787）への対処として決めた"
justifies:
  - "REQ-b4e4b65d-6e35-40c3-a00e-20c14043df6f"
references:
  - "ADR-f29d1256-5e14-4b83-ac77-4717b2210193"
  - "ADR-37191429-2099-4217-a552-47b16a7e504c"
  - "ADR-56325270-20bb-4959-8498-e3223c35e24b"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-15T01:25:42+09:00"
---
# recopy の退避先が既にあれば上書きせず error で止める

- ステータス: stable
- 日付: 2026-11-15
- 関連: [undo の発火を FS 反映の段に限り、世代コミットの後は対象にしない](undo-scope-ends-at-commit.md), [退避先が既にあれば conflict で止める](backup-existing-destination-conflict.md), [途中失敗した apply をインメモリ undo ジャーナルで巻き戻す](in-memory-undo-journal.md), GitHub Issue #677, #787
- 起点: Issue #677 の grilling（2026-11-14）。commit の失敗の後に残った recopy の退避物を、次の --recopy が上書きするか毎回失敗する欠陥（#787）への対処として決めた

## 背景

`apply --recopy` は、既存の copy target を同じ親ディレクトリの `<target>.layat-recopy-aside` へ rename で
退避してから再コピーする。退避物は run が成功すると消え、途中の段が失敗すると元へ戻る
（[途中失敗した apply をインメモリ undo ジャーナルで巻き戻す](in-memory-undo-journal.md)）。
世代の commit が失敗したときは巻き戻さない（[undo の発火を FS 反映の段に限り、世代コミットの後は対象にしない](undo-scope-ends-at-commit.md)）ので、
退避物はディスクに残る。journal はメモリにしか無いので、run の途中でプロセスが落ちたときも残る。

次の `--recopy` は退避先を確かめずに rename していた。target がファイルなら、残っていた退避物を
黙って上書きし、前の run が退避した元の内容が消える。target がディレクトリで退避物が空でない
ディレクトリなら、rename が失敗し、退避物を手で消すまで `--recopy` は毎回失敗していた。

## 決定

`--recopy` は既存の target を退避する前に退避先を `lstat` で確かめる。退避先が既にあれば、
上書きも削除もせず、その entry の失敗として止める。エラーの文は退避先のパスを示し、退避物を
移すか消してから再実行するよう案内する。commit の失敗で巻き戻さない動作は変えない。

この規範は [recopy の上書きは削除ではなく同一親内への rename 退避で行う](../requirements/recopy-rename-aside.md) に置く。

## 根拠

残っている退避物は、前の run が上書きする前の内容である。利用者に知らせずに上書きすれば、
元の内容を取り戻す手段が消える。`--backup` も、退避先が既にあれば上書きせずに止める
（[退避先が既にあれば conflict で止める](backup-existing-destination-conflict.md)）。recopy の退避物も同じく、繰り返して通ることより残った内容を守ることを優先する。

止める場所は entry の失敗とし、planner の conflict にはしない。recopy は planner を通らずに
実行の段で target を調べるので、実行の段で失敗させるのがその流れに合う。同じ run で先に再コピーした
entry は、ほかの entry の失敗と同じく undo で戻る。

## 棄却した代替案

### 残っている退避物を消してから退避する

`--recopy` は毎回通るようになるが、前の run が退避した元の内容が利用者に知らされずに消える。

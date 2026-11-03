---
id: "ADR-a86a8492-57fb-49d0-ba77-200f6dc49c3a"
type: adr
name: "reset --dryrun は副作用なしで撤去対象を示す"
status: 採用
issues:
  - "#584"
origin: "ADR-0021 §2（reset の --dryrun を副作用なしのプレビューにする）を書き直した"
justifies:
  - "REQ-6a950d6d-c191-4235-a1b4-73ffc7c2bb38"
supersedes:
  - "ADR-0021"
references:
  - "ADR-ddeafc2b-f5a9-4332-a0b9-b18bea6ce6f7"
  - "ADR-acc31876-df36-485e-b356-92ae60e26db7"
  - "ADR-f607b605-c69e-4c7e-ac89-75b7bce5bee1"
---
# ADR-a86a8492-57fb-49d0-ba77-200f6dc49c3a: reset --dryrun は副作用なしで撤去対象を示す

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-ddeafc2b-f5a9-4332-a0b9-b18bea6ce6f7, ADR-acc31876-df36-485e-b356-92ae60e26db7, ADR-f607b605-c69e-4c7e-ac89-75b7bce5bee1
- 起点: ADR-0021 §2（reset の --dryrun を副作用なしのプレビューにする）を書き直した

## 背景

reset は copy target を消すので、実行するとデータを失うことがある（ADR-ddeafc2b-f5a9-4332-a0b9-b18bea6ce6f7）。
削除前の確認は撤去対象を示すが、対話が前提で、CI やスクリプトからは使えない
（ADR-acc31876-df36-485e-b356-92ae60e26db7）。

apply はすでに `--dryrun` を持っている。

## 決定

reset に `--dryrun` を設ける。`--dryrun` は撤去する symlink と copy target を示して終わり、
FS の削除・確認・flock のどれも行わない。撤去対象があってもなくても、正常に終える。

規範は REQ-6a950d6d-c191-4235-a1b4-73ffc7c2bb38 に置く。終了コードの体系は
ADR-f607b605-c69e-4c7e-ac89-75b7bce5bee1 に従う。

## 根拠

非対話で「何が消えるか」を前もって確かめられれば、copy を失う前の安全な下見になる。CI でも使える。

`--dryrun` は FS を変えないので、並行する操作と直列化する必要が無い。flock を取らないのは
apply の `--dryrun` と同じ扱いである。

撤去対象があることは予告であって失敗ではない。対象の有無で終了コードを変えると、スクリプトから
見て失敗と区別できなくなる。

## 棄却した代替案

### `--dryrun` を設けず、確認の一覧で代える

対話でなら下見ができるが、非対話では事前に確かめられない。apply とも対称にならない。

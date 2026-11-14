---
id: "ADR-a86a8492-57fb-49d0-ba77-200f6dc49c3a"
type: adr
name: "reset --dryrun は副作用なしで撤去対象を示す"
section: placement
status: stable
sources:
  - id: "issue-584"
    resource: "https://github.com/yasunori0418/layat/issues/584"
    title: "#584"
origin: "ADR-0021 §2（reset の --dryrun を副作用なしのプレビューにする）を書き直した"
justifies:
  - "REQ-6a950d6d-c191-4235-a1b4-73ffc7c2bb38"
references:
  - "ADR-ddeafc2b-f5a9-4332-a0b9-b18bea6ce6f7"
  - "ADR-acc31876-df36-485e-b356-92ae60e26db7"
  - "ADR-f607b605-c69e-4c7e-ac89-75b7bce5bee1"
  - "ADR-45380022-d959-4b22-aa05-5993ac1f3c87"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# reset --dryrun は副作用なしで撤去対象を示す

- ステータス: stable
- 日付: 2026-11-03
- 関連: [reset で profile を触らずに配置物を撤去する](reset-fs-only-teardown.md), [reset は非対話の実行で --yes を必須にする](reset-confirmation-and-yes.md), [終了コードは 0 / 1 / 2 の 3 値にする](exit-code-scheme.md), [flock は明示実行で待ち、hook 実行でスキップする](flock-mode-explicit-wait-hook-skip.md)
- 起点: ADR-0021 §2（reset の --dryrun を副作用なしのプレビューにする）を書き直した

## 背景

reset は copy target を消すので、実行するとデータを失うことがある（[reset で profile を触らずに配置物を撤去する](reset-fs-only-teardown.md)）。
削除前の確認は撤去対象を示すが、対話が前提で、CI やスクリプトからは使えない
（[reset は非対話の実行で --yes を必須にする](reset-confirmation-and-yes.md)）。

apply はすでに `--dryrun` を持っている。

## 決定

reset に `--dryrun` を設ける。`--dryrun` は撤去する symlink と copy target を示して終わり、
FS の削除・確認・flock のどれも行わない。撤去対象があってもなくても、正常に終える。

規範は [reset --dryrun は副作用ゼロで削除対象を表示して終了する](../requirements/reset-dryrun-no-side-effect.md) に置く。終了コードの体系は
[終了コードは 0 / 1 / 2 の 3 値にする](exit-code-scheme.md) に従う。

## 根拠

非対話で「何が消えるか」を前もって確かめられれば、copy を失う前の安全な下見になる。CI でも使える。

`--dryrun` は FS を変えないので、並行する操作と直列化する必要が無い。flock を取らないのは
apply の `--dryrun` と同じ扱いである。

撤去対象があることは予告であって失敗ではない。対象の有無で終了コードを変えると、スクリプトから
見て失敗と区別できなくなる。

## 棄却した代替案

### `--dryrun` を設けず、確認の一覧で代える

対話でなら下見ができるが、非対話では事前に確かめられない。apply とも対称にならない。

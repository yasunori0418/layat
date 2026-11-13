---
id: "REQ-a47fbfb0-c808-42e9-b2ad-767c933494d6"
type: requirement
name: "prune --json はトップレベル info の removed に削除系列を、skipped に判定か削除に至れず残した系列を載せる"
section: cli-json
status: stable
derives_from:
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  Because `layat prune` names no config, its `--json` envelope SHALL keep `results` an
  empty array and SHALL carry the series inventory in the top-level `info` as
  `{removed, skipped}`. `removed` SHALL list the series deleted — under `--dryrun`, the
  series that would be deleted — and `skipped` SHALL list the series left alone because
  their verdict or deletion could not be reached; a series kept because its root exists
  SHALL NOT appear in either, and both SHALL be present as arrays even when empty. Each series
  SHALL be represented as `{roothash, root, names, dir}`: `roothash` the series directory
  name, `root` the absolute root path recorded by the backref `.root` (an empty string for a
  series skipped as `backref-unreadable`), `names` the `<name>` profileDirs under the series
  as an array (an empty array, never null, when there are none or when they could not be
  listed, as for `series-unreadable`), and `dir` the absolute series directory, which tells
  the two scan bases apart. Each element of
  `skipped` SHALL be `{series, reason, detail}`: `series` a series in the form above,
  `reason` one of `locked`, `backref-unreadable`, `series-unreadable`, `root-stat-failed`,
  and `permission-denied`, emitted verbatim from the engine and SHALL NOT be respelled by
  the CLI, and `detail` the message of the underlying failure. The key `pruned` SHALL NOT
  be used for deleted series. A run that fails before the scan, or one whose deletion was
  declined, SHALL NOT emit `info`.
specification_ja: |
  `layat prune` は config を名指ししないため、`--json` のエンベロープは `results` を空配列の
  ままとし、系列のインベントリをトップレベルの `info` に `{removed, skipped}` として載せなければ
  ならない。`removed` には削除した系列（`--dryrun` では削除予定の系列）を、`skipped` には判定か
  削除に至れずに残した系列を載せなければならず、root が実在するために残した系列はどちらにも
  載せてはならない。どちらも 0 件のときも配列として出さなければならない。各系列は
  `{roothash, root, names, dir}` で表さなければならない。`roothash` は系列ディレクトリ名、
  `root` は backref `.root` が記録する root の絶対パス（`backref-unreadable` で残した系列では
  空文字列）、`names` は系列配下の `<name>` profileDir の配列（無いとき、および
  `series-unreadable` のように列挙できなかったときは null ではなく空配列）、`dir` は系列
  ディレクトリの絶対パスで、2 つの走査基底のどちら由来かを区別する。`skipped` の各要素は `{series, reason, detail}` で
  なければならない。`series` は上の形の系列、`reason` は `locked`・`backref-unreadable`・
  `series-unreadable`・`root-stat-failed`・`permission-denied` のいずれかで engine の語彙を
  そのまま出し、CLI が綴り直してはならない。`detail` は根底の失敗のメッセージである。
  削除した系列に `pruned` というキーを使ってはならない。走査より前に失敗した実行と、削除の
  同意が得られなかった実行は `info` を出してはならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-01T21:09:38+09:00"
---
# prune --json はトップレベル info の removed に削除系列を、skipped に判定か削除に至れず残した系列を載せる

## 仕様

```json
{"removed": [{"roothash": "…", "root": "/abs/root", "names": ["docs"], "dir": "/abs/base/<roothash>"}],
 "skipped": [{"series": {"roothash": "…", "root": "…", "names": ["docs"], "dir": "…"},
              "reason": "locked", "detail": "…"}]}
```

- **トップレベルの `info`** — prune は主体（config）を持たないので、`init` と同じくインベントリを
  エンベロープ全体の `info` に載せる（→ [読み取り系の JSON ペイロードは dryRun パリティと info インベントリで表す](json-readonly-payload.md)）
- **`dir`** — 同じ `<roothash>` が state 基底と system 基底の両方に立ちうるので、由来を示せる
  のは `dir` だけ（→ [prune は backref の root が実在しない roothash 系列だけを系列ごと削除する](prune-target-selection.md)）
- **`reason` の語彙** — skip 理由の意味は [prune は backref の root が実在しない roothash 系列だけを系列ごと削除する](prune-target-selection.md) と
  [prune は dryrun・root 一覧付き確認・try-lock skip・--json の --yes 必須で削除を守る](prune-safety-gates.md) が定める
- **`pruned` を使わない** — `pruned` は空祖先ディレクトリの剪定で既に使われている
  （→ [target 除去後は空の親ディレクトリチェーンを root 境界まで保守的に剪定する](empty-parent-dir-pruning.md)）
- **`info` を出さない実行** — 走査前の失敗で空のインベントリを出すと「走査して 0 件だった」と
  読めてしまう。同意が得られなかった実行は何も削除していない

削除途中の失敗で、失敗より前に削除し終えた系列を `removed` に残すことと、`--json` の
`--yes` 必須は [prune は dryrun・root 一覧付き確認・try-lock skip・--json の --yes 必須で削除を守る](prune-safety-gates.md) の担当。

## 出典

`cmd/layat/prune.go` の `pruneInfo` / `pruneSeriesRow` / `pruneSkippedRow` が実装している
payload の形を、規範として起こした。

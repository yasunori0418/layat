# Issue tracker: GitHub

Issues and PRDs for this repo live as GitHub issues. Use the `gh` CLI for all operations.

## Filing an issue in layat

- **From the Web UI**, file through one of the three issue forms: バグ報告 (bug report),
  機能要望 (feature request), or ドキュメント改善 (docs improvement). Blank issues are
  disabled. Each form applies its type label (`bug` / `enhancement` / `documentation` —
  the same labels the Epics and Tasks sections call topical) and `needs-triage`
  automatically.
- **From `gh`**, the forms do not apply. Whoever files the issue applies the type label
  and a triage label themselves (see `triage-labels.md`).

## Epics

An epic bundles sub-issues under a milestone-level tracking issue (#147). Epics are
plain issues — there is no separate issue type — distinguished by convention:

- **Title prefix** `epic:` (e.g. `epic: sara によるドキュメントのグラフ構造化`).
- **Label** `epic`, plus a topical label (`enhancement` / `documentation`).
- **Registered** in the tracking issue's epic table. That table is the index; the label
  makes the same set queryable with `gh issue list --label epic --state all`.

Apply `epic` when creating an issue that the tracking issue will list as an epic. Keep it
on the issue after it closes — the label is how completed epics stay discoverable.

An epic row in the tracking table may exist before its issue does (the table can name a
planned epic with "issue は未起票"). Such a row has nothing to label until it is filed.

Epics may nest: an epic can be filed under another epic rather than directly under the
tracking issue (#283 under #203, for example). Label it `epic` all the same — depth is
expressed by which issue lists it, not by the label.

## Tasks

A task is a sub-issue under an epic or tracking issue — the unit actual work and PRs are
scoped to. Label it `task`.

- **Apply** `task` when decomposing an epic into sub-issues. This is required, not
  optional: `label:task` is how the work units stay enumerable.
- **Do not apply** it to a standalone issue that belongs to no epic (a one-off bug report
  or docs fix). Those carry only topical and triage labels.
- **Keep** it after the issue closes, same as `epic`.

`tracking` / `epic` / `task` are mutually exclusive — an issue is at most one of them.

PRs reference both levels: `Closes #<task>` + `Refs #<epic>`.

## Defect issues

Defects found in the product are managed here, not in the document graph — there is no
`defect` item type (→ [欠陥はグラフに置かず GitHub Issues で管理する](../sara-adr/defects-in-github-issues.md)). The graph holds norms (what should hold); a defect is an
event, and events belong to the tracker.

- **File** a defect as a GitHub issue with the `bug` label.
- **Fill in** the body template below. Maintainers (including agents) fill it in; the
  reporter is not asked to. The two headings are fixed; other sections are free-form.
  References are one-way — items in `docs/` never link back to defect issues.

  ```markdown
  ## トレース

  - 露見させたテストケース: [name](docs/test/<区分>/<slug>.md)
    （テストの実行で見つかった欠陥でなければ「なし」と書く）
  - 遡る先: [name](docs/requirements/<slug>.md) または [name](docs/design/<slug>.md)

  ## feed back

  クローズ前に書く。

  - 既存の RISK / TC: [name](docs/risks/<slug>.md) ／ なし
  - 追補した item: [name](docs/…) ／ なし（理由）
  ```

- **Trace** at triage: always write the requirement or design the defect traces back to.
  Write the revealing test case only when the defect was found by executing a test;
  otherwise write 「なし」.
- **Feed back** before closing, in the `feed back` section: ask whether a `risk` or
  `test_condition` was missing that would have caught this defect earlier. If so, add the
  missing item(s) in `docs/` — this shift-left loop is what keeps the risk and test
  analysis from decaying into a static snapshot.
- **Classify** at close: apply exactly one of `risk-expected` / `risk-unexpected`
  (meanings in `triage-labels.md`). Judge by the state when the issue was filed, not the
  state at close. An issue closed as not a defect (`wontfix`, duplicate) gets neither.

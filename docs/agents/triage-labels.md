# Triage Labels

The five triage labels in this repo's issue tracker.

| Label             | Meaning                                  |
| ----------------- | ---------------------------------------- |
| `needs-triage`    | Maintainer needs to evaluate this issue  |
| `needs-info`      | Waiting on reporter for more information |
| `ready-for-agent` | Fully specified, ready for an AFK agent  |
| `ready-for-human` | Requires human implementation            |
| `wontfix`         | Will not be actioned                     |

## Structural labels (not triage roles)

These are orthogonal to the triage roles above — they describe what an issue *is*, not
what it needs next. An issue carries both (e.g. `epic` + `needs-triage`).

| Label      | Color     | Meaning                                                     |
| ---------- | --------- | ----------------------------------------------------------- |
| `tracking` | `#5319e7` | Milestone-level issue aggregating epics (e.g. #147)         |
| `epic`     | `#6f42c1` | Bundles sub-issues under a tracking issue                   |
| `task`     | `#0e8a16` | A sub-issue under an epic or tracking issue — the work unit |

The three are mutually exclusive: an issue is at most one of them. A standalone issue
(a one-off bug or docs fix belonging to no epic) carries none.

All three are kept on closed issues, so `--state all` queries list the project's full
history: `label:epic` for every epic it has had, `label:task` for every work unit. See
`issue-tracker.md` for when to apply them.

## Type labels

What kind of issue it is, carried alongside a triage label. The issue forms apply it
automatically; when filing with `gh`, apply it yourself (see `issue-tracker.md`).

| Label           | Meaning                                          |
| --------------- | ------------------------------------------------ |
| `bug`           | A defect in the product                          |
| `documentation` | An error, inconsistency, or gap in the docs      |
| `enhancement`   | A new feature or an extension of an existing one |

## Defect classification labels

Applied to a `bug` issue when it is closed, to count how many defects the risk analysis
had foreseen.

| Label             | Meaning                                                                                   |
| ----------------- | ----------------------------------------------------------------------------------------- |
| `risk-expected`   | A `risk` or `test_condition` matching the defect already existed when the issue was filed |
| `risk-unexpected` | None existed when the issue was filed, whether or not the feed back added one afterwards  |

- Apply exactly one of the two at close. Judge by the state when the issue was filed.
- An issue closed as not a defect (`wontfix`, duplicate) gets neither.
- To judge, compare the issue's creation date with the date the matching `risk` or
  `test_condition` file was added (`git log --follow --diff-filter=A -- <path>`). When the feed back revised
  an existing `risk` / `test_condition` rather than adding one, the added date alone would
  wrongly say `risk-expected`; also check the item's history after the issue was filed.

```bash
gh issue list --label bug --label risk-unexpected --state closed --limit 1000 --json number --jq length
```

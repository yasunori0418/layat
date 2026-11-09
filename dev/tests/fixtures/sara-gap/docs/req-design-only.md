---
id: "REQ-dddd0004-0000-4000-8000-00000000000c"
type: requirement
name: "design 側にだけ risk がある要求"
status: stable
derives_from:
  - "UC-0ca11001-0000-4000-8000-000000000002"
specification: |
  The fixture SHALL keep risks only on the design that satisfies this requirement.
specification_ja: |
  fixture はこの要求を満たす design にだけ risk がある状態を保たなければならない。
---
# REQ-dddd0004: design 側にだけ risk がある要求

要求に直接 `threatens` する risk を持たない。`unthreatened` に `(design 側に risk あり)` を付けて出ることを固定する。

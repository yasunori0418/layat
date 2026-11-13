---
id: "REQ-cccc0003-0000-4000-8000-00000000000b"
type: requirement
name: "risk の対象外とした要求"
section: architecture
status: stable
derives_from:
  - "UC-0ca11001-0000-4000-8000-000000000002"
specification: |
  The fixture SHALL keep this requirement exempt from risk identification.
specification_ja: |
  fixture はこの要求を risk の識別の対象外に保たなければならない。
risk_exempt: "fixture の免除の分岐を確かめるための要求で、テストで確かめる対象を持たない。"
---
# REQ-cccc0003: risk の対象外とした要求

`unthreatened` に出ず `exempt` に出ることを固定する。到達率の分母からも外れる。

---
id: "CASE-66660002-0000-4000-8000-00000000000f"
type: test_case
name: "follows を持つテストケース"
status: stable
resource: "tests/fixture/followed_test.go"
covers:
  - "TC-33330001-0000-4000-8000-000000000008"
follows:
  - "TP-77770001-0000-4000-8000-00000000000e"
---
# CASE-66660002: follows を持つテストケース

`unfollowed` に出ないことを固定する。

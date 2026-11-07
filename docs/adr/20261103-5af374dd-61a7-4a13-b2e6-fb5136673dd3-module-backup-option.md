---
id: "ADR-5af374dd-61a7-4a13-b2e6-fb5136673dd3"
type: adr
name: "モジュールに layat.backup を起動配線のオプションとして持つ"
status: 採用
issues:
  - "#169"
  - "#172"
origin: "旧 epic #167 から統合された epic #172 の grilling セッション（2026-07-11・2026-07-12）で確定。ADR-0045 §8 を書き直した"
justifies:
  - "REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10"
  - "REQ-e1e1114b-ba07-4d57-8e04-6e30e39a5da3"
references:
  - "ADR-24488788-a5e6-47a5-9074-97192880642a"
  - "ADR-f58e9981-bacc-4937-98b8-bddce0a057db"
  - "ADR-2aaf72ab-1dfc-4cde-852d-4b1f9916780b"
  - "ADR-36bfbfbd-f1dd-4581-a75a-943d80f6ccf5"
  - "ADR-01abcd5a-464f-4c44-8afd-30950631487c"
  - "ADR-9506276e-2108-4962-9f90-efe82fec7f66"
  - "ADR-e8bb6153-2528-419e-8be0-c862a3eb4e05"
  - "ADR-5dad12a8-ada3-4a04-a3cf-1795a5b44c98"
  - "ADR-33d8c466-6b38-46b6-8696-ad7521c72786"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T23:00:56+09:00"
---
# ADR-5af374dd-61a7-4a13-b2e6-fb5136673dd3: モジュールに layat.backup を起動配線のオプションとして持つ

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-24488788-a5e6-47a5-9074-97192880642a, ADR-f58e9981-bacc-4937-98b8-bddce0a057db, ADR-2aaf72ab-1dfc-4cde-852d-4b1f9916780b, ADR-36bfbfbd-f1dd-4581-a75a-943d80f6ccf5, ADR-01abcd5a-464f-4c44-8afd-30950631487c, ADR-9506276e-2108-4962-9f90-efe82fec7f66, ADR-e8bb6153-2528-419e-8be0-c862a3eb4e05, ADR-5dad12a8-ada3-4a04-a3cf-1795a5b44c98, ADR-33d8c466-6b38-46b6-8696-ad7521c72786, GitHub Issue #169, #172
- 起点: 旧 epic #167 から統合された epic #172 の grilling セッション（2026-07-11・2026-07-12）で確定。ADR-0045 §8 を書き直した

## 背景

モジュールから使う利用者は、activation が起動する `layat apply` の引数を直接書かない。
`apply --backup`（ADR-33d8c466-6b38-46b6-8696-ad7521c72786）をモジュールの利用者にも使わせる
手段を決める必要があった。

## 決定

全モジュール共通のオプションとして `layat.backup`（`enable` と `suffix`）を持つ。`enable` が
true のとき、activation の `layat apply` 起動へ `--backup=<suffix>` を付ける。これは起動の配線
だけの変更で、manifest の契約にも entry の定義にも触れない。

オプションの型と既定値は REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10 に、起動への配線は
REQ-e1e1114b-ba07-4d57-8e04-6e30e39a5da3 に置く。

## 根拠

`--backup` は配置のロジックの一部ではなく、起動時の修飾である。モジュールは engine の CLI
引数を組み立てるだけで、engine の判定には関わらない。build 済みの link farm を `--manifest` で
渡す設計（ADR-24488788-a5e6-47a5-9074-97192880642a）と同じ分担になる。

CLI の `--backup` が apply 全体の修飾なので、モジュールのオプションも全体に 1 つとする。

## 棄却した代替案

### entry ごとに `backup` のオプションを持つ（`layat.entries.<key>.backup`）

entry ごとに退避の可否を変えられるが、実際の要求は apply 単位の一括である。使われない粒度の
ために manifest の契約へ項目を足すことになる。

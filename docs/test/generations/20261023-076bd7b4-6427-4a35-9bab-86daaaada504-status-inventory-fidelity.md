---
id: "TC-076bd7b4-6427-4a35-9bab-86daaaada504"
type: test_condition
name: "status が記録基底の全系列を漏れなく列挙し、記録の内容をそのまま反映し、状態を変えない"
mitigates:
  - "RISK-e92487a5-fff9-40db-844d-6590f1f36dec"
  - "RISK-f87d4909-6dea-4ead-8e72-994c498a14df"
---
# TC-076bd7b4-6427-4a35-9bab-86daaaada504: status が記録基底の全系列を漏れなく列挙し、記録の内容をそのまま反映し、状態を変えない

## テスト条件

engine の `Status` が記録基底を走査した結果について、網羅・内容・並び順と、読み取り専用で
あることを検証する。

**網羅と並び順**

- user 基底と system 基底の両方の、`<name>` 直キーと `<roothash>` 系列を全て返す
- 並びは user 基底 → system 基底、基底内は直キー → `<roothash>` 系列、それぞれ名前の昇順で、
  実行ごとに変わらない
- 基底が無ければその基底を空として扱い、error にしない
- system 基底が読めなければ warning を 1 行出してその基底を飛ばし、user 基底の結果は返す

**内容の反映**

- 直キーの Mode は基底に応じた `home` / `system`、Root は空
- `<roothash>` 系列の Mode は記録の manifest の `rootKind`、Root は backref `.root` の値
- `AppliedAt`・`LinkFarm`・manifest の entry は記録の値と一致する
- 不正な記録（壊れた JSON・未知のフィールド・`recordVersion` ≠ 1 など）と読めない backref は、
  その系列に error を立てて他の系列と並べ、列挙を止めない。「記録なし」として落とさない

**読み取り専用**

- 実行の前後で基底配下のエントリ一覧が増減しない（flock・`.pending`・`.root`・記録を作らない）

上位の規範は TP-e7c25263-6d2d-4a37-8275-26906889d912（`internal/engine/` を実 FS の tmpdir で駆動する統合レベル）。

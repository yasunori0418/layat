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
- 並びは user 基底 → system 基底、基底内は直キー（`<name>` の昇順）→ `<roothash>` 系列
  （`<roothash>`、続けて `<name>` の昇順）で、実行ごとに変わらない
- `last-applied.json` を持たない `<name>` は列挙しない
- `<name>` を列挙できない系列の 1 件は、その `<roothash>` の位置に並ぶ
- 基底が無ければその基底を空として扱い、error にしない
- system 基底が読めなければ warning を 1 行出してその基底を飛ばし、user 基底の結果は返す
- user 基底が存在するのに読めなければ error を返す

**内容の反映**

- 直キーの Mode は基底に応じた `home` / `system`、Root は空
- `<roothash>` 系列の Mode は記録の manifest の `rootKind`、Root は backref `.root` の値
- `AppliedAt`・`LinkFarm`・manifest の entry は記録の値と一致する
- 不正な記録（壊れた JSON・未知のフィールド・`recordVersion` ≠ 1 など）と読めない backref は、
  その系列に error を立てて他の系列と並べ、列挙を止めない。「記録なし」として落とさない
- `<roothash>` 系列の `<name>` で `last-applied.json` の stat が不在以外の理由で失敗したもの
  （`<name>` ディレクトリに検索権限が無いなど）は、error を立てて返し、落とさない
- backref が読めない系列は、記録を持つ `<name>` ごとに error を立て、Root は空、RootHash は系列の
  ディレクトリ名、Mode は基底に応じた `home` / `system` で返す
- `<name>` を列挙できない系列は、Name と Root が空で、RootHash と error を持ち、Mode が基底に
  応じた `home` / `system` の 1 件として返す。backref も読めない系列も同じ 1 件にまとめる
- backref が読めず、記録を持つ `<name>` も見つからない系列は、Name と Root が空で error を持つ
  1 件として返す（検索権限の無い `<name>` 直キーのディレクトリがこれに当たる）
- `<roothash>` 系列として返したディレクトリ（`.root` の有無を判定できず error になったものを
  含む）を、直キーとしても重ねて列挙しない

**読み取り専用**

- 実行の前後で `StateDir` と `SystemStateDir` の配下全体（`StateDir` 側は記録基底と profile
  基底の両方）のエントリ一覧が増減しない（flock・profileDir・`.pending`・`.root`・記録を作らない）
- 実行の前後で記録ファイルの内容が変わらない（記録を書き換えない）
- profileDir の lock を別に保持した状態でも、`Status` は待たずに返る（flock を取らない）

上位の規範は TP-e7c25263-6d2d-4a37-8275-26906889d912（`internal/engine/` を実 FS の tmpdir で駆動する統合レベル）。

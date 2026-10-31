---
id: "CASE-152e6f7b-cc49-4ac0-ad8e-0aba62dc1371"
type: test_case
name: "status_test.go — 記録基底の列挙順・基底の扱い・読めない系列・読み取り専用"
target: "internal/engine/status_test.go"
covers:
  - "TC-076bd7b4-6427-4a35-9bab-86daaaada504"
---
# CASE-152e6f7b-cc49-4ac0-ad8e-0aba62dc1371: status_test.go — 記録基底の列挙順・基底の扱い・読めない系列・読み取り専用

## 対象

`internal/engine/status_test.go`

実 FS（tmpdir）を使い nix は使わない統合テスト。`StateDir` と `SystemStateDir` に tmpdir を渡し、
記録は `writeRecord` で apply と同じ鍵で書く。読めない基底・系列は chmod で作り、root 実行では飛ばす。

## 検証内容

**網羅と並び順**

- user 基底 → system 基底の順に、直キー（`<name>` の昇順）→ `<roothash>` 系列（`<roothash>`、
  続けて `<name>` の昇順）で並び、2 回実行しても同じ結果になる
- 直キーの Mode は基底に応じた `home` / `system`、`<roothash>` 系列の Mode は記録の `rootKind`、
  Root は backref の値。`AppliedAt`・`LinkFarm`・manifest の entry が記録と一致する
- `last-applied.json` を持たない `<name>` は直キー・`<roothash>` 系列のどちらでも列挙しない

**基底の扱い**

- 基底が無い（state dir が空・存在しない）ときは空を返し、warning も error も出さない
- user 基底が読めなければ error を返す
- system 基底が読めなければ基底を名指しした warning を 1 行出して飛ばし、user 基底の結果を返す

**読めない記録・系列**

- 壊れた JSON・未知のフィールド・`recordVersion` が 1 でも 2 でもない記録・空の記録は `Err` を立てて他の系列と並ぶ。
  `<roothash>` 系列では Mode が基底に応じた値、Root が backref の値になる
- backref が読めない系列は、記録を持つ `<name>` ごとに Root 空・RootHash 入りの `Err` を返す。
  記録を持つ `<name>` が無ければ、Name と Root が空の 1 件として返す
- `<name>` を列挙できない系列は Name と Root が空の 1 件として返す
- backref も `<name>` も読めない系列は、両方の理由を持つ 1 件として、その `<roothash>` の位置
  （直キーの後、他の系列との名前順）に並ぶ
- 検索権限の無い直キーのディレクトリは `<roothash>` 系列側の 1 件としてだけ返し、直キーとして重ねない
- `last-applied.json` の stat が失敗した `<name>` は落とさず `Err` 付きで返す

**読み取り専用**

- 実行の前後で `StateDir`（記録基底と profile 基底）と `SystemStateDir` 配下のエントリ一覧と
  記録ファイルの内容が変わらない
- profileDir と記録ディレクトリの flock を別に保持した状態でも、待たずに返る

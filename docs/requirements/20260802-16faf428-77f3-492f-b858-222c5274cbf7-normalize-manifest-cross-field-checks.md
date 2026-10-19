---
id: "REQ-16faf428-77f3-492f-b858-222c5274cbf7"
type: requirement
name: "意図が矛盾する組み合わせをクロスフィールドチェックで評価時に拒否する"
derives_from:
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
  - "UC-403fbe32-b146-401b-8b53-fe67c1e169c5"
  - "UC-01b896b4-04b9-40d0-bf9e-966eaf64c3d4"
specification: |
  `normalizeManifest` SHALL reject, via `lib.throwIf`, a combination that a single-field
  type check cannot express: `method = "copy"` together with an out-of-store marker.
  Out-of-store means a live symlink while copy means a place-once snapshot, so the two
  intents contradict each other and SHALL fail early at evaluation time rather than being
  silently downgraded.
specification_ja: |
  `normalizeManifest` は単一フィールドの型検査では表現できない組み合わせを
  `lib.throwIf` で拒否しなければならない。対象は `method = "copy"` かつ out-of-store
  marker の組み合わせである。out-of-store は「ライブ symlink」、copy は「place-once
  スナップショット」で意図が矛盾するため、暗黙降格ではなく評価時に早期エラーとしなければ
  ならない。
---
# REQ-16faf428-77f3-492f-b858-222c5274cbf7: 意図が矛盾する組み合わせをクロスフィールドチェックで評価時に拒否する

## 仕様

`normalizeManifest` はクロスフィールドの `lib.throwIf` チェックを持つ。

- `method = "copy"` かつ out-of-store marker の組み合わせを拒否する。

規範は frontmatter が正。`root = systemRoot`（system mode）は `rootKind = "system"` として
評価を通り、このチェックの対象ではない（→ ADR-0036）。

単一フィールドの optionType では表現できない条件のため、型検査とは別に `throwIf` で
評価時に停止させる。out-of-store は「ライブ symlink」、copy は「place-once
スナップショット」で意図が矛盾するので、暗黙に降格させず早期エラーにする（→ ADR-0013 §4）。

## 出典

`docs/spec.md`「lib API」→「入力検査（`evalModules` + `normalizeManifest`）」。

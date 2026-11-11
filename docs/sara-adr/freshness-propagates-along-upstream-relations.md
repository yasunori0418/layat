---
id: "SDR-dfff5637-f0a4-4e5d-a57e-f4dd1e7dbb42"
type: sara_adr
name: "鮮度は上流の relation を辿って下流へ伝播させる"
description: "下流の item の基準時刻より上流の generated.at が新しければ、上流の relation のうち justifies を除く 7 種を辿って下流を再検証待ちとして挙げる"
status: stable
sources:
  - id: "issue-676"
    resource: "https://github.com/yasunori0418/layat/issues/676"
    title: "epic: 文書の鮮度と信頼を機械的に確かめる — 伝播する鮮度・generated.at と履歴の一致・verified と RISK の重さ"
origin: "epic #676 の grilling（2026-11-11）で確定"
justifies:
  - "QA-611977bd-3d11-4695-9fb2-a9f7211fcfbd"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-11T23:59:24+09:00"
---
# 鮮度は上流の relation を辿って下流へ伝播させる

- ステータス: stable
- 日付: 2026-11-11
- 関連: GitHub Issue #676, `docs/model.md`
- 起点: epic #676 の grilling（2026-11-11）で確定

## 背景

下流の item は上流の item を前提に書かれる。上流の本文が変わっても、下流の本文はそのまま残る。
`requirement` を書き換えたあとに、その `requirement` を脅かす risk や、risk に対処する test_condition が
古い前提のまま残っていても、sara の構造の検査（`sara check`）はそれを見分けない。

全ての item は `generated` を持ち、`generated.at` は本文を最後に意味のある形で変えた時刻を表す。
`verified` は人などが内容を確かめた記録で、`generated` と同じ形の要素を並べる。

## 決定

下流の item の基準時刻より上流の item の `generated.at` が新しければ、その下流の item を
再検証待ちとして挙げる。

- 基準時刻はその item の `generated.at` と、`generated.at` 以降の `verified[].at` のうち最も新しい
  時刻とする。`verified` の actor は問わない。`verified` が無い item では `generated.at` が基準時刻になる
- 辿る relation は `refines` / `derives_from` / `satisfies` / `threatens` / `mitigates` / `covers` /
  `follows` の 7 種とする。relation を宣言している item が下流、指されている item が上流になる
- `justifies` と、peer の relation（`depends_on` / `supersedes` / `references`）は辿らない
- `status` が `deprecated` の item は上流としても下流としても対象から外す
- 比べるのは直接の辺だけとする。上流の上流が新しくても、間の item が古ければ、間の item が挙がる
- 挙げた結果は報告に留め、`verified` を自動では足さない

## 根拠

7 種の relation は、下流の規範が上流の規範から導かれることを表す。上流の本文が変われば、下流が
前提にした上流の規範が、もう成り立っていない可能性がある。直接の辺だけを比べるのは、間の item の本文を書き換えればその
`generated.at` が進み、その下流も挙がるためである。間の item に確認の記録だけを足したときは、
その下流は挙げない。

基準時刻に `verified` を含めるのは、上流が変わったあとで下流を読み直し、本文を変える必要が
無かった場合に、その確認を記録する手段を残すためである。`generated.at` より前の `verified` は、
その後に本文が変わっているので、いまの本文を確かめた記録にならない。

`deprecated` の item はもう有効でないので、上流として変わっても下流の前提にならず、下流として
古くても読み直す意味が無い。

## 棄却した代替案

### `justifies` も辿る

ADR は決定の理由を残す記録で、`justifies` は寄与した item を指す。ADR の本文を後から整えても、
指された item の規範は変わらない。逆に item を書き換えても、ADR を書き直す必要は無い。辿ると、
規範に関わらない書き換えのたびに、多くの item が再検証待ちに挙がる。

### `verified` が無い下流を常に再検証待ちにする

着手時点で `verified` を持つ item は 1 件も無い。上流との新旧に関わらず全ての下流が挙がり、
上流が変わったために読み直すべき item が埋もれる。人の確認の有無は、level が high の risk について
[level が high の risk に人の verified を求める](human-verified-required-for-high-risk.md) で扱う。

### `verified` が無い下流を対象から外す

`verified` を持つ item だけを比べると、着手時点では何も挙がらない。上流が変わったあとに下流を
読み直していないという状態は、`verified` が無くても `generated.at` で判定できる。

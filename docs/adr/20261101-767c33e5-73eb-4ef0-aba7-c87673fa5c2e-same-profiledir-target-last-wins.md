---
id: "ADR-767c33e5-73eb-4ef0-aba7-c87673fa5c2e"
type: adr
name: "同一 profileDir への同時実行は後勝ちにする"
status: 採用
issues:
  - "#550"
origin: "ドキュメント計画段階のレビュー（2026-06-13）で確定。ADR-0013 §2 を書き直した"
justifies:
  - "REQ-60c6b7ea-e936-4ce8-bd75-ad35e9c693b9"
supersedes:
  - "ADR-0013"
references:
  - "ADR-400cadd4-0151-470b-92ae-9b11aa8e9bed"
  - "ADR-ad1c2507-1f42-4cd3-ab56-9183892f8988"
  - "ADR-15a358f4-1c81-49d5-b658-3a1fe39f0f75"
  - "ADR-55d8ecdf-86bb-45d0-8c60-ac9c630cb62d"
  - "ADR-90964d0f-3c2e-4201-a04b-5821f2f4c799"
  - "ADR-4dca6ef4-2791-48d5-aa82-7037630968a6"
  - "ADR-777fc3c8-4ca8-49da-aaf3-146b7d991ebe"
  - "ADR-1ea2b93a-f83f-46c6-b925-cb0b38f4fd3c"
---
# ADR-767c33e5-73eb-4ef0-aba7-c87673fa5c2e: 同一 profileDir への同時実行は後勝ちにする

- ステータス: 採用
- 日付: 2026-11-01
- 関連: ADR-400cadd4-0151-470b-92ae-9b11aa8e9bed, ADR-ad1c2507-1f42-4cd3-ab56-9183892f8988, ADR-15a358f4-1c81-49d5-b658-3a1fe39f0f75, ADR-55d8ecdf-86bb-45d0-8c60-ac9c630cb62d, ADR-90964d0f-3c2e-4201-a04b-5821f2f4c799, ADR-4dca6ef4-2791-48d5-aa82-7037630968a6, ADR-777fc3c8-4ca8-49da-aaf3-146b7d991ebe, ADR-1ea2b93a-f83f-46c6-b925-cb0b38f4fd3c, GitHub Issue #550
- 起点: ドキュメント計画段階のレビュー（2026-06-13）で確定。ADR-0013 §2 を書き直した

## 背景

同じ config の apply を 2 つ同時に実行すると、同じ profileDir の世代と配置物を取り合う。flock が
apply を直列にするので途中の状態が重なることはないが、どちらの配置を最終結果にするかは決めておく
必要がある。

## 決定

同一 profileDir への同時実行は、基本的に起きない前提としてユーザーの責任に置き、起きたときは、
lock を待って取る明示 apply どうしでは後勝ちにする。後から lock を取った apply が前世代を読み直して
適用し、その結果が残る。

後勝ちの規範と、実行フロー（eval 先行・flock・build の順序）の中での位置づけは
REQ-60c6b7ea-e936-4ce8-bd75-ad35e9c693b9 に置く。shellHook から lock を待たずにスキップする場合は
REQ-1c1526b1-59e3-4264-bb7c-65a10a4aa461 が扱う。

## 根拠

flock の待機で直列になるので、後勝ちにすれば追加の仕組みなしに整合した状態へ収束する。後の apply は
前の apply が書いた世代を前世代として読むため、前世代にだけある古い配置の除去も正しく働く。

profileDir は解決済みの root でキーするので（ADR-0005）、同じプロジェクトの別クローンは別の profileDir に
なり、この衝突に巻き込まれない。出所の異なる apply が profileDir を共有したときの警告は ADR-0070 が持ち、
後勝ちという決定は変えない。

## 棄却した代替案

### 後から来た明示 apply を待たせずスキップする

先に lock を取った apply の結果が残る形になるが、明示した apply が何もせずに終わるので、ユーザーは
適用されたと誤認する。

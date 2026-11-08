---
id: "ADR-6b7becf0-f029-41e3-895e-b672dce94d3d"
type: adr
name: "CLI は cobra で組む"
status: stable
sources:
  - id: "issue-552"
    resource: "https://github.com/yasunori0418/layat/issues/552"
    title: "#552"
origin: "engine / CLI の技術スタックの確定（2026-06-13）。ADR-0011 §2 を書き直した"
justifies:
  - "REQ-637599dc-a1ec-4af5-9e97-e882c7df56d0"
references:
  - "ADR-560852e1-3d07-46b9-ae1a-d15f82d28a5e"
  - "ADR-f85f1afd-9b87-4f7a-922e-97c62554f198"
  - "ADR-400cadd4-0151-470b-92ae-9b11aa8e9bed"
  - "ADR-88231ffc-4836-4252-be34-894944fd28c3"
  - "ADR-86a77232-dc99-4c04-bd6d-2127410e8828"
  - "ADR-5caf1c14-a054-4c8d-95a2-c7f89e0c5246"
  - "ADR-e683ac05-c0e0-41c1-8b6a-38b605972fe1"
  - "ADR-28605860-770e-41df-95dd-7dc05b20497a"
  - "ADR-955c2601-88fb-4be8-85de-ec7d0a8c727d"
  - "ADR-65b3439a-b90a-4056-b55a-522399bc71e6"
  - "ADR-d15a216f-b5d6-4df9-9825-b23a341bec1d"
  - "ADR-553eea60-6ab4-4da7-93a9-3acf44b2fae3"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T23:00:56+09:00"
---
# ADR-6b7becf0-f029-41e3-895e-b672dce94d3d: CLI は cobra で組む

- ステータス: stable
- 日付: 2026-11-01
- 関連: ADR-560852e1-3d07-46b9-ae1a-d15f82d28a5e, ADR-f85f1afd-9b87-4f7a-922e-97c62554f198, ADR-400cadd4-0151-470b-92ae-9b11aa8e9bed, ADR-88231ffc-4836-4252-be34-894944fd28c3, ADR-86a77232-dc99-4c04-bd6d-2127410e8828, ADR-5caf1c14-a054-4c8d-95a2-c7f89e0c5246, ADR-e683ac05-c0e0-41c1-8b6a-38b605972fe1, ADR-28605860-770e-41df-95dd-7dc05b20497a, ADR-955c2601-88fb-4be8-85de-ec7d0a8c727d, ADR-65b3439a-b90a-4056-b55a-522399bc71e6, ADR-d15a216f-b5d6-4df9-9825-b23a341bec1d, ADR-553eea60-6ab4-4da7-93a9-3acf44b2fae3, GitHub Issue #552
- 起点: engine / CLI の技術スタックの確定（2026-06-13）。ADR-0011 §2 を書き直した

## 背景

CLI は、サブコマンド・フラグ・help・シェル補完を持つ。engine は第三者依存を持たないが、CLI 層には
人間向けの使い勝手のために最小限の第三者依存を許している。

## 決定

CLI のフレームワークに cobra を使う。

CLI の依存の許可範囲と固定の方法は REQ-637599dc-a1ec-4af5-9e97-e882c7df56d0 に置く。

## 根拠

cobra は Go のエコシステムで事実上の標準で、nixpkgs でもパッケージが整っている。サブコマンド・
フラグ・補完・help の整形が成熟しており、`--help` で内部の nix コマンドを開示する要件（ADR-5caf1c14-a054-4c8d-95a2-c7f89e0c5246）にも
help の整形の質で寄与する。

## 棄却した代替案

### stdlib の flag と手書きの dispatch で組む

依存は増えないが、help と補完を自作する負荷が高い。

### kong / urfave-cli を使う

軽量だが、nixpkgs のパッケージや補完の成熟度で cobra に劣る。

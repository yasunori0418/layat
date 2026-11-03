---
id: "ADR-6b7becf0-f029-41e3-895e-b672dce94d3d"
type: adr
name: "CLI は cobra で組む"
status: 採用
issues:
  - "#552"
origin: "engine / CLI の技術スタックの確定（2026-06-13）。ADR-0011 §2 を書き直した"
justifies:
  - "REQ-637599dc-a1ec-4af5-9e97-e882c7df56d0"
supersedes:
  - "ADR-0011"
references:
  - "ADR-0007"
---
# ADR-6b7becf0-f029-41e3-895e-b672dce94d3d: CLI は cobra で組む

- ステータス: 採用
- 日付: 2026-11-01
- 関連: ADR-0007, GitHub Issue #552
- 起点: engine / CLI の技術スタックの確定（2026-06-13）。ADR-0011 §2 を書き直した

## 背景

CLI は、サブコマンド・フラグ・help・シェル補完を持つ。engine は第三者依存を持たないが、CLI 層には
人間向けの使い勝手のために最小限の第三者依存を許している。

## 決定

CLI のフレームワークに cobra を使う。

CLI の依存の許可範囲と固定の方法は REQ-637599dc-a1ec-4af5-9e97-e882c7df56d0 に置く。

## 根拠

cobra は Go のエコシステムで事実上の標準で、nixpkgs でもパッケージが整っている。サブコマンド・
フラグ・補完・help の整形が成熟しており、`--help` で内部の nix コマンドを開示する要件（ADR-0007）にも
help の整形の質で寄与する。

## 棄却した代替案

### stdlib の flag と手書きの dispatch で組む

依存は増えないが、help と補完を自作する負荷が高い。

### kong / urfave-cli を使う

軽量だが、nixpkgs のパッケージや補完の成熟度で cobra に劣る。

---
id: "ADR-0db061d1-fa30-436d-91d7-c0cec8804c52"
type: adr
name: "last-applied 記録に出所を持たせる"
status: stable
issues:
  - "#580"
  - "#498"
origin: "Issue #498（epic）の grilling（2026-10-31）。ADR-0070 §1 を書き直した"
justifies:
  - "REQ-ee5085a5-077b-4f5a-aa6a-841bc5515282"
references:
  - "ADR-767c33e5-73eb-4ef0-aba7-c87673fa5c2e"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# ADR-0db061d1-fa30-436d-91d7-c0cec8804c52: last-applied 記録に出所を持たせる

- ステータス: stable
- 日付: 2026-11-03
- 関連: ADR-767c33e5-73eb-4ef0-aba7-c87673fa5c2e, GitHub Issue #580, #498
- 起点: Issue #498（epic）の grilling（2026-10-31）。ADR-0070 §1 を書き直した

## 背景

home mode の config は、名前だけで profileDir と記録の鍵が決まる。home-manager モジュールは
`apply --all --manifest` で prebuilt 生成器を使い、CLI 単体は entrypoint を発見して nix
生成器を使う。両者が同じ config 名（典型は `default`）を使うと、profileDir と記録を共有する。

このとき片方の apply は、もう片方が置いた entry を「前回置いたが今回の manifest
に無い」ものとして除去する。交互に実行すると互いの entry を除去し合うが、
target の衝突ではないので foreign symlink の warning も出ず、利用者は何も知らされない。

## 決定

last-applied 記録に、その manifest がどこから来たかを示す出所（`origin`）を持たせる。出所は、
生成器の種別、nix 生成器なら発見した entrypoint、manifest が宣言していれば利用者が付けた識別子
`identity` で表す。フィールドの細目は REQ-ee5085a5-077b-4f5a-aa6a-841bc5515282 に置く。

## 根拠

記録は「この root に最後に置いたもの」を持ち、その出所も同じ事実の一部である。記録に持たせれば、
次の apply が今回の出所と比べて、別の出所からの上書きに気付ける。

profile や世代に持たせると、世代管理が layat の外へ出たときに読めなくなる。

## 棄却した代替案

### 出所を profile（世代）に持たせる

世代管理が layat の外へ出ると読めなくなる。

### 鍵に出所を含めて profileDir を分ける

共有そのものは無くなるが、移行するときに旧い profileDir と記録が孤立し、旧い出所が置いた entry
を誰も撤去しなくなる。

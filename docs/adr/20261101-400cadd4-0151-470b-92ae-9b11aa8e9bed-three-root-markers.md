---
id: "ADR-400cadd4-0151-470b-92ae-9b11aa8e9bed"
type: adr
name: "root は 3 つの marker で表す"
status: 採用
issues:
  - "#2"
origin: "起点 Issue: #2（root デフォルト）。ADR-0007 §2 の 3 マーカーと ADR-0013 §5 を書き直した"
justifies:
  - "REQ-37b56673-6e40-4a1b-a2a7-5d3c084e3e66"
  - "REQ-3f541d39-da41-4ef8-858b-707f54cf6a29"
  - "REQ-8d965ca2-f8fd-44a4-87f3-94e850e9f85b"
references:
  - "ADR-0036"
  - "ADR-0040"
  - "ADR-f85f1afd-9b87-4f7a-922e-97c62554f198"
---
# ADR-400cadd4-0151-470b-92ae-9b11aa8e9bed: root は 3 つの marker で表す

- ステータス: 採用
- 日付: 2026-11-01
- 関連: ADR-0036, ADR-0040, ADR-f85f1afd-9b87-4f7a-922e-97c62554f198, GitHub Issue #2, `docs/spec.md`
- 起点: 起点 Issue: #2（root デフォルト）。ADR-0007 §2 の 3 マーカーと ADR-0013 §5 を書き直した

## 背景

root は明示が必須である（ADR-f85f1afd-9b87-4f7a-922e-97c62554f198）。git toplevel や `$HOME`
は実行する環境によって変わるので、Nix の評価時にはパスへ展開できない。評価時に決まらない root
をどう書かせるかを決める必要があった。

## 決定

`root` は、3 つの marker（`projectRoot` / `homeRoot` / `systemRoot`）と、評価時に固定する
絶対パス文字列の union にする。`projectRoot` は git toplevel、`homeRoot` は `$HOME`、`systemRoot`
は `/` に、それぞれ engine が実行時に解決する。marker はパスを返す糖衣ではなく、解決の種別を
運ぶ入れ物で、種別は `manifest.json` の `rootKind` に記録される。`systemRoot` の実装は ADR-0036、
marker の `subdir` 適用形は ADR-0040 が定める。規範の細目は
REQ-37b56673-6e40-4a1b-a2a7-5d3c084e3e66・REQ-3f541d39-da41-4ef8-858b-707f54cf6a29・
REQ-8d965ca2-f8fd-44a4-87f3-94e850e9f85b に置く。

## 根拠

種別だけを評価時に決め、パスの解決を engine に任せれば、clone の場所やユーザーごとに違う root
を、pure な評価のまま扱える。project の root を git toplevel にすれば、どのサブディレクトリから
実行しても同じ root に解決される。絶対パス文字列を残すことで、任意の固定 root も書ける。

## 棄却した代替案

### marker を廃して絶対パス文字列だけにする

評価時に決まらない `$HOME` や git toplevel を書けない。`--impure` で環境変数を読むことになる。

### 絶対パス文字列を廃して 3 つの marker だけにする

任意の固定 root を書く手段を失う。

### `systemRoot` を API に載せない

marker を project と home の 2 つに絞れるが、`/` へ配置する手段が無くなる。

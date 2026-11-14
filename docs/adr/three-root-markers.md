---
id: "ADR-400cadd4-0151-470b-92ae-9b11aa8e9bed"
type: adr
name: "root は 3 つの marker で表す"
section: root-modes
status: stable
sources:
  - id: "issue-2"
    resource: "https://github.com/yasunori0418/layat/issues/2"
    title: "#2"
origin: "Issue #2（root デフォルト）。ADR-0007 §2 の 3 マーカーと ADR-0013 §5 を書き直した"
justifies:
  - "REQ-37b56673-6e40-4a1b-a2a7-5d3c084e3e66"
  - "REQ-3f541d39-da41-4ef8-858b-707f54cf6a29"
  - "REQ-8d965ca2-f8fd-44a4-87f3-94e850e9f85b"
references:
  - "ADR-15a358f4-1c81-49d5-b658-3a1fe39f0f75"
  - "ADR-f85f1afd-9b87-4f7a-922e-97c62554f198"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:26:59+09:00"
---
# root は 3 つの marker で表す

- ステータス: stable
- 日付: 2026-11-01
- 関連: [module と project では世代操作を公開しない](module-and-project-hide-generation-ops.md), [root の明示を mkManifest 層で必須にする](root-explicit-at-mkmanifest.md), GitHub Issue #2, `docs/spec.md`
- 起点: Issue #2（root デフォルト）。ADR-0007 §2 の 3 マーカーと ADR-0013 §5 を書き直した

## 背景

[root の明示を mkManifest 層で必須にする](root-explicit-at-mkmanifest.md)。git toplevel や `$HOME`
は実行する環境によって変わるので、Nix の評価時にはパスへ展開できない。評価時に決まらない root
をどう書かせるかを決める必要があった。

## 決定

`root` は、3 つの marker（`projectRoot` / `homeRoot` / `systemRoot`）と、評価時に固定する
絶対パス文字列の union にする。`projectRoot` は git toplevel、`homeRoot` は `$HOME`、`systemRoot`
は `/` に、それぞれ engine が実行時に解決する。marker はパスを返す糖衣ではなく、解決の種別を
運ぶ入れ物で、種別は `manifest.json` の `rootKind` に記録される。`systemRoot` の実装は [rootKind に system を足しても schemaVersion を上げない](system-rootkind-without-version-bump.md)、
marker の `subdir` 適用形は [root マーカーに subdir の適用形を持たせる](root-marker-subdir-applied-form.md) が定める。規範の細目は
[root は 3 マーカーと絶対パス文字列の union を取る](../requirements/root-marker-kinds.md)・[root マーカーは `kind` を運ぶ入れ物でパス解決は engine が行う](../requirements/root-markers-runtime-resolution.md)・
[home mode の root は層ごとに定まった供給元から解決する](../requirements/home-mode-root-resolution.md) に置く。

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

marker を project と home の 2 つに絞れる。`/` は固定の絶対パスとして書けるが、root 権限の確認や
system 用の profile の置き場所といった system mode の扱いを、marker で選べなくなる。

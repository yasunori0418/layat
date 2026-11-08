---
id: "ADR-8e253f2e-c255-482a-a0f8-0837e122e80f"
type: adr
name: "記録の読みは不正な記録を error にする"
status: stable
sources:
  - id: "issue-580"
    resource: "https://github.com/yasunori0418/layat/issues/580"
    title: "#580"
  - id: "issue-406"
    resource: "https://github.com/yasunori0418/layat/issues/406"
    title: "#406"
  - id: "issue-498"
    resource: "https://github.com/yasunori0418/layat/issues/498"
    title: "#498"
origin: "Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）と Issue #498（epic）の grilling（2026-10-31）。ADR-0058 §4d と ADR-0070 §2 を書き直した"
justifies:
  - "REQ-9d622539-a1e5-4719-af51-e0cd85b66f35"
references:
  - "ADR-0db061d1-fa30-436d-91d7-c0cec8804c52"
  - "ADR-55d8ecdf-86bb-45d0-8c60-ac9c630cb62d"
  - "ADR-767c33e5-73eb-4ef0-aba7-c87673fa5c2e"
  - "ADR-d97154cd-a2e5-4853-a682-800ff9435c89"
  - "ADR-fe035038-4415-4a1e-aa5c-509fdfce9955"
  - "ADR-ffa34ed1-71a6-4417-9f55-1102b307ce93"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# ADR-8e253f2e-c255-482a-a0f8-0837e122e80f: 記録の読みは不正な記録を error にする

- ステータス: stable
- 日付: 2026-11-03
- 関連: [last-applied 記録に出所を持たせる](record-carries-origin.md), [project mode では同じ link farm なら世代を積まず drift だけを直す](project-mode-generation-skip-and-drift-repair.md), [同一 profileDir への同時実行は後勝ちにする](same-profiledir-target-last-wins.md), [rollback は FS を収束させてからポインタを動かす](rollback-baseline-and-pointer-last.md), [1 config = 1 profile を全モードで自前に持つ](one-config-one-profile-all-modes.md), [古くなった配置の除去は manifest の差分で行う](stale-removal-by-manifest-diff.md), GitHub Issue #580, #406, #498
- 起点: Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）と Issue #498（epic）の grilling（2026-10-31）。ADR-0058 §4d と ADR-0070 §2 を書き直した

## 背景

記録は、古くなった配置を除去するときの比較の基準になる。読めない記録を「記録なし」として扱うと、
比較の基準は profile か空へ後退し、記録が正しければ消したはずの entry を残したり、消すべきでない
entry を消したりする。

記録には版がある。記録に出所を持たせると（[last-applied 記録に出所を持たせる](record-carries-origin.md)）記録の形が変わるが、
それより前に書かれた記録は手元に残る。

## 決定

apply は版 2 の記録を書く。読みは版 1 と 2 を受け付ける。版 1 は、出所を持たない記録として読む。
JSON として不正な記録、必須フィールドの欠落や型違いがある記録、未知のフィールドを含む記録、
それ以外の版の記録は error にする。「記録なし」には倒さない。

## 根拠

誤った比較の基準で古くなった配置を除去するより、止めて利用者に知らせる方が害が小さい。
古くなった配置の除去は、FS から symlink を消す。そのため比較の基準の誤りは、
利用者の手元のファイル構成を壊す。

版 1 を受け付けるのは、版を上げる前に書かれた記録で全員の apply が止まるのを避けるためである。
未知のフィールドは拒む。そうすれば、新しい engine が書いた記録を古い engine
が中途半端に解釈することも無い。

## 棄却した代替案

### 読めない記録を「記録なし」として扱い、profile へ後退する

誤った比較の基準で古くなった配置を除去する。

### 未知のフィールドを無視して読む

新しい版で足したフィールドの意味を無視したまま、比較の基準に使ってしまう。

### 版 1 の記録を拒む

版を上げた直後の apply が、記録を持つ全員の環境で止まる。

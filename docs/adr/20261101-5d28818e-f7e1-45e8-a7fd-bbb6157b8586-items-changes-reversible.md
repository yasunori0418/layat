---
id: "ADR-5d28818e-f7e1-45e8-a7fd-bbb6157b8586"
type: adr
name: "--json は items を全件、changes を差分で持ち可逆性を reversible で表す"
status: stable
issues:
  - "#552"
  - "yasunori0418/outturn#1"
origin: "layat の outturn 準拠化 grilling（2026-07-06）と outturn#1 の方針確定。ADR-0043 §4 を書き直した"
justifies:
  - "REQ-2ea19863-eaa2-466b-b1ed-3f56f6417c62"
  - "REQ-fa181aa6-29a2-48c3-ae07-cc1b9a3b0303"
references:
  - "ADR-6255a1ed-8df4-4230-b77d-cb9303a03e71"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:24:30+09:00"
---
# ADR-5d28818e-f7e1-45e8-a7fd-bbb6157b8586: --json は items を全件、changes を差分で持ち可逆性を reversible で表す

- ステータス: stable
- 日付: 2026-11-01
- 関連: ADR-6255a1ed-8df4-4230-b77d-cb9303a03e71, GitHub Issue #552, yasunori0418/outturn#1
- 起点: layat の outturn 準拠化 grilling（2026-07-06）と outturn#1 の方針確定。ADR-0043 §4 を書き直した

## 背景

outturn の `SubjectResult.result` は、処理単位の記録である `items` と、状態遷移の記録である `changes` を
分けて持つ。layat の apply は変更の無い entry も多く扱い、copy の上書きや削除のように戻せない操作も含む。

## 決定

`items` は対象の entry を変更の有無にかかわらずすべて並べるフルインベントリとし、`changes` は実際に
起きた差分だけを持つ（no-op を含めない）。可逆かどうかは `change.reversible` だけで表し、copy の
上書きと削除を `false` にする。不可逆を知らせる警告コードは設けない。

写像の規則と部分失敗の扱いは REQ-2ea19863-eaa2-466b-b1ed-3f56f6417c62 に、dryrun 側は
REQ-fa181aa6-29a2-48c3-ae07-cc1b9a3b0303 に置く。

## 根拠

`items` が全件を持てば、消費側は差分の無い entry も含めて状態を把握できる。`changes` を差分に絞るのは
outturn の規約どおりである。copy は世代の外に置き、hash で追跡しないので、旧内容を戻せない。
outturn §4 が「`reversible:false` を警告として扱うべき」と定めているので、警告コードを足すと同じ情報を
二重に表すことになる。

## 棄却した代替案

### 不可逆な変更に `W_IRREVERSIBLE` を付ける

`change.reversible` と同じ情報を二重に表すことになる。outturn#1 で棄却した。

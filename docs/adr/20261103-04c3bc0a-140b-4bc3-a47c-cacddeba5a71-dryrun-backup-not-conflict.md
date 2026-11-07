---
id: "ADR-04c3bc0a-140b-4bc3-a47c-cacddeba5a71"
type: adr
name: "--dryrun --backup は退避で解ける箇所を conflict にしない"
status: 採用
issues:
  - "#169"
  - "#172"
origin: "旧 epic #167 から統合された epic #172 の grilling セッション（2026-07-11・2026-07-12）で確定。ADR-0045 §6（出力規律）の dryrun の扱いを書き直した"
justifies:
  - "REQ-02a33511-0941-4813-b289-a05eb8e9aa57"
references:
  - "ADR-f607b605-c69e-4c7e-ac89-75b7bce5bee1"
  - "ADR-37191429-2099-4217-a552-47b16a7e504c"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:27:41+09:00"
---
# ADR-04c3bc0a-140b-4bc3-a47c-cacddeba5a71: --dryrun --backup は退避で解ける箇所を conflict にしない

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-f607b605-c69e-4c7e-ac89-75b7bce5bee1, ADR-37191429-2099-4217-a552-47b16a7e504c, GitHub Issue #169, #172
- 起点: 旧 epic #167 から統合された epic #172 の grilling セッション（2026-07-11・2026-07-12）で確定。ADR-0045 §6（出力規律）の dryrun の扱いを書き直した

## 背景

`apply --dryrun` は副作用なしに plan を示し、conflict があれば非ゼロで終わる。CI の事前の
関門に使える（ADR-f607b605-c69e-4c7e-ac89-75b7bce5bee1）。`--backup` と組み合わせたとき、
退避で解ける箇所をどう示すかを決める必要があった。

## 決定

`--dryrun --backup` では、`--backup` が無ければ conflict になる箇所を、退避と配置の予定として
示す。conflict ではないので、それだけなら終了コードは 0 になる。退避先が既にある箇所は、
dryrun でも conflict のままにする（ADR-37191429-2099-4217-a552-47b16a7e504c）。

表示と終了コードの規範は REQ-02a33511-0941-4813-b289-a05eb8e9aa57 に置く。

## 根拠

dryrun は、同じ引数で apply したときの plan を先に見せるためのものである。`--backup` 付きの
apply は退避して先へ進むので、dryrun が conflict を示すと予告と実際の結果が食い違う。

退避先が既にある箇所は、実際の apply でも止まる。plan の計算で判定できるので、dryrun でも
同じく conflict として示す。

## 棄却した代替案

### `--backup` を付けても dryrun では conflict として示す

dryrun の結果と実際の apply の結果が食い違う。CI の関門に `--backup` 付きの dryrun を使うと、
通るはずの apply を止めてしまう。

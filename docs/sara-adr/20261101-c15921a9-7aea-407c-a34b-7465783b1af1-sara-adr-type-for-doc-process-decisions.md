---
id: "SDR-c15921a9-7aea-407c-a34b-7465783b1af1"
type: sara_adr
name: "sara 品質管理機構の決定は sara_adr 型に分ける"
status: stable
issues:
  - "#499"
  - "#500"
origin: "Issue #499 の grilling（2026-11-01）で確定"
justifies:
  - "INF-659b139d-0cf8-4c65-b30d-93c5ee2dfc71"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-01T21:26:22+09:00"
---
# SDR-c15921a9-7aea-407c-a34b-7465783b1af1: sara 品質管理機構の決定は sara_adr 型に分ける

- ステータス: stable
- 日付: 2026-11-01
- 関連: GitHub Issue #499, #500, `docs/adr/README.md`
- 起点: Issue #499 の grilling（2026-11-01）で確定

## 背景

`docs/adr/` には、layat 本体の設計判断と、sara による品質管理機構（`docs/model.yaml`・グラフの検証・
起票ラッパー・item の運用規約）の設計判断が混在していた。後者は layat に固有ではなく、#500 で
別のリポジトリへ切り出す予定がある。

## 決定

sara 品質管理機構の決定は、`adr` とは別の型 `sara_adr`（prefix `SDR`、`docs/sara-adr/`）に記録する。
フィールドと relation は `adr` と同じにする。型とディレクトリの対応は `docs/adr/README.md` に置く。

## 根拠

型とディレクトリを分けておけば、切り出すときにディレクトリごと移せる。どの決定が切り出しの対象かを
1 本ずつ判定し直す手間が無い。

prefix を分けると、ID を見ただけでどちらの機構の決定かが分かる。

## 棄却した代替案

### `adr` 型のまま、ディレクトリだけを分ける

型が同じだと、切り出し先のモデルで `adr` が layat の設計判断を指すのか機構の決定を指すのかが
曖昧になる。ID からも区別できない。

### 切り出すときに分ける

切り出しの直前に全 ADR を仕分けることになる。起票の時点で分けておくほうが手間が少ない。

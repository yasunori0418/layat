---
id: "ADR-65e41903-5eee-4a00-a154-be8c830cdca9"
type: adr
name: "退避先を target に suffix を付けた名前に固定する"
status: stable
sources:
  - id: "issue-169"
    resource: "https://github.com/yasunori0418/layat/issues/169"
    title: "#169"
  - id: "issue-172"
    resource: "https://github.com/yasunori0418/layat/issues/172"
    title: "#172"
origin: "旧 epic #167 から統合された epic #172 の grilling セッション（2026-07-11・2026-07-12）で確定。ADR-0045 §1（CLI）の退避先の決め方を書き直した"
justifies:
  - "REQ-5dd5a4e9-6162-4fa5-b295-66844f5a4f3b"
references:
  - "ADR-33d8c466-6b38-46b6-8696-ad7521c72786"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T15:57:37+09:00"
---
# ADR-65e41903-5eee-4a00-a154-be8c830cdca9: 退避先を target に suffix を付けた名前に固定する

- ステータス: stable
- 日付: 2026-11-03
- 関連: ADR-33d8c466-6b38-46b6-8696-ad7521c72786, GitHub Issue #169, #172
- 起点: 旧 epic #167 から統合された epic #172 の grilling セッション（2026-07-11・2026-07-12）で確定。ADR-0045 §1（CLI）の退避先の決め方を書き直した

## 背景

`apply --backup`（ADR-33d8c466-6b38-46b6-8696-ad7521c72786）は、配置を塞ぐ記録外の実体を
rename で退避してから配置する。退避先の名前をどう決めるかを決める必要があった。

## 決定

退避先は、target のパスに `.` と suffix を付けた `<target>.<suffix>` に固定する。entry ごとに
退避先を変える手段は持たない。

退避先の書式と既定の suffix は REQ-5dd5a4e9-6162-4fa5-b295-66844f5a4f3b に置く。

## 根拠

target と同じディレクトリに置くので、退避は rename だけで済む。`ls` で target の隣に並ぶため、
何が退避されたかが一目で分かる。`--recopy` の上書きでも、同じ親ディレクトリの中に target の
名前へ接尾辞を付けた名前で退避しており、命名の規則がそろう。

## 棄却した代替案

### 隠しファイルの形（`.<target>.bak`）にする

退避物が隠しファイルになり、利用者が退避に気付きにくくなる。退避した事実を見えるようにする
方針と逆行する。

### entry ごとに退避先を指定させる

退避先の組み立てが entry の定義に入り込む。`--backup` は apply 全体の修飾で、entry 単位の
粒度は持たせていない。

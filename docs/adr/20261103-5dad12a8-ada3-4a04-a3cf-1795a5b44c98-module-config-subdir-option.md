---
id: "ADR-5dad12a8-ada3-4a04-a3cf-1795a5b44c98"
type: adr
name: "configs.<name>.subdir で config ごとに target の基点を寄せる"
status: 採用
issues:
  - "#585"
origin: "ADR-0064 §1・§2 を書き直した"
justifies:
  - "REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10"
  - "REQ-068f2fb3-5459-4325-9359-c931b5049e0d"
---
# ADR-5dad12a8-ada3-4a04-a3cf-1795a5b44c98: configs.<name>.subdir で config ごとに target の基点を寄せる

- ステータス: 採用
- 日付: 2026-11-03
- 関連: GitHub Issue #585, `docs/spec.md`
- 起点: ADR-0064 §1・§2 を書き直した

## 背景

root marker には subdir 適用形（`homeRoot { subdir = "..."; }`）があり、`mkManifest` は root を
動かさずに全 target へ subdir を前置できる。ところがモジュールは root を pin しているので、利用者は
適用形を渡せない。

配置の基点は config（= profile）ごとに違う。たとえば `ai-tools` は `~/.claude`、`desktop` は
`~/.config` の下にまとまる。基点が同じ entry 群には、毎回同じ接頭辞を書くことになる。

## 決定

共通オプション `layat.configs.<name>.subdir :: nullOr str`（既定 null）を設ける。モジュールは
subdir が null なら素の root marker を、非 null なら subdir 適用形を組み、その config の
`mkManifest` に root として渡す。root は pin したまま、target だけが `<subdir>/` の下に置かれる。
全 config 共通の `layat.subdir` は設けない。

共通オプションの型は REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10、home-manager での配線は
REQ-068f2fb3-5459-4325-9359-c931b5049e0d に置く。

## 根拠

- 配置の基点は役割（profile）ごとに違い、1 つの共通値では表せない。standalone の `layat.<name>` も
  config ごとに root を持つ
- モジュールが自分の性質で root を pin し、`root` オプションを公開しない原則と両立する。subdir 適用形は
  root を動かさず、target だけを変える
- subdir は評価時に target へ前置されるので、manifest には前置後の target が載る。configs 間の衝突の
  検査は、engine と `manifest.json` のスキーマを変えずにそのまま追従する
- 共通オプションの集合に置き、モジュールごとに集合を分岐させない

## 棄却した代替案

### 全 config 共通の `layat.subdir` を設ける

基点が config ごとに違う使い方に合わない。

### `root` オプションを公開する

利用者が root を指定し直せるようになり、モジュールが root を pin する原則が崩れる。target の基点を
変えるだけなら subdir で足りる。

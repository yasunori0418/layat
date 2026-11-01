---
id: "ADR-d97154cd-a2e5-4853-a682-800ff9435c89"
type: adr
name: "rollback は FS を収束させてからポインタを動かす"
status: 採用
origin: "ドキュメント計画段階の最終レビュー（grill）で surfaced した「実装前に決めておくべき」残細目の束。ADR-0015 §5 を書き直した"
justifies:
  - "REQ-0e341430-17f0-498b-9439-65491652163a"
references:
  - "ADR-0057"
  - "ADR-0058"
---
# ADR-d97154cd-a2e5-4853-a682-800ff9435c89: rollback は FS を収束させてからポインタを動かす

- ステータス: 採用
- 日付: 2026-11-01
- 関連: ADR-0057, ADR-0058
- 起点: ドキュメント計画段階の最終レビュー（grill）で surfaced した「実装前に決めておくべき」残細目の束。ADR-0015 §5 を書き直した

## 背景

layat は profile のディレクトリではなく任意の root に配置する。profile のポインタを前の世代へ
動かしても FS は変わらないので、rollback には再配置が要る。そのときの差分の基準と、FS の収束と
ポインタ移動の順序を決める必要がある。

## 決定

rollback は、離れる世代 N を baseline、戻る世代 N-1 を target とした通常の配置として行う。
apply の engine を（baseline, target）の差し替えで再利用し、FS を収束させてから、最後に
profile のポインタを N-1 へ移す。

baseline の読み方・実行段の順序・ポインタ移動の操作は
REQ-0e341430-17f0-498b-9439-65491652163a に置く。

## 根拠

ポインタを先に動かすと baseline が N-2 へずれ、stale 除去が誤った target を消す。ポインタを
最後にすれば、途中で失敗してもポインタは N を指したままで、次の操作の baseline が FS の現状と
対応し続ける。apply と同じ処理で反映するので、配置前の除去や保守的削除の不変条件も rollback で
そのまま守られる。

## 棄却した代替案

### profile のポインタだけを動かす

FS が変わらず、配置は戻らない。

### ポインタを先に動かしてから FS を収束させる

baseline が N-2 へずれて stale 除去が誤り、FS が profile と食い違う。

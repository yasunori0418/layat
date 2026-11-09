---
id: "ADR-d97154cd-a2e5-4853-a682-800ff9435c89"
type: adr
name: "rollback は FS を収束させてからポインタを動かす"
status: stable
origin: "ドキュメント計画段階の最終レビュー（grill）で surfaced した「実装前に決めておくべき」残細目の束。ADR-0015 §5 を書き直した"
justifies:
  - "REQ-0e341430-17f0-498b-9439-65491652163a"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:28:06+09:00"
---
# rollback は FS を収束させてからポインタを動かす

- ステータス: stable
- 日付: 2026-11-01
- 起点: ドキュメント計画段階の最終レビュー（grill）で surfaced した「実装前に決めておくべき」残細目の束。ADR-0015 §5 を書き直した

## 背景

layat は profile のディレクトリではなく任意の root に配置する。profile のポインタを前の世代へ
動かしても FS は変わらないので、rollback には再配置が要る。そのときの差分の基準と、FS の収束と
ポインタ移動の順序を決める必要がある。

## 決定

rollback は、離れる世代 N を基準世代、戻る世代 N-1 を target とした通常の配置として行う。
apply の engine を（基準世代, target）の差し替えで再利用し、FS を収束させてから、最後に
profile のポインタを N-1 へ移す。

基準世代の読み方・実行段の順序・ポインタ移動の操作は
[rollback は FS を先に収束させてから profile ポインタを最後に移す](../requirements/rollback-refit-then-pointer.md) に置く。

## 根拠

ポインタを先に動かすと基準世代が N-2 へずれ、古くなった配置の除去が誤った target を消す。
ポインタを最後にすれば、途中で失敗してもポインタは N を指したままで、次の操作の基準世代が
FS の現状と対応し続ける。apply と同じ処理で反映するので、配置前の除去や保守的削除の不変条件も
rollback でそのまま守られる。

## 棄却した代替案

### profile のポインタだけを動かす

FS が変わらず、配置は戻らない。

### ポインタを先に動かしてから FS を収束させる

基準世代が N-2 へずれ、古くなった配置の除去が誤り、FS が profile と食い違う。

---
id: "ADR-9723dbb3-6659-4a69-ac4f-411c98bfa818"
type: adr
name: ".pending の後片付けを gcroot を知らない engine の FS 操作として行う"
status: 採用
issues:
  - "#406"
origin: "Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0057 §7 を書き直した"
justifies:
  - "REQ-840b3641-6e76-46da-82e9-680cabd65abe"
references:
  - "ADR-51863e94-6a9d-4029-95e7-a7ad2bfa93c1"
---
# ADR-9723dbb3-6659-4a69-ac4f-411c98bfa818: .pending の後片付けを gcroot を知らない engine の FS 操作として行う

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-51863e94-6a9d-4029-95e7-a7ad2bfa93c1
- 起点: Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0057 §7 を書き直した

## 背景

既定の nix 生成器は、link-farm を `.pending` に out-link を張って build する。適用に成功したら engine が
`.pending` を消す（→ ADR-51863e94-6a9d-4029-95e7-a7ad2bfa93c1）。`.pending` が gcroot として働くのは
nix の事情で、世代操作を契約へ切り出す際に、この後片付けを engine と世代バックエンドのどちらが持つかが
問題になった。

## 決定

コミット成功後の `.pending` の除去は engine の素の FS 操作（`os.Remove`）のままとし、世代バックエンドには
渡さない。engine はこれを build 出力の一時リンクの片付けとして扱い、gcroot の概念を持たない。pending を
gcroot として張るのは生成器の事情として生成器に閉じる。

失敗時に残る pending の扱いは REQ-840b3641-6e76-46da-82e9-680cabd65abe に置く。

## 根拠

pending は生成器が張る一時リンクで、世代の概念ではない。世代バックエンドが後片付けを持つと、生成器の
事情が契約へ漏れる。engine にとっては固定パスのリンクを 1 本消すだけなので、gcroot を知る必要が無い。

## 棄却した代替案

### Commit に pending のパスを渡し、後片付けを世代バックエンドへ移す

世代バックエンドが生成器の一時リンクを知ることになり、契約に生成器の事情が漏れる。

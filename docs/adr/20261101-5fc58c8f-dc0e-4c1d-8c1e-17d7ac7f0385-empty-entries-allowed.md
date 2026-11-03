---
id: "ADR-5fc58c8f-dc0e-4c1d-8c1e-17d7ac7f0385"
type: adr
name: "空の entries を正当な全クリアとして許可する"
status: 採用
origin: "ADR-0018 の後に行った docs の横断検証で残っていた細目。ADR-0019 §4 を書き直した"
justifies:
  - "REQ-fc64de4c-c82b-419c-8706-07d8d97daa37"
supersedes:
  - "ADR-0019"
references:
  - "ADR-0002"
---
# ADR-5fc58c8f-dc0e-4c1d-8c1e-17d7ac7f0385: 空の entries を正当な全クリアとして許可する

- ステータス: 採用
- 日付: 2026-11-01
- 関連: ADR-0002, `docs/spec.md`
- 起点: ADR-0018 の後に行った docs の横断検証で残っていた細目。ADR-0019 §4 を書き直した

## 背景

`entries = {}` の config を apply すると、前の世代が置いた symlink が全て stale として除去され、
その config の配置が全て消える。これを正当な操作として扱うか、誤りとして止めるかを決める必要が
あった。

## 決定

空の `entries` は、その config の配置を全て撤去する正当な表現として許可する。評価時のエラーにも
警告にもしない。

規範は REQ-fc64de4c-c82b-419c-8706-07d8d97daa37 に置く。

## 根拠

stale の除去は保守的で、layat が置いた symlink だけを消し、実ファイルには触れない（ADR-0002）。
全て撤去しても利用者のデータは失われないので、空を特別に扱う理由が無い。

entry を 1 つ以上必須にすると、config の配置を消すという正当な手段が無くなる。

## 棄却した代替案

### 空の `entries` に警告を出すか、評価時のエラーにする

誤った全クリアに気付く助けにはなるが、意図した撤去を妨げるか、毎回の雑音になる。

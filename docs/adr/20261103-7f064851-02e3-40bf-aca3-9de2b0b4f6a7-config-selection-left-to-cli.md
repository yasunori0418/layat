---
id: "ADR-7f064851-02e3-40bf-aca3-9de2b0b4f6a7"
type: adr
name: "config の選別を CLI だけに任せる"
status: 採用
issues:
  - "#530"
origin: "Issue #530 の grilling（2026-10-25）。ADR-0068 §3 を書き直した"
justifies:
  - "REQ-6c51facc-fa7a-4b5e-939b-44e7578d7a4f"
references:
  - "ADR-04cc2a03-78dd-4bf3-9cc7-3b8ed7faf429"
  - "ADR-db293e0c-4789-45b7-93a9-adedc021344f"
---
# ADR-7f064851-02e3-40bf-aca3-9de2b0b4f6a7: config の選別を CLI だけに任せる

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-04cc2a03-78dd-4bf3-9cc7-3b8ed7faf429, ADR-db293e0c-4789-45b7-93a9-adedc021344f, GitHub Issue #530
- 起点: Issue #530 の grilling（2026-10-25）。ADR-0068 §3 を書き直した

## 背景

`mkApplyCommand`（ADR-db293e0c-4789-45b7-93a9-adedc021344f）は root を指定して呼ばれる。指定した
root の config だけを適用する方法は 2 つある。Nix 側で各 farm の `passthru.rootKind` を見て
`--manifest` に渡す config を絞る方法と、全 config を渡して CLI の root フィルタ
（ADR-04cc2a03-78dd-4bf3-9cc7-3b8ed7faf429）に絞らせる方法である。

## 決定

`mkApplyCommand` は Nix 側で config を選別せず、`manifests` の全 config を `--manifest` に並べる。
選別は CLI の root フィルタだけが行う。選ばれない root の farm もビルド依存に入ることは許容する。

並べ方の規範は REQ-6c51facc-fa7a-4b5e-939b-44e7578d7a4f に置く。

## 根拠

rootKind による選別は CLI がすでに持っている。Nix 側にも持たせると同じ判定が 2 か所に分かれ、
片方だけ変えたときに食い違う。

選ばれない farm をビルドする費用は、devShell の依存が増えるだけで、配置の結果には影響しない。

## 棄却した代替案

### Nix 側で `passthru.rootKind` を見て選別する

不要な farm をビルドせずに済むが、CLI の root フィルタと判定が二重になる。

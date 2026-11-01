---
id: "ADR-3aebefee-405d-4342-b754-eef8eac29b1b"
type: adr
name: "symlink の張替えは unlink と symlink"
status: 採用
origin: "ADR-0017 §4 を書き直した"
justifies:
  - "REQ-61856da1-8883-401e-ad57-9f326b96d400"
---
# ADR-3aebefee-405d-4342-b754-eef8eac29b1b: symlink の張替えは unlink と symlink

- ステータス: 採用
- 日付: 2026-11-01
- 関連: `docs/spec.md`
- 起点: ADR-0017 §4 を書き直した

## 背景

既存の symlink を別の先へ張り替えるとき、unlink と symlink の 2 操作で行うと、間でプロセスが
異常終了したときに target が一時的に無くなる。これを防ぐかどうかを決める必要があった。

## 決定

symlink の張替えは unlink と symlink の 2 操作で行い、rename による atomic な差し替えは採らない。
途中で異常終了して target が消えても、同じ apply をもう一度実行すれば収束する。規範の細目は
REQ-61856da1-8883-401e-ad57-9f326b96d400 に置く。

## 根拠

apply は冪等で、再実行すれば宣言どおりの状態に戻る。プロセスの異常終了自体がまれで、
起きても再実行で直るので、atomic な差し替えの実装コストに見合う安全性の向上は小さい。

## 棄却した代替案

### rename による atomic な差し替え

target が一時的に消えることは無くなるが、一時名の symlink の作成と後始末が増える。冪等な再実行で足りる。

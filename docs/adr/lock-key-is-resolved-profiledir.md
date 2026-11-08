---
id: "ADR-ac648001-7e22-4fe2-a457-596064e32717"
type: adr
name: "ロックのキーを解決後の profileDir にする"
status: stable
origin: "engine の実行時セマンティクスの確定（2026-06-13）。ADR-0013 §2 を書き直した"
justifies:
  - "REQ-2aa3abbc-90b2-486e-92de-d785554bdeb3"
  - "REQ-a8edc58f-4adc-4637-b888-ab8ccc7e73e4"
  - "REQ-60c6b7ea-e936-4ce8-bd75-ad35e9c693b9"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-01T23:16:14+09:00"
---
# ADR-ac648001-7e22-4fe2-a457-596064e32717: ロックのキーを解決後の profileDir にする

- ステータス: stable
- 日付: 2026-11-01
- 起点: engine の実行時セマンティクスの確定（2026-06-13）。ADR-0013 §2 を書き直した

## 背景

engine は同じ profile への並行実行を直列化するためにロックを取る。何をキーにロックするかで、
直列化される範囲が変わる。

profile は root ごとに分かれていて、同じ config 名でも root が違えば別の profileDir になる。同じ root の
下には、複数の config の profileDir が並ぶ。

## 決定

ロックのキーは、root を解決した後に確定する profileDir（config 専用のディレクトリ）とする。同じ profileDir への
apply / rollback / reset は直列化し、別の profileDir どうしは互いに待たない。同じ root の別の config も、
別の root の同名の config も、互いに待たない。

ロックを取る段は [実行フローの順序は eval 先行 → flock → build とし build をロック内に閉じる](../requirements/exec-flow-eval-lock-build.md)、ロックの単位となるディレクトリは
[profileDir は config 専用ディレクトリとし、profile リンク・世代・pending out-link をその中に並べる](../requirements/profile-on-disk-layout.md)、reset のロックは [reset は名指し必須で profileDir 単位の blocking flock を取る](../requirements/reset-named-only-and-flock.md) に置く。

## 根拠

競合するのは同じ profileDir の配置と世代のコミットだけなので、直列化もその範囲で足りる。profileDir は
root ごとに分かれているので、同じ project を 2 か所にクローンして両方の devShell に入っても、互いを
待たない。

## 棄却した代替案

### config 名だけでロックする

同じ project の 2 つのクローンは root が違うのに config 名が同じなので、互いをブロックする。

### root の階層（`<roothash>`）でロックする

同じ root の下の別の config まで直列化され、競合しない適用が待たされる。

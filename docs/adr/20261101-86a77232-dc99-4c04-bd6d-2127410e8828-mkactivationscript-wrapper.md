---
id: "ADR-86a77232-dc99-4c04-bd6d-2127410e8828"
type: adr
name: "mkActivationScript で config ごとのラッパー derivation を提供する"
status: 不採用
issues:
  - "#3"
origin: "Issue #3（CLI 化）。ADR-0007 §3 で棄却された提案（ADR-0002 の初期設計、ADR-0006 の決定節での再定義）を書き直した"
justifies:
  - "REQ-f4d7d4ab-fbdb-48c6-b29f-08dd88e72645"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-02T22:56:50+09:00"
---
# ADR-86a77232-dc99-4c04-bd6d-2127410e8828: mkActivationScript で config ごとのラッパー derivation を提供する

- ステータス: 不採用
- 日付: 2026-11-01
- 関連: GitHub Issue #3
- 起点: Issue #3（CLI 化）。ADR-0007 §3 で棄却された提案（ADR-0002 の初期設計、ADR-0006 の決定節での再定義）を書き直した

## 背景

ADR-0002 の初期設計では、`mkActivationScript` が config ごとに、配置ロジックを含む bash を生成する案だった。
ADR-0006 は決定節でこれを、`layat apply --manifest <manifest>` を呼ぶだけの薄いラッパー derivation に
再定義した。利用者はこのラッパーを `nix run .#<config>` で起動する想定だった。

実装に入る前に Issue #3 が出た。`mkActivationScript` を呼ばないと layat コマンドを作れないので、
engine とは別に CLI が欲しいという要望で、この提案は ADR-0007 §3 で棄却された。
lib に `mkActivationScript` が実装されたことは無い。

## 決定

`mkActivationScript` は提供しない。entrypoint の公開面は `lib.mkManifest` が返す manifest だけにし、
engine の起動は CLI とモジュールの配線が担う。

CLI と engine の 2 層構成は REQ-f4d7d4ab-fbdb-48c6-b29f-08dd88e72645 に置く。

## 根拠

ラッパーは engine の呼び出しを config の数だけ複製するだけなので、CLI が manifest を名前で引けるなら
要らない。提供しなければ、lib の公開面は配置データの生成だけに揃い、lib が engine のパッケージを
参照することもない。

## 棄却した代替案

### 提案どおりラッパーを主な起動手段として提供する

`nix run .#<config>` で config ごとに適用できるが、ラッパーを作らないと layat を使えないという
#3 の不満と、flake 以外の entrypoint への対応の要求を満たせない。

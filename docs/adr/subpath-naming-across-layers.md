---
id: "ADR-3417c5b0-8c5f-4c2c-827a-6f56bce301ab"
type: adr
name: "配置元の中のパスを全層で subpath と呼ぶ"
status: stable
origin: "「リポジトリ全体を示す文字列として \".\" 以外の表現は無いか」という問いの設計検討。ADR-0008 §1 と §5 を書き直した"
justifies:
  - "REQ-3e446ad9-a6f4-4229-b5c5-184754b0ef51"
  - "REQ-0b0cd1e3-bfeb-45c1-978d-e2e11c568336"
references:
  - "ADR-12ad7852-e689-44d7-b9fe-0667effb3d13"
  - "ADR-3e283fe2-0177-4caf-8b60-577c5ee3b10a"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-07T02:02:56+09:00"
---
# 配置元の中のパスを全層で subpath と呼ぶ

- ステータス: stable
- 日付: 2026-11-03
- 関連: [entry の未知のキーを評価時に拒否する](strict-unknown-keys-rejected.md), [entry の配置種別のフィールドを method と名付ける](mode-renamed-to-method.md), `docs/spec.md`, [subpath](../ubiquitous-language/subpath.md), [src](../ubiquitous-language/src.md)
- 起点: 「リポジトリ全体を示す文字列として `"."` 以外の表現は無いか」という問いの設計検討。ADR-0008 §1 と §5 を書き直した

## 背景

entry には「どの配置元か」を表す `src` と、「その中のどのパスか」を表すフィールドがあり、後者は
`source` と呼んでいた。`src` は英語で `source` の略そのものなので、読み手は 2 つが直交する概念だと
見分けられず、`src` を `source` の別名や typo と読む。リポジトリ全体を `"."` 以外でどう表すかという
問いも、`"."` の是非ではなくこの名前の近さから来ていた。

## 決定

配置元の中のパスを表すフィールドを `subpath` と呼ぶ。`src` は Nix の慣例どおり残す。
entries の記述・`manifest.json` のキー・Go engine のどこでも同じ `subpath` を使う。
entry のフィールドと旧名の扱いは [entry submodule は strict とし未知キーと旧名を評価時エラーにする](../requirements/entry-submodule-strict.md)、manifest の
キーは [manifest.json の entries は attrset を配列へ正規化し 5 フィールドを記録する](../requirements/manifest-entries-array.md) に置く。

## 根拠

「`src` はどの物か、`subpath` はその中のどのパスか」という直交が名前に出る。`subpath` は
file と dir のどちらも取れるという語義にも合う。Nix の記述・`manifest.json`・engine で名前が
揃えば、`manifest.json` を config と照らし合わせてデバッグするときに読み替えが要らない。
改名した時点では実装が無く、`manifest.json` のキーを変える移行の手間も無かった。

## 棄却した代替案

### `source` を残して `src` を `repo` などに改名する

`src` はリポジトリ以外の store パスや out-of-store のローカルパスも取るので、`repo` では狭い。
Nix の慣例からも離れる。

### 両方の名前を残して文書で区別する

変更は最小で済む。ただ、名前の衝突という根本が残り、読み手の学習コストが下がらない。

### `manifest.json` のキーだけ `source` のまま残す

層ごとに同じ概念の名前が変わる。Nix の記述と `manifest.json` を比べるたびに読み替えが要る。

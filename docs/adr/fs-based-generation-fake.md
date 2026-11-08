---
id: "ADR-4c8d459a-c371-4f0d-b009-16e2ca096e49"
type: adr
name: "世代バックエンドのテストダブルを FS ベースの Fake に統一する"
status: stable
sources:
  - id: "issue-406"
    resource: "https://github.com/yasunori0418/layat/issues/406"
    title: "#406"
origin: "Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0057 §2c を書き直した"
justifies:
  - "DSG-f8d9bba5-cfd5-4a19-b1fd-fd8ecc8a1814"
references:
  - "ADR-30f105c7-85a8-4fab-9dea-3277d932b524"
  - "ADR-419914ba-9896-4dd1-853b-fb514fcb9f38"
  - "ADR-51863e94-6a9d-4029-95e7-a7ad2bfa93c1"
  - "ADR-58fdbad9-5ed7-469d-b6b2-6462b2173888"
  - "ADR-5fc58c8f-dc0e-4c1d-8c1e-17d7ac7f0385"
  - "ADR-74423bb8-1453-4689-a25e-30bf46536afc"
  - "ADR-b1e1dd5e-61a4-4d04-8389-137ef2347c56"
  - "ADR-d97154cd-a2e5-4853-a682-800ff9435c89"
  - "ADR-f90509f2-d75a-4162-baa5-a5cfdb8a7c30"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# ADR-4c8d459a-c371-4f0d-b009-16e2ca096e49: 世代バックエンドのテストダブルを FS ベースの Fake に統一する

- ステータス: stable
- 日付: 2026-11-03
- 関連: ADR-30f105c7-85a8-4fab-9dea-3277d932b524, ADR-419914ba-9896-4dd1-853b-fb514fcb9f38, ADR-51863e94-6a9d-4029-95e7-a7ad2bfa93c1, ADR-58fdbad9-5ed7-469d-b6b2-6462b2173888, ADR-5fc58c8f-dc0e-4c1d-8c1e-17d7ac7f0385, ADR-74423bb8-1453-4689-a25e-30bf46536afc, ADR-b1e1dd5e-61a4-4d04-8389-137ef2347c56, ADR-d97154cd-a2e5-4853-a682-800ff9435c89, ADR-f90509f2-d75a-4162-baa5-a5cfdb8a7c30
- 起点: Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0057 §2c を書き直した

## 背景

engine と cmd のテストは、nix-env の代わりにそれぞれ独自の差し替え関数を持っていた。代替が
いくつもあると、どれが本番の挙動を写しているのかが分からない。世代操作が契約越しであることも、
テストで確かめられない。

engine は前世代の manifest などを読むために、profileDir のレイアウトを直接読む。

## 決定

世代バックエンドのテストダブルは、`internal/generation` に置く FS ベースの Fake 1 つに統一する。
Fake は profile リンクと世代のリンクを実際に FS 上に張る。engine と cmd のテストは、nix-env の代わりに
これだけを使う。

Fake の置き場所とパッケージの構成は DSG-f8d9bba5-cfd5-4a19-b1fd-fd8ecc8a1814 に置く。

## 根拠

Fake が実際にリンクを張るので、engine の読み取りは手を加えずに動き、テストは本番と同じコードで profileDir を読む。
テストダブルが 1 つなら、契約を変えたときに追従する箇所も 1 つで済む。

## 棄却した代替案

### メモリ上で世代を持つモック

engine が直接読むリンクが FS に無く、engine の読み取りをテストできない。読み取りまで契約に載せる必要が出る。

### テストごとの差し替え関数を残す

代替が乱立したまま残り、世代操作が契約越しであることを検証できない。

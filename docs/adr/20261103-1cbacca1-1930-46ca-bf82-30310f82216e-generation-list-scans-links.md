---
id: "ADR-1cbacca1-1930-46ca-bf82-30310f82216e"
type: adr
name: "世代の一覧を nix-env を経由せず世代リンクの走査で得る"
status: stable
issues:
  - "#129"
origin: "system mode の e2e で、root 所有の profileDir を非 root が `list-generations` すると nix-env の lock 取得で失敗することが分かった（2026-10-18）。ADR-0063 §1・§2 を書き直した"
justifies:
  - "REQ-833205ca-adfc-4c86-bf66-ae4d8fe4885b"
  - "DSG-f8d9bba5-cfd5-4a19-b1fd-fd8ecc8a1814"
references:
  - "ADR-ffe1afcb-0b51-4b87-b698-0a6dea307bdd"
  - "ADR-4c8d459a-c371-4f0d-b009-16e2ca096e49"
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
# ADR-1cbacca1-1930-46ca-bf82-30310f82216e: 世代の一覧を nix-env を経由せず世代リンクの走査で得る

- ステータス: stable
- 日付: 2026-11-03
- 関連: ADR-ffe1afcb-0b51-4b87-b698-0a6dea307bdd, ADR-4c8d459a-c371-4f0d-b009-16e2ca096e49, ADR-30f105c7-85a8-4fab-9dea-3277d932b524, ADR-419914ba-9896-4dd1-853b-fb514fcb9f38, ADR-51863e94-6a9d-4029-95e7-a7ad2bfa93c1, ADR-58fdbad9-5ed7-469d-b6b2-6462b2173888, ADR-5fc58c8f-dc0e-4c1d-8c1e-17d7ac7f0385, ADR-74423bb8-1453-4689-a25e-30bf46536afc, ADR-b1e1dd5e-61a4-4d04-8389-137ef2347c56, ADR-d97154cd-a2e5-4853-a682-800ff9435c89, ADR-f90509f2-d75a-4162-baa5-a5cfdb8a7c30, GitHub Issue #129
- 起点: system mode の e2e で、root 所有の profileDir を非 root が `list-generations` すると nix-env の lock 取得で失敗することが分かった（2026-10-18）。ADR-0063 §1・§2 を書き直した

## 背景

`nix-env --list-generations` は世代を列挙するだけでも profileDir に lock を取る。書けない profileDir では
列挙が失敗するので、root 所有の system mode の profile を非 root が `list-generations` すると exit 1 になる。
`list-generations` は読み取り専用のコマンドで、system mode でも非 root に世代を見せたい。

テストダブルの Fake は、世代の一覧を世代リンクの走査で返している。nix-env の出力をパースしていた
本番とは、列挙のロジックが別だった。

## 決定

nixprofile の List は nix-env を起動せず、profileDir の世代リンクを読み取り専用で走査して返す。走査は
`internal/generation` の 1 つの関数にまとめ、Fake の List もそれに委ねる。Commit と Switch は nix-env で
行う。

current の判定で、`profile` のリンク先が世代の名前でないときに実体の一致で探す規則は、Fake から引き継いだ
補助の規則として残す。nix-env の作る profile は常に世代の名前を指すので、本番ではこの分岐に入らない。

List の規範は REQ-833205ca-adfc-4c86-bf66-ae4d8fe4885b、走査の共用は DSG-f8d9bba5-cfd5-4a19-b1fd-fd8ecc8a1814 に置く。

## 根拠

nix 本体の世代の列挙も、世代リンクの走査とリンクの更新時刻の整形で、出力は同じになる。lock を取らない分、
書けない profileDir でも読める。

列挙のロジックを Fake と共用すれば、engine と cmd のテストが Fake で確かめた列挙が本番でも同じ結果になる。

profileDir のレイアウトは契約の固定部分（→ ADR-ffe1afcb-0b51-4b87-b698-0a6dea307bdd）で、engine はすでに
直接読んでいる。バックエンドの側で同じレイアウトを読んでも、契約の形は変わらない。

## 棄却した代替案

### lock の取得に失敗したときだけ走査へ切り替える

列挙のロジックが 2 つ残り、どちらが動いたかで結果が変わりうる。失敗の判別も nix-env のエラー文言に頼る
ことになる。

### nix-env での列挙を続ける

非 root による system mode の `list-generations` が失敗したまま残る。

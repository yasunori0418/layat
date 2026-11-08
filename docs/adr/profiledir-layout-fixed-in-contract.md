---
id: "ADR-ffe1afcb-0b51-4b87-b698-0a6dea307bdd"
type: adr
name: "profileDir の内部レイアウトを世代バックエンド契約の固定部分にする"
status: stable
sources:
  - id: "issue-406"
    resource: "https://github.com/yasunori0418/layat/issues/406"
    title: "#406"
origin: "Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0057 §3a を書き直した"
justifies:
  - "REQ-833205ca-adfc-4c86-bf66-ae4d8fe4885b"
  - "REQ-2aa3abbc-90b2-486e-92de-d785554bdeb3"
references:
  - "ADR-419914ba-9896-4dd1-853b-fb514fcb9f38"
  - "ADR-30f105c7-85a8-4fab-9dea-3277d932b524"
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
# ADR-ffe1afcb-0b51-4b87-b698-0a6dea307bdd: profileDir の内部レイアウトを世代バックエンド契約の固定部分にする

- ステータス: stable
- 日付: 2026-11-03
- 関連: [profile は専用ディレクトリレイアウトに置く](profile-directory-layout.md), [symlink farm を GC アンカー専用にする](farm-as-gc-anchor-only.md), [store path は out-link で受け取る](store-path-via-out-link.md), [.pending の残留を許容する](pending-leftover-tolerated.md), [空の entries を正当な全クリアとして許可する](empty-entries-allowed.md), [新規配置と張替えを先に、世代のコミットを最後に行う](placement-order-commit-last.md), [apply --all は失敗しても残りを続け、最後に集約して非 0 で終える](apply-all-continue-and-aggregate.md), [rollback は FS を収束させてからポインタを動かす](rollback-baseline-and-pointer-last.md), [世代は link farm と nix profile に乗せる](generations-on-link-farm-and-nix-profile.md)
- 起点: Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0057 §3a を書き直した

## 背景

世代操作を世代バックエンドの契約へ切り出すと、profileDir の中身（profile リンク・世代のリンク・
`.pending`）を誰が所有するかが問題になる。engine は前世代の manifest や drift の判定のために、
このレイアウトを直接読んでいる。

## 決定

profileDir の内部レイアウトは世代バックエンドの契約の固定部分とし、実装ごとに変えない。レイアウトの
定義は `internal/paths` に置き、engine とバックエンドの双方がそれに従う。

並びそのものは [profile は専用ディレクトリレイアウトに置く](profile-directory-layout.md) が決めており、細目は
[profileDir は config 専用ディレクトリとし、profile リンク・世代・pending out-link をその中に並べる](../requirements/profile-on-disk-layout.md) に置く。レイアウトを契約の固定部分とする規範は
[engine は世代操作を世代バックエンドの契約（Commit / List / Switch / UserBase / SystemBase）越しに行う](../requirements/generation-backend-contract.md) に置く。

## 根拠

レイアウトが固定なら、engine の読み取りは契約を経ずに動き、FS ベースの Fake もそのまま使える。
実装ごとにレイアウトを変えられるようにすると、engine の読み取りを全て契約の操作に置き換えなければならない。

## 棄却した代替案

### レイアウト全体をバックエンドの所有にする

engine の読み取りを契約に載せるか、Fake が実装ごとのレイアウトを模倣するかの二択になる。2 つ目の実装が
無い段階では、どちらも検証できない。

---
id: "ADR-8b8c9109-a479-4571-8b10-5a6079127441"
type: adr
name: "前世代の読み取りは契約に載せず engine が直接読む"
section: generations
status: stable
sources:
  - id: "issue-406"
    resource: "https://github.com/yasunori0418/layat/issues/406"
    title: "#406"
origin: "Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0057 §3b を書き直した"
justifies:
  - "REQ-833205ca-adfc-4c86-bf66-ae4d8fe4885b"
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
# 前世代の読み取りは契約に載せず engine が直接読む

- ステータス: stable
- 日付: 2026-11-03
- 関連: [symlink farm を GC アンカー専用にする](farm-as-gc-anchor-only.md), [profile は専用ディレクトリレイアウトに置く](profile-directory-layout.md), [store path は out-link で受け取る](store-path-via-out-link.md), [.pending の残留を許容する](pending-leftover-tolerated.md), [空の entries を正当な全クリアとして許可する](empty-entries-allowed.md), [新規配置と張替えを先に、世代のコミットを最後に行う](placement-order-commit-last.md), [apply --all は失敗しても残りを続け、最後に集約して非 0 で終える](apply-all-continue-and-aggregate.md), [rollback は FS を収束させてからポインタを動かす](rollback-baseline-and-pointer-last.md), [世代は link farm と nix profile に乗せる](generations-on-link-farm-and-nix-profile.md)
- 起点: Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0057 §3b を書き直した

## 背景

engine は、前世代の `manifest.json`・世代リンクの解決・drift の判定に使う前世代の link-farm を、
profileDir から直接読む。世代バックエンドの契約を決める際、これらの読み取りも契約の操作に含めるかが
問題になった。

## 決定

前世代の読み取りは契約の操作に載せず、engine が profileDir を直接読む。外部の世代バックエンドが
実在した時点で、読み取りを契約へ移す。

engine が直接読んでよいという規範は [engine は世代操作を世代バックエンドの契約（Commit / List / Switch / UserBase / SystemBase）越しに行う](../requirements/generation-backend-contract.md)、パッケージの分担は
[internal/generation は契約・Generation 型・FS ベースの Fake を親パッケージに、nix-env 実装を nixprofile サブパッケージに置き、engine は Backend 1 フィールドで受ける](../design/generation-package-layout.md) に置く。

## 根拠

実装が nix 1 つの段階で読み取りを契約に載せても、使われない操作が固まるだけで検証できない。
profileDir のレイアウトは契約の固定部分なので、engine が直接読んでもバックエンドとの食い違いは起きない。

## 棄却した代替案

### 読み取りの操作もいま契約に載せる

engine の読み取り 5 箇所と、それを模倣する Fake の実装が増える。2 つ目の実装が無いので、増えた操作を
検証できない。

---
id: "ADR-764ffff1-6b5a-4ed7-89d0-5700eb3ee883"
type: adr
name: "state の基底を設定で上書きさせない"
status: stable
sources:
  - id: "issue-406"
    resource: "https://github.com/yasunori0418/layat/issues/406"
    title: "#406"
  - id: "issue-140"
    resource: "https://github.com/yasunori0418/layat/issues/140"
    title: "#140"
origin: "Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）と Issue #140（epic #129: system mode の実装）の計画（2026-10-18）。ADR-0057 §6b・ADR-0062 §2 を書き直した"
justifies:
  - "REQ-833205ca-adfc-4c86-bf66-ae4d8fe4885b"
  - "REQ-0d1f557b-7ae7-41ea-b17b-cdb5c80935d4"
  - "REQ-2aa3abbc-90b2-486e-92de-d785554bdeb3"
  - "DSG-f8d9bba5-cfd5-4a19-b1fd-fd8ecc8a1814"
  - "REQ-9d622539-a1e5-4719-af51-e0cd85b66f35"
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
# state の基底を設定で上書きさせない

- ステータス: stable
- 日付: 2026-11-03
- 関連: [symlink farm を GC アンカー専用にする](farm-as-gc-anchor-only.md), [profile は専用ディレクトリレイアウトに置く](profile-directory-layout.md), [store path は out-link で受け取る](store-path-via-out-link.md), [.pending の残留を許容する](pending-leftover-tolerated.md), [空の entries を正当な全クリアとして許可する](empty-entries-allowed.md), [新規配置と張替えを先に、世代のコミットを最後に行う](placement-order-commit-last.md), [apply --all は失敗しても残りを続け、最後に集約して非 0 で終える](apply-all-continue-and-aggregate.md), [rollback は FS を収束させてからポインタを動かす](rollback-baseline-and-pointer-last.md), [世代は link farm と nix profile に乗せる](generations-on-link-farm-and-nix-profile.md)
- 起点: Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）と Issue #140（epic #129: system mode の実装）の計画（2026-10-18）。ADR-0057 §6b・ADR-0062 §2 を書き直した

## 背景

layat の state には 2 種類の基底がある。世代の profile の基底（user は `<state>/nix/profiles/layat`、
system は `/nix/var/nix/profiles/layat`）と、last-applied 記録の基底（user は `<state>/layat`、system mode は `/var/lib/layat`）。
基底を変えると、既存の世代と記録を見失う。

一方で、root 権限を user namespace で得る e2e は、実システムの基底を汚さずに配置と世代を検証したい。

## 決定

state の基底を設定ファイルのキーで上書きさせない。system mode の 2 つの基底（profile と記録）に限り、
engine の入力と、CLI がそれへ渡すテスト用の環境変数 `LAYAT_SYSTEM_PROFILE_BASE`・`LAYAT_SYSTEM_STATE_DIR`
から上書きできる。用途は、root 権限を user namespace で得る e2e が、実システムの基底を汚さずに配置と世代を
検証することに限る。

設定キーを設けない規範は [engine は世代操作を世代バックエンドの契約（Commit / List / Switch / UserBase / SystemBase）越しに行う](../requirements/generation-backend-contract.md)、基底を世代バックエンドが与える規範は
[profileDir は config 専用ディレクトリとし、profile リンク・世代・pending out-link をその中に並べる](../requirements/profile-on-disk-layout.md)、system mode の基底は [system mode は root を / に解決し、変更系操作の前に euid 0 を確認し、profile と last-applied 記録を system 基底に置く](../requirements/system-mode-root-gate-and-record-base.md) に置く。
profile 基底の上書きを engine の入力へ渡す配線は [internal/generation は契約・Generation 型・FS ベースの Fake を親パッケージに、nix-env 実装を nixprofile サブパッケージに置き、engine は Backend 1 フィールドで受ける](../design/generation-package-layout.md)、記録基底を
engine の入力で上書きする規範は [apply は最後に配置した manifest を last-applied 記録に書き、古くなった配置の除去の比較の基準は記録を優先して読む](../requirements/last-applied-record.md) に置く。

## 根拠

profile の基底は世代をどこに置くかで、ファイル配置の関心ではない。置き場所は世代バックエンドの責務で、
契約の UserBase・SystemBase がその所有を表す。設定キーを足すと、世代管理を layat から切り出すときに
layat 側へ互換の負債が残る。

記録の基底も、変えれば既存の記録を見失う点は同じなので、profile の基底に揃える。テストの隔離に
は環境変数があれば足りる。

## 棄却した代替案

### 設定ファイルのキーで基底を上書きできるようにする

利用者の需要が無く、上書きのたびに既存の世代と記録を見失うおそれがある。世代管理を切り出すときにも
互換の負債になる。

### テスト用の上書き手段も持たない

e2e が実システムの基底を書き換えずに system mode を検証できない。

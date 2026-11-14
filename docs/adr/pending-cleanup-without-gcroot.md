---
id: "ADR-9723dbb3-6659-4a69-ac4f-411c98bfa818"
type: adr
name: ".pending の後片付けを gcroot を知らない engine の FS 操作として行う"
section: generations
status: stable
sources:
  - id: "issue-406"
    resource: "https://github.com/yasunori0418/layat/issues/406"
    title: "#406"
origin: "Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0057 §7 を書き直した"
justifies:
  - "REQ-833205ca-adfc-4c86-bf66-ae4d8fe4885b"
  - "REQ-840b3641-6e76-46da-82e9-680cabd65abe"
  - "DSG-f8d9bba5-cfd5-4a19-b1fd-fd8ecc8a1814"
references:
  - "ADR-51863e94-6a9d-4029-95e7-a7ad2bfa93c1"
  - "ADR-30f105c7-85a8-4fab-9dea-3277d932b524"
  - "ADR-419914ba-9896-4dd1-853b-fb514fcb9f38"
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
# .pending の後片付けを gcroot を知らない engine の FS 操作として行う

- ステータス: stable
- 日付: 2026-11-03
- 関連: [store path は out-link で受け取る](store-path-via-out-link.md), [symlink farm を GC アンカー専用にする](farm-as-gc-anchor-only.md), [profile は専用ディレクトリレイアウトに置く](profile-directory-layout.md), [.pending の残留を許容する](pending-leftover-tolerated.md), [空の entries を正当な全クリアとして許可する](empty-entries-allowed.md), [新規配置と張替えを先に、世代のコミットを最後に行う](placement-order-commit-last.md), [apply --all は失敗しても残りを続け、最後に集約して非 0 で終える](apply-all-continue-and-aggregate.md), [rollback は FS を収束させてからポインタを動かす](rollback-baseline-and-pointer-last.md), [世代は link farm と nix profile に乗せる](generations-on-link-farm-and-nix-profile.md)
- 起点: Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0057 §7 を書き直した

## 背景

既定の nix 生成器は、link-farm を `.pending` に out-link を張って build する。適用に成功したら engine が
`.pending` を消す（→ [store path は out-link で受け取る](store-path-via-out-link.md)）。`.pending` が gcroot として働くのは
nix の事情で、世代操作を契約へ切り出す際に、この後片付けを engine と世代バックエンドのどちらが持つかが
問題になった。

## 決定

コミット成功後の `.pending` の除去は engine の素の FS 操作（`os.Remove`）のままとし、世代バックエンドには
渡さない。engine はこれを build 出力の一時リンクの片付けとして扱い、gcroot の概念を持たない。pending を
gcroot として張るのは生成器の事情として生成器に閉じる。

engine が `.pending` を消す規範は [engine は世代操作を世代バックエンドの契約（Commit / List / Switch / UserBase / SystemBase）越しに行う](../requirements/generation-backend-contract.md) と
[internal/generation は契約・Generation 型・FS ベースの Fake を親パッケージに、nix-env 実装を nixprofile サブパッケージに置き、engine は Backend 1 フィールドで受ける](../design/generation-package-layout.md)、失敗時に残る pending の扱いは
[失敗時に残る pending gcroot は config あたり最大 1 個に有界とし回収処理を持たない](../requirements/pending-gcroot-bounded.md) に置く。

## 根拠

pending は生成器が張る一時リンクで、世代の概念ではない。世代バックエンドが後片付けを持つと、生成器の
事情を契約が抱え込む。engine にとっては固定パスのリンクを 1 本消すだけなので、gcroot を知る必要が無い。

## 棄却した代替案

### Commit に pending のパスを渡し、後片付けを世代バックエンドへ移す

世代バックエンドが生成器の一時リンクを知ることになり、契約が生成器の事情を抱え込む。

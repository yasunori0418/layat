---
id: "ADR-efa7685c-6fa4-45df-b7a3-8f3f58e4118c"
type: adr
name: "Build の事後条件を世代バックエンドがコミットできる link-farm のパスに置く"
status: stable
sources:
  - id: "issue-406"
    resource: "https://github.com/yasunori0418/layat/issues/406"
    title: "#406"
origin: "Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0057 §5・ADR-0055 §4 を書き直した"
justifies:
  - "REQ-194e4209-d804-4a4b-a2b8-3d39c6c33729"
references:
  - "ADR-51863e94-6a9d-4029-95e7-a7ad2bfa93c1"
  - "ADR-28605860-770e-41df-95dd-7dc05b20497a"
  - "ADR-2f472fc4-23ed-4169-aa49-8bded9e336d9"
  - "ADR-30f105c7-85a8-4fab-9dea-3277d932b524"
  - "ADR-419914ba-9896-4dd1-853b-fb514fcb9f38"
  - "ADR-58fdbad9-5ed7-469d-b6b2-6462b2173888"
  - "ADR-5caf1c14-a054-4c8d-95a2-c7f89e0c5246"
  - "ADR-5fc58c8f-dc0e-4c1d-8c1e-17d7ac7f0385"
  - "ADR-63039f93-cd78-4356-8cd3-82c2f49a1aaa"
  - "ADR-6b7becf0-f029-41e3-895e-b672dce94d3d"
  - "ADR-74423bb8-1453-4689-a25e-30bf46536afc"
  - "ADR-b1e1dd5e-61a4-4d04-8389-137ef2347c56"
  - "ADR-c7ffc21a-4be1-4468-b63c-2f47eedd9ab3"
  - "ADR-d97154cd-a2e5-4853-a682-800ff9435c89"
  - "ADR-f40c3d37-638d-420a-a0fe-3908cbf0aec0"
  - "ADR-f90509f2-d75a-4162-baa5-a5cfdb8a7c30"
  - "ADR-fe035038-4415-4a1e-aa5c-509fdfce9955"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# ADR-efa7685c-6fa4-45df-b7a3-8f3f58e4118c: Build の事後条件を世代バックエンドがコミットできる link-farm のパスに置く

- ステータス: stable
- 日付: 2026-11-03
- 関連: [store path は out-link で受け取る](store-path-via-out-link.md), [apply --all は config ごとの情報を 1 回の eval でまとめて取る](apply-all-single-eval.md), [eval を先に行い、flock を取ってから build する](eval-first-then-flock-then-build.md), [symlink farm を GC アンカー専用にする](farm-as-gc-anchor-only.md), [profile は専用ディレクトリレイアウトに置く](profile-directory-layout.md), [.pending の残留を許容する](pending-leftover-tolerated.md), [内部コマンドの開示で透明性を保つ](internal-command-disclosure.md), [空の entries を正当な全クリアとして許可する](empty-entries-allowed.md), [nix の experimental-features は利用者が有効にしておく前提にする](experimental-features-user-prerequisite.md), [CLI は cobra で組む](cobra-cli-framework.md), [新規配置と張替えを先に、世代のコミットを最後に行う](placement-order-commit-last.md), [apply --all は失敗しても残りを続け、最後に集約して非 0 で終える](apply-all-continue-and-aggregate.md), [root の必須性を別の確認を書かずに標準のエラーで担保する](root-required.md), [rollback は FS を収束させてからポインタを動かす](rollback-baseline-and-pointer-last.md), [home mode の CLI と lib の版ずれを manifest の確認に任せて許容する](home-mode-skew-guarded-by-schemaversion.md), [世代は link farm と nix profile に乗せる](generations-on-link-farm-and-nix-profile.md), [1 config = 1 profile を全モードで自前に持つ](one-config-one-profile-all-modes.md)
- 起点: Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0057 §5・ADR-0055 §4 を書き直した

## 背景

manifest 生成器の Build は link-farm を build し、pending に張ってそのパスを返す。返したパスは世代として
コミットされる。nix profile への世代のコミット（`nix-env --set`）は store 外のパスを拒否する。

pending に張るのは、build からコミットまでの間に GC が実行されても link-farm が dangling にならないようにするため
（→ [store path は out-link で受け取る](store-path-via-out-link.md)）。

世代操作は世代バックエンドの契約越しになったが、Build の事後条件を「store path」と書くと、生成器の契約が
nix profile に結び付いたまま残る。

## 決定

Build は、選択中の世代バックエンドがコミットできる link-farm のパスを pending に張って返す。nixprofile が
受け入れるのは store path なので、いまは store path を返すことになる。

達成方法は生成器に任せる。非 nix の生成器向けに store へ入れるヘルパーや、store 外のパスを受け入れる実装は
作らない。

生成器の契約は [CLI は manifest 生成器の契約（Discover / Roots / Build / DryBuild）越しに manifest を得る](../requirements/manifest-generator-contract.md) に置く。

## 根拠

事後条件をバックエンドに委ねる書き方にしておけば、バックエンドが増えても生成器の契約を書き換えずに済む。
いま要るのは nixprofile が受け入れる store path だけで、それ以外のパスを受け入れる実装には使い道が無い。

## 棄却した代替案

### 事後条件を「世代コミット可能な store path」に固定する

世代バックエンドを差し替えられるようになっても、生成器の条件だけが nix profile に縛られたまま残る。

### store 外のパスを受け入れる実装や、store へ入れるヘルパーをいま作る

受け入れる世代バックエンドが無く、使われない実装になる。

---
id: "ADR-0057"
type: adr
name: "世代バックエンドの契約"
status: 採用
issues:
  - "#444"
  - "#406"
origin: "Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）"
justifies:
  - "REQ-833205ca-adfc-4c86-bf66-ae4d8fe4885b"
  - "REQ-6c4e174a-4d16-477a-96ff-17cb4eb5b564"
  - "REQ-706de717-4e47-471a-a1c0-448635be159c"
  - "REQ-4ffda99a-7062-4c00-915f-70b525cb215b"
  - "REQ-1be4d678-959c-44d7-a346-44bfd95af56e"
  - "REQ-2aa3abbc-90b2-486e-92de-d785554bdeb3"
  - "REQ-840b3641-6e76-46da-82e9-680cabd65abe"
  - "DSG-f8d9bba5-cfd5-4a19-b1fd-fd8ecc8a1814"
revises:
  - "ADR-0002"
  - "ADR-0006"
  - "ADR-0011"
  - "ADR-0015"
  - "ADR-0036"
  - "ADR-0055"
references:
  - "ADR-0023"
  - "ADR-0025"
  - "ADR-0034"
  - "ADR-0043"
  - "ADR-0056"
---
# ADR-0057: 世代バックエンドの契約

- ステータス: 採用
- 日付: 2026-10-11
- 関連: ADR-0002（世代管理を nix profile に乗せる）, ADR-0006（engine は `nix`（profile）と `git` をサブプロセスで叩く・却下案「エンジンが世代 symlink を直接管理」）, ADR-0011（`.pending` out-link と `nix-env --set` 後の削除）, ADR-0015（世代操作の `nix-env --profile` 系統一）, ADR-0023（eval 先行 → flock → build）, ADR-0025（profileDir の専用ディレクトリレイアウト）, ADR-0034（`layat prune` の走査基底）, ADR-0036（system mode の state 基底）, ADR-0043（`--json` の outturn 準拠）, ADR-0055（manifest 生成器の契約・Build 事後条件）, ADR-0056（生成器の選択機構）
- 改訂対象: ADR-0002 / ADR-0015 §3 が「世代操作は `nix-env --profile <dir>` 系」とした主体を世代バックエンドの契約へ移す（nix-env はその既定実装 nixprofile になる）。ADR-0006 の「`nix`（profile）と `git` のみサブプロセスで叩く」を `git` のみへ改訂し、却下案「エンジンが世代 symlink を直接管理」を契約の実装として再評価する。ADR-0011 §4 の「`nix-env --set` 成功後に `.pending` を削除する」の主体を engine の FS 操作として位置づけ直す。ADR-0036 §3 の system 基底 `/nix/var/nix/profiles/layat` の所有者を nixprofile にする。ADR-0055 §4 の Build 事後条件「世代コミット可能な store path」を「選択中の世代バックエンドがコミット可能な link-farm のパス」へ緩和する（文言のみ）
- 起点: Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）

## 背景

engine は世代操作を `nix-env --profile <p>` 系（`--set` / `--list-generations` / `--switch-generation`）で直接叩き、profile の基底 `<state>/nix/profiles/layat`（`paths.Base`）と system の基底 `/nix/var/nix/profiles/layat`（prune の定数）を固定で持つ。差し替え点は `CommitFunc` / `ListGenerationsFunc` / `SwitchGenerationFunc` の 3 関数型で、nil のとき nix-env へ暗黙にフォールバックする。manifest の取得側は ADR-0055 で生成器の契約越しになったが、世代側は nix 前提のまま engine に残っている。

### 北極星

世代管理は将来、独立した世代管理 CLI へ切り出し、layat は「ファイル配置を上手くやるアプリケーション」に専念する。終着点のトポロジーは、パイプラインが `layat apply --json` → `jq` → `generation commit` と外から順に呼ぶ形で、世代管理 CLI は状態識別子（link-farm のパス）の中身を解釈せずにそのまま受け渡し、outturn の JSON エンベロープが境界プロトコルになる。

本 ADR はその **in-process 段階**を決める。engine と世代バックエンドの境界を、そのままプロセス境界に載せられる形（入出力がパスと整数だけ）の Go インターフェイスとして固定する。**プロセス境界の契約は世代管理 CLI 側が定義する**。それが確定した時点で、本契約の Commit / List / Switch の配線を世代管理 CLI の呼び出しへ置き換える。

### 終着点で layat 側に残るもの（未決）

次は profileDir に由来する責務で、世代管理が layat から出た後も layat 側に残る。本 ADR では解かず、分離計画（Issue #447）の未決として扱う。

- flock の鍵（profileDir）
- `layat prune` の孤児系列の判定（基底の走査）
- stale 除去の baseline となる前世代 manifest の取得（→ ADR-0058 が記録で置き換える）
- `Switch` の分解: 終着点では `generation rollback` → `layat activate <STATE>` に分かれる

現状の事実として、`apply --json` の `generation` payload は profile リンクと before / after の番号だけで状態識別子を出していない。`apply --manifest <link-farm>` が「与えた状態 artifact の配置」に最も近い既存操作だが、世代コミットまで行う。どちらも本 ADR では変えない。

## 決定

### 1. スコープは差し替え可能化まで

契約 + nixprofile 実装 + テストダブルまでで止める。非 nix の世代バックエンドは作らない。世代管理 CLI への分離は独立 issue（#447）が持つ。

### 2. 契約は 5 操作・入出力はパスと整数のみ

パッケージ `internal/generation` に置く。Go コールバックを契約に含めない。

```go
type Generation struct{ Number int; Date string; Current bool }
type Backend interface {
    Commit(profileLink, linkFarm string) error
    List(profileLink string) ([]Generation, error)
    Switch(profileLink string, gen int) error
    UserBase(stateDir string) string   // nixprofile: <state>/nix/profiles/layat
    SystemBase() string                // nixprofile: /nix/var/nix/profiles/layat
}
```

- nix 実装は `internal/generation/nixprofile`（`New() generation.Backend`）。nix-env 3 コマンドと世代一覧のパース、2 基底を所有する。
- テストダブルは `internal/generation` に FS ベースの Fake を置く（`profile-<N>-link` を実際に張る）。engine / cmd のテストが持つ nix-env 代替はこれに統一する。
- engine 側は差し替え点の 3 関数型を `Backend` 1 フィールドに置き換える。profileDir を解決する全入口（apply / rollback / reset / prune / profile 解決）が `Backend` を要求し、nil は error。**nix-env への暗黙のフォールバックは廃止する**。cmd 層は常に nixprofile を注入する。

### 3. レイアウトは契約の固定部分。読み操作は将来移管する

profileDir 内部（`profile` / `profile-<N>-link` / `.pending`）は `internal/paths` に残し、契約の固定部分とする。engine の manifest 読み（前世代の `manifest.json`・世代リンクの解決・drift の前世代 link-farm 解決）は契約を経由せず直接読む。**外部バックエンドが実在した時点で読み操作を契約へ移す**。それまで読み操作を契約に載せても、使われない操作が固まるだけで検証できない。

### 4. engine が叩く外部コマンドは `git` のみ

engine が直接叩く外部コマンドは `git`（toplevel 解決）だけになる。`nix-env` は nixprofile の内側にだけ現れる。

### 5. Build 事後条件は文言のみ緩和する

ADR-0055 §4 の「世代コミット可能な link-farm（= store path）」を「**選択中の世代バックエンドがコミット可能な link-farm のパス**（nixprofile は store path を要求する）」に読み替える。実装は増やさない（非 store パスの受け入れは作らない）。

### 6. 選択機構と基底の設定キーは作らない

- バックエンドを選ぶフラグ / 環境変数 / `layat.toml` キーは足さない。**2 つ目のバックエンドが現れた時点で、ADR-0056 と同じ優先順位（フラグ → 環境変数 → プロジェクト設定 → ユーザー設定 → 既定）で選ぶ**。これは予約であり、本 ADR では値も項目名も決めない。
- state 基底（`<state>/nix/profiles/layat` / `/nix/var/nix/profiles/layat`）を設定ファイルで上書きする機構も足さない。置き場所は世代バックエンド（将来は世代管理アプリケーション）の責務で、layat の設定インターフェイスは増やさない。`UserBase` / `SystemBase` はその第一歩で、世代管理が layat から出るときは基底の所有ごと出る。

### 7. `.pending` の後片付けは engine に残す

コミット成功後の `.pending` の除去は engine の素の `os.Remove` のまま。説明は「build 出力の一時リンクの片付け」とし、gcroot の語を engine から外す。pending を gcroot として張るのは生成器（nixgen）の事情で、engine は gcroot を知らない。

## 根拠

- **プロセス内インターフェイス**: 生成器（ADR-0055）と同じ理由。実装が nix 1 つの段階で外部プログラム規約を決めると使われない仕様が固まる。
- **パスと整数だけ**: 契約がプロセス境界へ載る形なら、世代管理 CLI が確定したときに配線の置き換えで済む。Go コールバックを混ぜると in-process にしか載らない。
- **暗黙フォールバックの廃止**: nil で nix-env に落ちる差し替え点は「注入し忘れ」を本番挙動として通す。error にすれば全入口の注入をテストで検証できる。
- **レイアウト固定**: 読み操作まで契約に載せると、Fake が実 FS を模倣するか契約が肥大するかの二択になる。FS ベースの Fake なら engine の読みは無改変で動く。
- **基底の設定キーを作らない**: 基底は「世代をどこに置くか」で、layat の配置の関心ではない。設定キーを足すと世代管理を切り出すときに layat 側へ互換負債が残る。

## 影響

- 実装は Issue #445（`internal/generation` + nixprofile の抽出・engine の置換・Fake への統一）。last-applied 記録は ADR-0058 / Issue #448。CONTEXT / README / glossary への反映は Issue #446。
- ADR-0002 / 0006 / 0011 / 0015 / 0036 / 0055 に改訂注記を足す。
- 要求: 契約（REQ-833205ca-adfc-4c86-bf66-ae4d8fe4885b）を起票し、外部コマンド制約（REQ-6c4e174a-4d16-477a-96ff-17cb4eb5b564）を `git` のみへ、世代操作と GC（REQ-706de717-4e47-471a-a1c0-448635be159c）を契約越しへ改訂する。nix コマンドの開示と委譲（REQ-4ffda99a-7062-4c00-915f-70b525cb215b）・世代機構（REQ-1be4d678-959c-44d7-a346-44bfd95af56e）・profileDir レイアウト（REQ-2aa3abbc-90b2-486e-92de-d785554bdeb3）・pending の有界性（REQ-840b3641-6e76-46da-82e9-680cabd65abe）から engine が nix-env を直接叩く前提を外す。パッケージ構成は DSG-f8d9bba5-cfd5-4a19-b1fd-fd8ecc8a1814。

## 棄却した代替案

- **レイアウト全部をバックエンド所有にする（読み操作も契約に載せる）**: engine の manifest 読み 5 箇所と Fake の模倣が増えるだけで、2 つ目の実装が無い段階では検証できない。外部バックエンドが実在した時点で移す（§3）。
- **Commit に pending のパスを渡し後片付けをバックエンドへ移す**: pending は生成器が張る一時リンクで、世代の概念ではない。バックエンドが知ると生成器の事情が契約へ漏れる。
- **設定キーで基底を上書きする**: 世代管理が layat から出るときに互換負債になる（§6）。
- **非 nix バックエンドの同時実装（自前 symlink + 採番 + GC アンカー代替）**: GC 安全性を自前で背負う ADR-0006 の却下理由が残る。別アプリケーションとして検討する（#147 へ番外登録）。
- **差し替え点の 3 関数型を残し nil フォールバックも残す**: 注入し忘れが nix-env 直叩きとして通り、契約越しであることを検証できない。

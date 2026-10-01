---
id: "ADR-0055"
type: adr
name: "manifest 生成器の契約"
status: 採用
issues:
  - "#401"
  - "#396"
  - "#395"
origin: "Issue #395（epic: manifest 生成器のインターフェイス化）の grilling（2026-09-09）と、評価器の診断の扱いを決める Issue #396 の結論"
justifies:
  - "REQ-194e4209-d804-4a4b-a2b8-3d39c6c33729"
  - "REQ-badc7e10-0ba4-40d4-b334-6e40193119db"
  - "REQ-7a2f1ecf-4675-45aa-80c0-a8fc58db9edd"
  - "DSG-25ad3cce-f921-4db2-aae0-c0263d4b8295"
  - "REQ-f4d7d4ab-fbdb-48c6-b29f-08dd88e72645"
  - "REQ-f9920c87-8551-4aa3-bf03-26fdf4191ed6"
  - "REQ-4ffda99a-7062-4c00-915f-70b525cb215b"
revises:
  - "ADR-0006"
  - "ADR-0007"
  - "ADR-0023"
  - "ADR-0026"
  - "ADR-0031"
  - "ADR-0043"
references:
  - "ADR-0011"
  - "ADR-0025"
  - "ADR-0054"
---
# ADR-0055: manifest 生成器の契約

- ステータス: 採用
- 日付: 2026-10-02
- 関連: ADR-0006（engine の契約は `manifest.json` 1 本）, ADR-0007（CLI の責務・透明性）, ADR-0011（pending out-link による GC 窓の封鎖）, ADR-0023（eval 先行 → flock → build）, ADR-0025（experimental-features の案内）, ADR-0026（`--manifest` 経路）, ADR-0031（成功時沈黙・`--debug`）, ADR-0043（`--json` の outturn 準拠）, ADR-0054（改名。manifest の生成は nix でなくてもよい）
- 改訂対象: ADR-0006 / ADR-0007 が「CLI が内部で `nix build` / `nix eval` を回す」とした manifest 取得の主体を生成器インターフェイスへ移す。ADR-0023 §1 の rootKind 先読みを生成器の契約の前提条件に位置づける。ADR-0026 の `--manifest` 経路を prebuilt 生成器として再定義する（外面は不変）。ADR-0031 §3 の「内部 nix コマンド開示」の主体を生成器にし、生成器の stderr を沈黙規律の対象外にする。ADR-0043 の `E_*_BUILD` を「生成器の失敗」へ読み替える
- 起点: Issue #395（epic: manifest 生成器のインターフェイス化）の grilling（2026-09-09）と、評価器の診断の扱いを決める Issue #396 の結論

## 背景

engine（`internal/`）が受け取る契約は既に `manifest.json` 1 本で、engine は誰が manifest を生成したかに関知しない（→ ADR-0006）。改名（→ ADR-0054）も「manifest の生成は nix でなくてもよい」を理由に挙げた。

一方 `cmd/layat` は nix を前提に組まれたままで、ここが「nix でなくてもよい」を阻んでいる。nix 前提は `cmd/layat/nix.go` に集中している: entrypoint 発見（`flake.nix` / `shell.nix` / `default.nix`）、`<ref>#layat.<system>.<name>` の attr path と `currentSystem` の eval、rootKind の先読み eval（`evalRoot` / `evalAllRoots`）、`nix build --out-link <profileDir>/.pending` による link-farm 取得（`buildFunc`）と gcroot を張らない取得（`dryBuildFunc` / `buildManifestStorePath`）、`exec.Command("nix", ...)` の直呼び、nix 固有のエラー文字列による分類と案内文（`isExperimentalDisabled` / `experimentalGuidance`）。

診断の扱いにも構造的な欠陥がある（→ Issue #396）。eval 系（`runNixCapture`）は成功時に子プロセスの stderr を捨てるため、`mkManifest` の `lib.warn` のような評価器の警告が利用者に届かない。build 系（`runNixStream`）は stderr を素通しするので経路によって挙動が非対称で、失敗の分類と案内文は nix 固有の文字列判定に依存している。

nix 自体は実行時に残る。世代のバックエンドは nix profile（`nix-env --set`）で、`nix-env --set` は store 外のパスを拒否する（実測済み）。生成器の出力は最終的に store path でなければならない。

## 決定

### 1. 境界はプロセス内 Go インターフェイスとする

- 契約・prebuilt 実装・テストダブルを `internal/generator` に、nix 実装を `internal/generator/nixgen` に置く。
- **cmd 層に残るのは生成器の選択と engine への注入だけ**。選択機構は ADR-0056 が定める。
- 生成器を経由するサブコマンドは apply（`--dryrun` を含む）/ gitignore / list-generations / reset / rollback の全て。世代操作系が manifest 経路を使う理由は rootKind の取得だけで、Roots 操作で自然に乗る。prune は元々 nix を使わない。`init`（`nix flake init`）は生成器化の範囲外。
- engine（`internal/engine`）は変更しない。既存の `BuildFunc` 等の注入可能な seam をそのまま使う。

### 2. 必須操作は Discover / Roots / Build / DryBuild の 4 つ

| 操作 | 入力 | 出力・事後条件 |
|---|---|---|
| **Discover** | `-f` の値（無ければ cwd） | entrypoint を発見する。発見規則は実装が持つ（nixgen は既存の `flake.nix` → `shell.nix` → `default.nix`） |
| **Roots** | `name`（`--all` 用の一括取得は全件） | **build せずに** `manifest.Root` を返す。`RootKind` / `Root`（fixed のとき）に加え、正規化後の target 一覧 `Targets` を返す |
| **Build** | `name` + pending のパス | 世代コミット可能な link-farm（= store path）を pending に張って返す（→ §4） |
| **DryBuild** | `name` | link-farm の store path を **gcroot を張らずに**返す（`--dryrun` と gitignore が使う） |

- **契約の引数は `name` のみ**。`system`（`layat.<system>.<name>` の `<system>`）は nixgen の内部に隠す。
- 現行の `rootInfo`（`RootKind` / `Root` / `Targets`）は `manifest.Root` に統合する。`Targets []string` は **Go 構造体側のフィールド**で、manifest.json スキーマ v1 は変えない（prebuilt は `entries` から導出する）。

### 3. Roots の先読みは能力宣言ではなく契約の前提条件

Roots は全ての生成器が build 前に答えられなければならない。「先読みできる生成器だけが Roots を持つ」という能力宣言にはしない。apply の実行フロー「eval 先行（rootKind 取得 → root 解決 → profileDir 確定）→ flock → build（ロック内）」（→ ADR-0023 §1）は、build 前に profileDir を確定できることを前提にしている。先読みを任意にすると、先読みできない生成器では flock の前に build する経路が生まれ、ロック外 build による out-link 競合が戻る。

### 4. Build の事後条件は「世代コミット可能な store path」

- Build は **世代コミット可能な link-farm（= store path）を pending に張って返す**。達成方法は実装に任せ、非 nix 生成器向けの store 化ヘルパーは作らない。
- 根拠: 世代のバックエンドは nix profile で、`nix-env --set` は store 外のパスを拒否する。pending に張るのは、build 完了から `nix-env --set` までの間に並行 GC が走っても link-farm が dangling 化しないため（→ ADR-0011）。
- この事後条件は世代バックエンドが nix profile であることに由来する。**緩めるのは後継 epic #406**（世代バックエンドのインターフェイス化）の担当で、本 ADR では緩めない。

### 5. `--manifest` は prebuilt 生成器とする

- `--manifest <link-farm>` は prebuilt 実装に写す: Discover = 与えられたパス / Roots = その link-farm の `manifest.json` から読む / Build = そのパスを返す（DryBuild も同じ）。
- ADR-0026 の外面（フラグ・挙動・`-f` / `--all` との排他）は不変。テストダブル（`Fake`）も同じ契約の実装として持つ。
- prebuilt は `--generator` の選択肢に出さない（→ ADR-0056）。

### 6. 診断契約: 生成器は渡された writer へ書き、表示は CLI が決める

- 生成器は診断を **CLI から渡された `io.Writer`** へ書き、stdout には書かない（CLI は `os.Stderr` を渡す）。`--json` 時も stdout にはエンベロープ以外を出さない。
- **成功時の stderr は eval / build の両経路とも素通し**する。nixgen は両経路とも子プロセスの stderr を tee（writer へ素通し + キャプチャ）する。生成器の stderr は ADR-0031 の沈黙規律の対象外で、沈黙規律を生成器の警告を握り潰す根拠にしない。
- 失敗は `generator.Error{Generator, Stage, Kind, Message, Guidance, Stderr}` で返す。

  | フィールド | 内容 |
  |---|---|
  | `Generator` | 生成器名（`nix` / prebuilt） |
  | `Stage` | `discover` / `roots` / `build`（DryBuild の失敗も `build`） |
  | `Kind` | `PrerequisiteMissing`（前提条件の未充足。nix では experimental-features）/ `NotFound`（config 名が無い）/ `Failed`（それ以外） |
  | `Message` | 要約 1 行。内部コマンドの失敗では失敗したコマンドを含める（内部コマンドを持たない生成器では省く） |
  | `Guidance` | 解決策の案内（無ければ空） |
  | `Stderr` | キャプチャした生の診断 |

  `generator.Error` は原因の error を保持し、`Unwrap` で辿れるようにする（公開フィールドは上の 6 つから増やさない）。

- **nix 固有の文字列判定・案内文・NotFound の前置文は nixgen に閉じる**（`internal/generator/nixgen/` の外に出さない）。
- **表示は CLI が決める**。人間向けは要約 + 案内のみで、生の診断は再掲しない（既に writer へ流れている）。`--json` は `errors[].message` に要約とキャプチャを含める。

### 7. `--json` のコードは `E_*_BUILD` 1 本のまま

- `E_LAYAT_BUILD` の意味を「nix の失敗」から「**生成器の失敗**」へ読み替える。新しいコードは足さない。
- `E_LAYAT_BUILD` に分類するのは **prebuilt 以外の生成器の Stage `roots` / `build` の失敗**（生成器が manifest を評価・生成できなかった失敗）に限る。**Stage `discover` の失敗（entrypoint の不在・`-f` のパスの不在）と prebuilt の失敗（入力の link-farm・`manifest.json` を読めない）は、Stage によらず生成器化の前の分類を変えない**。CLI はこれらを原因の error チェーンから従来の規則で分類する: `fs.ErrNotExist` → `E_NOTFOUND`、`fs.ErrPermission` → `E_PERMISSION`、その他の FS エラー → `E_IO`、それ以外 → `E_LAYAT_FAILED`。Discover は entrypoint の不在を `fs.ErrNotExist` を包まずに返す（`E_LAYAT_FAILED`。`-f` のパス自体の不在は `fs.ErrNotExist` を包み `E_NOTFOUND`）。
- `Kind` は `--json` コードに影響させない。roots / build 段の失敗では `Kind = NotFound` も outturn 共通コードの `E_NOTFOUND` には写さず `E_LAYAT_BUILD` に分類する。
- エンベロープに生成器名は出さない（機械可読の契約面を生成器の実装事情から切り離す）。CLI が生成器名をフィールドや前置きとして足さないことを指し、`Message` が含む失敗した内部コマンド（`nix eval ...` 等）の一部として現れる分は対象外。

### 8. 内部コマンドの開示の分担

ADR-0031 §3 の `--debug` 開示は、主体を生成器に移す。ADR-0007 §3 の透明性（内部実行する nix コマンドの開示）は、`--help` の文面と `--debug` の逐次開示に分けて持つ。分担は次のとおり。

- **`--help` の透明性の文面は CLI が持つ**。既定の nix 生成器が実行するコマンドを示す静的な説明文で、失敗時の `Guidance` ではないため、§6 の「nix 固有の文字列判定・案内文・NotFound の前置文は nixgen に閉じる」の対象外。
- **実行中の逐次開示は生成器が書く**。CLI は `--debug` の有無を生成器へ渡し、生成器は実行する内部コマンドを生成器名付きで、渡された writer へ書く（何を実行するかを知っているのは生成器だけのため）。
- **失敗時の表示は CLI が組み立てる**。`--debug` のとき、CLI は人間向けの要約行に `generator.Error` の `Generator`（生成器名）を添える。失敗した内部コマンドは `Message` が含む（§6）。`--json` の `errors[].message` は `--debug` に左右されず、生成器名を前置きしない（§7）。

## 根拠

- **プロセス内インターフェイス**: 生成器は現状 nix 1 つで、外部プログラム規約（`layat-manifest-<lang>` 等）を先に決めると使われない仕様が固まる（YAGNI）。Go インターフェイスなら nix 実装の抽出と prebuilt・テストダブルの統一だけで、cmd 層から nix を剥がせる。
- **4 操作**: 既存の nix 呼び出しは全てこの 4 つに写る（`discoverEntrypoint` → Discover、`evalRoot` / `evalAllRoots` → Roots、`buildFunc` → Build、`dryBuildFunc` / `buildManifestStorePath` → DryBuild）。
- **事後条件の固定**: 世代バックエンドを変えずに生成器だけを一般化するには、出力側の制約を契約に書くしかない。根拠を明記しておけば、#406 が緩めるときに何が前提だったかを辿れる。
- **診断の分担**: 文字列判定と案内文は生成器ごとに異なり、CLI が持つと生成器を足すたびに CLI が変わる。逆に表示形式（人間向け / `--json` / `--debug`）は生成器によらないので CLI が持つ。

## 影響

- `cmd/layat/nix.go` を `internal/generator/nixgen` へ移し、`--manifest` 経路を prebuilt 実装へ置き換える（Issue #402）。診断契約の実装は Issue #403、選択機構は Issue #404。
- ADR-0006 / 0007 / 0023 / 0026 / 0031 / 0043 に改訂注記を足す。
- 要求: 生成器の契約（REQ-194e4209-d804-4a4b-a2b8-3d39c6c33729）・診断契約（REQ-7a2f1ecf-4675-45aa-80c0-a8fc58db9edd）を起票し、2 層構成（REQ-f4d7d4ab-fbdb-48c6-b29f-08dd88e72645）・experimental-features の前提（REQ-f9920c87-8551-4aa3-bf03-26fdf4191ed6）・内部コマンドの開示（REQ-4ffda99a-7062-4c00-915f-70b525cb215b）を生成器前提へ改訂する。パッケージ構成は DSG-25ad3cce-f921-4db2-aae0-c0263d4b8295。

## 棄却した代替案

- **外部プログラム規約（`layat-manifest-<lang>` を exec する）**: 2 つ目の実生成器が無い段階で wire 形式・発見規則・バージョニングを決めることになる。実生成器が現れたときに決める。
- **Roots を任意の能力にする**: §3 のとおり eval 先行 → flock → build の順序が壊れる。
- **Build の出力を任意のディレクトリにする**: `nix-env --set` が store 外を拒否するため世代コミットできない。緩めるなら世代バックエンドごと変える必要があり #406 の範囲。
- **`Kind` ごとに `--json` コードを分ける（`E_NOTFOUND` 等）**: 消費者にとって生成器の内部分類に依存するコードが増える。分類は `errors[].message` と人間向けの案内で足りる。
- **生成器の stderr も成功時は沈黙させる**: 評価器の警告・deprecation 通知が利用者に届かない現状の欠陥（Issue #396）を固定化する。

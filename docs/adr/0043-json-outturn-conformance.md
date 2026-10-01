---
id: "ADR-0043"
type: adr
name: "`--json` 機械可読出力を outturn 規約準拠にし、JSON 出力の outturn 準拠を恒常原則とする"
status: 採用
issues:
  - "yasunori0418/outturn#1"
origin: "nput の outturn 準拠化 grilling（2026-07-06）と、それを受けた outturn 側 grilling による outturn#1 の方針確定（batch エンベロープ + subject + §5 参照キー規約の 3 層化）。参照: outturn specVersion 1（yasunori0418/outturn）, outturn ADR-0013（mode 廃止・subject 常時必須）"
justifies:
  - "REQ-a5053191-1c6a-449b-9c5e-5ff49dc5aead"
  - "REQ-5c2e64c3-09a7-4ae8-b60c-4f1ccabce4fd"
  - "REQ-9341fa5d-836e-4023-af53-cc7d273438d1"
  - "REQ-2353259f-5878-452a-8e11-3445de69abc2"
  - "REQ-2a613337-7646-4ced-8807-e43bca18acf3"
  - "REQ-57137302-de29-4f71-a565-034cd5de080b"
  - "REQ-2ea19863-eaa2-466b-b1ed-3f56f6417c62"
  - "REQ-fa181aa6-29a2-48c3-ae07-cc1b9a3b0303"
  - "REQ-059eb4d5-63fb-4f8e-b705-11b5e2ed4ae5"
  - "TP-d3000054-42d9-4bac-912a-dd3abc38d3e9"
  - "TP-e7c25263-6d2d-4a37-8275-26906889d912"
  - "REQ-0a123b89-0399-4f76-b988-56a5f7e0becf"
revises:
  - "ADR-0023"
  - "ADR-0033"
references:
  - "ADR-0004"
  - "ADR-0018"
  - "ADR-0031"
  - "ADR-0042"
---
# ADR-0043: `--json` 機械可読出力を outturn 規約準拠にし、JSON 出力の outturn 準拠を恒常原則とする

- ステータス: 採用
- 日付: 2026-07-07
- 関連: ADR-0033, ADR-0023, ADR-0031, ADR-0018, ADR-0042, ADR-0004, ADR-0055, ADR-0056, `docs/concept.md`, `docs/design.md`, `docs/spec.md`, outturn specVersion 1（yasunori0418/outturn）, outturn ADR-0013（mode 廃止・subject 常時必須）, yasunori0418/outturn#1
- 改訂対象: ADR-0033 §1-3（独自エンベロープ `{"version":1,...}` を outturn エンベロープ準拠へ）/ ADR-0023 §2（「エラーは stdout に畳み込まず stderr 専有」を再改訂）。ストリーム規律の骨子（stdout=機械可読専有・warning/error 常時 stderr）と終了コード表 0/1/2 は不変
- 起点: nput の outturn 準拠化 grilling（2026-07-06）と、それを受けた outturn 側 grilling による outturn#1 の方針確定（batch エンベロープ + subject + §5 参照キー規約の 3 層化）。2026-07-07 に outturn 側の正式成果物（`spec/v1/spec.md`・`schema/v1/envelope.schema.json`・`go` module・`testdata/v1` 適合ベクタ）が確定し、**エンベロープは single / batch を問わずトップレベル常時 `results[]` に統一**された。さらに同日の outturn ADR-0013 で **`mode` 判別子は全廃・`SubjectResult.subject` は常時必須**へ改訂された（実行形態を切り替える判別子フィールドは持たない）。本 ADR は確定仕様に合わせて記述する

> **2026-09-05 改訂注記（ADR-0054）**: 本 ADR §6 が定めたツール別エラー / 警告コードの接頭辞
> `E_NPUT_*` / `W_NPUT_*` と、エンベロープの `tool.name` の値 `"nput"` は、**ツール名の改名により
> `E_LAYAT_*` / `W_LAYAT_*` と `"layat"` へ変わる**（outturn 仕様が `E_<TOOL>_<NAME>` を要求するため、改名は
> 選択ではなく必須）。これは `--json` 消費者にとって破壊的変更である。二層命名（共通コードは `E_LOCK` /
> `E_IO` 等をそのまま再利用）・参照キー 3 つ組 `(tool.name, subject, id)`・JSON 出力の outturn 準拠を恒常原則と
> する決定はいずれも不変。また改名予告の警告は **`--json` のエンベロープには入れず stderr にのみ出す**
> （機械可読の契約面にツールの都合の告知を混ぜない・→ ADR-0054 §6）。

> **2026-10-02 改訂注記（ADR-0055）**: 本 ADR §8 のツール別コード `E_*_BUILD`（現 `E_LAYAT_BUILD`）は存続し、意味を
> 「nix の失敗」から「**manifest 生成器の失敗**」へ読み替える。新しいコードは足さない。生成器の失敗分類
> （`generator.Error` の `Kind`）はコードに影響させず、`NotFound` も共通コード `E_NOTFOUND` には写さない。
> 読み替えの対象は prebuilt 以外の生成器の roots / build 段の失敗で、entrypoint 発見段の失敗と prebuilt の入力を
> 読めない失敗は従来の分類のまま。エンベロープに生成器名は出さない（→ ADR-0055 §7）。

> **2026-10-02 改訂注記（ADR-0056）**: 本 ADR §8 で予約した共通コード `E_INPUT` を初めて実装した。未知の生成器名・設定ファイルの
> 不正・フラグの組み合わせ不正（`--generator` + `--manifest`、`-f` / `--all` + `--manifest`）を `E_INPUT` に分類する。
> 既存の `-f` / `--all` + `--manifest` は `E_LAYAT_FAILED` から `E_INPUT` へ変わる（文面は不変・→ ADR-0056）。

## 背景

ADR-0033 は `--json` を「全サブコマンド一律のグローバルフラグ・stdout に単一 JSON オブジェクト（`{"version":1,"command":...}`）」として導入する決定をした。これは nput 単独の独自エンベロープであり、当時は外部規格が存在しなかった。

その後、n プレフィックスのツール群（nput / nboot / nwrap / nherd / nshadow / ncompose）が **stdout / stdin の JSON パイプで合成される** ことを前提にした共通規格 **outturn**（specVersion 1）が起こされた。エコシステムの北極星は「NixOS とは異なる Nix 版 Arch / Gentoo ——最小のコアとユーザーによる組み立て」であり、その実現手段が「**単一責務のツールを outturn 規格のパイプで合成する**」ことである（→ outturn `docs/ecosystem/overview.md`・本リポジトリ ADR-0004 の north-star）。nput はこのエコシステムで唯一 active なツールであり、ncompose が nput を含む各ツールの出力を固定順に合成・失敗時逆順 rollback する構想の**最初の適合実装**になる。

したがって nput の `--json` は、nput 単独の都合ではなく **outturn 規約の適合ツールとして** 出力すべきである。ADR-0033 の独自エンベロープはこの相互運用の前提を満たさない。

outturn#1 の grilling で outturn 側の仕様も確定した（batch エンベロープ・subject・§5 の参照キー規約 3 層化。詳細は yasunori0418/outturn#1）。これにより nput が outturn に準拠するための決定を本 ADR で確定する。

## 決定

### 1. 恒常原則: nput の全 JSON 出力は outturn 規約に準拠する（北極星要件）

- nput の `--json` 出力（**現在および将来のすべての機能**）は outturn specVersion 1 規約に準拠する。単発の `--json` 機能の決定ではなく、**JSON 出力の outturn 準拠を nput の恒常的な設計制約**とする。
- 今後の機能追加（`nput prune`・plan / verify・mkEnv 等）で機械可読出力を持つ場合、その出力は outturn エンベロープ規約に収める。ツール固有の情報は outturn の `info` 配下にのみ置く。
- 根拠は北極星: ncompose によるツール合成は「規格が契約（ツール間の会話は outturn 規約のみに依存）」で初めて成立する。nput が独自形状に逸脱すると合成の前提が崩れる。**outturn 準拠はエコシステム構築に向けた設計要件**であり、nput 内部の出力都合より優先する。
- concept.md（北極星節）・design.md（出力規約）・spec.md（出力ストリーム規律）に本原則を明記する（→ 影響）。

### 2. エンベロープ（常時 `results[]`・判別子なし）

- `--json` 指定時、stdout に **outturn エンベロープを 1 文書だけ**書く。トップレベルは `specVersion` / `tool{name,version}` / `command` / `status` / `dryRun` / `startedAt` / `finishedAt` / `errors[]` / **`results[]`**。命名は camelCase・時刻は RFC 3339。
- **single / batch を問わずトップレベルは常に `results[]`（`SubjectResult` の配列）**であり、単数の `result` はトップに存在しない（outturn §2）。`specVersion` / `tool` / `command` はトップに 1 度だけ置き、主体ごとに繰り返さない。
- `results[]` の各要素は `SubjectResult`＝`subject`（常時必須・§3）/ `status` / `startedAt` / `finishedAt` / `errors[]` / `result{items,changes,info}`。item / change / info は各 `SubjectResult.result` 配下に入る。
- **実行形態の判別子フィールド（旧 `mode`）は存在しない**（outturn ADR-0013 で全廃）。`results` の要素数は 0 以上で、起動の性質（単一 config か `--all` か）は形状を変えない。nput の挙動として単一 config 起動の `results` は高々 1 要素・`--all` は 0 以上になるが、これは形状契約ではなく、消費側が依存してよいのは常時 `results[]` の一様形のみ。
- `tool.version` は VERSION ファイル → ldflags 埋め込み（ADR-0042・`--version` 新設が前提）。
- outturn の `specVersion`（出力規格）・`manifest.json` の `schemaVersion`（engine 入力契約）・`tool.version`（nput リリース）は**独立**に進化する 3 つのバージョンとして扱う。

### 3. item identity（subject-scoped）

- item の `id` は outturn の `id = lowercase-hex(sha256(JCS(identity)))` で導出する。`identity = {kind, key}`。
- entry の identity は `kind="entry"`, `key={target}`（root 相対 target のみ）。**config 名は key に含めない**。
- outturn#1 §5（参照キー規約 3 層）に従い、id 値は subject を跨いで衝突してよく、consumer は **`(tool.name, subject, id)` の 3 つ組**で参照を解決する。`subject` は id 導出に関与しない弱い識別子。
- `subject` は各 `results[i]` 内に置く（`SubjectResult.subject:{name}`＝config 名）。**トップレベル `subject` は存在しない**。`subject` は単一 config / `--all` を問わず**全 `SubjectResult` で常時必須**（outturn ADR-0013・schema の `subjectResult.required` で強制）。これにより参照 3 つ組 `(tool.name, subject, id)` が常に揃う。

### 4. items / changes マッピングと reversible

- **items = フルインベントリ**: `apply` / `apply --dryrun` は manifest の全 entry を item として列挙する（変更の無い entry も `status:"success"`）。stale 除去された旧 entry も列挙する。
- **changes = 差分のみ**（outturn: noop を含めない）: place / copy 新規 → `add`、replace / recopy → `modify`、stale 除去 / reset → `remove`。
- **reversible**: symlink の add/modify/remove と copy の新規配置 = `true`（世代 rollback / 単純除去で戻せる）。copy の上書き（`--recopy`）と削除（`reset`）= `false`（copy は世代外で hash 追跡せず旧内容を復元できない）。
- 可逆性は `change.reversible` のみで表現し、`W_IRREVERSIBLE` 等の警告コードは付けない（outturn §4 が「consumer は `reversible:false` を警告として扱うべき」と定めており二重符号化になる・outturn#1 F）。

### 5. read-only 列挙は `result.info` インベントリ（id 導出 item にしない）

- `list-generations` の世代・`gitignore` の対象パスは **副作用の無い列挙**であり、`result.info` 配下のツール固有インベントリとして持つ（outturn#1 D/E）。id 導出 item にはしない（item は「処理単位の実行結果の記録」の定義のまま）。
  - `list-generations`: `info.generations = [{number, date, current}]`。世代番号はどの key にも入らない。
  - `gitignore`: `info.paths`（anchor 形 target）。**デフォルトの行指向 stdout 出力は不変**（`--json` は opt-in の第 2 契約・ADR-0033 の gitignore 二契約方針を継承）。
- これらのコマンドの `result.items` は空配列でよい。

### 6. エラーと終了コード

- エラーは **outturn エンベロープに構造化して載せる**。置き場は 2 層（outturn §2）: **トップレベル `errors[]` は主体列挙・解決の前段エラーのみ**（入力 parse 失敗・`specVersion` 不能・主体列挙自体の失敗）、**主体に紐づく全体エラー（その主体の build / lock 失敗等）は該当 `results[i].errors[]`**、item 起因は `item.error`。単一 config の実行でも主体起因エラーは `results[0].errors[]` に置きトップ `errors[]` には畳まない。**同時に stderr の人間向けテキスト（既存の op + 対象パス wrap 規約）も常時併存**させる（outturn §1 が stderr を診断チャネルとして許容）。これは ADR-0023 §2 / ADR-0033 §2 の「エラーは stdout に畳み込まず stderr 専有」を再改訂するもの。
- 終了コード表 0 / 1 / 2 は不変（POSIX・0=成功 / 非 0=失敗。1 = 一般エラー・`--all` 部分失敗、2 = `--dryrun` conflict は nput 内部の意味づけ）。outturn の `status` は exit 0 → `success` / exit 1・2 → `error` に連動する。outturn 消費側が依存してよいのは「0 ⇔ success / 非 0 ⇔ error」のみ。
- conflict は該当 entry を `item.status:"failed"` + `error.code:"E_NPUT_COLLISION"` で表し、1 件でもあれば `status:"error"`（`--dryrun` でも当該 item は failed）。

### 7. `--all` は複数 `SubjectResult` の列挙（形状は単一実行と同一）

- `apply --all` / `list-generations --all` / `gitignore --all` は `results[]` に config ごとの `SubjectResult` を列挙する。**形状は単一 config 実行と同一**（判別子フィールドは無い・outturn ADR-0013）で、対象が N=0 / 1 でも特別な形にはならない。「batch」は複数主体実行の非公式呼称としてのみ残る。
- `results[]` の各要素は `SubjectResult`（`subject` / `status` / `startedAt` / `finishedAt` / `errors[]` / `result`）。`specVersion` / `tool` / `command` は**トップレベルに 1 度だけ**置き、各 `SubjectResult` は持たない（切り出して単独 valid にはならない・outturn §2）。`subject={config名}` は各 `SubjectResult` で必須（常時必須・§3）。config 単位の build / lock 失敗（item 非依存）は該当 `results[i].errors[]` に置く。top の `errors[]` は主体列挙自体の失敗のみ。
- top の `status` は集約（sibling に 1 つでも error があれば error）。`results` が空（対象 0 件）でエラーが無ければ `success`。`--all --dryrun` の終了コード優先度（error 1 → conflict 2 → 0）は不変（ADR-0024）。
- **`gitignore --all --json` は cross-config dedup をしない**: 各 subject が自 config の paths を持ち、消費側が union + dedup して `.gitignore` を再構成する。一方 **テキスト既定出力は従来通り dedup + sort 済みの単一リスト**（ADR-0018 不変）。テキスト＝集約 / JSON＝per-config という非対称を受け入れる（JSON は「どの config 由来か」を保つ利点があり、機械側の dedup は自明）。

### 8. Go 依存・エラーコード・実装 gate

- `github.com/yasunori0418/outturn/go`（Envelope 汎用型 + `DeriveID`）を依存に追加する。ADR-0033 が課した「stdlib-only で emit」制約は、outturn/go 自体が stdlib-only の規格参照実装であること、および id 導出（JCS + sha256）を規格実装と共有して適合ベクタ（`id-vectors.json`）との乖離リスクを下げることを理由に緩和する。CLI 出力契約は cmd 層に閉じ、`engine.Result` 等を直接 marshal せず DTO 経由で outturn 型へ詰め替える。
- エラーコード: ツール別 `E_NPUT_COLLISION` / `E_NPUT_BUILD`・警告 `W_NPUT_FOREIGN_SYMLINK`（foreign symlink・outturn §3 の例と揃える）。共通コードは `E_LOCK` / `E_IO` / `E_NOTFOUND` / `E_PERMISSION` を再利用し、入力 parse 失敗は `E_INPUT`・`specVersion` 不能は `E_SPEC_VERSION` を用いる（outturn §6 の二層命名・共通レジストリ）。
- `reset --json` は破壊的操作の確認を機械消費で扱えないため **`--yes` を必須**とし、無ければ prompt せず `status:"error"`・非 0 で fail fast する。
- **実装の着手**は outturn の正式成果物（`spec/v1/spec.md`・`schema/v1/envelope.schema.json`〔single / batch 統合の単一 schema〕・`go` module〔`Envelope[TItem,TChange,TInfo]` 汎用型・`DeriveID`・`Subject` / `SubjectResult` 型〕・`testdata/v1/id-vectors.json` 適合ベクタ）の完成、および ADR-0042 の VERSION + `--version` を前提とする。本 ADR は決定の記録であり、実装は前提成果物の完成後に #130 以降で進める。

> **2026-07-17 追記**: 本節の `Envelope[TItem,TChange,TInfo]`（3 型パラメータ）は outturn 確定版の `go` module で `Envelope[TItem,TChange,TInfo,TEnvInfo]`（4 型パラメータ）に更新された。トップレベル `info`（実行全体・主体に紐づかないツール固有情報。§2 / outturn ADR-0018）用の `TEnvInfo` が、主体ごとの `result.info` 用 `TInfo` から分離されたことによる（`Result` 側は `[TItem,TChange,TInfo]` の 3 パラメータのまま）。sub-issue #126 / #130 は既に 4 型パラメータで記述済み。決定の骨子は不変で、実装は 4 型パラメータでインスタンス化する。

## 根拠

- **恒常原則にする理由**: outturn 準拠を「今回の `--json` 機能」に閉じた決定にすると、将来機能が独自形状に逸れる余地が残り、ncompose 合成の前提（規格が契約）が崩れる。北極星に向けては「全 JSON 出力が outturn 準拠」を制約として明文化しておく必要がある。
- **独自エンベロープを捨てる理由**: nput 単独の `{"version":1,...}` は outturn 消費側（ncompose 等）が解釈できない。エコシステムの相互運用は規格準拠が唯一の前提。
- **エラーをエンベロープに載せる理由（ADR-0033 §2 の反転）**: outturn はエラーを規格の一部（`errors[]` / `item.error`）として持つ。適合するには畳み込みが必要。stderr テキストは診断チャネルとして残すため人間向けの可読性は失わない。
- **read-only 列挙を info に置く理由**: 世代一覧・gitignore パスは実行結果ではなく列挙で、id 安定性の機構（§5）を要しない。item 化すると世代番号の key 問題等の無理が生じる。info インベントリが素直。
- **`--all` でも形状を変えない理由**: 件数や起動形態で形状が変わると消費側が出力形状を予測できない。outturn は当初「起動の性質で `mode` が決まる」としたが、容器が常時 `results[]` に統一された結果 `mode` は件数・subject 有無との二重管理に堕し、outturn ADR-0013 で全廃された。nput は常時一様な `results[]` 形のみに依存する。

## 影響

- **`docs/concept.md`**: 北極星節に「nput は outturn 規約でエコシステムに接続し、JSON 出力は outturn 準拠」を追記。
- **`docs/design.md`**: 出力・終了コード規約節の「`--json` は将来送り」を「outturn 準拠（→ ADR-0043）」へ更新。
- **`docs/spec.md`**: 出力ストリーム規律節に outturn 準拠の `--json` 契約（エンベロープ / エラーはエンベロープ + stderr 併存）を追記し、「`--json` は MVP では持たない」注記を削除。各コマンドの JSON ペイロード詳細は #130 以降で追記。
- **ADR-0033 / ADR-0023**: 本 ADR で改訂した箇所に blockquote 注記を書き戻す（同一 PR）。
- **実装（`cmd/nput/` ほか）**: outturn/go 依存追加、エンベロープ emit、engine の full-inventory 化、レポート系の data-first 化。詳細は #130 / #131 / #132 / #164。

## 棄却した代替案

- **ADR-0033 の独自エンベロープを維持し outturn は参照に留める**: 相互運用の前提を満たさず、北極星（ncompose 合成）に到達できない。
- **outturn 準拠を今回の `--json` 機能に閉じた決定にする**: 将来機能の逸脱余地が残る。恒常原則として明文化する方を採る。
- **エラーを stdout の outturn エンベロープに載せず stderr 専有のまま**（ADR-0033 §2 維持）: outturn 適合を満たさない。
- **read-only 列挙を id 導出 item にする**: 世代番号の key 問題等の無理が生じ、outturn#1 で却下済み。
- **`W_IRREVERSIBLE` 警告コードの追加**: `change.reversible` との二重符号化。outturn#1 で却下済み。

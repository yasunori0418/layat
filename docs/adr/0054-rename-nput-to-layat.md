---
id: "ADR-0054"
type: adr
name: "nput を layat へ改名し、2 週間の改名予告期間を挟む"
status: 採用
issues:
  - "#387"
  - "#385"
  - "#386"
origin: "Issue #385（epic）の grilling セッション（2026-09-04）で確定。改名の由来検討は 2026-08-25 の命名セッションが起点"
justifies:
  - "REQ-2aa3abbc-90b2-486e-92de-d785554bdeb3"
  - "REQ-5dd5a4e9-6162-4fa5-b295-66844f5a4f3b"
  - "REQ-c50df875-2cb0-4e72-8a21-858359a11cae"
  - "REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10"
  - "REQ-8085f194-c903-4ecb-abd8-c719fe7b3292"
  - "DSG-16373ec2-3496-4b12-b3b1-ef74e0435b58"
revises:
  - "ADR-0007"
  - "ADR-0024"
  - "ADR-0025"
  - "ADR-0029"
  - "ADR-0042"
  - "ADR-0043"
  - "ADR-0045"
references:
  - "ADR-0015"
  - "ADR-0053"
---
# ADR-0054: nput を layat へ改名し、2 週間の改名予告期間を挟む

- ステータス: 採用
- 日付: 2026-09-05
- 関連: ADR-0007（`nput` CLI 一次 UX・`nput.<name>` 名前空間・`nput init`）, ADR-0024（HM profile dir レイアウト）, ADR-0025（`nput init` の固定 ref・profile 専用ディレクトリ）, ADR-0029（flake-parts `name = "nput"`）, ADR-0042（`nput --version` の出力書式）, ADR-0043（`E_NPUT_*` / `tool.name`）, ADR-0045（`--backup` の既定 suffix `nput-backup`）, ADR-0015（foreign symlink の後勝ち置換）, ADR-0053（ADR のみ連番を維持する ID 規約）
- 改訂対象: ADR-0007 §3 / §4 / §6・ADR-0024 §2・ADR-0025 §3 / §4・ADR-0029・ADR-0042・ADR-0043 §6・ADR-0045 §1 が固定した「nput」という字句（CLI 名・名前空間・on-disk 名・エラーコード接頭辞・既定 suffix・固定 flake ref）を、決定の内容を変えずに「layat」へ改名する
- 起点: Issue #385（epic）の grilling セッション（2026-09-04）で確定。改名の由来検討は 2026-08-25 の命名セッションが起点

## 背景

nput の「n」は **nix** を指していた。設計の中心が「nix 側で規約としての manifest を構築し、engine 側で副作用（ファイル配置）を実行する」ことにあり、nix store 経由でハッシュ固定されたファイルを配置することが主用途だと見込んでいたためである。

しかし実装が進むにつれ、この前提は 2 点で崩れた。

1. **manifest の生成は nix でなくてもよい**。engine（`internal/`）が受け取る契約は `manifest.json` 1 本（→ ADR-0006）で、これを誰が生成したかに engine は関知しない。`lib/` は nixpkgs のみに依存する純関数群だが、engine から見れば manifest の生成系のひとつでしかない。実際 `--manifest` 経路（→ ADR-0026）は entrypoint を経由せず、ビルド済み link-farm を直接渡す。
2. **作者自身の運用が nix store 経由ではない**。dotfiles の配置には `mkOutOfStoreSymlink`（→ ADR-0001）による可変 symlink を使っており、store へのコピーを介していない。

結果として「n」が指すものが曖昧になった。ツールが解決しているペインは「フェッチ済みの内容を、manifest の通りに、root 相対の target へ置く」ことであって、その内容を誰がどうフェッチしたかではない。**nix に縛られない、解決するペインを体現する名前**へ改める。

改名は破壊的変更（module path・CLI 名・オプション名前空間・JSON エラーコード・on-disk 名がすべて変わる）である。VERSION 0.1.0 の実装フェーズで利用者は作者を含む少数だが、無告知の破壊は避ける。

## 決定

### 1. 新名は `layat`

**layat** = "**lay** \<src\> **at** \<target\>" の圧縮造語。manifest の通りに src を target へ置く、というツールの動作を一語で自己記述する。

`chroot`（change root）・`getopt`（get options）と同じ、**動詞句の圧縮**という UNIX コマンドの伝統的造語法に従う。5 文字、発音 "lay-at"、`layat apply skills` / `layat reset` として自然に読める。

### 2. 命名条件

エコシステム全体（本ツールに限らない）に適用する条件を次に固定する。

1. **日本語由来の語は避ける**。一度検討したが方針転換で除外した（下の「棄却した候補」の日本語ローマ字系を参照）
2. **既存英単語に限らず、造語・アルファベット表記の他言語も可**
3. **UNIX 哲学に沿った短いコマンド名**（3〜6 字を目安）
4. **既存プロダクトと被らないこと。必ず調査で確認する**

条件 4 の検証は次の 4 レジストリで行う（→ Issue #386 が手順と実施記録を持つ）。

| レジストリ | 照会 | 合格条件 |
|---|---|---|
| GitHub Search API | `/search/repositories?q=<name>+in:name` | `name` の完全一致が 0 件（部分文字列ヒットはノイズとして許容）|
| npm | `https://registry.npmjs.org/<name>` | 404 |
| crates.io | `https://crates.io/api/v1/crates/<name>`（**User-Agent ヘッダ必須**）| 未登録 |
| PyPI | `https://pypi.org/pypi/<name>/json` | 404 |

`layat` は 2026-08-25 の初回調査・2026-09-05 の再確認（→ #386）ともに 4 レジストリすべてで空きだった。GitHub の部分一致 69 件は全て LayaAir ゲームエンジン系（`layaTree` ★17 等）と無関係な人名・テスト用リポジトリで、完全一致は 0 件。

### 3. 棄却した候補

#### 実英単語系（既存プロダクトが占有）

| 候補 | 棄却理由 |
|---|---|
| billet | ブラジルの boleto 決済 SDK 群が "billet" タグを占有。Billetto / BilletWeb 等チケット会社も存在 |
| stow | GNU Stow と同一ドメインで完全衝突 |
| mise | mise en place は思想として適合するが jdx/mise が同発想で先取り済み |
| emplace | 語義は最適だが tversteeg/emplace（隣接領域のツール）が既存 |
| berth | 同名の Docker / SSH 系ツール複数 + crates 使用済み |
| layup | 60★ のページビルダーが既存 |
| belay | 273★ の Python ライブラリが既存 |
| inlay | LSP の inlay-hints 関連ツール群と衝突 |
| prelay | npm 使用済み |
| lath / dowel / tenon / cleat / batten / plinth / socle / sconce | 建築・木工メタファーは配置の語義に適合するが、全て既存ソフトウェアあり |
| situ / posto / lieu / sett / pone / sted / setl / stel / fixa / placet | 既存プロダクトまたは大量の検索ノイズ（SETL 言語・Stellarium・STED 顕微鏡等）|

#### 日本語ローマ字系（方針転換で除外）

`haizen`（配膳）・`suetsuke`（据付）は全レジストリ完全に空きで有力だったが、命名条件 1（日本語由来を避ける）により除外した。

#### 造語系（小衝突または語感の問題）

| 候補 | 棄却理由 |
|---|---|
| layt | 空きだが "late" と同音で口頭伝達に難がある |
| layn | 3★ の同名レイアウトエンジン（laynjs/layn）と完全一致 |
| verlay | 12★ の同名 Vercel ツールあり |
| theto | 6★ の同名あり |
| layto | Professor Layton と typo 距離 1 で検索性が最悪 |
| layon / layin | 空きに近いが、意味の乗りが弱い |
| laymap / layset | 空き。次点として悪くないが `layat` の自己記述性に劣る |
| plax / pono / okiba / enlay / posa | 既存プロダクト・npm スコープ占有等 |

### 4. 保持候補（衝突時の差し替え先・将来の別ツール名に流用可）

改名当日の再確認（→ #386 の 2 回目）で `layat` が先取りされていた場合、次の順で差し替え、本 ADR を改訂する。

- **laydo**（5 字）: laydown plan（配置計画）+ do。GitHub 完全一致 0 件・全レジストリ空き。`sudo` と同韻律。弱点は "Play-Doh" 連想
- **setz**（4 字）: 独語 setzen の命令形「置け」。完全一致プロダクト無し・npm / crates 空き（PyPI のみ使用済み）。弱点は Setzer（422★ の LaTeX エディタ）との検索混線と、`set` との typo 距離 1

### 5. 改名の適用範囲

改名は**機械的な全置換**として 1 PR に集約する（→ Issue #388）。旧名が残ってよいのは、**旧名を旧名として言及している箇所**だけである。

改名対象（網羅ではなく類型）:

- **Go**: module path `github.com/yasunori0418/layat`・`cmd/layat`・cobra `Use`・envelope の `tool.name`・エラー接頭辞 `"layat: "`・環境変数 `LAYAT_TEMPLATE_REF`・固定 flake ref `github:yasunori0418/layat`
- **JSON 契約**: `E_NPUT_*` / `W_NPUT_*` → `E_LAYAT_*` / `W_LAYAT_*`（niface 仕様が `E_<TOOL>_<NAME>` を要求するため改名は必須。→ ADR-0043 §6）
- **Nix**: `packages.layat`・`mainProgram`・flake 出力 `layat.<system>.<name>`・flake-parts `name = "layat"`・`options.layat.*`・`home.activation.layat`・`_layatMarker` / `layatSrc` / `layatRoot` などの内部識別子
- **on-disk**: 状態ディレクトリ `<state>/nix/profiles/layat/`・backup suffix `layat-backup`・`.layat-recopy-aside`・manifest derivation 名 `layat-manifest`
- **docs / CI / templates / tests**: 生きた文書と item 本文、workflow、issue テンプレ、starter テンプレ

置換の**除外**:

- **既存 ADR（`docs/adr/NNNN-*.md`）の本文**。ADR は決定が下された時点の記録であり、当時の名前で書かれていることに意味がある。改訂注記 blockquote の追記は通常運用どおり行う（→ `docs/adr/README.md`）
- **README の改名予告バナーと "Migrating from nput" 節**。旧名を旧名として案内する箇所である
- **Issue タイトル・コミット履歴**。編集不能、または編集すると既存の参照が壊れる

sara の item ID はツール名を含まない（フル UUIDv4・→ ADR-0053）ため、**改名による ID の変動は生じない**。

### 6. 互換シムは提供せず、代わりに 2 週間の改名予告期間を置く

旧名でのエイリアス・互換 module path・旧オプション名の `mkRenamedOptionModule` などの**互換シムは一切提供しない**。VERSION 0.1.0 の実装フェーズで利用者が限られ、シムの維持コストが便益を上回るためである。

代わりに、改名 PR（#388）のマージまでに **2 週間以上の改名予告期間**を置く。予告は次の 3 層で出す。

| 層 | 実装 | 届く相手 |
|---|---|---|
| `lib/manifest.nix` の `mkManifest` | `lib.warn` | `lib` を直接使う利用者（モジュールを経由しない経路）|
| `modules/common.nix` | `config.warnings` | home-manager / NixOS / nix-darwin モジュール利用者 |
| `cmd/nput/main.go` の `PersistentPreRun` | stderr 1 行 | CLI 利用者（全サブコマンド共通）|

予告の設計規約:

- **文面は英語**とする。README・`--help` と同じ一次言語に揃える
- **`--json` の envelope には入れない**。envelope は niface 仕様に適合した機械可読の契約面であり、ツールの都合の告知を混ぜない（→ ADR-0043）。`--json` 指定時も予告は stderr にだけ出し、stdout の JSON を汚さない
- **日付は "will be renamed on or after YYYY-MM-DD" の下限表記**にする。改名の着手条件は「予告 + 14 日」と「prune epic #127 の完了」の遅い方であり、上限を約束すると #127 が遅れたときに警告文が嘘になる
- **文面には旧名で留まる選択肢を含める**。`github:yasunori0418/nput/legacy-nput` への pin を案内する（次項）
- CLI 層の予告は cobra の補完リクエスト（`__complete`）では出さない。補完スクリプトの出力を汚さないため

### 7. 旧名で留まる利用者向けに annotated tag `legacy-nput` を打つ

改名予告 PR（#387）の**マージコミットに annotated tag `legacy-nput` を打つ**。GitHub Release は作らない。

- **GitHub Release ではなく git のタグにする**理由: VERSION 0.1.0 の実装フェーズであり、リリースを名乗れる成熟度に達していない。タグは「この時点の状態を指す名前」以上の意味を持たない
- **非 semver 名にする**理由: release workflow（→ ADR-0042）が `v0.1.0` 形式のタグを扱うため、名前空間を分ける
- 旧名で留まりたい利用者は flake input を `github:yasunori0418/nput/legacy-nput` へ pin する。GitHub のリポジトリ改名は旧 URL のリダイレクトを維持するため、改名後もこの参照は解決する（#388 の当日チェックリストで実地確認する）

### 8. 世代（状態ディレクトリ）はバイナリでは移行しない

改名後のバイナリは `<state>/nix/profiles/layat/` を使う。**旧 `<state>/nix/profiles/nput/` からの自動移行は行わない**。

自動移行を行わない理由:

- 移行対象は nix profile（`nix-env --profile` が管理する世代リンク群）であり、単純な `mv` では gcroot の登録が旧パスのまま残る。正しく移行するには `nix-store --add-root` による gcroot の再登録が要る。これをバイナリ内で行うと、失敗時の巻き戻しが profile 世代の整合性に直接触れる危険な処理になる
- 移行しなくても壊れない。改名後の初回 apply は新しい状態ディレクトリで世代 1 から始まり、既存の配置物は次項の挙動で収束する

代わりに次を提供する。

- **README の "Migrating from nput" 節**に手動の移行スクリプト（`mv` + `find ... nix-store --add-root --indirect -r` による gcroot 再登録）と、移行せず削除する場合の手順を載せる（→ Issue #393）
- **改名後のバイナリが旧状態ディレクトリを検出したら stderr に 1 行のヒントを出す**（移行はしない・抑止フラグも持たない・→ Issue #389）。この検出ヒントと README の移行節は、v0.1.0 リリース後の次 minor で撤去する（→ Issue #392）

### 9. 改名後の初回 `layat apply` の挙動

移行ガイドが書くべき事実として、次を確定させる。いずれも既存の決定から導かれるもので、新たな決定ではない。

| 対象 | 挙動 | 根拠 |
|---|---|---|
| 旧 nput が置いた symlink | 記録外（foreign）扱い → **後勝ちで上書き** + `W_LAYAT_FOREIGN_SYMLINK` 警告。失敗しない | ADR-0015 |
| copy entry | 実ファイルが既にあるため foreign skip。上書きには `--backup` が要る | ADR-0045 |
| 旧世代にのみあった entry | 前世代 manifest が無いため **stale 検出されず残る** | ADR-0002, ADR-0006 |
| `<state>/nix/profiles/nput/` と `<target>.nput-backup` | 残る。手で消す | §8 |

利用者側で書き換えが要るもの: flake input URL・`nput.*` オプション・`home.activation.nput`・`#nput.<system>.<name>`・`--json` 消費者の `E_NPUT_*` / `W_NPUT_*` と `tool.name`。

## 影響

- **`docs/adr/README.md`**: 改訂注記の追記対象として ADR-0007 / 0024 / 0025 / 0029 / 0042 / 0043 / 0045 の 7 本に本 ADR の blockquote を積む（本 ADR と同じ変更の中で実施）
- **`README.md` / `README.ja.md`**: 改名予告バナーと "Migrating from nput" 節を追加する（#387）。改名後に自己記述文へ書き直す（#390）
- **`docs/concept.md` / `docs/glossary.md` / `docs/glossary.ja.md`**: 命名の由来と本 ADR への索引を追加する（#390）
- **`lib/` / `modules/` / `cmd/`**: 予告の 3 層を追加する（#387）。改名当日に予告を撤去し、機械的全置換を行う（#388）
- **隣接リポジトリ**: dotfiles / niface の dev shell / skills の dev shell の flake input と参照を差し替える（#391）。niface の仕様側（`E_<TOOL>_` の例示・conformance fixture・ecosystem docs）は**対象外**とする

## 棄却した案

- **互換シムを提供する**（旧 module path のエイリアス・旧オプション名の rename module）: 利用者の書き換えを不要にできるが、実装フェーズの少数利用者に対して維持コストが見合わない。破壊的変更を正面から出せる最後の時期であり、シムを入れると撤去のための追加の破壊的変更が将来必要になる
- **改名せず "n" の意味を再定義する**（例: n = "node/名前" と読み替える）: 変更コストはゼロだが、名前が何も自己記述しない状態が残る。ツールが解決するペインを体現する名前へ変えるという動機を満たさない
- **状態ディレクトリをバイナリで自動移行する**: 利用者の手間は減るが、gcroot の再登録を伴う危険な処理をバイナリへ入れることになる（§8）。撤去予定の一時コードとしては複雑度が高すぎる
- **`nput` を残して新ツールとして `layat` を作る**: 旧利用者を壊さないが、同一実装が 2 つの名前で流通する状態は最も混乱を招く。実装フェーズで分岐を作る理由がない
- **改名を v0.1.0 リリース後に行う**: リリース済みの名前を変える方が影響が大きい。初リリースを新名 `layat` で出すため、改名を先に済ませる（VERSION は 0.1.0 のまま据え置く）

# 用語集（日本語版）

layat の正準（canonical）英語用語を日本語で対訳したもの。README・コードコメント・コマンド出力で概念に言及するときは、ここで定義した **canonical 用語**（見出しの英語表記）を使い、列挙した別名を避けて、プロジェクト全体で表記をそろえる。正準の綴り自体は英語版 `docs/glossary.md` が一次（authoritative）の参照先。

各エントリは canonical な綴りを 1 つ固定する。定義は意図的に短い。詳細な根拠は `docs/requirements/` の requirement item（索引は `docs/spec.md`）と `docs/adr/`（設計判断）を参照する。英語版（一次）は `docs/glossary.md` にある。

## 配置の中心抽象

### placement primitive
layat のコア。`root` 相対の `target` に Nix store のパスを配置する純粋関数。モジュール抽象の裏に隠さず、ユーザーが直接合成して使う。
- **Avoid**: 「placement framework（配置フレームワーク）」「configuration management（設定管理）」（layat は設定を生成しない）。

### engine
配置（ネイティブなファイルシステム操作）と stale 除去の両方を一手に所有する配置コア。`manifest.json` を入力に取り、外部コマンドは `git`（toplevel）のみを叩く。世代の記録・切替は **generation backend（世代バックエンド）** に委ねる。**layat CLI** が駆動する Go ライブラリとして実装される。config ごとに bash スクリプトを生成しない。ここでの「ライブラリ」は `internal/` 配下のバイナリ内部の層分離を指し、公開 import 可能な再利用モジュールではない——安定面は `manifest.json` 契約に閉じる。エンジンは stdlib-only。
- **Avoid**: 「per-config generated bash script（config ごとに生成される bash スクリプト）」「per-layer placement logic（層ごとの配置ロジック）」「a single flat implementation fused with the CLI（CLI と一体の平らな単一実装）」「importing the engine as a public Go module（engine を公開 Go モジュールとして import する）」。

### layat
ツールそのものの名前。"**lay** \<src\> **at** \<target\>"（manifest の言うとおりに src を target へ置く）の圧縮造語。`chroot`（change root）・`getopt`（get options）と同じ、動詞句をコマンド名へ圧縮する UNIX の伝統的な造語法に従う。旧名は **nput** で、その「n」は *nix* を指していたが、engine の契約が生成者を問わない `manifest.json` 1 本へ収束した時点でツールを記述しなくなった（→ [ADR-0054](adr/0054-rename-nput-to-layat.md)）。文頭でも小文字で表記する。
- **Avoid**: 移行文脈の旧名以外で「nput」を使うこと、「Layat」「LayAt」「layAt」と書くこと、名前を nix 固有のものと読むこと。

### layat CLI
ユーザーが直接触れる一次 UX。`PATH` 上の `packages.layat` バイナリ。**entrypoint** を発見し、**generator（生成器）**（既定 `nix`）経由で named manifest を取得し、エンジンに配置させる。サブコマンドは `apply [<name>]`・`apply --all`・`reset`・`rollback`・`list-generations`・`gitignore`・`prune`・`init`。
- **Avoid**: config ごとの `nix run .#x` ラッパーを一次 UX と説明すること、`apply` を「常に entrypoint を build する」と説明すること（ビルド済み link-farm は `--manifest` で適用できる）。

### generator
**layat CLI** が manifest を得る相手（生成器）。契約は `internal/generator` に置く 4 操作で、Discover（entrypoint の発見）・Roots（build せずに root の種別と root を返す）・Build（世代コミット可能な link-farm = store path を pending に張る）・DryBuild（gcroot を張らずに store path を返す）。nix 実装（nixgen・`internal/generator/nixgen`）が既定。生成器は明示指定でのみ選ばれ、`--generator`・`LAYAT_GENERATOR`・`layat.toml`・`$XDG_CONFIG_HOME/layat/config.toml`・既定 `nix` の順に解決する。エンジンはどの生成器が manifest を作ったかに関知しない（→ [ADR-0055](adr/0055-manifest-generator-contract.md), [ADR-0056](adr/0056-generator-selection-and-config.md)）。
- **Avoid**: CLI 自身が `nix build` / `nix eval` を回すと説明すること（回すのは nix 生成器）、置かれているファイルから生成器が推測されると説明すること。

### prebuilt
`apply --manifest <link-farm>` の実体である生成器。Discover は与えられたパス、Roots はその link-farm の `manifest.json` を読み、Build / DryBuild はそのパスを返す。`--generator` の選択肢には出ない（→ [ADR-0055](adr/0055-manifest-generator-contract.md)）。
- **Avoid**: `--generator` の値として挙げること、`--manifest` を生成器とは別の取得経路と説明すること。

### generation backend
**engine** が世代の記録・一覧・切替を委ねる相手（世代バックエンド）。契約は `internal/generation` に置く 5 操作で、入出力はパスと整数のみ。Commit（link-farm を新世代として記録し profile を向ける）・List（世代の一覧と current）・Switch（既存の世代へ profile を向ける）・UserBase（user state の profileDir の基底）・SystemBase（system mode の profileDir の基底）。nix 実装（nixprofile・`internal/generation/nixprofile`）が既定かつ唯一の実装で、`nix-env` を叩くのはここだけ（基底は `<state>/nix/profiles/layat` と `/nix/var/nix/profiles/layat`）。CLI は常に nixprofile を注入し、engine はバックエンドが渡されなければ error にする。エンジンは誰が世代を管理するかに関知しない。profileDir 内部のレイアウト（`profile` / `profile-<N>-link` / `.pending`）は契約の固定部分で、エンジンが直接読む。バックエンドを選ぶフラグ・環境変数・`layat.toml` キーも、基底を上書きする設定キーも持たない（→ [ADR-0057](adr/0057-generation-backend-contract.md)）。
- **Avoid**: エンジンが `nix-env` を叩くと説明すること、バックエンドの選択手段や基底の設定キーがあると説明すること。

### entrypoint
layat CLI が読む Nix の config ファイル。`flake.nix` / `shell.nix` / `default.nix` のいずれか。`layat.<name>` に named manifest を公開する。config は依然として Nix で書かれ、`nix build` で評価される。
- **Avoid**: 「layat が CWD から config の内容を発見する」と説明すること——発見するのは entrypoint *ファイル* であり、config は Nix 評価で確定する。

### module
standalone・home-manager・NixOS / nix-darwin モジュール・devShell の `shellHook` といった統合層。エンジンを起動する配線（wiring）に徹し、自身ではファイルを配置せず、`home.file` や `systemd.tmpfiles` へ翻訳することもない。
- **Avoid**: 「the module places files（モジュールがファイルを配置する）」「the module translates to native mechanisms（モジュールがネイティブ機構へ翻訳する）」。

## 配置の入出力

### entry / entries
配置定義。`entries` は **target** をキーとする attrset で、各値が entry（`{ src; subpath?; target?; method?; }`）。属性キーが識別子であり、既定の `target` を与える。手動の `name` フィールドは持たない。識別子の一意性は Nix の attrset キーで native に担保される。
- **Avoid**: 「file entry（ファイルエントリ）」（配置元はディレクトリのこともある）、entry に `name` フィールドを持たせること、`entries` を `{ name; … }` のリストと呼ぶこと（旧形）。

### src
**entry** の配置元（どの store パス / リポジトリか）。デフォルトは Nix ストアへの **store link**。out-of-store は明示マーカーで opt-in する。**subpath** とは直交する別概念であり、`src` を `source` の短縮形と読まない。
- **Avoid**: **subpath** と混同すること、`source` と呼ぶこと。

### subpath
**entry** で **src** 内のどのパスを取り出すかを選ぶ相対パス。file / dir 両対応。デフォルト `"."`。リポジトリ全体を選ぶ canonical な方法は **subpath を省略**すること（`subpath = "."` は明示形）。
- **Avoid**: 旧名 `source` / `dir` を使うこと、**src** と混同すること、「全体を表すには専用のトークン / marker が要る」と考えること（省略で表せる）。

### target
**entry** の配置先。**root** からの相対パスで指定する。`entries` の属性キーが既定の `target` であり、entry の identity（stale 除去の diff キーであり、一意性のキー）でもある。

### root
配置の基準パス。公開 API の `root` 引数で**明示的に**選ぶ（**暗黙のデフォルトは持たない**）。型は `string`（絶対パス・評価時に固定）と `marker`（実行時に解決）の union。マーカーは **projectRoot** / **homeRoot** / **systemRoot**。各マーカーは適用形 `homeRoot { subdir = "..."; }` も受け付ける。適用形でも root は動かず、全 target へ `<subdir>/` を前置するのと等価な評価時の糖衣である。
- **Avoid**: 「`$HOME` is fixed（`$HOME` 固定）」「defaults to home mode（既定は home mode）」「the default when `root` is omitted（`root` 省略時の既定）」と説明すること、マーカーを「パス文字列を返す糖衣（sugar）」と説明すること（実行時解決の種別であって、評価時にパスへ展開しない）、適用形を「root actually resolves to the subdir（root が実際に subdir へ解決される）」と説明すること（root と `rootKind` は単体形と同じで、変わるのは target だけ）。

## 配置モード

### home mode
**root** = `$HOME` の配置モード。`homeRoot` マーカーで明示選択する。standalone も home-manager 等の module もこのモードを使う。世代を毎回コミットし、`--rollback` をユーザーに公開する。
- **Avoid**: 「standalone-only（standalone 専用）」「the default when `root` is omitted（`root` 省略時の既定）」と説明すること。

### project mode
**root** = プロジェクトルートの配置モード。`projectRoot` マーカーで選ぶ。配置物は **ephemeral placement**（コミットしない）であり、世代・rollback はユーザーに公開しない。root は git toplevel に解決される。
- **Avoid**: 「relative to CWD（CWD 相対）」「relative to the config file（設定ファイル相対）」と説明すること。

### system mode
**root** = `/` の配置モード。`systemRoot` マーカーで選ぶ。distro 構想（root = `/`）に使う。

### projectRoot
**project mode** を選ぶ root マーカー。実行時に git toplevel を **root** に解決する（`--root` で上書き可）。`homeRoot` / `systemRoot` と並ぶ root マーカーの一つで、`mkOutOfStoreSymlink` と同じ「マーカーを渡して挙動を opt-in する」パターンに従う。適用形 `projectRoot { subdir = "..."; }` も受け付ける（意味は **root** の項）。
- **Avoid**: 設定ファイルの場所を指すと解釈すること。

### homeRoot
**home mode** を選ぶ root マーカー。実行時に `$HOME` を **root** に解決する。従来は暗黙だった `$HOME` の既定を明示マーカーへ昇格したもの。適用形 `homeRoot { subdir = ".config"; }` は root を `$HOME` のまま、全 target を `.config/` 配下へ置く。
- **Avoid**: 適用形の root が `$HOME/.config` へ解決されると説明すること。

### systemRoot
**system mode**（root = `/`）を選ぶ root マーカー。ADR-0004 が「将来の絶対パス文字列 seam」と呼んでいたものを第一級のマーカーへ昇格したもの。適用形 `systemRoot { subdir = "..."; }` も受け付ける（意味は **root** の項）。

### ephemeral placement
**project mode** における配置物の性質。クローンごとに再生成される前提で、プロジェクトにコミットされない。ゆえに activation は git 状態に干渉しない。
- **Avoid**: 「vendoring」や「成果物をコミットする配置」と混同すること。

## 配置の種別

### store link
コア・デフォルトの配置。配置先が Nix store パスである symlink。再現性を担保するデフォルト経路。「unification（統一）」とは、ストアをデフォルト / コアにし、out-of-store を明示的な例外へ降格することを意味する。
- **Avoid**: out-of-store symlink と混同すること、「copy（コピー）」と呼ぶこと。

### out-of-store symlink
ローカルパスへのライブ symlink。明示マーカーでのみ opt-in する。明示的な退避路（escape hatch）であり、第一級機能ではない。マーカーは 2 つある。`layat.lib.mkOutOfStoreSymlink "/abs/path"` はローカルの絶対パスを指す（開発中の dotfiles をライブ編集する用途）。
リポジトリ内の実体には、root 相対マーカー `layat.lib.mkRootRelativeSymlink "rel/path"` を使う。src は配置 root からの相対パスで、engine が実行時に解決した root を基準に、target の親を起点とする相対リンクを張る。
- **Avoid**: デフォルト挙動として扱うこと、`src` の型による暗黙の分岐で生み出すこと。

## 状態管理

### generation
rollback の単位。既定の **generation backend** である nixprofile が layat 自身の Nix profile（`nix-env --profile <dir>` 式）に乗せて管理する。nixprofile はコミット（`--set`）と切替（`--switch-generation`。rollback もこれを使う）を `nix-env --profile <dir>` で行い、一覧は nix-env を経由せず profileDir の世代リンク（`profile-<N>-link`）を読み取り専用で走査する。間引き（`--delete-generations`）と store GC（`nix-collect-garbage`）はユーザーが nix コマンドで行う。
- **Avoid**: 「stateless スクリプト」前提で語ること（初期方針からは覆っている）、新 `nix profile` CLI で管理すること（profile-manifest を要求し、`nix-env --set` 製の profile では動かない）。

### store manifest
「layat が何を配置したか」を記録する世代由来のデータ。実体は link-farm derivation 内の `manifest.json`（`schemaVersion` を持つ）で、Nix（`lib.mkManifest`）が生成し Go エンジンが読む——Nix↔Go の契約。エンジンの保守的な stale 除去（記録通りを指す layat 管理 symlink だけを削除し、ユーザーの実ファイルには決して触れない）を支える。

### last-applied record
「この root に layat が最後に置いたもの」という FS の事実を layat 自身が持つ記録（last-applied 記録）。世代ではない。置き場は `<state>/layat/<鍵>/last-applied.json` で、鍵は profileDir と同じ規則（`--root` なしの home は `<name>` 直、それ以外は `<roothash>/<name>` + backref `.root`）。内容は `recordVersion`（2。読みは 1 と 2 を受け付ける）・`linkFarm`・`appliedAt`・`manifest`（配置した `manifest.json` をそのまま（内容を変えずに）埋め込む）・`origin`（直近に記録を書いた apply の出所 `{generator, entrypoint, identity}`。`identity` は manifest が宣言していればその写し）で、store の GC に依存しない。記録に `origin` があり今回の出所と食い違えば（両方に `identity` があれば `identity` だけで比べる）、apply は warning を出して続行する（→ [ADR-0070](adr/0070-record-origin-and-same-name-warning.md)）。apply（generation skip を含む）と rollback は FS 操作の成功後・世代操作の前に、reset は撤去の後に書き直し、reset で全 entry が消えれば削除する。stale 除去・drift 修復・reset の撤去対象・rollback の離れる世代の baseline は**記録 → profile の current → 空**の順に取り、記録と profile の `linkFarm` が食い違えば記録を採って warning を 1 行出す。初回（どちらも無い）は baseline を空とし、「last-applied 記録なし」の warning を 1 行出す。壊れた記録は「記録なし」に倒さず error にする（→ [ADR-0058](adr/0058-last-applied-record.md)）。
- **Avoid**: 記録を世代と呼ぶこと（rollback の target は世代リンクから読む）、記録を flock の鍵と説明すること（鍵は profileDir）。

### method
配置の種別を選ぶ entry のフィールド（旧名 `mode` からの改名）。unix の file mode との誤読を避けるために改名した。
- **Avoid**: 旧名 `mode`。

## Flagged ambiguities

- **「symlink」単独では曖昧**。必ず **store link** か **out-of-store symlink** のどちらかに寄せる。デフォルトの store link 自体も symlink として実現されるため、裸の語では種別を識別できない。
- **「unification（統一）」は「removal（廃止）」ではない**。store link の統一は out-of-store を削除しない。デフォルトから降格し、明示的な関数の裏に隔離するだけ。
- **`src` と `subpath` は別概念**。`src` = どの物（store パス / リポジトリ）、`subpath` = その中のどのパス。名前は似ているが直交する概念。旧名 `source`（= 現在の `subpath`）は使わない。
- **「standalone」は配置モードではなく起動形態**。standalone（CLI を直接叩くこと）は配置モード（home / project / system）と直交する。モードは `root` マーカーが決める。

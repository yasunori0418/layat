# layat

フェッチ済み git リポジトリを、ユーザー環境の任意パスへ symlink / copy で配置する Nix ライブラリ・モジュール群の用語集。設定生成は行わない。

ここは glossary であり仕様書ではない。実装詳細は `docs/requirements/` の `requirement` item（索引は `docs/spec.md`）、設計判断は `docs/adr/` に置く。2026-06-07 の方針転換で意味が変わった語を中心に、正名と避けるべき同義語を固定する。

## Language

### 配置の中心抽象

**配置プリミティブ (placement primitive)**:
layat のコア。「nix store のパスを root 相対の target に配置する」純粋関数。モジュール抽象で隠さず、ユーザーが合成して使う（→ ADR-10dcf5a9-f298-4123-9a50-24f43e25d30c）。
_Avoid_: 「配置フレームワーク」「設定管理」（設定は生成しない）

**engine (layat エンジン)**:
配置（**ネイティブ FS 操作**）と古くなった配置の除去を一手に所有する**配置コア**。`manifest.json` を入力に取り、外部コマンドは `git`（toplevel）だけを実行する Go **ライブラリ**として実装し、**layat CLI** が import して駆動する（→ ADR-3d450af1-1c79-48ce-8ee9-b989dcb41ad7, ADR-2119433e-de69-4447-85a5-575d58246007, ADR-553eea60-6ab4-4da7-93a9-3acf44b2fae3）。config ごとに bash を生成しない。「ライブラリ」は **`internal/` のバイナリ内層分離**であって公開 import 可能な再利用モジュールではない（安定面は `manifest.json` 契約に閉じる）。**stdlib-only 厳守**（`syscall.Flock` / `filepath.WalkDir` ベースのコピー / `encoding/json`）。CLI が**生成器**の Build に世代コミット可能な link-farm（= store path）を `<profileDir>/.pending` へ張らせ（nix 生成器では `nix build --out-link`・→ ADR-efa7685c-6fa4-45df-b7a3-8f3f58e4118c）、配置〜世代 commit の **GC 窓**を indirect gcroot で塞ぐ（`profileDir` は config 専用ディレクトリで profile リンクは `<profileDir>/profile`・世代 `profile-N-link`・backref `.root` は `<roothash>` 階層・→ ADR-51863e94-6a9d-4029-95e7-a7ad2bfa93c1, ADR-419914ba-9896-4dd1-853b-fb514fcb9f38）。実行順は **eval 先行 → flock → build**で、生成器の Roots で `rootKind` を build 前に先取りして（nix 生成器では `mkManifest` が passthru する値を安価 eval）profileDir を確定し、flock を取ってから build を**ロック内**で行う（profileDir 未確定の循環と out-link 競合を同時に解消・→ ADR-2f472fc4-23ed-4169-aa49-8bded9e336d9）。非 build コマンド（`reset` / `rollback` / `list-generations`）も profileDir 確定のため Roots を先行し、`apply --all` は rootKind を一括の Roots で取る（nix 生成器では 1 回の一括 eval・→ ADR-28605860-770e-41df-95dd-7dc05b20497a）。並行実行は**解決後 profileDir 単位**の flock で直列化する（明示 apply は blocking wait / shellHook は try-lock skip・同一 profileDir 衝突はユーザー責任で後勝ち・→ ADR-ac648001-7e22-4fe2-a457-596064e32717, ADR-45380022-d959-4b22-aa05-5993ac1f3c87, ADR-767c33e5-73eb-4ef0-aba7-c87673fa5c2e）。配置前に target の祖先 component を lstat walk し、symlink ならネスト不可で error 停止する。target に foreign symlink（自身の前世代 manifest に記録の無い・別 config / 別ツール / 手動）があれば warning を出して後勝ちで置換する（→ ADR-095220c9-5a3d-4fd4-94bd-8fa1a02dd36d）。
_Avoid_: 「config ごとに生成される bash スクリプト」「各層がネイティブ機構へ翻訳する」「層ごとの配置ロジック」「CLI と一体の平らな単一実装」「engine を公開 Go モジュールとして外部から import する」（エンジンは `manifest.json` in 契約を保つ `internal/` 層）

**layat CLI**:
ユーザーが実行する一次 UX。PATH に常駐する `packages.layat`。**entrypoint** を発見（CWD 既定 / `-f` で上書き）し、**generator（生成器）**（既定 `nix`）経由で named manifest を得て **engine** に配置させる（→ ADR-87a2cf34-9592-4c81-8980-246bca579ca0, ADR-36cb1e78-90c0-47d2-bbb1-4bfec5db2964）。`apply [<name>]`（省略時は `layat.default`）/ `apply --all` / `reset`（配置物を無い状態へ戻す。名指し必須・→ ADR-ddeafc2b-f5a9-4332-a0b9-b18bea6ce6f7, ADR-56ac6298-db67-4f63-8b60-f14af3b9570a）/ `rollback`（home / system mode 限定）/ `list-generations`（home / system mode 限定）/ `gitignore`（project mode 限定・→ ADR-777fc3c8-4ca8-49da-aaf3-146b7d991ebe）/ `prune`（root が実在しない孤児 profile 系列を削除する。名指し不要・entrypoint 発見も nix eval / build もしない・→ ADR-b4be88b8-1c19-426d-885d-f5d80b95252f, ADR-e630482a-bc11-4ba6-a9cc-1ef46f54104d, ADR-11507e12-5b0f-4f8f-bae9-fb7fd41a9203, ADR-ad1c2507-1f42-4cd3-ab56-9183892f8988, ADR-7a3dc78d-8946-46ae-94ec-b752f2ca7c2e）/ `status`（last-applied 記録を走査して現在 apply されている内容を読み取り専用で表示する。entrypoint 発見も nix eval / build もしない・→ ADR-7abe9cc7-30cd-49c9-b10e-bb5e1790ff0d）/ `init` のサブコマンドを持つ。**`apply` の manifest 取得元は 2 つ**: 選択された生成器で entrypoint を build する（標準）か、ビルド済み link-farm を `--manifest` で直接渡す（`apply --manifest`・**prebuilt** 生成器・取得後の挙動は同一・→ ADR-24488788-a5e6-47a5-9074-97192880642a, ADR-1ddf424a-db2c-4bcb-8823-bf7e4326b980）。`--all` と合成するときは `--manifest <name>=<link-farm>` を繰り返し、ビルド済みの名前つき集合を `apply --all` の対象にする（module / host activation の engine kick・→ ADR-24488788-a5e6-47a5-9074-97192880642a）。既定の nix 生成器が実行する nix コマンドは `--help` で開示し、`--debug` では生成器が実行する内部コマンドを逐次開示する（→ ADR-5caf1c14-a054-4c8d-95a2-c7f89e0c5246, ADR-5be36ff2-ed08-4e8e-8476-62ff2b08a9dd）。
_Avoid_: config ごとの `nix run .#x` ラッパーを一次 UX と説明すること（per-config ラッパー `mkActivationScript` は廃止 → ADR-86a77232-dc99-4c04-bd6d-2127410e8828）、`apply` を「常に entrypoint を build する」と説明すること（`--manifest` でビルド済み link-farm も適用する・→ ADR-24488788-a5e6-47a5-9074-97192880642a）

**generator (生成器)**:
**layat CLI** が manifest を得る相手。契約は Discover（entrypoint の発見）/ Roots（build せずに rootKind・root を返す）/ Build（世代コミット可能な link-farm = store path を pending に張る）/ DryBuild（gcroot を張らずに store path を返す）の 4 操作で、`internal/generator` に置く（→ ADR-bac005fb-fe4d-4a3b-b9e1-517a62fe46c2）。nix 実装（nixgen・`internal/generator/nixgen`）が既定で、選択は明示指定のみ（`--generator` > `LAYAT_GENERATOR` > `layat.toml`（`-f` のディレクトリか cwd）> `$XDG_CONFIG_HOME/layat/config.toml` > 既定 `nix`・→ ADR-36cb1e78-90c0-47d2-bbb1-4bfec5db2964）。診断は CLI が渡す writer へ素通しし、失敗は `generator.Error` で返す。**engine** は誰が manifest を生成したかに関知しない（→ ADR-b06dd8ab-317e-41ee-9c80-dd70dcd7f2fd）。
_Avoid_: 「CLI が内部で `nix build` / `nix eval` を直接回す」と説明すること（nix を実行するのは nix 生成器）、置かれているファイルから生成器が推測されると説明すること

**prebuilt**:
`apply --manifest <link-farm>` の実体である生成器。Discover = 与えられたパス / Roots = その link-farm の `manifest.json` から読む / Build・DryBuild = そのパスを返す（→ ADR-1ddf424a-db2c-4bcb-8823-bf7e4326b980）。外面は ADR-24488788-a5e6-47a5-9074-97192880642a のまま。`--generator` の選択肢には出ない（→ ADR-9acd3e01-58f1-48ca-b84f-f33a5bc1aaa5）。
_Avoid_: `--generator prebuilt` のように選択肢として挙げること、`--manifest` を生成器とは別の取得経路と説明すること

**generation backend (世代バックエンド)**:
**engine** が世代の記録・一覧・切替を委ねる相手。契約は Commit（link-farm を新世代として記録し profile を向ける）/ List（世代の一覧と current）/ Switch（既存の世代へ profile を向ける）/ UserBase（user state の profileDir の基底）/ SystemBase（system mode の profileDir の基底）の 5 操作で、入出力はパスと整数のみ。`internal/generation` に置く（→ ADR-62540f8f-a8ba-4441-97b0-36b50d7862d8）。nix 実装（nixprofile・`internal/generation/nixprofile`）が既定かつ唯一の実装で、`nix-env` を実行するのはここだけ（基底は `<state>/nix/profiles/layat` / `/nix/var/nix/profiles/layat`）。CLI は常に nixprofile を注入し、engine はバックエンドが渡されなければ error にする。**engine** は誰が世代を管理するかに関知しない。profileDir 内部のレイアウト（`profile` / `profile-<N>-link` / `.pending`）は契約の固定部分で、engine が直接読む。バックエンドを選ぶフラグ / 環境変数 / `layat.toml` キーも、基底を上書きする設定キーも持たない。
_Avoid_: 「engine が `nix-env` を叩く」と説明すること、バックエンドの選択手段や基底の設定キーがあると説明すること

**entrypoint**:
layat CLI が読む nix の config ファイル。`flake.nix` / `shell.nix` / `default.nix` のいずれか。**`layat.<name>`** に named manifest（`mkManifest` の結果）を公開する。config は依然 Nix で書き `nix build` で評価される（→ ADR-e683ac05-c0e0-41c1-8b6a-38b605972fe1）。
_Avoid_: 「layat が CWD から config 内容そのものを発見する」と説明すること（発見するのは entrypoint *ファイル*。config は Nix 評価で確定する）

**module (モジュール / 配線)**:
CLI 単体・home-manager・NixOS / nix-darwin・devShell `shellHook` といった統合層。エンジンを起動する**配線**に徹し、自身では配置しない。`home.file` / `systemd.tmpfiles` へは翻訳しない（→ ADR-3d450af1-1c79-48ce-8ee9-b989dcb41ad7, ADR-90964d0f-3c2e-4201-a04b-5821f2f4c799）。**kick 方法は 2 クラス**（→ ADR-24488788-a5e6-47a5-9074-97192880642a）: **entrypoint 駆動**（CLI 単体 / devShell は `layat.<name>` を公開し `layat apply <name>` で build→配置）と、**ビルド済み manifest**（home-manager・NixOS / nix-darwin はモジュール評価時に `mkManifest` でビルドした link-farm を `layat apply --all --manifest <name>=<link-farm> … --jobs <n>` の 1 回の起動で配置）。配置〜世代コミットは両クラスで同じエンジンの処理。HM モジュールは **`layat.configs.<name>`** が正規の形で、config ごとに独立した link-farm と profile `<name>` を持つ（CLI 単体と同じく役割を分けられる）。configs 間の target 衝突は module eval ではなく activation の `apply --all` 前段検出で止める。`layat.entries` は `configs.default.entries` への deprecated な rename 糖衣。`layat.recopy` / `layat.jobs` / `layat.backup` は全 config に一律に掛かる共通オプションで、activation の `--recopy` / `--jobs` / `--backup` へ配線する（→ ADR-01abcd5a-464f-4c44-8afd-30950631487c, ADR-3447db4a-cdec-4573-bbad-05bbffc34a65, ADR-5af374dd-61a7-4a13-b2e6-fb5136673dd3, ADR-f58e9981-bacc-4937-98b8-bddce0a057db, ADR-1338a293-3f69-48f5-9164-efb4ec6994d6, ADR-7b9e4a7b-56d3-4356-9f67-bedbde8f0a52）。
_Avoid_: 「モジュールがファイルを配置する」「モジュールがネイティブ機構へ変換する」

### 配置の入出力

**entry / entries**:
配置定義。`entries` は **属性キー = target の attrset**で、各値が entry（`{ src; subpath?; target?; method?; }`）。**属性キーが識別子**で、`target` はキーから既定値を取る（home-manager `home.file` 同型・→ ADR-c494e86e-ceb9-4380-b983-4e09553b01d6）。識別子の一意性は Nix の attrset キー重複不可で native に担保され、手動 `name` フィールドは持たない。
_Avoid_: 「ファイルエントリ」（配置元はディレクトリのこともある）、entry に `name` フィールドを持たせること（識別子は属性キー = target・→ ADR-c494e86e-ceb9-4380-b983-4e09553b01d6）、entries を `{ name; … }` のリストと呼ぶこと（旧形）

**src**:
**entry** の配置元（どの store パス / リポジトリか）。デフォルトは Nix ストアへの **store link**。out-of-store は明示マーカーで opt-in する（→ ADR-73c24620-d2c0-4249-97fa-1b6296f1e79c）。**subpath**（その中のどのパスか）と直交する別概念で、`src` を `source` の短縮形と読まない（→ ADR-3417c5b0-8c5f-4c2c-827a-6f56bce301ab）。
_Avoid_: **subpath** と混同すること、`source` と呼ぶこと

**subpath**:
**entry** で **src** 内のどのパスを取り出すかを表す相対パス。file / dir 両対応。デフォルト `"."`。リポジトリ全体は **subpath を省略**するのが canonical（`subpath = "."` は明示形）。評価時に確定する subpath 選択であり、糖衣 marker（`wholeRepo` 等）は足さない（marker は実行時解決の種別を運ぶ入れ物・→ ADR-400cadd4-0151-470b-92ae-9b11aa8e9bed, ADR-2f5b8825-f13b-4f73-bdf1-fe14e0098d57）。
_Avoid_: 旧名 `source` / `dir` を使うこと、**src** と混同すること、「`"."` 以外で全体を表す専用トークン/marker が要る」と考えること（省略で表せる）

**target**:
**entry** の配置先パス。**root** からの相対パスで指定する。**entries の属性キーが target の既定値**であり、entry の identity（古くなった配置を除去するときの diff キー・一意性）でもある（→ ADR-c494e86e-ceb9-4380-b983-4e09553b01d6）。

**root**:
配置先の基準パス。公開 API の `root` 引数で**明示必須**に選ぶ（**暗黙デフォルトは持たない**）。型は `string（絶対パス・評価時固定）| marker（実行時解決）` の union。マーカーは **projectRoot** / **homeRoot** / **systemRoot** の 3 つ（→ ADR-400cadd4-0151-470b-92ae-9b11aa8e9bed）。各マーカーは適用形 `homeRoot { subdir = "..."; }` も受け付ける。適用形でも root は動かず、全 target へ `<subdir>/` を前置するのと等価な評価時の糖衣である（→ ADR-6669c002-1b90-47a2-8a84-a0f7bddfd965）。
_Avoid_: 「`$HOME` 固定」「既定は home mode」「root 省略時の既定」と説明すること（暗黙デフォルトは廃止し、root は明示必須になった）。マーカーを「パス文字列を返す糖衣」と説明すること（実行時解決の種別であって評価時にパスへ展開しない）。適用形で「root が実際に subdir へ解決される」（`$HOME/.config` が root になる）と説明すること（root と `rootKind` は単体形と同じで、変わるのは target だけ）

### 配置モード

**home mode**:
**root** = `$HOME` の配置モード。`homeRoot` マーカーで明示選択する（旧: `root` 省略時の既定。暗黙デフォルトは廃止）。CLI 単体も home-manager 等の module もこのモード。世代を毎回積み、CLI 単体では `layat rollback` をユーザーに公開する。module では公開せず、ユーザー向け rollback は host（`home-manager --rollback` 等）に一本化し、layat profile は旧内容の新世代を積む形で前進だけを続ける（→ ADR-15a358f4-1c81-49d5-b658-3a1fe39f0f75）。
_Avoid_: 「standalone 専用」「`root` 省略時の既定」と説明すること

**project mode**:
**root** = プロジェクトルートの配置モード。`projectRoot` マーカーで選ぶ。配置物は **ephemeral placement**（コミット対象外）で、世代・rollback はユーザーに公開しない（→ ADR-15a358f4-1c81-49d5-b658-3a1fe39f0f75）。
_Avoid_: 「CWD 相対」「設定ファイル相対」と説明すること（root は git toplevel で解決）

**projectRoot**:
**project mode** を選ぶ root マーカー。実行時に git toplevel を **root** に解決する（`--root` で上書き可）。`homeRoot` / `systemRoot` と並ぶ root マーカーの一つ。`mkOutOfStoreSymlink` と同じ「マーカーを渡して挙動を opt-in する」パターン（→ ADR-400cadd4-0151-470b-92ae-9b11aa8e9bed）。適用形 `projectRoot { subdir = "..."; }` も受け付ける（意味は **root** の項）。
_Avoid_: 設定ファイルの場所を指すと解釈すること（Nix では store path 化して成立しない）

**homeRoot**:
**home mode** を選ぶ root マーカー。実行時に `$HOME` を **root** に解決する。旧来「`root` 省略時の暗黙デフォルト」だった `$HOME` を明示マーカーへ昇格したもの（→ ADR-f85f1afd-9b87-4f7a-922e-97c62554f198）。適用形 `homeRoot { subdir = ".config"; }` は root を `$HOME` のまま、全 target を `.config/` 配下へ置く（→ ADR-6669c002-1b90-47a2-8a84-a0f7bddfd965）。
_Avoid_: 適用形の root が `$HOME/.config` へ解決されると説明すること

**systemRoot**:
**system mode**（root = `/`）を選ぶ root マーカー。distro 構想（root = `/`）の配置に使う。「将来の絶対パス文字列 seam」とされていたものを正式なマーカーへ昇格（→ ADR-400cadd4-0151-470b-92ae-9b11aa8e9bed）。適用形 `systemRoot { subdir = "..."; }` も受け付ける（意味は **root** の項）。

**ephemeral placement**:
**project mode** の配置物の性質。クローンごとに再生成される前提で、プロジェクトにコミットされない。ゆえに activation は git 状態に干渉しない（→ ADR-4dca6ef4-2791-48d5-aa82-7037630968a6）。
_Avoid_: 「vendoring」「成果物をコミットする配置」と混同すること

### 配置の種別

**store link**:
コア・デフォルトの配置。配置先が Nix ストアパスである symlink。再現性を担保する既定の使い方（→ ADR-73c24620-d2c0-4249-97fa-1b6296f1e79c）。「統一」とは「デフォルト/コアをストアにし、out-of-store を明示的な例外に降格する」こと。
_Avoid_: out-of-store symlink と混同すること、「コピー」と呼ぶこと

**out-of-store symlink**:
ローカルパスへのライブ symlink。明示マーカーでのみ opt-in する例外の手段で、第一級機能ではない（→ ADR-73c24620-d2c0-4249-97fa-1b6296f1e79c）。マーカーは 2 つある。`layat.lib.mkOutOfStoreSymlink "/abs/path"` はローカル絶対パスを指す（開発中 dotfiles のライブ編集用）。
root 相対マーカー `layat.lib.mkRootRelativeSymlink "rel/path"` はリポジトリ内の実体を指す。src は配置 root からの相対パスで、engine が実行時に解決した root を基準に、target の親を起点とする相対リンクを張る（→ ADR-b7f9b972-69b7-4bbe-affc-97978e1a55f7）。
_Avoid_: デフォルト挙動として扱うこと、`src` の型による暗黙分岐で生むこと

### 状態管理

**generation (世代)**:
ロールバック単位。既定の **generation backend**（nixprofile）が layat 自身の nix profile（`nix-env --profile <dir>` 式）に乗せて管理する（→ ADR-f90509f2-d75a-4162-baa5-a5cfdb8a7c30, ADR-9755ba76-4c02-4b4e-a8c0-031a24a3a5b9）。nixprofile はコミット（`--set`）と切替（`--switch-generation`。rollback もこれを使う）を `nix-env --profile <dir>` で行い、一覧は nix-env を経由せず profileDir の世代リンク（`profile-<N>-link`）を読み取り専用で走査する。間引き（`--delete-generations`）と store GC（`nix-collect-garbage`）はユーザーが nix コマンドで行う（→ ADR-9755ba76-4c02-4b4e-a8c0-031a24a3a5b9）。`rollback` は profile dir ≠ 配置先のため再配置が必須で、古くなった配置の除去は基準世代=離れる世代・ポインタ移動は最後（→ ADR-d97154cd-a2e5-4853-a682-800ff9435c89）。
_Avoid_: 「stateless スクリプト」前提の語り（初期方針からは覆っている）、`nix profile`（新CLI: `list` / `wipe-history` / `rollback`）で管理すること（profile-manifest を要求し `nix-env --set` 製 profile では動かない・→ ADR-9755ba76-4c02-4b4e-a8c0-031a24a3a5b9）

**store マニフェスト (store manifest)**:
「layat が配置した」記録を持つ世代由来のデータ。実体は link-farm derivation 内の **`manifest.json`**（`schemaVersion` 付き）で、Nix（`lib.mkManifest`）が生成し Go エンジンが読む **Nix↔Go の契約**。GC 参照は併存する symlink farm が明示的に張る。エンジンが古くなった配置を保守的に除去するときの不変条件（記録通りを指す layat 管理 symlink だけ削除し、ユーザーの実ファイルには触れない）を支える（→ ADR-9f2cc068-8edb-4e3d-958b-0dff8f6ed90c, ADR-b06dd8ab-317e-41ee-9c80-dd70dcd7f2fd）。

**last-applied 記録 (last-applied record)**:
「この root に layat が最後に置いたもの」という FS の事実を layat 自身が持つ記録。置き場は `<state>/layat/<鍵>/last-applied.json` で、鍵は profileDir と同じ規則（`--root` なしの home は `<name>` 直、それ以外は `<roothash>/<name>` + backref `.root`）。内容は `recordVersion`（2。読みは 1 と 2 を受け付ける）・`linkFarm`・`appliedAt`・`manifest`（配置した `manifest.json` をそのまま（内容を変えずに）埋め込む）・`origin`（直近に記録を書いた apply の出所 `{generator, entrypoint, identity}`。`identity` は manifest が宣言していればその写し）で、manifest を内容ごと持つので store の GC と無関係（→ ADR-52423842-bacf-4773-bbca-bcb34ae71013, ADR-0db061d1-fa30-436d-91d7-c0cec8804c52）。記録に `origin` があり今回の出所と食い違えば（両方に `identity` があれば `identity` だけで比べる）、apply は warning を出して続行する。apply（generation skip を含む）と rollback は FS 操作の成功後・世代操作の前に、reset は撤去の後に書き直す（reset で全 entry が消えれば削除）。古くなった配置の除去・drift 修復・reset の撤去対象・rollback の離れる世代の比較の基準は**記録 → profile の current → 空**の順に取り、記録と profile の `linkFarm` が食い違えば記録を採って warning を 1 行出す。初回（記録も profile も無い）は比較の基準を空とし、「last-applied 記録なし」の warning を 1 行出して配置する。破損した記録は「記録なし」に倒さず error にする。
_Avoid_: 記録を世代と呼ぶこと（rollback の target は世代リンクから読む）、記録を flock の鍵と説明すること（鍵は profileDir）

### エラー処理

**エラー wrap 規約 (error-wrap convention)**:
`internal/engine/` のエラーは**発生源（syscall / exec / parse に最も近い箇所）で 1 回だけ** `fmt.Errorf("layat: cannot <op> (<target>): %w", ...)` の形で op（何をしようとしたか）と対象パスを付けて wrap する。呼び出し元は素通し（bare `return err`）が既定で、伝搬の途中での再 wrap は `layat: ...: layat: ...` の二重 context ノイズと既存のエラー文字列 assert 破壊を招くため行わない。sentinel エラー（`ErrSkipped` = `lock.ErrLocked` 等）は常に `%w` で透過させ、呼び出し元が `errors.Is` で判定できる状態を保つ。監査の判定基準は「ユーザーに見える失敗が op＋対象を少なくとも 1 回含むか」であり、含まれていれば伝搬の途中に op＋対象が無くても規約違反ではない。
_Avoid_: 境界を跨ぐたびに context を積み増すこと、sentinel を wrap で握り潰し `errors.Is` を辿れなくすること

## Flagged ambiguities

- **「symlink」単独では曖昧**。必ず **store link** か **out-of-store symlink** のどちらかに寄せる。`entries` の `method = "symlink"`（旧名 `mode`・→ ADR-3e283fe2-0177-4caf-8b60-577c5ee3b10a）+ out-of-store marker の `src` で out-of-store を指すが、デフォルトの store link も symlink で実現される。
- **「統一」は「廃止」ではない**。store link 統一は out-of-store を消すことではなく、デフォルトから降格して明示関数に隔離すること。
- **project mode の配置物は untracked が前提**。**ephemeral placement** を「コミットする vendoring」と取り違えない。`.gitignore` への列挙は専用コマンドの stdout 出力から得る（activation は `.gitignore` に触れない）。
- **`src` と `subpath` は別概念**。`src` = どの物（store パス / リポジトリ）、`subpath` = その中のどのパス。名前が似ているが直交する。旧名 `source`（= 現 `subpath`）は使わない（→ ADR-3417c5b0-8c5f-4c2c-827a-6f56bce301ab）。
- **「standalone」は配置モードではなく起動形態**。standalone（= CLI を直接実行する起動形態）と配置モード（home / project / system）は直交する。standalone から project mode も home mode も使える。「standalone = home mode」と短絡しない（配置モードは `root` マーカーが決める）。
- **project mode の `layat` は devShell 同梱が canonical**。`templates/project` の devShell `packages` に pin 版 `layat` を入れ、CLI と `layat.lib`（manifest `schemaVersion`）を同一 flake 入力で一致させる。グローバル install は CLI 単体（home mode）の利便（→ ADR-88231ffc-4836-4252-be34-894944fd28c3）。
- **entry の配置種別フィールドは `method`**（旧名 `mode`）。unix file mode との誤読を避ける改名（→ ADR-3e283fe2-0177-4caf-8b60-577c5ee3b10a）。

## 会話例

> **Dev**: この dotfiles、編集しながら即反映したいんだけど src にローカルパス渡せばいい？
>
> **Maintainer**: 文字列で渡す暗黙分岐は廃止した。それは **out-of-store symlink** で、`mkOutOfStoreSymlink "/abs/path"` を `src` に渡して明示的に opt-in する。デフォルトは **store link**——ストアへの symlink で再現性がある方。
>
> **Dev**: 配置自体は home-manager のときは `home.file` に変換されるの？
>
> **Maintainer**: しない。全層で layat **エンジン**が**ネイティブ FS 操作**で配置する。home-manager **モジュール**はエンジンを起動する**配線**でしかない。だから振る舞いは層を跨いで同じ。
>
> **Dev**: 前の **世代** に戻したら、消える symlink がユーザーの実ファイルを巻き込んだりは？
>
> **Maintainer**: しない。**store マニフェスト**が「layat が置いた」と記録した symlink だけをエンジンが消す。**target** に元からある実ファイルには触れない。

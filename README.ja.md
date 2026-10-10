# layat

> **layat** — lays contents at root-relative targets, as the manifest says.
> （manifest の言うとおりに、内容を root 相対の target へ置く。）

*この文書は英語版 [`README.md`](README.md) の日本語訳。仕様・用語の一次参照は英語版とし、両者に差異があれば英語版が優先する。*

layat は、**フェッチ済みの Nix store パスの内容を `root` 相対の target へ配置する** Nix ライブラリ・モジュール群(symlink もしくは copy)。設定の生成は **行わない**。`root` は `projectRoot` / `homeRoot` / `systemRoot` マーカーで**明示的に**選ぶ(**暗黙のデフォルトは持たない**)。

> **ステータス: MVP / 実装フェーズ。** 実装済みの範囲は standalone CLI ＋ **project mode** をコアとし、**home mode** もサポート。NixOS / nix-darwin モジュールおよび **system mode** は将来対応。全体のマトリクスは [MVP ステータス](#mvp-ステータス) を参照。API は今後も変わりうる。

---

## なぜ layat か

layat は **フェッチ**(`src` は store パス)と **配置**(固定の実行時エンジン)を分離し、配置の挙動を **単一のコア** に閉じてユーザーが明示的に駆動する。

- **設定生成をしない。** リポジトリにすでにある内容を配置する。
- **独立した単位。** 各 config(`layat.<name>`)はそれ自体が独立した Nix profile。ある更新が別へ波及しない。
- **home-manager 非依存。** `lib/` は nixpkgs のみに依存する。module 統合はエンジンを *起動するだけ*。
- **readlink パターンマッチではなく自己記録の manifest。** stale 除去は前世代の manifest を使う。詳細は [`docs/concept.md`](docs/concept.md#home-manager-homefile-との配置意味論の差) を参照。

---

## 仕組み

```
[layat CLI]  packages.layat — PATH 上に乗る一次 UX
  · entrypoint を発見する(flake.nix / shell.nix / default.nix)
  · 生成器(既定 `nix`)を通じて named manifest を取得する
  · エンジンを駆動して配置・stale link 除去・profile 切替を行う
   ↓ manifest.json
[engine]    Go ライブラリ(stdlib-only)
  · manifest.json を入力に取り、`nix`(profile)と `git`(toplevel)のみを叩く
  · place / replace / remove のネイティブなファイルシステム操作 ＋ 保守的な stale 除去
```

- **engine** が配置と stale 除去を所有する。`manifest.json`——安定した Nix↔Go の契約——を読み、ネイティブなファイルシステム操作を実行する。
- `lib.mkManifest` は link-farm derivation(`manifest.json` ＋ symlink farm)を生成する **純粋関数**。副作用を持たない。
- **entrypoint** は CLI が読む Nix ファイル(`flake.nix` / `shell.nix` / `default.nix`)で、`layat.<name>` に named manifest を公開する。
- **生成器**は CLI が manifest を得る相手。既定は `nix`(「[生成器の選択](#生成器の選択)」を参照)。`apply --manifest` はビルド済み link-farm をそのまま使う *prebuilt* 生成器を通る。

---

## 要件

- **Nix**。環境で experimental features を有効化していること：`experimental-features = nix-command`(flake entrypoint には `flakes` も)。layat は `--extra-experimental-features` を黙って注入 **しない**。機能が無効なら、前提条件と有効化方法を明示するメッセージを出して停止する。
- **git** が `PATH` 上にあること(project mode でプロジェクトルートを解決するために使う)。

---

## インストール

### Standalone(home mode)

CLI をグローバルにインストールし、`layat` を `PATH` に乗せる。

```bash
nix profile install github:yasunori0418/layat
```

`schemaVersion` のずれを避けるため、CLI と flake が固定する `layat` を **同一 input** に揃える(グローバル CLI と flake 側の `layat.lib` は別 input で drift しうる。エンジンは自分より新しい `schemaVersion` を拒否する)。

### Project mode(canonical: devShell に固定する)

project mode の canonical な形は、プロジェクトの devShell に **固定した `layat` を同梱** すること。これで CLI と `layat.lib` が同一 flake input(`flake.lock` で固定)から来る。

```nix
devShells.${system}.default = pkgs.mkShell {
  packages  = [ layat.packages.${system}.layat ];   # 固定した layat を PATH に
  shellHook = "layat apply <name> --no-wait";        # `nix develop` / direnv 進入時に配置
};
```

### 新規プロジェクトの scaffold

`layat init` は `nix flake init -t` の透過的なラッパー。layat 自身は何も生成せず、既存ファイルを上書きしない。

```bash
layat init standalone   # homeRoot の例
layat init project      # projectRoot の例 ＋ devShell 配線 ＋ .gitignore ガイド
```

---

## Quickstart

### Project mode(中心的な使い方)

**project mode** では root はプロジェクトルート(git toplevel)。配置物は **ephemeral**——クローンごとに再生成され、コミットされない——なので activation が git 状態に干渉しない。これが layat の中心的な使い方：リポジトリに組み込み、devShell から起動して、store パスを任意の in-repo パスへ配置する。

```nix
# flake.nix — リポジトリ進入時に、フェッチ済み store パスから .claude/skills/nix を配置する
{
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  inputs.layat.url    = "github:yasunori0418/layat";

  inputs.claude-skills.url   = "github:someone/claude-skills";
  inputs.claude-skills.flake = false;

  outputs = { self, nixpkgs, layat, ... }@inputs:
    let
      system = "x86_64-linux";
      pkgs   = nixpkgs.legacyPackages.${system};
    in
    {
      layat.${system}.skills = layat.lib.mkManifest {
        inherit pkgs;
        root = layat.lib.projectRoot;        # 実行時に git toplevel へ解決される
        entries = {
          ".claude/skills/nix" = { src = inputs.claude-skills; subpath = "skills/nix"; };
        };
      };

      devShells.${system}.default = pkgs.mkShell {
        packages  = [ layat.packages.${system}.layat ];   # 固定した layat(canonical)
        shellHook = "layat apply skills --no-wait";       # shell 進入時に配置
      };
    };
}
```

```bash
# shell に入る — shellHook 経由で配置が自動実行される
nix develop          # または: direnv allow

# プロジェクト所有者が .gitignore に加えるべき target を一覧する(stdout のみ・書き込みなし)
layat gitignore skills >> .gitignore
```

- ここでは世代は内部機構であり、`rollback` / `list-generations` は project mode では **公開されない**(ephemeral な配置に rollback は無意味)。
- devShell では **named apply**(`layat apply skills`)か `layat apply --all --project-root` を使う。裸の `--all` は home-mode config も `$HOME` へ配置してしまうため、混在 entrypoint では footgun。

### Home mode(standalone、役割を別 profile に分ける)

**home mode** では root は `$HOME`(`homeRoot` で選ぶ)。各 config はそれ自体が独立した profile で、apply ごとにコミットされ、ユーザー公開の `rollback` を持つ。

```nix
# flake.nix — 各役割が named manifest = 独立した profile
outputs.layat.${system} = {
  vim-plugins = layat.lib.mkManifest {
    inherit pkgs;
    root = layat.lib.homeRoot;
    entries = {
      ".local/share/nvim/site/pack/foo/start/foo" = { src = inputs.vim-foo; };
      ".local/share/nvim/site/pack/bar/start/bar" = { src = inputs.vim-bar; };
    };
  };

  zsh-plugins = layat.lib.mkManifest {
    inherit pkgs;
    root = layat.lib.homeRoot;
    entries = {
      ".zsh/plugins/autosuggestions"     = { src = inputs.zsh-autosuggestions; };
      ".zsh/plugins/syntax-highlighting" = { src = inputs.zsh-syntax-highlighting; };
    };
  };
};
```

```bash
# 各役割を独立して更新 / 適用 / rollback できる(別 profile)
layat apply vim-plugins
layat rollback vim-plugins          # home mode のみ
layat apply zsh-plugins
layat list-generations vim-plugins
```

`src` を更新した後(flake input 更新、`npins update` など)は、影響を受けた config だけを再適用すれば、変更が他のツールに触れずに済む。

### home-manager モジュール

モジュールは `root = homeRoot` に固定される(`root` を再指定しない)。`home.activation` から `layat apply --manifest <link-farm>` でエンジンを起動し、`home.file` へは委譲しない。layat は自前の profile を **内部機構** として保持するが、ユーザー公開の rollback はホスト側に一本化される(`home-manager --rollback`)。

```nix
imports = [ inputs.layat.homeManagerModules.default ];

layat = {
  enable = true;
  entries = {
    # 外部リポジトリ(store link)
    ".claude/skills/nix" = { src = inputs.skills-repo; subpath = "skills/nix"; };
    # テーマを copy で(一度配置したら以降ユーザー管理)
    ".local/share/themes/dark" = { src = inputs.themes; subpath = "dark"; method = "copy"; };
    # ローカル dotfiles を out-of-store symlink でライブ編集
    ".config/nvim" = { src = layat.lib.mkOutOfStoreSymlink "/home/me/dotfiles"; subpath = "home/.config/nvim"; };
  };
};
```

> home-manager モジュールは MVP では **単一 manifest = 一つの profile**(固定名 `default`)であり、`<name>` 次元を持たないため、**役割分離はモジュール経由では使えない**。複数の独立 profile が要るなら standalone CLI 経路(`layat.<name>` entrypoint)を使う。

### 既存の flake に layat を追加する

`layat init` は新規プロジェクト向け。既存の `flake.nix` にあとから組み込むには、次の 4 ステップを手で行う(layat は flake を自動マージしない——「設定を生成しない」)。

1. **input を追加**：`inputs.layat.url = "github:yasunori0418/layat";`
2. **manifest を公開**：
   `outputs.layat.<system>.<name> = layat.lib.mkManifest { root = layat.lib.projectRoot; entries = { ... }; };`
3. **固定した layat を devShell に同梱**：`packages = [ layat.packages.${system}.layat ];`
4. **named apply を配線**：`shellHook = "layat apply <name> --no-wait";`

リポジトリが **flake-parts** を使うなら、ステップ 2 を flake module 経由で書く(`pkgs` を `perSystem` と整合させられる)。

```nix
imports = [ inputs.layat.flakeModules.default ];
perSystem = { pkgs, ... }: {
  layat.<name> = inputs.layat.lib.mkManifest {
    inherit pkgs;
    root = inputs.layat.lib.projectRoot;
    entries = { ... };
  };
};
# flake-parts はこれを flake.layat.<system>.<name> へ転置する — CLI のアドレッシングは変わらない。
```

> `nix flake check` は `warning: unknown flake output 'layat'` を報告するが exit 0(無害・想定どおり——`layat` 名前空間は manifest を `packages` の外に保つ)。成果物は `nix build .#layat.<system>.<name>` で検証する。

---

## `entries` スキーマ

`entries` は **`target` をキーとする attrset**——属性キーが識別子であり、既定の `target` を与える。一意性は Nix の attrset キーで native に担保され、手動の `name` フィールドは無い。

```nix
entries = {
  "<target>" = {
    src     = ...;          # 必須
    subpath = ".";          # 任意・デフォルト "."(リポジトリ全体)
    target  = "<target>";   # 任意・デフォルト = 属性キー
    method  = "symlink";    # 任意・デフォルト "symlink" | "copy"
  };
  # ...
};
```

entry submodule は **strict**(未知のキーを拒否する)。typo や旧名(`name` / `source` / `dir` / `mode`)は評価時エラー。

### `src` — 必須

配置元。デフォルトは **store link**(Nix store への symlink。再現性を担保する)。out-of-store は明示マーカーで opt-in する。

| `src` の値 | symlink が指す先 | 用途 |
|---|---|---|
| `path`(例：`inputs.myrepo`) | Nix store(不変) | バージョン固定の外部リポジトリ |
| `builtins.path { path = /home/...; name = "..."; }` | Nix store(ローカルを取り込み) | store 経由のローカル木 |
| `set`(例：`pkgs.fetchFromGitHub { ... }`) | Nix store(不変) | バージョン固定の外部リポジトリ |
| `marker`(`layat.lib.mkOutOfStoreSymlink "/abs/path"`) | ローカル FS(ライブ) | 開発中のローカル dotfiles |

```nix
src = inputs.myrepo;                                    # store link
src = pkgs.fetchFromGitHub { owner = "..."; repo = "..."; rev = "..."; hash = "..."; };
src = builtins.path { path = /path/to/dotfiles; name = "dotfiles"; };
src = layat.lib.mkOutOfStoreSymlink "/path/to/dotfiles"; # out-of-store(ライブ)・明示

# 廃止: 裸の文字列で暗黙の out-of-store にするのはサポートしない。
# src = "/path/to/dotfiles";   # エラー
```

### `subpath` — デフォルト `"."`

**`src` 内** のどのパスを取り出すかを選ぶ相対パス。file / dir 両対応。省略(または `"."`)でリポジトリ全体を選ぶ。

```nix
subpath = ".";                  # リポジトリ全体(明示形)
subpath = "skills/nix";         # サブディレクトリ
subpath = "themes/dark.json";   # 単一ファイル
```

`src` と `subpath` は直交する：`src` = *どの物*(store パス / リポジトリ)、`subpath` = *その中のどのパス*。

### `target` — デフォルト = 属性キー

配置先。**`root` からの相対パス**。属性キーが既定の `target` であり、entry の identity(stale 除去の diff キーであり、一意性のキー)でもある。キーを論理ラベルにして `target` を明示上書きすることもできる(`home.file` と同様)。

### `method` — デフォルト `"symlink"`

配置の種別を選ぶ。

| `method` | `src` の種類 | 挙動 | 世代 |
|---|---|---|---|
| `"symlink"` | path / set | Nix store への symlink(read-only) | あり(profile) |
| `"symlink"` | marker | ローカルパスへの out-of-store symlink(ライブ) | あり(link 先のみ) |
| `"copy"` | path / set | **一度だけ配置する** copy(書き込み可・ユーザー管理) | **なし** |
| `"copy"` | marker | 評価時エラー(矛盾) | — |

**copy は place-once でユーザー管理。** 一度実体化したら、layat は target に触れない。store の read-only モード(`0444` / `0555`)は保持しつつ owner-write を加えるので、copy は編集できる。上流 `src` の更新に追従するには `layat apply --recopy`(全 copy target を無条件に上書き)か `layat reset` 後に再適用する。copy は世代管理されず、rollback もされない。copy target に外部の実ファイルがすでにある場合、layat はスキップしつつ **警告** を出す(あなたのファイルを上書きしない)。

### entry を動的に生成する

名前を target キーに展開する。target は **あなたが** 制御する変数から導出する——`baseNameOf src` からではない(store パスは `/nix/store/<hash>-source` に解決されるので、`baseNameOf` は `<hash>-source` を返す)。

```nix
let plugins = [ "telescope" "treesitter" "cmp" ]; in
layat.lib.mkManifest {
  inherit pkgs;
  root = layat.lib.homeRoot;
  entries = builtins.listToAttrs (map (n: {
    name  = ".local/share/nvim/site/pack/plugins/start/${n}";  # キー = target
    value = { src = inputs.${n}; };
  }) plugins);
}
```

サブディレクトリを列挙するには、**実体化済みの store パス / `flake = false` input** を `builtins.readDir` する(生の `fetchFromGitHub` derivation は import-from-derivation を起こし、純粋な flake 評価を壊す)。

```nix
let
  skills = builtins.readDir "${inputs.claude-skills}/skills";
  names  = builtins.attrNames (nixpkgs.lib.filterAttrs (_: t: t == "directory") skills);
in
layat.lib.mkManifest {
  inherit pkgs;
  root = layat.lib.homeRoot;
  entries = builtins.listToAttrs (map (n: {
    name  = ".claude/skills/${n}";
    value = { src = inputs.claude-skills; subpath = "skills/${n}"; };
  }) names);
}
```

---

## コマンドリファレンス

CLI は CWD の entrypoint を発見する(`flake.nix` → `shell.nix` → `default.nix`)。`-f` で上書き可。各 `layat.<name>` は独立した profile で、`<name> = default` は名前省略時に `layat apply` が解決する。

```bash
layat apply [<name>]            # layat.<name> を適用(省略 = layat.default)。build・新世代コミット・配置を行う
layat apply <name> --dryrun     # read-only な計画: place/replace/remove/conflict/no-op、副作用ゼロ
layat apply <name> --recopy     # さらに全 copy target を src から無条件に上書きする
layat apply --manifest <farm>   # ビルド済み link-farm を直接適用(entrypoint 発見 / eval / build なし)
layat apply --all               # 全 layat.* を辞書順で適用。失敗しても続行する
layat apply --all --project-root # projectRoot config のみ適用(--home-root / --system-root も同様)
layat reset <name> [target...]  # 配置を撤去(profile 変更なし)。target 省略 = 全 entry
layat reset <name> --dryrun     # 削除対象を表示。副作用ゼロ
layat rollback <name>           # 直前の世代へ戻す(home mode のみ・name 必須)
layat list-generations <name>   # 世代を一覧(home mode のみ)
layat list-generations --all    # 全 home-mode config の世代を一覧
layat gitignore <name>          # 配置 target を .gitignore 用に stdout へ出力(書き込みなし・project mode のみ)
layat gitignore --all           # 全 projectRoot config の target をソート＋重複排除して出力
layat init <template>           # `nix flake init -t github:yasunori0418/layat#<template>` のラッパー
```

### グローバルフラグ

```text
-f, --file <path>   # entrypoint を明示指定(自動発見を上書き)
--root <path>       # 任意モードで解決済み root を上書き
--no-wait           # ロック競合時に待たずスキップ(shellHook 用。明示 apply は既定でブロック)
-v, --verbose       # 配置レポートを出力(サマリ＋target ごとの行)。既定は成功時サイレント
--version           # 埋め込みバージョンを表示して終了(cobra 既定書式 `layat version X.Y.Z`。-v は --verbose 割当済みのため短縮形なし)
--debug             # 内部 nix コマンドを stderr に出す(トラブルシュート用)
--project-root      # --all の限定子: projectRoot config のみ(--home-root / --system-root も同様)
--recopy            # apply の限定子: 全 copy target を src から上書き
--manifest <path>   # apply 専用: ビルド済み link-farm を直接適用
-y, --yes           # reset の確認プロンプトを省略(スクリプト / CI 用)
--generator <name>  # manifest 生成器(現状は `nix` のみ。「生成器の選択」を参照)
```

### 生成器の選択

生成器は明示指定でのみ選ばれ、置かれているファイルから推測されることはない。次のうち最初に指定されたものが使われる:

1. `--generator <name>`
2. 環境変数 `LAYAT_GENERATOR`
3. `-f` のディレクトリ(`-f` がファイルならそのファイルのあるディレクトリ)、`-f` 未指定なら CWD の `layat.toml`。親ディレクトリは探索しない
4. `$XDG_CONFIG_HOME/layat/config.toml`(既定 `~/.config/layat/config.toml`)
5. 既定の `nix`(現状唯一の生成器)

```toml
# layat.toml / config.toml — キーは `generator` のみ
generator = "nix"
```

空の値と `generator` を持たない設定ファイルは未指定とみなす。未知のキー・TOML の構文エラー・未知の生成器名は exit 1 と stderr 1 行で止まる(`--json` では `E_INPUT`)。`apply --manifest` はこれらを読まず `--generator`・`-f`・`--all` を拒否し、`prune` と `init` はこの仕組みを無視する。

### 出力と終了コード

- **既定では成功時サイレント。** 配置レポート、try-lock のスキップ通知、`apply --all` のサマリには `-v` が要る。`--debug` は内部 nix コマンドを表示する。`--quiet` は無い。
- **`--json`** はコマンド完了時に [outturn](https://github.com/yasunori0418/outturn) 準拠の JSON エンベロープ(1 文書)を stdout へ書く機械可読の第 2 契約。`-v`(stderr・人間向け)とは直交(全サブコマンドがペイロードを載せ、`--all` は config ごとに `SubjectResult` を列挙する。最小形のまま残るのは `reset --dryrun` のみ)。
- **ストリーム規律**：stdout は機械可読出力(`gitignore` 一覧、`apply --dryrun` 計画)専用で、既定 verbosity でも出力される。よって `layat gitignore <name> >> .gitignore` や `layat apply <name> --dryrun | ...` は安全にパイプできる。**警告とエラーは常に stderr へ出力され、サイレンスされない。**

| 終了コード | 意味 |
|---|---|
| `0` | 成功 / no-op / `--no-wait` の try-lock スキップ |
| `1` | 一般エラー(eval エラー、エンジン実行時エラー、`apply --all` の部分失敗) |
| `2` | `apply --dryrun` が conflict を検出(CI の事前ゲートに使える) |

### 挙動メモ

- **冪等。** 「配置したと記録し、かつ記録どおりを指している」stale link だけを除去する。外部 symlink は警告付きで置き換えられ、target にある実ファイル / ディレクトリはエラー。
- **世代** は layat 自前の Nix profile に乗り、`nix-env` / `nix-collect-garbage` で操作する。project mode は link-farm が不変なら新世代をコミットしないが、drift は修復する。
- **`apply --all`** は各 config を自分の profile 上で atomic に適用し、失敗しても続行し、いずれかが失敗すれば非ゼロで終了する。全体としては **atomic ではない**。
- **`reset`** は layat 管理 symlink と copy target を削除する(copy を消す唯一の手段)。name と確認または `-y` が必須で、profile / 世代には触れない。

---

## 他ツールとの比較

軸は「機能の有無」ではなく、**「配置をモジュール抽象の裏に隠すか、ユーザーが制御する純粋関数として露出させるか」**。

| ツール | 役割 | アプローチ | layat との違い |
|---|---|---|---|
| npins / niv | ソースのバージョン固定 | — | ファイルを配置しない(直交——layat と組み合わせる) |
| home-manager `home.file` | ファイル配置 ＋ 世代 | module(生成 / 宣言) | HM 必須。環境まるごとモデル。file モジュールを standalone に切り出せない |
| `mkOutOfStoreSymlink`(HM) | out-of-store symlink | module 内のヘルパー | HM 専用。layat は同等物を依存なしの明示関数として提供 |
| nixpkgs `linkFarm` / `symlinkJoin` | store 内 symlink ツリー | 純粋関数 | 出力は store *内* に留まり、任意の out-of-store パスへは置かない(layat は内部で使う) |
| `nix profile` | 世代管理 | — | 配置先が `~/.nix-profile` に固定。任意パス配置なし(layat はこれに乗る) |
| `systemd.tmpfiles`(`L`) | 任意パスへの宣言的 symlink | module(NixOS) | 低水準・NixOS 専用。copy / 世代 / フェッチ抽象なし |
| numtide/system-manager | 非 NixOS の `/etc` ＋ systemd ＋ パッケージ | module(`lib.evalModules`) | ドメインは重なるが **逆** のアプローチ。任意パス配置・HOME dotfiles・サブディレクトリ抽出なし |
| `git clone`(シェル) | clone して配置 | 命令的 | 再現性も Nix 統合もない |
| **layat** | フェッチ済みソースの独立配置 ＋ 世代 ＋ 明示的 out-of-store | **純粋関数・ユーザー管理** | — |

---

## MVP ステータス

| 領域 | ステータス |
|---|---|
| Standalone CLI(`apply` / `reset` / `rollback` / `list-generations` / `gitignore` / `init`) | 実装済み(コア) |
| project mode(`projectRoot`) | 実装済み(コア) |
| home mode(`homeRoot`) | 実装済み |
| home-manager モジュール | 実装済み——単一 profile(固定名 `default`)・役割分離なし |
| 世代 / rollback(home mode) | 実装済み |
| copy(place-once)/ out-of-store symlink | 実装済み |
| flake-parts モジュール | 実装済み |
| `manifest.json` スキーマ | v1 のみ。migration / 後方互換の仕組みはまだ無い |
| `--json` 機械可読出力 | 実装済み — 全サブコマンドで outturn 準拠エンベロープを返し、コマンドごとのペイロード(items / changes / info)と `--all` の config ごと `SubjectResult` も載る。最小形のまま残るのは `reset --dryrun` のみ |
| NixOS / nix-darwin モジュール | 将来 |
| system mode(`systemRoot` = `/`) | 将来(seam のみ。今選ぶと評価時エラー) |

**既知の制限 / 正直な注意点**

- boot / init / filesystem / partition 層は layat のドメインではない。
- クローンを削除すると `<state>/nix/profiles/layat/` 下に orphan な profile ディレクトリが残る(store は `nix-collect-garbage` で解放されるが、profile ディレクトリは残る)。MVP に `prune` コマンドは無い——手で消す。
- home-manager モジュールは MVP では役割を複数 profile に分けられない——その用途には standalone CLI を使う。

---

## ドキュメント

`docs/` は **README → 概要文書 → item の 3 層構造**。規範的な内容は item(1 ファイル 1 主張の
Markdown + YAML frontmatter)が持ち、概要文書は通読の入口として全体像と item への索引を担う。

- `docs/concept.md` — コンセプト(solution / use_case item と ADR への索引)、設計の哲学、north-star、既存ツールとの比較
- `docs/design.md` — 設計(design item への索引)
- `docs/spec.md` — 仕様(requirement item への索引)
- `docs/glossary.md` — 正準な英語用語(日本語対訳は `docs/glossary.ja.md`)

| ディレクトリ | 型 | prefix |
|---|---|---|
| `docs/solution/` | solution | `SOL` |
| `docs/use-cases/` | use_case | `UC` |
| `docs/requirements/` | requirement | `REQ` |
| `docs/design/` | design | `DSG` |
| `docs/infrastructure/` | infrastructure | `INF` |
| `docs/adr/` | adr(ツリーから分離) | `ADR` |

item はグラフを成す(型と関係の定義は `docs/model.yaml`)。辿るには devShell の `sara` を使う。

```bash
nix develop ./dev --command sara check    # グラフ全体の検証
```

グラフの検索(フル ID を取る `sara query`)は `docs/agents/domain.md` を参照。

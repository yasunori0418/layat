---
id: "DSG-f1d4c295-54ae-4def-951c-901b42dea008"
type: design
name: "buildEnv の合成環境は entry の src に置く配置 idiom で扱い、環境変数の配線は利用者側の idiom で行う"
satisfies:
  - "REQ-99ca5381-6c53-426c-b145-7b4297c53868"
  - "REQ-c1b3ca5f-d2f7-443c-bc4b-b18413ca97b9"
---
# DSG-f1d4c295-54ae-4def-951c-901b42dea008: buildEnv の合成環境は entry の src に置く配置 idiom で扱い、環境変数の配線は利用者側の idiom で行う

## 設計

「パッケージ群を `PATH` に載せたい」需要に対し、layat は専用の関数（`mkEnv` 等）を持たない
（→ ADR-0069）。`pkgs.buildEnv` の結果を entry の `src` に置けば、既存の配置プリミティブで
そのまま満たせるため。

- **REQ-99ca5381-6c53-426c-b145-7b4297c53868（`src` は path / set / marker の 3 種）**:
  `pkgs.buildEnv { … }` は derivation（set）なので、現行スキーマの `src` にそのまま渡せる。
  engine・manifest schema・passthru は変わらない
- **REQ-c1b3ca5f-d2f7-443c-bc4b-b18413ca97b9（モジュールは engine をキックするだけの配線）**:
  環境変数の設定はシェル・systemd・direnv の機構で、layat の関心外。layat は配線ファイル
  （env.sh / environment.d / `.envrc`）を生成せず、書き方を idiom として示すに留める

### 1. entry に buildEnv を直接指定する

```nix
dev-tools = layat.lib.mkManifest {
  inherit pkgs;
  root = layat.lib.homeRoot;
  entries.".local/share/layat/envs/dev-tools" = {
    src = pkgs.buildEnv {
      name = "dev-tools";
      paths = with pkgs; [ ripgrep fd ];
      extraOutputsToInstall = [ "man" ];
    };
  };
};
```

- **man が要るなら `extraOutputsToInstall = [ "man" ]` を明示する**。nixpkgs の多くのパッケージは
  man を別 output に分けており、素の buildEnv では `share/man` が空になる。closure が大きく
  なるので既定にはせず、利用者が判断する
- **rollback の粒度は config 全体になる**。buildEnv の entry を他の entry と同じ config に混ぜると、
  世代も rollback も config 単位になる
- **独立して update / rollback したい role は、上の例のように専用の config に 1 entry で置く**。
  config ごとに独立した profile になる

### 2. target の置き場所（例示）

| mode | target の例 |
|---|---|
| home | `.local/share/layat/envs/<name>` |
| project | `.layat/envs/<name>` |

これは例示であって規約ではない。layat は target の既定値を持たず、利用者が宣言した target を
そのまま使う。配線ファイルに書くパスも、利用者が書いた target そのものになる。

### 3. 環境変数の配線 idiom

以下のファイルは全て利用者が書く。layat の entry として配置してもよい。

**zsh（`.zshenv`）** — env ごとの `env.sh` を source するループ。`env.sh` を buildEnv の
`paths` に混ぜれば、配置先の `envs/<name>/env.sh` に置かれる。

```nix
paths = with pkgs; [ ripgrep fd ] ++ [
  (pkgs.writeTextDir "env.sh" ''
    export PATH="$HOME/.local/share/layat/envs/dev-tools/bin:$PATH"
    export MANPATH="$HOME/.local/share/layat/envs/dev-tools/share/man:''${MANPATH:-}"
  '')
];
```

```zsh
# ~/.zshenv
for f in ~/.local/share/layat/envs/*/env.sh(N); do source "$f"; done
```

- glob は名前順に展開される。順序が要るなら env 名に番号 prefix を付ける（`envs/10-dev-tools`）
- zsh の補完は `compinit` より前に fpath へ足す必要がある（`.zshrc` で
  `fpath=(~/.local/share/layat/envs/*/share/zsh/site-functions(N) $fpath)` → `compinit`）

**systemd user service（`environment.d`）** — フラグメントを置く。

```ini
# ~/.config/environment.d/50-layat-dev-tools.conf
PATH=${HOME}/.local/share/layat/envs/dev-tools/bin:${PATH}
```

- environment.d は `%h` 等の unit 指定子を解釈しない。`${HOME}` で書く（`%h` を解釈するのは unit
  ファイルの `Environment=`）

**direnv（project mode）** — `.envrc` で配置してから注入する。

```bash
# .envrc
use flake
watch_file flake.nix flake.lock
if command -v layat >/dev/null; then
  layat apply dev-tools --no-wait
fi
if [ -d .layat/envs/dev-tools ]; then
  PATH_add .layat/envs/dev-tools/bin
fi
```

- `layat apply <name>` は冪等で、変化が無ければ何もしない。cd のたびに実行してよい
- `watch_file` で flake の変更時に `.envrc` を再評価させる
- devShell に layat が無い・`--no-wait` で lock 競合をスキップした・apply が失敗した、のいずれかでは配置が無いことがある。存在チェックで分岐し、無ければ `PATH_add` をしない

### 4. 制約

- **direnv が注入できるのは環境変数だけ**。zsh の fpath / compinit はシェル内部の状態なので、
  project に居る間だけ補完を有効にする用途はスコープ外とする
- **`LD_LIBRARY_PATH` は設定しない**。Nix のバイナリは RPATH でライブラリを解決するので不要で、
  設定すると他のバイナリまで別版のライブラリを読み込むので、むしろ有害
- **layat は配線ファイルを生成しない**。env.sh / environment.d / `.envrc` は利用者の Nix 式や
  スクリプトで作る

## 出典

ADR-0069「`mkEnv` を足さず、buildEnv の合成環境は entry の配置 idiom で満たす」の決定節。
同 ADR は配置と環境変数配線の書き方を idiom として docs に置くと決めており、
本 item がその実体。

---
id: "ADR-0068"
type: adr
name: "lib に事前展開した apply コマンドを組む関数を公開する"
status: 採用
issues:
  - "#530"
  - "#493"
origin: "Issue #530 の grilling（2026-10-25）"
justifies:
  - "REQ-6c51facc-fa7a-4b5e-939b-44e7578d7a4f"
  - "DSG-e4d5db6b-5bac-4fd6-b2fb-ed07bdec30f5"
references:
  - "ADR-0029"
---
# ADR-0068: lib に事前展開した apply コマンドを組む関数を公開する

- ステータス: 採用
- 日付: 2026-10-25
- 関連: ADR-0029, GitHub Issue #530, #493
- 起点: Issue #530 の grilling（2026-10-25）

## 背景

project mode の devShell は `shellHook` で `layat apply <name>` を実行する。CLI はそのたびに
entrypoint を cwd から探し、`nix eval` / `nix build` を実行する。HM / NixOS / nix-darwin モジュールの
`modules/kick.nix` は、link-farm を評価時に作り `apply --all --manifest <name>=<farm> …` の 1 行を
組んで、これらを省いている。ただしこの組み立ては modules の内部にあり、devShell からは使えない。

`apply --all --manifest` は rootKind を制限せず、`--project-root` などの root フィルタも
使える。一方 `lib/` は nixpkgs.lib だけに依存する規約なので、pin 版 CLI（`packages.<system>.layat`）を
`lib/` の中から参照できない。

## 決定

### 1. `lib.mkApplyCommand { pkgs, layatPackage, manifests, root }` を公開する

`lib/apply-command.nix` に置き、`lib/default.nix` から公開する。戻り値は
`<layatPackage の実行ファイル> apply --all [--<root>-root] --manifest <name>=<farm> …` の文字列。
`manifests` は config 名 → link-farm の attrset で、flake-parts なら `config.layat`、素の flake なら
`self.layat.${system}` を渡す。CLI は `layatPackage` で受ける。

### 2. `root` は既定値のない必須の enum にする

`"all"` / `"project"` / `"home"` / `"system"` だけを受ける。`"all"` はフィルタを付けず、他は対応する
`--<root>-root` を付ける。省略と範囲外の値は評価時に throw する。

### 3. config の選別は CLI だけが行う

`--manifest` には常に `manifests` の全 config を名前の辞書順に並べる。Nix 側で `passthru.rootKind`
による選別はしない。選ばれない root の farm もビルド依存に入るが、それは許容する。

### 4. 実行時フラグは引数にしない

`--jobs` / `--recopy` / `--backup` / `--no-wait` は呼び出し側が文字列の後ろに足す。

### 5. kick.nix も `mkApplyCommand` を使う

`root = "all"` で呼び、`--jobs` / `--recopy` / `--backup` を後ろに足す。HM / NixOS / nix-darwin
モジュールの出力は変えない。flake-parts モジュール（ADR-0029）には option も注入も足さない。

## 根拠

- **事前展開**: devShell の依存として farm を先にビルドすれば、devShell に入るたびの eval / build /
  entrypoint 探索が要らなくなり、cwd にも依存しない。
- **CLI を引数で受ける**: `lib/` を純粋に保ったまま、利用者が pin した CLI を使える。
- **root を必須にする**: 素の `--all` を devShell から実行すると home mode の config まで `$HOME` に
  置かれる。既定値を持たせず、どの root を置くかを呼び出し側に明示させる。
- **CLI だけで選別する**: rootKind による選別は CLI がすでに持つ。Nix 側にも持つと、同じ判定が
  2 箇所に分かれる。
- **kick.nix の共用**: 文字列の組み立てを 1 箇所に寄せ、モジュールと devShell の間で形がずれない
  ようにする。

## 影響

- 要求: `mkApplyCommand` の要求（REQ-6c51facc-fa7a-4b5e-939b-44e7578d7a4f）を起こし、lib の依存
  制約（REQ-d85f0cef-0f1e-4897-a841-41b61a8dae51）の列挙に `mkApplyCommand` を加える。
- 設計: lib の分割（DSG-e4d5db6b-5bac-4fd6-b2fb-ed07bdec30f5）に `lib/apply-command.nix` を加える。
- 利用者文書: README に、devShell で `mkApplyCommand` を使う選択肢を足す。汎用の正準形
  `layat apply <name> --no-wait` は据え置く。

## 棄却した代替案

- **flake-parts モジュールに option を足して devShell へ注入する**: モジュールが devShell の形を
  決めることになる。関数を公開すれば、利用者が任意の shell に埋められる。
- **`lib` output に pin 版 CLI を既定で注入する**: `lib/` が自 flake の packages に依存し、
  nixpkgs.lib だけに依存する規約が崩れる。
- **Nix 側で rootKind を選別する**: CLI の root フィルタと判定が二重になる。
- **実行時フラグを引数に取る**: 関数の引数が CLI のフラグ追加に追従し続けることになる。
- **template に starter 例を足す**: 汎用の正準形は `layat apply <name>` のままにし、事前展開は
  README で選択肢として示すに留める。

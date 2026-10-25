---
id: "REQ-6c51facc-fa7a-4b5e-939b-44e7578d7a4f"
type: requirement
name: "lib.mkApplyCommand は事前展開した apply --all --manifest のコマンド行を返す"
derives_from:
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  The lib SHALL expose `mkApplyCommand { pkgs, layatPackage, manifests, root }`, which
  returns the string `<exe of layatPackage> apply --all [--<root>-root] --manifest
  <name>=<farm> ...`. `--manifest` SHALL list every config in `manifests` in lexicographic
  name order, each `<name>=<farm>` quoted with `lib.escapeShellArg`. `root` SHALL be a
  required argument taking only `"all"` / `"project"` / `"home"` / `"system"`: `"all"`
  SHALL add no root filter, and the others SHALL add `--project-root` / `--home-root` /
  `--system-root`. An omitted or out-of-range `root` SHALL throw at evaluation time. The
  function SHALL NOT select configs by `rootKind` on the Nix side, and SHALL NOT take
  runtime flags (`--jobs` / `--recopy` / `--backup` / `--no-wait`) as arguments.
specification_ja: |
  `lib` は `mkApplyCommand { pkgs, layatPackage, manifests, root }` を公開し、
  `<layatPackage の実行ファイル> apply --all [--<root>-root] --manifest <name>=<farm> …` の
  文字列を返さなければならない。`--manifest` は `manifests` の全 config を名前の辞書順に並べ、
  各 `<name>=<farm>` を `lib.escapeShellArg` で包まなければならない。`root` は `"all"` /
  `"project"` / `"home"` / `"system"` だけを受ける必須引数でなければならず、`"all"` は root
  フィルタを付けず、他はそれぞれ `--project-root` / `--home-root` / `--system-root` を付けなければ
  ならない。省略・範囲外の `root` は評価時に throw しなければならない。Nix 側で `rootKind` による
  config の選別をしてはならず、実行時フラグ（`--jobs` / `--recopy` / `--backup` / `--no-wait`）を
  引数に取ってはならない。
---
# REQ-6c51facc-fa7a-4b5e-939b-44e7578d7a4f: lib.mkApplyCommand は事前展開した apply --all --manifest のコマンド行を返す

## 仕様

`mkApplyCommand` は link-farm の集合を `apply --all --manifest` の 1 行へ事前展開する。
devShell の `shellHook` に埋めれば、devShell に入るたびの eval / build / entrypoint 探索が要らなくなり、
farm は devShell の依存として先にビルドされる。

- `manifests` は config 名 → link-farm の attrset。flake-parts なら `config.layat`、素の flake
  なら `self.layat.${system}` を渡す
- `layatPackage` は引数で受ける。`lib` は nixpkgs.lib だけに依存するので、pin 版 CLI を
  自分では参照しない（REQ-d85f0cef-0f1e-4897-a841-41b61a8dae51）
- config の選別は CLI の root フィルタ（REQ-d95b814f-aa7a-470e-9320-c14f9c14da7b）だけが行う。
  選ばれない root の farm もビルド依存に入る
- 実行時フラグは呼び出し側が文字列の後ろに足す。HM / NixOS / nix-darwin モジュールは
  `root = "all"` で呼び、`--jobs` / `--recopy` / `--backup` を足す

## 出典

ADR-0068「lib に事前展開した apply コマンドを組む関数を公開する」。

---
id: "RISK-81cadaa3-3ed2-4cb0-bc74-72ff3afa9cce"
type: risk
name: "--help の nix コマンド開示が実際に実行する nix 呼び出しからずれる"
section: cli-json
status: stable
likelihood: medium
impact: low
level: low
threatens:
  - "REQ-4ffda99a-7062-4c00-915f-70b525cb215b"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:36:11+09:00"
---
# --help の nix コマンド開示が実際に実行する nix 呼び出しからずれる

## リスク

`layat --help` の開示ブロック（`cmd/layat/main.go` の `rootCmdLong`）は、
[内部実行する nix コマンドを開示し世代の切替と GC は標準の nix コマンドへ委譲する](../requirements/nix-command-transparency-and-delegation.md) の「内部実行する nix コマンドを開示する」規範の唯一の実装になっている。
文面は手書きの定数で、実際に nix の引数を組み立てるのは `internal/generator/nixgen/nixgen.go` と
`cmd/layat/init.go` である。両者は独立に編集されるため、サブコマンドの追加、build の段数の変更、
フラグの差し替えがあっても文面は追随しないまま残りうる。

**顕在化したときに起きること**: 開示を頼りに手で nix を実行したユーザーが、layat の実際の
動作と違うコマンドを打ち、違う評価対象を読んだり gcroot の有無を取り違えたりする。nix を
実行するサブコマンドがブロックに載っていないと、そのサブコマンドが nix を実行すること自体が
利用者から見えなくなる。

## 評価

- likelihood: medium — 文面と組み立ては別ファイルにあり、nix 呼び出しの所在はこれまでにも
  移っている。サブコマンドの追加や build の処理の変更は通常の改修で起こりうる
- impact: low — ずれるのは開示の文面で、layat 自身の配置や記録は破損しない。手で実行した
  利用者が誤ったコマンドを打つ範囲に留まる

## 対処

[--help の開示ブロックがサブコマンドごとに実際の nix 呼び出しと一致する](../test/cli-json/help-disclosure-matches-nixgen.md)（開示ブロックとサブコマンドごとの nix 呼び出しの一致）で緩和する。

**緩和しきれない残余**:

- 開示ブロックの比較は、rootKind が fixed でない flake エントリポイントの使い方に限られる。
  fixed のときに追加で実行される `.root` の eval と、legacy エントリポイントの `-f` 形の文面は
  比較しない
- `gitignore --all` が実行する一括 eval と build は、ブロックに行が無く、比較の対象にもならない
- `apply --dryrun` が実行する build（`--no-link --print-out-paths`）は比較しない
- 比較するサブコマンドはテストに列挙したものに限られる。新しく nix を実行するサブコマンドが
  増えても、テストに足さない限り検出しない
- 世代バックエンドはテストで fake に差し替えるため、apply と rollback が実行する `nix-env` の
  呼び出しは比較しない。[内部実行する nix コマンドを開示し世代の切替と GC は標準の nix コマンドへ委譲する](../requirements/nix-command-transparency-and-delegation.md) が開示を求めるのは manifest 生成器の
  コマンドで、`nix-env` はブロックに載っていない
- prune / status が nix を実行しないことの比較で観測するのは `nix` 実行ファイルの呼び出しだけで、
  `nix-env` などほかの nix 系コマンドの実行は検出しない

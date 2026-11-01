---
id: "RISK-81cadaa3-3ed2-4cb0-bc74-72ff3afa9cce"
type: risk
name: "--help の nix コマンド開示が実際に実行する nix 呼び出しからずれる"
likelihood: medium
impact: low
level: low
threatens:
  - "REQ-4ffda99a-7062-4c00-915f-70b525cb215b"
---
# RISK-81cadaa3-3ed2-4cb0-bc74-72ff3afa9cce: --help の nix コマンド開示が実際に実行する nix 呼び出しからずれる

## リスク

`layat --help` の開示ブロック（`cmd/layat/main.go` の `rootCmdLong`）は、
REQ-4ffda99a-7062-4c00-915f-70b525cb215b の「内部実行する nix コマンドを開示する」規範の唯一の実装になっている。
文面は手書きの定数で、実際に nix の引数を組み立てるのは `internal/generator/nixgen/nixgen.go` と
`cmd/layat/init.go` である。両者は独立に編集されるため、サブコマンドの追加、build の段数の変更、
フラグの差し替えがあっても文面は追随しないまま残りうる。

**顕在化したときに起きること**: 開示を頼りに手で nix を実行したユーザーが、layat の実際の
動作と違うコマンドを打つ（違う評価対象を読む、gcroot の有無を取り違えるなど）。nix を
実行するサブコマンドがブロックから漏れると、そのサブコマンドが nix を実行すること自体が
利用者から見えなくなる。

## 評価

- likelihood: medium — 文面と組み立ては別ファイルにあり、nix 呼び出しの所在はこれまでにも
  移っている。サブコマンドの追加や build 経路の変更は通常の改修で起こりうる
- impact: low — ずれるのは開示の文面で、layat 自身の配置や記録は壊れない。手で実行した
  利用者が誤ったコマンドを打つ範囲に留まる

## 対処

TC-b4262641-776c-427d-8331-1a94b29fa568（開示ブロックとサブコマンドごとの nix 呼び出しの一致）で緩和する。

**緩和しきれない残余**:

- 開示ブロックの照合は、rootKind が fixed でない flake エントリポイントの経路に限られる。
  fixed のときに追加で実行される `.root` の eval と、legacy エントリポイントの `-f` 形の文面は
  照合しない
- `gitignore --all` が実行する一括 eval と build は、ブロックに行が無く、照合の対象にもならない
- `apply --dryrun` が実行する build（`--no-link --print-out-paths`）は照合しない
- 照合するサブコマンドはテストに列挙したものに限られる。新しく nix を実行するサブコマンドが
  増えても、テストに足さない限り検出しない
- 世代バックエンドはテストで fake に差し替えるため、apply と rollback が実行する `nix-env` の
  呼び出しは照合しない（REQ-4ffda99a-7062-4c00-915f-70b525cb215b が開示を求めるのは manifest 生成器の
  コマンドで、`nix-env` はブロックに載っていない）
- prune / status が nix を実行しないことの照合で観測するのは `nix` 実行ファイルの呼び出しだけで、
  `nix-env` などほかの nix 系コマンドの実行は検出しない

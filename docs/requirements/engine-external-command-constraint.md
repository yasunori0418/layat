---
id: "REQ-6c4e174a-4d16-477a-96ff-17cb4eb5b564"
type: requirement
name: "engine が叩く外部コマンドは git のみに限る"
section: architecture
status: stable
derives_from:
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  The engine SHALL invoke only `git` (for obtaining the toplevel) as an external command.
  Place / replace / remove and conservative stale removal SHALL be performed with native
  filesystem operations. Generation operations (commit, list, switch) SHALL go through the
  generation backend contract ([engine は世代操作を世代バックエンドの契約（Commit / List / Switch / UserBase / SystemBase）越しに行う](generation-backend-contract.md)) and SHALL NOT be
  invoked by the engine as external commands; the default backend nixprofile runs
  `nix-env` inside its own package.
specification_ja: |
  engine が外部コマンドとして叩いてよいのは `git`（toplevel の取得）のみでなければならない。
  place / replace / remove と古くなった配置の保守的な除去はネイティブ FS 操作で行わなければならない。
  世代操作（コミット・一覧・切替）は世代バックエンドの契約
  （[engine は世代操作を世代バックエンドの契約（Commit / List / Switch / UserBase / SystemBase）越しに行う](generation-backend-contract.md)）越しに行わなければならず、engine が外部コマンド
  として呼び出してはならない。既定バックエンド nixprofile は自パッケージの内側で `nix-env` を
  実行する。
risk_exempt: "engine が実行するコマンドを限る依存の制約であり、レビューで守る。"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T13:39:28+09:00"
---
# engine が叩く外部コマンドは git のみに限る

## 仕様

engine は `manifest.json` を入力に取り、外部コマンドは `git`（toplevel の取得）だけを叩く。

配置そのもの（place / replace / remove）と古くなった配置の保守的な除去は **ネイティブ FS 操作**で
行う。世代操作（コミット・一覧・切替）は世代バックエンドの契約越しで、engine は世代を誰が
管理するかに関知しない。既定バックエンド nixprofile が `nix-env` を自パッケージの内側で
実行する（[engine は世代操作を世代バックエンドの契約（Commit / List / Switch / UserBase / SystemBase）越しに行う](generation-backend-contract.md)）。

外部コマンドの依存を 1 つに閉じることで、engine の動作が実行環境のコマンド群に左右される
範囲を限定する。2 層構成そのものは [layat は CLI とエンジンの 2 層で構成する](two-layer-architecture.md) が規定する。

## 出典

`docs/spec.md`「アーキテクチャ概要」の構成図・engine 層の 2 行
（「`manifest.json` を入力に取り nix(profile)/git(toplevel)のみ叩く」「ネイティブ FS
操作で place/replace/remove、保守的 stale 除去、nix-env --set」）。

決定の実体は [engine が直接叩く外部コマンドを git に限る](../adr/engine-external-command-git-only.md)で、`rsync` / `ln` の runtime 依存を落とし FS 操作を
ネイティブ Go（`os.Symlink` / ネイティブ再帰コピー）で行うことを定めている。外部コマンドを
`git` のみに限り、profile 操作を世代バックエンドの契約へ移したのは [engine が直接叩く外部コマンドを git に限る](../adr/engine-external-command-git-only.md)。

---
id: "REQ-6c4e174a-4d16-477a-96ff-17cb4eb5b564"
type: requirement
name: "engine が叩く外部コマンドは git のみに限る"
derives_from:
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  The engine SHALL invoke only `git` (for obtaining the toplevel) as an external command.
  Place / replace / remove and conservative stale removal SHALL be performed with native
  filesystem operations. Generation operations (commit, list, switch) SHALL go through the
  generation backend contract (REQ-833205ca-adfc-4c86-bf66-ae4d8fe4885b) and SHALL NOT be
  invoked by the engine as external commands; the default backend nixprofile runs
  `nix-env` inside its own package.
specification_ja: |
  engine が外部コマンドとして叩いてよいのは `git`（toplevel の取得）のみでなければならない。
  place / replace / remove と保守的 stale 除去はネイティブ FS 操作で行わなければならない。
  世代操作（コミット・一覧・切替）は世代バックエンドの契約
  （REQ-833205ca-adfc-4c86-bf66-ae4d8fe4885b）越しに行わなければならず、engine が外部コマンド
  として呼び出してはならない。既定バックエンド nixprofile は自パッケージの内側で `nix-env` を
  実行する。
---
# REQ-6c4e174a-4d16-477a-96ff-17cb4eb5b564: engine が叩く外部コマンドは git のみに限る

## 仕様

engine は `manifest.json` を入力に取り、外部コマンドは `git`（toplevel の取得）だけを叩く。

配置そのもの（place / replace / remove）と保守的 stale 除去は **ネイティブ FS 操作**で
行う。世代操作（コミット・一覧・切替）は世代バックエンドの契約越しで、engine は世代を誰が
管理するかに関知しない。既定バックエンド nixprofile が `nix-env` を自パッケージの内側で
実行する（REQ-833205ca-adfc-4c86-bf66-ae4d8fe4885b）。

外部コマンドの依存を 1 つに閉じることで、engine の動作が実行環境のコマンド群に左右される
範囲を限定する。2 層構成そのものは REQ-f4d7d4ab-fbdb-48c6-b29f-08dd88e72645 が規定する。

## 出典

`docs/spec.md`「アーキテクチャ概要」の構成図・engine 層の 2 行
（「`manifest.json` を入力に取り nix(profile)/git(toplevel)のみ叩く」「ネイティブ FS
操作で place/replace/remove、保守的 stale 除去、nix-env --set」）。

決定の実体は ADR-0006「決定」節で、`rsync` / `ln` の runtime 依存を落とし FS 操作を
ネイティブ Go（`os.Symlink` / ネイティブ再帰コピー）で行うことを定めている。外部コマンドを
`git` のみに限り、profile 操作を世代バックエンドの契約へ移したのは ADR-0057「世代バックエンドの
契約」§4。

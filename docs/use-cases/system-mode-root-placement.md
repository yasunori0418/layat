---
id: "UC-6b8d31ea-3d5a-4334-aac7-8706df190e7d"
type: use_case
name: "system mode で `/` 配下へ配置し、root 権限で世代とともに管理する"
status: stable
refines:
  - "SOL-9fcd1d6e-6204-42e6-92bb-1faf966f0b3e"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:30:05+09:00"
---
# UC-6b8d31ea-3d5a-4334-aac7-8706df190e7d: system mode で `/` 配下へ配置し、root 権限で世代とともに管理する

## 使われ方

管理者は、外部リポジトリの中身を `/etc` などシステム全体の場所へ配置したい。config の
`root` に `layat.lib.systemRoot` を書き、root 権限で `apply` する。NixOS / nix-darwin の
モジュールを使わず CLI 単体で使う環境でも、home mode と同じ CLI で配置と世代管理ができる。

```bash
layat apply <name>             # root 権限で / 配下へ配置し、新しい世代を積む
layat list-generations <name>  # 非 root でも世代を閲覧できる
layat rollback <name>          # root 権限で前の世代へ戻す
layat reset <name>             # root 権限で配置を撤去する
layat apply <name> --dryrun    # 非 root でも計画を確認できる
```

システムの状態は特定ユーザーの home に置かない。世代は nix の system profile の慣習どおり
`/nix/var/nix/profiles/layat` の下に、最後に配置した内容の記録は `/var/lib/layat` の下に置く。
`/` 配下への配置は project mode と違って永続的なので、home mode と同じく世代を公開する。

非 root で変更系の操作を実行すると、何かに触れる前に「root 権限が要る」と明確に止まる。
最初に触れた target の権限エラーで途中まで進んだ状態にはならない。

## この使われ方が要求すること

- `systemRoot` の config の root が実行時に `/` へ解決され、`--root` で上書きできること
- profile と last-applied 記録が、実行ユーザーの state ではなくシステムの基底に置かれること
- 変更系の操作（`apply` / `reset` / `rollback`）が root 権限を前段で確認し、非 root では FS に
  触れずに止まること
- 計画確認（`apply --dryrun` / `reset --dryrun`）と世代閲覧（`list-generations`）は非 root でも
  できること
- home / project / fixed の config の配置と状態の置き場が変わらないこと

## 出典

[rootKind に system を足しても schemaVersion を上げない](../adr/system-rootkind-without-version-bump.md)、
[root は 3 つの marker で表す](../adr/three-root-markers.md)、
[system mode の変更系操作の前に engine が euid 0 を確認する](../adr/system-mode-euid-gate.md)、
[profile の基底ディレクトリをモードで切り替える](../adr/profile-base-directory-per-mode.md)、
[profile は解決済み root の roothash でキーする](../adr/profile-keyed-by-roothash-of-resolved-root.md)、
[prune は system 基底も走査し、権限の無い系列を見送る](../adr/prune-scans-system-base-skips-unpermitted.md)、
[module と project では世代操作を公開しない](../adr/module-and-project-hide-generation-ops.md)、
[system mode の last-applied 記録を /var/lib/layat に置く](../adr/system-record-base-var-lib.md)、
[state の基底を設定で上書きさせない](../adr/state-bases-not-configurable.md)。

---
id: "REQ-0d1f557b-7ae7-41ea-b17b-cdb5c80935d4"
type: requirement
name: "system mode は root を / に解決し、変更系操作の前に euid 0 を確認し、profile と last-applied 記録を system 基底に置く"
status: stable
specification: |
  When the root kind is `system`, the engine SHALL resolve the placement root to `/`, a
  `--root` override taking precedence as in every other mode. The profileDir of a
  system-mode config SHALL be keyed under the system profile base that the generation
  backend owns (`SystemBase`) instead of the user state base: directly on `<name>` without
  `--root`, and on `<roothash>/<name>` with the backref `.root` at the `<roothash>` level
  when `--root` is given. The last-applied record of a system-mode config SHALL be placed
  at `/var/lib/layat/<key>/last-applied.json`, where `<key>` is the same key as its
  profileDir. The mutating operations `apply`, `reset` and `rollback` on a system-mode
  config SHALL check that the effective user ID is 0 before creating any directory, taking
  the flock or otherwise touching the filesystem, and when it is not 0 they SHALL stop with
  an error that wraps the permission-denied condition and tells the user to rerun with root
  privileges. `apply --dryrun`, `reset --dryrun`, `list-generations` and `status` SHALL NOT perform
  this check. The locations of the profileDir and the record of home, project and fixed
  configs SHALL NOT change.
specification_ja: |
  rootKind が `system` のとき、engine は配置 root を `/` に解決しなければならない。`--root`
  上書きは他のモードと同じく優先しなければならない。system mode の config の profileDir は、
  ユーザー state 基底ではなく世代バックエンドが所有する system profile 基底（`SystemBase`）の
  下に、`--root` なしでは `<name>` 直キー、`--root` ありでは `<roothash>/<name>` と
  `<roothash>` 階層の backref `.root` でキーしなければならない。system mode の config の
  last-applied 記録は `/var/lib/layat/<鍵>/last-applied.json` に置かなければならず、`<鍵>` は
  その profileDir と同じ鍵でなければならない。system mode の config に対する変更系操作
  （`apply` / `reset` / `rollback`）は、ディレクトリの作成・flock の取得・その他 FS への接触
  より前に実効ユーザー ID が 0 であることを確認しなければならず、0 でなければ権限不足を wrap し
  root 権限での再実行を案内するエラーで停止しなければならない。`apply --dryrun` /
  `reset --dryrun` / `list-generations` / `status` はこの確認をしてはならない。home / project / fixed の
  config の profileDir と記録の置き場は変えてはならない。
derives_from:
  - "UC-6b8d31ea-3d5a-4334-aac7-8706df190e7d"
  - "UC-0b6f60cb-3e98-4ee7-8929-4d94a29f0af6"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:54:59+09:00"
---
# system mode は root を / に解決し、変更系操作の前に euid 0 を確認し、profile と last-applied 記録を system 基底に置く

## 仕様

| 対象 | `--root` なし | `--root` あり |
|---|---|---|
| root | `/` | 上書き後の絶対パス |
| profileDir | `<SystemBase>/<name>` | `<SystemBase>/<roothash>/<name>`（backref `<SystemBase>/<roothash>/.root`）|
| last-applied 記録 | `/var/lib/layat/<name>/last-applied.json` | `/var/lib/layat/<roothash>/<name>/last-applied.json`（backref 同階層）|

`<SystemBase>` は世代バックエンドの契約が返す system 基底（既定実装 nixprofile では
`/nix/var/nix/profiles/layat`）。鍵の規則は home と同型で、`<name>` 直キーは root が `/` に
固定されることで成り立つ。

- **euid の条件**: `apply` は dryrun の短絡より後・profileDir 作成より前、`reset` は dryrun
  以外のとき profileDir の解決直後、`rollback` は profileDir の解決直後に確認する。どれも
  flock を取らず、ディレクトリも記録も作らないまま止まる。エラーは `fs.ErrPermission` を wrap
  する（CLI はこれを `E_PERMISSION` に写す）。
- **確認しない場面**: `apply --dryrun` / `reset --dryrun` / `list-generations` / `status` は
  読み取りだけで動くので、非 root でも計画確認・世代閲覧・配置内容の確認ができる。
- **変えないもの**: home / project / fixed の profileDir と記録の置き場。

profileDir の鍵の一般規則は [profileDir は home と system のみ name 直キーとし、fixed root と --root 上書きは roothash でキーする](profiledir-keying.md) の担当。記録の内容・
書き込み・比較の基準の優先順位は [apply は最後に配置した manifest を last-applied 記録に書き、古くなった配置の除去の比較の基準は記録を優先して読む](last-applied-record.md) の担当で、system mode
でも変わらない。本 item が決めるのは置き場の基底だけ。

## 出典

[rootKind に system を足しても schemaVersion を上げない](../adr/system-rootkind-without-version-bump.md)・[root は 3 つの marker で表す](../adr/three-root-markers.md)・[system mode の変更系操作の前に engine が euid 0 を確認する](../adr/system-mode-euid-gate.md)・[profile の基底ディレクトリをモードで切り替える](../adr/profile-base-directory-per-mode.md)・[profile は解決済み root の roothash でキーする](../adr/profile-keyed-by-roothash-of-resolved-root.md)・[prune は system 基底も走査し、権限の無い系列を見送る](../adr/prune-scans-system-base-skips-unpermitted.md)、
[system mode の last-applied 記録を /var/lib/layat に置く](../adr/system-record-base-var-lib.md)・[state の基底を設定で上書きさせない](../adr/state-bases-not-configurable.md)。

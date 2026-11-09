---
id: "REQ-535b811d-dfc5-4eac-92db-737e70eb5415"
type: requirement
name: "apply --all は rootKind を 1 回の一括 eval で取る"
status: stable
derives_from:
  - "UC-1c280dce-7c72-44c0-95ea-d06344f62a47"
specification: |
  `apply --all` SHALL obtain the rootKinds in a single batched eval: a map from config
  name to `{ rootKind, targets }` (plus `root` for a fixed root) SHALL be obtained at once with
  `nix eval <ep>#layat.<system> --apply 'cs: builtins.mapAttrs (_: c: { rootKind = c.rootKind; targets = c.targets; } // …) cs' --json`
  (for a legacy entrypoint, which has no per-system dimension,
  `nix eval -f <ep> layat --apply 'cs: …' --json`), and each `profileDir` SHALL be
  determined from it. Filters such as `--project-root` SHALL also be dispatched from this
  result. Only the build SHALL be performed N times, once per config, for the sake of
  atomicity. This SHALL fix the eval process startup cost at 1 rather than N.
specification_ja: |
  `apply --all` は rootKind を 1 回の一括 eval で取らなければならない。
  `nix eval <ep>#layat.<system> --apply 'cs: builtins.mapAttrs (_: c: { rootKind = c.rootKind; targets = c.targets; } // …) cs' --json`
  （legacy は per-system 次元なし: `nix eval -f <ep> layat --apply 'cs: …' --json`）で
  config 名 → `{ rootKind, targets }`（fixed root は `root` も）マップを 1 回で取得し、
  各 profileDir を確定しなければならない。
  `--project-root` 等のフィルタもこの結果で振り分けなければならない。build だけは
  atomic 性のため config ごと N 回行わなければならない。eval プロセス起動コストを
  N→1 に固定しなければならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:29:58+09:00"
---
# REQ-535b811d-dfc5-4eac-92db-737e70eb5415: apply --all は rootKind を 1 回の一括 eval で取る

## 仕様

**`apply --all` は rootKind を 1 回の一括 eval で取る**。
`nix eval <ep>#layat.<system> --apply 'cs: builtins.mapAttrs (_: c: { rootKind = c.rootKind; targets = c.targets; } // …) cs' --json`
（legacy は per-system 次元なし: `nix eval -f <ep> layat --apply 'cs: …' --json`）で
config 名 → `{ rootKind, targets }`（fixed root は `root` も）マップを 1 回で取得し、各 profileDir を確定する。`--project-root` 等の
フィルタもこの結果で振り分ける。build だけは atomic 性のため config ごと N 回。
eval プロセス起動コストを N→1 に固定する。

フィルタそのものは [--all は root モードフィルタで対象 config を絞れる](apply-all-root-mode-filter.md) の担当。

## 出典

`docs/spec.md`「CLI 仕様」→「実行フロー」の `apply --all` 一括 eval の箇条書き。

決定の実体は [apply --all は config ごとの情報を 1 回の eval でまとめて取る](../adr/apply-all-single-eval.md)で、legacy entrypoint の eval 形は [legacy entrypoint の eval と build は新 CLI の -f 形で行う](../adr/legacy-eval-via-nix-file-flag.md)。
`targets` を取得に加えたのは [config 間の target 衝突は CLI が判定し engine に持ち込まない](../adr/cross-config-conflict-judged-by-cli.md)（前段での衝突の確認 [apply --all は build 前に選択 config 間の target 衝突を検出して停止する](apply-all-cross-config-preflight.md) が使う）。

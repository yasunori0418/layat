# layat E2E ハーネス（非 NixOS）

実 nix を使って `layat` を end-to-end に駆動し、**「非 NixOS でも nix さえあれば動く」**という
主張を検証する bash ハーネス（→ `docs/design.md`「テスト戦略」・ADR-0012）。

`lib`（nix-unit / namaka の評価テスト）や配置エンジン（Go の tmpdir 統合テスト）が
**nix を使わずに**純ロジック / FS 操作を検証するのに対し、ここでは **flake entrypoint からの
`nix build` / `nix eval` / `nix-env --set` を含む実経路**を一気通貫で回す。

## 実行

```bash
# CI（ubuntu-latest）と同じ起動方法。ci devShell が layat / git / jq / coreutils を提供する。
nix develop '.?dir=dev#ci' -c tests/e2e/run.sh
```

`run.sh` は `scenarios/*.sh` を辞書順に独立プロセスで実行し、1 つでも失敗すれば非ゼロ終了する。
各シナリオは隔離した一時 `$HOME` / `$XDG_STATE_HOME`（`mktemp -d`）下で動き、ランナーの実
profile / home を汚さない（`tests/e2e/lib.sh` の `e2e_isolate`）。`prune` が state 基底と並べて
走査する system 基底（→ ADR-0036 §3）は絶対パスで `$HOME` の差し替えでは動かないため、
`e2e_isolate` が `LAYAT_SYSTEM_PROFILE_BASE` で隔離先へ向ける（→ `cmd/layat/prune.go`）。偽 src は fixture flake
ディレクトリ内の相対パス（eval 時に store へコピー）か、out-of-store 用の live ディレクトリで用意する。
fixture flake は `layat` を `path:<repo>` input で参照し、`nixpkgs` / `home-manager` は layat の
`flake.lock` pin に `follows` させてオフライン評価する。

## シナリオ範囲

| シナリオ | 検証内容 |
|---|---|
| `01-project` | project mode。一時 git repo で `layat apply` → git toplevel 配下に store symlink 配置・再 apply の冪等性 |
| `02-home`    | home mode。仮 `$HOME` で apply → `$HOME` 配下配置 + profile 世代コミット、entry 入替で世代を進め `layat rollback` で前世代へ復帰 |
| `03-stale`   | stale 除去。entry を config から削除 → 再 apply で旧 symlink が消える（保守的不変条件） |
| `04-copy`    | copy place-once / out-of-store。copy が通常ファイル（書込可）・place-once 冪等（ローカル編集を破棄しない）・out-of-store の live symlink |
| `05-hm`      | HM module。home-manager standalone configuration を非 NixOS で評価・activate し、activation が engine を起動して配置すること。`layat.entries` と別名の config を併せ持つ構成で 2 profile が独立した世代を持つこと、2 config が同一 target を持つと何も置かずに非ゼロで止まること、後の config の target を外来ファイルで塞ぐと前の config だけ配置・コミットされ非ゼロになることを検証 |
| `06-init-templates` | init + templates。`layat init <t>` で standalone / project テンプレを展開し、展開後 flake が `nix flake check`（layat を局所 override）を通ること |
| `07-legacy`  | legacy entrypoint（shell.nix・passthru canonical 形・→ ADR-0032）。`NIX_PATH` を flake.lock の nixpkgs に pin し、`layat apply` / `apply --all` / 素の `nix-shell` 互換を検証 |
| `08-prune`   | prune（→ ADR-0034）。apply した 2 系列の一方の root を消して孤児にし、`--dryrun` の非破壊性・`--yes` の系列ごと削除・生存系列と配置物の不変・`print-roots` から gcroot が外れることを検証 |
| `09-apply-all` | `apply --all` の並列化（→ ADR-0038 / ADR-0039）。4 つの projectRoot config で `results[]` と `-v` の done 行が辞書順・`--jobs 1` と既定並列度で `results[]` が同一、同じ target を持つ 2 config が `--dryrun` 含め exit 1 で何も置かずに止まり `--json` ではトップ `errors[]` に載る、`--no-wait` 併用の初回 apply で全 config が applied になることを検証 |
| `10-generator-select` | 生成器の選択（→ ADR-0056）。`LAYAT_GENERATOR=bogus` と未知キー入りの `layat.toml` を cwd に置いても `apply --manifest` が配置すること（activation の環境不定に耐える）、`apply --generator bogus` と不正な `layat.toml` での `apply` が exit 1 + 1 行・`--json` で `E_INPUT` になること、`LAYAT_GENERATOR=nix` の `apply` が通常どおり配置することを検証 |
| `11-root-relative` | root 相対 out-of-store（`mkRootRelativeSymlink`）。project mode で apply → readlink が target 起点の相対文字列で内容へ到達できること、再 apply が warning 無しで冪等なこと、リポジトリを移動しても readlink が不変で、移動先での再 apply が foreign の last-wins 置換と記録なしの warning だけで成功すること、`--root` で上書きした root に対して src の実在判定と配置が行われること、entry を外した apply で stale 除去されることを検証 |
| `12-system` | system mode。root を user namespace（`unshare -Ur`、不可なら `sudo -n`）で取り、`systemRoot` の config を `--root` の sysroot へ apply → 隔離した system 基底に `<roothash>/<name>` の profile・backref と記録が立つこと、世代 2 から `rollback` で世代 1 へ戻り `reset --yes` で撤去されること、`--all` が `<roothash>` 系列を除外することを検証。非 root の `apply` / `rollback` / `reset --yes` が `E_PERMISSION` で何も変えずに止まり、`apply --dryrun` / `list-generations` は通ることも見る |

## 将来拡張

- **NixOS VM テスト（`runNixOSTest`）は将来拡張に回す**（→ `docs/design.md`・ADR-0012）。
  NixOS / nix-darwin モジュール経路の実 activate は VM / sandbox を要し、本ハーネス（非 NixOS の
  ubuntu-latest 単一ジョブ）のスコープ外。モジュール経路を本実装する段で別途追加する。

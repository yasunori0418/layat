# layat E2E ハーネス（非 NixOS）

実 nix を使って `layat` を end-to-end に駆動し、**「非 NixOS でも nix さえあれば動く」**という
主張を検証する bash ハーネス（→ `docs/design/README.md`「テスト戦略」・[テストを sandbox 可否で flake check と E2E に分ける](../../docs/adr/test-strategy-layers.md)）。

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
走査する system 基底（→ [profile の基底ディレクトリをモードで切り替える](../../docs/adr/profile-base-directory-per-mode.md), [profile は解決済み root の roothash でキーする](../../docs/adr/profile-keyed-by-roothash-of-resolved-root.md), [prune は system 基底も走査し、権限の無い系列を見送る](../../docs/adr/prune-scans-system-base-skips-unpermitted.md)）は絶対パスで `$HOME` の差し替えでは動かないため、
`e2e_isolate` が `LAYAT_SYSTEM_PROFILE_BASE` で隔離先へ向ける（→ `cmd/layat/prune.go`）。偽 src は fixture flake
ディレクトリ内の相対パス（eval 時に store へコピー）か、out-of-store 用の live ディレクトリで用意する。
fixture flake は `layat` を `path:<repo>` input で参照し、`nixpkgs` / `home-manager` は layat の
`flake.lock` pin に `follows` させてオフライン評価する。

## シナリオ範囲

| シナリオ | テストケース |
|---|---|
| `01-project` | [e2e 01-project — project mode の配置と --all の部分失敗を実 nix で通す](../../docs/test/integration/e2e-01-project.md) |
| `02-home` | [e2e 02-home — 実 nix での home mode 世代コミットと rollback / reset の往復](../../docs/test/generations/e2e-02-home.md) |
| `03-stale` | [tests/e2e/scenarios/03-stale.sh — 実 nix での古くなった配置の除去](../../docs/test/migration-stale/e2e-03-stale.md) |
| `04-copy` | [tests/e2e/scenarios/04-copy.sh — 実 nix での copy place-once と out-of-store](../../docs/test/copy/e2e-04-copy.md) |
| `05-hm` | [e2e 05-hm — 非 NixOS で HM configuration を実際に activate する](../../docs/test/integration/e2e-05-hm.md) |
| `06-init-templates` | [e2e 06-init-templates — テンプレを展開して展開物の評価が通ることを確かめる](../../docs/test/integration/e2e-06-init-templates.md) |
| `07-legacy` | [e2e 07-legacy — 非 flake entrypoint から同じ配置と出力契約を通す](../../docs/test/integration/e2e-07-legacy.md) |
| `08-prune` | [e2e 08-prune — 実 apply が作った系列を root 削除で孤児にし、prune が系列ごと消して gcroot が外れる](../../docs/test/generations/e2e-08-prune.md) |
| `09-apply-all` | [e2e 09-apply-all — apply --all の辞書順集約・並列度非依存・前段衝突 error・--no-wait 併用](../../docs/test/integration/e2e-09-apply-all.md) |
| `10-generator-select` | [e2e 10-generator-select — --manifest を使う場面の環境非依存と入力不正の E_INPUT](../../docs/test/integration/e2e-10-generator-select.md) |
| `11-root-relative` | [e2e 11-root-relative — root 相対 symlink の相対形・移動耐性・古くなった配置の除去を実 nix で通す](../../docs/test/integration/e2e-11-root-relative.md) |
| `12-system` | [e2e 12-system — user namespace の root で system mode の配置・世代・権限の条件を往復する](../../docs/test/generations/e2e-12-system.md) |
| `13-status` | [e2e 13-status — 実 apply が書いた記録を status が列挙し、[name] で絞り、system 基底の記録も拾う](../../docs/test/generations/e2e-13-status.md) |
| `14-flake-parts` | [e2e 14-flake-parts — mkApplyCommand を埋めた devShell が project config だけを配置する](../../docs/test/integration/e2e-14-flake-parts.md) |

## NixOS VM テスト（別系統）

NixOS モジュールで使うときの実 activate は、本ハーネスではなく flake check `checks.nixos-vm`
（`tests/modules/nixos-vm.nix`・`runNixOSTest`）で検証する（→ [NixOS モジュールの実 activate は runNixOSTest の checks.nixos-vm で検証し、E2E ハーネスとは別系統に置く](../../docs/design/nixos-vm-test-position.md)・[テストを sandbox 可否で flake check と E2E に分ける](../../docs/adr/test-strategy-layers.md)）。
実 activate には NixOS 自体が要り、非 NixOS の ubuntu-latest 単一ジョブで回す本ハーネスとは前提が違うため、
別系統の check に置いている。x86_64-linux でだけ VM を起動し、他の system では空の派生に差し替える。

nix-darwin モジュールは VM で回せないため、実 activate のテストを持たない。`checks.darwin-module` は
`postActivation.text` などの評価結果を確かめる check で、activate は実行しない。

---
id: "TP-229b69c0-cf5e-4fb6-a353-27e5064d93e9"
type: test_plan
name: "非 NixOS で動く主張を実 nix の一気通貫 E2E で検証する"
status: stable
derives_from:
  - "SOL-9fcd1d6e-6204-42e6-92bb-1faf966f0b3e"
specification: |
  The claim that layat runs on a non-NixOS system given only nix SHALL be verified against a
  real nix by an E2E harness that drives the actual path end to end, including `nix build`,
  `nix eval` and `nix-env --set` from the flake entrypoint, rather than by stubbing nix out.
  Each scenario SHALL run under an isolated temporary `$HOME` and `$XDG_STATE_HOME`, so that
  it does not touch the environment of the machine that runs it, and its source SHALL be
  supplied either as a relative path inside a fixture flake, which is copied into the store
  at evaluation time, or as a live out-of-store directory. The harness SHALL cover placement
  and idempotency in project mode; placement, generation commit and rollback in home mode;
  stale removal; copy place-once and out-of-store symlinks; evaluating and activating
  the home-manager module on a non-NixOS system; scaffolding the starter templates and
  evaluating what they produce; and the legacy non-flake entrypoint. Each specification it
  verifies is stated by the item responsible for it and SHALL NOT be restated here.
specification_ja: |
  「非 NixOS でも nix さえあれば動く」という主張は、nix をスタブせず、flake entrypoint
  からの `nix build` / `nix eval` / `nix-env --set` を含む実際の処理を一気通貫で回す E2E
  ハーネスにより、実 nix で検証しなければならない。各シナリオは隔離した一時 `$HOME` /
  `$XDG_STATE_HOME` の下で動かさなければならず（実行マシンの環境に触れないようにするため）、
  配置元は fixture flake 内の相対パス（評価時に store へコピーされる）か out-of-store の
  live ディレクトリで用意しなければならない。ハーネスは、project mode の配置と冪等性、home mode の
  配置・世代コミット・ロールバック、古くなった配置の除去、copy の place-once と out-of-store
  symlink、非 NixOS での home-manager モジュールの評価と activate、starter テンプレの
  展開と展開物の評価、および legacy な非 flake entrypoint を対象としなければならない。
  各シナリオが検証する仕様そのものは各担当 item の規範であり、ここでは再掲しない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:42:17+09:00"
---
# TP-229b69c0-cf5e-4fb6-a353-27e5064d93e9: 非 NixOS で動く主張を実 nix の一気通貫 E2E で検証する

## 仕様

「非 NixOS でも nix さえあれば動く」主張を実 nix で検証する E2E ハーネス（`tests/e2e/`・
bash・詳細は `tests/e2e/README.md`）が、flake entrypoint からの `nix build` / `nix eval` /
`nix-env --set` を含む実際の処理を一気通貫で回す。各シナリオは隔離した一時 `$HOME` /
`$XDG_STATE_HOME` 下で動き、偽 src は fixture flake 内の相対パス（eval 時に store へ
コピー）か out-of-store の live ディレクトリで用意する。

| シナリオ | 検証する仕様 |
|---|---|
| `01-project` | project mode。一時 git repo で `layat apply <name>` → git toplevel 配下に store symlink 配置・再 apply の冪等性 |
| `02-home` | home mode。仮 `$HOME` で apply → `$HOME` 配下配置 + profile 世代コミット、entry 入替で世代を進め `layat rollback` で前世代の配置へ復帰 |
| `03-stale` | stale 除去。entry を config から削除 → 再 apply で旧 symlink が消える |
| `04-copy` | copy place-once / out-of-store。copy が通常ファイル（書込可・mode に owner-write 付与）・place-once 冪等（ローカル編集を破棄しない）・out-of-store marker の live symlink |
| `05-hm` | HM module。home-manager standalone configuration を非 NixOS で評価・activate し、activation が `layat apply --manifest` で engine を起動して配置する |
| `06-init-templates` | init + templates。`layat init <t>` で standalone / project テンプレを展開し、展開後 flake が `nix flake check`（layat を局所 override）を通る |
| `07-legacy` | legacy entrypoint（shell.nix・passthru canonical 形）。`NIX_PATH` を flake.lock の nixpkgs に pin し、`layat apply` / `apply --all` / 素の `nix-shell` 互換を検証 |
| `08-prune` | prune。apply した 2 系列の一方の root を消して孤児にし、`--dryrun` の非破壊性・`--yes` の系列ごと削除・残す系列が変わらないこと・`print-roots` から gcroot が外れることを検証 |
| `09-apply-all` | `apply --all` の並列化。`results[]` が辞書順で並列度に依らないこと、同じ target を持つ 2 config が build 前に exit 1 で止まること、`--no-wait` 併用で全 config が applied になることを検証 |
| `10-generator-select` | 生成器の選択。`--manifest` 経路が `LAYAT_GENERATOR` / `layat.toml` に左右されないこと、未知の生成器名・不正な設定ファイルが exit 1 + 1 行・`--json` で `E_INPUT` になること、正しい指定で配置することを検証 |
| `11-root-relative` | root 相対 out-of-store（`mkRootRelativeSymlink`）。target 起点の相対リンク・再 apply の冪等性・`--root` 上書き後の実在判定と配置・リポジトリ移動後の到達性と再 apply・entry 削除時の stale 除去を検証 |
| `12-system` | system mode。root 権限を取って `--root` の sysroot へ apply → 隔離した system 基底に profile・backref・記録が作られ、`rollback` と `reset --yes` が通ること、非 root の変更系が `E_PERMISSION` で何も変えず、読み取り系は通ることを検証 |
| `13-status` | status。home / project の config を 1 つずつ apply し、`layat status` の列挙・`[name]` での絞り込み・`--json` の `info`、隔離した system の記録が `system:<name>` で出ることを検証 |
| `14-flake-parts` | flake-parts + `lib.mkApplyCommand`。devShell の shellHook に埋めた apply が `nix develop -c true` だけで project 側を git toplevel 配下に配置し、home 側は `$HOME` に何も出さないことを検証 |

> **表の `01-project`〜`07-legacy` の行は原文の写しで、規範は frontmatter が正**。その 7 行の右列が指す仕様そのものは、それぞれ
> [project mode の root は git toplevel から解決し、config 相対も CWD 相対も採らない](../requirements/project-mode-root-resolution.md)・[世代は link farm derivation を layat 自前 profile へコミットして積み、前世代 manifest から古くなった配置を除去する](../requirements/generation-mechanism-link-farm-manifest.md) / [rollback は FS を先に収束させてから profile ポインタを最後に移す](../requirements/rollback-refit-then-pointer.md)・[古くなった配置の除去は前世代の記録通りを指す symlink のみに限り、copy は消さず孤児を警告する](../requirements/stale-removal-invariants.md)・[copy は target 不在のときだけマテリアライズする place-once で世代管理の対象外とする](../requirements/copy-place-once.md) / [copy は元の mode を保存しつつ owner-write を付与する](../requirements/copy-mode-owner-write.md) / [out-of-store symlink は marker の絶対パスまたは root 相対パスを指し、版管理はリンク先マッピングのみとする](../requirements/out-of-store-link-mapping-only.md)・[home-manager モジュールの engine kick は activation から全 config のビルド済み link-farm を 1 回の起動で渡し、失敗で switch を止める](../requirements/hm-activation-contract.md)・[template は動く example を 1 config だけ置きバリエーションはコメントで示す](../requirements/init-template-contents.md)・
> [legacy entrypoint は mkShell passthru 形を canonical とし CLI の attr path を分岐させない](../requirements/legacy-entrypoint-passthru-canonical.md) の担当。`tests/e2e/` というパスと bash という実装手段は
> 現況であり、規範文では実装手段を固定していない。

> **2026-08-09 改訂（→ Issue #273）**: シナリオ表を実装の 7 本へ更新し、規範文の対象範囲へ
> `06-init-templates`（テンプレ展開と展開物の評価）と `07-legacy`（legacy 非 flake
> entrypoint）を追加した。表は 5 行のまま据え置かれており、E2E が現に覆っている範囲より
> 狭かった。シナリオ名の列も、実装ファイル（`tests/e2e/scenarios/*.sh`）と比べられる
> よう表記を揃えた。

## 出典

`docs/spec.md`「E2E 検証範囲（非 NixOS）」節の導入文と表。

決定の実体は [テストを sandbox 可否で flake check と E2E に分ける](../adr/test-strategy-layers.md)で、非 NixOS の
実 nix 上で E2E を回す検証範囲を定めている。

> **本 item は `requirement` から `test_plan` へ移設した**（→ Issue #238。旧 ID は
> `REQ-6419e4b0-629c-435f-b1e1-0bc996b5e61d`）。「非 NixOS で動く」主張を何をどこまで E2E で検証するかを定めるテスト
> スコープの規定であり、ユーザーの使われ方（use_case）から導かれるプロダクトの振る舞いでは
> ないため、use_case を親に持てず接続漏れになっていた（当時の判断は Issue #211）。テスト計画
> の型を新設して solution 直下で受けることにしたため、`derives_from` は
> [layat は nix store の物を任意パスへ置く配置プリミティブであり、設定を生成せずユーザーが配置を明示的に握る](../solution/layat-placement-primitive.md) を指す。[エンベロープの outturn 適合を Go テストと E2E の両方で検証する](json-conformance-verification.md)・[NixOS / nix-darwin モジュールで使うときの実 activate は E2E ハーネスの対象外とする](nixos-vm-test-future.md) も同じ経緯で移設した。[lib.__internal は private helper のテスト seam として公開する](lib-internal-test-seam.md) は
> #238 では「テスト計画そのものではない」と判断して見送ったが、Issue #239 でその判断を
> 改めて移設した（経緯は同 item の注記）。

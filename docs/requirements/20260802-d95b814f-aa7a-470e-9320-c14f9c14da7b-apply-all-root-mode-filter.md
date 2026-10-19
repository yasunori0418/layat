---
id: "REQ-d95b814f-aa7a-470e-9320-c14f9c14da7b"
type: requirement
name: "--all は root モードフィルタで対象 config を絞れる"
derives_from:
  - "UC-1c280dce-7c72-44c0-95ea-d06344f62a47"
specification: |
  When `--project-root` / `--home-root` / `--system-root` is given together with
  `apply --all`, only the configs of the corresponding root mode SHALL be applied
  (an opt-in filter named after the root markers). A bare `--all` SHALL apply every
  config. Because a devShell `shellHook` that runs a bare `--all` in an entrypoint mixing
  home mode and project mode configs would also place the home mode configs under `$HOME`,
  a devShell SHOULD use `layat apply --all --project-root` or an explicitly named
  `layat apply <name>`. The filter SHALL be a modifier of `--all` only, since a named apply
  already pins a single config. When `apply --all` runs with an effective uid of 0, without
  any of the three filters, and the selection includes home mode or project mode configs,
  the CLI SHOULD print a single warning line to stderr before applying, naming those configs
  with their root mode and pointing at splitting the run with `--system-root` and
  `--home-root` / `--project-root`; the warning SHALL NOT change the exit code or the
  `--json` envelope.
specification_ja: |
  `apply --all` に `--project-root` / `--home-root` / `--system-root` を付けたときは、
  `layat.*` のうち該当 root モードの config のみを適用しなければならない（root マーカー名に
  揃えたフィルタ・opt-in）。素の `--all` は全 config を適用しなければならない。home mode と
  project mode の config が混在する entrypoint で devShell の `shellHook` から素の `--all` を
  打つと home mode config も `$HOME` に配置される footgun があるため、devShell は
  `layat apply --all --project-root`（または名指し `layat apply <name>`）を使うべきである。
  フィルタは `--all` の修飾に限らなければならない（名指し apply では `<name>` が 1 config を
  pin するため無意味）。実効 uid が 0 の `apply --all` が 3 つのフィルタのいずれも伴わず、
  選択対象に home mode か project mode の config を含むときは、適用に入る前に stderr へ
  warning を 1 行出すべきである。warning は該当 config を root モード付きで挙げ、
  `--system-root` と `--home-root` / `--project-root` で実行を分けるよう案内する。warning は
  exit code と `--json` のエンベロープを変えてはならない。
---
# REQ-d95b814f-aa7a-470e-9320-c14f9c14da7b: --all は root モードフィルタで対象 config を絞れる

## 仕様

`apply --all` に **`--project-root` / `--home-root` / `--system-root`** を付けると、
`layat.*` のうち該当 root モードの config **のみ**を適用する（root マーカー名に揃えた
フィルタ・opt-in）。素の `--all` は全 config を適用する。

home mode と project mode の config が混在する entrypoint で devShell の `shellHook` から
`--all` を打つと home mode config も `$HOME` に配置される footgun があるため、devShell は
**`layat apply --all --project-root`**（または名指し `layat apply <name>`）を使う。
フィルタは `--all` 修飾で、名指し apply では `<name>` が 1 config を pin するため無意味。

root（実効 uid 0）で素の `--all` を打つと、home mode config は root の `$HOME` に root 所有で
置かれ、project mode config も repo 内に root 所有の symlink ができ、どちらも root の state に
世代が作られる。一般ユーザーの後続 apply とは baseline が食い違うため、選択対象に home / project
config が含まれるときは stderr に warning を 1 行出す（error にはしない。root 自身の dotfiles を
home mode で置く用途は正当）。fixed mode と system mode の config は warning の対象外で、
フィルタ付き・非 root のときも出さない。layat は `SUDO_USER` を見ない。

## 出典

`docs/spec.md`「CLI 仕様」→「サブコマンド体系」の `apply --all --project-root` の
箇条書き、およびグローバルフラグ表の `--project-root` / `--home-root` / `--system-root`。

決定の実体は ADR-0017「`--all` の root モードフィルタ」。root 混在時の warning は ADR-0036 の
system mode 実装で足した。

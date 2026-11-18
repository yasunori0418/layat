---
id: "REQ-50af28ba-ddf0-4488-abd7-244f1a5597cb"
type: requirement
name: "--json のエンベロープは暗黙に解決した entrypoint・generator・root を info に記録する"
description: "-f・--generator・--root を省いて解決した entrypoint・generator・root を info に記録し、明示した値・状態の置き場・実行環境の値は記録せず、出すキーが無い実行では info を出さない"
section: cli-json
status: stable
derives_from:
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  When `--json` is given, the CLI SHALL record in `info` only the implicitly resolved
  `entrypoint`, `generator` and `root`, as keys of those names, and SHALL NOT emit the key
  of a value given by an explicit argument. When no key is to be emitted, `info` SHALL be
  absent. The commands that record them SHALL be apply (including `--all` and `--dryrun`),
  reset, rollback, list-generations without `--all`, and gitignore.
  `generator` SHALL be placed in the top-level `info` when `--generator` is omitted, with
  the adopted generator's name, including when it falls back to `nix`; its source SHALL NOT
  be recorded. `entrypoint` SHALL be placed in the top-level `info` when `-f` is omitted and
  discovery finds an entrypoint, with the value `-f` accepts for the same entrypoint: the
  absolute path of the directory holding `flake.nix`, or the absolute path of the legacy
  file. Under `--manifest`, neither `entrypoint` nor `generator` SHALL be emitted. Under
  `--all`, `entrypoint` and `generator` SHALL appear once in the top-level `info`.
  `root` SHALL be placed in each subject's `result.info` when `--root` is omitted and the
  root kind is project or home, with the resolved root; it SHALL NOT be emitted for system
  or fixed, for gitignore, or for a subject that failed before its root was resolved.
  The CLI SHALL NOT record the state dir, the system base, the commands looked up on PATH,
  or init's template ref.
specification_ja: |
  `--json` 指定時、CLI は暗黙に解決した `entrypoint`・`generator`・`root` だけを、その名前の
  キーで `info` に記録しなければならず、明示引数で与えた値のキーを出してはならない。出すキーが
  無いとき、`info` を出してはならない。記録するコマンドは apply（`--all` と `--dryrun` を含む）・
  reset・rollback・`--all` を付けない list-generations・gitignore でなければならない。
  `generator` は、`--generator` を省いた実行で、採用した generator の名前をトップの `info` に
  置かなければならず、既定の `nix` になったときも置かなければならない。由来を記録しては
  ならない。`entrypoint` は、`-f` を省いて探索で見つかった実行で、同じ entrypoint を指す `-f` の
  値をトップの `info` に置かなければならない。値は `flake.nix` のあるディレクトリの絶対パスか、
  legacy のファイルの絶対パスである。`--manifest` の実行では `entrypoint` も `generator` も出しては
  ならない。`--all` では `entrypoint` と `generator` をトップの `info` に 1 度だけ置かなければならない。
  `root` は、`--root` を省いた実行で root の種類が project か home のとき、解決した root を主体ごとの
  `result.info` に置かなければならない。system と fixed、gitignore、root を解決する前に失敗した
  主体では出してはならない。CLI は state dir・system の基底・PATH から探すコマンド・init の
  テンプレートの ref を記録してはならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-18T23:46:43+09:00"
---
# --json のエンベロープは暗黙に解決した entrypoint・generator・root を info に記録する

## 仕様

`--json` のエンベロープは、明示引数を省いて layat が自分で解決した値を `info` に記録する。記録するのは
次の 3 つで、暗黙に解決した値だけをキーとして出す。明示引数で与えた値のキーは出さない。出すキーが
1 つも無い実行では、`info` そのものを出さない。

| キー | 置き場 | 出す条件 | 値 |
|---|---|---|---|
| `generator` | トップの `info` | `--generator` を省いた実行 | 採用した generator の名前（既定の `nix` を含む） |
| `entrypoint` | トップの `info` | `-f` を省き、探索で見つかった実行 | `flake.nix` のあるディレクトリか legacy のファイルの絶対パス |
| `root` | 主体ごとの `result.info` | `--root` を省き、root の種類が project か home の実行 | 解決した root（project は git の toplevel、home は `$HOME`） |

対象のコマンドは apply（`--all` と `--dryrun` を含む）・reset・rollback・`--all` を付けない
list-generations・gitignore である。

- `generator` は generator を選んだ時点で決まるので、その後の探索や評価で失敗した実行にも出る。
  環境変数・設定ファイル・既定値のどれで決まったかは記録しない。
- `entrypoint` の値は、`-f` に渡せば同じ entrypoint になる形である。探索に失敗した実行では出さない。
- `--manifest` の実行は entrypoint を探さず generator も選ばないので、`entrypoint` も `generator` も出さない。
  project の manifest で `--root` を省けば、`root` は出る。
- `--all` では、`entrypoint` と `generator` をトップの `info` に 1 度だけ出し、`root` を主体ごとに出す。
- gitignore は root を解決しないので `root` を出さない。root を解決する前に失敗した主体にも出さない。

次の値は環境から決まるが、記録しない。state dir と system の基底は状態の置き場、PATH から探すコマンドは
実行環境である。init のテンプレートの ref は固定の定数で、入力に数えない。

outturn-validate は、暗黙の入力が記録されているかを確かめない。記録の有無と値は layat のテストで確かめる。

`info` の置き場の全体は [--json は outturn 規約準拠のエンベロープを出す第 2 契約とする](json-outturn-envelope.md)、
変更系の `info` の形は [変更系の JSON ペイロードは engine 結果からフルインベントリと実差分を導く](json-change-payload.md)、
読み取り系の `info` の形は [読み取り系の JSON ペイロードは dryRun パリティと info インベントリで表す](json-readonly-payload.md) の担当。

## 出典

outturn の ADR-0028 で、暗黙の入力を使った実行は、その解決結果を `info` に記録することが MUST になった。

決定の実体は [暗黙の入力は配置を決める値だけを、暗黙に解決したときだけ info に記録する](../adr/implicit-inputs-recorded-in-info.md)。

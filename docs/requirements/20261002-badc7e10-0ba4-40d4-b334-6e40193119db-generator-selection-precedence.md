---
id: "REQ-badc7e10-0ba4-40d4-b334-6e40193119db"
type: requirement
name: "生成器は明示指定のみで選び、フラグ・環境変数・設定ファイル・既定の順に解決する"
derives_from:
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  The CLI SHALL select the manifest generator only from an explicit choice, and SHALL NOT
  infer it from which entrypoint files exist. The choice SHALL be resolved in this order,
  the first one present winning: the `--generator` flag, the `LAYAT_GENERATOR` environment
  variable, the project configuration `layat.toml`, the user configuration
  `$XDG_CONFIG_HOME/layat/config.toml` (`~/.config/layat/config.toml` when
  `XDG_CONFIG_HOME` is unset; skipped when neither `XDG_CONFIG_HOME` nor the home
  directory can be resolved), and the default `nix`. An empty `LAYAT_GENERATOR` and a
  configuration file without the `generator` key SHALL count as not given, and resolution
  SHALL continue to the next source. The only accepted value SHALL be
  `nix`; the prebuilt generator SHALL be selected by `--manifest` alone. When `--manifest`
  is given, the CLI SHALL NOT read the environment variable or any configuration file.
  The project configuration SHALL be looked up only in the directory given by `-f` (the
  directory containing it when `-f` names a file), or in the current directory when `-f`
  is absent, without searching parent directories. When the `-f` path cannot be stat'ed,
  the CLI SHALL NOT read either configuration file, and SHALL leave the failure to
  entrypoint discovery. The
  configuration files SHALL be TOML whose only key is `generator`, and an unknown key SHALL
  be an error. A configuration file that exists but cannot be read SHALL NOT be reported
  as `E_INPUT` and SHALL keep the classification of its read failure (`E_PERMISSION`,
  `E_IO`). An unknown generator name, an invalid configuration file, and the flag
  combinations `--generator` with `--manifest`, `-f` with `--manifest` and `--all` with
  `--manifest` SHALL be reported as `E_INPUT` under `--json`. `prune` and `init` SHALL
  ignore `--generator` without validating its value, and SHALL NOT read the environment
  variable or any configuration file.
specification_ja: |
  CLI は manifest 生成器を明示指定のみで選ばなければならず、どの entrypoint ファイルが
  あるかから推定してはならない。指定は次の順に解決し、最初に存在したものを採らなければ
  ならない: `--generator` フラグ、環境変数 `LAYAT_GENERATOR`、プロジェクト設定
  `layat.toml`、ユーザー設定 `$XDG_CONFIG_HOME/layat/config.toml`（`XDG_CONFIG_HOME`
  未設定時は `~/.config/layat/config.toml`。`XDG_CONFIG_HOME` もホームディレクトリも解決
  できないときはこの段を飛ばす）、既定 `nix`。空文字の `LAYAT_GENERATOR` と
  `generator` キーを持たない設定ファイルは指定なしとして次の段へ進まなければならない。
  受け付ける値は `nix` のみと
  しなければならず、prebuilt 生成器は `--manifest` だけで選ばなければならない。
  `--manifest` 指定時は環境変数・設定ファイルを読んではならない。プロジェクト設定は
  `-f` のディレクトリ（`-f` がファイルならそのファイルのあるディレクトリ）、無ければ cwd
  だけを探さなければならず、上方向に探索してはならない。`-f` のパスを stat できないときは
  設定ファイルをどちらも読んではならず、失敗は entrypoint の発見に委ねなければならない。
  設定ファイルは項目が `generator` だけの TOML とし、未知キーはエラーにしなければならない。
  存在するが読めない設定ファイルは `E_INPUT` として報告してはならず、読み込み失敗の分類
  （`E_PERMISSION` / `E_IO`）を保たなければならない。
  未知の生成器名・設定ファイルの不正・フラグの組み合わせ（`--generator` + `--manifest`、
  `-f` + `--manifest`、`--all` + `--manifest`）は `--json` で `E_INPUT` として報告しなければ
  ならない。`prune` と `init` は `--generator` の値を検証せずに無視し、環境変数・設定
  ファイルを読んではならない。
---
# REQ-badc7e10-0ba4-40d4-b334-6e40193119db: 生成器は明示指定のみで選び、フラグ・環境変数・設定ファイル・既定の順に解決する

## 仕様

| 優先 | 指定元 | 備考 |
|---|---|---|
| 1 | `--generator <name>` | persistent flag。prune / init は値を検証せず無視し、2〜4 も読まない |
| 2 | `LAYAT_GENERATOR` | 空文字は指定なし |
| 3 | `layat.toml` | `-f` のディレクトリ（ファイルならその親）、無ければ cwd。上方向探索なし。`-f` を stat できなければ 3・4 を読まない |
| 4 | `$XDG_CONFIG_HOME/layat/config.toml` | 未設定時は `~/.config/layat/config.toml`。どちらも解決できなければ飛ばす |
| 5 | 既定 `nix` | |

```toml
generator = "nix"
```

設定ファイルに `generator` キーが無ければ指定なしとして次の段へ進む。

- 値の enum は `nix` のみ。prebuilt は `--manifest` でだけ選ばれ、`--manifest` 時は
  2〜4 を読まない（モジュール activation が実行環境の環境変数・cwd に左右されないため）。
- 入力不正（未知値・TOML のパース失敗・未知キー・フラグの組み合わせ不正）は人間向けは
  exit 1 + 1 行、`--json` では共通コード `E_INPUT`。`-f` / `--all` + `--manifest` の既存の
  エラーは文面を変えずにコードだけ `E_INPUT` へ揃える。
- 存在するが読めない設定ファイル（権限なし・ディレクトリ）は入力不正にせず、読み込み失敗の
  分類（`E_PERMISSION` / `E_IO`）を保つ。
- 設定ファイルは生成器の選択にだけ使い、config の発見には使わない。

`--manifest` の外面と `-f` / `--all` との排他そのものは REQ-dec58330-6dad-47f7-8f56-2402764a89c7、TOML パーサの
依存許可は REQ-637599dc-a1ec-4af5-9e97-e882c7df56d0 の担当。

## 出典

ADR-0056「生成器の選択と設定ファイル」§1〜§6。

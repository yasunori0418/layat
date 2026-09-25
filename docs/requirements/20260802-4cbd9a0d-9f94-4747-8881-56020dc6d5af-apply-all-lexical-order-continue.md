---
id: "REQ-4cbd9a0d-9f94-4747-8881-56020dc6d5af"
type: requirement
name: "apply --all は選択された config を並列に適用し、集約表示と results[] を辞書順にし、部分失敗でも続行して最後に集約する"
derives_from:
  - "UC-1c280dce-7c72-44c0-95ea-d06344f62a47"
specification: |
  `layat apply --all` SHALL apply the selected configs of `layat.*` in parallel, their
  execution order being non-deterministic, and SHALL present the aggregated display and
  `results[]` in lexical (key-sorted, deterministic) order regardless of the order in which
  the configs complete. It SHALL continue with the remaining configs even when some of them
  fail, because each config is an independent atomic profile. It SHALL display an aggregated
  success / failure summary at the end, and SHALL exit non-zero when at least one config
  failed. `--all` itself SHALL NOT be atomic as a whole, because project mode does not
  expose rollback and the semantics would break.
specification_ja: |
  `layat apply --all` は選択された `layat.*` の config を並列に適用しなければならず（実行順は
  非決定）、集約表示と `results[]` は config の完了順に依らず辞書順（キーソート・決定的）に
  しなければならない。一部が失敗しても残りを続行しなければならない（各 config は独立 profile
  で atomic なため）。最後に成功 / 失敗を集約表示し、1 つでも失敗なら非ゼロ終了しなければ
  ならない。`--all` 自体は全体 atomic にしてはならない（project mode は rollback 非公開で
  意味論が崩れるため）。
---
# REQ-4cbd9a0d-9f94-4747-8881-56020dc6d5af: apply --all は選択された config を並列に適用し、集約表示と results[] を辞書順にし、部分失敗でも続行して最後に集約する

## 仕様

`apply --all` は **選択された config を並列に適用**し（同時実行数は `--jobs` の解決値。
実行順は非決定）、**一部が失敗しても残りを続行**する（各 config は独立 profile で atomic
なため）。各 config が独立 atomic なので実行順は結果に影響せず、順序が意味を持つのは
表示と集約だけである。そこで config ごとの出力（`-v` レポート・`--dryrun` の計画・失敗 /
skip の行）と `--json` の `results[]` は、**全 config の完了後に辞書順で**まとめて出す。
最後に成功 / 失敗を集約表示し、**1 つでも失敗なら非ゼロ終了**する。`--all` 自体は全体
atomic にしない（project mode は rollback 非公開で意味論が崩れるため）。

engine が配置中に出す warning（foreign symlink 等）と、ロック内 build が流す nix の出力・
`--debug` の開示行は辞書順集約の対象外で、発生順（非決定）に stderr へ流れる。

## 改訂の経緯

ADR-0039（`apply --all` の build・配置の両段階の並列化）による改訂。旧規範は
「`layat.*` を辞書順（キーソート・決定的）に**逐次**適用する」（ADR-0016「`--all` 適用順」）
で、適用順そのものを辞書順に固定していた。ADR-0038 の前段衝突検査で同一 entrypoint 内の
config 間衝突が排除され、実行順が結果に影響しなくなったため、辞書順の保証を実行順から
表示・集約の順へ移した。

## 出典

`docs/spec.md`「CLI 仕様」→「サブコマンド体系」の `apply --all` の箇条書き。

決定の実体は ADR-0039（並列適用・完了後の辞書順集約）、ADR-0016「`--all` 適用順」
（改訂前の辞書順・決定的）と ADR-0013（`--all` を全体 atomic にしない・部分失敗でも続行し
集約表示）。

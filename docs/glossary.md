# layat 用語索引

layat のユビキタス言語の索引。用語の正名・避ける綴り・定義は `docs/ubiquitous-language/` の
`ubiquitous_term` item が持ち、本文書は主題ごとの節に item へのリンクを並べるだけにする。

> **この文書の書き方（規約）**
>
> - 正名を決めるのは `ubiquitous_term` の item。英語の正名（`name`）を一次とし、日本語の
>   正名は `name_ja` に置く。英語のまま使う語は `name_ja` を持たない
> - 執筆では item の `avoid` にある綴りを使わない
> - `avoid` の綴りは契約テスト `dev/tests/term-avoid.sh`（textlint + prh）が検出する。対象は docs・README・CLAUDE.md で、ADR・SDR・用語 item は除く
> - 使いたい語の item が無ければ `sara-new ubiquitous_term <slug> docs/ubiquitous-language -- --description '<1 文の要約>'` で
>   起こし、本文書の該当する節に 1 行足す
> - 本文書には定義を書かない。1 語 1 行で `` - [<item の name>](<item>) — `<英語の正名>`（<日本語の正名>） `` の形に揃える

## Core placement abstraction

- [placement primitive](ubiquitous-language/placement-primitive.md) — `placement primitive`（配置プリミティブ）
- [engine](ubiquitous-language/engine.md) — `engine`
- [layat](ubiquitous-language/layat.md) — `layat`
- [layat CLI](ubiquitous-language/layat-cli.md) — `layat CLI`
- [generator](ubiquitous-language/generator.md) — `generator`（生成器）
- [prebuilt](ubiquitous-language/prebuilt.md) — `prebuilt`
- [generation backend](ubiquitous-language/generation-backend.md) — `generation backend`（世代バックエンド）
- [entrypoint](ubiquitous-language/entrypoint.md) — `entrypoint`
- [starter](ubiquitous-language/starter.md) — `starter`
- [module](ubiquitous-language/module.md) — `module`（モジュール）
- [`standalone`](ubiquitous-language/standalone.md) — `standalone`

## Placement input and output

- [entry / entries](ubiquitous-language/entry.md) — `entry / entries`
- [src](ubiquitous-language/src.md) — `src`
- [subpath](ubiquitous-language/subpath.md) — `subpath`
- [target](ubiquitous-language/target.md) — `target`
- [root](ubiquitous-language/root.md) — `root`

## Placement modes

- [home mode](ubiquitous-language/home-mode.md) — `home mode`
- [project mode](ubiquitous-language/project-mode.md) — `project mode`
- [system mode](ubiquitous-language/system-mode.md) — `system mode`
- [projectRoot](ubiquitous-language/project-root.md) — `projectRoot`
- [homeRoot](ubiquitous-language/home-root.md) — `homeRoot`
- [systemRoot](ubiquitous-language/system-root.md) — `systemRoot`
- [ephemeral placement](ubiquitous-language/ephemeral-placement.md) — `ephemeral placement`

## Placement kinds

- [store link](ubiquitous-language/store-link.md) — `store link`
- [out-of-store symlink](ubiquitous-language/out-of-store-symlink.md) — `out-of-store symlink`

## State management

- [generation](ubiquitous-language/generation.md) — `generation`（世代）
- [store manifest](ubiquitous-language/store-manifest.md) — `store manifest`（store マニフェスト）
- [last-applied record](ubiquitous-language/last-applied-record.md) — `last-applied record`（last-applied 記録）
- [method](ubiquitous-language/method.md) — `method`

## Error handling

- [error-wrap convention](ubiquitous-language/error-wrap-convention.md) — `error-wrap convention`

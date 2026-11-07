# layat 用語索引

layat のユビキタス言語の索引。用語の正名・避ける綴り・定義は `docs/ubiquitous-language/` の
`ubiquitous_term` item が持ち、本文書は主題ごとの節に item へのリンクを並べるだけにする。

> **この文書の書き方（規約）**
>
> - 正名を決めるのは `ubiquitous_term` の item。英語の正名（`name`）を一次とし、日本語の
>   正名は `name_ja` に置く。英語のまま使う語は `name_ja` を持たない
> - 執筆では item の `avoid` にある綴りを使わない
> - 使いたい語の item が無ければ `sara-new ubiquitous_term <slug> docs/ubiquitous-language` で
>   起こし、本文書の該当する節に 1 行足す
> - 本文書には定義を書かない。1 語 1 行で `` - [<ID>](<item>) — `<英語の正名>`（<日本語の正名>） `` の形に揃える

## Core placement abstraction

- [UBQ-3046d454-1ff1-48e2-9b05-0aad11c52ea6](ubiquitous-language/20261107-3046d454-1ff1-48e2-9b05-0aad11c52ea6-placement-primitive.md) — `placement primitive`（配置プリミティブ）
- [UBQ-4ddb0597-ec0e-47af-8988-6636b2e5aa29](ubiquitous-language/20261107-4ddb0597-ec0e-47af-8988-6636b2e5aa29-engine.md) — `engine`
- [UBQ-c3e66a6c-906d-4083-b7ba-19618539c56c](ubiquitous-language/20261107-c3e66a6c-906d-4083-b7ba-19618539c56c-layat.md) — `layat`
- [UBQ-285b8069-6e9d-4845-a9b5-634e91338bfe](ubiquitous-language/20261107-285b8069-6e9d-4845-a9b5-634e91338bfe-layat-cli.md) — `layat CLI`
- [UBQ-b3174f13-88f4-4163-8f56-33d41e21cf77](ubiquitous-language/20261107-b3174f13-88f4-4163-8f56-33d41e21cf77-generator.md) — `generator`（生成器）
- [UBQ-fbb72ad3-f0d8-4403-8775-705fb0275ade](ubiquitous-language/20261107-fbb72ad3-f0d8-4403-8775-705fb0275ade-prebuilt.md) — `prebuilt`
- [UBQ-1660e016-7a0a-4319-9346-466f759fe759](ubiquitous-language/20261107-1660e016-7a0a-4319-9346-466f759fe759-generation-backend.md) — `generation backend`（世代バックエンド）
- [UBQ-ba99246c-fd5b-4469-bb30-afcf33e83699](ubiquitous-language/20261107-ba99246c-fd5b-4469-bb30-afcf33e83699-entrypoint.md) — `entrypoint`
- [UBQ-3f190b7f-08dc-4b64-ab48-4be319a7d48f](ubiquitous-language/20261107-3f190b7f-08dc-4b64-ab48-4be319a7d48f-starter.md) — `starter`
- [UBQ-57c32ea5-8b3d-4b5d-b1f8-daee58ecbe40](ubiquitous-language/20261107-57c32ea5-8b3d-4b5d-b1f8-daee58ecbe40-module.md) — `module`（モジュール）
- [UBQ-f7799808-267e-40c4-aaa7-d85e568f11c4](ubiquitous-language/20261107-f7799808-267e-40c4-aaa7-d85e568f11c4-standalone.md) — `standalone`

## Placement input and output

- [UBQ-5caf1573-b91f-40dd-a414-58b5f49bf0d9](ubiquitous-language/20261107-5caf1573-b91f-40dd-a414-58b5f49bf0d9-entry.md) — `entry / entries`
- [UBQ-6e764ddd-03ce-4bd1-be53-64d651783241](ubiquitous-language/20261107-6e764ddd-03ce-4bd1-be53-64d651783241-src.md) — `src`
- [UBQ-12218337-953b-40c8-b9ec-6d8ee8ec0d4e](ubiquitous-language/20261107-12218337-953b-40c8-b9ec-6d8ee8ec0d4e-subpath.md) — `subpath`
- [UBQ-575f2527-64a9-4f8b-ae68-1131b053cf66](ubiquitous-language/20261107-575f2527-64a9-4f8b-ae68-1131b053cf66-target.md) — `target`
- [UBQ-e6466d28-6ab6-43be-a370-cca82eb1cfdb](ubiquitous-language/20261107-e6466d28-6ab6-43be-a370-cca82eb1cfdb-root.md) — `root`

## Placement modes

- [UBQ-229fb5db-7865-4679-917c-2c067ea571c9](ubiquitous-language/20261107-229fb5db-7865-4679-917c-2c067ea571c9-home-mode.md) — `home mode`
- [UBQ-dbbd91ee-1d41-438b-9c6f-6972643741cb](ubiquitous-language/20261107-dbbd91ee-1d41-438b-9c6f-6972643741cb-project-mode.md) — `project mode`
- [UBQ-4cf59313-b234-4534-95c6-d8dd3dbd99a2](ubiquitous-language/20261107-4cf59313-b234-4534-95c6-d8dd3dbd99a2-system-mode.md) — `system mode`
- [UBQ-d0d09a24-f641-447d-b490-d33ca57fe862](ubiquitous-language/20261107-d0d09a24-f641-447d-b490-d33ca57fe862-project-root.md) — `projectRoot`
- [UBQ-55db6ca4-63de-437d-a667-a0e71676e9f1](ubiquitous-language/20261107-55db6ca4-63de-437d-a667-a0e71676e9f1-home-root.md) — `homeRoot`
- [UBQ-9425b2c7-3f62-4994-9269-cd79dd1261f7](ubiquitous-language/20261107-9425b2c7-3f62-4994-9269-cd79dd1261f7-system-root.md) — `systemRoot`
- [UBQ-30bf2772-cc17-483e-9b0c-79bbeab2d0ed](ubiquitous-language/20261107-30bf2772-cc17-483e-9b0c-79bbeab2d0ed-ephemeral-placement.md) — `ephemeral placement`

## Placement kinds

- [UBQ-d0332595-e976-4942-b218-728e428ecc26](ubiquitous-language/20261107-d0332595-e976-4942-b218-728e428ecc26-store-link.md) — `store link`
- [UBQ-2340b0fd-6444-4946-94a6-43f874e37615](ubiquitous-language/20261107-2340b0fd-6444-4946-94a6-43f874e37615-out-of-store-symlink.md) — `out-of-store symlink`

## State management

- [UBQ-10bcf077-1e99-456b-bb0a-c9b87daf526e](ubiquitous-language/20261107-10bcf077-1e99-456b-bb0a-c9b87daf526e-generation.md) — `generation`（世代）
- [UBQ-836edea0-547e-4b96-a1b3-783c6ca86bde](ubiquitous-language/20261107-836edea0-547e-4b96-a1b3-783c6ca86bde-store-manifest.md) — `store manifest`（store マニフェスト）
- [UBQ-7e8e239e-e11b-414e-b6e7-295520d6a1a9](ubiquitous-language/20261107-7e8e239e-e11b-414e-b6e7-295520d6a1a9-last-applied-record.md) — `last-applied record`（last-applied 記録）
- [UBQ-bc180161-69ea-4ec6-bae5-9c0e5853a112](ubiquitous-language/20261107-bc180161-69ea-4ec6-bae5-9c0e5853a112-method.md) — `method`

## Error handling

- [UBQ-f5a31977-edde-4c0f-881f-28f709b0ef0b](ubiquitous-language/20261107-f5a31977-edde-4c0f-881f-28f709b0ef0b-error-wrap-convention.md) — `error-wrap convention`（エラー wrap 規約）

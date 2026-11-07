# layat の ADR

ADR の粒度・ステータス・`supersedes` と `deprecated`・`justifies` の選定基準といった汎用の規則は
[adr-rules.md](../agents/adr-rules.md) が持つ。この文書はそれを layat へ当てはめた部分
（型とディレクトリ・起票の手順・frontmatter の対応表・検算用の実データ）だけを持つ。

## 型とディレクトリ

| 型 | prefix | ID | ファイル | 置き場所 |
|---|---|---|---|---|
| `adr` | `ADR` | `ADR-<フル UUIDv4>` | `<YYYYMMDD>-<フル UUIDv4>-<slug>.md` | `docs/adr/` |
| `sara_adr` | `SDR` | `SDR-<フル UUIDv4>` | `<YYYYMMDD>-<フル UUIDv4>-<slug>.md` | `docs/sara-adr/` |

- `adr` は layat の設計判断を記録する
- `sara_adr` は sara による品質管理機構（`docs/model.yaml`・グラフの検証・起票ラッパー・item の運用規約）
  の設計判断を記録する。layat 本体とは別に切り出せる単位として分けている
- 連番で採番していた旧 ADR（`ADR-NNNN`）は削除済み。旧番号は新 ADR の `origin` と「起点:」行に来歴として残る

## 起票

`adr` と `sara_adr` は他の item と同じく `sara-new` で起票する。採番とファイル名の規約の適用が
一手で済む。

```bash
nix develop '.?dir=dev#sara' -c sara-new adr <slug> docs/adr -- \
  --name '<タイトル>' --description '<1 文の要約>' --status stable \
  --origin '…' \
  --justifies '<ID>' [--supersedes '<ADR>'] [--references '<ADR>']
```

`sara_adr` は型を `sara_adr`、配置ディレクトリを `docs/sara-adr` にする。本文のタイトル行は
`# ADR-<フル UUIDv4>: <タイトル>`（`sara_adr` は `# SDR-…`）とし、採番結果の ID で書く。

`--name` は必須フィールドなので省略しない。`sources` は `sara-new` のフラグで渡せないので、
起票後に frontmatter へ手で書く。

## frontmatter

各 ADR は本文のメタデータ行に加えて、機械検証用の YAML frontmatter を持つ。本文のメタデータ行が
人の読む一次情報で、frontmatter はその二重化。**両方を同じ変更の中で更新する**。

```yaml
---
id: "ADR-<フル UUIDv4>"
type: adr
name: "<タイトル行から `# ADR-…: ` を除いた部分>"
status: stable          # draft / stable / deprecated
sources:                # 「関連:」「起点 Issue:」「起点:」の GitHub Issue
  - id: "issue-123"
    resource: "https://github.com/yasunori0418/layat/issues/123"
    title: "#123"
origin: "…"             # 「起点:」の内容（無ければ「起点 Issue:」「参照:」）
justifies:              # この ADR が決めた item（必須・1 本以上）
  - "REQ-<フル UUIDv4>"
supersedes:             # 置き換えた ADR（置き換えた側を deprecated にする変更の中でだけ張る）
  - "ADR-<フル UUIDv4>"
references:             # 「関連:」の ADR
  - "ADR-<フル UUIDv4>"
---
```

フィールドの順序は上記のとおり（`id` / `type` / `name` / `status` / `sources` / `origin` /
`justifies` / `supersedes` / `references`）。

`supersedes` だけは本文のメタデータ行を持たず、frontmatter にだけ書く。読者への誘導は、置き換えられた
側の冒頭に置く後継一覧のバナーが担う。

本文のメタデータ行との対応:

| 本文 | frontmatter |
|---|---|
| `- ステータス:` | `status`。本文にも frontmatter と同じ英語の値を書き、カッコ書きは付けない |
| `- 日付:` | 持たない（機械検証に使わないため本文のみ）|
| `- 関連:` の ADR 参照 | `references` |
| `- 関連:` の GitHub Issue 参照 | `sources` |
| `- 関連:` の `docs/*.md` 参照 | 持たない（本文に残す）|
| `- 起点:` | `origin`。行が `Issue #N` で始まるときは `sources` にも入れる |
| `- 起点 Issue:` / `- 参照:` | `起点:` 行が無いときだけ `origin` へ入れる |

分類の細則:

- **他リポジトリの ADR は `references` に入れない**。`関連:` にリポジトリ名を前置した ADR 参照
  （例: ADR-79af04fe-7ba0-4953-bb0d-75bda1707cd7 の `関連:` にある outturn の ADR）は ID 空間が別なので、`references` に入れると自リポジトリの
  同番号 ADR への誤った辺になる（`sara check` は実在する ID なので通してしまう）。本文に残す
- **他リポジトリの Issue も `sources` へ入れる**（`関連:` / `起点:` / `起点 Issue:` のどの行にあっても）。
  `resource` を `https://github.com/owner/repo/issues/N`、`id` を `issue-<repo>-N` にすれば ID 空間が衝突しない
  （例: ADR-79af04fe-7ba0-4953-bb0d-75bda1707cd7 の `yasunori0418/outturn#1`）
- `起点:` 行の Issue は原則 `origin` のテキストに残す。行が `Issue #N` で始まる（その ADR の直接の
  起点である）ときだけ `sources` にも入れる（例: ADR-955c2601-88fb-4be8-85de-ec7d0a8c727d の `- 起点: Issue #4（…）`）。経緯として
  言及されるだけの Issue は入れない（例: ADR-56325270-20bb-4959-8498-e3223c35e24b の「旧 epic #167 から統合された」の `#167`）
- `origin` のフォールバックは `起点:` → `起点 Issue:` → `参照:` の順。フォールバック時は `"起点 Issue: #26"`
  `"参照: PR #40, …"` のようにラベルを残す
- 同じ ADR を `supersedes` と `references` の両方に入れない。
  前者を優先する（同一ペアに relation を二重に張らない）
- `docs/*.md` 参照は、その ADR が参照する記述が全て item へ分割され終わってから除く（→ #213）。
  `docs/concept.md` を指す参照は、「設計の哲学」など item を立てない節を指しうるので保留する

### `references` の向き

`references` は**後発 → 先行の片方向**に張る。同一ペアに相互に張ると `sara check` が cycle として
error にする。先行 ADR 側からは `sara query` の `Is referenced by` で逆引きできる。新しく起こす ADR は
常に最後発なので、`関連:` の ADR をそのまま `references` に入れてよい。

## `justifies` の実例

選定基準（寄与の有無・張らない 2 条件・監査の 2 方向）は [adr-rules.md](../agents/adr-rules.md) が
持つ。layat での判定の実例:

- 寄与した ADR が複数張る: `REQ-46fccb80-4bae-4d37-bc19-dded88e9a9c0` は ADR-a6a1ef1b-676c-484a-9c8a-3e6d49c1b444 /
  ADR-ad1c2507-1f42-4cd3-ab56-9183892f8988 / ADR-55d8ecdf-86bb-45d0-8c60-ac9c630cb62d /
  ADR-5265f879-e623-44f9-9c58-8c14598e8b10 / ADR-15a358f4-1c81-49d5-b658-3a1fe39f0f75 の 5 本が張る。
  `REQ-a5053191-1c6a-449b-9c5e-5ff49dc5aead` は帰属の ADR-6255a1ed-8df4-4230-b77d-cb9303a03e71 に加え、`--json` を導入した
  ADR-0346fdd2-bd05-4793-bcd0-9c3ce64c8657 と `tool.version` の供給元である ADR-9ab223eb-99dc-4aa2-a671-c07f75f7ff3c が張る
- item が `justifies` を受けない理由を明記している: `DSG-d2e17f4f-0d32-45c1-8125-17e589664c85` の
  「根拠が ADR 以外にあることによる」。一方 `REQ-3f541d39-da41-4ef8-858b-707f54cf6a29` の「`subdir`
  適用形は担当範囲外」は規範に含めないという宣言で、ADR に依拠していないことではない
- 背景で前提として言及するだけ: ADR-c109ac0c-d5df-4a15-acf0-d6d85ce5402e → `REQ-535b811d-dfc5-4eac-92db-737e70eb5415` の一括 eval
- 出典節が ADR を名指ししない item: `REQ-4fc98fa6-fd5f-4f25-8b10-60755bf49bd2` は ADR-cbac86c1-d774-4cf3-abe6-ca8a7c2f1995 の決定から見つけた
- 規範文が対称化の対象として直に述べるので張る: `REQ-07c3b735-3744-4778-a640-8c6fb66f4aa7` は symlink 側の
  foreign 警告を `REQ-622787dc-4512-4ce9-9c7d-7b32bbb70557` へ委譲しつつ、規範文が "symmetrically with the
  foreign warning for symlinks" と名指しするので ADR-095220c9-5a3d-4fd4-94bd-8fa1a02dd36d を張る
- 決定節が実際に決めているので張る: ADR-74423bb8-1453-4689-a25e-30bf46536afc →
  `REQ-61856da1-8883-401e-ad57-9f326b96d400` の「冪等再実行による収束の前提」

`justifies` を受けない理由を散文で書いている item は次で列挙できる。

```bash
rg -n 'justifies' docs/requirements/ docs/design/ docs/infrastructure/ docs/test-plan/
```

## セルフチェック

ADR を書き終えたら、グラフ全体の検証と、`supersedes` と `deprecated` の対応の検証を実行する。

```bash
nix develop '.?dir=dev#sara' -c sara check
nix develop '.?dir=dev#sara' -c dev/tests/adr-supersede.sh
```

`sara check` の合格条件は **warning / error とも 0 件で exit 0 になること**。`sara.toml` は
`strict_mode = true`（→ SDR-bf25a81b-d588-4c7e-9365-fe514ee4d626）なので、`Orphan item`（接続漏れ）も `Broken reference`・
`Duplicate identifier`・`Circular reference detected` と同様に error になる。ADR で接続漏れが出るのは
`justifies` の埋め忘れなので、張ってから完了とする。`justifies` に書いた ID のタイポは接続漏れでは
なく `Broken reference` として落ちる。

`adr-supersede.sh` は、`supersedes` で指された ADR が `deprecated` であることを確かめる。どちらも CI の `sara` job が実行するが、required status
check ではないため、赤くなってもマージはブロックしない（→ SDR-6cc54665-c239-4235-92a8-f73e26c5a5ef）。

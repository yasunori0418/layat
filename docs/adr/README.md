# layat の ADR

ADR の粒度・ステータス・`supersedes` と `廃止`・`justifies` の選定基準といった汎用の規則は
[adr-rules.md](../agents/adr-rules.md) が持つ。この文書はそれを layat へ当てはめた部分
（型とディレクトリ・起票の手順・frontmatter の対応表・検算用の実データ）だけを持つ。

## 型とディレクトリ

| 型 | prefix | ID | ファイル | 置き場所 |
|---|---|---|---|---|
| `adr` | `ADR` | `ADR-<フル UUIDv4>` | `<YYYYMMDD>-<フル UUIDv4>-<slug>.md` | `docs/adr/` |
| `sara_adr` | `SDR` | `SDR-<フル UUIDv4>` | `<YYYYMMDD>-<フル UUIDv4>-<slug>.md` | `docs/sara-adr/` |
| `legacy_adr` | `ADR` | `ADR-NNNN`（連番）| `NNNN-<slug>.md` | `docs/adr/` |

- `adr` は layat の設計判断を記録する
- `sara_adr` は sara による品質管理機構（`docs/model.yaml`・グラフの検証・起票ラッパー・item の運用規約）
  の設計判断を記録する。layat 本体とは別に切り出せる単位として分けている
- `legacy_adr` は連番で採番していた既存の ADR で、凍結している（→「`legacy_adr` の扱い」）

## 起票

`adr` と `sara_adr` は他の item と同じく `sara-new` で起票する。採番とファイル名の規約の適用が
一手で済む。

```bash
nix develop '.?dir=dev#sara' -c sara-new adr <slug> docs/adr -- \
  --name '<タイトル>' --status 採用 \
  --issues '#<sub-issue>' --origin '…' \
  --justifies '<ID>' [--supersedes '<ADR>'] [--references '<ADR>']
```

`sara_adr` は型を `sara_adr`、配置ディレクトリを `docs/sara-adr` にする。本文のタイトル行は
`# ADR-<フル UUIDv4>: <タイトル>`（`sara_adr` は `# SDR-…`）とし、採番結果の ID で書く。

`--name` は必須フィールドなので省略しない。`--issues` は `!list text` なので、複数の issue を
持つときはフラグを繰り返す。`legacy_adr` は `sara-new` が拒否する（新規には起こさない）。

## frontmatter

各 ADR は本文のメタデータ行に加えて、機械検証用の YAML frontmatter を持つ。本文のメタデータ行が
人の読む一次情報で、frontmatter はその二重化。**両方を同じ変更の中で更新する**。

```yaml
---
id: "ADR-<フル UUIDv4>"
type: adr
name: "<タイトル行から `# ADR-…: ` を除いた部分>"
status: 採用            # 提案 / 採用 / 不採用 / 廃止
issues:                 # 「関連:」「起点 Issue:」「起点:」の GitHub Issue 番号
  - "#123"
origin: "…"             # 「起点:」の内容（無ければ「起点 Issue:」「参照:」）
justifies:              # この ADR が決めた item（必須・1 本以上）
  - "REQ-8521e8f8-99cc-47f0-b7d0-70c24c837612"
supersedes:             # 置き換えた ADR（置き換えた側を 廃止 にする変更の中でだけ張る）
  - "ADR-0020"
references:             # 「関連:」の ADR
  - "ADR-0013"
---
```

フィールドの順序は上記のとおり（`id` / `type` / `name` / `status` / `issues` / `origin` /
`justifies` / `supersedes` / `references`）。`legacy_adr` は `supersedes` の位置に `revises` を持つ。

`supersedes` だけは本文のメタデータ行を持たず、frontmatter にだけ書く。読者への誘導は、置き換えられた
側の冒頭に置く後継一覧のバナーが担う。

本文のメタデータ行との対応:

| 本文 | frontmatter |
|---|---|
| `- ステータス:` | `status`（カッコ書きは付けない）|
| `- 日付:` | 持たない（機械検証に使わないため本文のみ）|
| `- 関連:` の ADR 参照 | `references` |
| `- 関連:` の GitHub Issue 参照 | `issues` |
| `- 関連:` の `docs/*.md` 参照 | 持たない（本文に残す）|
| `- 起点:` | `origin`。行が `Issue #N` で始まるときは `issues` にも入れる |
| `- 起点 Issue:` / `- 参照:` | `起点:` 行が無いときだけ `origin` へ入れる |
| `- 改訂対象:`（`legacy_adr` のみ）| `revises` |

分類の細則:

- **他リポジトリの ADR は `references` に入れない**。`関連:` にリポジトリ名を前置した ADR 参照
  （例: ADR-0043 の `outturn ADR-0013`）は ID 空間が別なので、`references` に入れると自リポジトリの
  同番号 ADR への誤った辺になる（`sara check` は実在する ID なので通してしまう）。本文に残す
- **他リポジトリの Issue は `issues` へ入れる**（`関連:` / `起点:` / `起点 Issue:` のどの行にあっても）。
  `owner/repo#N` 形式にすれば ID 空間が衝突しない（例: ADR-0043 の `yasunori0418/outturn#1`）
- `起点:` 行の Issue は原則 `origin` のテキストに残す。行が `Issue #N` で始まる（その ADR の直接の
  起点である）ときだけ `issues` にも入れる（例: ADR-0026 の `- 起点: Issue #17（…）`）。経緯として
  言及されるだけの Issue は入れない（例: ADR-0044 / ADR-0045 の「旧 epic #167 から統合された」の `#167`）
- `origin` のフォールバックは `起点:` → `起点 Issue:` → `参照:` の順（`legacy_adr` の実データでは
  `起点:` 31 本 / `起点 Issue:` 4 本 / `参照:` 5 本）。フォールバック時は `"起点 Issue: #26"`
  `"参照: PR #40, …"` のようにラベルを残す
- 同じ ADR を `supersedes`（`legacy_adr` では `revises`）と `references` の両方に入れない。
  前者を優先する（同一ペアに relation を二重に張らない）
- `docs/*.md` 参照は、その ADR が参照する記述が全て item へ分割され終わってから除く（→ #213）。
  `docs/concept.md` を指す参照は、「設計の哲学」など item を立てない節を指しうるので保留する

### `references` の向きと同着判定

`references` は**後発 → 先行の片方向**に張る。同一ペアに相互に張ると `sara check` が cycle として
error にする。先行 ADR 側からは `sara query` の `Is referenced by` で逆引きできる。新しく起こす ADR は
常に最後発なので、`関連:` の ADR をそのまま `references` に入れてよい。

`legacy_adr` どうしの先発・後発は、ADR 番号ではなく **author date** で判定する（committer date は
rebase で潰れる。`ADR-0033`〜`0042` の 10 本は committer date が同一時刻になる）。

```bash
git log --diff-filter=A --format=%at -1 -- docs/adr/<file>
```

番号と起票順は食い違うことがある（`ADR-0044`〜`0047` は起票順が 0046 → 0047 → 0044 → 0045）。
author date が同着の群は 5 つあり（ADR-0001〜0004 / ADR-0011・0012 / ADR-0033〜0037 /
ADR-0038・0039 / ADR-0040〜0042）、同着の中は ADR 番号順を後発とみなす。

この規則の結果、`legacy_adr` の本文 `関連:` にある前方参照（起票順で後発の ADR への参照）は
`references` に入らない。次はその検算リストで、列挙にある辺が frontmatter に無いのは規則どおり
（張り忘れではない）。23 本 46 辺:

```
ADR-0001 → 0059
ADR-0002 → 0003 / 0006 / 0015 / 0057 / 0058
ADR-0003 → 0004 / 0006            ADR-0004 → 0005 / 0007
ADR-0005 → 0007
ADR-0006 → 0007 / 0012 / 0013 / 0055 / 0056 / 0057
ADR-0007 → 0015 / 0055            ADR-0011 → 0055 / 0056 / 0057
ADR-0013 → 0015 / 0059            ADR-0014 → 0015
ADR-0015 → 0057 / 0058            ADR-0017 → 0058
ADR-0023 → 0055                   ADR-0025 → 0055
ADR-0026 → 0055 / 0060            ADR-0031 → 0055
ADR-0035 → 0060                   ADR-0036 → 0057
ADR-0038 → 0039 / 0040 / 0060     ADR-0039 → 0060
ADR-0043 → 0055 / 0056            ADR-0048 → 0049 / 0050 / 0053
ADR-0055 → 0057 / 0060
```

後発側の本文 `関連:` にその辺が無ければ、どちらの frontmatter にも現れない（ADR-0038 → 0040 が該当）。
本文に無い辺を frontmatter へ足すことはしない。

## `justifies` の実例

選定基準（寄与の有無・張らない 2 条件・監査の 2 方向）は [adr-rules.md](../agents/adr-rules.md) が
持つ。layat での判定の実例:

- 寄与した ADR が複数張る: `REQ-46fccb80-4bae-4d37-bc19-dded88e9a9c0` は ADR-0005 / 0013 / 0017 / 0022 の
  4 本が張る。`REQ-a5053191-1c6a-449b-9c5e-5ff49dc5aead` は帰属 ADR-0043 に加え、`--json` を導入した
  ADR-0033 と `tool.version` の供給元である ADR-0042 が張る
- item が `justifies` を受けない理由を明記している: `DSG-d2e17f4f-0d32-45c1-8125-17e589664c85` の
  「根拠が ADR 以外にあることによる」。一方 `REQ-3f541d39-da41-4ef8-858b-707f54cf6a29` の「`subdir`
  適用形は担当範囲外」は規範に含めないという宣言で、ADR に依拠していないことではない（線引きの実例は
  ADR-0040 本文の 2026-08-08 注記）
- 背景で前提として言及するだけ: ADR-0039 → `REQ-535b811d-dfc5-4eac-92db-737e70eb5415` の一括 eval
- 出典節が ADR を名指ししない item: `REQ-4fc98fa6-fd5f-4f25-8b10-60755bf49bd2` は ADR-0002 の決定節から引いた
- 委譲先で回収済みとして撤去した: ADR-0002 → `REQ-c847d1af-a437-46bb-bd64-42083810d034`・ADR-0026 →
  `REQ-c847d1af-a437-46bb-bd64-42083810d034`・ADR-0015 → `REQ-5923ac79-4a2d-43cd-b56c-2f1000c01b44`
  （それぞれ `REQ-844ee375-919f-4341-81e1-a5f89fd32840` / `REQ-dec58330-6dad-47f7-8f56-2402764a89c7` /
  `REQ-5c6b07da-3d06-414d-8770-4f438234b322` 側で回収済み。→ #254）
- 規範文が対称化の対象として直に述べるので張る: `REQ-07c3b735-3744-4778-a640-8c6fb66f4aa7` は symlink 側の
  foreign 警告を `REQ-622787dc-4512-4ce9-9c7d-7b32bbb70557` へ委譲しつつ、規範文が "symmetrically with the
  foreign warning for symlinks" と名指しするので ADR-0015 を張る
- 態度の供給元としてだけ挙がるので張らない: ADR-0018 → `REQ-c847d1af-a437-46bb-bd64-42083810d034` の
  「部分失敗を集約する姿勢の淵源」。決定節が実際に決めているので張る: ADR-0006 →
  `REQ-61856da1-8883-401e-ad57-9f326b96d400` の「冪等再実行による収束の前提」

`justifies` を受けない理由を散文で書いている item は次で列挙できる。

```bash
rg -n 'justifies' docs/requirements/ docs/design/ docs/infrastructure/ docs/test-plan/
```

## `legacy_adr` の扱い

`legacy_adr` は凍結している。新規には起こさず、本文も書き換えない。既存の `revises` と本文の blockquote
注記は当時の記録としてそのまま残り、新しくは足さない。

`legacy_adr` の決定を変えるときや、改訂を重ねた `legacy_adr` を整理するときは、
[adr-rules.md](../agents/adr-rules.md) の「統合と分割の手順」に従って決定ごとに `adr` を起こす。
全決定に後継が揃ったら、後継の `adr` から `supersedes` を張り、`legacy_adr` を `廃止` にする。
そのとき `legacy_adr` で変えるのは、本文のステータス行・frontmatter の `status`・冒頭の後継一覧の
バナーだけ。

## セルフチェック

ADR を書き終えたら、グラフ全体の検証と、`廃止` と `supersedes` の対応の検証を走らせる。

```bash
nix develop '.?dir=dev#sara' -c sara check
nix develop '.?dir=dev#sara' -c dev/tests/adr-supersede.sh
```

`sara check` の合格条件は **warning / error とも 0 件で exit 0 になること**。`sara.toml` は
`strict_mode = true`（→ ADR-0050）なので、`Orphan item`（接続漏れ）も `Broken reference`・
`Duplicate identifier`・`Circular reference detected` と同様に error になる。ADR で orphan が出るのは
`justifies` の埋め忘れなので、張ってから完了とする。`justifies` に書いた ID のタイポは orphan では
なく `Broken reference` として落ちる。

`adr-supersede.sh` は、`supersedes` で指された ADR が `廃止` であることと、`廃止` の ADR が 1 本以上の
`supersedes` から指されていることを確かめる。どちらも CI の `sara` job が走らせるが、required status
check ではないため、赤くなってもマージはブロックしない（→ ADR-0050）。

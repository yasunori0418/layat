# layat の ADR

ADR の粒度・本文の構成・`justifies` の選定基準は [ADR は 1 本に 1 つの決定を 4 節で書き、justifies は規範の形成に寄与した item へ張る](../quality/adr-one-decision-four-sections-justifies-by-contribution.md) が、
ステータス・出典・`supersedes` と `deprecated`・採番は [設計判断は ADR に記録し、改訂の書き戻しと item への接続を同じ変更の中で完了させる](../quality/adr-decisions-recorded-and-connected.md) が持つ。この文書はそれを layat へ当てはめた部分
（型とディレクトリ・起票の手順・frontmatter の対応表・検算用の実データ）だけを持つ。

## 型とディレクトリ

| 型 | prefix | ID | ファイル | 置き場所 |
|---|---|---|---|---|
| `adr` | `ADR` | `ADR-<フル UUIDv4>` | `<slug>.md` | `docs/adr/` |
| `sara_adr` | `SDR` | `SDR-<フル UUIDv4>` | `<slug>.md` | `docs/sara-adr/` |

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
`# <タイトル>` とし、ID を書かない。

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
    title: "<Issue のタイトル>"
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
  （例: [--json の item id は subject スコープで導く](subject-scoped-ids.md) の `関連:` にある outturn の ADR）は ID 空間が別なので、`references` に入れると自リポジトリの
  同番号 ADR への誤った辺になる（`sara check` は実在する ID なので通してしまう）。本文に残す
- **他リポジトリの Issue も `sources` へ入れる**（`関連:` / `起点:` / `起点 Issue:` のどの行にあっても）。
  `resource` を `https://github.com/owner/repo/issues/N`、`id` を `issue-<repo>-N` にすれば ID 空間が衝突しない
  （例: [--json の item id は subject スコープで導く](subject-scoped-ids.md) の `yasunori0418/outturn#1`）
- `起点:` 行の Issue は原則 `origin` のテキストに残す。行が `Issue #N` で始まる（その ADR の直接の
  起点である）ときだけ `sources` にも入れる（例: [shell.nix / default.nix の再現性は利用者の責任にする](shell-default-nix-best-effort.md) の `- 起点: Issue #4（…）`）。経緯として
  言及されるだけの Issue は入れない（例: [途中失敗した apply をインメモリ undo ジャーナルで巻き戻す](in-memory-undo-journal.md) の「旧 epic #167 から統合された」の `#167`）
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

選定基準（寄与の有無・張らない 2 条件・監査の 2 方向）は [ADR は 1 本に 1 つの決定を 4 節で書き、justifies は規範の形成に寄与した item へ張る](../quality/adr-one-decision-four-sections-justifies-by-contribution.md) が
持つ。layat での判定の実例:

- 寄与した ADR が複数張る: [project mode は世代を非公開にし、project / system mode は derivation 同一なら世代を積まず lstat ドリフト修復だけ行う](../requirements/project-mode-generation-skip.md) へは [世代スキップには profile に加え記録の一致も要する](generation-skip-requires-both-match.md) /
  [profile は解決済み root の roothash でキーする](profile-keyed-by-roothash-of-resolved-root.md) / [project mode では同じ link farm なら世代を積まず drift だけを直す](project-mode-generation-skip-and-drift-repair.md) /
  [世代スキップを project mode と system mode で行う](generation-skip-project-and-system.md) / [module と project では世代操作を公開しない](module-and-project-hide-generation-ops.md) の 5 本が張る。
  [--json は outturn 規約準拠のエンベロープを出す第 2 契約とする](../requirements/json-outturn-envelope.md) へは帰属の [--json は outturn 規約に常時準拠する](outturn-conformance-principle.md) に加え、`--json` を導入した
  [--json を全サブコマンド共通の persistent フラグにする](json-global-flag-for-all-subcommands.md) と `tool.version` の供給元である [VERSION ファイルをバージョンの一次情報にする](version-file-single-source.md) が張る
- item が `justifies` を受けない理由を明記している: [動的 entry 生成のヘルパを lib に置く](../design/lib-no-dynamic-entries.md) の
  「根拠が ADR 以外にあることによる」。一方 [root マーカーは `kind` を運ぶ入れ物でパス解決は engine が行う](../requirements/root-markers-runtime-resolution.md) の「`subdir`
  適用形は担当範囲外」は規範に含めないという宣言で、ADR に依拠していないことではない
- 背景で前提として言及するだけ: [apply --all の build を worker pool で並列に実行する](apply-all-parallel-build-worker-pool.md) → [apply --all は rootKind を 1 回の一括 eval で取る](../requirements/apply-all-batched-eval.md)
- 出典節が ADR を名指ししない item: [一部の entry だけを適用する --only を提供する](../requirements/no-only-flag.md) は [一部の entry だけを適用する --only を提供する](partial-apply-only-flag.md) の決定から見つけた
- 規範文が対称化の対象として直に述べるので張る: [copy が foreign 実ファイルを skip したときは warning で可視化する](../requirements/copy-foreign-skip-warning.md) は symlink 側の
  foreign 警告を [symlink 配置は親 dir を作り配置元/subpath を指すリンクを張り、foreign symlink は警告して後勝ちする](../requirements/symlink-placement-procedure.md) へ委譲しつつ、規範文が "symmetrically with the
  foreign warning for symlinks" と名指しするので [cross-config の同一 target は後勝ちにし、foreign symlink の上書きで警告する](cross-config-last-wins-with-foreign-warning.md) を張る
- 決定節が実際に決めているので張る: [新規配置と張替えを先に、世代のコミットを最後に行う](placement-order-commit-last.md) →
  [既存 symlink の張替えは unlink + symlink の 2 操作で行い冪等な再実行で収束させる](../requirements/symlink-replace-unlink-symlink.md) の「冪等再実行による収束の前提」

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
`strict_mode = true`（→ [sara check を strict モードで回す](../sara-adr/sara-check-strict-mode.md)）なので、`Orphan item`（接続漏れ）も `Broken reference`・
`Duplicate identifier`・`Circular reference detected` と同様に error になる。ADR で接続漏れが出るのは
`justifies` の埋め忘れなので、張ってから完了とする。`justifies` に書いた ID のタイポは接続漏れでは
なく `Broken reference` として落ちる。

`adr-supersede.sh` は、`supersedes` で指された ADR が `deprecated` であることを確かめる。どちらも CI の `sara` job が実行するが、required status
check ではないため、赤くなってもマージはブロックしない（→ [sara check を required status check にしない](../sara-adr/sara-check-not-required.md)）。

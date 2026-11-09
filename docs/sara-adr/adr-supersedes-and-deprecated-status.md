---
id: "SDR-3ff9f355-a667-476e-b10d-087248cdbfcf"
type: sara_adr
name: "ADR の失効は supersedes と 廃止 で表す"
status: stable
sources:
  - id: "issue-499"
    resource: "https://github.com/yasunori0418/layat/issues/499"
    title: "#499"
origin: "Issue #499 の grilling（2026-11-01）で確定。ADR-0048 §2・§3 を書き直した"
justifies:
  - "QA-030f926b-5ae7-4543-b8cb-e66aae5e2b5e"
references:
  - "SDR-1afa3a31-224d-492d-96d9-d3f6c9cb20cd"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:08:50+09:00"
---
# ADR の失効は supersedes と 廃止 で表す

- ステータス: stable
- 日付: 2026-11-01
- 関連: [ADR は単独で覆せる 1 つの決定だけを書く](adr-one-decision-granularity.md), GitHub Issue #499, `docs/agents/adr-rules.md`, `docs/adr/README.md`
- 起点: Issue #499 の grilling（2026-11-01）で確定。ADR-0048 §2・§3 を書き直した

## 背景

`docs/model.yaml` は、ADR 間の関係として節単位の部分改訂を表す `revises` だけを持っていた。
文書丸ごとの失効を表す `supersedes` と、ステータスの `廃止` は持たなかった。当時は丸ごと失効した
ADR の実例が無く、実例の無い関係を先に置くと部分改訂に誤用されると考えたためである。

[ADR は単独で覆せる 1 つの決定だけを書く](adr-one-decision-granularity.md) という規則を採ると、
決定の変更は ADR の丸ごとの置き換えになる。部分改訂という扱いは要らなくなる。

## 決定

ADR の失効は、後継の ADR から張る `supersedes` と、失効した側のステータス `廃止` で表す。新しく
起こす ADR は `revises` と改訂注記を持たない。改訂を伴わない参照は、これまでどおり `references` で表す。

`廃止` にできる条件・`supersedes` を張る時期・廃止した ADR に残すものは
`docs/agents/adr-rules.md` の「改訂は丸ごとの置き換えで表す」節に置く。`廃止` と `supersedes` の
対応は `dev/tests/adr-supersede.sh` が frontmatter から確かめる。

## 根拠

置き換えだけで改訂を表せば、有効な ADR は本文だけで読める。廃止した ADR は冒頭のバナーで後継へ
誘導するので、どの決定がいまも有効かを注記から組み立て直す必要が無い。

`supersedes` を誤用する心配は、粒度の規則で解消した。1 本に 1 決定なら、決定の変更は必ず文書
丸ごとの失効になる。

## 棄却した代替案

### `revises` を新しい ADR にも残す

部分改訂が残り、粒度の規則の利点が消える。新しい ADR 同士の改訂も `supersedes` だけで表す。

### 決定ごとに後継を揃えず、1 決定の後継ができた時点で `廃止` にする

残りの決定に後継が無いまま `廃止` にすると、有効な決定が「読まなくてよい」ADR の中に取り残される。
全決定に後継が揃うまで `廃止` にしない。

---
id: "ADR-1ea2b93a-f83f-46c6-b925-cb0b38f4fd3c"
type: adr
name: "project mode では copy と out-of-store も ephemeral として扱う"
section: placement
status: stable
origin: "ADR-0005 の決定「copy / out-of-store も同原則を継承」と ADR-0019 §3 を書き直した"
justifies:
  - "REQ-60787ed2-4176-4bdd-800f-1600c0315551"
references:
  - "ADR-c4ad9987-850e-4834-bbea-d8de3b073548"
  - "ADR-4dca6ef4-2791-48d5-aa82-7037630968a6"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:26:59+09:00"
---
# project mode では copy と out-of-store も ephemeral として扱う

- ステータス: stable
- 日付: 2026-11-01
- 関連: [copy は世代の外に place-once で置く](copy-outside-generations-place-once.md), [ephemeral 配置は git に干渉しない](ephemeral-no-git-interference.md), `docs/concept.md`, `docs/spec.md`
- 起点: ADR-0005 の決定「copy / out-of-store も同原則を継承」と ADR-0019 §3 を書き直した

## 背景

project mode の配置物は ephemeral である。[ephemeral 配置は git に干渉しない](ephemeral-no-git-interference.md)。リポジトリにはコミットしない。
一方で copy は利用者が持ち主となって編集する実ファイルで（[copy は世代の外に place-once で置く](copy-outside-generations-place-once.md)）、out-of-store は store の外を指す。
この 2 つの method も ephemeral として扱うのか、コミットしてよいのかを決める必要があった。

## 決定

project mode では copy と out-of-store も ephemeral として扱い、新しい意味論を持ち込まない。
copy target は clone ごとに place-once で置き直され、その編集は clone の中だけで使い捨てになる。
`layat gitignore` は method を区別せず、copy target を含む全 target を列挙する。copy target を
コミットして共有したい場合は layat の責務外とする。規範の細目は
[gitignore は method を区別せず copy target も含めて全 target を列挙する](../requirements/gitignore-includes-copy-targets.md) に置く。

## 根拠

method ごとにコミットの可否を分けると、「project mode の配置物はコミットしない」という単純な
不変条件が崩れる。copy target だけをコミットすると、place-once による配置と git による管理が
同じファイルに重なり、どちらが正なのかが曖昧になる。

## 棄却した代替案

### project mode の copy target はコミットする（gitignore から外す）

編集をチームで共有できるが、ephemeral の原則に例外ができる。clone 時の place-once と、
コミット済みのファイルが衝突する。

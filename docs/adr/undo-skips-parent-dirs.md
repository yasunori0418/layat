---
id: "ADR-c45921e1-e543-47ef-8d34-b52f2810eee3"
type: adr
name: "配置の前に作った中間ディレクトリを undo の対象にしない"
status: stable
sources:
  - id: "issue-168"
    resource: "https://github.com/yasunori0418/layat/issues/168"
    title: "#168"
  - id: "issue-172"
    resource: "https://github.com/yasunori0418/layat/issues/172"
    title: "#172"
origin: "旧 epic #167 から統合された epic #172 の grilling セッション（2026-07-11・2026-07-12）で確定。ADR-0044 §1（インメモリ undo ジャーナル）の中間ディレクトリの扱いを書き直した"
justifies:
  - "REQ-5e75aabc-0e8f-4a6c-92bd-a712dc68a940"
references:
  - "ADR-56325270-20bb-4959-8498-e3223c35e24b"
  - "ADR-74423bb8-1453-4689-a25e-30bf46536afc"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:26:59+09:00"
---
# 配置の前に作った中間ディレクトリを undo の対象にしない

- ステータス: stable
- 日付: 2026-11-03
- 関連: [途中失敗した apply をインメモリ undo ジャーナルで巻き戻す](in-memory-undo-journal.md), [新規配置と張替えを先に、世代のコミットを最後に行う](placement-order-commit-last.md), GitHub Issue #168, #172
- 起点: 旧 epic #167 から統合された epic #172 の grilling セッション（2026-07-11・2026-07-12）で確定。ADR-0044 §1（インメモリ undo ジャーナル）の中間ディレクトリの扱いを書き直した

## 背景

[途中失敗した apply をインメモリ undo ジャーナルで巻き戻す](in-memory-undo-journal.md)。巻き戻す対象は FS への書き込みである。配置の直前には、target の親ディレクトリを
`mkdir -p` 相当で作る。この中間ディレクトリもジャーナルに載せるかを決める必要があった。

## 決定

配置の直前に新しく作った中間ディレクトリはジャーナルに記録しない。巻き戻しで leaf（symlink /
copy）を消した後に残る空のディレクトリは、そのまま残す。

ジャーナルに載せる FS 変更の範囲は [途中失敗した apply / rollback はインメモリ undo ジャーナルで全 FS 変更を巻き戻す](../requirements/undo-journal-rollback.md) に置く。

## 根拠

leaf を消せば、target は呼び出し側から見て不在に戻る。残るのは空のディレクトリだけで、
データは失われない。entry が config から消えたときに残る空ディレクトリと見分けが付かず、
後の除去で、空の祖先ディレクトリを消す仕組みがこれも消す。

複数の leaf が同じ中間ディレクトリを新しく共有することがある。ジャーナルで扱うなら、どの
leaf の巻き戻しでディレクトリを消すかを重複排除して決める必要がある。データ損失の恐れが無い
ことに対して、その仕組みは割に合わない。

## 棄却した代替案

### 中間ディレクトリの作成も記録し、巻き戻しで rmdir する

適用前の状態に厳密に戻せるが、leaf 間で共有する中間ディレクトリの重複排除をジャーナルに
持ち込むことになる。得られるのは空ディレクトリが残らないことだけである。

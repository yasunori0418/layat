---
id: "ADR-b7f9b972-69b7-4bbe-affc-97978e1a55f7"
type: adr
name: "root 相対マーカーのリンクを target 起点の相対パスで張る"
status: 採用
issues:
  - "#362"
origin: "Issue #362（src 側のプロジェクトルート実行時解決）の grilling（2026-10-12）。ADR-0059 §1 と §3 を書き直した"
justifies:
  - "REQ-a77181dd-a59c-40bb-855e-3430ad970eb6"
  - "DSG-fd39b9ee-a5d2-46c2-bdcd-fc399a1e0296"
  - "REQ-a8a923ad-07fb-4582-b90a-07a6e0c41baa"
  - "REQ-622787dc-4512-4ce9-9c7d-7b32bbb70557"
  - "REQ-c9ab91c1-f778-4f87-a2ea-c66d6b3c2575"
references:
  - "ADR-5e46ebd6-e44b-48dc-a93f-a8a3a3903967"
  - "ADR-6d2327f7-f513-42c3-89f8-808467db1914"
  - "ADR-b06dd8ab-317e-41ee-9c80-dd70dcd7f2fd"
  - "ADR-f40c3d37-638d-420a-a0fe-3908cbf0aec0"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:28:06+09:00"
---
# ADR-b7f9b972-69b7-4bbe-affc-97978e1a55f7: root 相対マーカーのリンクを target 起点の相対パスで張る

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-5e46ebd6-e44b-48dc-a93f-a8a3a3903967, ADR-6d2327f7-f513-42c3-89f8-808467db1914, ADR-b06dd8ab-317e-41ee-9c80-dd70dcd7f2fd, ADR-f40c3d37-638d-420a-a0fe-3908cbf0aec0, `docs/design.md`
- 起点: Issue #362（src 側のプロジェクトルート実行時解決）の grilling（2026-10-12）。ADR-0059 §1 と §3 を書き直した

## 背景

リポジトリの中の実体（`dev/skills/<名前>/`）を、同じリポジトリの `.claude/skills/<名前>` へ可変の
symlink で置きたいという需要が出た。`mkOutOfStoreSymlink` の引数は評価時に決まる絶対パスで、
pure eval の flake は自分のチェックアウトの絶対パスを得られない。

engine の前世代との比較・古くなった配置の除去・drift 修復・`--json` の `old` / `new` は、どれもディスクの
readlink の結果と、Entry だけから計算するリンク先の文字列の完全一致で判定している。

## 決定

engine は root 相対マーカーの entry を、target の親ディレクトリから `src/subpath` への相対パスを
中身とする symlink として張る。この形は、配置の時点で target の祖先が実ディレクトリであるという
不変条件に依存する。foreign な祖先 symlink は配置前に止め、自己記録の古くなった祖先 symlink は
配置前に除く。リンク先の計算は DSG-fd39b9ee-a5d2-46c2-bdcd-fc399a1e0296、祖先 symlink の扱いは
REQ-c9ab91c1-f778-4f87-a2ea-c66d6b3c2575 に置く。

## 根拠

リンク先が root を含まない文字列になるので、比較する処理に手を入れずに済む。planner・古くなった配置の除去・
drift 修復・`--json` は相対の文字列をそのまま比べ、前世代の記録からも同じ文字列を計算し直せる。
リンクの中身が root の絶対パスを含まないから、worktree を動かしたり改名したりしてもリンクは
有効なまま残る。home mode でも HOME に依らない書き方になる。

相対リンクは、target の親が symlink だとその先を起点に解決され、別の場所を指す。祖先の不変条件は
これを防ぐ前提なので、この不変条件を緩める変更は root 相対マーカーの正しさを壊す。

## 棄却した代替案

### engine が `<root>/<src>/<subpath>` を絶対パスにして張る

リンク先の計算に root が要り、planner・古くなった配置の除去・drift 修復・`--json` の比較する処理を全て
root 付きに書き換えることになる。worktree を動かすとリンクが切れ、home mode では HOME に依存する
文字列が残る。

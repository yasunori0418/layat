---
id: "REQ-622787dc-4512-4ce9-9c7d-7b32bbb70557"
type: requirement
name: "symlink 配置は親 dir を作り配置元/subpath を指すリンクを張り、foreign symlink は警告して後勝ちする"
status: stable
derives_from:
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  To place a symlink the engine SHALL first create the parent directory of the target
  (equivalent to `mkdir -p`), then handle an existing symlink at the target: one recorded
  by the entry's own previous-generation manifest SHALL be replaced without a warning,
  falling under the ordinary output discipline of the placement report, whereas an
  unrecorded (foreign) symlink — belonging to another layat profile, another tool, or
  placed by hand — SHALL be replaced as well but SHALL emit a warning, the last writer
  thereby winning. It SHALL then create a symlink at `<root>/<target>` pointing at
  `<placement source>/<subpath>`, the placement source being the Nix store path for a
  store link and the absolute path held by the marker for an out-of-store symlink
  (`srcKind: "outOfStore"`). For a root-relative symlink (`srcKind: "rootRelative"`) the
  link SHALL instead point at the relative path from the directory of the target to
  `<src>/<subpath>`, computed as specified by [rootRelative の LinkDest は target の親から src/subpath への相対パスを Entry だけから計算し、root が要るのは配置前の確認だけにする](../design/root-relative-linkdest.md). The
  processing SHALL be the same whether the subpath denotes a file or a directory. Two
  different configs (profiles) aiming at the same target SHALL be treated as something not
  meant to happen; the last writer SHALL be tolerated and the overwriting of a foreign
  symlink SHALL be made visible by the warning. As an exception, a conflict between configs
  selected by one `apply --all` of a single entrypoint SHALL be stopped before the build by
  the preflight of [apply --all は build 前に選択 config 間の target 衝突を検出して停止する](apply-all-cross-config-preflight.md).
specification_ja: |
  symlink の配置にあたり engine は、まず target の親ディレクトリを作成し（`mkdir -p` 相当）、
  次に target の既存 symlink を扱わなければならない。当該 entry 自身の前世代 manifest が記録
  した symlink は warning を出さずに置き換え（配置レポートの通常の出力規律に従う）、記録の無い
  （foreign な）symlink——他 layat profile / 他ツール / 手動——も置き換えるが warning を出す
  （後勝ち）。そのうえで `<配置元>/<subpath>` を
  指す symlink を `<root>/<target>` に作成しなければならない。配置元は store link では Nix
  ストアパス、out-of-store symlink（`srcKind: "outOfStore"`）では marker の絶対パスで
  なければならない。root 相対の symlink（`srcKind: "rootRelative"`）では、target の親
  ディレクトリから `<src>/<subpath>` への相対パスを指さなければならない（計算は
  [rootRelative の LinkDest は target の親から src/subpath への相対パスを Entry だけから計算し、root が要るのは配置前の確認だけにする](../design/root-relative-linkdest.md) の担当）。subpath がファイル・
  ディレクトリのいずれでも処理は同じでなければならない。別 config（別 profile）が同一
  target を狙うことは起きない前提として扱わなければならず、後勝ちを許容しつつ foreign
  symlink の上書きは warning で可視化しなければならない。例外として、同一 entrypoint の
  `apply --all` で選択された config 間の衝突は [apply --all は build 前に選択 config 間の target 衝突を検出して停止する](apply-all-cross-config-preflight.md) の前段の確認で build 前に
  止まらなければならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T22:03:12+09:00"
---
# REQ-622787dc-4512-4ce9-9c7d-7b32bbb70557: symlink 配置は親 dir を作り配置元/subpath を指すリンクを張り、foreign symlink は警告して後勝ちする

## 仕様

```
1. target の親ディレクトリを作成（mkdir -p 相当。緩和対象の祖先 symlink / 実 dir target は PreRemove 除去済み・foreign は 0 で弾き済み）
2. target が既存 symlink のとき:
   - 自身の前世代 manifest が記録した symlink → そのまま置き換える（silent）
   - 記録の無い symlink（foreign = 他 layat profile / 他ツール / 手動）→ warning を出して置き換える（後勝ち）
3. <配置元>/<subpath> を指す symlink を <root>/<target> に作成（os.Symlink）
   - store link:        配置元 = Nix ストアパス
   - out-of-store:      配置元 = marker の絶対パス
   - root 相対:         target の親から <src>/<subpath> への相対パス
```

- subpath がファイル・ディレクトリどちらでも同じ処理
- 別 config（別 profile）が同一 target を狙うのは基本「衝突させない前提」。後勝ちを許容しつつ
  foreign symlink 上書きは warning で可視化する

> **上は原文の写しで、規範は frontmatter が正**。手順 0〜0.7（配置前除去・退避）は
> [祖先 symlink は自己記録の古くなったものだけを配置前除去し、それ以外はエラーで停止する](ancestor-symlink-nest-migration.md) / [実 dir の target は全 leaf が除去可能なときだけ全体を配置前除去して symlink 化する](real-dir-target-migration.md) / [method 変更は symlink→copy のみ配置前除去で移行し、copy→symlink は移行しない](method-change-symlink-to-copy-migration.md) / [backup 退避は配置前除去の後・配置の前に置き、drift 修復の処理でも同じく実施する](backup-stage-position.md)、張替えを unlink + symlink の
> 2 操作で行うことは [既存 symlink の張替えは unlink + symlink の 2 操作で行い冪等な再実行で収束させる](symlink-replace-unlink-symlink.md) の担当。`shellHook` 再入で起きる cross-config の振動は
> [同一 target を複数 config で狙うことによる振動はユーザー責任とし warning で可視化するに留める](cross-config-target-oscillation.md) の担当。記録済み symlink の張替えを既定 silent とし `-v` で可視化するという
> 出力規律そのもの（→ [指定が無いときは成功時に何も出力しない](../adr/silent-on-success.md)・[-v で配置レポートを opt-in で出す](../adr/verbose-opts-into-placement-report.md)）は [成功時はデフォルト沈黙とし warning と error は常時 stderr へ出す](silent-on-success.md) / [冗長度は -v、デバッグは --debug に分離し --json と直交させる](verbose-and-debug-separation.md) の担当で、本 item は
> 「記録済みは warning にせず配置レポート側で扱い、foreign は warning を出す」という
> 対比だけを規範とする。

## 出典

`docs/spec.md`「配置動作仕様」→「symlink モード」の手順 1〜3 と、同節の箇条書き
第 3・4 項。

決定の実体は [cross-config の同一 target は後勝ちにし、foreign symlink の上書きで警告する](../adr/cross-config-last-wins-with-foreign-warning.md)。root 相対の配置元は [root 相対マーカーのリンクを target 起点の相対パスで張る](../adr/root-relative-link-as-relative-path.md)が加えた。

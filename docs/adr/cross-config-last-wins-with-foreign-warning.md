---
id: "ADR-095220c9-5a3d-4fd4-94bd-8fa1a02dd36d"
type: adr
name: "cross-config の同一 target は後勝ちにし、foreign symlink の上書きで警告する"
section: placement
status: stable
sources:
  - id: "issue-550"
    resource: "https://github.com/yasunori0418/layat/issues/550"
    title: "#550"
origin: "ドキュメント計画段階の最終レビュー（2026-06-14）で確定。ADR-0015 §2 を書き直した"
justifies:
  - "REQ-622787dc-4512-4ce9-9c7d-7b32bbb70557"
  - "REQ-5c6b07da-3d06-414d-8770-4f438234b322"
  - "REQ-07c3b735-3744-4778-a640-8c6fb66f4aa7"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:27:41+09:00"
---
# cross-config の同一 target は後勝ちにし、foreign symlink の上書きで警告する

- ステータス: stable
- 日付: 2026-11-01
- 関連: GitHub Issue #550
- 起点: ドキュメント計画段階の最終レビュー（2026-06-14）で確定。ADR-0015 §2 を書き直した

## 背景

別の config（別の profile）や別のツールが同じ target に配置することがある。engine が知っているのは
自分の前世代 manifest だけで、他の profile がその target を所有しているかは分からない。警告なしに
上書きすると、ユーザーは配置が奪われたことに気付けない。

## 決定

別 config が同じ target に配置することは起きない前提とし、実行時は後勝ちにする。engine は、自分の前世代
manifest に記録の無い symlink（foreign）を置き換えるときに warning を出し、停止はしない。

target の状態ごとの扱い（不在・自分の記録した symlink・foreign symlink・実ファイルや実ディレクトリ）は
[symlink 配置は親 dir を作り配置元/subpath を指すリンクを張り、foreign symlink は警告して後勝ちする](../requirements/symlink-placement-procedure.md) に置く。同一 manifest 内の衝突を eval で止めるか、
この実行時の後勝ちで扱うかの分け方は [target 衝突の検出処理を同一 manifest 内と cross-config で分ける](../requirements/target-collision-eval-detection.md) に置く。copy が
foreign の実ファイルを見つけて配置を見送るときの、symlink 側と対になる warning は
[copy が foreign 実ファイルを skip したときは warning で可視化する](../requirements/copy-foreign-skip-warning.md) に置く。

## 根拠

所有者を管理するには、store の外に可変な registry を持つ必要がある。layat は配置の状態を profile と
世代の manifest にしか持たない方針で、その方針を変えてまで防ぐ価値は無い。

warning にとどめれば、意図しない上書きは見えるようになり、意図した乗り換え（別ツールから layat への
移行など）は止まらない。1 回の `apply --all` に載る config どうしの衝突のうち、eval で検出できる
ものは [apply --all は選択した config 間の target 重複を build 前に止める](apply-all-preflight-target-conflict.md) が前段で止める。この ADR の後勝ちが受け持つのは、それ以外の衝突である。

## 棄却した代替案

### グローバルな registry で target の所有者を管理する

store の外に可変な状態が要り、registry と実際のファイルシステムのずれという新しい故障点が
増える。

### foreign symlink を上書きせずエラーで止める

別ツールからの乗り換えのたびに手で symlink を消す必要があり、移行の手間が増える。実ファイル・
実ディレクトリは失うと戻せないので止めるが、symlink は張り直せば戻る。

### 警告せずに上書きする

配置が奪われたことにユーザーが気付けない。

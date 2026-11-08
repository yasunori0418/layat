---
id: "ADR-56325270-20bb-4959-8498-e3223c35e24b"
type: adr
name: "途中失敗した apply をインメモリ undo ジャーナルで巻き戻す"
status: stable
sources:
  - id: "issue-168"
    resource: "https://github.com/yasunori0418/layat/issues/168"
    title: "#168"
  - id: "issue-172"
    resource: "https://github.com/yasunori0418/layat/issues/172"
    title: "#172"
origin: "旧 epic #167 から統合された epic #172 の grilling セッション（2026-07-11・2026-07-12）で確定。ADR-0044 §1（インメモリ undo ジャーナル）を書き直した"
justifies:
  - "REQ-5e75aabc-0e8f-4a6c-92bd-a712dc68a940"
  - "REQ-b4e4b65d-6e35-40c3-a00e-20c14043df6f"
  - "TP-deb05610-44bc-4962-8939-952392e5fbd0"
references:
  - "ADR-74423bb8-1453-4689-a25e-30bf46536afc"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:24:30+09:00"
---
# ADR-56325270-20bb-4959-8498-e3223c35e24b: 途中失敗した apply をインメモリ undo ジャーナルで巻き戻す

- ステータス: stable
- 日付: 2026-11-03
- 関連: [新規配置と張替えを先に、世代のコミットを最後に行う](placement-order-commit-last.md), GitHub Issue #168, #172
- 起点: 旧 epic #167 から統合された epic #172 の grilling セッション（2026-07-11・2026-07-12）で確定。ADR-0044 §1（インメモリ undo ジャーナル）を書き直した

## 背景

世代のコミットは FS への反映が全て成功した後の 1 点に置いている
（[新規配置と張替えを先に、世代のコミットを最後に行う](placement-order-commit-last.md)）。途中で失敗すれば世代は進まないが、FS には
部分的な配置が残る。この残骸は 2 つの点で困る。

1. 失敗した run が新しく置いた symlink や copy は前世代の manifest に記録が無い。次の apply は
   前世代との差分で古くなった配置を除去するので、これを除去の対象にできない
2. 初回の apply が途中で失敗すると、`reset` は entries の定義から対象を決めるので巻き戻しに
   使えず、残骸がそのまま残る

## 決定

apply と rollback は、FS へ 1 回書き込むたびに、その逆操作をプロセスのメモリ上のジャーナルへ
1 件記録する。途中のどこかで失敗したら、それまでの記録を逆順にたどって適用前の状態へ戻して
から error を返す。この巻き戻しは常に有効で、フラグでは切り替えない。

順操作と逆操作の対応、`--recopy` の上書きを rename 退避で行う手順は
[途中失敗した apply / rollback はインメモリ undo ジャーナルで全 FS 変更を巻き戻す](../requirements/undo-journal-rollback.md) と [recopy の上書きは削除ではなく同一親内への rename 退避で行う](../requirements/recopy-rename-aside.md) に置く。
故障を誘発して検証する方針は [原子性は実 FS の条件で故障を誘発して不変条件ごとに検証する](../test-plan/fault-injection-atomicity.md) に置く。

## 根拠

逆操作に要る情報（張替え前のリンク先・退避先のパスなど）は、順操作を実行した時点で手元に
ある。メモリに積むだけなので永続化のコストが掛からず、コストも変更した量に比例する。

記録を逆順にたどれば、生成と厳密に逆の順で戻せる。たとえば祖先の symlink を除去して配下に
子を置いた run は、子の削除、祖先の再作成の順に戻る。順序を巻き戻し側で組み立て直す必要が
無い。

常に有効にするのは、「失敗した apply は FS に痕跡を残さない」を apply の意味論として扱う
ためである。

## 棄却した代替案

### apply の前に対象のツリー全体を退避しておく

失敗時には退避物を戻すだけで済むが、target ごとに一時コピーが要り、大きなツリーではコストが
跳ね上がる。

### 残骸は冪等な再実行に任せる

前世代に記録の無い残骸は次の apply の古くなった配置の除去に掛からず、初回 apply の失敗では `reset` も
使えない。再実行では片付かない残骸が残る。

### 巻き戻しをフラグで有効にする

失敗した run の残骸が残るかどうかが、起動のしかたで変わってしまう。

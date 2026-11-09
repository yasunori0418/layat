---
id: "ADR-509e15a2-9097-404d-bc3c-171fe63dd3c8"
type: adr
name: "クラッシュ耐性のための永続 WAL を持たない"
status: stable
sources:
  - id: "issue-168"
    resource: "https://github.com/yasunori0418/layat/issues/168"
    title: "#168"
  - id: "issue-172"
    resource: "https://github.com/yasunori0418/layat/issues/172"
    title: "#172"
origin: "旧 epic #167 から統合された epic #172 の grilling セッション（2026-07-11・2026-07-12）で確定。ADR-0044 §4 を書き直した"
justifies:
  - "REQ-61856da1-8883-401e-ad57-9f326b96d400"
  - "REQ-5e75aabc-0e8f-4a6c-92bd-a712dc68a940"
references:
  - "ADR-74423bb8-1453-4689-a25e-30bf46536afc"
  - "ADR-3aebefee-405d-4342-b754-eef8eac29b1b"
  - "ADR-56325270-20bb-4959-8498-e3223c35e24b"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T17:28:05+09:00"
---
# クラッシュ耐性のための永続 WAL を持たない

- ステータス: stable
- 日付: 2026-11-03
- 関連: [新規配置と張替えを先に、世代のコミットを最後に行う](placement-order-commit-last.md), [symlink を unlink と symlink の 2 操作で張り替える](relink-unlink-then-symlink.md), [途中失敗した apply をインメモリ undo ジャーナルで巻き戻す](in-memory-undo-journal.md), GitHub Issue #168, #172
- 起点: 旧 epic #167 から統合された epic #172 の grilling セッション（2026-07-11・2026-07-12）で確定。ADR-0044 §4 を書き直した

## 背景

undo ジャーナル（[途中失敗した apply をインメモリ undo ジャーナルで巻き戻す](in-memory-undo-journal.md)）は、プロセスが error を返せる
失敗を巻き戻す。SIGKILL や電源断、回収されない panic でプロセスが終わると、メモリ上の
ジャーナルも失われる。この場合まで保証するかを決める必要があった。

## 決定

ジャーナルはプロセスのメモリ上にだけ置き、ディスクへ永続化しない。クラッシュで終わった run は
巻き戻さない。クラッシュ後の保証は、世代を未コミットのまま残すことと、冪等な再実行で FS が
収束すること（[新規配置と張替えを先に、世代のコミットを最後に行う](placement-order-commit-last.md)）が担う。

クラッシュを巻き戻しの対象外とする規範は [途中失敗した apply / rollback はインメモリ undo ジャーナルで全 FS 変更を巻き戻す](../requirements/undo-journal-rollback.md) に、
張替えの間にクラッシュしたときの扱いは [既存 symlink の張替えは unlink + symlink の 2 操作で行い冪等な再実行で収束させる](../requirements/symlink-replace-unlink-symlink.md) に置く。

## 根拠

error を返したときの巻き戻しとクラッシュ耐性は別の保証である。前者は、apply が失敗を自分で
報告できた場合を全て適用前の状態へ戻す。後者には永続 WAL、fsync の規律、クラッシュ後の
回復手順が要る。

クラッシュの場合は、世代が進まず、同じ manifest の再適用で FS が収束する。symlink の張替えも
2 操作の間のクラッシュを再実行で収束させている（[symlink を unlink と symlink の 2 操作で張り替える](relink-unlink-then-symlink.md)）。
この下支えで運用できているので、ジャーナルはその上に「プロセスが生きている間の失敗は
すぐ元に戻す」という品質を足すだけにとどめる。

## 棄却した代替案

### ディスク上に undo ログ（永続 WAL）を持つ

クラッシュまで含めて巻き戻せるが、fsync の規律・回復手順・WAL 自体が破損したときの扱いが要る。
途中失敗を元に戻すという目的に対して、実装と保守のコストが大きすぎる。

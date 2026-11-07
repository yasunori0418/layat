---
id: "ADR-f90509f2-d75a-4162-baa5-a5cfdb8a7c30"
type: adr
name: "世代は link farm と nix profile に乗せる"
status: stable
origin: "ADR-0002 §決定（nix profile に乗せる）を書き直した"
justifies:
  - "REQ-1be4d678-959c-44d7-a346-44bfd95af56e"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:28:06+09:00"
---
# ADR-f90509f2-d75a-4162-baa5-a5cfdb8a7c30: 世代は link farm と nix profile に乗せる

- ステータス: stable
- 日付: 2026-11-01
- 起点: ADR-0002 §決定（nix profile に乗せる）を書き直した

## 背景

layat は配置を世代として積み、前の世代へ戻せるようにする。世代管理には profile symlink の
差し替えや GC root の登録が要る。世代の実体を何で持つかを決める必要がある。

## 決定

1 回の適用で置くものを、lib が生成する store 内の link farm derivation 1 つにまとめる。engine は
これを layat の profile の新しい世代としてコミットし、世代番号・GC root・rollback を Nix の profile
機構から得る。

コミットは世代バックエンドの契約を通して行い、既定のバックエンドが nix profile を更新する。
link farm の中身とコミットの手順は REQ-1be4d678-959c-44d7-a346-44bfd95af56e に置く。

## 根拠

nix profile は profile symlink の差し替えだけで atomic な切り替えと rollback を実現し、各世代が
そのまま GC root になる。世代の採番と GC に対する安全性を自前で担わずに済み、Nix エコシステムの
機構に乗るという要求にも合う。任意のパスへの配置と世代管理を組み合わせた既存ツールは無く、
この組み合わせを layat が提供する。

## 棄却した代替案

### store 外の可変 JSON マニフェストで世代を表す

実装は単純だが、GC root を作らない。参照中の store パスが GC で消えて rollback できなくなるか、
GC root の管理を自前で持つことになる。

### 独自の軽量な世代ログを持つ

nix profile の機構を作り直すことになり、GC に対する安全性も自前で担保しなければならない。

### 世代を取らずバックアップ（`.bak`）だけを残す

前の状態へ戻せるという要求に応えられない。

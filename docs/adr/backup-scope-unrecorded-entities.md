---
id: "ADR-c42a51ff-fc51-4d19-b8c6-2b79704dd30f"
type: adr
name: "--backup は配置を塞ぐ記録外の実体だけを退避する"
status: stable
sources:
  - id: "issue-169"
    resource: "https://github.com/yasunori0418/layat/issues/169"
    title: "#169"
  - id: "issue-172"
    resource: "https://github.com/yasunori0418/layat/issues/172"
    title: "#172"
origin: "旧 epic #167 から統合された epic #172 の grilling セッション（2026-07-11・2026-07-12）で確定。ADR-0045 §2 を書き直した"
justifies:
  - "REQ-5dd5a4e9-6162-4fa5-b295-66844f5a4f3b"
  - "REQ-9b0046e0-8ddc-4c0b-940e-3fe6f36d0e98"
references:
  - "ADR-33d8c466-6b38-46b6-8696-ad7521c72786"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:39:38+09:00"
---
# ADR-c42a51ff-fc51-4d19-b8c6-2b79704dd30f: --backup は配置を塞ぐ記録外の実体だけを退避する

- ステータス: stable
- 日付: 2026-11-03
- 関連: [--backup を = 区切りで suffix を取る省略可能な値のフラグにする](backup-optional-value-flag.md), GitHub Issue #169, #172
- 起点: 旧 epic #167 から統合された epic #172 の grilling セッション（2026-07-11・2026-07-12）で確定。ADR-0045 §2 を書き直した

## 背景

layat は、自分が前世代に置いて古くなったものなら配置前に除去して自動で移行する。一方、記録外の実体
（他人のデータ）は判定なしに動かすと危険なので、自動移行の対象から外して conflict で止めて
いる。`--backup` はこの隙間を埋める手段で（[--backup を = 区切りで suffix を取る省略可能な値のフラグにする](backup-optional-value-flag.md)）、
何を退避の対象にするかを決める必要があった。

## 決定

`--backup` が退避するのは、`--backup` が無ければ conflict で止まるか、copy の foreign として
skip される記録外の実体に限る。実ディレクトリの target の移行が成り立たないときは、
ディレクトリ全体を 1 回の rename で退避する。祖先の symlink による conflict は退避の対象に
しない。

対象の外延（symlink モードの foreign な実体、copy の構造不一致、copy の foreign な実ファイル、
method の copy から symlink への変更など）は [backup 退避は配置前除去の後・配置の前に置き、drift 修復の処理でも同じく実施する](../requirements/backup-stage-position.md) と
[apply --backup は配置を塞ぐ記録外実体を rename 退避してから配置する](../requirements/apply-backup-rename-aside.md) に置く。

## 根拠

退避で解けるのは、target の位置に実体が 1 つあるという問題だけである。祖先の symlink による
conflict は、ネストの構造がそもそも成り立たないという問題で、退避しても解けない。祖先を
退避すると、配下にある別の entry のネストの前提まで崩れうる。1 つの target を 1 つの退避先へ
rename するモデルでは表せず、無理に適用すると別の問題を隠すことになる。解消の手段は entry の
定義を見直すことである。

実ディレクトリを丸ごと 1 回で退避するのは、移行を部分的には行わない方針に合わせるためである。
ディレクトリの一部だけを退避すると、退避の残りと新しい配置が入り混じり、何が起きたかを
利用者が把握しにくくなる。

## 棄却した代替案

### 祖先の symlink による conflict も退避の対象にする

祖先を退避すると、配下の複数の entry のネストの前提が同時に崩れうる。1 対 1 の rename では
表しきれない。

### 実ディレクトリの中で、移行できない leaf だけを退避する

退避した leaf と移行した leaf と新しい配置が 1 つのディレクトリに混在し、何が起きたかを
追いにくくなる。

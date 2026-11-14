---
id: "ADR-a56f1b79-1e49-4898-bedd-9020e8a42142"
type: adr
name: "除去の直前の再確認で起きた I/O の失敗は drift と区別して error にする"
description: "除去の直前の再確認で、不在と symlink でなくなったこと以外の I/O の失敗は drift と扱わず、元エラーを残したエラーで止める"
section: removal-migration
status: stable
sources:
  - id: "issue-677"
    resource: "https://github.com/yasunori0418/layat/issues/677"
    title: "#677 Go のエラー処理を規約に揃える"
origin: "Issue #677 の grilling（2026-11-14）。除去の直前の再確認が I/O の失敗を drift と読み替える欠陥（#785）への対処として決めた"
justifies:
  - "REQ-16aef46b-7bb8-4ca1-b962-e9f3ed1fd1d2"
references:
  - "ADR-2d4fdda4-b2e1-4161-8b71-3cff5b265f2f"
  - "ADR-64cff3f4-5997-49f4-b23a-e6ebeced7df9"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-15T01:16:41+09:00"
---
# 除去の直前の再確認で起きた I/O の失敗は drift と区別して error にする

- ステータス: stable
- 日付: 2026-11-15
- 関連: [配置前除去は drift を見つけたら error で止める](preremove-drift-errors.md), [エラーコードは共通コードを再利用し固有の意味だけを E_LAYAT_* にする](error-code-scheme.md), GitHub Issue #677, #785
- 起点: Issue #677 の grilling（2026-11-14）。除去の直前の再確認が I/O の失敗を drift と読み替える欠陥（#785）への対処として決めた

## 背景

古くなった配置の除去と配置前除去は、symlink を消す直前に、それがまだ記録どおりの先を指す symlink かを
確かめ直す。この再確認は `lstat` と `readlink` の失敗を、理由を問わず drift として扱っていた。
古くなった配置の除去は drift を警告して続け、apply は 0 で終わる。その target は last-applied 記録から
外れるので、以後の apply はその symlink を除去しない。配置前除去は drift で止まる
（[配置前除去は drift を見つけたら error で止める](preremove-drift-errors.md)）が、元のエラーが連鎖に無いので、
`--json` の code は原因を表さない。

計画は不在以外の `lstat` の失敗をすでにエラーにしている。再確認でこの失敗が起きるのは、計画から
実行までの間に権限が変わったときや、EIO のときに限られる。

## 決定

除去の直前の再確認で drift と扱うのは、target が無いとき、target が symlink でなくなったとき、
symlink が記録と違う先を指すときだけにする。symlink でなくなったことは、`lstat` の mode と、
`readlink` の EINVAL で判定する。それ以外の I/O の失敗は drift と扱わず、元のエラーを残したエラーで
止める。古くなった配置の除去と配置前除去のどちらも、そのエラーを entry の失敗として返す。

この規範は [古くなった配置の除去は前世代の記録通りを指す symlink のみに限り、copy は消さず孤児を警告する](../requirements/stale-removal-invariants.md) に置く。

## 根拠

drift は記録と実体が食い違っていると分かったことを表す。I/O の失敗は実体を調べられなかったことを
表すので、食い違いの有無は分からない。調べられなかったものを drift として残すと、記録から外れた
symlink が以後も除去されずに残り、apply は成功を報告する。

元のエラーを残せば、`--json` の code が原因から分類される
（[エラーコードは共通コードを再利用し固有の意味だけを E_LAYAT_* にする](error-code-scheme.md)）。権限の失敗は
`E_PERMISSION` になり、利用者は直す手がかりを得る。

## 棄却した代替案

### I/O の失敗も drift として警告し、続ける

今の動作である。調べられなかった symlink が記録から外れ、以後も除去されずに残る。apply が 0 で
終わるので、利用者は残ったことに気付けない。

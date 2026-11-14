---
id: "ADR-00cc3497-654f-420a-928f-99f201e7efb0"
type: adr
name: "絶対パスと .. エスケープは eval 時に拒否する"
section: lib-entries
status: stable
origin: "ADR-0019 §1 を書き直した"
justifies:
  - "REQ-6911eab6-12b4-457c-9db4-d7430a9e9b3f"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:27:41+09:00"
---
# 絶対パスと .. エスケープは eval 時に拒否する

- ステータス: stable
- 日付: 2026-11-01
- 関連: `docs/spec.md`
- 起点: ADR-0019 §1 を書き直した

## 背景

entry の target は root からの相対パス、subpath は src の中の相対パスである。ところが `/` で
始まる値や、`.claude/../../etc/x` のように `..` で root や src の外へ出る値を確認していなかった。
そのまま通ると、root の外へ書き込める。

## 決定

target と subpath が絶対パスである場合と、正規化すると `..` で root（target）や src（subpath）の
外へ出る場合を、`normalizeManifest` が評価時に拒否する。規範の細目は
[target / subpath のパス安全性を評価時に確認する](../requirements/path-safety-validation.md) に置く。

## 根拠

root の外への書き込みは、未定義の挙動のなかで最も危険である。絶対パスか、`..` で外へ出るかは
root の実際の値に依らず静的に判定できるので、実行時まで待つ理由が無い。評価時に確認すれば、
CLI から使ってもモジュールから使っても、不正なパスは engine に届く前に止まる。

## 棄却した代替案

### engine の実行時にだけ確認する

`--dryrun` や CI まで不正が見えず、評価時に拒否できる誤りの発見を遅らせる。

### 評価時と実行時の 2 段で確認する

最も堅牢だが、静的に判定できる絶対パスと `..` について engine が同じ確認を重ねるのは過剰である。

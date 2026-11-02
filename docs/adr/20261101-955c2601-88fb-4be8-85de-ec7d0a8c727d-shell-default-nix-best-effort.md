---
id: "ADR-955c2601-88fb-4be8-85de-ec7d0a8c727d"
type: adr
name: "shell.nix / default.nix の再現性は利用者の責任にする"
status: 採用
issues:
  - "#4"
origin: "Issue #4（flake.nix 以外の entrypoint）。ADR-0007 §5 を書き直した"
justifies:
  - "REQ-67095391-eab2-45d2-b75b-b428d481bcc2"
supersedes:
  - "ADR-0007"
---
# ADR-955c2601-88fb-4be8-85de-ec7d0a8c727d: shell.nix / default.nix の再現性は利用者の責任にする

- ステータス: 採用
- 日付: 2026-11-01
- 関連: GitHub Issue #4
- 起点: Issue #4（flake.nix 以外の entrypoint）。ADR-0007 §5 を書き直した

## 背景

CLI は flake.nix に加えて shell.nix / default.nix も entrypoint として読む（#4）。flake は flake.lock が
入力を固定し、pure eval になる。shell.nix / default.nix は NIX_PATH（channels）に依存する impure eval に
なり、layat の側には再現性を保証する手段が無い。

## 決定

shell.nix / default.nix の entrypoint は best-effort で扱い、impure eval を許容する。再現性は、
nixpkgs と layat lib を利用者が固定する責任とする。

flake と legacy の entrypoint の eval と再現性の区別は REQ-67095391-eab2-45d2-b75b-b428d481bcc2 に置く。

## 根拠

flake を使わない利用者も layat を使えるようにするには、legacy の entrypoint を読む必要がある。
その再現性まで layat が担おうとすると、入力を固定する機構を CLI が持つことになる。npins・
fetchTarball・flake-compat など固定の手段は既にあり、推奨として案内すれば足りる。

flake の側は、root を engine の実行時に解決するので、eval を pure のまま保てる。

## 棄却した代替案

### flake.nix だけを entrypoint にする

再現性は常に flake.lock で保てるが、flake 以外の entrypoint からも動かしたいという #4 の要求を
満たせない。

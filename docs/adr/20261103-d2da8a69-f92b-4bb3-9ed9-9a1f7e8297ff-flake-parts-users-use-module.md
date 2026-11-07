---
id: "ADR-d2da8a69-f92b-4bb3-9ed9-9a1f7e8297ff"
type: adr
name: "flake-parts の利用者には module 経由を正準にする"
status: 採用
issues:
  - "#26"
origin: "Issue #26。ADR-0029 §6a と ADR-0029 §6b を書き直した"
justifies:
  - "REQ-c50df875-2cb0-4e72-8a21-858359a11cae"
references:
  - "ADR-e6820da0-fce3-4aef-8b79-d53abbca6051"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T15:55:37+09:00"
---
# ADR-d2da8a69-f92b-4bb3-9ed9-9a1f7e8297ff: flake-parts の利用者には module 経由を正準にする

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-e6820da0-fce3-4aef-8b79-d53abbca6051, GitHub Issue #26
- 起点: Issue #26。ADR-0029 §6a と ADR-0029 §6b を書き直した

## 背景

flake-parts module（ADR-e6820da0-fce3-4aef-8b79-d53abbca6051）を公開すると、named manifest を
公開する書き方が 2 つになる。`perSystem.layat.<name>` に書く module 経由と、
`outputs.layat.<system>.<name>` に `mkManifest` の結果を直書きする形である。どちらも CLI からは
`nix build .#layat.<system>.<name>` で同じように見える。

## 決定

flake-parts を使う repo では module 経由を正準にする。直書きは、plain flake と
`shell.nix` / `default.nix` での正準として残す。`layat init` が展開するテンプレートは plain flake
のまま変えず、module 経由は既に flake-parts を使っている repo 向けの導線にとどめる。

両形の関係と CLI のアドレッシングの規範は REQ-c50df875-2cb0-4e72-8a21-858359a11cae に置く。

## 根拠

flake-parts を使う repo では、module 経由にすると `pkgs` が perSystem から来て `packages.layat` と
揃い、将来 overlay や config を足しても追従する。

flake-parts を使わない repo や `shell.nix` / `default.nix` には module を読み込む場所が無いので、直書きを
正準として残す必要がある。

テンプレートは最初の一歩の雛形なので、依存を増やさない最小の形に保つ。flake-parts は starter には
過剰で、使っている repo は自分の flake に module を import すれば足りる。

## 棄却した代替案

### 直書きを廃して module 経由に一本化する

書き方は 1 つになるが、plain flake の repo と `shell.nix` / `default.nix` では使えなくなる。

### テンプレートを flake-parts ベースにする

flake-parts を使う repo にはそのまま使える雛形になるが、使わない利用者にも flake-parts の依存と
書き方を持ち込む。starter としては過剰である。

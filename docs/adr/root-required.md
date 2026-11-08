---
id: "ADR-c7ffc21a-4be1-4468-b63c-2f47eedd9ab3"
type: adr
name: "root の必須性を別の確認を書かずに標準のエラーで担保する"
status: stable
origin: "manifest の型検査を NixOS モジュールシステムで行えるかという確認。ADR-0010 §3（クロスエントリ制約と必須性）の root 必須の項を書き直した"
justifies:
  - "REQ-d1b5b3f5-10a0-400d-9f03-ba00c63d1c34"
  - "REQ-4ec3accc-8bb6-461f-9024-dcf0027849e4"
references:
  - "ADR-560852e1-3d07-46b9-ae1a-d15f82d28a5e"
  - "ADR-f85f1afd-9b87-4f7a-922e-97c62554f198"
  - "ADR-400cadd4-0151-470b-92ae-9b11aa8e9bed"
  - "ADR-88231ffc-4836-4252-be34-894944fd28c3"
  - "ADR-86a77232-dc99-4c04-bd6d-2127410e8828"
  - "ADR-5caf1c14-a054-4c8d-95a2-c7f89e0c5246"
  - "ADR-e683ac05-c0e0-41c1-8b6a-38b605972fe1"
  - "ADR-28605860-770e-41df-95dd-7dc05b20497a"
  - "ADR-955c2601-88fb-4be8-85de-ec7d0a8c727d"
  - "ADR-65b3439a-b90a-4056-b55a-522399bc71e6"
  - "ADR-d15a216f-b5d6-4df9-9825-b23a341bec1d"
  - "ADR-553eea60-6ab4-4da7-93a9-3acf44b2fae3"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:30:50+09:00"
---
# ADR-c7ffc21a-4be1-4468-b63c-2f47eedd9ab3: root の必須性を別の確認を書かずに標準のエラーで担保する

- ステータス: stable
- 日付: 2026-11-01
- 関連: [project mode を第一の使い方とする](project-first.md), [root の明示を mkManifest 層で必須にする](root-explicit-at-mkmanifest.md), [root は 3 つの marker で表す](three-root-markers.md), [CLI が一次 UX で devShell に pin 版を同梱する](cli-primary-ux-devshell-pinned.md), [mkActivationScript で config ごとのラッパー derivation を提供する](mkactivationscript-wrapper.md), [内部コマンドの開示で透明性を保つ](internal-command-disclosure.md), [named manifest を専用の layat.<name> 名前空間に公開する](layat-name-namespace.md), [apply --all は config ごとの情報を 1 回の eval でまとめて取る](apply-all-single-eval.md), [shell.nix / default.nix の再現性は利用者の責任にする](shell-default-nix-best-effort.md), [apply の config は名前省略で default を選び、--all で全件を選ぶ](config-selection-default-and-all.md), [templates と layat init は固定 ref で導入する](templates-and-init-fixed-ref.md), [engine は cmd/ と internal/ に置きライブラリ化しない](engine-internal-cmd-layout.md), `docs/spec.md`
- 起点: manifest の型検査を NixOS モジュールシステムで行えるかという確認。ADR-0010 §3（クロスエントリ制約と必須性）の root 必須の項を書き直した

## 背景

`root` は暗黙の既定値を持たず、利用者が明示しなければならない（[root の明示を mkManifest 層で必須にする](root-explicit-at-mkmanifest.md)）。この必須性をどの仕組みで
確認するかを決める必要があった。

Nix の関数は、既定値の無い引数を渡されずに呼ばれると、評価を止める。モジュールシステムも、既定値の
無いオプションが設定されないまま使われると評価を止める。

## 決定

`root` には既定値を持たせず、`normalizeManifest` と `mkManifest` の必須の引数にする。省略されたときは
Nix の標準のエラーで止め、必須性のための確認を別に書かない。

必須性の規範は [root は明示必須で暗黙デフォルトを持たない](../requirements/root-explicit-required.md) に、`mkManifest` の確認の規範は
[mkManifest 自身が evalModules で入力を確認する単一の箇所になる](../requirements/mkmanifest-input-validation.md) に置く。

## 根拠

Nix の標準の挙動だけで、省略を評価時に止められる。同じことを手書きの確認で
重ねる理由が無い。

## 棄却した代替案

### `root` の省略を `lib.throwIf` で別に確認する

標準のエラーで既に止まる誤りを二重に確認することになる。

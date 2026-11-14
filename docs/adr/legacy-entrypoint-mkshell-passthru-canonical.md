---
id: "ADR-925b0011-e2a2-43ac-a5d3-5e4a10f52470"
type: adr
name: "legacy entrypoint は mkShell の passthru 形を正準にする"
section: cli-commands
status: stable
sources:
  - id: "issue-112"
    resource: "https://github.com/yasunori0418/layat/issues/112"
    title: "#112"
  - id: "issue-107"
    resource: "https://github.com/yasunori0418/layat/issues/107"
    title: "#107"
origin: "ADR-0032 §1 を書き直した"
justifies:
  - "REQ-c890ce4a-6528-4ab3-ac86-23d7aebff7da"
references:
  - "ADR-28605860-770e-41df-95dd-7dc05b20497a"
  - "ADR-6b7becf0-f029-41e3-895e-b672dce94d3d"
  - "ADR-c7ffc21a-4be1-4468-b63c-2f47eedd9ab3"
  - "ADR-e683ac05-c0e0-41c1-8b6a-38b605972fe1"
  - "ADR-f40c3d37-638d-420a-a0fe-3908cbf0aec0"
  - "ADR-fe035038-4415-4a1e-aa5c-509fdfce9955"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:24:30+09:00"
---
# legacy entrypoint は mkShell の passthru 形を正準にする

- ステータス: stable
- 日付: 2026-11-03
- 関連: [apply --all は config ごとの情報を 1 回の eval でまとめて取る](apply-all-single-eval.md), [CLI は cobra で組む](cobra-cli-framework.md), [root の必須性を別の確認を書かずに標準のエラーで担保する](root-required.md), [named manifest を専用の `layat.<name>` 名前空間に公開する](layat-name-namespace.md), [home mode の CLI と lib の版ずれを manifest の確認に任せて許容する](home-mode-skew-guarded-by-schemaversion.md), [1 config = 1 profile を全モードで自前に持つ](one-config-one-profile-all-modes.md), GitHub Issue #112, #107
- 起点: ADR-0032 §1 を書き直した

## 背景

CLI は flake.nix に加えて shell.nix / default.nix も entrypoint として読む。legacy entrypoint の書き方は
トップレベルの attrset `{ layat.<name> = mkManifest { ... }; }` とだけ想定されていた。project mode では、
同じ shell.nix を devShell と layat の entrypoint に兼ねるのが典型である。トップレベルの attrset を
返すと、その shell.nix は `nix-shell` でシェルに入れない。

## 決定

legacy entrypoint の正準形は、素の `mkShell` の derivation を返し、`passthru.layat.<name>` に named
manifest を載せる形にする。トップレベルの attrset を返す形も有効な代替形として残す。

legacy entrypoint の書き方と CLI が読む attr path は [legacy entrypoint は mkShell passthru 形を canonical とし CLI の attr path を分岐させない](../requirements/legacy-entrypoint-passthru-canonical.md) に置く。

## 根拠

`passthru` は derivation に情報を足すだけで、derivation の型や振る舞いを変えない。`nix-shell` は
返り値が derivation であればシェルに入れ、`passthru` は無視する。project の devShell と layat の
entrypoint を同じファイルに置ける。

トップレベルの attrset 形は、`mkShell` を使わず devShell も兼ねない最小の構成ではこちらのほうが
単純なので、廃止しない。

## 棄却した代替案

### トップレベルの attrset 形を正準にする

`mkShell` を返さない前提になり、project の devShell を兼ねる主要な使い方で `nix-shell` からシェルに入れなくなる。

### トップレベルの attrset 形を廃止する

書き方は 1 つに揃うが、devShell を持たない構成にまで `mkShell` を強いることになる。

---
id: "ADR-925b0011-e2a2-43ac-a5d3-5e4a10f52470"
type: adr
name: "legacy entrypoint は mkShell の passthru 形を正準にする"
status: 採用
issues:
  - "#112"
  - "#107"
origin: "ADR-0032 §1 を書き直した"
justifies:
  - "REQ-c890ce4a-6528-4ab3-ac86-23d7aebff7da"
---
# ADR-925b0011-e2a2-43ac-a5d3-5e4a10f52470: legacy entrypoint は mkShell の passthru 形を正準にする

- ステータス: 採用
- 日付: 2026-11-03
- 関連: GitHub Issue #112, #107
- 起点: ADR-0032 §1 を書き直した

## 背景

CLI は flake.nix に加えて shell.nix / default.nix も entrypoint として読む。legacy entrypoint の書き方は
トップレベルの attrset `{ layat.<name> = mkManifest { ... }; }` とだけ想定されていた。project mode では、
同じ shell.nix を devShell と layat の entrypoint に兼ねるのが典型である。トップレベルの attrset を
返すと、その shell.nix は `nix-shell` で入室できない。

## 決定

legacy entrypoint の正準形は、素の `mkShell` の derivation を返し、`passthru.layat.<name>` に named
manifest を載せる形にする。トップレベルの attrset を返す形も有効な代替形として残す。

legacy entrypoint の書き方と CLI が読む attr path は REQ-c890ce4a-6528-4ab3-ac86-23d7aebff7da に置く。

## 根拠

`passthru` は derivation に情報を足すだけで、derivation の型や振る舞いを変えない。`nix-shell` は
返り値が derivation であれば入室でき、`passthru` は無視する。project の devShell と layat の
entrypoint を同じファイルに置ける。

トップレベルの attrset 形は、`mkShell` を使わず devShell も兼ねない最小の構成ではこちらのほうが
単純なので、廃止しない。

## 棄却した代替案

### トップレベルの attrset 形を正準にする

`mkShell` を返さない前提になり、project の devShell を兼ねる主要な使い方で `nix-shell` が壊れる。

### トップレベルの attrset 形を廃止する

書き方は 1 つに揃うが、devShell を持たない構成にまで `mkShell` を強いることになる。

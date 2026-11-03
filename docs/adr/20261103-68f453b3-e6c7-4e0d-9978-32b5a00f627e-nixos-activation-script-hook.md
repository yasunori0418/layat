---
id: "ADR-68f453b3-e6c7-4e0d-9978-32b5a00f627e"
type: adr
name: "NixOS モジュールは system.activationScripts.layat から起動する"
status: 採用
issues:
  - "#519"
  - "#493"
origin: "Issue #493（epic: NixOS / nix-darwin モジュールの実装）の grilling（2026-10-24）。ADR-0066 §2a を書き直した"
justifies:
  - "REQ-c1b3ca5f-d2f7-443c-bc4b-b18413ca97b9"
references:
  - "ADR-0066"
---
# ADR-68f453b3-e6c7-4e0d-9978-32b5a00f627e: NixOS モジュールは system.activationScripts.layat から起動する

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-0066, GitHub Issue #519, #493
- 起点: Issue #493（epic: NixOS / nix-darwin モジュールの実装）の grilling（2026-10-24）。ADR-0066 §2a を書き直した

## 背景

NixOS モジュールは host の activation から engine を起動する。NixOS は activation の処理を名前付きの
スクリプトとして並べ、`deps` で順序を付ける。engine は system mode で配置するので、`/etc` やユーザーが
用意されてから動く必要がある。`nixos-rebuild dry-activate` でも、変更の予定を示せることが望ましい。

## 決定

NixOS モジュールは `system.activationScripts.layat` を定義し、そこから engine を起動する。
`deps` には `"etc"` と `"users"` を入れ、`supportsDryActivation = true` とする。

フック名と起動の規範は REQ-c1b3ca5f-d2f7-443c-bc4b-b18413ca97b9 に置く。

## 根拠

layat 専用の名前でスクリプトを持てば、他のスクリプトと独立に順序を付けられ、利用者の設定とも衝突しない。
`deps` に `etc` と `users` を入れることで、配置先のディレクトリやファイルの所有者が揃ったあとに engine が動く。

`supportsDryActivation = true` が無いと、スクリプトは dry-activate 用の activate に含まれず、
dry-activate のときに何も示せない。

## 棄却した代替案

### `postActivation` のような既存の固定名スクリプトへ追記する

NixOS では名前付きのスクリプトを足せるので、固定名に寄せる理由が無い。追記にすると `deps` で順序を
付けられない。

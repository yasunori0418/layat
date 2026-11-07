---
id: "TC-1b492e41-ee81-4945-a42b-0669dd574809"
type: test_condition
name: "NixOS の実 activate で system mode の配置・世代・記録が host の構成に収束する"
status: stable
mitigates:
  - "RISK-70fc1db8-56ec-457d-af12-553d80a04d7d"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:27:22+09:00"
---
# TC-1b492e41-ee81-4945-a42b-0669dd574809: NixOS の実 activate で system mode の配置・世代・記録が host の構成に収束する

## 条件

NixOS モジュールを組み込んだ host を VM で起動し、構成の切り替えを実際に activate して、
その後のファイルシステムに対してアサートする。評価時の配線だけでは観測できない、
activate 後の結果を確かめる。

- **実配置** — 構成の switch で `/etc` 配下の target が store の配置元を指す symlink になり、
  system 基底に profile の世代と last-applied の記録が立つこと
- **世代を積まない再 activate** — 同じ構成で activation をもう一度実行しても世代が増えないこと
  （NixOS は boot のたびに activation を実行するため、ここが崩れると世代が増え続ける）
- **構成の切り替え** — entry の違う構成へ switch すると target が入れ替わり、旧 target が
  古くなった配置として除去され、世代が 1 件増えること
- **収束** — 元の構成へ switch し直すと配置がその構成に戻ること

初回 boot の activation は nix DB のロードより先に実行されるため、boot 時の結果はこの条件の外に置く。

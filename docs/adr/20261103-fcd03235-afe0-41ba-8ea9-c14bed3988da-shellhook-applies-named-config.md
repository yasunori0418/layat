---
id: "ADR-fcd03235-afe0-41ba-8ea9-c14bed3988da"
type: adr
name: "devShell の shellHook は config を名指しして apply する"
status: 採用
origin: "ADR-0018 §3a を書き直した"
justifies:
  - "REQ-196ddabf-6569-4303-942e-050872972501"
---
# ADR-fcd03235-afe0-41ba-8ea9-c14bed3988da: devShell の shellHook は config を名指しして apply する

- ステータス: 採用
- 日付: 2026-11-03
- 起点: ADR-0018 §3a を書き直した

## 背景

project template は devShell を持ち、入室時の shellHook で config を配置する。shellHook から呼ぶ
`layat apply` を、config を名指しする形にするか、`--all` で一括する形にするかは決まっていなかった。

## 決定

project template の shellHook は `layat apply <name> --no-wait` と config を名指しする形を既定にする。
複数の config を一括で置く `layat apply --all --project-root --no-wait` はコメントで示すに留める。

shellHook を含む project template の構成は REQ-196ddabf-6569-4303-942e-050872972501 に置く。

## 根拠

template の example は 1 config なので、名指しが最も明示的で、何が配置されるかが読めばわかる。
`--all` は root の種類で絞らないと、同じ entrypoint にある別の mode の config まで拾う。名指しを
既定にすれば、その取り違えが template の既定から起きない。

`--no-wait` は、並行して入室したときにロック待ちで shell が止まらないようにする。

## 棄却した代替案

### `--all` を shellHook の既定にする

config を足しても shellHook を書き換えずに済むが、`--project-root` を付け忘れると別の mode の
config を配置しうる。example が 1 config の template では、一括の利点も生きない。

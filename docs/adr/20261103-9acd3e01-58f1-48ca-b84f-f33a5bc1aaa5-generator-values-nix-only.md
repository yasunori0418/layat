---
id: "ADR-9acd3e01-58f1-48ca-b84f-f33a5bc1aaa5"
type: adr
name: "生成器の選択値を nix だけにする"
status: stable
sources:
  - id: "issue-395"
    resource: "https://github.com/yasunori0418/layat/issues/395"
    title: "#395"
origin: "Issue #395（epic: manifest 生成器のインターフェイス化）の grilling（2026-09-09）。ADR-0056 §2 を書き直した"
justifies:
  - "REQ-badc7e10-0ba4-40d4-b334-6e40193119db"
references:
  - "ADR-36cb1e78-90c0-47d2-bbb1-4bfec5db2964"
  - "ADR-1ddf424a-db2c-4bcb-8823-bf7e4326b980"
  - "ADR-ef9e4099-0cf4-409a-9e73-518866aea676"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# ADR-9acd3e01-58f1-48ca-b84f-f33a5bc1aaa5: 生成器の選択値を nix だけにする

- ステータス: stable
- 日付: 2026-11-03
- 関連: ADR-36cb1e78-90c0-47d2-bbb1-4bfec5db2964, ADR-1ddf424a-db2c-4bcb-8823-bf7e4326b980, ADR-ef9e4099-0cf4-409a-9e73-518866aea676, GitHub Issue #395
- 起点: Issue #395（epic: manifest 生成器のインターフェイス化）の grilling（2026-09-09）。ADR-0056 §2 を書き直した

## 背景

生成器は明示の指定で選ぶ（→ ADR-36cb1e78-90c0-47d2-bbb1-4bfec5db2964）。実装する生成器は nix・prebuilt・
テストダブルで、prebuilt は `--manifest` に写る（→ ADR-1ddf424a-db2c-4bcb-8823-bf7e4326b980）。HM などの
モジュールの activation は `apply --manifest` で engine を起動し、その実行環境の環境変数やカレント
ディレクトリは利用者の制御外にある。

## 決定

生成器として選べる値は `nix` だけにし、2 つ目の実生成器が現れたときに足す。prebuilt は `--manifest`
でだけ選ばれ、`--generator` の値には出さない。`--generator` と `--manifest` の同時指定はエラーにする。
`--manifest` のときは、環境変数と設定ファイルを読まない。

選択の規範は REQ-badc7e10-0ba4-40d4-b334-6e40193119db に置く。

## 根拠

prebuilt を値に出すと、`--manifest` のパス指定と選び方が二重になり、片方だけを指定したときの意味が
定まらない。`--manifest` のときに環境変数や設定ファイルを読むと、activation が利用者の手元の設定に
左右される。

## 棄却した代替案

### prebuilt を選択値に出す

`--generator prebuilt` だけでは link-farm が決まらず、`--manifest` だけのときとの関係も定まらない。

### `--manifest` のときも環境変数と設定ファイルを読む

activation の結果が、実行環境に残った環境変数や cwd にある設定ファイルで変わる。

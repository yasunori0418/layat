---
id: "ADR-0061"
type: adr
name: "モジュールに `layat.recopy` オプションを足し、`layat.jobs` とともに共通オプションへ載せる"
status: 採用
issues:
  - "#128"
  - "#471"
origin: "Issue #128（HM 複数 profile・ADR-0035 の実装）の着手前 grilling（2026-10-15）"
justifies:
  - "REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10"
references:
  - "ADR-0020"
  - "ADR-0045"
  - "ADR-0060"
---
# ADR-0061: モジュールに `layat.recopy` オプションを足し、`layat.jobs` とともに共通オプションへ載せる

- ステータス: 採用
- 日付: 2026-10-15
- 関連: ADR-0020, ADR-0045, ADR-0060, GitHub Issue #128, #471
- 起点: Issue #128（HM 複数 profile・ADR-0035 の実装）の着手前 grilling（2026-10-15）

## 背景

`apply` の修飾フラグのうち、モジュール経由で使えるのは `--backup`（`layat.backup`・→ ADR-0045 §8）だけで、`--recopy`（→ ADR-0020 §1）は HM activation から使えなかった。copy entry を src の更新へ追従させるには、利用者が activation の外で `layat apply --manifest … --recopy` を手で叩くしかない。

また ADR-0060 で HM activation は `apply --all` の 1 回起動になり、`--jobs` を常に付ける。その値を利用者が指定する口が要る。

## 決定

### 1. `layat.recopy :: bool`（既定 false）を共通オプションに足す

- `true` のとき、activation の `apply --all` 起動へ `--recopy` を付ける。`false` のときは付けない。
- 有効にすると activation のたびに全 copy target を src から上書きし、ローカルの編集を捨てる。オプションの説明文にこれを明記する。
- `layat.backup` と同じく起動配線レイヤーのオプションで、manifest（entries・manifest v1 契約）には影響しない。entry 単位や config 単位の上書きは持たず、1 回の起動に載る全 config へ一律に効く。

### 2. `layat.jobs :: ints.unsigned`（既定 0）を共通オプションに足す

- 値は activation の `apply --all` 起動へ `--jobs <n>` として渡す（付け方は ADR-0060 §1）。0 は論理 CPU 数を意味し、`--jobs` の解決規則に従う。
- 型を `ints.unsigned` にして、`--jobs` が拒否する負数を module eval の型検査で弾く。

### 3. config 間の差を持たせない

- `layat.backup` / `layat.recopy` / `layat.jobs` はいずれも global とし、`configs.<name>` ごとの上書きは持たない。activation は 1 回の起動で、修飾フラグは起動全体に効くため。

## 根拠

- **bool にする理由**: `--recopy` は値を取らないフラグで、`layat.backup` のような submodule にする中身が無い。汎用のパススルー（`extraFlags` 等）は型検査が効かず、フラグの意味を module 側の説明文で伝えられないため採らない。
- **既定 false の理由**: recopy はローカル編集を破棄する破壊的な修飾で、CLI でも明示の opt-in（→ ADR-0020 §1）。activation のたびに黙って上書きする既定は採れない。
- **jobs を明示オプションにする理由**: `apply --all` の並列度は CLI の既定（論理 CPU 数）で十分な場面が多いが、switch 時の負荷を抑えたい利用者が下げられる口が要る。

## 影響

- **`modules/common.nix`**: `layat.recopy` / `layat.jobs` を追加する。
- **`modules/home-manager.nix`**: activation の `apply --all` 起動へ `--recopy` を条件付きで、`--jobs <n>` を常に付ける。
- **要求**: 共通オプション集合（REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10）を改訂し、recopy の配線レイヤーの要求を起票する。

## 棄却した代替案

- **`layat.extraFlags :: listOf str` で任意のフラグを渡す**: 型検査も説明文も効かず、`--manifest` や `--all` のような配線側が持つフラグとの衝突も防げない。
- **`configs.<name>.recopy` で config ごとに切り替える**: activation は 1 回の `apply --all` 起動で、`--recopy` は起動全体の修飾。config 単位にするには CLI 側に config ごとの修飾を足す必要があり、要求が出ていない。
- **`layat.jobs` を持たず CLI の既定に任せる**: 並列度を下げたい利用者が activation に介入する手段が無くなる。

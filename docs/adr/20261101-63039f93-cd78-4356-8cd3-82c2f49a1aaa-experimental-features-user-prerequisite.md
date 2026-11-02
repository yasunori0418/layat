---
id: "ADR-63039f93-cd78-4356-8cd3-82c2f49a1aaa"
type: adr
name: "nix の experimental-features は利用者が有効にしておく前提にする"
status: 採用
origin: "/grill-me による実装前の横断検査（実装前残セマンティクスの第 7 巡）。ADR-0025 §1 を書き直した"
justifies:
  - "REQ-f9920c87-8551-4aa3-bf03-26fdf4191ed6"
references:
  - "ADR-0055"
---
# ADR-63039f93-cd78-4356-8cd3-82c2f49a1aaa: nix の experimental-features は利用者が有効にしておく前提にする

- ステータス: 採用
- 日付: 2026-11-01
- 関連: ADR-0055
- 起点: /grill-me による実装前の横断検査（実装前残セマンティクスの第 7 巡）。ADR-0025 §1 を書き直した

## 背景

既定の manifest 生成器である nix 生成器は、内部で `nix eval` / `nix build`（新 CLI）を使う。これらは
`experimental-features = nix-command`（flake の entrypoint ではさらに `flakes`）が無効だと動かず、
有効化されていない環境では分かりにくい nix のエラーで落ちる。

## 決定

`nix-command`（flake の entrypoint ではさらに `flakes`）の experimental-features は、利用者の環境で
有効化済みであることを前提にする。layat は `--extra-experimental-features` を自動で付けない。有効化されていないときは、前提条件と
有効化の方法を案内するエラーで止まる。判定と案内文を nix 生成器に閉じることは ADR-0055 が決める。

前提条件とエラーの規範は REQ-f9920c87-8551-4aa3-bf03-26fdf4191ed6 に置く。

## 根拠

experimental-features は nix.conf・Determinate Nix・組織のポリシーで管理されていることがある。layat が
フラグを自動で付けると、その設定を利用者に知らせずに上書きし、環境の違いを隠してしまう。余計な前提や state を
持ち込まないという layat の方針にも反する。前提を明示して、案内付きのエラーで止まるほうが透明である。

## 棄却した代替案

### `--extra-experimental-features` を自動で付ける

未設定の環境でも追加の手間なく動くが、利用者や組織による experimental-features の管理を知らせずに
上書きし、設定の違いを隠す。

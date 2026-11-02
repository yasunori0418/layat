---
id: "ADR-40dd7945-5792-4b80-bdfa-43f95f482554"
type: adr
name: "flake check の警告は無害として許容する"
status: 採用
origin: "ドキュメント計画段階の最終レビュー（grill）。ADR-0015 §8 を書き直した"
justifies:
  - "REQ-d0aef5af-e922-400b-b250-ca38719c480b"
supersedes:
  - "ADR-0015"
---
# ADR-40dd7945-5792-4b80-bdfa-43f95f482554: flake check の警告は無害として許容する

- ステータス: 採用
- 日付: 2026-11-01
- 起点: ドキュメント計画段階の最終レビュー（grill）。ADR-0015 §8 を書き直した

## 背景

利用者の flake は named manifest を `outputs.layat.<system>.<name>` に公開する。`layat` は nix 本体が
知らない output なので、`nix flake check` は `warning: unknown flake output 'layat'` を出す。

nix 2.27 / 2.34 で実際に動かした結果と、nix のソース（`src/nix/flake.cc` の `CmdFlakeCheck`）から、
次のことを確かめた。

- 終了コードは 0 で、CI は落ちない
- `layat` 直下の値は評価されるが、未知の output の配下までは再帰して調べない。配下の derivation は build も
  eval もされない
- flake-parts を経由しても警告は消えない。消す仕組みである flake-schemas（nix PR #8892）は未マージで、
  Determinate Nix でしか使えない
- `lib` 以外のどの名前にしても unknown の警告が出る

## 決定

`layat` output に対する `nix flake check` の unknown 警告は、想定内で無害なものとして許容し、
docs に明記する。`layat` の成果物の主な検証は `nix build .#layat.<system>.<name>` で行う。

許容の条件と主検証の方法は REQ-d0aef5af-e922-400b-b250-ca38719c480b に置く。

## 根拠

警告は CI を壊さず、配下の derivation を誤って build することもない。`layat` 直下の評価が成功するかは
むしろ確かめられる。警告を消す手段はいずれも、upstream の安定版では使えないか、特定の
ディストリビューションへの依存を招くので、許容するのがいちばん安い。flake-schemas が upstream にマージされたら、
`schemas.layat` を提供して警告を消す余地は残る。

## 棄却した代替案

### flake-schemas で警告を消す

nix PR #8892 は未マージで、Determinate Nix でしか使えない。upstream の nix の利用者を切り捨てる。

### 別の output 構造へ逃げる

`lib` 以外の名前では警告は消えず、`packages` を流用すると manifest を誤って build する危険がある。

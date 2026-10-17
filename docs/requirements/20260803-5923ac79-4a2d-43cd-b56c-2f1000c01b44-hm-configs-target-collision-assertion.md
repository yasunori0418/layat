---
id: "REQ-5923ac79-4a2d-43cd-b56c-2f1000c01b44"
type: requirement
name: "単一 HM config 内の configs 間 target 衝突は activation の apply --all 前段検査で停止する"
derives_from:
  - "UC-1c280dce-7c72-44c0-95ea-d06344f62a47"
specification: |
  Where `configs.<A>` and `configs.<B>` of a single home-manager configuration duplicate a
  normalized target — the string resolved after the attribute-key default and any explicit
  `target` override — the activation SHALL be stopped by the preflight of the `apply --all`
  it invokes (REQ-38506419-c23d-4fce-b167-30715ce6a69d), before any config is built or
  placed, so that the activation exits non-zero and the switch stops. The module
  evaluation SHALL NOT detect this collision. This is possible because every config of one
  module configuration rides on that single `apply --all` invocation. It SHALL be an
  exception to the general cross-config collision, which does not ride on a single
  invocation and which REQ-5c6b07da-3d06-414d-8770-4f438234b322 and
  REQ-fc1118b1-b0e8-4ddf-80f6-c70956651693 leave to last-writer-wins plus a foreign symlink
  warning at engine runtime; how a collision across separate entrypoints, machines or
  tools is treated is stated by those two items and is not restated or altered here.
  Without this detection, the two configs would take the same target from each other at
  every activation, making a flip-flop permanent, the introduction of role separation
  itself being what breeds the collision.
specification_ja: |
  単一の home-manager config 内の `configs.<A>` と `configs.<B>` が正規化後 target
  （属性キー既定値・明示 `target` 上書きを解決した後の文字列）を重複させた場合、activation が
  起動する `apply --all` の前段検査（REQ-38506419-c23d-4fce-b167-30715ce6a69d）で、どの
  config の build・配置にも入る前に停止させなければならない。activation は非ゼロで終わり
  switch が止まる。モジュール eval はこの衝突を検出してはならない。1 つのモジュール config の
  全 config が単一の `apply --all` 起動に載るため検出が可能である。これは単一の起動に載らない
  一般の cross-config 衝突（REQ-5c6b07da-3d06-414d-8770-4f438234b322 /
  REQ-fc1118b1-b0e8-4ddf-80f6-c70956651693 が engine 実行時の後勝ち + foreign symlink warning に
  委ねるもの）に対する例外でなければならない。別 entrypoint・別マシン・別ツールに跨る衝突を
  どう扱うかは同 2 item の担当であり、本 item では再掲も変更もしない。
  検出しなければ、activation のたびに A と B が同じ target を交互に奪い合う
  フリップフロップが恒常化する。役割分離の導入自体が衝突の温床になるためである。
---
# REQ-5923ac79-4a2d-43cd-b56c-2f1000c01b44: 単一 HM config 内の configs 間 target 衝突は activation の apply --all 前段検査で停止する

## 仕様

- 同一 HM config 内の `configs.<A>` と `configs.<B>` が**正規化後 target**を重複させた場合、
  activation の `apply --all` 前段検査で、build・配置の前に停止する。activation は非ゼロで
  終わり switch が止まる
- **モジュール eval では検出しない**（nix の `assertions` を持たない）
- 一般の cross-config 衝突（別 entrypoint・別マシン・別ツール）が「engine 実行時の
  後勝ち + foreign symlink warning」であることは**不変**
- 検出しなければ、activation のたびに A と B が同じ target を交互に奪い合うフリップフロップが
  恒常化する

> **REQ-5c6b07da-3d06-414d-8770-4f438234b322 / REQ-fc1118b1-b0e8-4ddf-80f6-c70956651693 との整合**: 両 item は規範文を「単一の起動に載らない
> 場合」へ限定し、単一の起動に載る複数 config については「配置前の検出を妨げない」までを
> 述べる。本 item の検出はその場合にあたり、activation の `apply --all` 前段検査が配置の前に
> 止めるので、両 item が engine 実行時に委ねる範囲の外にある。両 item が何を規定するかはそれぞれの specification が正で、
> ここでは再掲しない。
>
> **REQ-c5dfcae6-6094-4850-99e5-bf14530bc60a との整合**: 同 item が Nix 評価時に検出するのは、
> `mkManifest` が 1 つの manifest の設定だけで判定できる誤りである。configs 間の衝突は複数の
> manifest に跨るため `mkManifest` では判定できず、CLI の前段検査が engine の起動前に止める。
> engine を起動させないという層分けの趣旨は保たれる。
>
> **他 item との担当分界**: 正規化後 target 文字列の同値判定と属性キー = target という
> 識別の体系は REQ-cb77ea05-bab8-4ccf-b09e-d23d8f71cdc7 / REQ-b232ec98-af3b-41f3-a050-29d417322002、`layat.configs` オプションの定義は
> REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10、`<name>` 次元と役割分離そのものは REQ-c6891aeb-13c0-4ae7-9ad1-5c343735266a、foreign symlink warning
> と単発の後勝ちは REQ-622787dc-4512-4ce9-9c7d-7b32bbb70557、前段検査のバケット規則と
> エラーの形は REQ-38506419-c23d-4fce-b167-30715ce6a69d の担当。

## 出典

ADR-0060「HM activation を prebuilt manifest 群への `apply --all` 1 回起動にし、configs 間の
衝突検出を CLI 前段へ委譲する」§3。旧版（module eval の assertion で停止する形）の出典は
ADR-0035 §4 で、ADR-0060 が改訂した。

`docs/spec.md` には対応記述が無いため、原文の写しは持たない（規範は frontmatter が正で、
上の箇条書きは ADR 本文の要約）。cross-config を engine 実行時の後勝ち + foreign symlink
warning とするのは ADR-0015、`apply --all` の前段検査は ADR-0038 が定めている。

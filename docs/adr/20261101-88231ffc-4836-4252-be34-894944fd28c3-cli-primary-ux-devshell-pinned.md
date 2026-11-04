---
id: "ADR-88231ffc-4836-4252-be34-894944fd28c3"
type: adr
name: "CLI が一次 UX で devShell に pin 版を同梱する"
status: 採用
issues:
  - "#3"
origin: "Issue #3（CLI 化）と、ドキュメント計画段階の最終レビュー（grill）。ADR-0007 §3・ADR-0015 §1 を書き直した"
justifies:
  - "REQ-14f0aec9-abae-4621-82f3-40536a1ad904"
---
# ADR-88231ffc-4836-4252-be34-894944fd28c3: CLI が一次 UX で devShell に pin 版を同梱する

- ステータス: 採用
- 日付: 2026-11-01
- 関連: GitHub Issue #3
- 起点: Issue #3（CLI 化）と、ドキュメント計画段階の最終レビュー（grill）。ADR-0007 §3・ADR-0015 §1 を書き直した

## 背景

配置を行う engine は固定のバイナリで、利用者はそれを何らかのコマンドから起動する。config ごとに起動用の
ラッパーを用意させる形では、ラッパーを作らないと layat を使えず、flake 以外の entrypoint にも
対応しにくい（#3）。

CLI を PATH に置くと、今度は CLI の版と、利用者の flake が pin する `layat.lib` の版が別々に決まる。
両者がずれると、lib が生成する manifest の `schemaVersion` と CLI が受け付ける版が食い違う。

## 決定

`layat` CLI（`packages.layat`）を PATH に常駐する一次 UX にする。project mode の正規の導入方法は、
`templates/project` の devShell に flake.lock で pin した `layat` を同梱することとする。グローバルな
導入（`nix profile install`）は CLI 単体で使うときの利便として残す。

導入形態ごとの規範は REQ-14f0aec9-abae-4621-82f3-40536a1ad904 に置く。

## 根拠

PATH 上の CLI があれば、利用者は `layat apply <name>` で任意の config を適用でき、config ごとの
起動手段を持たなくてよい。

devShell に同梱すれば、CLI と `layat.lib` が同じ flake 入力から来るので `schemaVersion` が必ず一致する。
flake.lock による再現性が CLI まで届き、project-first の方針とも合う。CLI 単体で使うときは PATH 常駐の
利便を優先し、同梱を強制しない。

## 棄却した代替案

### config ごとのラッパーを主な起動手段にする

`nix run .#<config>` で起動するラッパーを主な起動手段にする案。ラッパーを作らないと engine を起動できず、
flake 以外の entrypoint や CLI 単体での使い勝手の要求を満たせない。

### project mode でもグローバルな導入を前提にする

導入は 1 回で済むが、flake.lock の pin を CLI の側で破り、`schemaVersion` のずれを招く。

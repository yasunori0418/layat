---
id: "ADR-6d2327f7-f513-42c3-89f8-808467db1914"
type: adr
name: "out-of-store はマッピングだけを追跡する"
status: stable
origin: "ADR-0002 の決定「out-of-store は世代追跡する」を書き直した"
justifies:
  - "REQ-a8a923ad-07fb-4582-b90a-07a6e0c41baa"
references:
  - "ADR-5e46ebd6-e44b-48dc-a93f-a8a3a3903967"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:24:30+09:00"
---
# ADR-6d2327f7-f513-42c3-89f8-808467db1914: out-of-store はマッピングだけを追跡する

- ステータス: stable
- 日付: 2026-11-01
- 関連: [copy と out-of-store の併用を eval 時に拒否する](copy-with-out-of-store-eval-error.md), `docs/spec.md`
- 起点: ADR-0002 の決定「out-of-store は世代追跡する」を書き直した

## 背景

out-of-store symlink は、store ではなく作業ツリーやホームの中のパスを指す symlink である。
指す先の内容は利用者がその場で編集し、変更がすぐ反映されることを期待して選ばれる。世代管理が
何を記録し、何を記録しないかを決める必要があった。

## 決定

out-of-store symlink は世代で追跡する。ただし版管理するのは target からリンク先へのマッピング
だけで、リンク先の内容はスナップショットしない。リンク先は絶対パスでも、配置 root に対する
相対パス（[root 相対マーカーのリンクを target 起点の相対パスで張る](root-relative-link-as-relative-path.md)）でもよい。規範の細目は [out-of-store symlink は marker の絶対パスまたは root 相対パスを指し、版管理はリンク先マッピングのみとする](../requirements/out-of-store-link-mapping-only.md) に置く。

## 根拠

マッピングを世代に載せれば、entry の増減や張り先の変更は symlink と同じく古くなった配置の除去と rollback
の対象になる。リンク先の内容をライブのまま保つことが out-of-store を選ぶ理由なので、内容を
固定すればその利点が失われる。

## 棄却した代替案

### リンク先の内容もスナップショットする

rollback で内容まで戻せるが、store の外にある任意のファイルを複製して保持することになる。
ライブ編集という目的とも矛盾する。

### out-of-store を世代の外に置く

copy と同じ扱いにする案。マッピングの変更が古くなった配置の除去の対象から外れ、古いリンクが残る。

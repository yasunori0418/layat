---
id: "SDR-29ef8bb4-9349-47db-9e79-926bc71bbdd0"
type: sara_adr
name: "infrastructure を技術基盤として再定義し root から外す"
status: stable
sources:
  - id: "issue-583"
    resource: "https://github.com/yasunori0418/layat/issues/583"
    title: "#583"
  - id: "issue-237"
    resource: "https://github.com/yasunori0418/layat/issues/237"
    title: "#237"
  - id: "issue-203"
    resource: "https://github.com/yasunori0418/layat/issues/203"
    title: "#203"
origin: "epic #203 の grilling セッション（2026-08-08）で確定。ADR-0049 §3a を書き直した"
justifies:
  - "INF-659b139d-0cf8-4c65-b30d-93c5ee2dfc71"
references:
  - "SDR-a0095ee8-570c-4f77-a6c9-bbc162a5c4cb"
  - "SDR-34916a83-0cff-4524-a8ff-5a3195af52c3"
  - "SDR-2457aa00-fab7-4c99-af53-2e6643c21c79"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T22:44:41+09:00"
---
# infrastructure を技術基盤として再定義し root から外す

- ステータス: stable
- 日付: 2026-11-03
- 関連: [root は solution と ADR 系の型だけにする](roots-solution-and-adr-types-only.md), [テスト計画の型を品質方針の型から分けて設ける](quality-and-test-plan-types.md), [model.yaml で組み込みモデルを全面置換し、汎用の型名を使う](model-yaml-full-replacement.md), GitHub Issue #583, #237, #203, `docs/model.yaml`
- 起点: epic #203 の grilling セッション（2026-08-08）で確定。ADR-0049 §3a を書き直した

## 背景

infrastructure は root の型として定義されていた。satisfy できる相手が design しか無く、CI やトレーサビリティ
検証のような開発の基盤は design の木に属さないため、root にせざるを得なかった。その結果、
`docs/infrastructure/` の item は全件が構造上の接続漏れだった。

また、infrastructure には品質系の文書（DoD・トレーサビリティ検証）と基盤系の文書（CI/CD・配信）が同居
していた。

## 決定

infrastructure を「プロダクトを開発・提供・稼働させる技術基盤」と定義する。CI/CD・リリース・開発環境に
加え、配信・ホスティング・クラウドを含む。

infrastructure の親は quality と design にし、`satisfies` で張る。開発の基盤は quality を、稼働の基盤は
design を満たす。これで infrastructure は root から外れる。

## 根拠

quality を設けた（[テスト計画の型を品質方針の型から分けて設ける](quality-and-test-plan-types.md)）ことで、開発の基盤にも接続先ができた。
quality と design のどちらかを経由して solution の木へ届くので、root の型を solution と ADR 系に絞れる
（[root は solution と ADR 系の型だけにする](roots-solution-and-adr-types-only.md)）。

品質系の文書が quality へ移ったため、infrastructure は技術基盤だけを受ける型として定義し直せる。
配信やクラウドまで含めるのは、web アプリケーションのような横展開でも同じ型で受けられるようにするため。

## 棄却した代替案

### infrastructure を root のまま残す

`docs/infrastructure/` の item が構造上の接続漏れであり続ける。warning に常在の項目が 1 件でもあると、
本物の張り忘れが埋もれる。

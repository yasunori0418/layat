---
id: "ADR-08a6e1be-e424-4cfe-a529-e123a50cd6ca"
type: adr
name: "CLI が評価する attr path を legacy entrypoint の書き方で分岐させない"
status: stable
sources:
  - id: "issue-112"
    resource: "https://github.com/yasunori0418/layat/issues/112"
    title: "#112"
  - id: "issue-107"
    resource: "https://github.com/yasunori0418/layat/issues/107"
    title: "#107"
origin: "ADR-0032 §2 を書き直した"
justifies:
  - "REQ-c890ce4a-6528-4ab3-ac86-23d7aebff7da"
  - "REQ-496b1a07-5b74-416b-9e5f-3952b4c03737"
  - "DSG-92f54490-872a-42ac-bbd7-d06e9ee381c6"
references:
  - "ADR-28605860-770e-41df-95dd-7dc05b20497a"
  - "ADR-6b7becf0-f029-41e3-895e-b672dce94d3d"
  - "ADR-c7ffc21a-4be1-4468-b63c-2f47eedd9ab3"
  - "ADR-e683ac05-c0e0-41c1-8b6a-38b605972fe1"
  - "ADR-f40c3d37-638d-420a-a0fe-3908cbf0aec0"
  - "ADR-fe035038-4415-4a1e-aa5c-509fdfce9955"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-05T00:29:04+09:00"
---
# ADR-08a6e1be-e424-4cfe-a529-e123a50cd6ca: CLI が評価する attr path を legacy entrypoint の書き方で分岐させない

- ステータス: stable
- 日付: 2026-11-03
- 関連: [apply --all は config ごとの情報を 1 回の eval でまとめて取る](apply-all-single-eval.md), [CLI は cobra で組む](cobra-cli-framework.md), [root の必須性を別の確認を書かずに標準のエラーで担保する](root-required.md), [named manifest を専用の `layat.<name>` 名前空間に公開する](layat-name-namespace.md), [home mode の CLI と lib の版ずれを manifest の確認に任せて許容する](home-mode-skew-guarded-by-schemaversion.md), [1 config = 1 profile を全モードで自前に持つ](one-config-one-profile-all-modes.md), GitHub Issue #112, #107
- 起点: ADR-0032 §2 を書き直した

## 背景

legacy entrypoint には、`mkShell` の `passthru` に named manifest を載せる正準形と、トップレベルの
attrset を返す代替形の 2 つの書き方がある。CLI がこの 2 つを見分けて attr path を変えるかどうかを
決める必要があった。

## 決定

CLI は entrypoint の種別（flake / legacy）でだけ attr path を変え、legacy の 2 つの書き方は区別
しない。どちらの書き方でも、CLI が評価する attr path は `layat.<name>` になる。

種別ごとの attr path は [legacy entrypoint は mkShell passthru 形を canonical とし CLI の attr path を分岐させない](../requirements/legacy-entrypoint-passthru-canonical.md) と
[entrypoint は `layat.<name>` に named manifest を公開し CLI は形ごとの attr path で build する](../requirements/named-manifest-addressing.md) に、分岐を attr path の組み立てに閉じる設計は
[legacy entrypoint の分岐は attr path 組み立てに閉じ、nix 呼び出しヘルパを共通で再利用する](../design/legacy-nix-invocation-reuse.md) に置く。

## 根拠

derivation は attrset で、`passthru` の各フィールドはその attrset にマージされる。Nix の属性選択は
`passthru` にもトップレベルの属性と同じように届くので、2 つの書き方で同じ attr path が通る。

CLI に書き方を判定する処理を持ち込む必要が無く、利用者がどちらで書いても同じコードパスを通る。

## 棄却した代替案

### 書き方を判定して attr path を切り替える

属性選択が `passthru` にそのまま届く以上、判定しても得るものが無い。判定を誤るという新しい失敗の
原因と、それを確かめるテストが増えるだけになる。

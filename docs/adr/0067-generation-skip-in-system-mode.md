---
id: "ADR-0067"
type: adr
name: "世代スキップを system mode へ拡張する"
status: 採用
issues:
  - "#519"
  - "#493"
origin: "Issue #493（epic: NixOS / nix-darwin モジュールの実装）の grilling（2026-10-24）"
justifies:
  - "REQ-46fccb80-4bae-4d37-bc19-dded88e9a9c0"
revises:
  - "ADR-0005"
  - "ADR-0006"
references:
  - "ADR-0017"
  - "ADR-0036"
  - "ADR-0058"
  - "ADR-0066"
---
# ADR-0067: 世代スキップを system mode へ拡張する

- ステータス: 採用
- 日付: 2026-10-24
- 関連: ADR-0017, ADR-0036, ADR-0058, ADR-0066, GitHub Issue #519, #493
- 改訂対象: ADR-0005「世代スキップ短絡を project mode の必須要件にする」の「世代スキップは project mode 限定」を project / system mode へ広げる（home mode の「適用のたびに新世代」は不変）。ADR-0006「実行フロー」4 の「project mode かつ」を「project / system mode かつ」へ改める
- 起点: Issue #493（epic: NixOS / nix-darwin モジュールの実装）の grilling（2026-10-24）

## 背景

NixOS の activation は switch のときだけでなく、boot のたびにも実行される（stage-2 が `$systemConfig/activate` を実行する）。NixOS モジュール（→ ADR-0066）はこの activation から `layat apply --all --manifest …` を起動するので、構成を変えていなくても boot のたびに engine が呼ばれる。

engine が世代を積まずに済ませる世代スキップは project mode に限られており（→ ADR-0005）、system mode では apply のたびに新しい世代を積む。このままでは、boot を重ねるだけで system mode の profile の世代が増え続ける。

## 決定

### 1. 世代スキップを project mode と system mode の両方で発火させる

- 判定と動作は project mode と同じ機構を使う。新しい link-farm が profile の current 世代に一致し、last-applied 記録があればその `linkFarm` にも一致するとき（→ ADR-0017, ADR-0058）、新しい世代を積まない。各 target の `lstat` 検査と、記録どおりでない entry だけの再配置は行う。
- home mode は従来どおり、適用のたびに新しい世代を積む。
- 新しい CLI フラグは足さない。

## 根拠

- **system mode へ広げる理由**: boot 時の再 activation は、内容の同じ apply を高い頻度で繰り返す点で devShell の `shellHook` と同じ性質を持つ。project mode で世代の無限増殖を止めた理由が、そのまま当てはまる。
- **同じ機構を使う理由**: スキップ判定とドリフト修復は project mode で実装・検証済みで、判定条件を mode ごとに変える理由が無い。boot 時に foreign tool が target を書き換えていても、lstat 検査がそれを直す。
- **home mode を据え置く理由**: 本 ADR が扱うのは、system mode の NixOS モジュールで起きる boot 時の世代増殖である。home mode で同じ内容の apply でも世代を残す従来の挙動は、この変更のついでに変えない。home mode で同じ増殖が起きるか（NixOS の中で使う HM の activation を含む）は本 ADR では扱わない。
- **system mode の世代公開との関係**: system mode も standalone では rollback / list-generations を公開する（→ ADR-0036）。スキップは新しい世代を積まないだけで既存の世代を消さないので、rollback 先は失われない。
- **host rollback との関係**: host が旧構成へ戻すと旧 link-farm は profile の current 世代と異なるので、スキップされず新しい世代が積まれる。module での「layat の profile は前進だけで追従する」は変わらない。

## 影響

- **engine**: スキップ条件が project と system の両 rootKind を通す。
- **要求**: 世代スキップの要求（REQ-46fccb80-4bae-4d37-bc19-dded88e9a9c0）の適用範囲を project / system mode に改める。
- **テスト**: `internal/engine/drift_test.go` に system / home の両 rootKind の場合を、`tests/e2e/scenarios/12-system.sh` に同一構成の再 apply で世代が増えないことを足す。

## 棄却した代替案

- **NixOS モジュールが boot 時（`NIXOS_ACTION` 未設定）の起動を止める**: boot 時のドリフト修復の機会を失う。モジュール側に起動条件の分岐を持ち込むことにもなる。
- **全 mode でスキップする**: home mode の世代公開の挙動を、本 ADR の範囲を超えて変えてしまう。
- **スキップを opt-in の CLI フラグにする**: モジュールが必ず付けるフラグになり、system mode の既定にしない理由が無い。

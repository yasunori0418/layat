---
id: "TC-3c28934c-d098-40a3-b598-1b08117f9cf1"
type: test_condition
name: "baseline を last-applied 記録 → profile → 空の順で読み、各操作が記録を FS の事実どおりに保つ"
mitigates:
  - "RISK-ac5cd9b4-58f9-43d8-9fae-66f7b5940beb"
---
# TC-3c28934c-d098-40a3-b598-1b08117f9cf1: baseline を last-applied 記録 → profile → 空の順で読み、各操作が記録を FS の事実どおりに保つ

## テスト条件

stale 除去・ドリフト修復・reset・rollback が使う baseline（この root に layat が最後に
置いたもの）について、読みの優先順位と、各操作が記録を書き換える時機と内容を検証する。

**読み**

- 記録があれば記録の manifest を baseline にする。profile の current と `linkFarm` が
  食い違えば記録を採り、warning を 1 行出す
- 記録が無ければ profile の current を baseline にする
- 両方無ければ baseline は空で、「last-applied 記録なし」の warning を 1 行出す
- 壊れた JSON・必須フィールドの欠落や型違い・未知のフィールド・`recordVersion` ≠ 1 の記録は
  「記録なし」に倒さず error にする

**書き**

- apply は全 FS 操作の成功直後・世代 commit の前に、そのまま（内容を変えずに）埋め込んだ manifest・
  `linkFarm`・`appliedAt` を書く。FS 操作が失敗すれば書かない。記録を書けなければ apply は失敗し、
  undo journal が配置を巻き戻す
- generation skip は新 link-farm が profile の current に一致し、記録があればその `linkFarm` にも
  一致するときだけ行い、その経路でも記録を更新する。profile が無いときや両者が食い違う間は、
  記録を baseline に通常経路を通る
- rollback は FS の収束後・profile ポインタ移動の前に N-1 の manifest と link-farm で書き直す
- reset は撤去対象のうち target が無くなった entry だけを除いて書き直し（foreign として
  残した entry・途中失敗で未到達の entry・`linkFarm` は保持）、全撤去なら記録を削除する
- reset は FS 操作の途中失敗でも同じ規則で記録を書き直し、書き直しも失敗したときは撤去の
  error と書き直しの error を併せて返す
- profile が無く記録だけがあるとき、reset は flock を取る前に profileDir（roothash 系列は
  backref `.root` も）を apply と同じ手順で作る
- generation skip・rollback・reset も、記録を書けなければ失敗する

上位の規範は TP-e7c25263-6d2d-4a37-8275-26906889d912（`internal/engine/` を実 FS の tmpdir で駆動する統合レベル）。

## 対応する CASE

CASE-233855f7-912c-4477-af32-9a54c82353c3（`internal/engine/record_test.go`）。実 nix での記録の
位置と rollback 後の書き直しは CASE-2e7b9ea2-4e4d-48ca-9e1e-0bfbb6223ff2（e2e 02-home）が見る。

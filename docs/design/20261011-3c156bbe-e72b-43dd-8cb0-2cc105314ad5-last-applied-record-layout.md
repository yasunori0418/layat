---
id: "DSG-3c156bbe-e72b-43dd-8cb0-2cc105314ad5"
type: design
name: "last-applied 記録は <state>/layat/<鍵>/last-applied.json に manifest を verbatim で埋め込み、baseline は記録 → profile → 空の順に読む"
satisfies:
  - "REQ-9d622539-a1e5-4719-af51-e0cd85b66f35"
---
# DSG-3c156bbe-e72b-43dd-8cb0-2cc105314ad5: last-applied 記録は <state>/layat/<鍵>/last-applied.json に manifest を verbatim で埋め込み、baseline は記録 → profile → 空の順に読む

## 設計

```
<state>/layat/<name>/last-applied.json                 # home（--root なし）
<state>/layat/<roothash>/.root                         # backref（profileDir と同じ規則）
<state>/layat/<roothash>/<name>/last-applied.json      # project / fixed / --root 上書き
```

```json
{
  "recordVersion": 1,
  "linkFarm": "/nix/store/<hash>-layat-<name>",
  "appliedAt": "2026-10-11T12:34:56+09:00",
  "manifest": { ...配置した manifest.json を verbatim... }
}
```

- **鍵の解決は `internal/paths`** が profileDir と同じ規則で行う（基底だけが
  `<state>/layat`）。backref `.root` も同じ階層に置く。
- **読み書きは engine が持つ**。読みは `recordVersion` 不一致・壊れた JSON を error にする。
  書きは temp + rename で atomic に行い、全 FS 操作の成功直後・世代 commit の前に呼ぶ。
  削除は reset の全撤去時だけ。
- **baseline の優先順位は 1 つの関数に集める**。apply（stale 除去と drift / generation skip の
  link-farm 比較）・rollback（離れる世代 N）・reset（撤去対象）が同じ関数で baseline を得る。
  記録があれば記録の manifest、無ければ profile リンクの current、どちらも無ければ空。
  両方あって `linkFarm` が profile の指す先と食い違えば記録を採り warning を 1 行出す。
- **更新する経路**: apply の通常経路と generation skip 経路・rollback 成功後（N-1 の
  manifest）・reset 成功後（撤去 entry を除いた manifest）。rollback の target（N-1）は
  引き続き `profile-<N-1>-link` から読む。
- 記録を flock の鍵にしない。flock の鍵・prune の対象・`--json` の payload は従来どおり
  profileDir 側に残る。

## 出典

ADR-0058「last-applied 記録」§1〜§4。

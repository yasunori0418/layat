---
id: "RISK-b7159827-3d29-405f-87ef-cdce7bb9a72f"
type: risk
name: "配置並列で集約順・カウント・エンベロープが完了順に依存して非決定化するかレースする"
likelihood: medium
impact: high
level: high
threatens:
  - "REQ-4cbd9a0d-9f94-4747-8881-56020dc6d5af"
  - "REQ-059eb4d5-63fb-4f8e-b705-11b5e2ed4ae5"
  - "REQ-b7bb09d6-74c4-44d6-905f-cb5e8383ea32"
---
# RISK-b7159827-3d29-405f-87ef-cdce7bb9a72f: 配置並列で集約順・カウント・エンベロープが完了順に依存して非決定化するかレースする

## リスク

配置（engine 呼び出し）を config 単位の worker pool で並列にすると、辞書順の逐次ループでは
構造的に起きなかった壊れ方が集約側に生じる。

- **集約順の非決定化** — config ごとの出力（`-v` レポート・`--dryrun` の計画・`apply <name>
  failed` / `skipped apply <name>` の行）を worker から直接書くと、完了順に並んで実行ごとに
  入れ替わる。subject を worker 内で登録すると `results[]` も完了順になり、辞書順集約を
  求める REQ-4cbd9a0d-9f94-4747-8881-56020dc6d5af と、単一実行と同一形状・選択順という
  REQ-059eb4d5-63fb-4f8e-b705-11b5e2ed4ae5 の前提が崩れる
- **カウント・エンベロープのレース** — applied / skipped / failures と error / conflict の
  判定を保護なしで共有すると更新が失われ、失敗があるのに exit 0 になる・error が conflict に
  負ける（REQ-b7bb09d6-74c4-44d6-905f-cb5e8383ea32 の優先度が崩れる）。`nifaceRun` の
  subject 一覧への append が並行すると `results[]` から config が欠ける
- **意味論の取り違え** — 並列化の書き換えで try-lock の skip を失敗へ数える・部分失敗時に
  成功 config の subject を捨てる、といった退行が入り込む

## 影響

impact は high とする。集約の取りこぼしは CI で `apply --all` を回す利用者に、配置されて
いない config を成功と報告する沈黙した誤りになる。順序の非決定は `--json` の差分比較や
ログ突合を壊す。

likelihood を medium とするのは、subject の事前登録・index 付き結果スライス・完了後の
flush という対策が集約関数の中に閉じている一方、集約関数を書き換える変更（後続の検証
強化や出力形式の追加）でこの境界が崩れる余地が残るため。

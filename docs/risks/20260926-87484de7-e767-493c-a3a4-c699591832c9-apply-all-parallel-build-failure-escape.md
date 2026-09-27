---
id: "RISK-87484de7-e767-493c-a3a4-c699591832c9"
type: risk
name: "並列 build の失敗が集約から抜けるか並列度が制御されず nix daemon を圧迫する"
likelihood: medium
impact: high
level: high
threatens:
  - "REQ-91f0a9a7-c2b2-4cda-8cce-cbf5d8c5d04d"
  - "REQ-059eb4d5-63fb-4f8e-b705-11b5e2ed4ae5"
---
# RISK-87484de7-e767-493c-a3a4-c699591832c9: 並列 build の失敗が集約から抜けるか並列度が制御されず nix daemon を圧迫する

## リスク

build を worker pool で先行並列にすると、逐次ループなら構造的に起きなかった 2 種類の壊れ方が
生じる。

- **失敗が集約から抜ける** — stage 1 の結果を config 名で引き当て損ねる・並行書き込みで
  結果を取りこぼす・build 失敗の config を subject 登録せずに捨てる、といった実装では、
  失敗した config が `results[]` にも failures にも現れず、`apply --all` が exit 0 で終わる。
  逆に失敗 config を stage 2 へ流すと、先行 build の意味が消えて in-lock build でもう一度
  失敗する（あるいは stage 1 と 2 の結果が食い違う）。subject の登録位置がずれると
  `results[]` の辞書順が `--jobs` によって変わり、単一実行と同一形状という
  REQ-059eb4d5-63fb-4f8e-b705-11b5e2ed4ae5 の前提が崩れる
- **並列度が制御されない** — worker 数が `--jobs` を超える・`--jobs` の解決（0 → CPU 数・
  負値拒否）を誤ると、config 数だけ `nix build` が同時に立ち上がり、nix daemon のキューと
  メモリを圧迫する。`--jobs` が `--all` 以外で黙って受理されると、効かないフラグを利用者が
  効いていると誤解する

## 影響

impact は high とする。失敗の取りこぼしは CI で `apply --all` を回す利用者にとって、配置
されていない config を成功と報告する沈黙した誤りになる。並列度の暴走は同じマシン上の
他のビルドを巻き込んで遅延・OOM を起こし得る。

likelihood を medium とするのは、stage 2 は既存の集約関数をそのまま使い、失敗の短絡も
1 箇所の包み関数に閉じている一方、後続の配置並列化（Issue #154）で集約が並行化され、
この境界が組み替わる余地があるため。

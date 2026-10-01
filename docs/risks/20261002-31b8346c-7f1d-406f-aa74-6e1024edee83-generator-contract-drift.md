---
id: "RISK-31b8346c-7f1d-406f-aa74-6e1024edee83"
type: risk
name: "生成器の実装が契約からずれ、経路ごとに manifest の取得や失敗の分類が食い違う"
likelihood: medium
impact: medium
level: medium
threatens:
  - "REQ-194e4209-d804-4a4b-a2b8-3d39c6c33729"
  - "REQ-7a2f1ecf-4675-45aa-80c0-a8fc58db9edd"
---
# RISK-31b8346c-7f1d-406f-aa74-6e1024edee83: 生成器の実装が契約からずれ、経路ごとに manifest の取得や失敗の分類が食い違う

## リスク

CLI は manifest を生成器の契約（Discover / Roots / AllRoots / Build / DryBuild）越しに得る
（→ REQ-194e4209-d804-4a4b-a2b8-3d39c6c33729）。実装は nix・prebuilt（`--manifest`）・テスト
ダブルの 3 つで、CLI はどれが渡されても同じ手順で engine へ注入する。実装の 1 つが契約から
ずれると、その実装を通る経路だけが壊れる。

ずれ方は 2 方向ある。

- **操作の意味** — prebuilt の Build / DryBuild が与えられた link-farm 以外を返す、Roots が
  manifest.json の root を取り違える、テストダブルが呼び出しを記録し損ねる。`--manifest` 経路
  （モジュール activation）だけが別の link-farm を適用し、あるいは cmd 層のテストが実際とは
  違う呼び出し列を緑にする
- **失敗の構造** — `generator.Error` が原因の error を保持し損ねる、Stage を取り違える
  （→ REQ-7a2f1ecf-4675-45aa-80c0-a8fc58db9edd）。`--json` の分類は Stage と原因の chain で
  決まるため、entrypoint の不在や `-f` のパスの不在が `E_LAYAT_BUILD` に化け、あるいは逆に
  評価・build の失敗が `E_LAYAT_FAILED` へ落ちる

## 影響

意図しない link-farm の適用、あるいは機械可読出力の誤分類。impact を medium とするのは、
適用されるのは layat の通常の配置で、記録外の実体は conflict で停止し、正しい link-farm での
再 apply が意図した状態へ収束させるため。誤分類は消費側の判断を誤らせるが、再実行で正しい
コードが得られる。

likelihood を medium とするのは、契約の実装が生成器化の直後で、診断契約（Issue #403）・
選択機構（Issue #404）が続けて同じ型を触るため。

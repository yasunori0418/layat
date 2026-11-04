---
id: "RISK-8241dd1d-f856-4ed5-b572-05b88232a199"
type: risk
name: "前段衝突検査のバケット判定を誤り衝突をすり抜けさせるか誤って停止させる"
threatens:
  - "REQ-38506419-c23d-4fce-b167-30715ce6a69d"
  - "REQ-fc1118b1-b0e8-4ddf-80f6-c70956651693"
likelihood: medium
impact: medium
level: medium
---
# RISK-8241dd1d-f856-4ed5-b572-05b88232a199: 前段衝突検査のバケット判定を誤り衝突をすり抜けさせるか誤って停止させる

## リスク

前段の確認の正しさは「どの config 同士を同じ root とみなすか」というバケットの切り方で決まる。
切り方を誤ると 2 方向に失敗する。

- **すり抜け** — 同じ root に解決する config を別バケットへ分けると、衝突が検出されないまま
  配置に進み、後勝ち置換のフリップフロップが恒常化する。`--root` 一律上書き時に rootKind
  ごとに分けたまま確認する、fixed root の値比較を取り違えて同じ root 値の config を別
  バケットへ分ける、などがこの方向
- **誤検出** — 別の root に解決する config を同じバケットへ入れると、何も衝突しない
  `apply --all` が止まる。fixed root を root 値ではなく rootKind だけで束ねる、project と
  fixed を同一視する、選択外の config の衝突で選択側を止める、などがこの方向

加えて、停止の位置がずれると確認の意味が失われる。build や配置の後に判定すると、止めた
時点で一部の config が既に置かれている。`--dryrun` の処理だけ確認を迂回する実装も同じ。

別々の config が同じ target を狙うことによる振動は、engine の実行時には止めない
（→ REQ-fc1118b1-b0e8-4ddf-80f6-c70956651693）。1 回の `apply --all` に載る config どうしでは、この
前段の確認が配置前に止める唯一の場所になる。確認が衝突を取りこぼすと、apply のたびに config どうしが
target を取り合い、foreign 警告が出続ける。

## 影響

impact は重い方向（すり抜け）で medium とする。すり抜けても配置は従来どおりの後勝ち +
foreign symlink warning になるだけで、config の重複を解消して再 apply すれば収束する。
誤検出も設定を直せば（あるいは確認の誤りを直せば）再実行で通る。どちらもユーザー環境を
不可逆に壊さない。すり抜けは warning 以外に何も知らせないため沈黙しやすい。

likelihood を medium とするのは、判定は純粋関数に切り出されテストで覆われている一方、
`runApplyAll` は後続の並列化（ADR-c109ac0c-d5df-4a15-acf0-d6d85ce5402e / ADR-e7a57f51-41da-460c-8f51-c84e94643f3d）で build・配置の段が大きく組み替わり、確認の
呼び出し位置（root フィルタ適用後・`--dryrun` 分岐と build の前）がずれる余地があるため。

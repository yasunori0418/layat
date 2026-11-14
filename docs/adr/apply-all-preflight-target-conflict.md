---
id: "ADR-75df92ae-0694-452b-8fb4-350994e31b25"
type: adr
name: "apply --all は選択した config 間の target 重複を build 前に止める"
section: cli-commands
status: stable
sources:
  - id: "issue-584"
    resource: "https://github.com/yasunori0418/layat/issues/584"
    title: "#584"
origin: "ADR-0038 §1（--all で選択した config 集合内の正規化後 target の重複を build 前に error 停止する）と §3（単一の apply は確かめず、検出できない衝突は後勝ちのままにする）を書き直した"
justifies:
  - "REQ-38506419-c23d-4fce-b167-30715ce6a69d"
  - "REQ-5c6b07da-3d06-414d-8770-4f438234b322"
references:
  - "ADR-095220c9-5a3d-4fd4-94bd-8fa1a02dd36d"
  - "ADR-28605860-770e-41df-95dd-7dc05b20497a"
  - "ADR-af37a636-a798-4c03-8fa4-e73ed6375a92"
  - "ADR-b3f3b68e-5096-46d1-bafd-53a9ab4711f4"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:26:59+09:00"
---
# apply --all は選択した config 間の target 重複を build 前に止める

- ステータス: stable
- 日付: 2026-11-03
- 関連: [cross-config の同一 target は後勝ちにし、foreign symlink の上書きで警告する](cross-config-last-wins-with-foreign-warning.md), [apply --all は config ごとの情報を 1 回の eval でまとめて取る](apply-all-single-eval.md), [apply --all の集約は config 名の辞書順に並べる](apply-all-lexicographic-order.md), [配置の振動はユーザー責任とし、文書で注意するに留める](oscillation-is-user-responsibility.md)
- 起点: ADR-0038 §1（--all で選択した config 集合内の正規化後 target の重複を build 前に error 停止する）と §3（単一の apply は確かめず、検出できない衝突は後勝ちのままにする）を書き直した

## 背景

別々の config が同じ target を持つと、engine はそれぞれ自分の前世代の manifest としか比べないので、
相手の symlink を foreign と見る。[cross-config の同一 target は後勝ちにし、foreign symlink の上書きで警告する](cross-config-last-wins-with-foreign-warning.md)。`apply --all` を実行するたびに所有者が入れ替わり、
衝突はいつまでも解消しない。古くなった配置の除去の保守的な条件は衝突があっても保たれるので、実害は
どちらの symlink が残るかに限られる。それでも設定の誤りであることに変わりはない。

一方、`apply --all` が扱う config の正規化後 target は、build の前に分かる。同じ entrypoint から
選んだ config なら一括 eval で取れ（[apply --all は config ごとの情報を 1 回の eval でまとめて取る](apply-all-single-eval.md)）、`--manifest` で渡された
ビルド済みの集合なら各 link-farm の `manifest.json` から読める。

## 決定

`apply --all` は、選択した config 群の正規化後 target に重複が無いかを build の前に確かめ、
重複があれば build にも配置にも入らずエラーで止める。`--manifest` で渡されたビルド済みの集合も
同じ確認を通る。名指しの `apply <name>` ではこの確認をしない。
この確認で見えない衝突は、これまでどおり実行時の後勝ちと foreign symlink の warning で扱う。

確認の対象・バケットの分け方・終了コードは [apply --all は build 前に選択 config 間の target 衝突を検出して停止する](../requirements/apply-all-cross-config-preflight.md) に、
同一 manifest 内の衝突と cross-config の衝突を別々に扱うことは
[target 衝突の検出処理を同一 manifest 内と cross-config で分ける](../requirements/target-collision-eval-detection.md) に置く。判定を置く層は
[config 間の target 衝突は CLI が判定し engine に持ち込まない](cross-config-conflict-judged-by-cli.md) が定める。

## 根拠

所有者の入れ替わりは設定の誤りで、warning を出して通しても直らない。同じ起動に載る config 同士
なら build の前に分かるので、その段で止めれば、利用者は所有者が入れ替わる前に誤りに気付ける。

`apply --all` を並列に実行すると、後勝ちの順序が実行ごとに変わる。検出できる衝突を前段で取り除いて
おけば、並列化しても挙動の決定性を失わない。

名指しの apply は対象の config しか eval しない。兄弟の config まで調べると、名指しのたびに全 config
を eval することになり、単一 config の apply の軽さを失う。名指しの apply での衝突は実行時の
warning で見える。

別の entrypoint・別のツール・手作業の配置との衝突は、eval の段では見えない。fixed root の値が
たまたま project root と一致する場合も同じである。これらはもともと競合であり、前段の確認では
扱えない。

## 棄却した代替案

### warning を出して続ける

設定の誤りが残り、並列化した後は勝つ側が実行ごとに変わる。

### 名指しの `apply <name>` でも兄弟の config を確かめる

名指しのたびに全 config の eval を実行することになる。`--all` での確認だけで足りる。

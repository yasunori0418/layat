---
id: "RISK-7808768d-bd87-429a-af42-52e8559d940c"
type: risk
name: "Nix と engine の唯一の契約である manifest の生成側と消費側がずれる"
status: stable
threatens:
  - "REQ-79ce0a09-f9bd-4e61-ba7f-45fb5643137b"
  - "REQ-250d936c-1df0-491d-a7af-708f38b61f33"
  - "REQ-dec58330-6dad-47f7-8f56-2402764a89c7"
  - "REQ-dd10d820-e453-4099-a47a-ffb9a7de02fb"
  - "REQ-2b0c2bb8-964f-4e36-a121-c6ea0d4be1c4"
likelihood: low
impact: high
level: medium
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T23:37:58+09:00"
---
# Nix と engine の唯一の契約である manifest の生成側と消費側がずれる

## リスク

`manifest.json` は Nix 側の評価と Go 側の engine をつなぐ唯一の安定契約であり、両側が同じ
文書を同じ意味で読めることに全体が乗っている。ここがずれると、評価は成功しビルドも通るのに
配置の段で意味を取り違える、という発見の遅い形で失敗する。

ずれ方は 2 方向ある。

- **バージョンの skew** — 利用者が pin した flake（生成側）と手元の CLI（消費側）の版が
  食い違い、消費側の知らない版の文書が渡る。このとき「破損した JSON」ではなく「読めるが意味の
  違う文書」として通ってしまうと、ずれに気づく機会が失われる。移行機構を持たない方針を
  採っている以上、拒否そのものが契約の一部であり、しかも原因が skew だと分かる形で拒否
  できなければ利用者は対処できない
- **契約の暗黙拡張** — 消費側が未知のフィールドを利用者に知らせずに無視すると、生成側が足したつもりの
  情報が届かないまま実行が進む。届いていないことに誰も気づかない

root の種別ごとの扱いも同じ場所にある。評価時に絶対パスが決まる種別と、実行時に解決する種別が
あり、この区別を取り違えると配置先そのものが変わる。本 item はこの区別を消費側が受理の段で
守ること（→ [manifest.json の root は rootKind を持ち fixed のときだけ絶対パスを併記する](../requirements/manifest-root-object.md)）を射程に含む。生成側が同じ区別で文書を組み立てる側は
[manifest.json の形が engine の読む契約から気付かれずにずれる](manifest-contract-drift.md) の担当で、同じ `requirement` を別の facet から脅かす。

## 影響

意図しない場所への配置、あるいは原因の分からない失敗。impact を high とするのは、root の種別を
取り違えた受理が配置先そのものを変え、書き込み先が layat の記録に残らないため。どこへ撒いたかを
辿れない以上、契約を直して再 apply しても撒いた側は回収されない。ずれが「読めるが意味の違う
文書」として利用者に知らせずに通り、統合層すべてが前提とするこの契約を通じて layat のどの使い方にも波及
することは、その回収に着手する契機まで奪う。

likelihood を low に留めるのは、本 item の射程が消費側の受理
（版の拒否・未知フィールドの拒否・root の種別が空である文書の拒否・実行時に解決する種別が
パスを併記した文書の拒否）と統合層の配線に限られ、前者は Go の契約テスト（[manifest の読み取りが v1 を正確に復元し版の境界・未知フィールド・不整合な root 種別を拒否する](../test/integration/manifest-decode-schema-guard.md)）が、
後者は activation が root と entry とオプションを engine へ届けることの評価アサート
（[activation が評価時点で engine 起動へ配線され root と entry とオプションを届ける](../test/integration/activation-engine-kick-wiring.md)）が覆っているため。評価層が文書の形を変えてしまう側は本 item ではなく
[manifest.json の形が engine の読む契約から気付かれずにずれる](manifest-contract-drift.md) の担当で、そちらはフィールド追加・既定値変更という通常の改修で動くぶん
likelihood が一段高い。

消費側の受理が見るのは root の種別が空でないことと、実行時に解決する種別がパスを併記して
いないこと（→ [manifest.json の root は rootKind を持ち fixed のときだけ絶対パスを併記する](../requirements/manifest-root-object.md)）までで、種別が既知の 4 値のいずれかであることは見ていない。
そちらは engine の root 解決が未知の種別を弾いて止まるため、利用者に知らせずに通る処理にはならない。

[mkManifest は配置データを生成する純粋関数である](../requirements/mkmanifest-pure-function.md) は、配置を engine の実行時の責務とし、生成した結果を CLI が engine へ渡すことも定める。
受け渡す文書の読み方がずれると純粋性が破れなくてもこの分担は成り立たず、この懸念は受け渡しの実装を置き換えても残るので `requirement` へ張る。

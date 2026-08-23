---
id: "TC-bc653f83-70ce-4920-b7cd-9b70a4cd2cae"
type: test_condition
name: "prune が孤児系列だけを系列ごと消し、生存・判定不能な系列は理由付きで残す"
mitigates:
  - "RISK-f522a51b-db4b-4164-bd20-63fa4a0cb27d"
---
# TC-bc653f83-70ce-4920-b7cd-9b70a4cd2cae: prune が孤児系列だけを系列ごと消し、生存・判定不能な系列は理由付きで残す

## テスト条件

`Prune`（→ DSG-096dc893-21f4-45e3-9347-986e9275b4d1）を、複数の系列を並べた state dir の
tmpdir に対して駆動して検証する。判定の入力は backref `.root` と root の実在だけなので、
fixture は state dir の形と root ディレクトリの有無で作る。

**孤児の削除** — `.root` が指す root を消した系列は、`<roothash>` ディレクトリごと消える。
配下の全 `<name>` profile・世代リンク・`.pending`・`.root` が残らないことまで見る。

**世代を持たない系列** — `.pending` と `.root` だけで世代リンクが 1 つも無い系列も、root 不在
なら同じく削除される。世代の有無を条件にしていないことの担保。

**生存系列** — root が実在する系列は残る。`Removed` に現れず、失敗でもないので `Skipped` にも
入らない。

**`.root` を持たない dir** — home mode / system mode の `<name>` 直キー系列は走査に混ざらない。
判定すべき root パスが導けないため構造的に対象外という前提を、実際の列挙で確かめる。

**判定不能な系列** — backref が読めない・空・絶対パスでない系列は残り、`Skipped` に
`backref-unreadable` として理由付きで現れる。root の stat が不在以外の理由で失敗する系列も
同様に残り、`root-stat-failed` になる。判定材料が欠けたときに削除へ倒れないことの担保。

**dangling symlink の root** — symlink 越しにしか辿れず、その先が存在しない root は不在として
扱われ、系列が削除される。

**配置物** — root が実在する系列の配置先も、削除対象系列とは無関係な FS 上のファイルも、
`Prune` の前後で変わらない。

上位の規範は TP-e7c25263-6d2d-4a37-8275-26906889d912（`internal/engine/` を実 FS の tmpdir で
駆動する統合レベル）。安全機構の側（dryrun・確認・lock）は
TC-a9857bf7-f7f9-41f9-b42c-9993fd16a5e9 の担当。

## 対応する CASE

未着手（テスト資産と同じコミットで起こす）。

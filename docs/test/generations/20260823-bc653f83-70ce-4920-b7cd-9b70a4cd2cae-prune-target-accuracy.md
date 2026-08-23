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

**世代を持たない系列** — `<name>` の下に `.pending` があるだけで世代リンクが 1 つも無い系列も、
root 不在なら同じく削除される。世代の有無を条件にしていないことの担保。`<roothash>` 直下が
`.root` 単独（`<name>` が 1 つも無い）の形も置く。削除手順が `<name>` のループ → `.root` →
`<roothash>` の rmdir に縮退したときも、`<roothash>` まで消え切ることを見る。

**生存系列** — root が実在する系列は残る。`Removed` に現れず、失敗でもないので `Skipped` にも
入らない。

**`.root` を持たない dir** — home mode / system mode の `<name>` 直キー系列は走査に混ざらない。
判定すべき root パスが導けないため構造的に対象外という前提を、実際の列挙で確かめる。走査基底が
2 つある（ユーザー state 基底と system 基底 → ADR-0036 §3）ので、**両方の基底で同じ列挙条件が
効くこと**も見る。基底はどちらも `PruneOptions` のフィールドで受けるので、system 基底も tmpdir へ
差し替えて駆動する（→ DSG-096dc893-21f4-45e3-9347-986e9275b4d1。実 `/nix/var/nix/profiles/` は
触らない）。列挙そのもの（`ListRootHashSeries`）と backref 読み取り（`ReadBackref`）は
`internal/paths` に置かれるが FS を読むので、同層の純関数と違って table では覆えず、この TC の
tmpdir 駆動が覆う。

**基底が無い / 読めない** — 基底ディレクトリ自体が存在しないときは、その基底の孤児 0 件として
正常に続く（system mode を使ったことがない環境では system 基底が無いのが普通）。両基底とも
不在でもエラーにしない。基底が存在するのに `ReadDir` が失敗するときは、その基底を名指しした
warning が出て、もう一方の基底の処理は続く。**「基底が無い」と「基底が読めない」が区別される**
ことの担保（→ RISK-f522a51b-db4b-4164-bd20-63fa4a0cb27d）。読めない側の誘発は root でも成立する
側を採る（基底のパスに通常ファイルを置いて ENOTDIR を出す。mode 0 の権限誘発は root で素通り
する → TP-deb05610-44bc-4962-8939-952392e5fbd0 の横断規約）。

削除段の失敗（権限不足を含む）は TC-a9857bf7-f7f9-41f9-b42c-9993fd16a5e9 の担当。

**判定不能な系列** — backref が読めない・空・絶対パスでない系列は残り、`Skipped` に
`backref-unreadable` として理由付きで現れる。3 条件は実装上も別分岐（I/O エラー / 長さ 0 /
`filepath.IsAbs` false）になるので、**同じ理由へ落ちることを 3 ケース独立に確かめる**。空と
相対パスは `ReadFile` の err チェックだけを書いた実装が素通りする経路で、代表 1 件では覆えない。
併せて末尾改行付きの絶対パス（`.root` をファイルとして書けばまず起きる形）を境界として置く。
これは trim されて**通常どおり判定される**べきケースで、trim しない実装では `os.Stat` が失敗して
`root-stat-failed` へ落ちる。孤児でない系列を残す方向なので実害は小さいが、実在する root が
「判定不能」と報告され、孤児の系列は消えなくなる。
root の stat が不在以外の理由で失敗する系列も同様に残り、`root-stat-failed` になる。判定材料が
欠けたときに削除へ倒れないことの担保。

**系列単位の判定段失敗** — 列挙が系列ごとに載せる失敗（`BackrefErr` / `NamesErr` →
`internal/paths` の `RootHashSeries`）は、どちらも基底全体を落とさず、その系列だけを残して
理由付きで `Skipped` に載せる。`BackrefErr` は `backref-unreadable`、`NamesErr` は
`series-unreadable`（→ DSG-096dc893-21f4-45e3-9347-986e9275b4d1）。同じ基底の健全な孤児系列は
変わらず削除されることを併せて見る。**`NamesErr` の系列を独立に置くのが要点**で、この系列は
`Names` が nil のまま返るため、`NamesErr` を見ずに nil だけで分岐する実装では「`<name>` を
持たない `.root` 単独の系列」と区別が付かず、lock を取らずに削除する経路へ落ちる。削除されず
`series-unreadable` になることまで固定する。誘発は権限を落として行うため root 実行では skip
する（→ TP-deb05610-44bc-4962-8939-952392e5fbd0 の横断規約）。

**dangling symlink の root** — symlink 越しにしか辿れず、その先が存在しない root は不在として
扱われ、系列が削除される。

**root がディレクトリでない** — root だったパスが通常ファイルに置き換わっている系列は「実在
する」側に落ちて残る。判定条件が「実在しない」だけで種別を見ないことの帰結を 1 ケース固定する
（後から「ディレクトリかどうかも見るべき」と足すと削除側へ静かに倒れるため）。

**配置物** — root が実在する系列の配置先も、削除対象系列とは無関係な FS 上のファイルも、
`Prune` の前後で変わらない。

**warning** — `Skipped` に入るケース（`backref-unreadable` / `series-unreadable` /
`root-stat-failed`）では、`Skipped`
への計上だけでなく **`Warnf` が理由付きで呼ばれる**ことも各ケースで見る。DSG が「`Skipped` に
入る全ケースで `Warnf` を呼ぶ」を規律にしているのに `locked` だけ warning を確かめる形にすると、
検証の側が禁じられた不揃いと同じ形になる。

上位の規範は TP-e7c25263-6d2d-4a37-8275-26906889d912（`internal/engine/` を実 FS の tmpdir で
駆動する統合レベル）。安全機構の側（dryrun・確認・lock・削除段の失敗）は
TC-a9857bf7-f7f9-41f9-b42c-9993fd16a5e9 の担当。

## 対応する CASE

- CASE-49ac6d68-ed29-41aa-b939-dcfdc8711082（`internal/engine/prune_test.go`）— `Prune` を
  tmpdir の state dir / system 基底で駆動し、この TC の判定分岐を覆う主戦力
- CASE-ff4a842e-dd25-4c75-82ef-185507781d02（`internal/paths/paths_test.go`）— 列挙
  （`ListRootHashSeries`）と backref 読み取り（`ReadBackref`）そのものを名指しで覆う。
  `.root` の有無による振り分け・基底が「無い」と「読めない」の区別・系列単位の
  `BackrefErr` / `NamesErr` はこちらが持ち、`Prune` 経由の結合は上の CASE が見る

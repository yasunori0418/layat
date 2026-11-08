---
id: "TC-a9857bf7-f7f9-41f9-b42c-9993fd16a5e9"
type: test_condition
name: "prune の dryrun が副作用ゼロで、非同意は中断し、lock 中の系列は skip され、削除段の失敗は進捗と理由で振り分けられる"
status: stable
mitigates:
  - "RISK-0909e121-8ec4-40c8-ab1c-562e61f31067"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:34:43+09:00"
---
# TC-a9857bf7-f7f9-41f9-b42c-9993fd16a5e9: prune の dryrun が副作用ゼロで、非同意は中断し、lock 中の系列は skip され、削除段の失敗は進捗と理由で振り分けられる

## テスト条件

`Prune`（→ [prune を paths の列挙・engine の判定と削除・CLI の同意解釈に分け、engine API を Reset と同型にする](../../design/prune-orphan-series.md)）の安全機構を、孤児系列を含む state dir の
tmpdir で駆動して検証する。何を対象と判定するかは
[prune が孤児系列だけを系列ごと消し、生存・判定不能な系列は理由付きで残す](prune-target-accuracy.md) の担当で、ここは「対象と判定したものを実際に消してよいか」
だけを見る。

**`--dryrun` の非破壊性** — `DryRun` で駆動すると FS が一切変わらず、削除予定が `Removed` に
入って返る。lock も取らないので、dryrun 中に他の engine が同じ系列の lock を取れる（dryrun が
lock を取ると本番の apply をブロックする脅威は [dryrun が副作用を残す、または本番 apply と異なる結果を報告する](../../risks/dryrun-diverges-from-apply.md) が
同型で持つ）。preview が preview であることの担保。

**確認へ渡る情報** — `Confirm` コールバックに渡る `PruneResult` に、削除予定系列の root パスが
入っている。CLI が root パス一覧を表示できる材料が engine から出ていることの担保
（アンマウント中 root の誤判定に対する唯一の手段 → [prune は削除前に root パスの一覧を示して確認を取る](../../adr/prune-confirms-with-root-list.md)）。**材料が出ていることまでが
この TC の範囲**で、CLI が実際にそれを表示するかは cli-json 区分の残余になる（既存の
[破壊的操作の確認が対話性と同意フラグの組み合わせで決まり非対話で止まらない](../cli-json/confirm-policy-interactivity.md) はプロンプトを出すか出さないかの方針、
[stdout を機械可読出力が専有し人間向けの出力が stderr に留まる](../cli-json/stream-ownership.md) は stdout / stderr の排他しか見ておらず、プロンプトの
中身はどちらの範囲にも入らない）。

**非同意の中断** — `Confirm` が false を返したら削除せず、`Aborted` を立てて返る。FS は
変わらない。

**lock 規律の全分岐** — 系列内のいずれかの `<name>` profileDir を別に lock した状態で駆動すると、
その系列は削除されず `Skipped` に `locked` として現れ、warning が出る（① 1 つでも取れなければ
系列ごと skip・既に取れていた分は解放される）。`ErrLocked` 以外の `Acquire` 失敗も同じく skip +
warning になる（② 判定不能を削除として扱わない）。全部取れた系列は lock を保持したまま削除され、
削除後に解放される（③）。③ は終点だけでなく**保持区間**を見る — 削除の最中に別プロセスが同じ
profileDir の lock を取れないことまで確かめないと、削除前に解放する実装も同じく通ってしまう
（早すぎる解放は [flock による直列化が破れ、同一 config への同時実行が互いの配置を壊す](../../risks/lock-serialization-broken.md) と同型の実害を生む）。`<name>` を
1 つも持たない `.root` だけの系列は lock なしで削除される（④）。同じ実行に含まれる他の孤児系列は
削除される（1 系列の skip が全体を止めない）。lock の解放を待たないことも併せて見る。

**削除段の失敗** — DSG の 2×2（そこまでに何か消えたか × 理由が権限か）を全セル固定する。
`Skipped` になるのは 1 セルだけで、残る 3 セルはすべてエラーになる。

- **何も消えていない × 権限** — `Skipped` に `permission-denied` として現れ、warning が出て、
  他の基底・他の系列の削除は続く。[prune は system 基底も走査し、権限の無い系列を見送る](../../adr/prune-scans-system-base-skips-unpermitted.md) の非 root 実行がここに当たる。root で回すと権限
  失敗を誘発できず何も確かめないまま通るので、`os.Geteuid() == 0` を検出して skip する
  （→ [原子性は実 FS の条件で故障を誘発して不変条件ごとに検証する](../../test-plan/fault-injection-atomicity.md) の横断規約）
- **何も消えていない × 権限以外** — `Skipped` に入らずエラーで返る。`errors.Is` を省いて「着手前の
  失敗は理由を問わず skip」と書いた実装は、このセルが無ければ全ケースを通ってしまい、read-only
  FS 上の系列が「権限を直せば消える」と読める形で報告される。

  **このセルだけは「root でも成立する誘発を採る」を適用しない。** 削除の最初の一手は列挙が実在を
  確認済みの `<name>` ディレクトリの unlink で、そこを静的な fixture から非権限の errno で失敗させる
  手段が無い（ENOTDIR / ENOTEMPTY はいずれも `.root` 削除後 = 「1 つ以上消した」側にしか現れず、
  `.root` を非空ディレクトリにする案は `ReadBackref` が先に失敗して `backref-unreadable` になる）。
  production コードへフォールト注入の seam を足すのは採らない（2026-08-23 の裁定）。代わりに
  **「着手前の失敗で理由が権限でないときエラーになる」ことを permission 側と対にした形で検証する**
  — 系列直下に stray file を置いて最終 rmdir を ENOTEMPTY で失敗させ（root でも成立する）、非権限の
  失敗が `Skipped{permission-denied}` へ丸められずエラーで返ることを固定する。分岐としては下の
  「1 つ以上消した × 権限以外」と同じ処理を通るので、着手前後の別を担保するのは permission 側
  （「何も消えていない × 権限」と「1 つ以上消した × 権限」の対）が持つ
- **1 つ以上消した × 権限以外** — `Skipped` に入らない。`Removed` にも数えられず、系列を特定
  できる形でエラーが返り、後続の系列へは進まない。誘発は 2 通りを併用する。**非権限の errno は
  root でも成立する側**（系列直下に stray file を置いて最終 rmdir を ENOTEMPTY で失敗させる。前セルの
  代替検証と同じケース）。**「中途半端な状態の系列が『残した』と報告されない」報告側の担保は、複数
  `<name>` を持つ系列の 2 つ目を権限で失敗させたケース**が持つ（`Removed` に失敗前の系列だけが
  残る・`.root` が残る・当該系列が `Skipped` に無い、を同時に見る。権限誘発なので
  `os.Geteuid() == 0` で skip する）。errno の別と報告の別を 1 ケースに束ねると、root 環境で
  後者と一緒に前者の確認も skip される
- **1 つ以上消した × 権限** — 同じくエラー。理由が権限でも `Skipped{permission-denied}` には
  ならない。**このセルを省くと、`errors.Is` を先に見て権限なら無条件 skip する実装（DSG が定めた
  「状態を先に見る」順序を逆にした実装）が他の 3 セルをすべて通過してしまう**。1 つ目の `<name>`
  を消させた後、2 つ目で権限失敗を誘発する（root では誘発できないので `os.Geteuid() == 0` で
  skip する）
- **`Removed` の正の側** — 上のエラーケースを複数系列の fixture で駆動し、**失敗より前に完了した
  系列が `Removed` に残る**ことを同じケースで見る。失敗時に結果ごと破棄する実装では、既に消えた
  系列が結果に現れず、ユーザーは何が消えて何が残ったか分からないまま部分削除された FS を渡される
- **`.root` の残存** — 失敗した系列に `.root` が残っており、再実行で同じ系列がふたたび走査対象に
  なる。削除順を破る実装（`<roothash>` を一発で `RemoveAll` する等）は「`Removed` に入らず
  エラーが返る」だけなら通ってしまうが、系列は二度と辿れなくなる
  （→ [prune の dryrun・確認・lock の安全機構が働かず、同意なく削除が実行される](../../risks/prune-safety-gate-bypass.md)）

安全機構としての意味は「消し切れなかったことがユーザーに届く」ことなので、どの場面でも失敗が握り潰され
ないことを見る。

engine が見るのは `Confirm` コールバックだけで、TTY 判定と `--yes` / `--json` の解釈は CLI 側の
責務（cli-json 区分の担当）。

上位の規範は [エンジンとコマンド層の Go テストは nix を介さない実 FS 統合テストを主戦力とする](../../test-plan/go-test-layering.md)（`internal/engine/` を実 FS の tmpdir で
駆動する統合レベル）。

## 対応する CASE

- [prune_test.go — 孤児系列の判定と系列ごとの削除・dryrun / 確認 / lock / 削除段の失敗](prune-test-go.md)（`internal/engine/prune_test.go`）
- [e2e 08-prune — 実 apply が作った系列を root 削除で孤児にし、prune が系列ごと消して gcroot が外れる](e2e-08-prune.md)（`tests/e2e/scenarios/08-prune.sh`）— `--dryrun` の
  非破壊性を実 CLI を通して見るだけ。lock 規律と削除段の失敗は誘発が実機で成立しないので上の
  CASE が持つ

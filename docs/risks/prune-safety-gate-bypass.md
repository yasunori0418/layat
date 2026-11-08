---
id: "RISK-0909e121-8ec4-40c8-ab1c-562e61f31067"
type: risk
name: "prune の dryrun・確認・lock の安全機構が働かず、同意なく削除が実行される"
status: stable
threatens:
  - "REQ-42fe312c-927c-4da3-9346-f7ca2f3a58ed"
likelihood: medium
impact: high
level: high
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-05T00:26:05+09:00"
---
# RISK-0909e121-8ec4-40c8-ab1c-562e61f31067: prune の dryrun・確認・lock の安全機構が働かず、同意なく削除が実行される

## リスク

`layat prune` の判定が正しくても、削除を守る条件（→ [prune は dryrun・root 一覧付き確認・try-lock skip・--json の --yes 必須で削除を守る](../requirements/prune-safety-gates.md)）が
破れれば同じ結果になる。ここでの脅威は「何を消すと決めたか」ではなく「決めたものを、ユーザーの
同意なしに、あるいは消してはいけない状況で消してしまう」ことで、判定側の
[prune が生存系列や判定不能な系列を孤児と読み、不可逆に削除する](prune-misjudges-live-series.md) とは別の脅威を成す。

**preview の破れ** — `--dryrun` が副作用を持てば、削除予定を確認するつもりの操作が削除そのものに
なる。prune はこの操作を安全に試す唯一の手段で、しかも root の実在判定という機械にしか
見えない基準で対象を選ぶため、preview の重みは `reset` の場合より大きい。

**同意の条件の破れ** — 確認を出さない、確認の中断（no を選ぶ・非 TTY で `--yes` が無い）でも
処理が止まらず進む、あるいは確認は出るが root パス一覧を欠く。root パス一覧の表示はアンマウント中の
誤判定に対する唯一の防波堤で（→ [prune は削除前に root パスの一覧を示して確認を取る](../adr/prune-confirms-with-root-list.md)）、一覧が欠けた確認は「何を消すか分からないまま
yes を押させる」ことになり、確認が無いのとほぼ変わらない。

**lock 中系列の削除** — try-lock が取れない系列は engine が動いている = root が生きている強い
兆候。ここを skip せず削除すれば、実行中の apply の足元から世代を消すことになる。blocking wait
へ倒す誤りも、「使用中の系列を消すために待つ」倒錯した挙動として同じ場所に現れる。

**`--json` の同意迂回** — `--json` が `--yes` 無しで実行できてしまえば、機械消費の使い方が確認を
迂回して削除に到達する。

**削除中の失敗の握り潰し** — 削除が途中で失敗した系列を「消えた」として `Removed` に数えれば、
半端に破損した系列がユーザーに報告されないまま残る。「手を付けなかった」を意味する `Skipped` へ
落とすのも同じ誤りで、実際には破損しているものが無傷だと読める。

**削除順の規律の破れ** — 削除は backref `.root` を最後に消す規律で守られている
（→ [prune を paths の列挙・engine の判定と削除・CLI の同意解釈に分け、engine API を Reset と同型にする](../design/prune-orphan-series.md)）。この順を破った実装（`<roothash>` を一発で
`RemoveAll` する等）が途中で失敗すると、`.root` が先に消えた系列は以後の走査対象から外れ、
prune でも手動でも辿れなくなる。失敗の報告だけを見ていると規律の破れは観測できない。

## 実現性

**likelihood: medium** — 確認・dryrun の処理は `reset` の `Confirm` コールバックと同型に置く設計
（→ [prune を paths の列挙・engine の判定と削除・CLI の同意解釈に分け、engine API を Reset と同型にする](../design/prune-orphan-series.md)）で、既存のテストと規律に相乗りする。判定側と違い
blind spot ではないが、prune 固有の分岐（`--json` の `--yes` 必須・系列単位の try-lock）は新規で、
`reset` 側の変更が波及もしうる。

**impact: high** — 判定が正しくても削除が実行されてしまう脅威なので、失われるのは
[prune が生存系列や判定不能な系列を孤児と読み、不可逆に削除する](prune-misjudges-live-series.md) と違って「孤児と判定された系列」ではなく
ユーザーが止めるつもりだった削除の対象、あるいは lock 中 = 実行中の apply の足元にある世代。
どちらも不可逆で、確認を挟んだはずのユーザーには気付く機会が無かったぶん質が悪い。

## 緩和

[prune の dryrun が副作用ゼロで、非同意は中断し、lock 中の系列は skip され、削除段の失敗は進捗と理由で振り分けられる](../test/generations/prune-safety-gates.md)（prune の安全機構）が緩和する。ただし覆えるのは
engine 層まで（`Confirm` コールバックが false を返したときの中断・dryrun の非破壊性・lock skip・
削除中の失敗を `Removed` に数えないこと）で、TTY 判定と `--yes` / `--json` の解釈は CLI 側の
責務のため、同意の条件の残り半分は cli-json 区分の TC が担う
（[reset の teardown が管理外の実体まで消し、preview と同意の条件も機能しない](reset-teardown-overreach.md) と同じ分担）。

cli-json 側の受け皿は [破壊的操作の確認が対話性と同意フラグの組み合わせで決まり非対話で止まらない](../test/cli-json/confirm-policy-interactivity.md)（確認ポリシーと非対話フォール
バック。コマンド横断で書かれているので `prune` 追加で射程に入る）。`--json` が `--yes` 無しで
fail fast すること自体は同 TC の射程だが、同 TC の CASE が実際に `prune` を駆動対象へ足す
ことが条件で、足さなければ射程に入っているのは文面だけになる。確認プロンプトの中身
（root パス一覧を表示するか）はどの既存 TC も持たない残余。どちらも prune の CLI 側 CASE を
起こす段（→ #134）で埋める。

---
id: "RISK-a1622cdd-c7e2-4178-a50d-85bc2a35b29f"
type: risk
name: "巻き戻し中の 1 件の失敗が残りの巻き戻しを道連れにし、復旧できない項目も報告されない"
status: stable
threatens:
  - "REQ-9fca28c9-d3b1-4ad7-8f24-13b2ec7aeab2"
likelihood: medium
impact: high
level: high
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:28:26+09:00"
---
# 巻き戻し中の 1 件の失敗が残りの巻き戻しを道連れにし、復旧できない項目も報告されない

## リスク

undo ジャーナルの巻き戻しで個々の逆操作が失敗したとき、そこで巻き戻し全体が中断する。
本来の契約は [巻き戻し自体の失敗は best-effort で続行し、全件を stderr へ報告して停止する](../requirements/undo-rollback-best-effort.md) である。報告には元のエラーと巻き戻せなかった項目の一覧を
まとめる。

失敗する場面が現実にあるのがこのリスクの前提になる。逆操作の対象パスは、前方操作の
あとに別のプロセス・別のツールによって動かされうるし、巻き戻しを引き起こしたのと同じ
permission 障害が逆操作側でも起きることがある。中断すると、まだ巻き戻せたはずの項目が
FS に残り、リスク [途中失敗した run の巻き戻しが不完全で、FS に部分適用の痕跡が残る](undo-journal-rollback-incomplete.md) と同じ状態を別の処理で作る。

報告側の欠落も同じ重さで扱う。巻き戻せなかった項目が報告されないと、ユーザーは「巻き戻しは
全部成功した」と読んで手当てしないまま次の apply へ進む。この報告は失敗時の処理であり、
成功時沈黙の対象外として常時 stderr へ出す必要がある。journal が空のとき（FS 書き込みの
前に失敗した場合）に「N 件復旧できず」という虚偽の行を出すのも、同じ信頼性の毀損として
本リスクに含む。

## 実現性

**likelihood: medium** — 巻き戻し自体の失敗は正常系では起きず、故障注入で初めて実行される処理。
そのぶん実装の回帰が通常のテストをすり抜けやすい。

**impact: high** — 中断で巻き戻されずに残った変更は、前世代 manifest に記録が無いため次の
apply の古くなった配置の除去でも対象にならず（→ [途中失敗した run の巻き戻しが不完全で、FS に部分適用の痕跡が残る](undo-journal-rollback-incomplete.md)）、再実行では消えない。PreRemove が消した実
ディレクトリのように復元手段が無い項目も含む。復旧はユーザーの手作業になる。巻き戻せなかった
項目が報告されないことは、その手作業に着手する契機まで奪う。

## 緩和

[unwind が LIFO で巻き戻し、1 件の失敗でも継続して全件報告し、成功時は journal を破棄する](../test/atomicity/unwind-lifo-and-journal-lifecycle.md) のうち「best-effort 継続」と「全件報告（空 journal で虚偽の件数行を出さない
ことを含む）」の 2 条件が緩和する。同 TC の LIFO 順序と journal 破棄の条件は
[途中失敗した run の巻き戻しが不完全で、FS に部分適用の痕跡が残る](undo-journal-rollback-incomplete.md) 側の担当。

---
id: "RISK-5f3c98b8-bd78-4bee-95af-b35e6ade35f9"
type: risk
name: "project mode の配置物が ephemeral に保たれず、git の追跡対象へ漏れる"
likelihood: medium
impact: medium
level: medium
threatens:
  - "REQ-e79178f5-5865-4444-a05d-3ab06f33cd6d"
  - "REQ-60787ed2-4176-4bdd-800f-1600c0315551"
  - "REQ-1f128917-4424-4e37-8a88-e0bb23a09da7"
  - "REQ-eaa8c0df-af44-4f52-9603-cd2bc22a67e9"
---
# RISK-5f3c98b8-bd78-4bee-95af-b35e6ade35f9: project mode の配置物が ephemeral に保たれず、git の追跡対象へ漏れる

## リスク

project mode の配置物は commit の対象外で、activation は `.gitignore` に触れず git の状態に
干渉しない（→ REQ-e79178f5-5865-4444-a05d-3ab06f33cd6d）。利用者は `layat gitignore` の出力を
`.gitignore` へ取り込んで配置物を追跡から外す。これが崩れる形は次のとおり。

- **列挙の漏れ** — `gitignore` が `method` で target を選り分け、copy target を出さない
  （→ REQ-60787ed2-4176-4bdd-800f-1600c0315551）。copy の配置物は通常ファイルなので、
  `git add -A` でそのまま commit に入る
- **全件出力の揺れ** — 行指向の `gitignore --all` が config ごとの出力を並べたままにするか、
  並び替えと重複除去を欠く（→ REQ-1f128917-4424-4e37-8a88-e0bb23a09da7）。同じ target が二重に現れ、
  出力を取り込むたびに `.gitignore` の差分が揺れる
- **project mode 外での出力** — home / fixed の config を指定しても止まらず、git toplevel を
  前提にした anchor 形の行を出す（→ REQ-eaa8c0df-af44-4f52-9603-cd2bc22a67e9）。`.gitignore` の
  位置と無関係な行が書き込まれ、意図しないパスが追跡から外れる
- **activation の干渉** — activation が `.gitignore` を書き換えるか git の index に触れる。
  利用者の手元に、自分が書いていない変更が commit 候補として現れる

## 影響

impact を medium とするのは、漏れた配置物や書き換わった `.gitignore` が git の操作で元へ戻せる
ため。commit に入った配置物は `git rm --cached` で外せ、`.gitignore` の行は消せば済む。
push 済みの履歴に配置物が残る場合も、配置物は src から再生成できる内容で、利用者の実データは
失われない。likelihood を medium とするのは、列挙の対象を method や rootKind で絞る変更が
入りやすい一方、copy target を含むこと・project mode 外で止まることを確かめる検証条件が無いため。
likelihood medium × impact medium から level は medium。

## 緩和

TC-0c7281f8-cffc-4a53-8fe0-f52d9ea6bd12 が、名前の省略・単一・全件という対象選択の経路の区別を
緩和する。anchor 形への変換と重複除去は `cmd/layat/gitignore_test.go` が確かめているが、それを
束ねる検証条件は無い。copy target を列挙に含むこと、project mode 以外の config で止まること、
activation が `.gitignore` に触れないことを確かめる検証条件も無く、これらは残余として残る。

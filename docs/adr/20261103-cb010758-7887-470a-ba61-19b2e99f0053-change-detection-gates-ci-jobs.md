---
id: "ADR-cb010758-7887-470a-ba61-19b2e99f0053"
type: adr
name: "変更検出で nix を回さない PR の必須の確認を skip で通す"
status: 採用
origin: "ADR-0030 §2 と ADR-0027 §2b（docs だけの変更で nix を回さない）を書き直した"
justifies:
  - "INF-8b97573f-d4d6-4abf-85e2-d859afbd96c6"
  - "INF-d1230e1f-8ba8-49d8-8386-409bfbb7dd27"
references:
  - "ADR-bff01b65-ddb8-46ce-8ffe-ed33ad354511"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:28:06+09:00"
---
# ADR-cb010758-7887-470a-ba61-19b2e99f0053: 変更検出で nix を回さない PR の必須の確認を skip で通す

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-bff01b65-ddb8-46ce-8ffe-ed33ad354511
- 起点: ADR-0030 §2 と ADR-0027 §2b（docs だけの変更で nix を回さない）を書き直した

## 背景

docs だけを変える PR で、flake check の 3 環境の matrix と E2E を回すのは無駄になる。一方、main
へのマージには flake check と `e2e` の check を必須にしている
（ADR-bff01b65-ddb8-46ce-8ffe-ed33ad354511）。required status check は、生成されない check を
`Expected` のまま扱い、マージをブロックし続ける。ジョブが生成されたうえで `if:` によって skip
された check は、成功として扱われる。

## 決定

`test.yml` はトリガ段の `paths` フィルタを持たず、全ての PR で起動する。変更の検出は専用の
composite action にまとめ、`changes` ジョブがそれを呼ぶ。後続のジョブは `changes` の outputs を
実行の条件にし、対象外の PR では nix を実行せずに skip する。

実行の条件を置く段はジョブごとに分ける。matrix を持つ `flake-check` はステップ段の `if:` に条件を置き、
3 つの check を必ず生成する。`e2e` と `go-coverage` はジョブ段の `if:` に条件を置く。

ジョブごとの条件の段は INF-d1230e1f-8ba8-49d8-8386-409bfbb7dd27 に、required status check
との関係は INF-8b97573f-d4d6-4abf-85e2-d859afbd96c6 に置く。

## 根拠

skip した check が成功として扱われる挙動を使えば、docs だけの PR で nix を回さずに必須の check を
満たせる。検出のロジックを composite action に集めると、複数の workflow から使っても検出の
規則が 1 箇所にまとまる。

ジョブ段の `if:` で skip すると matrix が展開されず、ruleset が要求する `(os, system)` 付きの
3 つの check が生成されない。そのため `flake-check` だけはステップ段に条件を置く。

## 棄却した代替案

### トリガ段の `paths` フィルタで docs だけの PR を除外する

workflow ごと起動しないので必須の check が生成されず、docs だけの PR がマージできなくなる。

### 変更検出をやめ、全ての PR で nix を実行する

docs だけの PR でも 3 環境の flake check と E2E が実行され、CI の時間を浪費する。

### `paths` と `paths-ignore` で同じ名前の check を持つ workflow を 2 つ定義する

skip 用にも matrix を重ねて定義することになり、保守が二重になる。

### `flake-check` もジョブ段の `if:` に条件を置く

skip すると matrix が展開されず、必須の 3 つの check が `Expected` のまま残ってマージを
ブロックする。

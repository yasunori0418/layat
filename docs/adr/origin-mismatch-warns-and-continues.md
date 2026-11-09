---
id: "ADR-ab47f502-0fc8-4714-9af4-61e0a64732f8"
type: adr
name: "出所の食い違いは警告して apply を続ける"
status: stable
sources:
  - id: "issue-580"
    resource: "https://github.com/yasunori0418/layat/issues/580"
    title: "#580"
  - id: "issue-498"
    resource: "https://github.com/yasunori0418/layat/issues/498"
    title: "#498"
origin: "Issue #498（epic）の grilling（2026-10-31）。ADR-0070 §3 と §6 を書き直した"
justifies:
  - "REQ-ee5085a5-077b-4f5a-aa6a-841bc5515282"
references:
  - "ADR-767c33e5-73eb-4ef0-aba7-c87673fa5c2e"
  - "ADR-0db061d1-fa30-436d-91d7-c0cec8804c52"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-05T00:40:25+09:00"
---
# 出所の食い違いは警告して apply を続ける

- ステータス: stable
- 日付: 2026-11-03
- 関連: [同一 profileDir への同時実行は後勝ちにする](same-profiledir-target-last-wins.md), [last-applied 記録に出所を持たせる](record-carries-origin.md), GitHub Issue #580, #498
- 起点: Issue #498（epic）の grilling（2026-10-31）。ADR-0070 §3 と §6 を書き直した

## 背景

同じ profileDir への apply は後勝ちにしている（[同一 profileDir への同時実行は後勝ちにする](same-profiledir-target-last-wins.md)）。
欠けているのは、後勝ちが別の出所からの上書きとして起きていることを利用者が知る手段である。

同じ設定を home-manager と CLI 単体の両方から使う場合、generator と entrypoint
は使い方ごとに必ず違う。

## 決定

apply（`--dryrun` と世代スキップの処理を含む）は、
記録の出所と今回の出所が両方あって食い違えば、警告を出して続行する。後勝ちは変えない。

食い違いは、両方に `identity` があれば `identity` だけで比べ、どちらかに無ければ generator と
entrypoint で比べる。版 1 の記録は出所を持たないので比べない。警告のコードと判定の細目は
[apply は last-applied 記録に出所を持たせ、記録と今回の出所が食い違えば警告して続行する](../requirements/apply-origin-mismatch-warning.md) に置く。

## 根拠

共有を error にすると、モジュールから CLI 単体へ（またはその逆へ）移行する正当な操作も止まる。
警告なら移行は通り、意図しない共有だけが気付かれる。

`identity` を優先するのは、同じ設定かどうかを知っているのが利用者だけだからである。
両方の使い方で同じ `identity` を宣言すれば、往復するたびに警告が出ることは無い。版 1
の記録と推測で比べると、版を上げた直後の初回の apply で全員に誤った警告が出る。

## 棄却した代替案

### 食い違いを error にして止める

移行の正当な操作も止まり、解除のフラグが要る。後勝ちの方針を実質的に覆す。

### 食い違いで apply を止めるモードを設け、`--force` 系のフラグで解除する

止めないのを標準にすると使う場面が無く、止めるのを標準にすると error にするのと同じ問題が出る。

### `status` に出所を出す

気付けるのは `status` を実行した利用者だけで、上書きが起きた apply の時点では見えない。apply
で警告すれば足りる。

### link-farm の一致で警告を抑える

使い方が違えば、同じ設定でも link-farm の store パスが一致するとは限らない。逆に、
一致しても出所が違えば、上書きを知らせる意味は残る。

### 警告を抑えるフラグを設ける

activation と手元の実行の両方にフラグを渡す必要があり、意図しない共有の警告まで出なくなる。

### entrypoint から `identity` の既定値を導く

prebuilt 生成器を使う場合は entrypoint が無く、両方の使い方で同じ値にならない。既定値を暗黙に導くと、
一致・不一致の理由が利用者に見えなくなる。

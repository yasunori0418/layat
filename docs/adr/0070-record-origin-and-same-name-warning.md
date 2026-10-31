---
id: "ADR-0070"
type: adr
name: "記録の出所と同名 config の警告"
status: 採用
issues:
  - "#498"
origin: "Issue #498（epic: v1.0 前に残っている挙動の穴を埋める）の grilling（2026-10-31）"
justifies:
  - "REQ-ee5085a5-077b-4f5a-aa6a-841bc5515282"
revises:
  - "ADR-0058"
references:
  - "ADR-0013"
---
# ADR-0070: 記録の出所と同名 config の警告

- ステータス: 採用
- 日付: 2026-10-31
- 関連: ADR-0013（同一 profileDir は後勝ち）, Issue #498
- 改訂対象: ADR-0058 §2 の記録の内容（`origin` を足し `recordVersion` を 2 に上げる）と §4 の読みの規則（版 1 と 2 を受け付ける）
- 起点: Issue #498（epic: v1.0 前に残っている挙動の穴を埋める）の grilling（2026-10-31）

## 背景

home mode の config は名前だけで profileDir と last-applied 記録の鍵が決まる（→ ADR-0058 §1）。
home-manager モジュールは `apply --all --manifest` で prebuilt 生成器を使い、standalone の CLI は
entrypoint を発見して nix 生成器を使う。両者が同じ config 名（典型は `default`）を使うと、
profileDir と記録を共有する。片方の apply は、もう片方が置いた entry を「前回置いたが今回の
manifest に無い」ものとして stale 除去する。交互に実行すると互いの entry を除去し合うが、
target の衝突ではないので foreign symlink の warning（→ ADR-0015）の対象にもならず、何も出ない。

同一 profileDir の後勝ちは ADR-0013 §2 が決めた方針で、これ自体は覆さない。欠けているのは、
後勝ちが「別の出所からの上書き」として起きていることを利用者が知る手段である。

## 決定

### 1. 記録に出所を持たせる

last-applied 記録に `origin = {generator, entrypoint, identity}` を持たせる。`generator` は `nix`（nix 生成器が
link-farm を組んだ）か `prebuilt`（`--manifest` で渡された）。`entrypoint` は `nix` のとき発見した
entrypoint の絶対パスで、`prebuilt` では省く。`identity` は manifest のトップレベルの任意フィールド
`identity`（利用者が `mkManifest { identity }` か `layat.configs.<name>.identity` で宣言する設定の
同一性の識別子・自由文字列）の写しで、manifest が持たなければ省く。

### 2. `recordVersion` を 2 に上げる

apply は版 2 で書く。読みは 1 と 2 を受け付け、それ以外の版と未知のフィールドは従来どおり
error にする。版 1 の記録は origin 無しとして読み、出所の比較をしない。

### 3. apply は食い違いを警告して続行する

apply（`--dryrun` と generation skip 経路を含む）は、記録の origin と今回の origin が両方あって
食い違えば、stderr に 1 行と subject 警告 `W_LAYAT_ORIGIN_MISMATCH`（`detail = {recorded, current}`）を
出して続行する。後勝ち（ADR-0013）は維持する。

食い違いの判定は、記録と今回の両方に `identity` があれば `identity` だけで比べる（一致なら
generator / entrypoint が違っても警告せず、不一致なら同じでも警告する）。どちらかに無ければ
generator / entrypoint で比べる。

### 4. reset / rollback は記録済みの出所を引き継ぐ

reset と rollback が記録を書き直すときは記録済みの origin を引き継ぐ。記録に origin が無ければ
`origin` を省いて版 2 で書く。どちらの操作も新しい manifest を生成しないので、今回の出所を
持たない。

### 5. 版 1 しか読めない engine との非互換を受容する

版 1 しか読めない engine は版 2 の記録で error になる。home-manager が pin した `layatPackage` と
PATH 上の CLI のバージョンがずれる場合がこれに当たる。pre-1.0 の既知事項として受容する。

manifest の任意フィールド `identity` の追加では `schemaVersion` を 1 のまま据え置く。`identity` を
持たない manifest の形は変わらず、旧 manifest は新 CLI でそのまま読める。一方 `identity` 付きの
manifest は、未知のフィールドを拒む旧 CLI では読めない。これも同じく pre-1.0 の既知事項として受容する。

### 6. status への露出と停止モードは採らない

`status` は origin を出さない。食い違いで apply を止めるモード（`--force` 系で解除する形）も
設けない。

## 根拠

- **後勝ちを保ったまま見えるようにする**: 同名 config の共有を error にすると、モジュールから
  standalone へ（またはその逆へ）移行する正当な操作も止まる。警告なら移行は通り、意図しない
  共有だけが気付かれる。
- **出所は記録が持つのが自然**: 「この root に最後に置いたもの」を持つのが記録で、その出所も
  同じ事実の一部である。profile や世代に持たせると、世代管理が layat から出る終着点
  （→ ADR-0057）で読めなくなる。
- **版 1 と比べない**: 版 1 の記録には出所が無い。推測で比べると、版を上げた直後の初回 apply で
  全員に誤った警告が出る。
- **identity は利用者が宣言する**: 同じ設定を home-manager と standalone の両経路で共有すると、
  generator / entrypoint は経路ごとに必ず違い、往復するたびに警告が出る。同じ設定かどうかは
  利用者しか知らないので、利用者が名前を付け、両方にあればそれを優先する。
- **非互換の受容**: 読みを緩めて未知の版を「記録なし」に倒すと、誤った baseline で stale 除去
  する。ADR-0058 §4 の「止める」方針は変えず、版のずれは pre-1.0 の範囲で受け入れる。

## 影響

- 要求: REQ-ee5085a5-077b-4f5a-aa6a-841bc5515282 を起票する。REQ-9d622539-a1e5-4719-af51-e0cd85b66f35 の記録の内容と読みの規則、
  REQ-2ea19863-eaa2-466b-b1ed-3f56f6417c62 の警告コード一覧を改訂する。レイアウトは DSG-3c156bbe-e72b-43dd-8cb0-2cc105314ad5 に反映する。
  manifest の `identity` は REQ-dedd2c28-bba3-4ecf-80c9-8c77347e8e1f・REQ-60e6b49c-9ba1-4552-a0ec-d340421ec281・
  REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10 に足す。
- ADR-0058 と ADR-0013 に改訂注記を足す。

## 棄却した代替案

- **食い違いを error にして止める**: 移行の正当な操作も止まり、解除フラグが要る。後勝ちの方針
  （ADR-0013）を実質的に覆す。
- **`status` に出所を出す**: 気付くのは `status` を実行した利用者だけで、上書きが起きた apply の時点
  では見えない。警告を apply に出せば足りる。
- **鍵に出所を含めて profileDir を分ける**: 共有そのものは無くなるが、移行時に旧 profileDir と
  記録が孤立し、旧出所が置いた entry を誰も撤去しなくなる。
- **出所を profile（世代）に持たせる**: 世代管理が layat から出ると読めなくなる（→ ADR-0058 の
  根拠と同じ）。
- **link-farm の一致で警告を抑える**: 経路が違えば同じ設定でも link-farm の store path は一致すると
  限らない。逆に、一致しても出所が違えば上書きを知らせる意味は残る。
- **警告を抑えるフラグを設ける**: activation と手元の実行の両方にフラグを渡す必要があり、意図しない
  共有の警告まで出なくなる。
- **entrypoint から identity の既定値を導く**: prebuilt 経路には entrypoint が無く、両経路で同じ値に
  ならない。既定値を暗黙に導くと、一致・不一致の理由が利用者に見えなくなる。

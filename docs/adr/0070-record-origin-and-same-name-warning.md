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
home-manager モジュールは `apply --all --manifest` で prebuilt 生成器を通り、standalone の CLI は
entrypoint を発見して nix 生成器を通る。両者が同じ config 名（典型は `default`）を使うと、
profileDir と記録を共有する。片方の apply は、もう片方が置いた entry を「前回置いたが今回の
manifest に無い」ものとして stale 除去する。交互に走らせると互いの entry を除去し合うが、
target の衝突ではないので foreign symlink の warning（→ ADR-0015）にも掛からず、何も出ない。

同一 profileDir の後勝ちは ADR-0013 §2 が決めた方針で、これ自体は覆さない。欠けているのは、
後勝ちが「別の出所からの上書き」として起きていることを利用者が知る手段である。

## 決定

### 1. 記録に出所を持たせる

last-applied 記録に `origin = {generator, entrypoint}` を持たせる。`generator` は `nix`（nix 生成器が
link-farm を組んだ）か `prebuilt`（`--manifest` で渡された）。`entrypoint` は `nix` のとき発見した
entrypoint の絶対パスで、`prebuilt` では省く。

### 2. `recordVersion` を 2 に上げる

apply は版 2 で書く。読みは 1 と 2 を受け付け、それ以外の版と未知のフィールドは従来どおり
error にする。版 1 の記録は origin 無しとして読み、出所の比較をしない。

### 3. apply は食い違いを警告して続行する

apply（`--dryrun` と generation skip 経路を含む）は、記録の origin と今回の origin が両方あって
食い違えば、stderr に 1 行と subject 警告 `W_LAYAT_ORIGIN_MISMATCH`（`detail = {recorded, current}`）を
出して続行する。後勝ち（ADR-0013）は維持する。

### 4. reset / rollback は記録済みの出所を引き継ぐ

reset と rollback が記録を書き直すときは記録済みの origin を引き継ぐ。記録に origin が無ければ
`origin` を省いて版 2 で書く。どちらの操作も新しい manifest を生成しないので、今回の出所を
持たない。

### 5. 版 1 しか読めない engine との非互換を受容する

版 1 しか読めない engine は版 2 の記録で error になる。home-manager が pin した `layatPackage` と
PATH 上の CLI の版がずれる場合がこれに当たる。pre-1.0 の既知事項として受容する。

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
  全員に偽の警告が出る。
- **非互換の受容**: 読みを緩めて未知の版を「記録なし」に倒すと、誤った baseline で stale 除去
  する。ADR-0058 §4 の「止める」方針を崩さず、版のずれは pre-1.0 の範囲で受け入れる。

## 影響

- 要求: REQ-ee5085a5-077b-4f5a-aa6a-841bc5515282 を起票する。REQ-9d622539-a1e5-4719-af51-e0cd85b66f35 の記録の内容と読みの規則、
  REQ-2ea19863-eaa2-466b-b1ed-3f56f6417c62 の警告コード一覧を改訂する。レイアウトは DSG-3c156bbe-e72b-43dd-8cb0-2cc105314ad5 に反映する。
- ADR-0058 と ADR-0013 に改訂注記を足す。

## 棄却した代替案

- **食い違いを error にして止める**: 移行の正当な操作も止まり、解除フラグが要る。後勝ちの方針
  （ADR-0013）を実質的に覆す。
- **`status` に出所を出す**: 気付くのは `status` を打った利用者だけで、上書きが起きた apply の時点
  では見えない。警告を apply に出せば足りる。
- **鍵に出所を含めて profileDir を分ける**: 共有そのものは無くなるが、移行時に旧 profileDir と
  記録が孤立し、旧出所が置いた entry を誰も撤去しなくなる。
- **出所を profile（世代）に持たせる**: 世代管理が layat から出ると読めなくなる（→ ADR-0058 の
  根拠と同じ）。

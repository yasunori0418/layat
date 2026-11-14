---
id: "ADR-51863e94-6a9d-4029-95e7-a7ad2bfa93c1"
type: adr
name: "store path は out-link で受け取る"
section: generations
status: stable
origin: "engine / CLI の技術スタックの確定（2026-06-13）。ADR-0011 §4 を書き直した"
justifies:
  - "REQ-60c6b7ea-e936-4ce8-bd75-ad35e9c693b9"
  - "REQ-7a71a049-5876-4cfc-a65e-44e9a0349856"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# store path は out-link で受け取る

- ステータス: stable
- 日付: 2026-11-01
- 起点: engine / CLI の技術スタックの確定（2026-06-13）。ADR-0011 §4 を書き直した

## 背景

apply は link-farm を build して store path を受け取り、そこを指す symlink を配置してから世代をコミットする。
build してからコミットするまでの間に `nix-collect-garbage` が実行されると、まだどの世代からも参照されていない
link-farm が回収され、配置したばかりの symlink が dangling になる。

profileDir に素の symlink を置いても GC root にはならない。

## 決定

既定の nix 生成器の Build は、link-farm を profileDir の中の固定パス `.pending` に out-link を張って build し、
そのリンクから store path を返す。適用に成功したら、engine が `.pending` を消す。`--dryrun` は配置しないので、out-link を張らない。

out-link を張る段は [実行フローの順序は eval 先行 → flock → build とし build をロック内に閉じる](../requirements/exec-flow-eval-lock-build.md)、`--dryrun` の扱いは
[--dryrun は root を解決するが flock も pending gcroot も取らない](../requirements/dryrun-no-lock-no-gcroot.md) に置く。

## 根拠

out-link は store path の取得と gcroot の登録を 1 手で行う。build からコミットまでの間も link-farm が GC から
守られるので、配置した symlink が dangling になる窓が無くなる。コミット後は世代のリンクが link-farm を
参照するので、`.pending` は不要になる。

パスを固定すると、out-link は前回のものを上書きする。失敗で残っても config あたり 1 個を超えない。

`--dryrun` は配置しないので、dangling になる symlink を作らない。gcroot を張る理由が無い。

## 棄却した代替案

### `--print-out-paths` の出力から store path を読む

作業ディレクトリに `result` リンクを残さずに済むが、build からコミットまでの GC の窓を塞げない。

### 出力を読んでから `nix-store --add-root --indirect` で gcroot を張る

窓は塞げるが、取得と gcroot の登録が 2 コマンドに分かれ、out-link 1 手より冗長になる。

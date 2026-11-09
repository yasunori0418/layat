---
id: "DSG-3c156bbe-e72b-43dd-8cb0-2cc105314ad5"
type: design
name: "last-applied 記録は `<state>/layat/<鍵>/last-applied.json` に manifest をそのまま埋め込み、比較の基準は記録 → profile → 空の順に読む"
status: stable
satisfies:
  - "REQ-9d622539-a1e5-4719-af51-e0cd85b66f35"
  - "REQ-ee5085a5-077b-4f5a-aa6a-841bc5515282"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:53:31+09:00"
---
# last-applied 記録は `<state>/layat/<鍵>/last-applied.json` に manifest をそのまま埋め込み、比較の基準は記録 → profile → 空の順に読む

## 設計

```
<state>/layat/<name>/last-applied.json                 # home（--root なし）
<state>/layat/<roothash>/.root                         # backref（profileDir と同じ規則）
<state>/layat/<roothash>/<name>/last-applied.json      # project / fixed / --root 上書き
```

```json
{
  "recordVersion": 2,
  "linkFarm": "/nix/store/<hash>-layat-<name>",
  "appliedAt": "2026-10-11T12:34:56+09:00",
  "manifest": { ...配置した manifest.json をそのまま... },
  "origin": { "generator": "nix", "entrypoint": "/home/<user>/dotfiles/flake.nix", "identity": "dotfiles" }
}
```

`origin` は prebuilt 生成器（`--manifest`）なら `{ "generator": "prebuilt" }` で、`entrypoint` を
持たない。`identity` は配置した manifest の `identity` の写しで、manifest が持たなければ省く。reset / rollback が版 1 の記録を書き直す場合や、記録が無く profile から比較の基準を
取って書き直す場合のように、出所が分からないときは `origin` 自体を省く。

- **鍵の解決は `internal/paths`** が profileDir と同じ規則で行う（基底だけが
  `<state>/layat`）。backref `.root` も同じ階層に置く。
- **読み書きは engine が持つ**。読みは破損した JSON・必須フィールドの欠落や型違い・
  未知のフィールド・`recordVersion` が 1 でも 2 でもない記録を error にする（「記録なし」に
  倒さない）。`origin` がオブジェクトでも `null` でもなければ error で、`null` は `origin` 無しとして
  読む。`origin` の中の `generator` の欠落・各フィールドの型違い・未知のフィールドも error で、それ以外の
  値は読んだまま比較に使う。版 1 は `origin` 無しとして読む。書きは常に版 2。
  書きは temp + rename で atomic に行い、全 FS 操作の成功直後・世代 commit の前に呼ぶ。
  記録ファイルは rename 前に 0644 にする（root の種類で分けない。ディレクトリは無ければ 0755 で作り、umask に従う）。
  削除は reset の全撤去時だけ。
- **比較の基準の優先順位は 1 つの関数に集める**。apply（古くなった配置の除去と drift 修復）・rollback
  （離れる世代 N）・reset（撤去対象）が同じ関数で比較の基準を得る。記録があれば記録の
  manifest、無ければ profile リンクの current、どちらも無ければ空。両方あって `linkFarm` が
  profile の指す先と食い違えば記録を採り warning を 1 行出す。generation skip の判定は新
  link-farm を profile の current と比べ、記録があればその `linkFarm` とも比べる。どちらかが
  食い違うか profile が無ければ（記録だけがあっても）skip せず通常の処理に入る。
- **更新する処理**: apply の通常の処理と generation skip の処理（失敗規則は同じ）・rollback
  （FS 収束後・ポインタ移動前に N-1 の manifest と link-farm。失敗は undo journal で巻き戻す）・
  reset（撤去対象のうち target が消えた entry だけを除いた manifest。FS 操作の途中失敗でも
  書き直し、書き直しの失敗は撤去の error と併せて返す。`linkFarm` は保持。記録が無ければ
  profile の current の link-farm。失敗は error で undo journal は無い。profile が無く記録
  だけがあるときは flock の鍵である profileDir（roothash 系列は backref `.root` も）を apply
  と同じ手順で作ってから flock を取る）。`appliedAt` は書いた操作の時刻。rollback の target（N-1）は引き続き
  `profile-<N-1>-link` から読む。
- **出所の比較**: apply（`--dryrun` と generation skip の処理を含む）は記録の `origin` と
  今回の origin が両方あって食い違えば、stderr に 1 行と
  subject 警告 `W_LAYAT_ORIGIN_MISMATCH`（`detail = {recorded, current}`）を出して続行する。
  両方に `identity` があれば `identity` だけで、どちらかに無ければ `generator` / `entrypoint` で比べる。
  記録に `origin` が無ければ比べない。reset / rollback の書き直しは記録済みの `origin` を
  そのまま引き継ぐ。
- 記録を flock の鍵にしない。flock の鍵・prune の対象・`--json` の payload は従来どおり
  profileDir 側に残る。

## 出典

[last-applied 記録を profileDir と同じ鍵で state 基底に置く](../adr/last-applied-record-location.md)、
[last-applied 記録は manifest を内容ごと持つ](../adr/record-embeds-manifest.md)、
[記録を書き直す操作ごとに linkFarm の値を定める](../adr/record-rewrite-linkfarm.md)、
[古くなった配置の除去は、比較の基準を記録から優先して読む](../adr/baseline-prefers-record.md)、
[世代スキップには profile に加え記録の一致も要する](../adr/generation-skip-requires-both-match.md)、
[rollback の戻り先は世代リンクから読む](../adr/rollback-target-from-generation-link.md)、
[apply は世代の commit の前に記録を書く](../adr/apply-writes-record-before-commit.md)、
[rollback はポインタ移動の前に戻り先の manifest で記録を書き直す](../adr/rollback-rewrites-record-before-pointer.md)、
[reset は撤去できなかった entry を残して記録を書き直す](../adr/reset-rewrites-record-with-survivors.md)、
[記録の読みは不正な記録を error にする](../adr/record-read-strictly.md)。
`origin` と版 2 は
[last-applied 記録に出所を持たせる](../adr/record-carries-origin.md)、
[記録の読みは不正な記録を error にする](../adr/record-read-strictly.md)、
[出所の食い違いは警告して apply を続ける](../adr/origin-mismatch-warns-and-continues.md)、
[記録を書き直す操作は記録済みの出所を引き継ぐ](../adr/reset-rollback-inherit-origin.md)。

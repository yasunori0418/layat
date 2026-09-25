---
id: "TC-e42a5438-039b-41b3-9a39-0e8dc1f25cb3"
type: test_condition
name: "並列 build の失敗が config ごとに集約され並列度が --jobs を超えない"
mitigates:
  - "RISK-87484de7-e767-493c-a3a4-c699591832c9"
---
# TC-e42a5438-039b-41b3-9a39-0e8dc1f25cb3: 並列 build の失敗が config ごとに集約され並列度が --jobs を超えない

## 条件

`apply --all` の stage 1（先行並列 build）が、並列度を `--jobs` で抑えたまま、失敗を
取りこぼさずに stage 2 の集約へ渡すことを確かめる。

- **並列度の上限** — 同時に走る build が min(`--jobs`, config 数) に達し、それを超えない
- **失敗の集約** — stage 1 で失敗した config は stage 2 に渡らず、自身の subject に error で
  載り、failures に数えられ、stderr に報告される。成功 config の結果はそのまま残る。
  `--dryrun` でも同じ
- **集約の同一性** — `--jobs 1` と `--jobs N` で件数・エンベロープが一致し、stage 2 を
  並列度 1 で回したときの呼び出し順は辞書順である
- **`--jobs` の値域** — 負値はエラーで一括 eval 前に止まり、0 は論理 CPU 数に解決され、
  `--all` 無しの指定（明示の 0 を含む）はエラー。persistent flag ではない
- **`--debug` の識別** — stage 1 の nix 開示行に `[<name>] ` が付き、一括 eval の開示行は
  prefix 無しのまま

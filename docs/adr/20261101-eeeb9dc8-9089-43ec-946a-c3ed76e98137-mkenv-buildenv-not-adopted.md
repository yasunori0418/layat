---
id: "ADR-eeeb9dc8-9089-43ec-946a-c3ed76e98137"
type: adr
name: "layat.lib.mkEnv による buildEnv 合成環境は採らない"
status: 不採用
issues:
  - "#162"
  - "#537"
origin: "Issue #162 の job-plan grilling（2026-10-30）。ADR-0069 の本題を書き直した"
justifies:
  - "DSG-f1d4c295-54ae-4def-951c-901b42dea008"
references:
  - "ADR-0004"
  - "ADR-0040"
  - "ADR-0041"
---
# ADR-eeeb9dc8-9089-43ec-946a-c3ed76e98137: layat.lib.mkEnv による buildEnv 合成環境は採らない

- ステータス: 不採用
- 日付: 2026-11-01
- 関連: ADR-0004, ADR-0040, ADR-0041, GitHub Issue #162, #537, `docs/concept.md`
- 起点: Issue #162 の job-plan grilling（2026-10-30）。ADR-0069 の本題を書き直した

## 背景

Issue #162 は、`pkgs.buildEnv` で合成したパッケージ環境（bin / share / man のツリー）を任意の target に
置き、role ごとに `PATH` を管理する `layat.lib.mkEnv` を提案した。

現行の entry の `src` は derivation を取れるので、`src = pkgs.buildEnv { … }` と書けば、合成環境は
今のスキーマのまま配置できる。

## 決定

`mkEnv` と薄関数 `mkEnvEntry` を lib に足さない。buildEnv の合成環境は、entry の `src` に
`pkgs.buildEnv { … }` を置く配置 idiom で満たす。env.sh や `.envrc` などの配線ファイルは layat が
生成せず、利用者が書く。engine・manifest schema・passthru は変えない。

配置と配線の書き方は DSG-f1d4c295-54ae-4def-951c-901b42dea008 に置く。

## 根拠

`mkEnv` は 1 entry の mkManifest へ脱糖する設計で、engine・schema・passthru に差が出ない。足されるのは
buildEnv 呼び出しの短縮と、`binDir` などの派生パスだけである。派生パスは利用者が `target` に書いた
値から導けるので、新しい情報を持たない。API の表面積が増える一方、素の buildEnv の式より表現力は
狭まる（入れ子の buildEnv や `postBuild` など）。

hook の需要を engine に持ち込まず配線層の idiom で満たした ADR-0041 と同じ判断になる。`PATH` の管理は、
配置を既存のプリミティブで行い、環境変数の設定をシェル・systemd・direnv の機構に任せれば足りる。

再提案するときは、entry の配置 idiom で満たせない需要を具体的に示すことを前提条件にする。

## 棄却した代替案

### `mkEnv` を 1 entry の mkManifest への脱糖として足す

#162 が実装の対象にした案。ADR-0040 の root marker `subdir` と同じく lib 層の糖衣で、engine は
変えない。role ごとに独立した profile は得られるが、それは専用の config に 1 entry を置けば今でも
得られる。API の表面積に見合う価値が無い。

### 薄関数 `mkEnvEntry` を足す

entry を 1 つ返す関数。buildEnv 呼び出しの短縮しか足さず、buildEnv の引数をそのまま使える素の式に
劣る。

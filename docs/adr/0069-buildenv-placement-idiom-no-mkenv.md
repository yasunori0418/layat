---
id: "ADR-0069"
type: adr
name: "`mkEnv` を足さず、buildEnv の合成環境は entry の配置 idiom で満たす"
status: 採用
status_note: "不採用の記録"
issues:
  - "#162"
  - "#537"
origin: "Issue #162 の job-plan grilling（2026-10-30）"
justifies:
  - "DSG-f1d4c295-54ae-4def-951c-901b42dea008"
references:
  - "ADR-0040"
  - "ADR-0041"
---
# ADR-0069: `mkEnv` を足さず、buildEnv の合成環境は entry の配置 idiom で満たす

- ステータス: 採用（不採用の記録）
- 日付: 2026-10-30
- 関連: ADR-0040, ADR-0041, Issue #162, Issue #537, `docs/concept.md`
- 改訂対象: なし（新規提案の棄却判断）
- 起点: Issue #162 の job-plan grilling（2026-10-30）

## 背景

Issue #162 は `pkgs.buildEnv` で合成したパッケージ環境（bin / share / man のツリー）を layat で
任意の target に置き、role 別に `PATH` を管理する `layat.lib.mkEnv` を提案した。起票時
（2026-07-05）は README の Known limitations が「package install + PATH はスコープ外」と書いて
おり、それを撤回するスコープ変更として ADR を要した。現在その文は README に無く、スコープ文は
`docs/concept.md` の north-star 節「『関数ベースのパッケージ導入・PATH 追加』の具体機構は未定義」に
移っている。依存していた epic（#150 / #152 / #161 / #131）は全て完了済み。

## 決定

**`mkEnv` / `mkEnvEntry` を lib に足さない。** buildEnv の合成環境は、既存の entry の `src` に
`pkgs.buildEnv { … }` を置く配置 idiom で満たす。

- 配置は entry の `src = pkgs.buildEnv { … }` で書く。現行スキーマの `src` は derivation を取れるので、
  この書き方は今日そのまま動く
- env.sh / environment.d フラグメント / `.envrc` などの配線ファイルは layat が生成しない。
  利用者が自分の Nix 式やスクリプトで書く
- engine / manifest schema / passthru は変えない
- 配置と配線の書き方は design item（DSG-f1d4c295-54ae-4def-951c-901b42dea008）に idiom として置く

### 判断根拠

1. **`mkEnv` が足すのは糖衣と派生文字列だけ**。`mkEnv` は 1 entry の mkManifest へ脱糖する
   設計で、engine・schema・passthru（`rootKind` / `targets`）に差が出ない。足されるのは
   buildEnv 呼び出しの短縮と、`binDir` / `manDir` 等の派生パスの passthru だけ。派生パスも、
   利用者が `target` に書いた値から算術的に導けるので、新しい情報を持たない
2. **B 案の却下理由が `mkEnv` 自身にも当たる**。#162 では薄関数 `mkEnvEntry`（B 案）を
   「buildEnv 呼び出しの短縮以上の価値がなく、API 表面積だけが増える」として却下した。
   `mkEnv` もこの評価を同じ強さで受ける。doc-comments・docs site・テストの面が増える一方、
   素の buildEnv 式を書く経路より表現力は狭まる（nested buildEnv・`postBuild` 等）
3. **ADR-0041 と同型の判断**。ADR-0041 は hook 需要を engine に持ち込まず、配線層の idiom で
   満たした。`PATH` 管理も、配置は既存プリミティブ、環境変数の設定はシェル・systemd・direnv の
   機構で足りる。layat が配線を背負う理由が無い

## 影響

- **docs**: `docs/concept.md` north-star 節のスコープ文を「専用機構は持たない。buildEnv の合成環境を
  entry として配置し、環境変数の配線は利用者側で行う」に改める。README（英日）の Home mode 節に
  利用者向けの小節を足す
- **lib / engine / tests**: 変更なし。buildEnv を `src` に置けることは現行スキーマと既存テストが
  担保している
- **将来の再提案への基準**: 再提案するときは、entry の配置 idiom で満たせない需要を具体的に示す
  ことを前提条件とする

## 棄却した代替案

- **A 案: `mkEnv` = 1 entry の mkManifest への脱糖 + 派生パスの passthru**（#162 の実装対象案）:
  ADR-0040（root marker `subdir`）と同じく lib 層の糖衣で engine を変えない案。role 別の独立
  profile は得られるが、それは専用 config に 1 entry を置くだけで今日得られる。判断根拠 1・2 の
  とおり、API 表面積に見合う価値が無い
- **薄関数 `mkEnvEntry`**（#162 の B 案）: entry 1 個を返す関数。buildEnv 呼び出しの短縮しか
  足さず、buildEnv の引数を透過的に使える素の式より劣る

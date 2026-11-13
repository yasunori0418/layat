---
id: "DSG-d2e17f4f-0d32-45c1-8125-17e589664c85"
type: design
name: "動的 entry 生成のヘルパを lib に置く"
section: flake-outputs
status: deprecated
satisfies:
  - "REQ-d85f0cef-0f1e-4897-a841-41b61a8dae51"
  - "REQ-2b0c2bb8-964f-4e36-a121-c6ea0d4be1c4"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-10T00:44:25+09:00"
---
# 動的 entry 生成のヘルパを lib に置く

## 設計

この設計は棄却した。「リポジトリ内のファイルを列挙して entries を組み立てる」需要に対し、**lib は関数を
提供しない**。既に realise 済みの store パスまたは flake input を `builtins.readDir`
する idiom をドキュメント側（`docs/spec.md` 応用節）で示すに留める。

lib へヘルパを置かない理由は、置いた場合に **IFD（import from derivation）を
誘発する API になる**こと。列挙対象を「まだ realise していない derivation」で
受け取れる形にすると、評価中にビルドが実行される処理を lib の公開 API から呼べてしまう。

これは次の 2 つの要求の実現手段になっている。

- **[lib は nixpkgs.lib のみに依存する純データ生成器である](../requirements/lib-pure-data-generator.md)**: 列挙ヘルパは
  実質的に「FS を読む」機能で、純データ生成の枠を越える。API 面に置かないことで
  lib の役割が entries を受け取って manifest を組むことに限定され続ける
- **[mkManifest は配置データを生成する純粋関数である](../requirements/mkmanifest-pure-function.md)**: 動的生成を lib の内側へ入れると、
  `mkManifest` の入力が「評価時に確定したデータ」でなくなる余地が生まれる。
  列挙をユーザー側の評価コードに置けば、`mkManifest` が受け取る `entries` は
  常に確定済みの attrset になる

代償として、列挙の書き方はユーザーが自分で書くことになる。これを idiom として
ドキュメントで補うのが本設計の対になる措置。

## 出典

`docs/design.md`「flake.nix outputs 設計」の `lib` 公開 API 一覧末尾のコメント行。

> **本 item の根拠は ADR ではなく `docs/publication-roadmap.md` ①**: `listFilesInSrc` を
> lib から除去し readDir idiom へ降格する決定は同ロードマップのブロッカー ①（完了済み）が
> 持つ。`justifies` を受けていないのは張り忘れではなく、根拠が ADR 以外にあることによる。

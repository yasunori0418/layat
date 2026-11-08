---
id: "REQ-a33a11e3-830d-4142-88ed-4c1fc35e7f74"
type: requirement
name: "entry submodule のフィールドは src / subpath / target / method の 4 つとする"
status: stable
derives_from:
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  The value of each `entries` attribute SHALL be an entry submodule holding exactly four
  fields: `src` (`path | set | marker`, required), `subpath` (`string`, optional),
  `target` (`string`, optional), and `method` (`"symlink" | "copy"`, optional). `target`
  SHALL be a path relative to the `root` selected explicitly in `mkManifest`. The
  concrete default of each optional field, and which component validates and applies
  those defaults, are stated by REQ-b232ec98-af3b-41f3-a050-29d417322002, REQ-cb77ea05-bab8-4ccf-b09e-d23d8f71cdc7 and REQ-d1b5b3f5-10a0-400d-9f03-ba00c63d1c34 and are NOT
  restated here.
specification_ja: |
  `entries` の各属性値は entry submodule とし、フィールドは `src`
  （`path | set | marker`・必須）・`subpath`（`string`・省略可）・`target`
  （`string`・省略可）・`method`（`"symlink" | "copy"`・省略可）の 4 つでなければ
  ならない。`target` は `mkManifest` で明示選択した `root` からの相対パスでなければ
  ならない。
  各省略可フィールドの具体的なデフォルト値と、それを確認・適用する主体は
  REQ-b232ec98-af3b-41f3-a050-29d417322002 / REQ-cb77ea05-bab8-4ccf-b09e-d23d8f71cdc7 / REQ-d1b5b3f5-10a0-400d-9f03-ba00c63d1c34 が規定しており、本 item では重ねて
  規定しない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:36:56+09:00"
---
# REQ-a33a11e3-830d-4142-88ed-4c1fc35e7f74: entry submodule のフィールドは src / subpath / target / method の 4 つとする

## 仕様

`entries` の各値は entry submodule で、フィールドは次の 4 つ。

```
entry :: {
  src    : path | set | marker # 必須（type/marker によって挙動が変わる）
  subpath: string              # 省略可、デフォルト: "."（省略 = リポジトリ全体）
  target : string              # 省略可、デフォルト: 属性キー
  method : "symlink"
         | "copy"              # 省略可、デフォルト: "symlink"
}
```

上のコードブロックは原文の写しで、規範は frontmatter が正。各省略可フィールドの
デフォルト値そのもの（コメント欄の記述）は規範に含めない（理由は下の注記）。

`target` は配置先パスで、`mkManifest` の `root` で明示選択した基準からの相対パスで
指定する。省略可なので、属性キーをそのまま target とするのが canonical だが、キーを
論理ラベルにして `target` を明示上書きすることもできる（home-manager の `home.file` と
同型）。

各フィールドは `lib/types.nix` の entry submodule（`lib.types`）として定義され、
`mkManifest` の `evalModules` が確認・デフォルト適用する。この確認主体の規範は
[mkManifest 自身が evalModules で入力を確認する単一の箇所になる](mkmanifest-input-validation.md)（`mkManifest` が入力を確認する単一の箇所）と [normalizeManifest が確認・デフォルト適用・marker 変換を行い mkManifest が derivation を組む](normalize-manifest-defaults.md)（`normalizeManifest` が
確認・デフォルト適用を担う）が持つ。

各フィールドの値域・意味も個別 item が持つ（`src` は [src は path / set / marker の 3 種を取り store link を既定として out-of-store は marker で opt-in する](src-kinds-store-default.md)、`subpath` は
[subpath は src 内の相対パスとし、リポジトリ全体は省略で表して専用トークンを設けない](subpath-omission-whole-repo.md)、`target` の識別子性とデフォルトは [entry の識別子は属性キー = target とし一意性は Nix が担保する](entry-identity-by-target-key.md)、`method` の動作は
[method は配置方法を選び symlink は世代管理下・copy は世代管理外になる](method-src-matrix.md)）。本 item はフィールドの構成そのもの（4 つで過不足なく、どれが必須で
どれが省略可か）を規定する。5 つ目のフィールドが評価時エラーになること（strict）は
[entry submodule は strict とし未知キーと旧名を評価時エラーにする](entry-submodule-strict.md) の規範。

> **デフォルト値を規範に含めない理由**: `subpath` → `"."` / `method` → `"symlink"` /
> `target` → 属性キーの 3 つは、いずれも既存 item が規範として持つ（[normalizeManifest が確認・デフォルト適用・marker 変換を行い mkManifest が derivation を組む](normalize-manifest-defaults.md) が
> `normalizeManifest` のデフォルト適用として 3 つとも、[entry の識別子は属性キー = target とし一意性は Nix が担保する](entry-identity-by-target-key.md) が `target` の
> 分を）。本 item で重ねて規定すると規範の所在が二箇所になるため、frontmatter からは
> 落として委譲した。

## 出典

`docs/spec.md`「entries スキーマ仕様」→「フィールド定義（entry submodule）」、
および `#### target` の「root からの相対パス」。

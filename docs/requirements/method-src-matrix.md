---
id: "REQ-77689c68-953c-4cbb-ab31-1ac1e4f5f2fe"
type: requirement
name: "method は配置方法を選び symlink は世代管理下・copy は世代管理外になる"
status: stable
derives_from:
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-403fbe32-b146-401b-8b53-fe67c1e169c5"
  - "UC-01b896b4-04b9-40d0-bf9e-966eaf64c3d4"
specification: |
  The `method` field SHALL select how an entry is placed. With `method = "symlink"` and a
  store-backed `src` the entry SHALL be a read-only symlink into the Nix store, and with
  an out-of-store marker it SHALL be a live symlink to the local path; both SHALL be
  under generation management. With `method = "copy"` and a store-backed `src` the entry
  SHALL be a place-once copy that the user owns and may edit, and it SHALL NOT be under
  generation management.
specification_ja: |
  `method` フィールドは entry の配置方法を選ぶものでなければならない。
  `method = "symlink"` では、store 由来の `src` に対して Nix ストアへの読み取り専用
  symlink、out-of-store marker に対してローカルパスへのライブ symlink とならなければならず、
  いずれも世代管理下に置かれなければならない。`method = "copy"` では、store 由来の `src` に
  対して place-once コピー（書き込み可・ユーザー管理）とならなければならず、世代管理の
  対象外でなければならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:28:26+09:00"
---
# REQ-77689c68-953c-4cbb-ab31-1ac1e4f5f2fe: method は配置方法を選び symlink は世代管理下・copy は世代管理外になる

## 仕様

`method` は `"symlink"` | `"copy"`・省略可・デフォルト `"symlink"`。`src` の種別との
組み合わせで動作が決まる。

| method | `src` の種別 | 動作 | 世代管理 |
|---|---|---|---|
| `"symlink"` | path / set | Nix ストアへの symlink（読み取り専用）| あり（profile）|
| `"symlink"` | marker | ローカルパスへの out-of-store symlink（ライブ）| あり（リンク先のみ）|
| `"copy"` | path / set | place-once コピー（書き込み可・ユーザー管理）| **なし** |
| `"copy"` | marker | **eval 時エラー**（意図矛盾・`lib.throwIf`）| — |

上の表は原文の写しで、規範は frontmatter が正。最終行（`"copy"` + marker の eval 時
エラー）は規範に含めない。この規範は [意図が矛盾する組み合わせをクロスフィールドの確認で評価時に拒否する](normalize-manifest-cross-field-checks.md) が持ち、
本 item で二重に定義しないため。

旧名 `mode` からの改名理由（unix file mode との誤読回避・→ [entry の配置種別のフィールドを method と名付ける](../adr/mode-renamed-to-method.md)）は原文の注記で、
規範ではない。旧名が評価時エラーになることは [entry submodule は strict とし未知キーと旧名を評価時エラーにする](entry-submodule-strict.md) が持つ。

out-of-store symlink が世代管理下に置かれるといっても、版管理されるのは「どのリンク先を
指すか」のマッピングのみで、リンク先のファイル内容は対象外（原文の表の
「あり（リンク先のみ）」）。この限定の規範は [out-of-store symlink は marker の絶対パスまたは root 相対パスを指し、版管理はリンク先マッピングのみとする](out-of-store-link-mapping-only.md) の担当。

配置動作そのもの（symlink / copy の FS 操作・place-once の意味）と世代管理の機構は
[symlink 配置は親 dir を作り配置元/subpath を指すリンクを張り、foreign symlink は警告して後勝ちする](symlink-placement-procedure.md) / [copy は target 不在のときだけマテリアライズする place-once で世代管理の対象外とする](copy-place-once.md) / [世代は link farm derivation を layat 自前 profile へコミットして積み、前世代 manifest から古くなった配置を除去する](generation-mechanism-link-farm-manifest.md) 等の担当。本 item は method の値と、それが選ぶ
配置方法・世代管理の有無までを規定する。

## 出典

`docs/spec.md`「entries スキーマ仕様」→「フィールド定義（entry submodule）」→
`#### method`。

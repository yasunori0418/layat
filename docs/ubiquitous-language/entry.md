---
id: "UBQ-5caf1573-b91f-40dd-a414-58b5f49bf0d9"
type: ubiquitous_term
name: "entry / entries"
status: stable
avoid:
  - "file entry"
  - "ファイルエントリ"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-07T01:45:24+09:00"
---

# entry / entries

## Definition

A placement definition. `entries` is an attrset keyed by `target`, where each value is an entry (`{ src; subpath?; target?; method?; }`). The attribute key is the identifier and supplies the default `target`. There is no manual `name` field; key uniqueness is guaranteed natively by Nix attrset keys.

## Avoid

"file entry" (the source may be a directory); giving an entry a `name` field; calling `entries` a list of `{ name; … }` (the old form).

## 定義

配置定義。`entries` は target をキーとする attrset で、各値が entry（`{ src; subpath?; target?; method?; }`）。属性キーが識別子であり、既定の `target` を与える。手動の `name` フィールドは持たない。識別子の一意性は Nix の attrset キーで native に担保される。

## 避ける表現

「ファイルエントリ」（配置元はディレクトリのこともある）、entry に `name` フィールドを持たせること、`entries` を `{ name; … }` のリストと呼ぶこと（旧形）。

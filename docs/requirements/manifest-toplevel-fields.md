---
id: "REQ-dedd2c28-bba3-4ecf-80c9-8c77347e8e1f"
type: requirement
name: "manifest.json のトップレベルは必須の schemaVersion / root / entries と任意の identity とする"
status: stable
derives_from:
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  The top level of `manifest.json` SHALL consist of three required fields: `schemaVersion`
  (int, the contract version), `root` (object, the kind of the placement base), and
  `entries` (array of object, the placement definitions), and MAY additionally carry the
  optional field `identity` (non-empty string, the user-declared identifier of the config).
  When `identity` is not declared, the field SHALL be absent.
specification_ja: |
  `manifest.json` のトップレベルは `schemaVersion`（int・契約バージョン）・
  `root`（object・配置先基準の kind）・`entries`（array of object・配置定義）の
  必須 3 フィールドで構成しなければならず、加えて任意フィールド `identity`（空でない string・
  利用者が宣言する設定の識別子）を持ってもよい。`identity` を宣言しないときはフィールド自体が
  無くなければならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-10-31T21:49:26+09:00"
---
# REQ-dedd2c28-bba3-4ecf-80c9-8c77347e8e1f: manifest.json のトップレベルは必須の schemaVersion / root / entries と任意の identity とする

## 仕様

| フィールド | 型 | 説明 |
|---|---|---|
| `schemaVersion` | int | 契約バージョン。v1 は `1` |
| `root` | object | 配置先基準の kind |
| `entries` | array of object | 配置定義 |
| `identity`（任意） | string | 利用者が宣言する設定の識別子。宣言しなければ無い |

上の表は原文の写しで、規範は frontmatter が正。`schemaVersion` の値と engine の拒否
挙動は [manifest.json が Nix と engine の唯一の安定契約であり schemaVersion は 1 に固定する](manifest-single-stable-contract.md)、`root` オブジェクトの中身は [manifest.json の root は rootKind を持ち fixed のときだけ絶対パスを併記する](manifest-root-object.md)、`entries[]` の中身は
[manifest.json の entries は attrset を配列へ正規化し 5 フィールドを記録する](manifest-entries-array.md) が持つ。

```json
{
  "schemaVersion": 1,
  "root": { "rootKind": "project" },
  "entries": [
    {
      "srcKind": "store",
      "src": "/nix/store/abcd1234...-source",
      "subpath": "skills/nix",
      "target": ".claude/skills/nix",
      "method": "symlink"
    },
    {
      "srcKind": "outOfStore",
      "src": "/home/me/dotfiles",
      "subpath": "home/.config/nvim",
      "target": ".config/nvim",
      "method": "symlink"
    }
  ]
}
```

## 出典

`docs/spec.md`「manifest.json スキーマ（v1・Nix↔Go 契約）」→「トップレベル」・「例」。

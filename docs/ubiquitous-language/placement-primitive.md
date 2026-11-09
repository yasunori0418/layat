---
id: "UBQ-3046d454-1ff1-48e2-9b05-0aad11c52ea6"
type: ubiquitous_term
name: "placement primitive"
status: stable
name_ja: "配置プリミティブ"
avoid:
  - "placement framework"
  - "configuration management"
  - "配置フレームワーク"
  - "設定管理"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-07T01:40:14+09:00"
---

# placement primitive

## Definition

The core of layat: a pure function that places a Nix store path at a `root`-relative `target`. It is not hidden behind a module abstraction — users compose it directly.

## Avoid

"placement framework", "configuration management" (layat does not generate configuration).

## 定義

layat のコア。`root` 相対の `target` に Nix store のパスを配置する純粋関数。モジュール抽象の裏に隠さず、ユーザーが直接合成して使う。

## 避ける表現

「配置フレームワーク」「設定管理」（layat は設定を生成しない）。

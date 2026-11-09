---
id: "UBQ-f7799808-267e-40c4-aaa7-d85e568f11c4"
type: ubiquitous_term
name: "`standalone`"
status: stable
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T10:11:12+09:00"
---

# UBQ-f7799808-267e-40c4-aaa7-d85e568f11c4: `standalone`

## Definition

A word with two meanings. First, the CLI on its own (invoking the CLI directly): a launch form, not a placement mode. It is orthogonal to placement mode (home / project / system); the mode is decided by the `root` marker, and the CLI on its own can use both project mode and home mode. Second, home-manager's standalone configuration (using HM without the NixOS module), the sense used by `tests/e2e/scenarios/05-hm` and [home-manager モジュールの配線は build sandbox 内の評価アサートで検証する](../test-plan/hm-module-eval-assert.md).

Japanese prose writes the first as 「CLI 単体」 and keeps the second as the code spelling `standalone`. The template name `templates/standalone` and the command `layat init standalone` are identifiers for the first meaning and also stay in code spelling.

## Avoid

Short-circuiting "`standalone` = home mode" (the placement mode is decided by the `root` marker).

## 定義

2 つの意味を持つ語。1 つ目は CLI 単体（= CLI を直接実行する起動形態）で、配置モードではない。配置モード（home / project / system）とは直交し、モードは `root` マーカーが決める。CLI 単体から project mode も home mode も使える。2 つ目は home-manager の `standalone` configuration（HM を NixOS モジュール無しで使う構成）で、`tests/e2e/scenarios/05-hm` と [home-manager モジュールの配線は build sandbox 内の評価アサートで検証する](../test-plan/hm-module-eval-assert.md) がこの意味で使う。

日本語の散文では前者を「CLI 単体」と書き、後者はコード表記の `standalone` のまま書く。テンプレート名の `templates/standalone` と `layat init standalone` は前者の意味の識別子で、これもコード表記のまま書く。

## 避ける表現

「`standalone = home mode`」と短絡すること（配置モードは `root` マーカーが決める）。

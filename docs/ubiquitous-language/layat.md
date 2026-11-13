---
id: "UBQ-c3e66a6c-906d-4083-b7ba-19618539c56c"
type: ubiquitous_term
name: "layat"
section: core
status: stable
avoid:
  - "nput"
  - "Layat"
  - "LayAt"
  - "layAt"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-07T21:38:16+09:00"
---

# layat

## Definition

The name of the tool itself. A compressed coinage of "**lay** \<src\> **at** \<target\>": it places src at target, as the manifest says. It follows the traditional UNIX way of compressing a verb phrase into a command name, as `chroot` (change root) and `getopt` (get options) do. The tool was formerly named nput, whose "n" stood for *nix*; that prefix stopped describing the tool once the engine's only contract became `manifest.json`, regardless of who generates it (→ [ツール名を layat にする](../adr/tool-named-layat.md)). Spell it lowercase, including at the start of a sentence.

## Avoid

"nput" for anything other than the old name in migration context; "Layat" / "LayAt" / "layAt"; reading the name as nix-specific. The old name in migration context is written as a code span (`nput`), which the term-avoid check skips.

## 定義

ツールそのものの名前。"**lay** \<src\> **at** \<target\>"（manifest の言うとおりに src を target へ置く）の圧縮造語。`chroot`（change root）・`getopt`（get options）と同じ、動詞句をコマンド名へ圧縮する UNIX の伝統的な造語法に従う。旧名は nput で、その「n」は *nix* を指していたが、engine の契約が生成者を問わない `manifest.json` 1 本へ収束した時点でツールを記述しなくなった（→ [ツール名を layat にする](../adr/tool-named-layat.md)）。文頭でも小文字で表記する。

## 避ける表現

移行文脈の旧名以外で「nput」を使うこと、「Layat」「LayAt」「layAt」と書くこと、名前を nix 固有のものと読むこと。移行文脈で旧名を書くときは、term-avoid が見ない code span（`nput`）にする。

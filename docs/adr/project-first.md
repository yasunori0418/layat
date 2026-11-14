---
id: "ADR-560852e1-3d07-46b9-ae1a-d15f82d28a5e"
type: adr
name: "project mode を第一の使い方とする"
section: concept
status: stable
sources:
  - id: "issue-2"
    resource: "https://github.com/yasunori0418/layat/issues/2"
    title: "#2"
  - id: "issue-3"
    resource: "https://github.com/yasunori0418/layat/issues/3"
    title: "#3"
  - id: "issue-4"
    resource: "https://github.com/yasunori0418/layat/issues/4"
    title: "#4"
origin: "Issue #2（root デフォルト）, #3（CLI 化）, #4（flake.nix 以外の entrypoint）。ADR-0007 §1 を書き直した"
justifies:
  - "REQ-c1b3ca5f-d2f7-443c-bc4b-b18413ca97b9"
  - "DSG-4a84f282-76ea-47c9-aede-deac12ff5257"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-01T23:16:56+09:00"
---
# project mode を第一の使い方とする

- ステータス: stable
- 日付: 2026-11-01
- 関連: GitHub Issue #2, #3, #4, `docs/concept.md`
- 起点: Issue #2（root デフォルト）, #3（CLI 化）, #4（flake.nix 以外の entrypoint）。ADR-0007 §1 を書き直した

## 背景

layat は store の物を任意のパスへ置く。置き先が `$HOME` なら dotfiles の管理で、home-manager と
正面から競合する。root を `$HOME` に固定した既定はスコープを絞った利用に合わず（#2）、CLI と
flake 以外の entrypoint への要望（#3・#4）も出ていた。layat の中心にどの使い方を置くかを
決める必要があった。

## 決定

layat の中心となる使い方を、プロジェクトに組み込んで repo 内の任意のパスへ配置すること
（project mode）に置く。`$HOME` への配置と `/` への配置は、root マーカーで明示して選ぶ使い方と
位置付ける。モジュール（HM・NixOS・nix-darwin）への対応は、ほかのモジュールシステムの switch と
一括で動かしたい用途を受け持つものにとどめる。規範の細目は
[全モジュールと devShell は engine をキックするだけの配線とし、ネイティブ機構へ翻訳しない](../requirements/modules-are-engine-wiring.md) と [実装スコープを CLI 単体 + project mode + home mode + system mode と、home-manager / NixOS / nix-darwin モジュールとする](../design/implementation-scope.md) に置く。

## 根拠

project 内への配置は、既存のツールが埋めていない領域である。home-manager と `$HOME` を
取り合わずに済み、PATH 上の CLI から `layat apply <name>` で使う形とも素直に噛み合う。実装の
順序も、CLI と project mode を最初に完成させればよいと決まる。

## 棄却した代替案

### home mode を中心に置く

dotfiles の用途には合うが、home-manager と正面から競合する。layat 固有の価値が薄れる。

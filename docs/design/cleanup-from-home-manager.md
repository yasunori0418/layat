---
id: "DSG-0b94308c-6b15-46d1-854c-5945ae587b0d"
type: design
name: "配置と cleanup のアルゴリズムは home-manager の linkGeneration / cleanup を範として Go で再実装する"
section: core-logic
status: stable
satisfies:
  - "REQ-16aef46b-7bb8-4ca1-b962-e9f3ed1fd1d2"
  - "REQ-1be4d678-959c-44d7-a346-44bfd95af56e"
  - "REQ-c1b3ca5f-d2f7-443c-bc4b-b18413ca97b9"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:22:37+09:00"
---
# 配置と cleanup のアルゴリズムは home-manager の linkGeneration / cleanup を範として Go で再実装する

## 設計

engine の配置・古くなった配置の除去のアルゴリズムは、home-manager の `linkGeneration` /
`cleanup` を**参考にして Go で再実装する**。`home.file` そのものは再利用しない。

「範とする」対象は、世代由来のマニフェストを前世代と diff して消えた entry の
symlink だけを保守的に消すという**アルゴリズムの骨格**であり、実装コードでも
インタフェースでもない。

この選択が実現手段になっているのは次の点。

- **[全モジュールと devShell は engine をキックするだけの配線とし、ネイティブ機構へ翻訳しない](../requirements/modules-are-engine-wiring.md) という要求がある**以上、
  home-manager の機構へ委譲する選択肢は最初から取れない。委譲すれば HM の使い方だけ
  振る舞いが HM 由来になり、CLI 単体 / devShell と二重化する。したがって
  自前実装が前提になり、残る判断は「アルゴリズムをゼロから設計するか、既に
  実運用で枯れた設計を範とするか」になる
- **[古くなった配置の除去は前世代の記録通りを指す symlink のみに限り、copy は消さず孤児を警告する](../requirements/stale-removal-invariants.md) という不変条件**は、
  まさに home-manager の cleanup が長年扱ってきた問題で、範とする価値がある部分に
  当たる
- **前世代マニフェストとの diff → 配置 → コミットという順序**（[世代は link farm derivation を layat 自前 profile へコミットして積み、前世代 manifest から古くなった配置を除去する](../requirements/generation-mechanism-link-farm-manifest.md)）も
  同じ骨格に載る。layat 固有なのは、前世代の出所を layat 自身の profile に
  統一する点（ホストの `oldGenPath` に依存しない）

再実装であって移植ではないため、`home.file` の内部表現・HM のオプション体系・
activation の作法は持ち込まない。これらを持ち込むと lib / engine が
home-manager へ依存することになり、独立性が崩れる。

## 出典

`docs/design.md`「コアロジック設計」→「世代管理と state」節の箇条書き。

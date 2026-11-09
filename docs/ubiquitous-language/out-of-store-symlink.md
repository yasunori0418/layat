---
id: "UBQ-2340b0fd-6444-4946-94a6-43f874e37615"
type: ubiquitous_term
name: "out-of-store symlink"
status: stable
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-07T01:45:24+09:00"
---

# out-of-store symlink

## Definition

A live symlink to a local path, opted into only via an explicit marker. An explicit escape hatch, not a first-class feature. There are two markers. `layat.lib.mkOutOfStoreSymlink "/abs/path"` points to a local absolute path (for live editing of dotfiles under development).
For in-repo trees, the root-relative marker `layat.lib.mkRootRelativeSymlink "rel/path"` takes a path relative to the placement root; the engine resolves it against the root at runtime and creates a relative link from the target's parent.

## Avoid

Treating it as default behavior; producing it via implicit branching on the type of `src`.

"symlink" alone is ambiguous. Always disambiguate to either store link or out-of-store symlink. Out-of-store is expressed by `method = "symlink"` in `entries` (formerly `mode`) together with an out-of-store marker in `src`, but the default store link is itself realized as a symlink, so the bare word does not identify the kind.

## 定義

ローカルパスへのライブ symlink。明示マーカーでのみ opt-in する。明示的な退避路（escape hatch）であり、第一級機能ではない。マーカーは 2 つある。`layat.lib.mkOutOfStoreSymlink "/abs/path"` はローカルの絶対パスを指す（開発中の dotfiles をライブ編集する用途）。
リポジトリ内の実体には、root 相対マーカー `layat.lib.mkRootRelativeSymlink "rel/path"` を使う。src は配置 root からの相対パスで、engine が実行時に解決した root を基準に、target の親を起点とする相対リンクを張る。

## 避ける表現

デフォルト挙動として扱うこと、`src` の型による暗黙の分岐で生み出すこと。

「symlink」単独では曖昧。必ず store link か out-of-store symlink のどちらかに寄せる。`entries` の `method = "symlink"`（旧名 `mode`）+ out-of-store marker の `src` で out-of-store を指すが、デフォルトの store link も symlink で実現される。

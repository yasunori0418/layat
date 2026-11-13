---
id: "DSG-fd39b9ee-a5d2-46c2-bdcd-fc399a1e0296"
type: design
name: "rootRelative の LinkDest は target の親から src/subpath への相対パスを Entry だけから計算し、root が要るのは配置前の確認だけにする"
description: "rootRelative の LinkDest は target の親から src/subpath への相対パスを Entry だけから計算する純関数とし、root は配置前に配置元の実在を確かめるときだけ使う"
section: core-logic
status: stable
satisfies:
  - "REQ-a77181dd-a59c-40bb-855e-3430ad970eb6"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:53:31+09:00"
---
# rootRelative の LinkDest は target の親から src/subpath への相対パスを Entry だけから計算し、root が要るのは配置前の確認だけにする

## 設計

```go
// srcKind == "rootRelative"
LinkDest(e) = filepath.Rel(filepath.Dir(e.Target), filepath.Join(e.Src, e.Subpath))
```

| `target` | `src` | `subpath` | `LinkDest` |
|---|---|---|---|
| `.claude/skills/sara-docs` | `dev/skills/sara-docs` | `.` | `../../dev/skills/sara-docs` |
| `.claude/skills/x` | `dev/skills/x` | `sub` | `../../dev/skills/x/sub` |

- **`LinkDest` は root を受け取らない純関数のまま**。`target` と `src` はどちらも同じ root に
  対する相対パスなので、両者の相対関係は root に依存しない。store / outOfStore の返り値は
  変えない。
- **相対リンクが正しく解決される条件は「配置時点で `target` の祖先が実ディレクトリ」**。祖先が
  symlink なら相対リンクは symlink の先を起点に解決され、別の場所を指す。layat は foreign な
  祖先 symlink を conflict として止め、自己記録の古くなった祖先 symlink は配置前に除去する
  （[祖先 symlink は自己記録の古くなったものだけを配置前除去し、それ以外はエラーで停止する](../requirements/ancestor-symlink-nest-migration.md)）ので、配置の時点でこの条件は常に成り立つ。
- **planner・古くなった配置の除去・drift 修復・`--json` の `old` / `new` は手を入れない**。いずれも
  ディスク上の readlink 結果と `LinkDest` の文字列完全一致で判定するので、相対の文字列が
  そのまま比較の両辺に乗る。last-applied 記録は manifest をそのまま保持するので、前世代の
  rootRelative entry からも同じ `LinkDest` が再計算できる。
- **root が要るのは配置前の確認だけ**。`checkOutOfStore` は rootRelative entry について
  `filepath.Join(root, e.Src, e.Subpath)` を Lstat し、不在なら dangling を作らずに停止する
  （エラー文は `target` と解決後の絶対パスを含む）。`copyStructureMismatch` は copy 専用で、
  rootRelative と copy の併用は評価時に拒否されるので触らない。
- **manifest の `src` は root 相対文字列をそのまま書く**（`srcKind: "rootRelative"`）。
  `validate()` は `srcKind` / `method` / `rootKind` の既知値以外を明示エラーにする。
- **link farm の GC アンカーは張らない**。farm は `srcKind == "store"` の entry だけを対象に
  するので、rootRelative は自然に対象外になる。

## 出典

[root 相対マーカーのリンクを target 起点の相対パスで張る](../adr/root-relative-link-as-relative-path.md)、
[root 相対マーカーの基準点を manifest の配置 root にする](../adr/root-relative-base-is-placement-root.md)、
[srcKind の列挙値は schemaVersion を上げずに足す](../adr/srckind-added-without-version-bump.md)、
[root 相対で指定するパスすべてにパス安全性の確認を課す](../adr/path-safety-for-root-relative-inputs.md)、
[copy と out-of-store の併用を eval 時に拒否する](../adr/copy-with-out-of-store-eval-error.md)、
[src が target に重なる root 相対 entry を評価時に拒否する](../adr/reject-overlapping-src-and-target.md)、
[root 相対 entry で engine が root を使うのは配置前の実在確認だけにする](../adr/root-only-in-preflight.md)。

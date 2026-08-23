---
id: "REQ-42fe312c-927c-4da3-9346-f7ca2f3a58ed"
type: requirement
name: "prune は dryrun・root 一覧付き確認・try-lock skip・--json の --yes 必須で削除を守る"
specification: |
  `nput prune --dryrun` SHALL have no side effect and SHALL write the series it would
  delete — the `<roothash>`, the root path, and the `<name>` profiles under it — to stdout.
  Before deleting for real, `nput prune` SHALL always present the root paths of the series
  it is about to delete and SHALL ask for confirmation, which `--yes` SHALL skip. On a
  non-TTY without `--yes`, it SHALL abort rather than delete. Each series SHALL be locked
  with a try-lock before deletion; a series whose lock cannot be taken SHALL be skipped
  with a warning, and `nput prune` SHALL NOT wait for a lock to be released. `nput prune
  --json` SHALL require `--yes`, and SHALL fail fast with `status:"error"` and a non-zero
  exit when it is absent.
specification_ja: |
  `nput prune --dryrun` は副作用を持ってはならず、削除予定の系列（`<roothash>`・root
  パス・配下の `<name>` profile 一覧）を stdout へ出力しなければならない。実削除の前に
  `nput prune` は削除対象の root パス一覧を必ず提示して確認を求めなければならず、
  `--yes` はこれをスキップしなければならない。非 TTY で `--yes` が無いときは削除せず
  中止しなければならない。各系列は削除前に try-lock で lock しなければならず、lock を
  取れない系列は warning を出して skip しなければならない。lock の解放を待っては
  ならない。`nput prune --json` は `--yes` を必須とし、無ければ `status:"error"` +
  非ゼロで fail fast しなければならない。
derives_from:
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
---
# REQ-42fe312c-927c-4da3-9346-f7ca2f3a58ed: prune は dryrun・root 一覧付き確認・try-lock skip・--json の --yes 必須で削除を守る

## 仕様

**`--dryrun`** — 削除予定系列（`<roothash>` / root パス / 配下 `<name>` 一覧）を stdout へ出力して
終了する。副作用ゼロが preview である担保で、lock も取らない。

**確認** — 実削除の前に削除対象の **root パス一覧を必ず表示**し、`reset` と同型の確認を出す
（→ ADR-0021）。`--yes` でスキップできる。非 TTY では確認が取れないため、`--yes` が無ければ
中止する（`reset` の非 TTY 規律 → REQ-31dae599-f3a3-4bbe-b367-c955535265da と同型）。

root パス一覧を必ず見せるのは、out-of-store な root（リムーバブルディスク・ネットワーク
マウント上のプロジェクト）がアンマウント中に「一時的に実在しない」と誤判定されうるため。
機械側でマウントポイント判定等の区別は行わず、この表示が唯一の防波堤になる。

**try-lock skip** — 各系列の削除前に profileDir の lock を try-lock で取り、取れない系列は
warning を出して skip する。lock が取れない = その系列で engine が動いている = root が実在して
使われている、なので削除対象である可能性が極めて低い。blocking wait は「使用中の系列を消す
ために待つ」倒錯した挙動になるので持たない。

**`--json`** — 確認プロンプトは機械消費で扱えないため、`--json` は `--yes` を必須とし、無ければ
即 `status:"error"` + 非ゼロで fail fast する（`reset --json` の REQ-2a613337-7646-4ced-8807-e43bca18acf3 と同型）。
`--json` 出力が stdout を専有することは REQ-2353259f-5878-452a-8e11-3445de69abc2 の担当。

> 何を削除対象とみなすかは REQ-c44433a1-7ee7-459a-9aae-7cc42166876f が持つ。本 item は
> 「対象と判定したものを実際に消してよいか」を守るゲートだけを述べる。

## 出典

決定の実体は ADR-0034 §2「安全機構 = dryrun・確認プロンプト・flock・アンマウント caveat」。
`--json` の `--yes` 必須は同 §2 末尾の `--json` 対応と ADR-0043 §8 の帰結。

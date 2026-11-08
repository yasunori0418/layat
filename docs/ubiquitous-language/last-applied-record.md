---
id: "UBQ-7e8e239e-e11b-414e-b6e7-295520d6a1a9"
type: ubiquitous_term
name: "last-applied record"
status: stable
name_ja: "last-applied 記録"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-07T01:45:24+09:00"
---

# UBQ-7e8e239e-e11b-414e-b6e7-295520d6a1a9: last-applied record

## Definition

layat's own record of what it last placed under a root — a filesystem fact, not a generation. It lives at `<state>/layat/<key>/last-applied.json`, keyed like the profileDir (`<name>` for home without `--root`, otherwise `<roothash>/<name>` with the backref `.root`). It holds `recordVersion` (2; versions 1 and 2 are both read), `linkFarm`, `appliedAt`, `manifest` (the placed `manifest.json` embedded verbatim), and `origin` (the source of the most recent apply that wrote the record, `{generator, entrypoint, identity}`; `identity` is copied from the manifest when it declares one), so it does not depend on store GC. When the record has an `origin` that differs from the current one (compared by `identity` alone when both carry one), apply warns and continues (→ [ADR-ab47f502-0fc8-4714-9af4-61e0a64732f8](../adr/origin-mismatch-warns-and-continues.md)). apply (including generation skip) and rollback rewrite it after the filesystem operations succeed and before the generation operation; reset rewrites it after removal and deletes it once every entry is gone. The baseline for stale removal, drift repair, reset's removal set, and rollback's departing generation is taken from the record, else the profile's current generation, else empty; when the record and the profile disagree on `linkFarm`, the record wins with a one-line warning. On the first apply (neither exists) the baseline is empty and a one-line "no last-applied record" warning is printed. A broken record is an error, not treated as absent (→ [ADR-8e253f2e-c255-482a-a0f8-0837e122e80f](../adr/record-read-strictly.md)).

## Avoid

Calling the record a generation (the rollback target is read from the generation link); describing it as the flock key (that is the profileDir).

## 定義

「この root に layat が最後に置いたもの」という FS の事実を layat 自身が持つ記録（last-applied 記録）。世代ではない。置き場は `<state>/layat/<鍵>/last-applied.json` で、鍵は profileDir と同じ規則（`--root` なしの home は `<name>` 直、それ以外は `<roothash>/<name>` + backref `.root`）。内容は `recordVersion`（2。読みは 1 と 2 を受け付ける）・`linkFarm`・`appliedAt`・`manifest`（配置した `manifest.json` をそのまま（内容を変えずに）埋め込む）・`origin`（直近に記録を書いた apply の出所 `{generator, entrypoint, identity}`。`identity` は manifest が宣言していればその写し）で、store の GC に依存しない。記録に `origin` があり今回の出所と食い違えば（両方に `identity` があれば `identity` だけで比べる）、apply は warning を出して続行する（→ [ADR-ab47f502-0fc8-4714-9af4-61e0a64732f8](../adr/origin-mismatch-warns-and-continues.md)）。apply（generation skip を含む）と rollback は FS 操作の成功後・世代操作の前に、reset は撤去の後に書き直し、reset で全 entry が消えれば削除する。古くなった配置の除去・drift 修復・reset の撤去対象・rollback の離れる世代で使う比較の基準は記録 → profile の current → 空の順に取り、記録と profile の `linkFarm` が食い違えば記録を採って warning を 1 行出す。初回（どちらも無い）は比較の基準を空とし、「last-applied 記録なし」の warning を 1 行出す。破損した記録は「記録なし」として扱わず error にする（→ [ADR-8e253f2e-c255-482a-a0f8-0837e122e80f](../adr/record-read-strictly.md)）。

## 避ける表現

記録を世代と呼ぶこと（rollback の target は世代リンクから読む）、記録を flock の鍵と説明すること（鍵は profileDir）。

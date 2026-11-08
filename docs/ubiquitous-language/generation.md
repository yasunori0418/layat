---
id: "UBQ-10bcf077-1e99-456b-bb0a-c9b87daf526e"
type: ubiquitous_term
name: "generation"
status: stable
name_ja: "世代"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-07T01:45:24+09:00"
---

# UBQ-10bcf077-1e99-456b-bb0a-c9b87daf526e: generation

## Definition

The unit of rollback. The default generation backend, nixprofile, manages it on layat's own Nix profile (the `nix-env --profile <dir>` style). nixprofile commits (`--set`) and switches (`--switch-generation`, which rollback also uses) generations with `nix-env --profile <dir>`, and lists them by scanning the generation links (`profile-<N>-link`) in the profileDir read-only, without `nix-env`; pruning (`--delete-generations`) and store GC (`nix-collect-garbage`) are left to the user's nix commands.

## Avoid

Narrating it under a "stateless script" premise (reversed since the initial direction); managing it with the new `nix profile` CLI (which requires a profile-manifest and does not work on `nix-env --set` profiles).

## 定義

rollback の単位。既定の generation backend である nixprofile が layat 自身の Nix profile（`nix-env --profile <dir>` 式）に乗せて管理する。nixprofile はコミット（`--set`）と切替（`--switch-generation`。rollback もこれを使う）を `nix-env --profile <dir>` で行い、一覧は nix-env を経由せず profileDir の世代リンク（`profile-<N>-link`）を読み取り専用で走査する。間引き（`--delete-generations`）と store GC（`nix-collect-garbage`）はユーザーが nix コマンドで行う。

## 避ける表現

「stateless スクリプト」前提で語ること（初期方針からは覆っている）、新 `nix profile` CLI で管理すること（profile-manifest を要求し、`nix-env --set` 製の profile では動かない）。

---
id: "UBQ-1660e016-7a0a-4319-9346-466f759fe759"
type: ubiquitous_term
name: "generation backend"
status: stable
name_ja: "世代バックエンド"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-07T01:45:24+09:00"
---

# UBQ-1660e016-7a0a-4319-9346-466f759fe759: generation backend

## Definition

What the engine leaves recording, listing, and switching generations to. Its contract has five operations under `internal/generation`, with paths and integers as the only inputs and outputs: Commit (record a link-farm as a new generation and point the profile at it), List (list the generations, marking the current one), Switch (point the profile at an existing generation), UserBase (the base the user-state profileDirs are keyed under), and SystemBase (the base the system-mode profileDirs are keyed under). The nix implementation (nixprofile, `internal/generation/nixprofile`) is the default and only backend, and the only place that runs `nix-env`; its bases are `<state>/nix/profiles/layat` and `/nix/var/nix/profiles/layat`. The CLI always injects nixprofile, and the engine errors when given no backend. The engine does not care who manages generations; the profileDir layout (`profile` / `profile-<N>-link` / `.pending`) is the fixed part of the contract that the engine reads directly. There is no flag, environment variable, or `layat.toml` key that chooses a backend, and no setting that overrides its bases (→ [世代操作はパスと整数だけを受け渡す 5 操作の契約越しに行う](../adr/generation-backend-contract-paths-and-ints.md), [世代バックエンドを nix 実装 1 つに留める](../adr/single-nix-generation-backend.md), [state の基底を設定で上書きさせない](../adr/state-bases-not-configurable.md)).

## Avoid

Saying the engine runs `nix-env`; describing a way to choose the backend or a setting for its bases.

## 定義

engine が世代の記録・一覧・切替を委ねる相手（世代バックエンド）。契約は `internal/generation` に置く 5 操作で、入出力はパスと整数のみ。Commit（link-farm を新世代として記録し profile を向ける）・List（世代の一覧と current）・Switch（既存の世代へ profile を向ける）・UserBase（user state の profileDir の基底）・SystemBase（system mode の profileDir の基底）。nix 実装（nixprofile・`internal/generation/nixprofile`）が既定かつ唯一の実装で、`nix-env` を叩くのはここだけ（基底は `<state>/nix/profiles/layat` と `/nix/var/nix/profiles/layat`）。CLI は常に nixprofile を注入し、engine はバックエンドが渡されなければ error にする。エンジンは誰が世代を管理するかに関知しない。profileDir 内部のレイアウト（`profile` / `profile-<N>-link` / `.pending`）は契約の固定部分で、エンジンが直接読む。バックエンドを選ぶフラグ・環境変数・`layat.toml` キーも、基底を上書きする設定キーも持たない（→ [世代操作はパスと整数だけを受け渡す 5 操作の契約越しに行う](../adr/generation-backend-contract-paths-and-ints.md), [世代バックエンドを nix 実装 1 つに留める](../adr/single-nix-generation-backend.md), [state の基底を設定で上書きさせない](../adr/state-bases-not-configurable.md)）。

## 避ける表現

エンジンが `nix-env` を叩くと説明すること、バックエンドの選択手段や基底の設定キーがあると説明すること。

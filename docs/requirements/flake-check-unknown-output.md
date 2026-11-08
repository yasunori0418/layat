---
id: "REQ-d0aef5af-e922-400b-b250-ca38719c480b"
type: requirement
name: "layat カスタム output は nix flake check の unknown 警告を許容し主検証は nix build で行う"
status: stable
derives_from:
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  The consumer-side `outputs.layat.<system>.<name>` (a dedicated namespace that avoids
  polluting `packages`) causes `nix flake check` to emit
  `warning: unknown flake output 'layat'` while exiting 0; this SHALL be accepted as
  expected and harmless, and SHALL NOT break CI. `flake check` inspects the evaluation
  soundness of the attrset directly under `layat`, but SHALL neither build nor evaluate the
  `.<system>.<name>` derivations beneath it, so nothing is built by mistake. Transposing
  to `flake.layat.<system>` with the flake-parts module SHALL NOT remove the warning, since
  nix itself hardcodes the known outputs, and renaming the output SHALL NOT remove it
  either. The primary verification of a `layat` artifact SHALL therefore be
  `nix build .#layat.<system>.<name>`.
specification_ja: |
  consumer 側の `outputs.layat.<system>.<name>`（`packages` 汚染を避ける専用 namespace）は
  `nix flake check` で `warning: unknown flake output 'layat'` を出すが exit 0 になる。
  これは想定内・無害として許容しなければならず、CI を壊してはならない。`flake check` は
  `layat` 直下 attrset の eval 健全性は確認するが、配下の `.<system>.<name>` derivation は
  build も eval もしてはならない（誤 build しない）。flake-parts module で `flake.layat.<system>` へ
  transpose しても警告は消えない（nix 本体が known-output を hardcode するため）。
  output 名を変えても消えない。`layat` 成果物の主検証は
  `nix build .#layat.<system>.<name>` で行わなければならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:31:01+09:00"
---
# REQ-d0aef5af-e922-400b-b250-ca38719c480b: layat カスタム output は nix flake check の unknown 警告を許容し主検証は nix build で行う

## 仕様

**`nix flake check` と `layat` カスタム output**: consumer の
`outputs.layat.<system>.<name>`（`packages` 汚染回避の専用 namespace）は `nix flake check` で
**`warning: unknown flake output 'layat'` を出すが exit 0**（CI を壊さない・想定内・無害）。
`flake check` は `layat` 直下 attrset の eval 健全性は確認するが、**配下の
`.<system>.<name>` derivation は build も eval もしない**（誤 build しない）。
**flake-parts module で `flake.layat.<system>` へ transpose しても警告は消えない**
（nix 本体が known-output を hardcode するため）。output 名を変えても消えない
（`lib` 以外は unknown 警告）。`layat` 成果物の主検証は `nix build .#layat.<system>.<name>`
で行う。将来 upstream の flake-schemas（PR #8892）がマージされたら `schemas.layat` で
消す余地を残す。

専用 namespace を使う判断そのものは [config 名 default を慣例の解決先とし専用 layat 名前空間で packages を汚さない](default-config-name-and-namespace.md) の担当。

## 出典

`docs/spec.md`「CLI 仕様」→「再現性スタンス」節の blockquote「`nix flake check` と `layat`
カスタム output」。

決定の実体は [flake check の警告は無害として許容する](../adr/flake-check-warnings-tolerated.md)「flake check の警告は無害として許容する」（unknown 警告は想定内・主検証は `nix build`。
transpose しても警告が消えないことの確認を含む）。

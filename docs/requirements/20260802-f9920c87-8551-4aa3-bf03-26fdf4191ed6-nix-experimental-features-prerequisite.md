---
id: "REQ-f9920c87-8551-4aa3-bf03-26fdf4191ed6"
type: requirement
name: "nix experimental-features は nix 生成器の前提条件とし、自動付与せず案内エラーで停止する"
derives_from:
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  The nix generator, the default manifest generator, uses `nix eval` / `nix build` (the
  new CLI) internally, and therefore SHALL require that the user environment has already
  enabled `experimental-features = nix-command` (and additionally `flakes` for flake
  entrypoints). The nix generator SHALL NOT add `--extra-experimental-features`
  automatically, so that it does not silently override nix.conf / Determinate Nix /
  organizational policy settings. When nix returns a feature-not-enabled error, the nix
  generator SHALL report it as a missing prerequisite together with how to enable it, the
  CLI SHALL stop with an error that states the prerequisite and how to enable it, and the
  raw nix error SHALL NOT be swallowed.
specification_ja: |
  既定の manifest 生成器である nix 生成器は内部で `nix eval` / `nix build`（新 CLI）を
  使うため、ユーザー環境で `experimental-features = nix-command`（flake entrypoint は
  さらに `flakes`）が有効化済みであることを前提としなければならない。nix 生成器は
  `--extra-experimental-features` を自動付与してはならない（nix.conf / Determinate Nix /
  組織ポリシーの設定を黙って上書きしないため）。未有効で nix が機能未有効エラーを
  返したときは、nix 生成器が前提条件の未充足として有効化方法と共に報告し、CLI は前提条件と
  有効化方法を案内するエラーで停止しなければならず、生の nix エラーを握り潰してはならない。
---
# REQ-f9920c87-8551-4aa3-bf03-26fdf4191ed6: nix experimental-features は nix 生成器の前提条件とし、自動付与せず案内エラーで停止する

## 仕様

nix 生成器（既定の manifest 生成器）は内部で `nix eval` / `nix build`（新 CLI）を使うため、ユーザー環境で
`experimental-features = nix-command`（flake entrypoint はさらに `flakes`）が
**有効化済みであることを前提**とする。

nix 生成器は `--extra-experimental-features` を自動付与しない（nix.conf / Determinate Nix /
組織ポリシーの設定を黙って上書きしないため）。未有効で nix が機能未有効エラーを
返したら**前提条件と有効化方法を案内する分かりやすいエラーで停止**する（生の nix
エラーを握り潰さない）。

機能未有効の判定と案内文は nix 生成器の中に閉じ、`generator.Error` の
`Kind = PrerequisiteMissing` と `Guidance` で CLI へ渡す。表示の分担は
REQ-7a2f1ecf-4675-45aa-80c0-a8fc58db9edd の担当。

## 出典

`docs/spec.md`「CLI 仕様（一次 UX）」の blockquote「前提条件: nix experimental-features」。

決定の実体は ADR-0025「nix experimental-features 前提」。主語を nix 生成器に絞ったのは
ADR-0055「manifest 生成器の契約」§6。

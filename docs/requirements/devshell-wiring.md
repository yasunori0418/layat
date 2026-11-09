---
id: "REQ-a0bdf6db-6c0c-476c-916a-61ee4e4510d9"
type: requirement
name: "devShell は shellHook から engine を起動する配線で、シェル入室のたびに project mode で配置する"
status: stable
derives_from:
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  The devShell SHALL kick the engine from the `shellHook` of `devShells.<name>`, of the
  same form as the home-manager module. The timing of activation it supplies SHALL be
  entry into the shell: it SHALL be kicked the moment a shell is entered by `nix develop`
  or by direnv (`use flake`), and SHALL place within the project on the git toplevel as
  root. Because a `shellHook` runs at high frequency, the generation-skip short circuit of
  project mode, which makes an unchanged application a no-op, SHALL be a premise of this
  wiring. That it holds no placement logic of its own and supplies only the root and the
  timing of activation is stated by [全モジュールと devShell は engine をキックするだけの配線とし、ネイティブ機構へ翻訳しない](modules-are-engine-wiring.md), and that it pins `projectRoot` by
  [全モジュールは共通オプションの同一集合を公開し、entries は configs 経由・root はモジュールが pin する](module-common-options.md); neither is restated here.
specification_ja: |
  devShell は `devShells.<name>` の `shellHook` から engine を起動しなければならない
  （HM モジュールと同型）。供給する activation タイミングはシェル入室でなければならず、
  `nix develop` / direnv（`use flake`）でシェルに入った瞬間にキックされなければならず、
  git toplevel を root にプロジェクト内へ配置しなければならない。`shellHook` は高頻度で
  走るため、変更なしなら no-op になる project mode の世代スキップ短絡をこの配線の前提と
  しなければならない。配置ロジックを持たず root と
  activation タイミングだけを供給することは [全モジュールと devShell は engine をキックするだけの配線とし、ネイティブ機構へ翻訳しない](modules-are-engine-wiring.md)、`projectRoot` を pin することは
  [全モジュールは共通オプションの同一集合を公開し、entries は configs 経由・root はモジュールが pin する](module-common-options.md) の担当で、いずれも本 item では規定しない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T23:05:37+09:00"
---
# devShell は shellHook から engine を起動する配線で、シェル入室のたびに project mode で配置する

## 仕様

`devShells.<name>` の `shellHook` から engine を起動する配線。HM モジュールと同型で、配置
ロジックは持たず root（project mode の git toplevel）と activation タイミング（シェル入室）を
供給するだけ。root は `projectRoot` を pin。

```nix
devShells.default = pkgs.mkShell {
  shellHook = ''
    layat apply skills --no-wait
  '';
};
```

- `nix develop` / direnv（`use flake`）でシェルに入った瞬間にキックされ、git toplevel を root に
  プロジェクト内へ配置する
- `shellHook` は高頻度で走るため、project mode の**世代スキップ短絡**（変更なしなら no-op）が
  前提

> **上は原文の写しで、規範は frontmatter が正**。世代スキップ短絡そのもの（derivation 同一なら
> 世代を積まず lstat ドリフト修復だけ行う）は [project mode は世代を非公開にし、project / system mode は derivation 同一なら世代を積まず lstat ドリフト修復だけ行う](project-mode-generation-skip.md)、project mode の root 解決は
> [project mode の root は git toplevel から解決し、config 相対も CWD 相対も採らない](project-mode-root-resolution.md)、モジュールと devShell が一律配線であることは [全モジュールと devShell は engine をキックするだけの配線とし、ネイティブ機構へ翻訳しない](modules-are-engine-wiring.md)、`--no-wait` の
> try-lock 挙動は [flock の取得は既定 blocking とし --no-wait のときだけ try-lock でスキップする](flock-acquisition-mode.md) の担当。上のコードブロックの `layat apply skills --no-wait` は
> 使用例であり、devShell が `--all --project-root` を使うべきであることは [--all は root モードフィルタで対象 config を絞れる](apply-all-root-mode-filter.md)、
> devShell への CLI 同梱が project mode の canonical であることは [layat CLI は PATH 常駐の一次 UX で、project mode は devShell 同梱を canonical とする](cli-primary-ux-installation.md) が規定する。

## 出典

`docs/spec.md`「モジュール別動作仕様」→「devShell（配線）」節。

決定の実体は [devShell の shellHook で配線する](../adr/devshell-shellhook-wiring.md)
（devShell からの project mode キックと世代スキップ）と、devShell を配線層として位置づける
[root の明示を mkManifest 層で必須にする](../adr/root-explicit-at-mkmanifest.md)。

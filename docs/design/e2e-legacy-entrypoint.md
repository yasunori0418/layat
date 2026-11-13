---
id: "DSG-bb6e03fd-e1c4-47c4-bfa0-a73e9457b184"
type: design
name: "E2E に legacy entrypoint シナリオを置き、NIX_PATH を flake.lock の nixpkgs に pin して検証する"
section: test-strategy
status: stable
satisfies:
  - "REQ-c890ce4a-6528-4ab3-ac86-23d7aebff7da"
  - "REQ-da253cab-34d4-4d6e-96f0-de99e012b376"
  - "REQ-67095391-eab2-45d2-b75b-b428d481bcc2"
  - "TP-229b69c0-cf5e-4fb6-a353-27e5064d93e9"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:22:37+09:00"
---
# E2E に legacy entrypoint シナリオを置き、NIX_PATH を flake.lock の nixpkgs に pin して検証する

## 設計

[非 NixOS で動く主張を実 nix の一気通貫 E2E で検証する](../test-plan/e2e-harness-scope.md) がハーネスに課すシナリオ 5 件（project mode / home mode / 古くなった配置の除去 /
copy place-once・out-of-store / HM module）に加えて、**legacy entrypoint シナリオ**を
E2E に置く。

| 項目 | 内容 |
|---|---|
| entrypoint | `shell.nix`（passthru 形）|
| 検証する使い方 | `layat apply` / `layat apply --all` / 素の `nix-shell` 互換 |
| nixpkgs の供給 | `NIX_PATH` を flake.lock の nixpkgs に pin |

**このシナリオを足す理由**は、legacy entrypoint が flake の使い方とは**別の実行パスを通る**
ことにある。`mkShell` passthru 形（[legacy entrypoint は mkShell passthru 形を canonical とし CLI の attr path を分岐させない](../requirements/legacy-entrypoint-passthru-canonical.md)）は、
`nix build -f <ep> layat.<name>` という flake とは異なる attr path 解決を経る
（内部の共通化は [legacy entrypoint の分岐は attr path 組み立てに閉じ、nix 呼び出しヘルパを共通で再利用する](legacy-nix-invocation-reuse.md)）。attr path の同一性を CLI 内部で保っていても、
実際に nix が解決できるかは実 nix で叩かないと分からない。

**素の `nix-shell` 互換まで検証範囲に含める**のは、passthru 形を canonical にした
狙いが「`shell.nix` が mkShell を兼ねられる」ことだからである。`layat` を通さない
素の `nix-shell` が動かなくなっていたら、passthru 形を選んだ意味が失われる。

**`NIX_PATH` を flake.lock の nixpkgs に pin する**のは、legacy entrypoint が
[flake は pure eval で flake.lock が固定し legacy は impure を許容しユーザー責任とする](../requirements/reproducibility-stance.md) の通り NIX_PATH 依存の impure eval を許容する best-effort だからである。
テスト側で pin しないと、実行マシンのチャンネル状態でシナリオの結果が変わり、
再現性のない E2E になる。pin する先を flake.lock に揃えれば、flake の使い方の
シナリオと同じ nixpkgs で実行される。

[legacy entrypoint では相対 path の src が自動で store 化されない](../requirements/legacy-src-not-auto-stored.md) という性質も、
このシナリオで実際に踏まれる（fixture の組み方がそれを前提にする必要がある）。

## 出典

`docs/design.md`「テスト戦略」の E2E シナリオ列挙のうち legacy entrypoint の項。

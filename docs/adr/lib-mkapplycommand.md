---
id: "ADR-db293e0c-4789-45b7-93a9-adedc021344f"
type: adr
name: "lib.mkApplyCommand で事前展開した apply コマンドを組む"
status: stable
sources:
  - id: "issue-530"
    resource: "https://github.com/yasunori0418/layat/issues/530"
    title: "#530"
  - id: "issue-493"
    resource: "https://github.com/yasunori0418/layat/issues/493"
    title: "#493"
origin: "Issue #530 の grilling（2026-10-25）。ADR-0068 §1 を書き直した"
justifies:
  - "REQ-6c51facc-fa7a-4b5e-939b-44e7578d7a4f"
  - "DSG-e4d5db6b-5bac-4fd6-b2fb-ed07bdec30f5"
references:
  - "ADR-521765d7-7e31-42f4-b548-979185a7586a"
  - "ADR-88231ffc-4836-4252-be34-894944fd28c3"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T17:26:08+09:00"
---
# ADR-db293e0c-4789-45b7-93a9-adedc021344f: lib.mkApplyCommand で事前展開した apply コマンドを組む

- ステータス: stable
- 日付: 2026-11-03
- 関連: ADR-521765d7-7e31-42f4-b548-979185a7586a, ADR-88231ffc-4836-4252-be34-894944fd28c3, GitHub Issue #530, #493
- 起点: Issue #530 の grilling（2026-10-25）。ADR-0068 §1 を書き直した

## 背景

project mode の devShell は pin 版の CLI を同梱し（ADR-88231ffc-4836-4252-be34-894944fd28c3）、
`shellHook` で `layat apply <name>` を実行する。CLI はそのたびに entrypoint を cwd から探し、
`nix eval` と `nix build` を実行する。

HM / NixOS / nix-darwin のモジュールは、link-farm を評価時に作り、`apply --all --manifest` の
1 行を組んでこれらを省いている。ただしこの組み立てはモジュールの内部にあり、devShell からは
使えない。また `lib` は nixpkgs.lib だけに依存するので（ADR-521765d7-7e31-42f4-b548-979185a7586a）、
pin 版の CLI を `lib` の中から参照できない。

## 決定

`lib` に `mkApplyCommand` を公開し、config 名から link-farm への attrset を受けて、事前展開した
`apply --all --manifest` のコマンド行を組む。CLI は引数 `layatPackage` で受け取る。

引数・戻り値の規範は REQ-6c51facc-fa7a-4b5e-939b-44e7578d7a4f に、`lib/` の中での置き場所は
DSG-e4d5db6b-5bac-4fd6-b2fb-ed07bdec30f5 に置く。

## 根拠

farm を devShell の依存として先にビルドしておけば、devShell に入るたびの eval・build・entrypoint の
探索が要らなくなり、cwd にも依存しなくなる。

CLI を引数で受ければ、`lib` を nixpkgs.lib だけの依存に保ったまま、利用者が pin した CLI を使える。

## 棄却した代替案

### `layatPackage` を省いたときに `lib` が pin 版 CLI を使う

利用者は CLI を渡さずに済むが、`lib` が layat 自身の flake の packages に依存し、nixpkgs.lib だけに
依存する規約が崩れる。

### template に事前展開の starter 例を足す

template の shellHook は名指しの `layat apply <name>` を汎用の正準形として保ち、事前展開は README で
選択肢として示すにとどめる。

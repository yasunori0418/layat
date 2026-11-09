---
id: "RISK-c9197b73-483b-4d42-8541-49bee039f5ee"
type: risk
name: "フラグ解釈と nix 呼び出しの引数組み立てが仕様とずれ意図しない対象へ作用する"
status: stable
threatens:
  - "REQ-1cc080f6-ae91-4c1f-973e-b7054cfc0198"
  - "REQ-c2d44626-d8f4-446a-a80a-319a500129b4"
  - "REQ-d95b814f-aa7a-470e-9320-c14f9c14da7b"
  - "REQ-687e225f-5046-46db-88fb-f9e527a1e97a"
  - "REQ-6be1cbf1-6c6e-498b-8acb-7f4b80037169"
  - "REQ-9ed6b500-a11f-414e-a763-adb47c89f7d4"
  - "REQ-a480c183-40ce-4201-93b5-65a7a59c1b9e"
  - "REQ-205d744d-5a53-4511-bc09-892ba01d4e6f"
  - "REQ-496b1a07-5b74-416b-9e5f-3952b4c03737"
  - "REQ-535b811d-dfc5-4eac-92db-737e70eb5415"
  - "REQ-774cef80-2872-4ea1-937b-a0fbabc305a9"
  - "REQ-89c7baf9-9be0-417b-bd2d-b2e4edabe796"
  - "REQ-1767b250-b475-4276-a551-20dc79e75a30"
likelihood: medium
impact: medium
level: medium
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T23:36:39+09:00"
---
# フラグ解釈と nix 呼び出しの引数組み立てが仕様とずれ意図しない対象へ作用する

## リスク

CLI の表層は「どの対象に」「どの修飾で」作用するかを決める層であり、ここがずれると、正しい
ロジックが誤った対象に対して正しく動く。engine より下は無傷なので、テストが下層に偏っていると
見逃される。

ずれ方は 3 系統ある。

- **フラグの解釈** — 値を省略できるフラグが直後の位置引数を値として飲み込み、対象の名前が
  消える。あるいは省略時の既定値が配線されず、フラグを付けた意味が失われる。root の種別を
  絞るフラグどうしの排他が機能せず、複数指定のうち片方だけが利用者に知らせずに採用される
- **対象の選択** — 名前省略・明示・全件の 3 つの使い方の区別が崩れる。全件と修飾フラグの合成が
  機能しない
- **nix 呼び出しの引数組み立て** — entrypoint の探索順序（flake → legacy）がずれる、あるいは
  flake 形と legacy 形で attr path の組み立てが食い違い、存在しない attribute を参照する。
  テンプレート名の受理集合が緩むと、意図しない参照先を `nix flake init` へ渡す

サブコマンドの登録し忘れも同じ層にある。コマンドを足したときに機械可読出力の開始・公開へ
配線し忘れると、そのコマンドだけ契約から外れる。

attr path の解決も同じ層にある。名前を省いたときに `default` 以外へ解決する、manifest を
`layat` 名前空間ではなく `packages` に探す（→ [config 名 default を慣例の解決先とし専用 layat 名前空間で packages を汚さない](../requirements/default-config-name-and-namespace.md)）、flake の
`layat.<system>.<name>` と legacy の `-f` の使い方で attr path を取り違える
（→ [entrypoint は `layat.<name>` に named manifest を公開し CLI は形ごとの attr path で build する](../requirements/named-manifest-addressing.md)）、`apply --all` の rootKind を一括 eval で取らずフィルタの
振り分けを誤る（→ [apply --all は rootKind を 1 回の一括 eval で取る](../requirements/apply-all-batched-eval.md)）と、意図しない config を build する。
entrypoint や `layat.<name>` が見つからないときに別の config を推測するか空の manifest で続ける
（→ [要求された操作が成立しないときは CLI がエラーで停止し、暗黙のフォールバックを採らない](../requirements/cli-stop-conditions.md)）、`rollback` が `--all` を受け付ける
（→ [rollback は名指し必須で --all に対応しない](../requirements/rollback-named-only.md)）のも、利用者が求めていない対象へ作用する形である。
このうち flake と legacy の attr path の組み立てと、entrypoint が見つからないときのエラー（明示指定・
CWD の自動探索とも）は [entrypoint の探索順序と nix へ渡す引数列が形ごとに正しく組み立てられる](../test/cli-json/nix-argv-assembly.md) が、名前を省略した既定の使い方の区別は
[フラグの解釈と対象選択が仕様どおりで位置引数を飲み込まない](../test/cli-json/flag-parsing-target-selection.md) が覆う。省略した名前が `default` という名前そのものへ解決すること、
`layat` 名前空間、一括 eval からのフィルタの振り分け、`layat.<name>` が無いときの停止、`rollback --all` の
拒否を確かめる検証条件はこの risk に無く、残余として残る。

[config は Nix で書き nix build で評価する](../requirements/config-written-in-nix.md) は、CLI が発見するのは entrypoint のファイルで、その評価は `nix build` に任せることを定める。
探索や引数列がずれて別の config を評価する懸念は、組み立ての実装を置き換えても残るので `requirement` へ張る。

> **[entrypoint の形ごとに root 解決と配置の意味論が食い違う](entrypoint-root-resolution.md) との境界**。本 item が扱うのは CLI 表層で引数列を組み立てる過程そのもので、
> どの起点を選んでも同じ配置の意味論が成り立つかは [entrypoint の形ごとに root 解決と配置の意味論が食い違う](entrypoint-root-resolution.md) の担当。分岐を attr path の
> 組み立て 1 箇所へ閉じ込めるという設計選択そのものへの懸念は、設計を入れ替えれば消える性質の
> ため [entrypoint の形ごとに root 解決と配置の意味論が食い違う](entrypoint-root-resolution.md) が [legacy entrypoint の分岐は attr path 組み立てに閉じ、nix 呼び出しヘルパを共通で再利用する](../design/legacy-nix-invocation-reuse.md) へ張って表す。ただしその閉じ込めが実際に成立している
> ことの検証は、引数列を組み立てる層の話なので本 item 側の [entrypoint の探索順序と nix へ渡す引数列が形ごとに正しく組み立てられる](../test/cli-json/nix-argv-assembly.md) が持つ。

## 影響

意図しない config が適用される、あるいはコマンドが実行できない。impact を medium とするのは、
誤った対象へ作用しても、そこで行われるのは layat の通常の配置——記録外の実体は conflict で
停止し、置かれるのは layat 管理下の配置物——であり、引数の組み立てを直したうえでの `reset` と
再 apply が意図した状態へ収束させるため。フラグが位置引数を飲み込む種類の誤りでは対象の名前が
消えたことが可視の失敗として現れず、別の config へ作用したことに気づく契機が無い。この無音性は
本 risk の中心だが、気づいた後の回復可能性は下げない。

likelihood を medium とするのは、フラグの解釈・対象の選択・nix 引数の組み立てという 3 系統が
コマンドを足すたびに増え、機械可読出力への配線を含めて登録し忘れが入りうるため。

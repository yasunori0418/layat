---
id: "QA-a5f7f088-a459-4bb2-9674-82b1a4a52053"
type: quality
name: "マージ前の自動検証を必須にし、プラットフォーム差が効く層は全プラットフォームで通す"
status: stable
derives_from:
  - "SOL-9fcd1d6e-6204-42e6-92bb-1faf966f0b3e"
specification: |
  Every change to `main` SHALL be verified automatically before it is merged. The test
  layers whose behaviour varies by platform — nix evaluation and build among them — SHALL
  be verified on every platform the project declares as verified in CI, and the layers that
  do not SHALL be verified on at least one such platform. Passing that verification SHALL
  be a technical precondition of merging rather than a rule contributors are asked to
  follow, and no participant SHALL be able to bypass it. A change that does not affect the
  verified sources MAY skip the verification work itself, but SHALL NOT thereby become
  unmergeable.
specification_ja: |
  main へのすべての変更は、マージ前に自動的に検証されなければならない。振る舞いが
  プラットフォームによって変わるテスト層（nix の評価とビルドはこれにあたる）は、
  プロジェクトが CI での検証対象として宣言する全プラットフォーム上で検証されなければ
  ならず、そうでない層は、そのうち少なくとも 1 つのプラットフォーム上で検証されなければ
  ならない。検証の成功は、参加者に遵守を求める運用ルールではなくマージの技術的な必須条件で
  なければならず、いかなる参加者もこれを迂回できてはならない。検証対象のソースに影響しない
  変更は検証の実行自体を省いてもよいが、それによってマージ不能になってはならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:28:03+09:00"
---
# QA-a5f7f088-a459-4bb2-9674-82b1a4a52053: マージ前の自動検証を必須にし、プラットフォーム差が効く層は全プラットフォームで通す

## 仕様

layat は複数の OS / system で動くことを主張するツールであり、その主張は「手元の 1 環境で
通った」では裏づけられない。マージ前の自動検証を必須にし、プラットフォーム差が影響する層は
全プラットフォームで通す。

**全プラットフォームを要求する範囲を層ごとに分ける**のが要点になる。振る舞いが OS / system で
変わる層（nix の評価・ビルド）は全環境で通さなければ主張の裏づけにならないが、そうでない層まで
全環境へ広げるのは検証時間を払うだけで得るものが無い。どの層がどちらかは実装が決めるため、
規範は分岐の存在だけを固定して割り当ては基盤（INF）へ委ねる。

必須化を**技術的な条件**として課す点も同様に要点になる。「main へ直接コミットしない」
「PR 経由でマージする」といった運用ルールは、テストが失敗していてもマージできる状態を残す。
方針として求めるのはその状態を塞ぐことであり、bypass の余地を残さないことを含む。

検証対象のソースに影響しない変更（ドキュメントのみの変更など）まで検証を実走させる必要はない。
ただし「実走を省く」ことが「マージできなくなる」に転化してはならない — 省略の手段はこの制約と
両立するものに限られる。

「検証対象として宣言するプラットフォーム」の実体は INF-d1230e1f-8ba8-49d8-8386-409bfbb7dd27 が持つ。これは `flake.nix` の
`perSystem` が定義する system の集合とは一致せず、**CI で検証すると宣言した範囲**を指す。各
テストレベルの範囲・アプローチも本 item は規定せず、何をどこまでテストするかは test_plan
（`docs/test-plan/`）が、実装形は design が持つ。

## 出典

ADR-ab400e58-afbd-40a0-8eae-239d3cfda211、ADR-b07bbe19-b8a7-4401-98b0-5ccaf798ad05、
ADR-dfdcc588-2e23-4d3b-ae17-b508a8d85ff0、ADR-fb5e9bee-8132-431b-9948-6ad9995c9a9a、
ADR-f2e72cfa-f46e-46b7-9d50-a98d60d8980f、ADR-3eb35e63-5465-44cd-8299-c2ac3e6034de、
ADR-cb010758-7887-470a-ba61-19b2e99f0053、ADR-6f0383d9-0176-47c4-a9da-1a41e1fd3747、
ADR-bff01b65-ddb8-46ce-8ffe-ed33ad354511、ADR-8035e7c6-7dc5-4aa8-9a37-403998fa7f9b が置いた方針を、
基盤（INF）から分離して
quality item として立てたもの。具体的なジョブ構成・check 名・ruleset の設定は上記 INF が持つ。

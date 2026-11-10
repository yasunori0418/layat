---
id: "QA-83707ee2-5937-405d-ba7c-2d19109155d0"
type: quality
name: "規範の型は、それが誰を縛るかで決める"
description: "規範の型は、それが誰を縛るかで requirement・quality・test_plan から選ぶ"
status: stable
derives_from:
  - "SOL-9fcd1d6e-6204-42e6-92bb-1faf966f0b3e"
specification: |
  A norm that binds the product's behaviour and contracts as a user meets them SHALL be
  placed in a `requirement`; a norm that binds the development process, conventions and
  governance SHALL be placed in a `quality`; a norm that binds the scope, levels and
  approach of testing SHALL be placed in a `test_plan`. Between the two types under the
  `solution`, the `test_plan` question SHALL be asked first: an item that exists because
  something has to be verified SHALL be a `test_plan`, and `quality` SHALL take what
  remains. The norm itself SHALL be held by a `quality`, and the machinery that upholds it
  SHALL be held by an `infrastructure`.
specification_ja: |
  製品の振る舞いと契約を利用者から見える形で縛る規範は `requirement` に置かなければならず、
  開発の進め方・規約・ガバナンスを縛る規範は `quality` に置かなければならず、テストの範囲・
  レベル・方法を縛る規範は `test_plan` に置かなければならない。`solution` 直下の 2 型の間では
  `test_plan` の問いを先に当てなければならない。何かを検証する必要があるから存在する item は
  `test_plan` としなければならず、残りを `quality` としなければならない。規範そのものは
  `quality` が持たなければならず、それを支える仕組みは `infrastructure` が持たなければならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-10T23:06:41+09:00"
---
# 規範の型は、それが誰を縛るかで決める

## 仕様

`specification` / `specification_ja` の組を持ち、グラフの仕様側にぶら下がる型は 3 つある。
`use_case` の下の `requirement`、`solution` の下の `quality` と `test_plan` である。この 3 型を
機械的に見分ける仕組みは無い。形さえ正しければ誤った親も正しい親と区別が付かないので、誤った
型に置いた item も `sara check` を通る。RFC2119 キーワードの検査も助けにならない。キーワードの
有無は、空でない `specification` を持つ item なら型を問わず確認される。したがって型の選択は
判断であり、次の規則で決め、レビューだけで守る。

> その規範が**誰を縛るか**を問う。利用者が出会う製品を縛る規範は `requirement`、製品を作る
> 人と進め方を縛る規範は `quality`、何をどうテストするかを縛る規範は `test_plan` とする。

型ごとの判別は次のとおり。

| 型 | 規範が縛るもの | 親 | 判定の問い |
|---|---|---|---|
| `requirement` | 製品の振る舞いと契約。このリポジトリを読まない利用者にも観測できる。 | `use_case` | 外すと、layat の何らかの*使い方*が変わる。 |
| `quality` | 開発の進め方・規約・ガバナンス。機能に付くのではなく作業を横断し、それ自体は何かの検証を目的としない。 | `solution` | 外しても気付くのは開発者だけで、出荷物は変わらず、失われる検証も無い。 |
| `test_plan` | テストの範囲・レベル・方法。テスト対象へ手が届くようにすることを含む。 | `solution` | そもそも存在する理由が、何かを検証しなければならないことにある。 |

`solution` 直下の 2 型は、開発者の問いで重なる。どちらを外しても利用者には見えない。そこで
`test_plan` の問いを先に当てる。何かを検証する必要があるから存在する item は `test_plan` で、
残りが `quality` になる。

例: 終了コードの意味、`manifest.json` のフィールド構成、衝突したら止まることは `requirement`。
「`specification` は英語で書く」「CI で全プラットフォームの `flake check` を回す」は `quality`。
E2E ハーネスの範囲、*テストの*対象外とする宣言、テストできるようにする手当ては `test_plan`。

### 灰色領域

ほとんどの item は上の規則だけで決まる。決まらないのは次の場合で、並列のレーンで判断が割れない
よう、先例で決着させる。

**テストのために開いた API の面。** [lib.__internal で private helper をテストへ公開する](../test-plan/lib-internal-test-seam.md)（もとは
`requirement`）は Issue #239 で `test_plan` に決まった。製品が公開する実在の attrset なので契約に
見えるが、動機は評価テストを private helper へ届かせることだけにある。その `specification` は
SHALL NOT を、この面を安定性と後方互換の対象から*外す*ことに使っている。何かをテストできるように
するためだけに存在し、契約が負う保証を放棄している面は、テストできることが優先し `test_plan` になる。

**layat の作り方についての規範。** 「エンジンは標準ライブラリだけを使う」「CLI の第三者依存は
cobra に限り vendorHash で固定する」「`lib` は nixpkgs.lib だけに依存する」といった制約は進め方の
規則に見えるが、`requirement` のままにする。利用者が出会うアーキテクチャの境界だからである。
`lib` が nixpkgs.lib だけに依存するから、利用者は home-manager を引き込まずにどの Nix 環境へも
取り込める。判別するのは判定の問いで、規範を外すと layat の使い方が変わる。本当の `quality` に
あたるのは、*リポジトリの進め方*についての規則（レビューの慣行、CI の義務）のように、利用者の
誰にも観測できないものである。

**何かを対象外とする宣言。** *誰の*範囲かを問う。製品が何かを提供しないという宣言は
`requirement` である。「`--only` フラグは無い」「MVP に片付けのコマンドは無い」「2 つの設定を
同じ配置先へ向けるのは利用者の責任」がこれにあたり、利用者ができることの範囲を区切り、利用者は
その欠落に直接出会う。何かを*検証しない*という宣言は `test_plan` である。
[NixOS / nix-darwin モジュールで使うときの実 activate は E2E ハーネスの対象外とする](../test-plan/nixos-vm-test-future.md) は NixOS / nix-darwin モジュールの経路を
実際に activate することを E2E ハーネスの外に置く。製品はその機能を提供したままで、テストが
そこまで届かないだけである。

**エラーと警告の出力についての規範。** 「エラーは対処の案内を 1 行持つ」「成功は既定で何も
出さない」「警告は stderr へ出し、stdout は機械可読の出力に充てる」はどれも `requirement` で、
文言・出力先・出力の多さのどれを縛るかを問わない。出力は CLI の観測できる振る舞いの一部で、
利用者もパイプもそれに直接出会う。`quality` になるのは、どのコマンドの契約も依存しない、
リポジトリ全体の書き方の規約である場合に限る。

**バージョンと互換性の方針。** 「`schemaVersion` は 1 に固定する」「MVP は v1 だけを出力・受理
し、移行の仕組みを作らない」は `requirement` である。`manifest.json` の利用者が読む契約を縛り、
後者はさらにエンジンが実行時に何を受理するかも縛る。製品が出力・受理するものを変えずに、開発者が
いつバージョンを上げてよいかだけを定める方針なら `quality` になる。

**nix の作業手順についての規範。** 縛る作業手順が*利用者の*ものなら `requirement` である。
「`nix flake check` が `unknown flake output` を警告するのは受け入れ、主な検証は `nix build`
とする」は、利用者自身の flake が出すものを縛る。layat が利用者に採ってもらう出力の名前空間の
性質である。「flake の入口は純粋に評価し、legacy の入口は利用者の責任で非純粋に評価してよい」と
「`experimental-features` は CLI が肩代わりしない前提条件である」は、layat がそもそも動くために
利用者の nix 環境が満たすべきものを縛る。どれかがこのリポジトリ自身の CI についての規則なら
`quality` になる（守る規範ではなくそれを回すパイプラインについての規則なら `infrastructure`）。

### `quality` と `infrastructure` の区別

`quality` は規範を持ち、`infrastructure` はそれを支える仕組みを持つ。「マージ前に全プラット
フォームの `flake check` が通る」は `quality`、それを回す CI パイプラインは `infrastructure` で
ある。迷ったら、道具を差し替えてもその item が成り立つかを問う。成り立つなら規範である。

## 出典

[テスト計画の型を品質方針の型から分けて設ける](../sara-adr/quality-and-test-plan-types.md) と
[infrastructure を技術基盤として再定義し root から外す](../sara-adr/infrastructure-redefined-non-root.md) が置いた型の区分を、Issue #240 で判別の規則として
定め、灰色領域の先例（Issue #239 の決着を含む）とともに運用してきたもの。Issue #649 で
quality item として立てた。

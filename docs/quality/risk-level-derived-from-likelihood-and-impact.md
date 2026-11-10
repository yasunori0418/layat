---
id: "QA-b14e0a14-b9c9-4cc6-9b7a-0ce95d30d289"
type: quality
name: "risk の level は likelihood と impact から導き、手で決めない"
description: "risk の likelihood と impact を決まった尺度で採点し、level はその 2 つから導く"
status: stable
derives_from:
  - "SOL-9fcd1d6e-6204-42e6-92bb-1faf966f0b3e"
specification: |
  The `likelihood` of a `risk` SHALL be scored by how easily a regression that realises
  the threat gets in, and its `impact` SHALL be scored by how recoverable the threat is
  once someone notices it. A risk with several failure directions SHALL be scored by the
  heaviest direction, and a direction SHALL NOT be left out unless another risk
  explicitly holds it as its own. A risk that a verification gate silently stops
  verifying SHALL inherit the `impact` of what the gate guarded and SHALL name that source
  with a link. Whether anyone would notice the threat SHALL NOT be scored and SHALL be
  stated in the prose instead.
  The `level` SHALL be derived from `likelihood` and `impact` by the matrix in
  `dev/tests/risk-matrix.tsv` and SHALL NOT be judged by hand. The priority of a
  `test_condition` SHALL be derived from the highest `level` among the risks it mitigates,
  and a `test_condition` treated at a different priority SHALL state the reason in its
  prose.
specification_ja: |
  `risk` の `likelihood` は脅威を実現する退行がどれだけ入り込みやすいかで採点しなければ
  ならず、`impact` は気付いた後に回復できるかで採点しなければならない。失敗の向きが複数ある
  risk は最も重い向きで採点しなければならない。ある向きは、別の risk がその向きを自身のものと
  して明示的に持つときを除き、外してはならない。自動検証が何も確かめなくなる risk は、その検証が
  守っていた対象の `impact` を引き継がなければならず、引き継ぎ元をリンクで名指ししなければならない。
  気付かれにくさは採点してはならず、本文に書かなければならない。`level` は
  `dev/tests/risk-matrix.tsv` の表で `likelihood` と `impact` から導かなければならず、手で
  決めてはならない。`test_condition` の優先度は、`mitigates` 先の risk のうち最も高い `level`
  から導かなければならず、導いた優先度と違う扱いをする `test_condition` は、その理由を本文に
  書かなければならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-10T23:27:44+09:00"
---
# risk の level は likelihood と impact から導き、手で決めない

## 仕様

`docs/model.yaml` はどの `risk` にも 3 つのフィールドを求め、それぞれを `high` / `medium` /
`low` に限る。しかし enum だけでは値の選び方が何も決まらない。決めずにおくと、並列のレーンが
それぞれ自前の尺度を作り、item 群は何も順位付けしなくなる。`specification` の書き方と
`threatens` の張り方を規約として持つのと同じ理由である。以下の規則で尺度を固定し、`level` は
判断せずに導く。

### `likelihood`: 脅威を実現する退行がどれだけ入り込みやすいか

「対象のコードがどれだけ頻繁に変わるか」×「既存の検証がその退行をどれだけ見逃すか」で判断する。
頻繁に変わり、テストの死角にあるコードは `high`、安定していて機械的な検査がすでに覆うコードは
`low` である。

### `impact`: 実現したら回復できるか

| 値 | 懸かっているもの |
|---|---|
| `high` | 利用者に届き、その環境や実データを壊す。または、気付いても再実行や再ビルドでは回復できない状態を残す。 |
| `medium` | 出荷物が誤動作する（気付かれるかどうかを問わない）が、気付けば再実行や再ビルドで回復できる。 |
| `low` | 開発の中に閉じ、出荷物には届かない。 |

この軸が問うのは 1 つだけで、**誰かが気付いた後に回復できるか**である。誰かが*気付くかどうか*は
あえて採点しない（→ [沈黙性を risk のスコア軸にせず本文に書く](../sara-adr/silence-not-a-score-axis.md)）。気付かれにくさは実在し、記録する
価値もあるが、layat の脅威はほぼすべて気付かれにくい。配置は知らせずにずれていく。これを採点
すると尺度が飽和して何も順位付けしなくなる。代わりに item の本文に書く。そうすれば、item 群を
並べ替えるふりをせずに読める形で残る。`medium` の行に「気付かれるかどうかを問わない」と明記
するのも同じ理由である。以前の尺度は気付かれにくい脅威をすべて `high` に送っており、その読み方を
習慣で残してはならない。

行は上から読み、最初に当てはまる行を採る。行は主語で互いに重ならない。`high` と `medium` は
利用者が出会うものについて、`low` はリポジトリの外へ出ないものについての行である。したがって
開発の中に閉じる脅威は `medium` ではなく `low` になる。例外は後述の引き継ぎの規則だけである。

**失敗の向きが複数ある risk は、最も重い向きで採点する。** 向きごとに表の行へ当て、最も高い
行に落ちた向きで risk を採点する。後の読み手が軽い向きを全体と取り違えないよう、そのことを本文に
書く。平均を取ったり、最も起こりやすい向きで採点したりすると、その risk が警告するために存在
する、まさにその場合が抜け落ちる。

ある向きをこの判断から外してよいのは、**別の risk がその向きを自身のものとして明示的に持つ**
ときに限る。兄弟の item を確かめ、それを引用する。確かめずに外すと、重い向きを、範囲が実際には
そこまで届かない risk へ委ねることになり、重みが item 群から丸ごと抜ける。どちらの item もそれを
採点しない。

**黙って何も確かめなくなった自動検証は、守っていた対象を引き継ぐ。** 何も確かめずに成功を報告
するテスト群は開発の中に閉じるので、`low` の行が採りそうになる。そうさせてはならない。失われる
のは検証そのものではなく、それが与えていた保護であり、その保護の価値は守っていた対象の価値に
等しい。その検証が捕まえるはずだった退行として risk を採点する。後で引き継ぎ元を採点し直した
とき、それを引き継ぐすべての item へ辿れるよう、退行を持つ `risk` への Markdown リンクで引き継ぎ
元を本文に名指しする。検証が複数の退行を守るときは、引き継ぎ元にも最も重い向きの規則が働く。
パスの安全性の検査（`high`）と文書の形の契約（`medium`）の両方を覆うテスト群は `high` を
引き継ぐ。

### `level`: 2 つから導き、手で判断しない

**正は `dev/tests/risk-matrix.tsv` である。** このファイルが 9 つのセルを機械可読な形で並べる。
下の表は読み物で、データファイルが持つ関係を目で読める形に並べたものにすぎない。導出を変える
ときは TSV とこの表を同じコミットで直す（`dev/tests/test-categories.tsv` と `CLAUDE.md` の
8 区分の表の関係と同じ取り決め）。

| `likelihood` \ `impact` | `high` | `medium` | `low` |
|---|---|---|---|
| **`high`** | `high` | `high` | `medium` |
| **`medium`** | `high` | `medium` | `low` |
| **`low`** | `medium` | `low` | `low` |

この item の他の規則と違い、`level` の導出はレビューでは守らない。`dev/tests/risk-matrix.sh`
が TSV を読み、`docs/risks/` のすべての item をそれと突き合わせる。この検査は `nix flake check ./dev` の
`checks.risk-matrix` と CI の `sara` ジョブに組み込まれている（→ #303）。`sara check` は各
フィールドの enum を検証するが、3 つの間の関係は見ない。その隙間をこのテストが埋める。frontmatter
の 3 フィールドの純粋な対応づけで、解釈する散文が無いので、スクリプトが完全に判定できる。

セルと合わない `level` は item の欠陥であり、熟慮した上書きとしては扱わない。ある item について表が
おかしいと感じるなら、採点を誤っているのは `likelihood` か `impact` である。

### test_condition の優先度

`test_condition` は優先度をフィールドに持たない。優先度は `mitigates` 先の risk の `level` から
導き、複数の risk を `mitigates` するなら、そのうち最も高い `level` を採る。優先度を risk と
test_condition の両方に持つと二重管理になり、片方だけが古くなる。導いた優先度と違う扱いをしたい
例外があるときだけ、その根拠を test_condition の本文に書く。`level` と
違って優先度はフィールドを持たず機械検査の対象でもないので、例外を本文で扱える。`level` は理由を
書いても上書きできない。

## 出典

[risk の impact を検出後の回復可能性で採点する](../sara-adr/risk-impact-as-recoverability.md)、[複数の失敗方向を持つ risk は最も重い方向で採点する](../sara-adr/risk-heaviest-direction.md)、
[自動検証の形骸化は守備対象の impact を継承し、継承元をリンクで名指しする](../sara-adr/gate-risk-inherits-impact-by-link.md)、[沈黙性を risk のスコア軸にせず本文に書く](../sara-adr/silence-not-a-score-axis.md) が Issue #303 で
定めた採点の尺度と `level` の導出に、test_condition の優先度の導き方を合わせ、Issue #649 で
quality item として立てた。

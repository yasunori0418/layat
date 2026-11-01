---
id: "ADR-fb5e9bee-8132-431b-9948-6ad9995c9a9a"
type: adr
name: "キャッシュ投入は matrix ビルド"
status: 採用
origin: "参照: ~/src/github.com/yasunori0418/cryoflow の CI 構成（同一メンテナの実績構成）。ADR-0012 §4 を書き直した"
justifies:
  - "INF-d1230e1f-8ba8-49d8-8386-409bfbb7dd27"
references:
  - "ADR-0028"
---
# ADR-fb5e9bee-8132-431b-9948-6ad9995c9a9a: キャッシュ投入は matrix ビルド

- ステータス: 採用
- 日付: 2026-11-01
- 関連: ADR-0028
- 起点: 参照: ~/src/github.com/yasunori0418/cryoflow の CI 構成（同一メンテナの実績構成）。ADR-0012 §4 を書き直した

## 背景

layat を使う側（activation や CI）は layat のバイナリを必要とする。キャッシュが無いと、利用者の
手元で Go のビルドを実行することになる。どの system のビルドを、どうやってキャッシュへ投入するかを決める必要がある。

## 決定

キャッシュへの投入は、x86_64-linux・aarch64-linux・aarch64-darwin の matrix で `nix build .#layat` を
実行し、cachix のキャッシュ `yasunori0418` へ送って行う。各 system のビルドは独立したジョブで、
対象を選別するジョブは挟まない。投入の契機は ADR-0028 が決めている。

CI のジョブ構成は INF-d1230e1f-8ba8-49d8-8386-409bfbb7dd27 に置く。

## 根拠

キャッシュを置く system を、CI がテストしている system にそろえる。ネイティブのランナーで
ビルドするので、クロスビルドの設定が要らない。投入するパッケージは 1 つだけなので、未キャッシュの
ものを選別するジョブを挟むと往復のコストの方が高い。既にキャッシュにある store パスは
`cachix-action` が送らずに済ませる。

## 棄却した代替案

### nur-packages の matrix 選別の仕組みを踏襲する

複数パッケージを前提にした仕組みで、パッケージが 1 つの layat には過剰になる。

---
id: "SDR-e9de8d3d-fcb8-456d-a3af-2338f8080f4f"
type: sara_adr
name: "test_case の target を resource に改名する"
description: "test_case のテスト資産の正準表記を OKF の resource キーで持ち、値はリポジトリ根からのパスのままにする"
section: frontmatter-fields
status: stable
sources:
  - id: "issue-662"
    resource: "https://github.com/yasunori0418/layat/issues/662"
    title: "#662"
origin: "epic #662 の grilling（2026-11-07）で確定"
justifies:
  - "QA-6bf957d9-17d9-4660-92b7-ebd6eeb71a8c"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:08:50+09:00"
---
# test_case の target を resource に改名する

- ステータス: stable
- 日付: 2026-11-07
- 関連: GitHub Issue #662, `docs/model.md`, `dev/tests/test-doc-map.sh`
- 起点: epic #662 の grilling（2026-11-07）で確定

## 背景

test_case は、対応するテスト資産の正準表記を `target` フィールドに持つ。値はリポジトリ根からの
パスか、flake check の `checks.<name>` の形で、`dev/tests/test-doc-map.sh` が CASE とテスト資産の
1:1 を確かめるのに使う。

OKF には、item が説明する対象を指す `resource` キーがある。[OKF v0.2 と互換の frontmatter キーを item に足す](okf-compatible-frontmatter.md) と
決めたので、`target` が持つ情報を OKF の読み手にも読める名前にしたい。

## 決定

test_case の `target` を `resource` に改名する。値の規則は変えない。

- 値はリポジトリ根からのパス（ディレクトリは末尾にスラッシュ）か、`checks.<name>` の形にする
- OKF の `resource` は URI を想定するが、layat の `resource` は URI ではない。index.md のバンドル根は
  `docs/` だが、`resource` の根はリポジトリ根である。この違いは layat の拡張の意味論として扱い、
  `docs/model.md` に書く
- test_case 以外の型には `resource` を足さない

## 根拠

`target` と `resource` は同じ情報で、名前だけが違う。OKF と同じ名前にすれば、sara を持たない
読み手も「この item はどのファイルを説明しているか」を読み取れる。

値をリポジトリ根からのパスのままにするのは、テスト資産の列挙（`dev/scripts/test-inventory.sh`）と
同じ表記のまま比べるためである。`docs/` を根にすると全ての値が `../` で始まり、比べる
前に変換が要る。`checks.<name>` はファイルではないので、URI にすると規則が 2 つに割れる。

## 棄却した代替案

### `target` を残し、`resource` を併記する

同じ値を 2 つのキーに書くことになり、片方だけを直す食い違いが起きる。

### 値を OKF の URI に寄せる

`file:` などの形にすると、テスト資産の列挙と比べる前に変換が要る。flake check は URI で
表せる場所を持たない。

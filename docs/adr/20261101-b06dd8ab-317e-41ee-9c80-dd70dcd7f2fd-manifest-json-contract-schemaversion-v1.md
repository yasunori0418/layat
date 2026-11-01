---
id: "ADR-b06dd8ab-317e-41ee-9c80-dd70dcd7f2fd"
type: adr
name: "manifest.json を schemaVersion 1 の唯一の契約として固定する"
status: 採用
origin: "実装前のレビューで詰めた manifest 契約の細目。ADR-0006 の決定節（Nix↔Go の契約の項）・ADR-0013 §1・ADR-0015 §7 を書き直した"
justifies:
  - "REQ-79ce0a09-f9bd-4e61-ba7f-45fb5643137b"
  - "REQ-250d936c-1df0-491d-a7af-708f38b61f33"
  - "REQ-dedd2c28-bba3-4ecf-80c9-8c77347e8e1f"
  - "REQ-dd10d820-e453-4099-a47a-ffb9a7de02fb"
  - "REQ-0b0cd1e3-bfeb-45c1-978d-e2e11c568336"
references:
  - "ADR-0014"
  - "ADR-0059"
---
# ADR-b06dd8ab-317e-41ee-9c80-dd70dcd7f2fd: manifest.json を schemaVersion 1 の唯一の契約として固定する

- ステータス: 採用
- 日付: 2026-11-01
- 関連: ADR-0014, ADR-0059, `docs/spec.md`
- 起点: 実装前のレビューで詰めた manifest 契約の細目。ADR-0006 の決定節（Nix↔Go の契約の項）・ADR-0013 §1・ADR-0015 §7 を書き直した

## 背景

layat は、配置定義を Nix の lib が評価してデータにし、Go の engine がそのデータを読んで配置する
2 層で動く。2 層は別々に実装・テストされ、別々の版で配られることもある。層の境目で何を受け渡すかを
決め、lib・engine・テストが同じものを見て作業できるようにする必要があった。

## 決定

`lib.mkManifest` が生成する `manifest.json` を、Nix と engine の唯一の安定契約にする。`manifest.json` は
`schemaVersion` を持ち、その値は `1` に固定する。engine は自身の対応版より新しい `schemaVersion` を
拒否する。v1 の後継版を切るときの移行方式は、そのときに別の決定として決める。

トップレベル・`root`・`entries` の各フィールドは REQ-dedd2c28-bba3-4ecf-80c9-8c77347e8e1f・
REQ-dd10d820-e453-4099-a47a-ffb9a7de02fb・REQ-0b0cd1e3-bfeb-45c1-978d-e2e11c568336 に置き、
版の扱いは REQ-79ce0a09-f9bd-4e61-ba7f-45fb5643137b・REQ-250d936c-1df0-491d-a7af-708f38b61f33 に置く。

## 根拠

契約が 1 点に閉じていれば、lib は Nix の評価テストで、engine は Go のテストで、互いを起動せずに
検証できる。契約を実装の後で Go の構造体に合わせて決めると、lib・engine・テストの並行作業に
手戻りが出る。先に item へ固定しておけば、3 者が同じ一次情報を見る。

`schemaVersion` があれば、古い engine が新しい形式の manifest を読み違えることを、明示的なエラーで
止められる。版を上げずに列挙値を足したときは、engine が未知の値を拒否する検査で同じように
止める（ADR-0059）。後継版がまだ無いうちに移行方式を設計しても、必要になる形を当てられないので、v1 だけを
対象にする。

## 棄却した代替案

### manifest の形を実装時に Go の構造体と擦り合わせて決める

契約が最後に決まるため、lib と engine とテストを並行して作れない。

### JSON Schema のファイルを一次情報にする

文書とコードの二重管理は減るが、仕様の段階で JSON Schema の記法を持ち込むのは重い。item の
規範文と JSON 例で足りる。

### v1 の段階で後継版への移行方式まで決める

まだ存在しない版の形を推測して設計することになる。

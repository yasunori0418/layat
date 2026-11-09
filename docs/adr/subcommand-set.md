---
id: "ADR-a51897c5-f709-4b7a-8db2-61e6ebb0040c"
type: adr
name: "独立した操作はサブコマンドにし、操作の修飾はフラグにする"
status: stable
sources:
  - id: "issue-550"
    resource: "https://github.com/yasunori0418/layat/issues/550"
    title: "#550"
origin: "Go エンジン化の設計（2026-06-11）で確定。ADR-0006「CLI 構造（サブコマンド体系）」節を書き直した"
justifies:
  - "REQ-02a33511-0941-4813-b289-a05eb8e9aa57"
references:
  - "ADR-acc31876-df36-485e-b356-92ae60e26db7"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:28:06+09:00"
---
# 独立した操作はサブコマンドにし、操作の修飾はフラグにする

- ステータス: stable
- 日付: 2026-11-01
- 関連: [reset は非対話の実行で --yes を必須にする](reset-confirmation-and-yes.md), GitHub Issue #550
- 起点: Go エンジン化の設計（2026-06-11）で確定。ADR-0006「CLI 構造（サブコマンド体系）」節を書き直した

## 背景

当初の CLI は `layat --rollback` のようなフラグ式で、配置・世代の切り替え・一覧・`.gitignore` 向けの
出力といった性格の違う操作を、同じ階層のフラグで表していた。engine を Go のバイナリにするにあたり、
操作をどう並べるかを決め直す必要があった。

## 決定

独立した操作（apply・reset・rollback・list-generations・gitignore など）はサブコマンドにする。
ある操作の挙動を変えるだけのもの（`--dryrun` など）は、サブコマンドにせずその操作のフラグにする。
操作を追加するときもこの区別に従う。

dry-run を apply のフラグとして持つ規範は [apply --dryrun は読み取り専用で conflict 検出時に非ゼロ終了する](../requirements/apply-dryrun-readonly.md) に置く。

## 根拠

操作ごとにサブコマンドを分けると、操作ごとに受け付ける引数とフラグを分けて定義でき、help も操作
単位で読める。

dry-run は操作の修飾の典型である。apply を副作用なしで実行して見せるモードで、独立した操作では
ない。apply のフラグにすると、apply と同じ config の選択や `--all` がそのまま使え、サブコマンドの数も増えない。

## 棄却した代替案

### フラグ式の CLI を続ける（`layat --rollback` など）

性格の違う操作が同じ階層のフラグに並び、操作ごとに有効なフラグの組み合わせが分かりにくい。

### dry-run を `layat plan` サブコマンドにする

terraform 風で読みやすいが、dry-run は独立した操作ではなく apply のモードである。サブコマンドに
すると、apply の選択やフラグを plan 側にも重ねて持つことになる。

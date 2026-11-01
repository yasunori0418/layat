---
id: "ADR-777fc3c8-4ca8-49da-aaf3-146b7d991ebe"
type: adr
name: "gitignore コマンド"
status: 採用
origin: "ADR-0005 の決定「.gitignore 生成は専用コマンドで stdout 出力のみ」・ADR-0013 §6・ADR-0023 §5 の (5a) を書き直した"
justifies:
  - "REQ-a480c183-40ce-4201-93b5-65a7a59c1b9e"
  - "REQ-eaa8c0df-af44-4f52-9603-cd2bc22a67e9"
references:
  - "ADR-4dca6ef4-2791-48d5-aa82-7037630968a6"
---
# ADR-777fc3c8-4ca8-49da-aaf3-146b7d991ebe: gitignore コマンド

- ステータス: 採用
- 日付: 2026-11-01
- 関連: ADR-4dca6ef4-2791-48d5-aa82-7037630968a6, `docs/spec.md`
- 起点: ADR-0005 の決定「.gitignore 生成は専用コマンドで stdout 出力のみ」・ADR-0013 §6・ADR-0023 §5 の (5a) を書き直した

## 背景

project mode の配置物はコミットせず、activation は `.gitignore` に触れない
（ADR-4dca6ef4-2791-48d5-aa82-7037630968a6）。それでも利用者は、配置物を git に無視させるために
target の一覧を `.gitignore` へ書く必要がある。その一覧を手で作らずに済む手段が要った。

## 決定

配置 target を `.gitignore` 向けに列挙する専用のコマンド `layat gitignore` を設ける。出力は
stdout だけで、ファイルには書き込まない。各行は root 相対の target に先頭 `/` を付けたアンカー
形式にする。対象は project mode の config に限り、それ以外の config を指定したらエラーで止める。
`.gitignore` を更新する責務はプロジェクトの管理者が持つ。規範の細目は
REQ-a480c183-40ce-4201-93b5-65a7a59c1b9e と REQ-eaa8c0df-af44-4f52-9603-cd2bc22a67e9 に置く。

## 根拠

stdout に出すだけなら、layat は配置以外のファイルを書き換えないツールのままでいられる。
利用者はリダイレクトで `.gitignore` へ追記できる。project mode の root は git toplevel で、
`.gitignore` の置き場所と一致するので、先頭 `/` のアンカーは正しい位置を指し、別の階層の同名の
パスを誤って無視しない。home mode や fixed root ではこの前提が成り立たず、出力しても意味が無い。

## 棄却した代替案

### `.gitignore` へ自動で追記する（`--write`）

手間は省けるが、layat がファイルを生成・改変するツールになり、設定を生成しない方針に反する。

### 全モードで受け付ける

home mode や fixed root ではアンカーの基準が `.gitignore` の場所と合わない。意味の無い出力を
利用者が誤って `.gitignore` に貼る危険がある。

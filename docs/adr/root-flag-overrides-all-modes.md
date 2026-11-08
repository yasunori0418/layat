---
id: "ADR-27f99ee6-5fb4-4dbe-b0fd-8267eecaa3a1"
type: adr
name: "--root は全モードで root を上書きする"
status: stable
origin: "ADR-0017 §2 を書き直した"
justifies:
  - "REQ-61c05e09-0bde-4f74-9a96-03185f9df606"
references:
  - "ADR-400cadd4-0151-470b-92ae-9b11aa8e9bed"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:27:41+09:00"
---
# ADR-27f99ee6-5fb4-4dbe-b0fd-8267eecaa3a1: --root は全モードで root を上書きする

- ステータス: stable
- 日付: 2026-11-01
- 関連: ADR-400cadd4-0151-470b-92ae-9b11aa8e9bed, `docs/spec.md`
- 起点: ADR-0017 §2 を書き直した

## 背景

engine は marker が運ぶ種別から root を実行時に解決する（ADR-400cadd4-0151-470b-92ae-9b11aa8e9bed）。
CLI の `--root <path>` は解決した root を上書きするフラグで、当初は project mode の root を
差し替える用途だけを想定していた。home mode や fixed root の config に渡したときの扱いを
決める必要があった。

## 決定

`--root <path>` は、manifest の root の種別（project / home / system / fixed）に関わらず、engine が
解決する root を一律に上書きする。規範の細目は REQ-61c05e09-0bde-4f74-9a96-03185f9df606 に置く。

## 根拠

project mode だけに限る技術的な理由は無い。テストで home mode の config を一時ディレクトリへ
逃がすといった正当な用途があり、限定するとその用途を塞ぐ。

## 棄却した代替案

### project mode に限る

home mode の root が変わらないことは守れるが、テストなどの正当な用途を塞ぐ。

### project mode 以外では無視して warning を出す

指定したのに root が変わらず、利用者が驚く。

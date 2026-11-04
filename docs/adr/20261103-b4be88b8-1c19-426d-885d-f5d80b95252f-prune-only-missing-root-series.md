---
id: "ADR-b4be88b8-1c19-426d-885d-f5d80b95252f"
type: adr
name: "prune は root が実在しない roothash 系列だけを削除する"
status: 採用
issues:
  - "#580"
origin: "次期マイルストーン計画の grilling（2026-07-04）。公開後の実運用で孤児系列が溜まり始めたことを消費側の要求と判断した。ADR-0034 §1a を書き直した"
justifies:
  - "REQ-d41b1d0a-c6d5-41cc-93f9-e5cc7f152da4"
  - "REQ-c44433a1-7ee7-459a-9aae-7cc42166876f"
references:
  - "ADR-ad1c2507-1f42-4cd3-ab56-9183892f8988"
  - "ADR-fadf64ef-2308-40e5-8060-34d63e6ddb66"
---
# ADR-b4be88b8-1c19-426d-885d-f5d80b95252f: prune は root が実在しない roothash 系列だけを削除する

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-ad1c2507-1f42-4cd3-ab56-9183892f8988, ADR-fadf64ef-2308-40e5-8060-34d63e6ddb66, GitHub Issue #580
- 起点: 次期マイルストーン計画の grilling（2026-07-04）。公開後の実運用で孤児系列が溜まり始めたことを消費側の要求と判断した。ADR-0034 §1a を書き直した

## 背景

project mode と `--root` で root を上書きした config の profile は `<roothash>` 階層に積まれ、
backref `.root` が root の絶対パスを記録する（ADR-ad1c2507-1f42-4cd3-ab56-9183892f8988）。
プロジェクトのクローンを消すと、その系列は世代 symlink のディレクトリとして残り続ける。

孤児になった系列は store を占有しない（store は `nix-collect-garbage` で解放される）。
それでも系列のディレクトリは消えず、手で消すには backref の中身を一つずつ確かめる手間がかかる。
公開後の実運用で、こうした孤児系列が実際に溜まり始めた。

## 決定

`layat prune` は、backref `.root` を持つ roothash 系列のうち、記録された root パスが FS
上に実在しない系列だけを、`<roothash>` ディレクトリごと削除する。判定に使うのは root
パスの実在だけで、root が実在する系列は、ほかに使われていない兆候があっても削除しない。`.root`
を持たない `<name>` 直キーの系列は判定する root パスが無いので、構造的に対象から外れる。

走査する基底と、root パスを決められない系列の扱いは REQ-c44433a1-7ee7-459a-9aae-7cc42166876f に、
孤児系列を prune まで放置してよいとする前提は REQ-d41b1d0a-c6d5-41cc-93f9-e5cc7f152da4 に置く。

## 根拠

「root パスが実在しない」は機械的に決まる安全側の条件で、誤って消す余地は backref
の記録の誤りとアンマウントの 2 つしかない。backref は apply が書くので記録の誤りは構造的に起きず、
アンマウントは削除前の確認で防げる。

世代の間引きは `nix-env --delete-generations` が既にできる仕事で、prune に持たせると二重になる。
prune の仕事を「root が消えた系列の片付け」に絞れば、root が実在する系列には一切触れないと言い切れる。

## 棄却した代替案

### 世代の間引き（`--older-than` / `--keep N`）も prune に含める

`nix-env --delete-generations` の仕事と二重になる。root が実在する系列の世代の管理は prune
の仕事にしない。

### root が実在しても、entrypoint の消失などから使われていない系列を推定して候補に出す

「使われていない」は機械的に判定できない。誤検出の害は削除系のコマンドで最も大きいので、
安全側の条件だけを扱う。

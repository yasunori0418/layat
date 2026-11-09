---
id: "ADR-7a3dc78d-8946-46ae-94ec-b752f2ca7c2e"
type: adr
name: "prune は system 基底も走査し、権限の無い系列を見送る"
status: stable
origin: "次期マイルストーン計画の grilling（2026-07-04）。ADR-0034 §1b と ADR-0036 §3c を書き直した"
justifies:
  - "REQ-c44433a1-7ee7-459a-9aae-7cc42166876f"
  - "DSG-096dc893-21f4-45e3-9347-986e9275b4d1"
references:
  - "ADR-b4be88b8-1c19-426d-885d-f5d80b95252f"
  - "ADR-95b45d12-d4a7-4374-8624-f5ceed7044ee"
  - "ADR-fadf64ef-2308-40e5-8060-34d63e6ddb66"
  - "ADR-8517e311-c461-41aa-80ac-48760a7762b2"
  - "ADR-63ec1446-d328-42d2-b8d2-e8a11d752ed7"
  - "ADR-e630482a-bc11-4ba6-a9cc-1ef46f54104d"
  - "ADR-25d54346-f13a-4c12-9895-ab93dc6c183b"
  - "ADR-11507e12-5b0f-4f8f-bae9-fb7fd41a9203"
  - "ADR-04cc2a03-78dd-4bf3-9cc7-3b8ed7faf429"
  - "ADR-15a358f4-1c81-49d5-b658-3a1fe39f0f75"
  - "ADR-400cadd4-0151-470b-92ae-9b11aa8e9bed"
  - "ADR-419914ba-9896-4dd1-853b-fb514fcb9f38"
  - "ADR-ad1c2507-1f42-4cd3-ab56-9183892f8988"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# prune は system 基底も走査し、権限の無い系列を見送る

- ステータス: stable
- 日付: 2026-11-03
- 関連: [prune は root が実在しない roothash 系列だけを削除する](prune-only-missing-root-series.md), [prune は配置物に触れない](prune-leaves-placements-untouched.md), [prune は削除前に root パスの一覧を示して確認を取る](prune-confirms-with-root-list.md), [prune はロックを取れない系列を待たずに飛ばす](prune-skips-locked-series.md), [prune は新設時から --json に対応する](prune-json-output.md), [prune は引数を取らない](prune-command-without-arguments.md), [system mode の変更系操作の前に engine が euid 0 を確認する](system-mode-euid-gate.md), [profile の基底ディレクトリをモードで切り替える](profile-base-directory-per-mode.md), [apply --all に root モードのフィルタを持たせる](apply-all-root-mode-filter.md), [module と project では世代操作を公開しない](module-and-project-hide-generation-ops.md), [root は 3 つの marker で表す](three-root-markers.md), [profile は専用ディレクトリレイアウトに置く](profile-directory-layout.md), [profile は解決済み root の roothash でキーする](profile-keyed-by-roothash-of-resolved-root.md)
- 起点: 次期マイルストーン計画の grilling（2026-07-04）。ADR-0034 §1b と ADR-0036 §3c を書き直した

## 背景

`layat prune` は、root が実在しなくなった `<roothash>` 系列を削除するコマンドである。profile の基底は
ユーザー state 基底だけではなく、system mode の config では system profile 基底になる。system mode でも
`--root` で root を上書きした系列は `<roothash>` でキーされるので、上書き先のディレクトリが消えれば
ユーザー state 基底の系列と同じように孤児になる。

一方、system 基底の下にある系列の削除には root 権限が要る。prune を非 root で実行すると、
system 基底の孤児系列は削除しようとした時点で権限エラーになる。

## 決定

`layat prune` はユーザー state 基底に加えて system profile 基底も走査する。非 root で実行して権限が
無いために削除へ着手できなかった系列は、warning を出して見送り、他の系列の処理を続ける。

走査の対象と見送りの条件は [prune は backref の root が実在しない roothash 系列だけを系列ごと削除する](../requirements/prune-target-selection.md) と
[prune を paths の列挙・engine の判定と削除・CLI の同意解釈に分け、engine API を Reset と同型にする](../design/prune-orphan-series.md) に置く。

## 根拠

system mode の `--root` 系列も、ユーザー state 基底の系列と同じ理由で孤児になる。基底が違うことを
理由に走査から外すと、その孤児を消す手段が手作業しか残らない。

system 基底の系列を 1 つ消せないことを理由に全体を止めると、消せるはずの系列まで残る。見送った系列は warning で
示されるので、root で実行し直せば消せることが分かる。

## 棄却した代替案

### system 基底を走査の対象から外す

root = `/` の `<name>` 直キー系列は孤児にならないが、`--root` で上書きした系列は孤児になる。
これを消す手段が無くなる。

### 権限エラーで prune 全体を失敗させる

非 root の実行では system 基底に孤児が 1 つあるだけで毎回失敗し、ユーザー state 基底の掃除もできなくなる。

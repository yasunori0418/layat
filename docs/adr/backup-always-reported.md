---
id: "ADR-c0a91b81-f54f-4790-9295-f90656f53de8"
type: adr
name: "退避の発生は常に stderr へ出す"
status: stable
sources:
  - id: "issue-169"
    resource: "https://github.com/yasunori0418/layat/issues/169"
    title: "#169"
  - id: "issue-172"
    resource: "https://github.com/yasunori0418/layat/issues/172"
    title: "#172"
origin: "旧 epic #167 から統合された epic #172 の grilling セッション（2026-07-11・2026-07-12）で確定。ADR-0045 §6（出力規律）の退避の報告を書き直した"
justifies:
  - "REQ-8ef34101-8150-4124-92d5-94fabe6b5d90"
  - "REQ-5dd5a4e9-6162-4fa5-b295-66844f5a4f3b"
references:
  - "ADR-36bfbfbd-f1dd-4581-a75a-943d80f6ccf5"
  - "ADR-3dc21349-5504-48a4-ab88-812b92c3bdf9"
  - "ADR-a19bf484-368d-4d01-b539-52de32510fa5"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T23:16:45+09:00"
---
# ADR-c0a91b81-f54f-4790-9295-f90656f53de8: 退避の発生は常に stderr へ出す

- ステータス: stable
- 日付: 2026-11-03
- 関連: [指定が無いときは成功時に何も出力しない](silent-on-success.md), [-v で配置レポートを opt-in で出す](verbose-opts-into-placement-report.md), [内部コマンドの開示を --debug に分ける](debug-discloses-internal-commands.md), GitHub Issue #169, #172
- 起点: 旧 epic #167 から統合された epic #172 の grilling セッション（2026-07-11・2026-07-12）で確定。ADR-0045 §6（出力規律）の退避の報告を書き直した

## 背景

layat は成功したときに何も出さず（[指定が無いときは成功時に何も出力しない](silent-on-success.md)）、配置のレポートは `-v` を付けたときだけ出す（[-v で配置レポートを opt-in で出す](verbose-opts-into-placement-report.md)）。
`--backup` による退避をこの沈黙の対象にするかを決める必要があった。

## 決定

退避が起きた事実は、成功時に何も出さない扱いの対象にせず、`-v` が無くても warning 級として stderr へ出す。
copy の foreign な実ファイルを退避したときは、skip のときに出す warning の代わりにこの報告を
出す。

出力の規律は [成功時はデフォルト沈黙とし warning と error は常時 stderr へ出す](../requirements/silent-on-success.md) と
[apply --backup は配置を塞ぐ記録外実体を rename 退避してから配置する](../requirements/apply-backup-rename-aside.md) に置く。

## 根拠

退避は、利用者の所有物を layat が動かしたという重い出来事である。成功の一部として出さずにおくと、
ファイルが別名へ移ったことに利用者が気付けない。warning と同じく常に見せる。

copy の foreign な実ファイルは、`--backup` が無ければ skip して warning を出す。退避した場合は
退避の報告が可視化を兼ねるので、warning を重ねて出さない。

## 棄却した代替案

### 配置のレポートと同じく、`-v` を付けたときだけ見せる

配置前除去による移行のような意図された安全な操作と違い、退避は利用者のファイルを動かす。
`-v` を付けない限り気付けないのでは、退避の可視性を保てない。

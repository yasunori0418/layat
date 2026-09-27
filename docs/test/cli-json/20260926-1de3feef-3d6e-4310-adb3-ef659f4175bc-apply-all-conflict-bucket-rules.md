---
id: "TC-1de3feef-3d6e-4310-adb3-ef659f4175bc"
type: test_condition
name: "前段衝突検査がバケット規則と選択集合に従って衝突を判定する"
mitigates:
  - "RISK-8241dd1d-f856-4ed5-b572-05b88232a199"
---
# TC-1de3feef-3d6e-4310-adb3-ef659f4175bc: 前段衝突検査がバケット規則と選択集合に従って衝突を判定する

## 条件

`apply --all` の前段検査が、同じ root に解決する config 同士だけを突き合わせ、衝突したら
何も build せずに止まることを確かめる。

- **バケット規則** — 同じ rootKind（project / home / system）同士は衝突し、rootKind が違えば
  衝突しない。fixed は root 値が同じときだけ衝突する。fixed と project は同じ root に偶然
  解決しても別扱い（検出不能として実行時に委ねる）。`--root` 指定時は rootKind・root 値に
  関わらず全体が 1 つのバケットになる
- **選択集合** — root フィルタで選択外になった config 同士の衝突では止まらない
- **停止の位置** — 衝突時は一括 eval の後で止まり、build が 1 回も起動しない。`--dryrun`
  でも同じ
- **単一 apply の非検査** — 名指しの apply は一括 eval を足さず、従来どおりの eval 1 回と
  build に進む
- **エラーの形** — メッセージが衝突 target と両 config 名を含む。`--json` ではトップレベル
  `errors[]` に `E_LAYAT_FAILED` として 1 件載り、`results[]` は空

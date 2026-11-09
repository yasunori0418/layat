---
id: "ADR-fb01d942-2875-4dcb-a7df-879e18f63883"
type: adr
name: "apply --recopy で全 copy target を src から無条件に上書きする"
status: stable
sources:
  - id: "issue-584"
    resource: "https://github.com/yasunori0418/layat/issues/584"
    title: "#584"
origin: "ADR-0020 §1（apply --recopy で copy target を上書き再コピーする）を書き直した"
justifies:
  - "REQ-7cc32a2b-eee4-4a29-8dc1-a1dc23e7a065"
  - "REQ-d2277c7a-7992-49af-a9dc-4cc73843a6f9"
  - "REQ-b4e4b65d-6e35-40c3-a00e-20c14043df6f"
references:
  - "ADR-c4ad9987-850e-4834-bbea-d8de3b073548"
  - "ADR-a51897c5-f709-4b7a-8db2-61e6ebb0040c"
  - "ADR-acc31876-df36-485e-b356-92ae60e26db7"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# apply --recopy で全 copy target を src から無条件に上書きする

- ステータス: stable
- 日付: 2026-11-03
- 関連: [copy は世代の外に place-once で置く](copy-outside-generations-place-once.md), [独立した操作はサブコマンドにし、操作の修飾はフラグにする](subcommand-set.md), [reset は非対話の実行で --yes を必須にする](reset-confirmation-and-yes.md)
- 起点: ADR-0020 §1（apply --recopy で copy target を上書き再コピーする）を書き直した

## 背景

[copy は世代の外に place-once で置く](copy-outside-generations-place-once.md)。一度置いた
target には以後触れないので、`src` を更新しても copy した側には反映されない。反映するには、
target を手で消してから apply をやり直すしかなかった。

layat は copy した内容の hash を記録していない。そのため、どの target が `src` から変わったかを
後から判定する材料が無い。

## 決定

`apply` に修飾フラグ `--recopy` を設ける。`--recopy` を付けた apply は、通常の配置に加えて、
config の全 copy target を現在の `src` から無条件に上書きする。copy は世代の外のままで、
`--recopy` が世代を増やすことはない。

上書きの対象と手順は [apply --recopy は config 内の全 copy target を src から無条件に上書き再コピーする](../requirements/apply-recopy.md) に、上書き時の退避は
[recopy の上書きは削除ではなく同一親内への rename 退避で行う](../requirements/recopy-rename-aside.md) に、place-once の既定は
[copy は target 不在のときだけマテリアライズする place-once で世代管理の対象外とする](../requirements/copy-place-once.md) に置く。

## 根拠

apply のフラグにすれば、symlink の更新と copy の追従を 1 回のコマンドで済ませられる。

内容の hash を持たない以上、差分を判定せず全部を上書きするしかない。ローカルで編集した内容は
失われるが、`src` に追従させたいという `--recopy` の目的そのものなので、確認は求めない。
フラグを付けること自体が利用者の明示の同意になる。

## 棄却した代替案

### 専用のサブコマンド `layat recopy` にする

symlink を含めて一括で更新するのに 2 回のコマンドが要る。

### 変わった target だけを上書きする

編集を不要に上書きせずに済むが、mode の保存や copy 内の symlink の複製を考えると差分の判定が自明でなく、
実装が重い。利用者の責任で上書きするという意図ともずれる。

### 専用の機能を設けず、reset のあとに apply をやり直してもらう

CLI の表面は最小になるが、手順が 2 段になり、symlink と同時に更新できない。

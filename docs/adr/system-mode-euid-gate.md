---
id: "ADR-25d54346-f13a-4c12-9895-ab93dc6c183b"
type: adr
name: "system mode の変更系操作の前に engine が euid 0 を確認する"
status: stable
origin: "次期マイルストーン計画の grilling（2026-07-04）。ADR-0036 §2 を書き直した"
justifies:
  - "REQ-0d1f557b-7ae7-41ea-b17b-cdb5c80935d4"
references:
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
# ADR-25d54346-f13a-4c12-9895-ab93dc6c183b: system mode の変更系操作の前に engine が euid 0 を確認する

- ステータス: stable
- 日付: 2026-11-03
- 関連: [profile の基底ディレクトリをモードで切り替える](profile-base-directory-per-mode.md), [apply --all に root モードのフィルタを持たせる](apply-all-root-mode-filter.md), [module と project では世代操作を公開しない](module-and-project-hide-generation-ops.md), [root は 3 つの marker で表す](three-root-markers.md), [profile は専用ディレクトリレイアウトに置く](profile-directory-layout.md), [profile は解決済み root の roothash でキーする](profile-keyed-by-roothash-of-resolved-root.md)
- 起点: 次期マイルストーン計画の grilling（2026-07-04）。ADR-0036 §2 を書き直した

## 背景

system mode は root = `/` に配置し、profile も system 基底に置くので、root 権限が無ければ書き込みは
必ず失敗する。確認をしなければ、失敗は flock の取得や一部の確認が進んだあとで、最初に触った target の
`EACCES` として現れる。どの target で失敗するかは偶然に左右され、sudo での再実行を促す案内も無い。

## 決定

system mode の config に対する変更系の操作（`apply` / `reset` / `rollback`）では、engine がディレクトリの
作成や flock の取得より前に実効ユーザー ID が 0 であることを確認し、0 でなければ root 権限での再実行を
案内するエラーで停止する。`apply --dryrun` や `list-generations` のような読み取りだけの操作は確認しない。

確認の対象とエラーの扱いは [system mode は root を / に解決し、変更系操作の前に euid 0 を確認し、profile と last-applied 記録を system 基底に置く](../requirements/system-mode-root-gate-and-record-base.md) に置く。

## 根拠

前段で確認すれば、権限不足による失敗が決定的になり、エラーも原因を直接示せる。確認を CLI ではなく
engine に置くのは、NixOS / nix-darwin モジュールの activation が `apply --manifest` で起動したときも、
同じ確認を通るようにするためである。

見るのは実効ユーザー ID だけで、sudo・doas・root でのログインを区別しない。`SUDO_USER` などの
環境変数に頼ると、起動のしかたによって結果が変わる。

読み取りだけの操作を確認から外すのは、非 root の利用者でも計画の確認と世代の閲覧ができるようにするため
である。eval と build は nix daemon を通すので、非 root でも実行できる。

## 棄却した代替案

### 確認をせず、書き込みの失敗をそのまま返す

失敗が最初に触った target の `EACCES` という偶然の形になり、sudo での再実行への案内も無い。
前段で止める方が、利用者にとってもテストにとっても扱いやすい。

### 確認を CLI に置く

エントリポイントを評価する起動と `--manifest` での起動のように、engine を呼ぶ起動のしかたごとに確認を
足すことになる。足し忘れた起動のしかたでは確認されない。

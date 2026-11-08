---
id: "ADR-1743c957-a732-4a44-a89a-4c03ab277f0c"
type: adr
name: "NixOS の dry-activate では --dryrun を付けて起動する"
status: stable
sources:
  - id: "issue-519"
    resource: "https://github.com/yasunori0418/layat/issues/519"
    title: "#519"
  - id: "issue-493"
    resource: "https://github.com/yasunori0418/layat/issues/493"
    title: "#493"
origin: "Issue #493（epic: NixOS / nix-darwin モジュールの実装）の grilling（2026-10-24）。ADR-0066 §4a を書き直した"
justifies:
  - "REQ-a220bb28-89de-42b0-a854-e7468755ecb1"
  - "DSG-0e186e89-c8c6-4bca-9daa-03855e8d5cda"
references:
  - "ADR-f58e9981-bacc-4937-98b8-bddce0a057db"
  - "ADR-1338a293-3f69-48f5-9164-efb4ec6994d6"
  - "ADR-0da9d817-c870-4f48-8a6d-8fb9b55e2a70"
  - "ADR-15a358f4-1c81-49d5-b658-3a1fe39f0f75"
  - "ADR-abfe1b12-1a7d-42e1-8520-c1e9fcf5846c"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:27:41+09:00"
---
# ADR-1743c957-a732-4a44-a89a-4c03ab277f0c: NixOS の dry-activate では --dryrun を付けて起動する

- ステータス: stable
- 日付: 2026-11-03
- 関連: [モジュールの activation は apply --all --manifest を 1 回だけ起動する](module-activation-single-apply-all.md), [モジュールは configs 間の target 衝突を評価時に確認しない](modules-no-eval-conflict-assertion.md), [NixOS の activation の PATH に nix を前置する](nixos-activation-path-nix.md), [module と project では世代操作を公開しない](module-and-project-hide-generation-ops.md), [system mode の last-applied 記録を /var/lib/layat に置く](system-record-base-var-lib.md), GitHub Issue #519, #493
- 起点: Issue #493（epic: NixOS / nix-darwin モジュールの実装）の grilling（2026-10-24）。ADR-0066 §4a を書き直した

## 背景

NixOS の activation スクリプトは、`nixos-rebuild switch` のときだけでなく、`dry-activate` のときと
boot のときにも実行される。どの契機で実行されたかは環境変数 `NIXOS_ACTION` で分かる。boot のときは
この変数が設定されない。`dry-activate` は FS を変えずに変更の予定を示す操作である。

## 決定

NixOS モジュールは、`NIXOS_ACTION` が `dry-activate` のときだけ engine の起動に `--dryrun` を付ける。
変数が設定されていない boot のときを含め、他の値では通常どおり起動する。nix-darwin モジュールには
この分岐を持たせない。

起動の規範は [NixOS / nix-darwin モジュールは activation から全 config のビルド済み link-farm を systemRoot で 1 回の起動に渡す](../requirements/system-module-activation-contract.md) と [NixOS / nix-darwin モジュールは modules/kick.nix を介した system mode の配線とする](../design/module-stubs.md) に置く。

## 根拠

`dry-activate` は FS を変えないことを前提にした操作なので、layat もその前提に合わせる。`--dryrun` を
付ければ、layat の配置の予定も他の activation の予定と一緒に示せる。

`dry-activate` 以外で通常どおり起動するのは、boot のときの起動で配置のずれを直す機会を残すためである。

`NIXOS_ACTION` は NixOS の activation が設定する変数で、nix-darwin の activation では設定されない。
nix-darwin に同じ分岐を持たせても使われない。

## 棄却した代替案

### `NIXOS_ACTION` が `switch` のときだけ通常どおり起動する

変数が設定されない boot のときに起動しなくなり、配置のずれを直す機会を失う。

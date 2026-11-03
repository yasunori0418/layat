---
id: "ADR-1743c957-a732-4a44-a89a-4c03ab277f0c"
type: adr
name: "NixOS の dry-activate では --dryrun を付けて起動する"
status: 採用
issues:
  - "#519"
  - "#493"
origin: "Issue #493（epic: NixOS / nix-darwin モジュールの実装）の grilling（2026-10-24）。ADR-0066 §4a を書き直した"
justifies:
  - "REQ-a220bb28-89de-42b0-a854-e7468755ecb1"
  - "DSG-0e186e89-c8c6-4bca-9daa-03855e8d5cda"
supersedes:
  - "ADR-0066"
references:
  - "ADR-f58e9981-bacc-4937-98b8-bddce0a057db"
  - "ADR-1338a293-3f69-48f5-9164-efb4ec6994d6"
  - "ADR-0da9d817-c870-4f48-8a6d-8fb9b55e2a70"
  - "ADR-15a358f4-1c81-49d5-b658-3a1fe39f0f75"
  - "ADR-abfe1b12-1a7d-42e1-8520-c1e9fcf5846c"
---
# ADR-1743c957-a732-4a44-a89a-4c03ab277f0c: NixOS の dry-activate では --dryrun を付けて起動する

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-f58e9981-bacc-4937-98b8-bddce0a057db, ADR-1338a293-3f69-48f5-9164-efb4ec6994d6, ADR-0da9d817-c870-4f48-8a6d-8fb9b55e2a70, ADR-15a358f4-1c81-49d5-b658-3a1fe39f0f75, ADR-abfe1b12-1a7d-42e1-8520-c1e9fcf5846c, GitHub Issue #519, #493
- 起点: Issue #493（epic: NixOS / nix-darwin モジュールの実装）の grilling（2026-10-24）。ADR-0066 §4a を書き直した

## 背景

NixOS の activation スクリプトは、`nixos-rebuild switch` のときだけでなく、`dry-activate` のときと
boot のときにも実行される。どの契機で実行されたかは環境変数 `NIXOS_ACTION` で分かる。boot のときは
この変数が設定されない。`dry-activate` は FS を変えずに変更の予定を示す操作である。

## 決定

NixOS モジュールは、`NIXOS_ACTION` が `dry-activate` のときだけ engine の起動に `--dryrun` を付ける。
変数が設定されていない boot のときを含め、他の値では通常どおり起動する。nix-darwin モジュールには
この分岐を持たせない。

起動の規範は REQ-a220bb28-89de-42b0-a854-e7468755ecb1 と DSG-0e186e89-c8c6-4bca-9daa-03855e8d5cda に置く。

## 根拠

`dry-activate` は FS を変えないことを前提にした操作なので、layat もその前提に合わせる。`--dryrun` を
付ければ、layat の配置の予定も他の activation の予定と一緒に示せる。

`dry-activate` 以外を通常の起動に倒すのは、boot のときの起動で配置のずれを直す機会を残すためである。

`NIXOS_ACTION` は NixOS の activation が設定する変数で、nix-darwin の activation では設定されない。
nix-darwin に同じ分岐を持たせても使われない。

## 棄却した代替案

### `NIXOS_ACTION` が `switch` のときだけ通常どおり起動する

変数が設定されない boot のときに起動しなくなり、配置のずれを直す機会を失う。

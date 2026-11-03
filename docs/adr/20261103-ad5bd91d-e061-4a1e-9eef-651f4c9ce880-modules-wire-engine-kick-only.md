---
id: "ADR-ad5bd91d-e061-4a1e-9eef-651f4c9ce880"
type: adr
name: "モジュールは root と起動の時機だけを供給して engine を起動する"
status: 採用
issues:
  - "#585"
origin: "ADR-0003 §b を書き直した"
justifies:
  - "REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10"
  - "REQ-8d965ca2-f8fd-44a4-87f3-94e850e9f85b"
  - "REQ-8085f194-c903-4ecb-abd8-c719fe7b3292"
  - "DSG-98d7fa5d-c590-441a-b682-0ab7afa14233"
---
# ADR-ad5bd91d-e061-4a1e-9eef-651f4c9ce880: モジュールは root と起動の時機だけを供給して engine を起動する

- ステータス: 採用
- 日付: 2026-11-03
- 関連: GitHub Issue #585, `docs/spec.md`
- 起点: ADR-0003 §b を書き直した

## 背景

配置は全層で engine が行い、`home.file` や `systemd.tmpfiles` などのネイティブ機構には委譲しない。
一方で、統合層ごとにモジュール（home-manager / NixOS / nix-darwin）がある。モジュールの受け持ちを
決めておかないと、衝突の検出や古い配置の除去といった配置の振る舞いがモジュール側へ漏れ、層ごとに
分岐していく。

## 決定

モジュールは、root と起動の時機（`home.activation` / `system.activationScripts`）だけを供給し、
switch のときに engine を起動する配線に徹する。配置・衝突の検出・古い配置の除去・世代のコミットは
モジュールに持たせない。

root の解決は REQ-8d965ca2-f8fd-44a4-87f3-94e850e9f85b と
REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10、起動の契約は REQ-8085f194-c903-4ecb-abd8-c719fe7b3292、
起動の 2 クラスへの束ね方は DSG-98d7fa5d-c590-441a-b682-0ab7afa14233 に置く。

## 根拠

振る舞いの源を engine 1 つに集めれば、どのモジュールから使っても配置は同じ挙動になる。
モジュールに残すのは、プラットフォームごとに異なる「どこが root か」「いつ起動するか」だけである。
配置の規則をモジュールに書くと nix と bash で実装することになり、Go の engine に比べてテストしにくい。
集合の扱いやチェックも CLI に寄せれば、3 つのモジュールで同じ規則を写し直さずに済む。

## 棄却した代替案

### モジュールがネイティブ機構へ翻訳する

`home.file` や `systemd.tmpfiles` へ翻訳すれば各プラットフォームの追跡や GC を使えるが、配置の
振る舞いが層ごとに二重になる。`systemd.tmpfiles` は作成だけを扱い古い配置を除去しないため、
どのみち削除は layat が担うことになる。

### 配置の一部をモジュールが nix / bash で持つ

config のループや部分失敗の集約、衝突の検出をモジュール側に書くと、CLI と同じ規則を 2 か所で
保つことになり、片方だけが変わって分岐する。

---
id: "ADR-ad5bd91d-e061-4a1e-9eef-651f4c9ce880"
type: adr
name: "モジュールは root と起動の時機だけを供給して engine を起動する"
status: stable
sources:
  - id: "issue-585"
    resource: "https://github.com/yasunori0418/layat/issues/585"
    title: "#585"
origin: "ADR-0003 §b を書き直した"
justifies:
  - "REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10"
  - "REQ-8d965ca2-f8fd-44a4-87f3-94e850e9f85b"
  - "REQ-8085f194-c903-4ecb-abd8-c719fe7b3292"
  - "DSG-98d7fa5d-c590-441a-b682-0ab7afa14233"
references:
  - "ADR-2119433e-de69-4447-85a5-575d58246007"
  - "ADR-59f63d1f-25df-4290-913f-02804c5eeade"
  - "ADR-90964d0f-3c2e-4201-a04b-5821f2f4c799"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# ADR-ad5bd91d-e061-4a1e-9eef-651f4c9ce880: モジュールは root と起動の時機だけを供給して engine を起動する

- ステータス: stable
- 日付: 2026-11-03
- 関連: [engine を単一の固定 Go バイナリにする](engine-fixed-go-binary.md), [root の型を mkManifest だけに持たせる](roottype-mkmanifest-only.md), [devShell の shellHook で配線する](devshell-shellhook-wiring.md), GitHub Issue #585, `docs/spec.md`
- 起点: ADR-0003 §b を書き直した

## 背景

配置は全層で engine が行い、`home.file` や `systemd.tmpfiles` などのネイティブ機構には委譲しない。
一方で、統合層ごとにモジュール（home-manager / NixOS / nix-darwin）がある。モジュールの受け持ちを
決めておかないと、衝突の検出や古くなった配置の除去といった配置の振る舞いがモジュール側へ漏れ、層ごとに
分岐していく。

## 決定

モジュールは、root と起動の時機（`home.activation` / `system.activationScripts`）だけを供給し、
switch のときに engine を起動する配線に徹する。配置・衝突の検出・古くなった配置の除去・世代のコミットは
モジュールに持たせない。

root の解決は [home mode の root は層ごとに定まった供給元から解決する](../requirements/home-mode-root-resolution.md) と
[全モジュールは共通オプションの同一集合を公開し、entries は configs 経由・root はモジュールが pin する](../requirements/module-common-options.md)、起動の契約は [home-manager モジュールの engine kick は activation から全 config のビルド済み link-farm を 1 回の起動で渡し、失敗で switch を止める](../requirements/hm-activation-contract.md)、
起動の 2 クラスへの束ね方は [engine の起動を entrypoint 駆動とビルド済み manifest の 2 クラスに束ね、統合層の差を link-farm の取得方法だけに閉じる](../design/integration-layer-kick-classes.md) に置く。

## 根拠

振る舞いの源を engine 1 つに集めれば、どのモジュールから使っても配置は同じ挙動になる。
モジュールに残すのは、プラットフォームごとに異なる「どこが root か」「いつ起動するか」だけである。
配置の規則をモジュールに書くと nix と bash で実装することになり、Go の engine に比べてテストしにくい。
集合の扱いや確認も CLI に寄せれば、3 つのモジュールで同じ規則を写し直さずに済む。

## 棄却した代替案

### モジュールがネイティブ機構へ翻訳する

`home.file` や `systemd.tmpfiles` へ翻訳すれば各プラットフォームの追跡や GC を使えるが、配置の
振る舞いが層ごとに二重になる。`systemd.tmpfiles` は作成だけを扱い古くなった配置を除去しないため、
どのみち削除は layat が担うことになる。

### 配置の一部をモジュールが nix / bash で持つ

config のループや部分失敗の集約、衝突の検出をモジュール側に書くと、CLI と同じ規則を 2 か所で
保つことになり、片方だけが変わって分岐する。

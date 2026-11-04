---
id: "ADR-70debe94-c7ce-492b-b938-4c62f7a781e7"
type: adr
name: "out-of-store のリンクは lib の関数が包むだけにして engine が張る"
status: 採用
origin: "ADR-0001 §d と ADR-0003 §d を書き直した"
justifies:
  - "REQ-eb363122-385a-499c-a074-c95efb949d07"
  - "REQ-81249072-56b8-42f5-807b-ea623c1efe99"
references:
  - "ADR-73c24620-d2c0-4249-97fa-1b6296f1e79c"
  - "ADR-2119433e-de69-4447-85a5-575d58246007"
  - "ADR-451679d5-473b-4faf-bf2d-a6f6d3e1e131"
  - "ADR-59f63d1f-25df-4290-913f-02804c5eeade"
  - "ADR-90964d0f-3c2e-4201-a04b-5821f2f4c799"
---
# ADR-70debe94-c7ce-492b-b938-4c62f7a781e7: out-of-store のリンクは lib の関数が包むだけにして engine が張る

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-73c24620-d2c0-4249-97fa-1b6296f1e79c, ADR-2119433e-de69-4447-85a5-575d58246007, ADR-451679d5-473b-4faf-bf2d-a6f6d3e1e131, ADR-59f63d1f-25df-4290-913f-02804c5eeade, ADR-90964d0f-3c2e-4201-a04b-5821f2f4c799, `docs/design.md`
- 起点: ADR-0001 §d と ADR-0003 §d を書き直した

## 背景

out-of-store を明示関数で選ばせると決めると、その関数が何をするかを決める必要がある。
home-manager の `config.lib.file.mkOutOfStoreSymlink` は、外部パスを指す symlink を持つ
derivation を作り、`home.file` で配置する。layat の lib は nixpkgs だけに依存し、
配置は全層で engine が行う。HM モジュールの中でだけ HM の関数へ委ねると、out-of-store の配置が
層によって別の機構になる。

## 決定

`mkOutOfStoreSymlink` は、渡されたパスをマーカーに包んで返すだけの純粋関数とする。リンクを実際に
張るのは engine で、HM モジュールでも HM の `mkOutOfStoreSymlink` には委ねない。マーカーの形は
REQ-eb363122-385a-499c-a074-c95efb949d07、src の解決の時点は
REQ-81249072-56b8-42f5-807b-ea623c1efe99 に置く。

## 根拠

lib を純粋関数に留めれば、nixpkgs だけに依存する制約を保ったまま nix-unit で確かめられる。
リンクの生成を engine に集めれば、out-of-store の配置・古くなった配置の除去・drift 修復が store link と
同じ engine の処理で行われ、どの統合層から使っても挙動が変わらない。

## 棄却した代替案

### HM モジュールでは HM の `mkOutOfStoreSymlink` へ委ねる

HM の上でだけ配置の機構が変わり、CLI 単体や他のモジュールと挙動が分かれる。古くなった配置の除去も
HM と layat の 2 か所に分かれる。

### lib が外部パスを指す symlink の derivation を作る

リンクがストア内の中継を経由する 2 段になり、engine が記録・比較するリンク先が中継のパスになる。
マーカーのまま manifest にパスを書けば、engine はリンク先を直接扱える。

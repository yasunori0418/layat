---
id: "ADR-2e454c1f-6cad-41d8-8311-2c8e5b8efdf6"
type: adr
name: "mkManifest 自身が evalModules で入力を検査する"
status: 採用
origin: "manifest の型検査を NixOS モジュールシステムで行えるかという確認。ADR-0010 §1（検査機構）を書き直した"
justifies:
  - "REQ-d1b5b3f5-10a0-400d-9f03-ba00c63d1c34"
  - "REQ-c5dfcae6-6094-4850-99e5-bf14530bc60a"
---
# ADR-2e454c1f-6cad-41d8-8311-2c8e5b8efdf6: mkManifest 自身が evalModules で入力を検査する

- ステータス: 採用
- 日付: 2026-11-01
- 関連: `docs/spec.md`
- 起点: manifest の型検査を NixOS モジュールシステムで行えるかという確認。ADR-0010 §1（検査機構）を書き直した

## 背景

`lib.mkManifest` の入力（`entries` と `root`）が流れ込む経路は 2 つある。CLI や entrypoint から
`mkManifest` を直接呼ぶ経路と、home-manager / NixOS のモジュールの `layat.entries` オプションを
通る経路である。`mkOption` に型を書いても、型の検査が行われるのは `lib.evalModules` を通したときだけ
なので、型をオプションに書くだけでは直接呼ぶ経路の入力が検査されないまま通る。

`lib.types`・`mkOption`・`lib.evalModules` は nixpkgs の lib の中核で、NixOS や home-manager の
ものではない。

## 決定

`mkManifest` 自身が `lib.evalModules` で `entries` と `root` を検査・正規化する。設定だけで
判定できる誤りは、この検査で評価時に止める。

検査の主体と層の分け方の規範は REQ-d1b5b3f5-10a0-400d-9f03-ba00c63d1c34・
REQ-c5dfcae6-6094-4850-99e5-bf14530bc60a に置く。

## 根拠

layat の中心の使い方は、CLI と project mode で `mkManifest` を直接呼ぶ経路である。ここを検査
しなければ型定義の意味が薄れる。`mkManifest` が検査を持てば、どちらの経路でも検査が行われる。

モジュールシステムは nixpkgs の lib だけで使えるので、「lib は nixpkgs の lib だけに依存する」という
制約を破らない。既定値の適用・未知キーの検出・型エラーのメッセージを、自前で書かずに得られる。

## 棄却した代替案

### 型をモジュールのオプションに書くだけにする

検査がモジュール経路でしか行われず、中心の使い方である直接呼ぶ経路が無検査になる。

### 全ての検査を手書きの `assert` / `throw` で行う

既定値の適用・未知キーの検出・型エラーのメッセージを自前で作り直すことになり、コードが増える
割に品質が落ちる。

### NixOS の `assertions` オプションを lib に持ち込む

この仕組みが要るほど検査の数が無く、nixpkgs の lib だけで済む軽さを損なう。

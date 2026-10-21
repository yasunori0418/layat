---
id: "ADR-0064"
type: adr
name: "HM モジュールに `layat.configs.<name>.subdir` を足し、root marker の subdir 適用形へ配線する"
status: 採用
issues:
  - "#507"
  - "#150"
origin: "Issue #507（epic #150 の B3。lib に入った `homeRoot { subdir = \"...\"; }` をモジュールから使う口）"
justifies:
  - "REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10"
  - "REQ-068f2fb3-5459-4325-9359-c931b5049e0d"
revises:
  - "ADR-0035"
references:
  - "ADR-0007"
  - "ADR-0040"
  - "ADR-0061"
---
# ADR-0064: HM モジュールに `layat.configs.<name>.subdir` を足し、root marker の subdir 適用形へ配線する

- ステータス: 採用
- 日付: 2026-10-22
- 関連: ADR-0007, ADR-0040, ADR-0061, GitHub Issue #507, #150
- 改訂対象: ADR-0035 §1 の `configs` submodule の形 `{ entries }` を `{ entries, subdir }` へ広げる。「1 config = 1 profile = 1 manifest」と `layat.entries` の糖衣は不変
- 起点: Issue #507（epic #150 の B3。lib に入った `homeRoot { subdir = "..."; }` をモジュールから使う口）

## 背景

root marker の subdir 適用形（`homeRoot { subdir = "..."; }`・→ ADR-0040）は lib に入り、`mkManifest` は root を動かさずに全 target へ subdir を前置できるようになった。一方 `modules/home-manager.nix` は `root = layatLib.homeRoot` を固定しており、HM 利用者には適用形を渡す口が無い。

config（= profile）ごとに配置の基点は異なる。たとえば `ai-tools` は `~/.claude`、`desktop` は `~/.config` の下にまとまる。基点が同じ entry 群には毎回同じ接頭辞を書くことになる。standalone の `layat.<name>` と同じく、この接頭辞を subdir 適用形で 1 か所にまとめられるようにしたい。

## 決定

### 1. `layat.configs.<name>.subdir :: nullOr str`（既定 null）を共通オプションに足す

- 定義は `modules/common.nix` の `configs` submodule に置き、全モジュール共通の集合に載せる。NixOS / nix-darwin は stub のままで、オプションだけが共通集合に入る。
- config ごとに持つ。全 config 共通の `layat.subdir` は設けない。

### 2. HM モジュールは subdir から root の適用形を組む

- `c.subdir == null` なら素の `homeRoot`、非 null なら `homeRoot { subdir = c.subdir; }` を、その config の `mkManifest` の root に渡す。
- root は home のまま pin され、変わるのは target だけになる（全 target が `$HOME/<subdir>/` 配下になる）。`root` オプションは引き続き公開しない。

### 3. subdir の検査は lib に委ねる

- モジュールは subdir の検査・正規化・空文字の読み替えを持たない。絶対パス・`..` による脱出・空は lib の検査が `layat: invalid root subdir (...)` の eval エラーとして止める。`""` は null と同一視せず、そのまま lib へ渡す。

## 根拠

- **config ごとに持つ理由**: 配置の基点は役割（profile）ごとに異なり、1 本の共通値では表せない。standalone の `layat.<name>` も config ごとに root を持つ。
- **`root` を公開しない理由**: モジュールが自分の性質で root を pin する原則（→ ADR-0007）を保つため。subdir 適用形は root を動かさず target だけを変えるので、この原則と両立する。
- **検査を lib に委ねる理由**: 同じ規則を 2 か所で持つと分岐する。lib の検査は standalone と同じ経路で効くので、モジュールでもう一度検査しても得るものが無い。
- **`apply --all` の衝突検出**: subdir 適用形は eval 時に target へ前置する糖衣で（→ ADR-0040）、manifest には前置後の target が載る。config 間の target 衝突検出は engine と manifest.json スキーマを変えずに追従する。
- **オプションの置き場所**: `layat.recopy` / `layat.jobs`（→ ADR-0061）と同じく共通集合へ足し、モジュールごとに集合を分岐させない。

## 影響

- **`modules/common.nix`**: `configs` submodule に `subdir` を足す。
- **`modules/home-manager.nix`**: config ごとに root の適用形を組んで `mkManifest` へ渡す。
- **要求**: 共通オプション集合（REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10）の `configs` 型表記に `subdir` を加え、HM の subdir 配線の要求（REQ-068f2fb3-5459-4325-9359-c931b5049e0d）を起票する。
- engine・manifest.json スキーマ・lib は変えない。

## 棄却した代替案

- **全 config 共通の `layat.subdir`**: 基点が config ごとに異なる用途に合わない。必要になった時点で別に検討する。
- **`root` オプションの公開**: 利用者が root を再指定できるようになり、モジュールが root を pin する原則を崩す。subdir だけで target の基点は変えられる。
- **モジュール側での再検査**: lib と同じ規則を二重に持つことになり、片方だけが変わると挙動が分岐する。

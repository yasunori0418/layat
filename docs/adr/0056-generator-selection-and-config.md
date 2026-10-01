---
id: "ADR-0056"
type: adr
name: "生成器の選択と設定ファイル"
status: 採用
issues:
  - "#401"
  - "#395"
origin: "Issue #395（epic: manifest 生成器のインターフェイス化）の grilling（2026-09-09）"
justifies:
  - "REQ-badc7e10-0ba4-40d4-b334-6e40193119db"
  - "REQ-637599dc-a1ec-4af5-9e97-e882c7df56d0"
revises:
  - "ADR-0006"
  - "ADR-0043"
references:
  - "ADR-0007"
  - "ADR-0011"
  - "ADR-0026"
  - "ADR-0055"
---
# ADR-0056: 生成器の選択と設定ファイル

- ステータス: 採用
- 日付: 2026-10-02
- 関連: ADR-0006（設定ファイルから config を発見する機構は足さない）, ADR-0007（entrypoint 発見の規律）, ADR-0011（CLI の依存許可と vendorHash）, ADR-0026（`--manifest` 経路と `-f` / `--all` との排他）, ADR-0043（`E_INPUT` の予約）, ADR-0055（manifest 生成器の契約）
- 改訂対象: ADR-0006 の「設定ファイルから config を発見する機構は足さない」に、生成器の選択に限って設定ファイルを読む例外を足す（config の発見は引き続き足さない）。ADR-0043 §8 が予約した共通コード `E_INPUT` を初めて実装し、入力不正の分類先にする
- 起点: Issue #395（epic: manifest 生成器のインターフェイス化）の grilling（2026-09-09）

## 背景

ADR-0055 で manifest の取得は生成器インターフェイス越しになった。実装する生成器は nix と prebuilt（`--manifest`）とテストダブルで、利用者が選べる実生成器は当面 nix だけだが、選ぶ仕組みが無いと 2 つ目の生成器を足すたびに CLI の外面を決め直すことになる。

選び方には 2 通りある。各生成器の Discover を順に試して最初に entrypoint を見つけたものを使う暗黙の方式（claim 方式）と、利用者が明示する方式である。claim 方式は entrypoint の置き方次第で生成器が切り替わり、同じディレクトリに複数の生成器の entrypoint があると結果が試行順に依存する。

明示の手段には、毎回のフラグのほか、プロジェクトやユーザー単位で固定する設定ファイルが要る。layat はこれまで設定ファイルを読まず、ADR-0006 は「CWD / 設定ファイルから config を発見する機構は足さない」と決めている。

HM 等のモジュール activation は `apply --manifest` で engine を起動する（→ ADR-0026）。activation の実行環境の環境変数・カレントディレクトリは利用者の制御外で、ここで設定ファイルを読むと activation が利用者の手元の設定に左右される。

## 決定

### 1. 生成器は明示指定でのみ選ぶ

claim 方式は採らない。優先順位は次のとおりで、上で値が決まれば下は読まない。

1. `--generator <name>`（persistent flag）
2. 環境変数 `LAYAT_GENERATOR`
3. プロジェクト設定 `layat.toml`
4. ユーザー設定 `$XDG_CONFIG_HOME/layat/config.toml`（`XDG_CONFIG_HOME` 未設定時は `~/.config/layat/config.toml`）
5. 既定 `nix`

### 2. 値の enum は `nix` のみ。prebuilt は出さない

- 選択可能な値は `nix` だけ。2 つ目の実生成器が現れたときに足す。
- prebuilt は `--manifest` でだけ選ばれ、`--generator` の値に出さない。`--generator` と `--manifest` の同時指定はエラーにする。
- **`--manifest` 時は環境変数・設定ファイルを参照しない**。activation の環境不定（背景）から `--manifest` 経路を切り離すため。

### 3. 設定ファイルの探索

- プロジェクト設定 `layat.toml` は **`-f` で指定したディレクトリ、無ければ cwd** だけを見る。**上方向探索はしない**。entrypoint 発見と同じ規律で、「設定を読むには entrypoint が要り、entrypoint 発見には生成器が要る」循環を避ける。
- 設定ファイルが読むのは生成器の選択だけで、**config の発見には使わない**（ADR-0006 の決定は不変）。

### 4. 形式は TOML で strict

- 形式は TOML。パーサに `github.com/pelletier/go-toml/v2` を CLI の依存として許可する（engine は stdlib-only のまま）。
- 未知キーはエラー（strict）。version 項目は持たない。項目は当面 `generator` のみ。

```toml
generator = "nix"
```

### 5. 入力不正は `E_INPUT`

次はいずれも入力不正として人間向けは exit 1 + 1 行、`--json` では共通コード `E_INPUT`（ADR-0043 §8 で予約済み・初実装）に分類する。

- 未知の enum 値（`--generator` / `LAYAT_GENERATOR` / 設定ファイルのいずれでも）
- 設定ファイルの不正（TOML のパース失敗・未知キー）
- フラグの組み合わせ不正: `--generator` + `--manifest`、既存の `-f` + `--manifest` と `--all` + `--manifest`（既存 2 件は `E_LAYAT_FAILED` から `E_INPUT` へ変わる。文面は不変）

### 6. `--generator` は prune / init で無視する

`--generator` は persistent flag とし、manifest を得ない prune と生成器化の範囲外の init（→ ADR-0055 §1）では無視する（エラーにしない）。persistent flag `-f` と同じ扱い。

## 根拠

- **明示のみ**: 生成器の切り替えは配置結果を変える操作で、ファイルの置き方から暗黙に決まると原因を追えない。
- **環境変数と 2 段の設定ファイル**: CI・一時的な切り替えは環境変数、リポジトリ単位の固定はプロジェクト設定、マシン単位の既定はユーザー設定に置ける。フラグが常に勝つので、どの層の設定も 1 回の実行で上書きできる。
- **strict / version なし**: 項目が 1 つの段階で前方互換の仕組みを持つと使われない仕様が固まる。未知キーをエラーにしておけば、将来の項目追加は旧バージョンで明示的に失敗する。
- **`E_INPUT`**: 利用者の入力の誤りで、生成器の失敗（`E_LAYAT_BUILD`）や汎用の失敗（`E_LAYAT_FAILED`）と区別できると消費者が再試行の要否を判断できる。

## 影響

- 実装は Issue #404（`cmd/layat` の選択機構・`go.mod` / `vendorHash`・e2e シナリオ）。
- ADR-0006 / 0043 に改訂注記を足す。
- 要求: 選択機構（REQ-badc7e10-0ba4-40d4-b334-6e40193119db）を起票し、CLI の依存許可（REQ-637599dc-a1ec-4af5-9e97-e882c7df56d0）に go-toml/v2 を足す。

## 棄却した代替案

- **claim 方式（Discover を順に試す）**: 背景のとおり、entrypoint の置き方と試行順で生成器が決まる。
- **設定ファイルの上方向探索**: 循環を生み、どの設定が効いたかが cwd の深さに依存する。
- **`--manifest` 時も設定を読む**: activation が利用者の手元の設定に左右される。
- **prebuilt を enum に出す**: `--manifest` のパス指定と二重の入口になり、片方だけ指定したときの意味が定まらない。
- **JSON / YAML**: JSON はコメントを書けず手書きの設定に向かない。YAML は暗黙の型変換があり、strict にしても値の誤読が残る。

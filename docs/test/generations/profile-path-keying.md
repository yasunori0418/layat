---
id: "TC-40b762fc-3b48-471c-b495-c5ffe1f584b5"
type: test_condition
name: "state 基底・profileDir のキーイング・世代リンク名が root と mode ごとに正しく決まる"
status: stable
mitigates:
  - "RISK-ac5cd9b4-58f9-43d8-9fae-66f7b5940beb"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:27:49+09:00"
---
# TC-40b762fc-3b48-471c-b495-c5ffe1f584b5: state 基底・profileDir のキーイング・世代リンク名が root と mode ごとに正しく決まる

## テスト条件

profile の on-disk レイアウトを決めるパス解決層を、環境変数と入力の組み合わせに対する
関数として検証する。

**`<state>` の解決** — `$XDG_STATE_HOME` があればそれを使い（`$HOME` の値によらず）、
無ければ `$HOME/.local/state` を使う。`$HOME` も解決できなければエラーにする
（利用者に知らせずに相対パスや空文字を使わない）。

**`<roothash>` の性質** — 同じ絶対 root に対して決定的で、長さが固定であること。
決定的でなければ apply のたびに別の世代系列が生まれ、長さが揺れれば FS 上の名前として
使えない。

**キーイングの分岐** — mode と `--root` の有無で、`<name>` 直キーと `<roothash>` キーの
どちらを取るかが決まる。

| 入力 | キー |
|---|---|
| project mode | `<roothash>` |
| home mode（`--root` なし）| `<name>` 直キー |
| home mode + `--root` 上書き | `<roothash>` |
| system mode（`--root` なし）| `<SystemBase>/<name>` 直キー |
| system mode + `--root` 上書き | `<SystemBase>/<roothash>/<name>` + backref |
| fixed root mode | `<roothash>` |

**世代リンクの命名** — profile リンクに対する `profile-N-link` の形が `nix-env` の
規約どおりで、同一ディレクトリ内の兄弟として並ぶこと。

**兄弟の配置** — `.pending` out-link が profileDir 内に置かれること。backref `.root` は
roothash 階層（`<name>` dir の親）に置かれ、`<roothash>` キーのモードでのみ生じること
（home / system の `<name>` 直キーでは空。→ [profileDir は config 専用ディレクトリとし、profile リンク・世代・pending out-link をその中に並べる](../../requirements/profile-on-disk-layout.md)）。

上位の規範は [エンジンとコマンド層の Go テストは nix を介さない実 FS 統合テストを主戦力とする](../../test-plan/go-test-layering.md) の「FS に一切触れずに判断する層」。`internal/paths/` の解決と
ハッシュ化は入力が列挙可能なので、統合テストがたまたま到達する範囲に委ねず、ユニット
レベルで table として覆う。

## 対応する CASE

[paths_test.go — state 基底の解決・roothash の性質・profileDir キーイング・世代リンク名と兄弟パス](paths-test-go.md)（`internal/paths/paths_test.go`）。root 解決そのもの（`rootKind` と
`--root` から絶対 root を得る側）は [internal/engine/engine_test.go — 実 tmpdir に対する apply の統合テスト](../engine-core/engine-test-go.md) も隣接して検証する。

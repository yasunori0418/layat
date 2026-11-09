---
id: "TC-746cb5b9-fe27-4d09-a51d-eb03d56a93a7"
type: test_condition
name: "profileDir の世代リンクを番号順に列挙して current を判定し、読めない profileDir は黙認しない"
status: stable
mitigates:
  - "RISK-ac5cd9b4-58f9-43d8-9fae-66f7b5940beb"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:27:49+09:00"
---
# profileDir の世代リンクを番号順に列挙して current を判定し、読めない profileDir は黙認しない

## テスト条件

profileDir に並ぶ `profile-<N>-link` を読み取り専用で走査して世代一覧を作る層を、tmpdir に
張った世代リンクに対して検証する。

- **列挙と順序** — `profile-<N>-link` の兄弟を世代番号の昇順に返す（名前の辞書順ではない）。
  世代の名前の形をしていない兄弟（`profile.lock`・番号が数値でない名前・別 profile の名前）は
  世代に数えない
- **日時** — 各世代の日時はリンク自身の lstat の mtime で、リンク先の mtime ではない
- **current** — `profile` リンクが指す世代だけを current と判定する
- **profileDir 不在** — profileDir がまだ無いときは世代 0 件で、エラーにしない
- **読めない profileDir** — 不在以外の理由で profileDir を読めないときはエラーを返す。
  「世代 0 件」と見なすと、前世代が無いという最も危険な結論を利用者に知らせずに出すことになる

上位の規範は [state 基底・profileDir のキーイング・世代リンク名が root と mode ごとに正しく決まる](profile-path-keying.md) と同じく [エンジンとコマンド層の Go テストは nix を介さない実 FS 統合テストを主戦力とする](../../test-plan/go-test-layering.md)。
世代リンクのレイアウトは tmpdir 上の実 FS に張って確かめ、nix-env は起動しない。

## 対応する CASE

[nixprofile_test.go — 世代リンクの走査と 2 基底](nixprofile-test-go.md)（`internal/generation/nixprofile/nixprofile_test.go`）。

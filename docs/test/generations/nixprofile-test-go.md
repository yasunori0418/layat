---
id: "CASE-80a327c3-5fdd-48d4-9446-4b3e0e262105"
type: test_case
name: "nixprofile_test.go — 世代リンクの走査と 2 基底"
status: stable
resource: "internal/generation/nixprofile/nixprofile_test.go"
covers:
  - "TC-746cb5b9-fe27-4d09-a51d-eb03d56a93a7"
follows:
  - "TP-e7c25263-6d2d-4a37-8275-26906889d912"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-15T00:21:56+09:00"
---
# nixprofile_test.go — 世代リンクの走査と 2 基底

## 対象

`internal/generation/nixprofile/nixprofile_test.go`

世代バックエンドの nix 実装（`internal/generation/nixprofile`）のユニットテスト。実際の nix-env は
起動せず、tmpdir に張った世代リンクに対する `List`（Fake と共用する走査）と、バックエンドが
所有する 2 つの基底（user / system）の値と、PATH の先頭に置いた偽の nix-env を失敗させた
ときのエラーを検証する。

## 検証内容

**世代リンクの走査**（[profileDir の世代リンクを番号順に列挙して current を判定し、読めない profileDir は黙認しない](generation-listing-parse.md)）

- `profile-<N>-link` を番号の昇順に返し、`profile` の指す世代だけを current にする
- 世代でない兄弟名（`profile.lock` 等）を無視する
- 日時はリンク自身の mtime（リンク先の mtime ではない）
- profileDir が無ければ 0 件でエラーにしない
- profileDir を読めなければ（ディレクトリでない等）、列挙の失敗であることと profileDir を
  示すエラーにする

**nix-env の失敗**

- `Commit` と `Switch` が失敗すると、操作（`--set`、`--switch-generation N`）と profile を
  添えた `CommandError` を返す。nix-env が stderr へ書いた内容は `Stderr` に取り込み、エラーの
  文には含めない。原因の終了状態を `errors.As` で取り出せる
- nix-env が PATH に無いときも、操作と profile を添えたエラーを返し、not-found を `errors.Is`
  で判別できる

**基底**

- `UserBase(<state>)` が `<state>/nix/profiles/layat`、`SystemBase()` が
  `/nix/var/nix/profiles/layat` を返す（engine と CLI はこの値を直接持たない）

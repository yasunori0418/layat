---
id: "DSG-f8d9bba5-cfd5-4a19-b1fd-fd8ecc8a1814"
type: design
name: "internal/generation は契約・Generation 型・FS ベースの Fake を親パッケージに、nix-env 実装を nixprofile サブパッケージに置き、engine は Backend 1 フィールドで受ける"
satisfies:
  - "REQ-833205ca-adfc-4c86-bf66-ae4d8fe4885b"
  - "REQ-f4d7d4ab-fbdb-48c6-b29f-08dd88e72645"
---
# DSG-f8d9bba5-cfd5-4a19-b1fd-fd8ecc8a1814: internal/generation は契約・Generation 型・FS ベースの Fake を親パッケージに、nix-env 実装を nixprofile サブパッケージに置き、engine は Backend 1 フィールドで受ける

## 設計

```
cmd/layat                           ← バックエンドの注入だけ（newGenerationBackend()・常に nixprofile）
  │ import
  ├─ internal/generation            ← 契約（Backend インターフェイス・Generation 型）
  │                                   + テストダブル Fake（fake.go・非 _test ファイル。profile-<N>-link を実際に張る）
  ├─ internal/generation/nixprofile ← nix-env 実装（New() generation.Backend。
  │                                   nix-env 3 コマンド・世代一覧のパース・2 基底）
  ├─ internal/paths                 ← profileDir 内部のレイアウト（profile / profile-<N>-link / .pending）。
  │                                   Resolve は基底を引数に取る
  └─ internal/engine                ← Options / RollbackOptions / ResetOptions / PruneOptions /
                                      ProfileOptions が Backend 1 フィールドを持つ。nil は error
```

```go
type Generation struct{ Number int; Date string; Current bool }
type Backend interface {
    Commit(profileLink, linkFarm string) error
    List(profileLink string) ([]Generation, error)
    Switch(profileLink string, gen int) error
    UserBase(stateDir string) string
    SystemBase() string
}
```

- **契約と nix-env 実装をパッケージで分ける**。`nixprofile` は `generation` を import する側に
  立ち、逆向きの依存を持たない。`nix-env` の呼び出し・世代一覧のパース・基底
  `<state>/nix/profiles/layat` と `/nix/var/nix/profiles/layat` は `nixprofile` の外へ出ない。
- **Fake は親パッケージの非 `_test` ファイルに置く**。`cmd/layat` のテストからも import する
  ため。Fake は FS ベースで `profile-<N>-link` と `profile` リンクを実際に張り、engine の
  レイアウト読みがそのまま動く。
- **engine は差し替え点の 3 関数型を `Backend` 1 フィールドに置き換える**。profileDir を解決する
  全入口が `Backend` を要求し、nil は error。Prune も `Backend` から user / system の両基底を
  導く（`PruneOptions.SystemDir` は環境変数 `LAYAT_SYSTEM_PROFILE_BASE` 由来の上書きだけに
  使う）。`paths.Base` と engine 側の `Generation` 型は無くなり、`paths.Resolve` は基底を受け取る。
- **cmd 層は `newGenerationBackend()` の注入点だけを持つ**。選択機構は無く、常に
  nixprofile を組み立てて各サブコマンドへ渡す。
- `.pending` の後片付けは engine の `os.Remove` のまま（build 出力の一時リンクの片付け）。
- 2 層構成の境界（CLI と engine の間は `manifest.json`）は変えない。世代バックエンドは
  engine の内側の構成で、CLI から見た契約は従来どおり。

## 出典

ADR-0057「世代バックエンドの契約」§2〜§4・§7。

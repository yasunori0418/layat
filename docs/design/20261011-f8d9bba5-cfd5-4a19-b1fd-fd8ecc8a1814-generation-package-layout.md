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
  │                                   + 世代リンクの走査 ListLinks（scan.go。Fake と nixprofile が共用）
  │                                   + テストダブル Fake（fake.go・非 _test ファイル。profile-<N>-link を実際に張る）
  ├─ internal/generation/nixprofile ← nix-env 実装（New() generation.Backend。
  │                                   Commit / Switch の nix-env 2 コマンド・List は ListLinks・2 基底）
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
  立ち、逆向きの依存を持たない。`nix-env` の呼び出しと基底
  `<state>/nix/profiles/layat` と `/nix/var/nix/profiles/layat` は `nixprofile` の外へ出ない。
- **世代一覧は親パッケージの走査を共用する**。`ListLinks` は `profile-<N>-link` を読み取り専用で
  走査し、Fake と `nixprofile` の List がどちらもこれを返す。`nix-env` は列挙でも profileDir に
  lock を取るため、List には使わない。
- **Fake は親パッケージの非 `_test` ファイルに置く**。`cmd/layat` のテストからも import する
  ため。Fake は FS ベースで `profile-<N>-link` と `profile` リンクを実際に張り、engine の
  レイアウト読みがそのまま動く。
- **engine は差し替え点の 3 関数型を `Backend` 1 フィールドに置き換える**。profileDir を解決する
  全入口が `Backend` を要求し、nil は error。Prune も `Backend` から user / system の両基底を
  導く（`SystemDir` は全コマンド共通の環境変数 `LAYAT_SYSTEM_PROFILE_BASE` 由来の上書きだけに
  使う）。`paths.Base` と engine 側の `Generation` 型は無くなり、`paths.Resolve` は基底を受け取る。
- **cmd 層は `newGenerationBackend()` の注入点だけを持つ**。選択機構は無く、常に
  nixprofile を組み立てて各サブコマンドへ渡す。
- `.pending` の後片付けは engine の `os.Remove` のまま（build 出力の一時リンクの片付け）。
- 2 層構成の境界（CLI と engine の間は `manifest.json`）は変えない。世代バックエンドは
  engine の内側の構成で、CLI から見た契約は従来どおり。

## 出典

ADR-62540f8f-a8ba-4441-97b0-36b50d7862d8「世代操作はパスと整数だけを受け渡す 5 操作の契約越しに行う」、
ADR-1cbacca1-1930-46ca-bf82-30310f82216e「世代の一覧を nix-env を経由せず世代リンクの走査で得る」、
ADR-4c8d459a-c371-4f0d-b009-16e2ca096e49「世代バックエンドのテストダブルを FS ベースの Fake に統一する」、
ADR-eb1aa25f-a0ca-4efe-bcd9-4cbe11368764「世代バックエンドの注入を必須にし nix-env への暗黙のフォールバックを持たない」、
ADR-ffe1afcb-0b51-4b87-b698-0a6dea307bdd「profileDir の内部レイアウトを世代バックエンド契約の固定部分にする」、
ADR-8b8c9109-a479-4571-8b10-5a6079127441「前世代の読み取りは契約に載せず engine が直接読む」、
ADR-1908099c-87f7-4df1-bda5-7731df7d9ed7「engine が直接叩く外部コマンドを git に限る」、
ADR-9723dbb3-6659-4a69-ac4f-411c98bfa818「.pending の後片付けを gcroot を知らない engine の FS 操作として行う」。

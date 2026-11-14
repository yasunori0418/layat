---
id: "ADR-6c477b0f-cd25-417d-91a0-172dd0ed084f"
type: adr
name: "キャッシュへの投入を main への push で起動する"
section: ci-release-docs
status: stable
origin: "参照: https://github.com/yasunori0418/nur-packages の `.github/workflows/cachix-push.yml`。ADR-0028 §1 を書き直した"
justifies:
  - "INF-af33c5a1-f846-4f09-8e43-c6ea042d6e06"
references:
  - "ADR-fb5e9bee-8132-431b-9948-6ad9995c9a9a"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:26:59+09:00"
---
# キャッシュへの投入を main への push で起動する

- ステータス: stable
- 日付: 2026-11-03
- 関連: [キャッシュは 3 system の matrix ビルドで投入する](ci-cache-via-matrix-build.md)
- 起点: 参照: https://github.com/yasunori0418/nur-packages の `.github/workflows/cachix-push.yml`。ADR-0028 §1 を書き直した

## 背景

layat のバイナリの投入先は cachix のキャッシュ `yasunori0418` である。[キャッシュは 3 system の matrix ビルドで投入する](ci-cache-via-matrix-build.md)。layat は flake input として消費される CLI で、
利用者の CI や手元は main の最新を参照することが多い。投入をどのイベントで起動するかを決める
必要がある。

## 決定

キャッシュへの投入は、main への push のうち layat のバイナリの内容を決める入力が変わったときに
起動する。docs だけの変更では起動しない。手動の投入は `workflow_dispatch` で行う。

対象とする入力の一覧は [バイナリキャッシュ（main push 時の cachix 投入）](../infrastructure/binary-cache.md) に置く。

## 根拠

main の最新を都度キャッシュすれば、利用者は layat を手元でビルドせずにキャッシュから取得できる。
ビルドの時間が減り、結果の再現性も上がる。

この workflow は main への push の後に実行され、required status check を持たない。トリガ段の `paths`
フィルタで除外してもマージをブロックしないので、`paths` をそのまま使える。push と
`workflow_dispatch` でだけ起動するため fork の PR からは実行されず、`CACHIX_AUTH_TOKEN` が無い状況は
生じない。

## 棄却した代替案

### tag の push で投入する

リリースの間に main へ入った変更がキャッシュされず、開発中の利用者が手元でのビルドを強いられる。
layat が flake input として main の最新を消費される形と噛み合わない。

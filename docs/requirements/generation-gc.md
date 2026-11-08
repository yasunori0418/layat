---
id: "REQ-706de717-4e47-471a-a1c0-448635be159c"
type: requirement
name: "世代操作は世代バックエンドの契約越しに行い、GC root の間引きと store 解放は nix コマンドへ分けて委譲する"
status: stable
derives_from:
  - "UC-0b6f60cb-3e98-4ee7-8929-4d94a29f0af6"
specification: |
  Every generation of the profile SHALL be a GC root. The generation operations that
  layat performs — the commit, switching to a generation (rollback included) and listing —
  SHALL go through the generation backend contract
  ([engine は世代操作を世代バックエンドの契約（Commit / List / Switch / UserBase / SystemBase）越しに行う](generation-backend-contract.md)); the default backend nixprofile SHALL
  implement the commit and the switch with the `nix-env --profile <profileDir>/profile`
  family (`--set`, `--switch-generation`) and the listing by scanning the generation links
  read-only. Thinning out old generations SHALL be
  done through `nix-env --profile <profileDir>/profile --delete-generations <gens>`, and
  the freeing of store paths that have thereby become unreferenced SHALL be done through
  `nix-collect-garbage`; neither SHALL be a layat subcommand. The new `nix profile`
  subcommands SHALL NOT be used against a nixprofile profile, being incompatible: they
  require a profile-manifest of their own directly under the profile, whereas a profile
  made by `nix-env --set` has a link farm there, containing layat's `manifest.json`.
specification_ja: |
  profile の各世代は GC root でなければならない。layat が行う世代操作（コミット・世代の切替
  （rollback を含む）・一覧）は世代バックエンドの契約（[engine は世代操作を世代バックエンドの契約（Commit / List / Switch / UserBase / SystemBase）越しに行う](generation-backend-contract.md)）
  越しに行わなければならず、既定バックエンド nixprofile はコミットと切替を
  `nix-env --profile <profileDir>/profile` 系（`--set` / `--switch-generation`）で、一覧を
  世代リンクの読み取り専用の走査で実装しなければならない。旧世代の間引きは
  `nix-env --profile <profileDir>/profile --delete-generations <gens>` で行わなければならず、
  それにより無参照になった store パスの解放は `nix-collect-garbage` で行わなければならない。
  いずれも layat のサブコマンドにしてはならない。新 `nix profile` サブコマンドを nixprofile
  の profile に対して使ってはならない（profile 直下に自身の profile-manifest を要求するのに
  対し、`nix-env --set` 製 profile は profile 直下が link-farm であり非互換のため）。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T23:05:37+09:00"
---
# REQ-706de717-4e47-471a-a1c0-448635be159c: 世代操作は世代バックエンドの契約越しに行い、GC root の間引きと store 解放は nix コマンドへ分けて委譲する

## 仕様

profile の各世代は GC root。`nix-env --profile <profileDir>/profile --delete-generations <gens>`
で旧世代を間引き、`nix-collect-garbage` で無参照になった store パスを解放する。

> **世代操作は `nix-env --profile <profileDir>/profile` 系で統一する**。コミット（`--set`）・
> rollback（`--rollback`）・任意世代切替（`--switch-generation`）・
> 間引き（`--delete-generations`）。store GC のみ `nix-collect-garbage`。一覧だけは nix-env を
> 経由せず、世代リンクを読み取り専用で走査する（nix-env は列挙でも profileDir に lock を取るため）。
> 新 `nix profile`（`list` / `wipe-history` / `rollback`）は profile 直下に自身の profile-manifest を
> 要求するが、`nix-env --set` 製 profile は profile 直下が link-farm（layat の `manifest.json` を
> 含む）なので **`nix profile` サブコマンドとは非互換**。

> **上は原文の写しで、規範は frontmatter が正**。原文が併記する「可変 JSON 方式は GC root を
> 作らず参照中 store パスが GC で消えるため採らない」は棄却案の理由であり要求ではない
> （manifest を store 内へ埋め込むことは [世代は link farm derivation を layat 自前 profile へコミットして積み、前世代 manifest から古くなった配置を除去する](generation-mechanism-link-farm-manifest.md) の担当）。`<profileDir>` のレイアウトは
> [profileDir は config 専用ディレクトリとし、profile リンク・世代・pending out-link をその中に並べる](profile-on-disk-layout.md)、`rollback` の手順は [rollback は FS を先に収束させてから profile ポインタを最後に移す](rollback-refit-then-pointer.md) の担当。

> **2026-10-11 追記（[世代操作はパスと整数だけを受け渡す 5 操作の契約越しに行う](../adr/generation-backend-contract-paths-and-ints.md)）**: 上の写しの「世代操作は `nix-env --profile` 系で統一する」は、
> layat が行う世代操作（コミット・切替・一覧）を世代バックエンドの契約越しに行い、`nix-env` 系は
> その既定実装 nixprofile の内側に閉じる、と読み替える。間引きと store GC を nix コマンドへ
> 委譲すること、`nix profile`（新 CLI）との非互換は変わらない。契約は
> [engine は世代操作を世代バックエンドの契約（Commit / List / Switch / UserBase / SystemBase）越しに行う](generation-backend-contract.md) の担当。

## 出典

`docs/spec.md`「世代管理仕様」→「GC とストレージ解放」節。

決定の実体は、世代操作を `nix-env --profile` 系で統一していた旧決定と、
`<profileDir>` レイアウトを確定した [profile は専用ディレクトリレイアウトに置く](../adr/profile-directory-layout.md)。世代操作を契約越しにしたのは
[世代操作はパスと整数だけを受け渡す 5 操作の契約越しに行う](../adr/generation-backend-contract-paths-and-ints.md)。

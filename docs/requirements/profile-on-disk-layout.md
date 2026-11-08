---
id: "REQ-2aa3abbc-90b2-486e-92de-d785554bdeb3"
type: requirement
name: "profileDir は config 専用ディレクトリとし、profile リンク・世代・pending out-link をその中に並べる"
status: stable
derives_from:
  - "UC-0b6f60cb-3e98-4ee7-8929-4d94a29f0af6"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
  - "UC-403fbe32-b146-401b-8b53-fe67c1e169c5"
specification: |
  `profileDir` SHALL be a directory dedicated to each config, holding within it the
  profile link, the generations and the build out-link. Its base `<state>` SHALL be
  `$XDG_STATE_HOME` when that is set and `~/.local/state` otherwise, consistently with the
  default profile location of nix itself; the base directory of profiles under `<state>`
  SHALL be given by the generation backend (`<state>/nix/profiles/layat` for the default
  nixprofile backend). The profile link SHALL be `<profileDir>/profile` and SHALL be the
  profile link handed to the generation backend (the subject of `nix-env --profile` under
  nixprofile); the generations `profile-N-link` created by the backend and the build
  out-link SHALL sit as its siblings inside the same directory, so that neither passes
  through the profile link and no breakage arising from passing through a read-only store
  path can occur structurally.
  The pending out-link SHALL be named `.pending`, there being at most one per config
  inside the dedicated directory. The backref file `.root`, which records the absolute
  path of the original root, SHALL be placed at the roothash level — the parent of the
  `<name>` directory — and SHALL be shared by the several `<name>` directories under it.
  `profileDir` SHALL also be the flock key, serializing apply / reset / rollback for the
  same config while remaining independent of a different `<name>` directory under the same
  roothash.
specification_ja: |
  `profileDir` は各 config 専用のディレクトリでなければならず、その中に profile リンク・世代・
  build out-link を置かなければならない。基底 `<state>` は `$XDG_STATE_HOME` があればそれ、
  無ければ `~/.local/state` としなければならない（nix 本体の profile 既定と整合）。`<state>`
  配下の profile の基底は世代バックエンドが与えなければならない（既定の nixprofile では
  `<state>/nix/profiles/layat`）。profile リンクは `<profileDir>/profile` でなければならず、
  これを世代バックエンドへ渡す profile リンクとしなければならない（nixprofile では
  `nix-env --profile` の対象）。バックエンドが作る世代 `profile-N-link` と build out-link の
  対象は、同じディレクトリ内の兄弟として並べなければならず、profile リンクを貫通させてはならない
  （read-only な store パスを貫通する破綻を構造的に起こさないため）。pending out-link は専用
  ディレクトリ内に config あたり最大 1 個であり、名は `.pending` としなければならない。元 root の
  絶対パスを記録した backref ファイル `.root` は roothash 階層（`<name>` dir の親）に置かなければ
  ならず、複数の `<name>` で共有しなければならない。`profileDir` は flock キーでもあり、
  同一 config の apply / reset / rollback を直列化しつつ、同 roothash でも別 `<name>` dir とは
  独立させなければならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T23:05:37+09:00"
---
# REQ-2aa3abbc-90b2-486e-92de-d785554bdeb3: profileDir は config 専用ディレクトリとし、profile リンク・世代・pending out-link をその中に並べる

## 仕様

```
<state>/nix/profiles/layat/<roothash>/.root              # backref（roothash 階層・複数 <name> で共有）
<state>/nix/profiles/layat/<roothash>/<name>/             # ← profileDir（flock キー）
<state>/nix/profiles/layat/<roothash>/<name>/profile        # profile リンク（nix-env --profile <profileDir>/profile の対象）
<state>/nix/profiles/layat/<roothash>/<name>/profile-N-link # 世代（nix-env が profile の兄弟に作成）
<state>/nix/profiles/layat/<roothash>/<name>/.pending       # nix build --out-link（profile を貫通しない兄弟）
# home（--root なし）: <state>/nix/profiles/layat/<name>/{profile, profile-N-link, .pending}
```

> profile の基底 `<state>` は **`$XDG_STATE_HOME` があればそれ、無ければ `~/.local/state`**
> （nix 本体の profile 既定と整合）。

- profile 操作は `nix-env --profile <profileDir>/profile ...`、build は
  `nix build --out-link <profileDir>/.pending`。世代兄弟 `profile-N-link` と `.pending` が
  profile リンクを貫通せず兄弟として並ぶため、read-only な store パスを貫通する破綻が構造的に
  起きない
- **pending out-link は専用ディレクトリ内に 1 個なので名は `.pending`**（`<name>` 次元は
  ディレクトリ階層が表す）。`--set` 前失敗で残る pending は config あたり最大 1
- **flock キー = profileDir（専用ディレクトリ）**。同一 config の apply / reset / rollback を
  直列化し、同 roothash でも別 `<name>` dir とは独立する
- **backref `.root` は roothash 階層**（`<name>` dir の親）に置き、複数 `<name>` で共有する
  （孤児 profile 逆引き seam）

> **上は原文の写しで、規範は frontmatter が正**。`profileDir` の**キー**そのものは、home
> （`<name>` 直キー）・fixed・`--root` 上書き（`<roothash>/<name>`）が [profileDir は home と system のみ name 直キーとし、fixed root と --root 上書きは roothash でキーする](profiledir-keying.md)、project mode
> （解決済み root でキーすること）とその理由・`<roothash>`（解決後の絶対 root パスの sha256
> 短縮 hex）の定義が [project mode は世代を非公開にし、project / system mode は derivation 同一なら世代を積まず lstat ドリフト修復だけ行う](project-mode-generation-skip.md)、孤児 profile の扱いが [孤児 profile は backref で逆引き可能なまま放置許容とし、`layat prune` が root 不在の系列を削除する](project-mode-orphan-profile.md) の担当。
> HM モジュール経由も home mode の `<name>` 直キーに乗る（原文が挙げる「固定名 `default`」の
> MVP 限定は [HM モジュールは `layat.configs.<name>` ごとに独立した profile を持つ](../adr/hm-configs-independent-profiles.md) が反転済み・[HM モジュール経由でも名前つき config ごとに役割分離した独立 profile を取れる](hm-named-configs-profiles.md)）。本 item は「キーで指す先が
> ディレクトリであり、その中に何をどう並べるか」という物理形を規定する。flock を blocking で
> 取るか try-lock で取るかは [flock の取得は既定 blocking とし --no-wait のときだけ try-lock でスキップする](flock-acquisition-mode.md)、pending out-link が gcroot を持ち config あたり
> 最大 1 に抑えられることは [失敗時に残る pending gcroot は config あたり最大 1 個に有界とし回収処理を持たない](pending-gcroot-bounded.md) の担当。

> **2026-10-11 追記（[profileDir の内部レイアウトを世代バックエンド契約の固定部分にする](../adr/profiledir-layout-fixed-in-contract.md)）**: 上の写しの `nix-env --profile <profileDir>/profile ...` と
> 基底 `<state>/nix/profiles/layat` は、既定の世代バックエンド nixprofile の内部と読み替える。
> profile リンク・世代 `profile-N-link`・`.pending` が兄弟として並ぶレイアウトは世代バックエンドの
> 契約の固定部分で（[engine は世代操作を世代バックエンドの契約（Commit / List / Switch / UserBase / SystemBase）越しに行う](generation-backend-contract.md)）、engine はこれを直接読む。基底は
> 契約の `UserBase` が与え、設定ファイルでは上書きできない。

## 出典

`docs/spec.md`「世代管理仕様」→「機構」→「profile のオンディスクレイアウト」節と、
同「機構」節の機構表直後の blockquote（`<state>` の定義）。

決定の実体は [profile は専用ディレクトリレイアウトに置く](../adr/profile-directory-layout.md)「profile は専用ディレクトリレイアウトに置く」。`<state>` の定義は [profile の基底ディレクトリをモードで切り替える](../adr/profile-base-directory-per-mode.md)、backref
`.root` による孤児逆引きは [profile は解決済み root の roothash でキーする](../adr/profile-keyed-by-roothash-of-resolved-root.md) が定めている。基底の所有を世代バックエンドへ移し、レイアウトを
契約の固定部分としたのは [profileDir の内部レイアウトを世代バックエンド契約の固定部分にする](../adr/profiledir-layout-fixed-in-contract.md)・[state の基底を設定で上書きさせない](../adr/state-bases-not-configurable.md)。

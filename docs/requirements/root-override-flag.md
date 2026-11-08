---
id: "REQ-61c05e09-0bde-4f74-9a96-03185f9df606"
type: requirement
name: "--root は全モード共通で解決 root を明示上書きする"
status: stable
derives_from:
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  The `--root <path>` flag SHALL explicitly override the resolved root, in every mode
  alike: with it, project mode SHALL NOT use the git toplevel and home mode SHALL NOT use
  `$HOME`. How `profileDir` is keyed when it is given is stated by [profileDir は home と system のみ name 直キーとし、fixed root と --root 上書きは roothash でキーする](profiledir-keying.md) and is not
  restated here.
specification_ja: |
  `--root <path>` は解決 root を明示上書きするフラグでなければならず、全モード共通で
  効かなければならない（project は git toplevel を、home は `$HOME` を使ってはならない）。
  明示時に profileDir を
  どうキーするかは [profileDir は home と system のみ name 直キーとし、fixed root と --root 上書きは roothash でキーする](profiledir-keying.md) の担当で、本 item では規定しない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:29:58+09:00"
---
# REQ-61c05e09-0bde-4f74-9a96-03185f9df606: --root は全モード共通で解決 root を明示上書きする

## 仕様

```bash
--root <path>       # 解決 root を明示上書き（全モード共通。project は git toplevel を、home は $HOME を使わない）
```

> **上は原文の写しで、規範は frontmatter が正**。原文のフラグ表は上のコメントに続けて
> 「明示時は全モードで profileDir を上書き後 root の `<roothash>` でキー」を併記するが、これは
> **[profileDir は home と system のみ name 直キーとし、fixed root と --root 上書きは roothash でキーする](profiledir-keying.md) の担当**なので写しから落とした。profileDir のキーは root の種別ごとに定まる
> 一つの体系（home 直キー / fixed・project・`--root` は roothash）であり、`--root` の行だけを
> 切り出して二重に規範化すると [profile は解決済み root の roothash でキーする](../adr/profile-keyed-by-roothash-of-resolved-root.md) の改訂時に片方だけが直って食い違うため、
> [profileDir は home と system のみ name 直キーとし、fixed root と --root 上書きは roothash でキーする](profiledir-keying.md) へ一本化した。profileDir のレイアウト（`<roothash>/<name>` の中に何を並べるか）は
> [profileDir は config 専用ディレクトリとし、profile リンク・世代・pending out-link をその中に並べる](profile-on-disk-layout.md)、`<roothash>`（解決後の絶対 root パスの sha256 を短縮した hex）の定義は
> [project mode は世代を非公開にし、project / system mode は derivation 同一なら世代を積まず lstat ドリフト修復だけ行う](project-mode-generation-skip.md) の担当。

## 出典

`docs/spec.md`「CLI 仕様」→「サブコマンド体系」のグローバルフラグ表 `--root <path>`、および
「root の解決」→「project mode（`root = projectRoot`）」節の箇条書き第 3 項（`--root` は
project mode に限らず全モードの解決 root を一律上書きする）。

決定の実体は [--root は全モードで root を上書きする](../adr/root-flag-overrides-all-modes.md)。

---
id: "REQ-46fccb80-4bae-4d37-bc19-dded88e9a9c0"
type: requirement
name: "project mode は世代を非公開にし、project / system mode は derivation 同一なら世代を積まず lstat ドリフト修復だけ行う"
status: stable
derives_from:
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  In project mode the profile SHALL be keyed on the resolved root, so that cloning the
  same entrypoint into several places does not make the profiles collide and the stale
  removal of one clone does not destroy the placement of another — a problem that does not
  arise in home mode, where there is one per user. The `<roothash>` SHALL be a shortened
  hex of the sha256 of the resolved absolute root path, being thereby of fixed length and
  safe for a filesystem. Generations SHALL NOT be exposed to the user in project mode:
  the profile SHALL be confined to being the internal mechanism for stale removal and the
  generation-skip decision, and neither `rollback` nor `list-generations` SHALL be offered,
  the placed objects being ephemeral and there being no host generation to return to when
  kicked from a devShell. When the new link farm derivation is identical to the current
  generation of the profile and, when a last-applied record exists, also to its `linkFarm`
  ([apply は最後に配置した manifest を last-applied 記録に書き、古くなった配置の除去の比較の基準は記録を優先して読む](last-applied-record.md)), no new generation SHALL be stacked; the
  skip SHALL never be taken when no profile exists, and while the profile and the record
  disagree the full apply path SHALL run instead. Skipping is mandatory, since under
  devShell / direnv the `shellHook` runs on every shell re-entry and stacking a generation
  each time would make them grow without bound. The same generation skip and drift repair
  SHALL apply in system mode, since NixOS activation also runs on every boot and stacking a
  generation each time would make them grow with every boot. Home mode SHALL remain as it
  is, stacking a new generation on every application, the generation skip being confined
  to project and system mode. Even when the generation is skipped, a lightweight filesystem inspection SHALL
  still be performed rather than making it a complete no-op: the target of each entry
  SHALL be inspected with `lstat` against the baseline (the last-applied record when one
  exists, otherwise the current generation of the profile), and any entry that is not as
  recorded — rewritten by a
  foreign tool, or gone — SHALL be re-placed, and that entry alone, with a warning where
  it is a foreign symlink. The implication "identical derivation, therefore identical
  filesystem" breaks down under a foreign rewrite, so the filesystem SHALL be made to
  converge without stacking a generation; the `lstat` comparison is cheap and withstands
  the high frequency at which a `shellHook` runs. This drift repair SHALL cover both
  symlinks and copies. A copy entry SHALL be re-materialized under place-once only while
  its target is absent, and SHALL NOT be touched where the target exists but its content
  differs through the user having edited it, matching the behaviour of place-once in home
  mode; following the source SHALL remain confined to `apply --recopy`, and no comparison
  of content hashes SHALL be performed.
specification_ja: |
  project mode では profile を解決済み root でキーしなければならない。同一 entrypoint を複数箇所へ
  クローンしても profile が衝突せず、古くなった配置の除去が互いのクローンの配置を破壊しないようにするため
  である（home mode（1 ユーザー 1 つ）では起きない問題）。`<roothash>` は解決後の絶対 root パスの
  sha256 を短縮した hex（固定長・FS 安全）でなければならない。project mode では世代をユーザーに
  公開してはならない。profile は古くなった配置の除去と世代スキップ判定の内部機構に留めなければならず、
  `rollback` / `list-generations` を出してはならない（配置物が ephemeral で rollback の意味が薄く、
  devShell キック時は戻し先 host 世代も無いため）。新 link farm derivation が profile の
  current 世代に一致し、last-applied 記録（[apply は最後に配置した manifest を last-applied 記録に書き、古くなった配置の除去の比較の基準は記録を優先して読む](last-applied-record.md)）が
  あればその `linkFarm` にも一致するなら新世代を積んではならない。profile が無ければ skip
  してはならず、profile と記録が食い違う間は通常の apply の処理を走らせなければならない。skip は必須であり、devShell / direnv 運用では `shellHook` がシェル
  再入のたびに走るため、毎回新世代を積むと世代が無限増殖するからである。system mode でも同じ
  世代スキップとドリフト修復を行わなければならない。NixOS の activation は boot のたびにも走り、
  毎回新世代を積むと boot ごとに世代が増えるからである。home mode は従来通り「適用のたびに
  新世代」のままでなければならず、世代スキップは project / system mode 限定としなければならない。ただし世代スキップ時も完全 no-op にせず、FS の確認だけは軽量に行わなければならない。
  各 entry の target を比較の基準（last-applied 記録があれば記録、無ければ profile の current
  世代）に照らして `lstat` で確認しなければならず、記録通りでない（foreign tool に
  書き換えられた・消えた）entry があればその entry だけ再張りしなければならない
  （foreign symlink なら warning）。「derivation 同一 ⇒ FS 同一」は foreign 書き換えで崩れるため、
  新世代を積まずに FS だけ収束させなければならない。lstat 比較は安価で `shellHook` の高頻度実行に
  耐える。このドリフト修復は symlink と copy の両方を対象としなければならない。copy entry は
  target が不在のときだけ place-once で再マテリアライズしなければならず、存在するが内容が異なる
  （ユーザー編集）場合は触ってはならない（home mode の place-once と振る舞いを一致させる）。
  src 追従は `apply --recopy` 限定としなければならず、内容ハッシュ比較はしてはならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:54:59+09:00"
---
# project mode は世代を非公開にし、project / system mode は derivation 同一なら世代を積まず lstat ドリフト修復だけ行う

## 仕様

- **profile は解決済み root でキーする**（例: `<state>/nix/profiles/layat/<roothash>/<name>`）。
  同一 entrypoint を複数箇所にクローンしても profile が衝突せず、stale 除去が互いのクローンの
  配置を破壊しない。home mode（1 ユーザー 1 つ）では起きない問題。`<roothash>` は **解決後の
  絶対 root パスの sha256 を短縮した hex**（固定長・FS 安全）
- **世代はユーザーに公開しない**。profile は stale 除去 + 世代スキップ判定の内部機構に留め、
  `rollback` / `list-generations` を出さない。配置物が ephemeral で rollback の意味が薄く、
  devShell キック時は戻し先 host 世代も無いため
- **世代スキップ短絡（必須）**: 新 link farm derivation が前世代と同一なら**新世代を積まない**。
  devShell / direnv 運用では `shellHook` がシェル再入のたびに走るため、毎回新世代を積むと世代が
  無限増殖する。home mode は従来通り「適用のたびに新世代」のまま（世代スキップは project mode 限定）
- **ただし世代スキップ時も FS 検査だけは軽量に行う**（完全 no-op にしない）。各 entry の target を
  **lstat で検査**し、記録通りでない（foreign tool に書き換えられた・消えた）entry があればその
  entry **だけ**再張りする（foreign symlink なら warning）。「derivation 同一 ⇒ FS 同一」は
  foreign 書き換えで崩れるため、新世代を積まずに FS だけ収束させる。lstat 比較は安価で
  `shellHook` 高頻度実行に耐える
- この lstat ドリフト修復は **symlink と copy の両方**を対象にする。**copy entry は target が
  不在のときだけ** place-once で再マテリアライズし、**存在するが内容が異なる（ユーザー編集）
  場合は触らない**（home mode の place-once と振る舞いを一致させる。src 追従は `apply --recopy`
  限定・内容ハッシュ比較はしない）

> **上は原文の写しで、規範は frontmatter が正**。原文が参照する次の規範は本 item の担当ではない。
>
> - `<roothash>` 階層に backref `.root` を置き孤児 profile を逆引きできるようにすること →
>   [profileDir は config 専用ディレクトリとし、profile リンク・世代・pending out-link をその中に並べる](profile-on-disk-layout.md)（レイアウト）・[孤児 profile は backref で逆引き可能なまま放置許容とし、`layat prune` が root 不在の系列を削除する](project-mode-orphan-profile.md)（孤児 profile の扱い）
> - flock のキーが解決後 profileDir であること → [profileDir は config 専用ディレクトリとし、profile リンク・世代・pending out-link をその中に並べる](profile-on-disk-layout.md) / [flock の取得は既定 blocking とし --no-wait のときだけ try-lock でスキップする](flock-acquisition-mode.md)
> - `<state>` の定義 → [profileDir は config 専用ディレクトリとし、profile リンク・世代・pending out-link をその中に並べる](profile-on-disk-layout.md)
> - cross-config で同一 target を狙ったときの振動がユーザー責任であること → [同一 target を複数 config で狙うことによる振動はユーザー責任とし warning で可視化するに留める](cross-config-target-oscillation.md)
> - ドリフト修復の処理でも `--backup` の退避段が適用されること → [backup 退避は配置前除去の後・配置の前に置き、drift 修復の処理でも同じく実施する](backup-stage-position.md)
> - `rollback` / `list-generations` が home / system mode 限定であること（CLI 側の規範）→ [rollback と list-generations は home / system mode 限定にする](rollback-list-generations-home-only.md)
> - foreign symlink の warning そのもの → [symlink 配置は親 dir を作り配置元/subpath を指すリンクを張り、foreign symlink は警告して後勝ちする](symlink-placement-procedure.md)、place-once そのもの → [copy は target 不在のときだけマテリアライズする place-once で世代管理の対象外とする](copy-place-once.md)
>
> **2026-10-11 追記（[世代スキップには profile に加え記録の一致も要する](../adr/generation-skip-requires-both-match.md)）**: 上の写しの「新 link farm derivation が前世代と同一なら」は、新 link-farm が
> profile の current 世代に一致し、last-applied 記録があればその `linkFarm` にも一致するなら、と読み替える
> （profile が無ければ skip しない）。profile と記録が食い違う間は通常の処理で記録を比較の基準に古くなった配置の除去とコミットを行う。
> lstat での確認の「記録通り」も記録を優先した比較の基準に照らす。記録の規範は [apply は最後に配置した manifest を last-applied 記録に書き、古くなった配置の除去の比較の基準は記録を優先して読む](last-applied-record.md) の担当。
>
> **2026-10-24 追記**: 上の写しの「世代スキップは project mode 限定」は「project / system mode 限定」と読み替える。
> system mode も同じ条件で世代を積まずドリフト修復だけ行い、home / fixed mode は従来どおり適用のたびに新世代を積む。
> NixOS の activation は boot のたびにも実行されるため、system mode で毎回世代を積むと boot ごとに世代が増える。

## 出典

`docs/spec.md`「世代管理仕様」→「project mode の世代」節の箇条書き第 1〜4 項。

決定の実体は [module と project では世代操作を公開しない](../adr/module-and-project-hide-generation-ops.md)、
[ephemeral 配置は git に干渉しない](../adr/ephemeral-no-git-interference.md)と [project mode では同じ link farm なら世代を積まず drift だけを直す](../adr/project-mode-generation-skip-and-drift-repair.md)。
profile を解決済み root でキーすることと `<roothash>` の定義は [profile は解決済み root の roothash でキーする](../adr/profile-keyed-by-roothash-of-resolved-root.md)、`<state>` の確定は
[profile の基底ディレクトリをモードで切り替える](../adr/profile-base-directory-per-mode.md)、copy のドリフト修復の限定は [世代スキップを project mode と system mode で行う](../adr/generation-skip-project-and-system.md) が定めている。世代スキップの判定元を profile と
last-applied 記録の両方一致にし、比較の基準を記録優先にしたのは [世代スキップには profile に加え記録の一致も要する](../adr/generation-skip-requires-both-match.md)・[古くなった配置の除去は、比較の基準を記録から優先して読む](../adr/baseline-prefers-record.md)。

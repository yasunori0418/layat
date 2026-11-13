---
id: "REQ-0bd55dfc-b28c-4ad2-ad50-bb1f76b2246c"
type: requirement
name: "copy は src ツリー内の symlink を deref せず symlink のまま複製する"
section: placement-copy
status: stable
derives_from:
  - "UC-403fbe32-b146-401b-8b53-fe67c1e169c5"
specification: |
  A symlink inside the source tree SHALL be reproduced as a symlink and SHALL NOT be
  dereferenced, so as to avoid cycles and size inflation. A relative symlink SHALL be kept
  as it stands.
specification_ja: |
  src ツリー内の symlink は symlink のまま複製しなければならず、deref してはならない
  （循環・サイズ膨張を避けるため）。相対 symlink はそのまま保たなければならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:29:58+09:00"
---
# copy は src ツリー内の symlink を deref せず symlink のまま複製する

## 仕様

**src ツリー内の symlink は symlink のまま複製する**（deref しない。循環・サイズ膨張回避）。
相対 symlink はそのまま保つ。

> **上は原文の写しで、規範は frontmatter が正**。原文はこれに続けて「ただし store 内への
> 絶対 symlink を複製すると store 依存（read-only / GC 後 dangling）が残る点に注意」と
> 注記するが、これは規範ではなく利用した結果起きることへの注意なので規範文には含めない。

この複製規則は通常の place-once コピーと `apply --recopy` の再コピーの双方に適用される。
place-once そのものは [copy は target 不在のときだけマテリアライズする place-once で世代管理の対象外とする](copy-place-once.md)、recopy そのものは [apply --recopy は config 内の全 copy target を src から無条件に上書き再コピーする](apply-recopy.md) / [recopy の上書きは削除ではなく同一親内への rename 退避で行う](recopy-rename-aside.md) の担当。

## 出典

`docs/spec.md`「配置動作仕様」→「copy モード（place-once・ユーザー管理）」節の箇条書き第 5 項
（および「recopy」節の「symlink 複製は通常 copy と同じ」）。

決定の実体は [copy 内の symlink は symlink のまま複製する](../adr/symlink-inside-copy-kept.md)。

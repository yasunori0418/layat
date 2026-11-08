---
id: "RISK-1a8f91c8-dfaa-4598-805a-a3b71cb967f0"
type: risk
name: "非 root の system mode の変更系操作が権限の確認を経ずに FS へ触れる"
status: stable
likelihood: medium
impact: medium
level: medium
threatens:
  - "REQ-0d1f557b-7ae7-41ea-b17b-cdb5c80935d4"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:53:31+09:00"
---
# RISK-1a8f91c8-dfaa-4598-805a-a3b71cb967f0: 非 root の system mode の変更系操作が権限の確認を経ずに FS へ触れる

## リスク

system mode の変更系操作は、euid 0 の確認を通ってから FS に触れる
（→ REQ-0d1f557b-7ae7-41ea-b17b-cdb5c80935d4）。これが崩れる形は 3 つある。

- **条件の抜け・順序の誤り** — `apply` / `reset` / `rollback` のどれかで確認が無いか、
  profileDir や backref の作成・flock の取得より後に置かれる。非 root の実行が途中まで進み、
  最初に触れた場所の権限エラーという偶発的な形で止まる。`--root` を書き込み可能な場所へ
  向けていれば、配置まで進んでしまう
- **条件の過剰** — `apply --dryrun` / `reset --dryrun` / `list-generations` まで確認し、
  非 root の計画確認と世代閲覧ができなくなる
- **基底の取り違え** — system mode の profileDir や記録が実行ユーザーの state 基底に解決される、
  あるいは基底の切替が home / project / fixed にも波及し、既存の profile と記録が別の場所へ移る。
  前者はシステムの状態が個人の領域に置かれ、後者は既存の世代と比較の基準を見失わせる

## 影響

impact を medium とするのは、どの形でも利用者の実データは上書きも削除もされないため。
条件が抜けた非 root の実行は、既定の基底（`/nix/var/nix/profiles/layat` と `/var/lib/layat`）
への書き込みで権限エラーになり、undo journal が配置を巻き戻す。残るのは空のディレクトリか、
説明の乏しいエラーにとどまる。基底の取り違えは、既存の世代と記録を消さずに別の場所から
始め直す形になり、正しい基底へ戻せば復旧する。

likelihood を medium とするのは、確認を 3 つの操作へ別々に置き、操作ごとに置く位置
（dryrun の短絡の後・解決の直後）が違うため。基底の切替も profile と記録の 2 か所に分かれる。

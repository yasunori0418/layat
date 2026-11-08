---
id: "RISK-f522a51b-db4b-4164-bd20-63fa4a0cb27d"
type: risk
name: "prune が生存系列や判定不能な系列を孤児と読み、不可逆に削除する"
status: stable
threatens:
  - "REQ-c44433a1-7ee7-459a-9aae-7cc42166876f"
  - "REQ-d41b1d0a-c6d5-41cc-93f9-e5cc7f152da4"
likelihood: high
impact: high
level: high
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-05T00:25:57+09:00"
---
# RISK-f522a51b-db4b-4164-bd20-63fa4a0cb27d: prune が生存系列や判定不能な系列を孤児と読み、不可逆に削除する

## リスク

`layat prune` は「root が実在しない」の 1 条件だけで系列を丸ごと消す（→ [prune は backref の root が実在しない roothash 系列だけを系列ごと削除する](../requirements/prune-target-selection.md)）。
判定を誤ると、消してはいけない系列が消える。削除は `<roothash>` ディレクトリごとで、世代
リンクも `.pending` も `.root` も一度に失われるため、後から戻す手段が無い。

**生存系列の誤削除** — root が実在するのに不在と読む。`os.Stat` の戻りの扱い（`nil` を生存と
見るか、エラーの種類を区別するか）を 1 箇所でも取り違えると、使用中のプロジェクトの世代が
消える。

**`.root` を持たない系列への波及** — home mode / system mode の `<name>` 直キー系列は判定すべき
root パスを持たないため構造的に対象外だが、列挙が `.root` の有無ではなくディレクトリの形で
行われると、`$HOME` の profile 系列が走査に含まれる。走査基底が 2 つある（ユーザー state 基底と
system 基底 `/nix/var/nix/profiles/layat/` → [prune は system 基底も走査し、権限の無い系列を見送る](../adr/prune-scans-system-base-skips-unpermitted.md)）ぶん、基底ごとに列挙条件がずれる余地も
ある。

**system 基底が権限で読めない場合** — 非 root 実行では system 基底の列挙そのものが失敗しうる。
これを「孤児が 0 件」と読んで利用者に知らせずに何もしなければ、ユーザーは prune が実行されたのに何も消えない
理由を知る手掛かりを持たない。逆に一部だけ読めた状態で判定を続ければ、読めなかった系列の
存在を無視した結論になる。列挙の失敗は「対象なし」と区別して報告される必要がある。

**アンマウント中 root の誤判定** — リムーバブルディスク・ネットワークマウント上のプロジェクトは
アンマウント中「一時的に実在しない」。機械側でマウントポイント判定は行わない設計なので
（→ [prune は削除前に root パスの一覧を示して確認を取る](../adr/prune-confirms-with-root-list.md)）、判定そのものはこの誤りを含んだまま通る。

**判定不能を削除へ倒す** — backref が読めない・空・絶対パスでない、root の stat が権限等で
失敗する。これらは「孤児と確認できていない」だけであって「孤児である」ではない。skip + warning
以外へ倒すと、判定材料が欠けたときに最も危険な方向へ落ちる。

**孤児の逆引きの喪失** — clone を消した後の profile が backref `.root` を失うか `<roothash>` の階層に
置かれず、どの root 由来か辿れなくなる。prune はその系列を判定できず、孤児は手で消すしかなくなる
（→ [孤児 profile は backref で逆引き可能なまま放置許容とし、`layat prune` が root 不在の系列を削除する](../requirements/project-mode-orphan-profile.md)）。

## 実現性

**likelihood: high** — 変更頻度の側では、判定が依存する profile レイアウト（`<roothash>` 階層・
backref `.root` の位置・走査基底）が繰り返し動いてきた領域で
（現在の決定は → [profile は専用ディレクトリレイアウトに置く](../adr/profile-directory-layout.md) / [prune は system 基底も走査し、権限の無い系列を見送る](../adr/prune-scans-system-base-skips-unpermitted.md) /
[profile の基底ディレクトリをモードで切り替える](../adr/profile-base-directory-per-mode.md)）、
`layat status`（→ #198）が同じ列挙に相乗りする予定もあるため、この先も触られる。自動検証の
側では、判定ロジックが新規で既存の配置系の条件（planner の分類・古くなった配置の除去の保守的不変条件
→ [古くなった配置を保守的に除去するときの不変条件が破れ、layat が作っていない実体を消す](stale-removal-invariant-broken.md)）のどれにも乗らず、blind spot にあたる。両軸とも
高いので high。

**impact: high** — 系列削除は不可逆で、消えるのは実際に使われているプロジェクトの世代履歴。
再 apply で最新世代は作り直せても、rollback 先の履歴は戻らない。

## 緩和

[prune が孤児系列だけを系列ごと消し、生存・判定不能な系列は理由付きで残す](../test/generations/prune-target-accuracy.md)（prune の対象判定の正確さ）が緩和する。判定と削除は
engine 層に閉じているので、この TC で覆える。アンマウント中 root の誤判定だけは機械側で
区別しない設計上の残余で、確認プロンプトの root パス一覧が唯一の防波堤になる
（→ [prune の dryrun・確認・lock の安全機構が働かず、同意なく削除が実行される](prune-safety-gate-bypass.md)）。

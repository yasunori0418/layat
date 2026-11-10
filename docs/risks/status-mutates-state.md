---
id: "RISK-f87d4909-6dea-4ead-8e72-994c498a14df"
type: risk
name: "status が flock を取り、ディレクトリを作り、記録を書き換えて読み取り専用でなくなる"
status: stable
threatens:
  - "REQ-e8d911ff-8ee2-4d68-87d0-26f6d6817173"
likelihood: medium
impact: medium
level: medium
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-11T00:30:18+09:00"
---
# status が flock を取り、ディレクトリを作り、記録を書き換えて読み取り専用でなくなる

## リスク

`layat status` は読み取り専用でなければならない（→ [status は記録基底を走査して last-applied 記録を読み取り専用で表示する](../requirements/status-command.md)）。
engine には記録や profileDir を扱う関数が既にあり、その多くは apply・reset のために
ディレクトリを作るか lock を取る。status がそれらを流用すると、読むだけのはずの操作が
状態を変える。

**flock の取得** — status が profileDir の lock を取ると、同時に実行される apply・reset を待たせる。
system 基底では lock ファイルを作れず、非 root の status が失敗する。

**ディレクトリ・ファイルの生成** — profileDir・`.pending`・backref `.root` を作ると、profile 基底や
記録基底に apply していない系列の痕跡が残る。profile 基底に `.root` だけの系列ができれば prune の
走査対象になり、status を実行しただけで prune の候補が増える。非 root で system 基底に書こうとすれば権限エラーで止まる。

**記録の書き換え** — 記録を書き直すと `appliedAt` が status の実行時刻になり、manifest や
`linkFarm` を取り違えれば、次の apply の古くなった配置の除去が誤った比較の基準から始まる。

この脅威は status をどう設計しても残る要求側の脅威なので、`design` ではなく `requirement` に張る
（→ [risk はまず requirement に張り、設計を差し替えると消える懸念だけを design に張る](../quality/risk-attaches-to-requirement-by-default.md)）。

## 実現性

**likelihood: medium** — 変更頻度の側では、status は新規のコードで、記録を読む既存の処理
（`loadBaseline` など）は profileDir の解決やディレクトリの作成と隣り合っている。読み取りに
必要な部分だけを取り出すつもりで、作成まで含む関数を呼んでしまう誤りは入りやすい。検証の
条件の側では、実行の前後で基底配下が変わらないことは tmpdir で機械的に確かめられる。
両軸を合わせて medium。

**impact: medium** — 最も重い方向は記録の書き換えで、次の apply の古くなった配置の除去を誤った比較の基準
から始めさせる。古くなった配置の除去は layat が置いた symlink だけを消す保守的な不変条件
（→ [古くなった配置を保守的に除去するときの不変条件が破れ、layat が作っていない実体を消す](stale-removal-invariant-broken.md)）に守られており、利用者の実データには届かない。
誤った配置は、正しい記録を書き直す再 apply で収束する。lock の取得とディレクトリの生成も、
後から消せば元に戻る。

## 緩和

[status が記録基底の全系列を漏れなく列挙し、記録の内容をそのまま反映し、状態を変えない](../test/generations/status-inventory-fidelity.md) が緩和する。実行の前後で `StateDir` / `SystemStateDir`
配下のエントリが増減しないこと、記録ファイルの内容が変わらないこと、profileDir の lock を
保持した状態でも待たずに返ることを検証する。

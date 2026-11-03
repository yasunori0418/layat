---
id: "ADR-abfe1b12-1a7d-42e1-8520-c1e9fcf5846c"
type: adr
name: "system mode の last-applied 記録を /var/lib/layat に置く"
status: 採用
issues:
  - "#140"
origin: "Issue #140（epic #129: system mode の実装）の計画（2026-10-18）。ADR-0062 §1・ADR-0066 の既知の注意点（macOS の記録基底）を書き直した"
justifies:
  - "REQ-0d1f557b-7ae7-41ea-b17b-cdb5c80935d4"
references:
  - "ADR-0036"
  - "ADR-0058"
---
# ADR-abfe1b12-1a7d-42e1-8520-c1e9fcf5846c: system mode の last-applied 記録を /var/lib/layat に置く

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-0036, ADR-0058
- 起点: Issue #140（epic #129: system mode の実装）の計画（2026-10-18）。ADR-0062 §1・ADR-0066 の既知の注意点（macOS の記録基底）を書き直した

## 背景

last-applied 記録は、実行ユーザーの state（`$XDG_STATE_HOME`、既定は `$HOME/.local/state`）の下に置く。
このままだと、root で実行した system mode の記録は `/root/.local/state/layat` に置かれる。system mode の
profile は、システムの状態が root の個人領域に住む形を避けて `/nix/var/nix/profiles/layat` に置いている。
記録も同じ理由で個人領域に置けない。

一方で記録は nix profile ではなく、GC root でもない。profile と同じ基底に置く理由も無い。

macOS には `/var/lib` が無く、記録の置き場には `/var/db` を使う慣習がある。

## 決定

system mode の last-applied 記録は `/var/lib/layat` の下に、profileDir と同じ形の鍵で置く。macOS でも
基底は `/var/lib/layat` のままとし、`/var/db` へ切り替える分岐は足さない。`/var/lib/layat` が無ければ
root で実行する engine が作る。

home / project / fixed の記録の置き場と、記録の内容・書き込みの手順は system mode でも変えない。

置き場のパスと鍵の形は REQ-0d1f557b-7ae7-41ea-b17b-cdb5c80935d4 に置く。

## 根拠

記録は layat が自分で読み書きする可変の状態で、FHS の `/var/lib/<パッケージ>` がその置き場にあたる。
nix profile でも GC root でもないので、nix の慣習に合わせる必要は無い。

profile の基底は世代バックエンドの所有物で、記録は世代と独立した layat の所有物になる。user mode でも
記録（`<state>/layat`）と profile（`<state>/nix/profiles/layat`）は別の基底にある。同居させると、
`layat prune` の走査対象に記録のディレクトリが入る。

macOS で分岐しなければ、記録の置き場が OS ごとに分かれず、engine の解決と e2e の検証が 1 通りで済む。
`/var/lib` が無くても、root の engine が作れば足りる。

## 棄却した代替案

### 実行ユーザーの state のまま置く

root なら `/root/.local/state/layat` になる。実装の差分は無いが、システムの状態が root の個人領域に置かれ、
将来のモジュール導入時に記録を移す痛みが残る。

### `/nix/var/nix/profiles/layat` の下に同居させる

置き場の数は減るが、世代バックエンドの所有する基底に layat 自前の状態を持ち込み、prune の走査と衝突する。

### macOS では `/var/db/layat` に置く

macOS の慣習には合うが、記録の置き場が OS ごとに分かれ、解決と検証を OS ごとに 2 通り持つことになる。

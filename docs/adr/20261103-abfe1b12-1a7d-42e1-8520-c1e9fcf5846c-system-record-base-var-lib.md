---
id: "ADR-abfe1b12-1a7d-42e1-8520-c1e9fcf5846c"
type: adr
name: "system mode の last-applied 記録を /var/lib/layat に置く"
status: stable
sources:
  - id: "issue-140"
    resource: "https://github.com/yasunori0418/layat/issues/140"
    title: "#140"
origin: "Issue #140（epic #129: system mode の実装）の計画（2026-10-18）。ADR-0062 §1・ADR-0066 の既知の注意点（macOS の記録基底）を書き直した"
justifies:
  - "REQ-0d1f557b-7ae7-41ea-b17b-cdb5c80935d4"
  - "REQ-9d622539-a1e5-4719-af51-e0cd85b66f35"
references:
  - "ADR-9bdf2604-c7b3-468a-b1e4-2478ef37477c"
  - "ADR-400cadd4-0151-470b-92ae-9b11aa8e9bed"
  - "ADR-25d54346-f13a-4c12-9895-ab93dc6c183b"
  - "ADR-11507e12-5b0f-4f8f-bae9-fb7fd41a9203"
  - "ADR-ad1c2507-1f42-4cd3-ab56-9183892f8988"
  - "ADR-7a3dc78d-8946-46ae-94ec-b752f2ca7c2e"
  - "ADR-15a358f4-1c81-49d5-b658-3a1fe39f0f75"
  - "ADR-df8ab6ff-7636-42ff-8b47-6ee1209dfd15"
  - "ADR-52423842-bacf-4773-bbca-bcb34ae71013"
  - "ADR-10c6dc33-d188-406b-beb4-d10ccbe3c645"
  - "ADR-a91585ad-971a-4558-9f25-b7fcd6e57671"
  - "ADR-a6a1ef1b-676c-484a-9c8a-3e6d49c1b444"
  - "ADR-be5865a8-000a-491f-8438-6308076ea727"
  - "ADR-facd3173-a5ce-4d81-90a9-370fb26428ff"
  - "ADR-f3d06307-ae5b-4f17-94ad-db92e80e3415"
  - "ADR-fd3353cf-00af-4343-bd3f-3369c21d7194"
  - "ADR-8e253f2e-c255-482a-a0f8-0837e122e80f"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:28:06+09:00"
---
# ADR-abfe1b12-1a7d-42e1-8520-c1e9fcf5846c: system mode の last-applied 記録を /var/lib/layat に置く

- ステータス: stable
- 日付: 2026-11-03
- 関連: ADR-9bdf2604-c7b3-468a-b1e4-2478ef37477c, ADR-400cadd4-0151-470b-92ae-9b11aa8e9bed, ADR-25d54346-f13a-4c12-9895-ab93dc6c183b, ADR-11507e12-5b0f-4f8f-bae9-fb7fd41a9203, ADR-ad1c2507-1f42-4cd3-ab56-9183892f8988, ADR-7a3dc78d-8946-46ae-94ec-b752f2ca7c2e, ADR-15a358f4-1c81-49d5-b658-3a1fe39f0f75, ADR-df8ab6ff-7636-42ff-8b47-6ee1209dfd15, ADR-52423842-bacf-4773-bbca-bcb34ae71013, ADR-10c6dc33-d188-406b-beb4-d10ccbe3c645, ADR-a91585ad-971a-4558-9f25-b7fcd6e57671, ADR-a6a1ef1b-676c-484a-9c8a-3e6d49c1b444, ADR-be5865a8-000a-491f-8438-6308076ea727, ADR-facd3173-a5ce-4d81-90a9-370fb26428ff, ADR-f3d06307-ae5b-4f17-94ad-db92e80e3415, ADR-fd3353cf-00af-4343-bd3f-3369c21d7194, ADR-8e253f2e-c255-482a-a0f8-0837e122e80f
- 起点: Issue #140（epic #129: system mode の実装）の計画（2026-10-18）。ADR-0062 §1・ADR-0066 の既知の注意点（macOS の記録基底）を書き直した

## 背景

last-applied 記録は、実行ユーザーの state（`$XDG_STATE_HOME`、既定は `$HOME/.local/state`）の下に置く。
このままだと、root で実行した system mode の記録は `/root/.local/state/layat` に置かれる。system mode の
profile は、システムの状態が root の個人領域に置かれる形を避けて `/nix/var/nix/profiles/layat` に置いている。
記録も同じ理由で個人領域に置けない。

一方で記録は nix profile ではなく、GC root でもない。profile と同じ基底に置く理由も無い。

macOS には `/var/lib` が無い。

## 決定

system mode の last-applied 記録は `/var/lib/layat` の下に、profileDir と同じ形の鍵で置く。macOS でも
基底は `/var/lib/layat` のままとし、`/var/db` へ切り替える分岐は足さない。`/var/lib/layat` が無ければ
root で実行する engine が作る。

home / project / fixed の記録の置き場と、記録の内容・書き込みの手順・古くなった配置を除去するときの基準をどこから読むかの優先順位は system mode でも変えない。

置き場のパスと鍵の形は REQ-0d1f557b-7ae7-41ea-b17b-cdb5c80935d4 と
REQ-9d622539-a1e5-4719-af51-e0cd85b66f35 に置く。

## 根拠

記録は layat が自分で読み書きする可変の状態で、FHS の `/var/lib/<パッケージ>` がその置き場にあたる。
nix profile でも GC root でもないので、nix の慣習に合わせる必要は無い。

profile の基底は世代バックエンドの所有物で、記録は世代と独立した layat の所有物になる。user mode でも
記録（`<state>/layat`）と profile（`<state>/nix/profiles/layat`）は別の基底にある。同居させると、
`layat prune` の走査対象に記録のディレクトリが入る。

macOS でも分岐を足さないので、記録の置き場は OS によらず 1 つになる。`/var/lib` が無くても、root で
実行する engine が作る。

## 棄却した代替案

### 実行ユーザーの state のまま置く

root なら `/root/.local/state/layat` になる。実装の差分は無いが、システムの状態が root の個人領域に置かれ、
将来のモジュール導入時に記録を移す手間が残る。

### `/nix/var/nix/profiles/layat` の下に同居させる

置き場の数は減るが、世代バックエンドの所有する基底に layat 自前の状態を持ち込み、prune の走査と衝突する。

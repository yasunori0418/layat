---
id: "RISK-238bc9da-353b-4699-94f2-ea871e305f45"
type: risk
name: "root 相対 src が誤った基準点に解決される、または target と重なって symlink ループになる"
likelihood: medium
impact: medium
level: medium
threatens:
  - "REQ-a77181dd-a59c-40bb-855e-3430ad970eb6"
  - "DSG-fd39b9ee-a5d2-46c2-bdcd-fc399a1e0296"
---
# RISK-238bc9da-353b-4699-94f2-ea871e305f45: root 相対 src が誤った基準点に解決される、または target と重なって symlink ループになる

## リスク

root 相対 src は `target` と同じ配置 root に対して解決され、`target` の親から
`<src>/<subpath>` への相対リンクとして張られる（→ REQ-a77181dd-a59c-40bb-855e-3430ad970eb6・
DSG-fd39b9ee-a5d2-46c2-bdcd-fc399a1e0296）。これが崩れる形は 3 つある。

- **基準点の取り違え** — `LinkDest` が `target` の親ではなく root や cwd を起点に相対化する、
  あるいは preflight が root ではなく cwd を基準に実在を確かめる。リンクは別の場所を指すか
  dangling になり、preflight は誤った実在判定で止まらない
- **前世代との照合の崩れ** — 相対の文字列が世代間で揺れる（`filepath.Clean` の有無・subpath
  `.` の扱い）と、readlink と `LinkDest` の完全一致が外れる。同じ設定の再 apply が foreign
  置換の warning を出し、entry を外しても stale 除去されずに残る
- **target との重なり** — `<src>/<subpath>` と `target` が一致するか一方が他方の祖先のとき、
  評価時ゲートが漏れるとリンクが自分自身か自分の祖先を指し、symlink ループになる

## 影響

impact を medium とするのは、どの形でも誤って張られたり残ったりするのは symlink だけで、
利用者の実データは上書きも削除もされないため。誤ったリンクは設定を直して再 apply すれば置き換わり、残った
リンクは手で消せる。気付かれにくい点（相対リンクは worktree 内では正しく見えることが多い）は
採点に入れない。

likelihood を medium とするのは、`LinkDest`・preflight・評価時ゲートの 3 箇所が同時に新設
され、相対化の基準点と文字列の正規化を既存テストが見ていないため。

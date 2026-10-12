---
id: "TC-f1da084f-c0f0-4aba-85d9-f273c04eb373"
type: test_condition
name: "root 相対 symlink が配置 root に対して相対形で張られ、worktree 移動後も有効で、前世代との照合が崩れない"
mitigates:
  - "RISK-238bc9da-353b-4699-94f2-ea871e305f45"
---
# TC-f1da084f-c0f0-4aba-85d9-f273c04eb373: root 相対 symlink が配置 root に対して相対形で張られ、worktree 移動後も有効で、前世代との照合が崩れない

## 条件

root 相対 src の解決が配置 root と `target` の位置だけで決まることを確かめる。

- **相対形で張られる** — 配置後の readlink が `target` の親から `<src>/<subpath>` への相対
  文字列に一致し（subpath があってもなくても）、リンク越しに内容へ到達できること
- **基準点が root である** — preflight が `<root>/<src>/<subpath>` の実在で判定し、不在なら
  dangling を作らずに停止すること。cwd を root 以外に置いても結果が変わらないこと
- **上書き後の root が基準点になる** — `--root` で root を上書きしたとき、実在判定と配置が
  上書き後の root に対して行われ、元の root には何も置かれないこと
- **移動に耐える** — リポジトリのディレクトリごと移動した後も readlink が変わらず、内容へ
  到達できること
- **前世代との照合** — 同じ設定の再 apply が warning 無しで readlink を変えないこと、entry を
  足した再 apply で既存の root 相対リンクが不変であること、entry を外した apply で stale
  除去されること
- **重なりの拒否** — `<src>/<subpath>` と `target` が一致するか一方が他方の祖先である entry が
  評価時に拒否されること

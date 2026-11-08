---
id: "RISK-e916f742-59a3-4202-a316-603a908fd0db"
type: risk
name: "entrypoint の形ごとに root 解決と配置の意味論が食い違う"
status: stable
threatens:
  - "REQ-9cb26ffd-071e-4c68-a6fc-faac6373b75e"
  - "REQ-6506bc82-d1e1-4dbf-8c57-d5d1babf218a"
  - "REQ-c890ce4a-6528-4ab3-ac86-23d7aebff7da"
  - "REQ-da253cab-34d4-4d6e-96f0-de99e012b376"
  - "DSG-92f54490-872a-42ac-bbd7-d06e9ee381c6"
  - "REQ-c50df875-2cb0-4e72-8a21-858359a11cae"
likelihood: medium
impact: high
level: high
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:36:11+09:00"
---
# RISK-e916f742-59a3-4202-a316-603a908fd0db: entrypoint の形ごとに root 解決と配置の意味論が食い違う

## リスク

layat は flake の entrypoint と legacy の非 flake entrypoint という 2 つの起点を持ち、
project mode ではさらに git から root を解決する。起点が増えるほど「同じ配置のはずなのに
起点によって結果が違う」という食い違いが入り込む。

- **root 解決の誤り** — project mode の root は git の toplevel であり、config の位置でも
  カレントディレクトリでもない。リポジトリ内のどこから呼んでも同じ場所へ解決されなければ、
  呼び出した場所によって配置先が変わる。逆にリポジトリの外で呼ばれたときは、曖昧に成功せず
  停止しなければならない。祖先がたまたまリポジトリであるだけで誤って成功するのが典型的な
  失敗の形である
- **起点ごとの意味論の分岐** — legacy の使い方でも同じ config の指定・全件適用・出力契約が成り
  立つ必要がある。分岐が attr path の組み立てを越えて広がると、起点ごとに別実装を抱えること
  になる。この分岐の閉じ込め方は設計上の選択であり、別の設計を採れば消える `risk` なので
  `design` にも張る
- **既存の利用体験の破壊** — legacy の起点へ layat のための属性を足したことで、素の nix-shell
  としての利用ができなくなる
- **store 化の差の取り違え** — flake の使い方では相対パスの配置元が store へ入るのに対し、
  legacy の impure な評価では作業木のパスのまま解決される。この差は意図されたものであり、
  同一の性質を両方の使い方へ期待すると誤った検証をすることになる
- **flake-parts の使い方の分岐** — `perSystem.layat.<name>` の transpose が直書きと別の derivation を
  生むか、perSystem と別の `pkgs` で評価され、flake-parts を使うときだけ CLI のアドレッシングが
  変わる（→ [flake-parts での使い方は直書きと同一の derivation を生み CLI のアドレッシングを変えない](../requirements/flake-parts-module-path.md)）

## 影響

意図しない場所への配置、あるいは起点によって再現しない挙動。impact を high とするのは、
root 解決の誤りが「利用者の意図しないディレクトリへ書き込む」形で現れ、どこへ書いたかが
記録に残らない以上、原状回復の手立てが無いため。

likelihood を medium とするのは、project mode の root 解決は [project mode の root が git toplevel へ一意に解決し外では曖昧に成功しない](../test/integration/project-root-resolution.md) が、legacy 起点の
配置と出力契約の同等性は [legacy entrypoint が canonical 形で同じ配置と出力契約を保ち素の nix-shell を壊さない](../test/integration/legacy-entrypoint-parity.md) が覆っている一方、起点が flake / legacy の 2 つに
分かれ、そこへ git 解決が掛かるため、片方の起点だけを見た変更が他方との食い違いを作りうる
ため。engine 層で同じ root 解決を扱う [root 解決を誤り、意図しないディレクトリツリーへ配置する](root-resolution-wrong-tree.md) が low なのに対して本 item が medium に
留まるのは、起点ごとの分岐という掛け算の側を持つぶん混入の余地が広いことによる。

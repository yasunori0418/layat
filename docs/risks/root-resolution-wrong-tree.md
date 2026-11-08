---
id: "RISK-24e0805d-53cd-40dd-9e7a-b5c4bbf2a298"
type: risk
name: "root 解決を誤り、意図しないディレクトリツリーへ配置する"
status: stable
threatens:
  - "REQ-9cb26ffd-071e-4c68-a6fc-faac6373b75e"
  - "REQ-6506bc82-d1e1-4dbf-8c57-d5d1babf218a"
  - "REQ-8d965ca2-f8fd-44a4-87f3-94e850e9f85b"
  - "REQ-61c05e09-0bde-4f74-9a96-03185f9df606"
likelihood: low
impact: high
level: medium
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:36:11+09:00"
---
# RISK-24e0805d-53cd-40dd-9e7a-b5c4bbf2a298: root 解決を誤り、意図しないディレクトリツリーへ配置する

全ての配置先は `<root>/<target>` として組み立てられるため、root の解決を誤ると entry 単位の
正しさに関わらず配置全体が別のツリーへ流れ込む。project mode の root を CWD 相対や config
相対で採ってしまえば、リポジトリ外へ配置物を撒く。home mode の root を層ごとに定まった供給元
以外から採れば、HM activation と CLI 直叩きで配置先が食い違う。

git から root を解決できないときに engine 実行時ではなく評価時にエラーで止まる／逆に利用者に知らせずに fallback
する、という層分けの崩れも同じ脅威に含む。fallback すると「git repo の外で apply した」ことに
ユーザーが気づけない。

## 想定する失敗

- project mode で git toplevel ではなく CWD を root に採り、リポジトリ外へ配置する
- git repo 外での apply が停止せず、任意のディレクトリを root に採る
- `fixed` で root パスを省いた文書を受理したまま停止せず、root 未決で配置へ進む
  （契約の消費側は受理の段でこの組み合わせを判断しない層分けのため、この形の失敗は root 解決の
  側にしか現れない）
- home mode の root 供給元が層ごとに揃わず、HM 経由と CLI 経由で配置先が分岐する
- profile ディレクトリの作成・backref 書き込みの失敗が握り潰され、逆引き不能な状態で配置が進む
- `--root` を渡しても、project mode が git toplevel、home mode が `$HOME` を採り続け、明示した
  root が使われない（→ [--root は全モード共通で解決 root を明示上書きする](../requirements/root-override-flag.md)）

## 評価

- likelihood: low — 解決規則はモードごとに 1 箇所へ集約されており、[root 解決と profile ディレクトリ準備の成功時・失敗時の処理が仕様どおりであること](../test/engine-core/root-resolution.md) が成功する場面と
  失敗する場面（git repo 外での停止・`fixed` でのパス無しの拒否・profile 準備の失敗）の双方を
  機械的に覆っている。ただし同 TC は「想定する失敗」4 点目に現れる root 供給元の層分けそのものを
  `integration` 区分へ明示的に委ねている（rootKind から絶対 root への解決分岐は home mode 側も
  含めて engine 層で覆う）。委ねたこの 1 点は残余だが、engine 層の解決規則そのものが
  崩れる場面ではない
- impact: high — 誤った root への配置はユーザー環境の任意の場所へ書き込むことであり、
  どこへ撒いたかが記録に残らない以上、原状回復の手立てが無い

## 張り先の判断

4 本とも `requirement` へ張る。root をどこから解決するかはユーザーが直接触れる契約であり、
解決の処理の実装を差し替えても「意図しないツリーへ配置される」懸念は消えない。

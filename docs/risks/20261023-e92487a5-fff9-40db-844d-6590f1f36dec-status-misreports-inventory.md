---
id: "RISK-e92487a5-fff9-40db-844d-6590f1f36dec"
type: risk
name: "status が系列を取りこぼし、mode / root を取り違え、不正な記録を記録なしと読む"
threatens:
  - "REQ-e8d911ff-8ee2-4d68-87d0-26f6d6817173"
likelihood: medium
impact: medium
level: medium
---
# RISK-e92487a5-fff9-40db-844d-6590f1f36dec: status が系列を取りこぼし、mode / root を取り違え、不正な記録を記録なしと読む

## リスク

`layat status` は「今なにが apply されているか」を確かめるためのコマンドで
（→ REQ-e8d911ff-8ee2-4d68-87d0-26f6d6817173）、一覧が実態とずれると、利用者はずれた一覧を
もとに判断する。どの方向のずれも、status 自身はエラーを出さない。

**系列の取りこぼし** — 記録基底は `<name>` 直キーと `<roothash>/<name>` 系列の 2 つの形を持ち、
基底も user と system の 2 つある。どちらかの形、どちらかの基底を走査し損ねると、apply 済みの
config が一覧から消える。system 基底が読めないのを「記録 0 件」と読むのも同じ結果になる。

**mode / root の取り違え** — subject 名の mode は、直キーでは走査した基底から、`<roothash>`
系列では記録の manifest の `rootKind` から決まる。root は backref `.root` の値である。取り出す
場所を取り違えると、project mode の config を home と表示したり、別の root の系列を同じ root の
ものとして表示したりする。

**不正な記録を「記録なし」と読む** — 壊れた記録を警告なしに飛ばすと、その config は一覧に現れず、
記録が壊れていることに気付く機会を失う。次の apply は同じ記録を baseline として読んで失敗する
ので、利用者は status では見えなかった問題に apply の時点で初めて気付く。

## 実現性

**likelihood: medium** — 変更頻度の側では、記録基底のレイアウトは ADR-0058 / ADR-0062 で
最近動いた領域で、`ListRootHashSeries` は prune と共有するので、prune 側の都合で列挙条件が
変わりうる。検証ゲートの側では、status の走査は新規のコードで、既存のテストはどれも覆って
いない。ただし列挙と並び順は tmpdir の実 FS で決定的に検証でき、blind spot に
なりにくい。両軸を合わせて medium。

**impact: medium** — status は読み取り専用で、ずれた一覧は利用者の環境を壊さない。誤りに
気付けば、修正後の再実行で正しい一覧が得られる。最も重い方向は不正な記録の見落としだが、
これも次の apply が error で止めるので、取り返しのつかない状態には至らない。

## 緩和

TC-076bd7b4-6427-4a35-9bab-86daaaada504（一覧の網羅と記録の反映）が engine の走査を、
TC-1bc3e39a-116a-4f26-96b4-1d15e6f837bd（CLI の出力契約）が subject 名・絞り込み・error 時の
続行を緩和する。

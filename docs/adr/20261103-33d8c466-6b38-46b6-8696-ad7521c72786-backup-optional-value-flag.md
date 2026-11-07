---
id: "ADR-33d8c466-6b38-46b6-8696-ad7521c72786"
type: adr
name: "--backup を = 区切りで suffix を取る省略可能な値のフラグにする"
status: stable
sources:
  - id: "issue-169"
    resource: "https://github.com/yasunori0418/layat/issues/169"
    title: "#169"
  - id: "issue-172"
    resource: "https://github.com/yasunori0418/layat/issues/172"
    title: "#172"
origin: "旧 epic #167 から統合された epic #172 の grilling セッション（2026-07-11・2026-07-12）で確定。ADR-0045 §1（CLI）のフラグの形を書き直した"
justifies:
  - "REQ-5dd5a4e9-6162-4fa5-b295-66844f5a4f3b"
  - "REQ-687e225f-5046-46db-88fb-f9e527a1e97a"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T15:57:37+09:00"
---
# ADR-33d8c466-6b38-46b6-8696-ad7521c72786: --backup を = 区切りで suffix を取る省略可能な値のフラグにする

- ステータス: stable
- 日付: 2026-11-03
- 関連: GitHub Issue #169, #172
- 起点: 旧 epic #167 から統合された epic #172 の grilling セッション（2026-07-11・2026-07-12）で確定。ADR-0045 §1（CLI）のフラグの形を書き直した

## 背景

layat は、配置先に記録外の実体があると conflict で止まり、上書きしない。利用者のデータを
誤って失わないための既定だが、「このファイルは確かめたので退かしてよい」という場面も多い
（dotfiles を layat の管理へ移す初回、手で作った設定ファイルを置き換えるときなど）。これまでは
手で `mv` してから再実行するしかなく、conflict が複数あると確認・退避・再実行を 1 件ずつ
繰り返すことになった。そこで、利用者が明示したときだけ記録外の実体を退避する `apply` の
フラグを足す。そのフラグの形を決める必要があった。

## 決定

`apply --backup[=<suffix>]` とする。値を付けなければ既定の suffix を使い、`=` で付ければ
その suffix を使う。cobra の省略可能な値のフラグ（`NoOptDefVal`）で実装するので、スペース
区切りの `--backup bak` は suffix として扱わない。`--backup` は apply 全体の修飾とし、entry
ごとには指定させない。

既定の suffix とフラグの書式は REQ-5dd5a4e9-6162-4fa5-b295-66844f5a4f3b に、`--all` との
合成は REQ-687e225f-5046-46db-88fb-f9e527a1e97a に置く。

## 根拠

ふだんは `--backup` だけで済み、退避先の名前を変えたいときだけ値を書けばよい。スペース
区切りを受け付けないのは `NoOptDefVal` に共通する制約で、受け入れて仕様に明記する。

「この apply だけ安全側に倒して退かしたい」という要求は、ほぼ全 entry に対する一括の要求に
なる。apply 全体の修飾にしておけば、`--recopy` と同じく `--all` とも直交して合成できる。

## 棄却した代替案

### entry ごとに退避の可否を指定させる

柔軟に見えるが、実際の要求は apply 単位の一括である。使われない粒度のために設定項目を
増やすことになる。

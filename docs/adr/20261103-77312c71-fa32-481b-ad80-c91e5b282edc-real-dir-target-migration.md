---
id: "ADR-77312c71-fa32-481b-ad80-c91e5b282edc"
type: adr
name: "実ディレクトリの target は全 leaf を除去できるときだけ移行する"
status: 採用
issues:
  - "#175"
  - "#172"
origin: "epic #172 の grilling セッション（2026-07-12）で確定。ADR-0047 §1 を書き直した"
justifies:
  - "REQ-7cee95dd-bc5a-4e86-bebc-6080ef78fe26"
  - "REQ-053cfed2-265a-4997-a27b-97b0bed10d8a"
references:
  - "ADR-b06dd8ab-317e-41ee-9c80-dd70dcd7f2fd"
  - "ADR-3b8175b9-5ebf-4123-8855-f698dc843d6b"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-05T00:35:05+09:00"
---
# ADR-77312c71-fa32-481b-ad80-c91e5b282edc: 実ディレクトリの target は全 leaf を除去できるときだけ移行する

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-b06dd8ab-317e-41ee-9c80-dd70dcd7f2fd, ADR-3b8175b9-5ebf-4123-8855-f698dc843d6b, GitHub Issue #175, #172
- 起点: epic #172 の grilling セッション（2026-07-12）で確定。ADR-0047 §1 を書き直した

## 背景

前世代が `foo/main.sh` をファイルごとに置き、新しい世代がその親の `foo` をディレクトリの
symlink にすると、`foo` は実ディレクトリとして残っているので conflict になる。2026-07-12 の
実障害（`.claude/hooks/<name>/main.sh` を `.claude/hooks/<name>` のディレクトリ symlink に
まとめた変更）がこの形だった。祖先の symlink の移行だけでは、target そのものが実ディレクトリの
場合を救えない。配置前除去（ADR-3b8175b9-5ebf-4123-8855-f698dc843d6b）の対象を広げるかを
決める必要があった。

## 決定

symlink の entry の target が実ディレクトリのとき、配下の全 leaf が「自己記録の古くなった
symlink」か「空のディレクトリ」なら、ディレクトリごと配置前に除去して symlink を置く。それ
以外の leaf が 1 つでもあれば、ディレクトリ全体を conflict にし、部分的には除去しない。空の
ディレクトリは由来を問わない。copy の entry には適用しない。

leaf の判定と除去の順は REQ-7cee95dd-bc5a-4e86-bebc-6080ef78fe26 に、実ディレクトリ以外の
実体で止まる規範は REQ-053cfed2-265a-4997-a27b-97b0bed10d8a に置く。

## 根拠

空のディレクトリの rmdir は空のときにしか成功せず、データを失うことが原理的に無い。利用者は
その target を config で宣言している。layat が作ったディレクトリかどうかを見分けるコストを
払う理由が無い。

中身のある foreign な leaf が 1 つでもあれば全体を止めるのは、部分的な除去が利用者の意図
しない中間の状態を生むからである。移行できるところだけ移行する設計は、エラーに気付きにくく
する。

## 棄却した代替案

### 自動で作った親ディレクトリを manifest に記録する

layat が作ったディレクトリを追えれば判定は楽になるが、manifest の schemaVersion 1 の契約
（ADR-b06dd8ab-317e-41ee-9c80-dd70dcd7f2fd）を破る。空のディレクトリは由来を問わず除去できる
ので、記録しなくても解ける。

### foreign な symlink が含まれていても除去する

他のツールが管理する symlink を巻き込みうる。

### 除去できる leaf だけを除去し、残りを conflict にする

除去済みの leaf と残った実体が混在する中間の状態が残り、何が起きたかを追いにくくなる。

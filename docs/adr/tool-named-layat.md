---
id: "ADR-6cd587e8-7aa2-4624-be78-7f6501bf0d6f"
type: adr
name: "ツール名を layat にする"
section: concept
status: stable
sources:
  - id: "issue-385"
    resource: "https://github.com/yasunori0418/layat/issues/385"
    title: "#385"
  - id: "issue-386"
    resource: "https://github.com/yasunori0418/layat/issues/386"
    title: "#386"
origin: "Issue #385（epic）の grilling セッション（2026-09-04）で確定。ADR-0054 §1 を書き直した"
justifies:
  - "REQ-2aa3abbc-90b2-486e-92de-d785554bdeb3"
  - "REQ-5dd5a4e9-6162-4fa5-b295-66844f5a4f3b"
  - "REQ-c50df875-2cb0-4e72-8a21-858359a11cae"
  - "REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10"
  - "REQ-8085f194-c903-4ecb-abd8-c719fe7b3292"
  - "DSG-16373ec2-3496-4b12-b3b1-ef74e0435b58"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-07T02:02:56+09:00"
---
# ツール名を layat にする

- ステータス: stable
- 日付: 2026-11-03
- 関連: GitHub Issue #385, GitHub Issue #386, [layat](../ubiquitous-language/layat.md)
- 起点: Issue #385（epic）の grilling セッション（2026-09-04）で確定。ADR-0054 §1 を書き直した

## 背景

このツールの旧名は nput だった。「n」は nix を指し、nix store でハッシュを固定したファイルを
配置することを主な用途と見込んでいた。

実装が進むと、この前提が 2 点で崩れた。engine が受け取る契約は manifest 1 本で、それを誰が
生成したかに engine は関知しない。nix の `lib/` は生成系の 1 つでしかない。また、作者自身の
dotfiles は `mkOutOfStoreSymlink` による可変の symlink で配置しており、store へのコピーを
介していない。

ツールが解決しているのは「フェッチ済みの内容を、manifest のとおりに root 相対の target へ置く」
ことで、内容を誰がどうフェッチしたかではない。nix に縛られず、この役割を表す名前が要った。

## 決定

ツールの名前を **layat** とする。"**lay** \<src\> **at** \<target\>" を縮めた造語で、CLI 名・
Nix の名前空間・on-disk 名・エラーコードの接頭辞など、ツール名を含む字句はすべてこの名前で書く。

字句の具体は、状態ディレクトリを [profileDir は config 専用ディレクトリとし、profile リンク・世代・pending out-link をその中に並べる](../requirements/profile-on-disk-layout.md)、退避の既定 suffix を
[apply --backup は配置を塞ぐ記録外実体を rename 退避してから配置する](../requirements/apply-backup-rename-aside.md)、flake-parts のモジュールを [flake-parts での使い方は直書きと同一の derivation を生み CLI のアドレッシングを変えない](../requirements/flake-parts-module-path.md)、
モジュールの名前空間を [全モジュールは共通オプションの同一集合を公開し、entries は configs 経由・root はモジュールが pin する](../requirements/module-common-options.md) と
[home-manager モジュールの engine kick は activation から全 config のビルド済み link-farm を 1 回の起動で渡し、失敗で switch を止める](../requirements/hm-activation-contract.md)、flake outputs を [layat 自身の flake outputs は packages / templates / 各モジュール / flakeModules / lib の 5 系統で構成する](../design/flake-outputs-attrs.md) に置く。

## 根拠

名前は次の 4 条件で選んだ。

1. 日本語由来の語を避ける
2. 既存の英単語に限らず、造語やアルファベット表記の他言語も認める
3. UNIX 哲学に沿った短いコマンド名にする（3〜6 字を目安）
4. 既存のプロダクトと被らない。必ず調査で確かめる

layat は 5 字で、"lay-at" と発音でき、`layat apply skills` や `layat reset` と自然に読める。
`chroot`（change root）や `getopt`（get options）と同じく、動詞句を縮める UNIX コマンドの
伝統的な造語法に従い、manifest のとおりに src を target へ置くという動作を一語で表す。

条件 4 は GitHub Search API・npm・crates.io・PyPI の 4 つで確かめ、2 回の調査ともすべて空きだった。
照会の手順と記録は Issue #386 が持つ。

## 棄却した代替案

### 改名せず「n」の意味を読み替える

変更のコストは無いが、名前が何も表さない状態が残る。ツールの役割を表す名前にするという動機を
満たさない。

### 既存の英単語を使う

stow・mise・emplace・berth・belay など、配置の語義に合う語はどれも既存のプロダクトが使っていた。
建築や木工の比喩（lath・dowel・tenon など）や、situ・lieu などの短い語も、既存のソフトウェアか
大量の検索ノイズとぶつかる。

### 日本語のローマ字にする

`haizen`（配膳）と `suetsuke`（据付）は全レジストリで空いていたが、条件 1 に反する。

### 他の造語を使う

`layt` は "late" と同音で口頭で伝えにくい。`layn`・`verlay`・`theto` は同名のプロダクトがあり、
`layto` は Professor Layton と 1 字違いで検索しにくい。`laymap`・`layset` は空いていたが、
動作を表す力で layat に劣る。`laydo` と `setz` は次点として残したが、layat が空いていたので
使わなかった。

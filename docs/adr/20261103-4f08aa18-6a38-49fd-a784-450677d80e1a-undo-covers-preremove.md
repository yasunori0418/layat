---
id: "ADR-4f08aa18-6a38-49fd-a784-450677d80e1a"
type: adr
name: "配置前除去の除去も undo の対象にする"
status: stable
sources:
  - id: "issue-168"
    resource: "https://github.com/yasunori0418/layat/issues/168"
    title: "#168"
  - id: "issue-170"
    resource: "https://github.com/yasunori0418/layat/issues/170"
    title: "#170"
  - id: "issue-172"
    resource: "https://github.com/yasunori0418/layat/issues/172"
    title: "#172"
origin: "旧 epic #167 から統合された epic #172 の grilling セッション（2026-07-11・2026-07-12）で確定。ADR-0044 §5 と ADR-0046 §5（途中失敗の扱い）を書き直した"
justifies:
  - "REQ-5e75aabc-0e8f-4a6c-92bd-a712dc68a940"
  - "TP-deb05610-44bc-4962-8939-952392e5fbd0"
references:
  - "ADR-56325270-20bb-4959-8498-e3223c35e24b"
  - "ADR-74423bb8-1453-4689-a25e-30bf46536afc"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-05T00:35:05+09:00"
---
# ADR-4f08aa18-6a38-49fd-a784-450677d80e1a: 配置前除去の除去も undo の対象にする

- ステータス: stable
- 日付: 2026-11-03
- 関連: ADR-56325270-20bb-4959-8498-e3223c35e24b, ADR-74423bb8-1453-4689-a25e-30bf46536afc, GitHub Issue #168, #170, #172
- 起点: 旧 epic #167 から統合された epic #172 の grilling セッション（2026-07-11・2026-07-12）で確定。ADR-0044 §5 と ADR-0046 §5（途中失敗の扱い）を書き直した

## 背景

配置を塞ぐ、自己記録の古くなったもの（祖先の symlink、実ディレクトリの target の配下、method を変えた
symlink）は、配置の前に除去する。除去は symlink の unlink と空ディレクトリの rmdir で、子から
親の順に並ぶ。除去の後で配置が途中失敗すると、親を除去したのに子が置かれていない中間の
状態が残る。

## 決定

配置前除去の unlink と rmdir も undo ジャーナル（ADR-56325270-20bb-4959-8498-e3223c35e24b）に
記録し、apply と rollback が途中失敗したら他の FS 変更と同じく巻き戻す。記録は配置前除去を
実行した順のままにする。

逆操作（記録 dest での symlink の再作成、空ディレクトリの再作成）は
REQ-5e75aabc-0e8f-4a6c-92bd-a712dc68a940 に、故障を誘発した検証は
TP-deb05610-44bc-4962-8939-952392e5fbd0 に置く。

## 根拠

配置前除去も apply の中で行う FS 変更で、ジャーナルから外す理由が無い。外すと、配置前除去を
経る移行（dotfiles でよく起きるレイアウト変更）だけ、巻き戻しが粗いという非対称が生まれる。

配置前除去は子から親の順に並ぶので、ジャーナルを逆にたどれば親の再作成から子の再作成へと
正しい順で戻る。巻き戻し側で順序を並べ替える必要が無い。

## 棄却した代替案

### 配置前除去は巻き戻さず、冪等な再実行の収束に任せる

再実行すれば最終的な状態には収束するが、それまでは中間の状態が残る。ジャーナルがある以上、
配置前除去だけ保証を粗いままにする理由が無い。

---
id: "ADR-ad1c2507-1f42-4cd3-ab56-9183892f8988"
type: adr
name: "profile は解決済み root の roothash でキーする"
status: 採用
origin: "実装前の残セマンティクス確定（2026-06-11〜06-14）。ADR-0005 §「profile は解決済み root でキーする」・ADR-0013 §7・ADR-0023 §3・ADR-0024 §1 を書き直した"
justifies:
  - "REQ-d5a2e289-40bc-45a9-9d44-21b8dc561b81"
  - "REQ-46fccb80-4bae-4d37-bc19-dded88e9a9c0"
  - "REQ-d41b1d0a-c6d5-41cc-93f9-e5cc7f152da4"
references:
  - "ADR-0017"
  - "ADR-0034"
  - "ADR-0036"
---
# ADR-ad1c2507-1f42-4cd3-ab56-9183892f8988: profile は解決済み root の roothash でキーする

- ステータス: 採用
- 日付: 2026-11-01
- 関連: ADR-0017, ADR-0034, ADR-0036
- 起点: 実装前の残セマンティクス確定（2026-06-11〜06-14）。ADR-0005 §「profile は解決済み root でキーする」・ADR-0013 §7・ADR-0023 §3・ADR-0024 §1 を書き直した

## 背景

layat は config ごとに profile（世代の系列）を持ち、前世代の manifest との差分で、不要になった配置を除去する。
profile を config 名だけでキーすると、同じ config を別の root へ配置したときに 1 つの系列を共有する。
同一 entrypoint の複数クローン、`--root` で配置先を差し替えた適用、`root` に絶対パスを渡した fixed root が
これに当たる。

系列を共有すると、片方の適用が他方の配置を不要とみなして消すか、前回の root に置いた配置物を
どの系列も追跡しない孤児として残す。

## 決定

profileDir を config 名で直接キーするのは、root が 1 つに定まる home と system（いずれも `--root` なし）だけに
する。それ以外（project mode・fixed root・`--root` 明示時の全モード）は、解決済みの絶対 root パスから求めた
`<roothash>` と config 名でキーする。`<roothash>` は解決済み root パスの sha256 を短縮した hex とし、
`<roothash>` 階層には元の root パスを記録した backref ファイルを置く。

キーの対応表は REQ-d5a2e289-40bc-45a9-9d44-21b8dc561b81、`<roothash>` の定義は REQ-46fccb80-4bae-4d37-bc19-dded88e9a9c0、
backref による孤児の逆引きは REQ-d41b1d0a-c6d5-41cc-93f9-e5cc7f152da4 に置く。

## 根拠

root ごとに系列を分ければ、クローン間や root 間で 不要な配置の除去が互いの配置を壊すことも、前回の root の
配置物が気付かれずに残ることも構造的に起きない。home と system は root が 1 つに決まるので、直キーでも
衝突せず、「1 ユーザー 1 profile」の分かりやすさを保てる。

hash を固定長の hex にすれば、深いパスでもディレクトリ名がファイルシステムの長さ制限に触れない。hash
からは元の root が分からないので backref を置き、クローンの削除後に残った孤児系列を人と `layat prune` が
逆引きできるようにする。

`--root` は全モードで同じ規則に従うので、`--root` を付けて積んだ世代を操作するときも同じ `--root` が要る。
`--root $HOME` のように既定と同じ値を明示すると通常の home とは別系列になるが、`--root` は検証や特殊な
配置のための脱出路なので許容する。

## 棄却した代替案

### config 名だけでキーする

最も単純だが、同じ config を別の root に配置すると系列を共有し、不要な配置の除去が他のクローンの配置を壊すか、
前回の root の配置物を孤児として残す。

### fixed root と `--root` の上書きは文書の注意書きに留める

実装は最小で済むが、気付かれない孤児が残り、誤って壊さない・問題を隠さないという方針に劣る。

### fixed root を eval 時エラーにする

キーの問題は消えるが、任意の固定 root を使う余地を失い、root を一般化した方針に逆行する。

### roothash を短縮 hash だけにする

backref を置かないと、孤児系列がどの root 由来か分からず、手動でも `layat prune` でも削除の判断ができない。

### root パスを可逆にエンコードしてディレクトリ名にする

逆引きはできるが、深いパスではディレクトリ名がファイルシステムの長さ制限に触れうる。

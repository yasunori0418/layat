---
id: "RISK-ac5cd9b4-58f9-43d8-9fae-66f7b5940beb"
type: risk
name: "profile のキーイングか世代の読み取りが誤り、別 config の世代系列を取り違える"
status: stable
threatens:
  - "REQ-d5a2e289-40bc-45a9-9d44-21b8dc561b81"
  - "REQ-2aa3abbc-90b2-486e-92de-d785554bdeb3"
  - "REQ-1be4d678-959c-44d7-a346-44bfd95af56e"
  - "REQ-9c111c32-8f6c-4eda-859c-dae02c0624fc"
  - "REQ-833205ca-adfc-4c86-bf66-ae4d8fe4885b"
  - "REQ-2f9205ee-cec5-4072-ac3e-890caae79904"
  - "REQ-706de717-4e47-471a-a1c0-448635be159c"
  - "REQ-840b3641-6e76-46da-82e9-680cabd65abe"
likelihood: medium
impact: high
level: high
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T22:43:20+09:00"
---
# RISK-ac5cd9b4-58f9-43d8-9fae-66f7b5940beb: profile のキーイングか世代の読み取りが誤り、別 config の世代系列を取り違える

## リスク

`profileDir` の解決規則か、そこに並ぶ世代リンクの読み取りが誤ると、layat は「前世代」に
別 config の manifest を読む。前世代 manifest は古くなった配置の除去の入力そのものなので、
取り違えればすぐに「別 config が置いた symlink を、この config の古くなった配置として消す」という
誤削除に直結する。

キーイング側の破れ方は 2 方向ある。**分離すべきものが同じ系列に含まれる** — fixed root や `--root`
上書きが `<roothash>` でキーされず `<name>` 直キーになると、別 root の同名 config が
同じ世代系列を共有する（気付かれない孤児）。project mode で同一 entrypoint を複数箇所へ
クローンした場合も同様に衝突する。**分離すべきでないものが割れる** — 同じ root を指す
はずの呼び出しが毎回違う `<roothash>` を得ると（hash が決定的でない・長さが揺れる）、
apply のたびに新しい世代系列が生まれ、前世代が永久に見つからず古くなった配置の除去が機能しなくなる。

レイアウト側では、profile リンク・`profile-N-link` 世代・`.pending` out-link を同一
ディレクトリの兄弟として並べる規約（→ [profileDir は config 専用ディレクトリとし、profile リンク・世代・pending out-link をその中に並べる](../requirements/profile-on-disk-layout.md)）が崩れると、read-only な store パスを
貫通して書こうとし、書き込みに失敗する。`<state>` の解決（`$XDG_STATE_HOME` 優先・無ければ
`~/.local/state`・`$HOME` も解決できなければエラー）が誤れば、profile 一式が想定外の
場所に作られ、既存の世代が一切見えなくなる。

世代リンクの走査も同じ系統の脅威に属する。兄弟名から世代番号を解釈し損ねる
（`profile.lock` のような世代でない名前を世代に数える・番号順に並べない）か、`profile` の
指す世代を current と判定し損ねれば、rollback の戻り先も `list-generations` の表示も狂う。
profileDir を読めない失敗を黙認して「世代 0 件」と見なすのは、前世代が無いという最も
危険な結論を沈黙で作る。

キーイングの前提と世代を操作する処理も同じ系統にある。build 前に profileDir を決めるための
`passthru.rootKind`（→ [mkManifest の返り値は passthru で `rootKind` を露出する](../requirements/mkmanifest-passthru-rootkind.md)）が評価時に取れないか実際の root と
食い違うと、flock と前世代の読み取りが別の profileDir に対して行われる。build しないコマンドが eval 先行を
省くか、`--root` で作った世代を同じ `--root` 無しで操作できてしまうと、別の系列を読む
（→ [非 build コマンドも eval 先行を共通前段に持つ](../requirements/non-build-commands-eval-first.md)）。世代の commit・列挙・切替が世代バックエンドの契約
（→ [engine は世代操作を世代バックエンドの契約（Commit / List / Switch / UserBase / SystemBase）越しに行う](../requirements/generation-backend-contract.md)）を迂回するか、世代の間引きが契約と nix コマンドの
外で行われると（→ [世代操作は世代バックエンドの契約越しに行い、GC root の間引きと store 解放は nix コマンドへ分けて委譲する](../requirements/generation-gc.md)）、同じ系列を処理ごとに違う形で読み書き
する。失敗時に残る `.pending` が同名で上書きされず config あたり 1 個を超えて溜まるのも、
同じディレクトリ規約の崩れである（→ [失敗時に残る pending gcroot は config あたり最大 1 個に有界とし回収処理を持たない](../requirements/pending-gcroot-bounded.md)）。

## 実現性

**likelihood: medium** — キーイング規則はモード（home / project / fixed）と `--root` の
有無の掛け合わせで枝が多く、モード追加のたびに考慮し忘れが入りうる。

**impact: high** — 起きるのは前世代 manifest の取り違えであり、ユーザーの実配置の誤削除に
つながる。消される symlink 自体は別 config の再 apply で張り直せるが、取り違えの原因である
世代系列の破損はそれでは直らない。系列が割れれば（`<roothash>` が呼び出しごとに揺れる側）
前世代が永久に見つからず rollback の戻り先も古くなった配置の除去の入力も失われ、系列に他の config が含まれれば
（`<name>` 直キーになる側）別 root の config どうしが互いの世代を上書きし続ける。
どちらも profile 一式を捨てて配置をやり直すまで回復せず、その時点で世代の履歴は戻らない。

## 緩和

[state 基底・profileDir のキーイング・世代リンク名が root と mode ごとに正しく決まる](../test/generations/profile-path-keying.md)（`<state>` 解決・profileDir キーイング・世代リンク命名）と
[profileDir の世代リンクを番号順に列挙して current を判定し、読めない profileDir は黙認しない](../test/generations/generation-listing-parse.md)（世代リンクの走査）が緩和する。

キーイングの前提と世代を操作する処理の面のうち、`--root` 上書きのキーイングは
[state 基底・profileDir のキーイング・世代リンク名が root と mode ごとに正しく決まる](../test/generations/profile-path-keying.md) が、世代の列挙は [profileDir の世代リンクを番号順に列挙して current を判定し、読めない profileDir は黙認しない](../test/generations/generation-listing-parse.md) が覆う。
`passthru.rootKind` の値、build しないコマンドの eval 先行、commit・切替・間引きが契約を迂回しないこと、
`.pending` が config あたり 1 個に収まることを確かめる検証条件はこの risk に無く、残余として残る。

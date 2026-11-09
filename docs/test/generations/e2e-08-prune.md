---
id: "CASE-b19188cb-e3f0-4c52-8c92-6db96bbbb45a"
type: test_case
name: "e2e 08-prune — 実 apply が作った系列を root 削除で孤児にし、prune が系列ごと消して gcroot が外れる"
status: stable
resource: "tests/e2e/scenarios/08-prune.sh"
covers:
  - "TC-bc653f83-70ce-4920-b7cd-9b70a4cd2cae"
  - "TC-a9857bf7-f7f9-41f9-b42c-9993fd16a5e9"
follows:
  - "TP-229b69c0-cf5e-4fb6-a353-27e5064d93e9"
  - "TP-d3000054-42d9-4bac-912a-dd3abc38d3e9"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:27:49+09:00"
---
# e2e 08-prune — 実 apply が作った系列を root 削除で孤児にし、prune が系列ごと消して gcroot が外れる

## 対象

`tests/e2e/scenarios/08-prune.sh`

`e2e_isolate` で仮 `$HOME` と `$XDG_STATE_HOME` を切り、projectRoot の fixture flake を持つ
git repo を 2 つ用意して実 nix 経由で apply する E2E シナリオ。root（git toplevel）ごとに独立した
`<roothash>` 系列が立つので、一方の root を `rm -rf` して孤児にし、もう一方は生存させたまま
`layat prune` を駆動する。

判定分岐そのものは [prune_test.go — 孤児系列の判定と系列ごとの削除・dryrun / 確認 / lock / 削除段の失敗](prune-test-go.md)（tmpdir に手で組んだ fixture）が
覆っている。**このシナリオの持ち分は「実 apply が置いた本物のレイアウトを prune が対象にできるか」**で、
手組み fixture が再現しないのは `nix-env --set` が `/nix/var/nix/gcroots/auto` 経由で張る
indirect gcroot と、CLI の実際の処理（`--dryrun` / `--yes` / `--json`）を通ることの 2 点。engine の
ユニットだけでは、実機の gcroot が外れないまま系列だけ消える退行に気付けない。

`.pending` は逆に**ユニット側だけが持つ形**で、このシナリオには現れない。apply が成功すると
`cleanupPending` が `--set` 直後に消し、以後は世代リンクが gcroot を継承する
（→ `internal/engine/engine.go` 手順 10・[store path は out-link で受け取る](../../adr/store-path-via-out-link.md)）。`.pending` だけの系列を prune が対象にすることは
[prune_test.go — 孤児系列の判定と系列ごとの削除・dryrun / 確認 / lock / 削除段の失敗](prune-test-go.md) の `TestPruneRemovesSeriesWithoutGenerations` が覆う。

走査基底は 2 つある（→ [prune は system 基底も走査し、権限の無い系列を見送る](../../adr/prune-scans-system-base-skips-unpermitted.md)）。system 基底は絶対パスで `$HOME` / `$XDG_STATE_HOME` の
差し替えでは動かないため、`e2e_isolate` が `LAYAT_SYSTEM_PROFILE_BASE` で隔離先の tmpdir へ
向ける（→ `cmd/layat/prune.go`）。これにより破壊的な `--yes` の実行でも実機の
`/nix/var/nix/profiles/layat` には触れない。

## 検証内容

- **system 基底の隔離** — 隔離先の system 基底へ手で置いた孤児系列（`.root` だけの系列）が
  走査に載り、`--yes` で消えること。破壊的な `--yes` を回す前提条件そのものなので、シナリオ
  自身がアサートする。**env の値を見比べる形では足りない** — それはハーネスが export した
  ことしか確かめられず、`layat`（`run.sh` の `LAYAT` で差し替え可能）がその env を実際に読むかは
  観測できないため、seam を持たないバイナリを指した実行がテストを通ったまま実基底を走査する
- **apply が作るレイアウト** — project mode の apply が `<roothash>/.root`・`<roothash>/docs/profile`
  （symlink）・`<roothash>/docs/profile-1-link`（symlink）を置く。系列ディレクトリは roothash の
  計算式をシナリオへ写さず、置かれた `.root` の内容から読み取る（実装との二重管理を避ける）
- **prune 前の gcroot** — `nix-store --gc --print-roots` に孤児予定系列の世代リンクが現れる。
  `<state>` 配下の世代リンクは `/nix/var/nix/gcroots/auto` 経由の indirect root で、系列を消すまで
  store path を保持し続ける（GC 回収の陽性対照）。`print-roots` の取得失敗と空出力は
  fail として扱い、後段の陰性判定がコマンド失敗と区別できない形にならないようにする
- **`prune --dryrun` の非破壊性** — 孤児系列を `remove-series` 行に root パス付きで挙げ、生存
  系列は挙げない。行はフィールド構造ごと比較する。実行後も `.root` と世代リンクが残る
  （→ [prune は削除前に root パスの一覧を示して確認を取る](../../adr/prune-confirms-with-root-list.md)）
- **`prune --dryrun --json`** — payload の形は [prune --json はトップレベル info の removed に削除系列を、skipped に判定か削除に至れず残した系列を載せる](../../requirements/prune-json-payload.md) が定める。同じ在庫をエンベロープの `info.removed` が含む。prune は config を
  名指ししないので `results` は空配列のまま。孤児系列の `root` と `names`（`["docs"]`）が載り、
  生存系列は載らない。`info.skipped` が空であることも併せて見る（想定外の skip が `removed` の
  陽性に紛れて通らないように）。走査基底は 2 つとも隔離下にあり素性の分かる系列しか置いていない
  ので、基底で絞らず全件 0 で固定する。dryrun の `removed` は候補そのもので skip 済みを含まない
  ため、載っている側の確認ではこの取りこぼしを代替できない（→ issue #134）
- **`prune --yes` の削除** — 孤児系列は `<roothash>` ごと消え、生存系列の `.root` と profile
  symlink は残る
- **配置物の不変** — 生存 root 配下の配置 symlink と内容は prune の前後で変わらない
- **GC 回収可能になったこと** — `print-roots` から孤児系列配下の gcroot が消え、生存系列の
  gcroot は保持される。実 GC は回さず、root として数えられなくなったことで見る
- **再実行の no-op** — 孤児が無い状態での再実行が**削除 0 件・skip 0 件**を報告し、生存系列の
  `.root` と `<name>` 側の profile symlink の両方が残る。件数まで固定するのは、skip 側になった
  系列も `Removed` が空で `no-op` を出すため、「正しく生存と判定して残した」と「判定できずに
  残した」が区別できなくなるから

破壊的実行のインベントリ（`--yes` の結果を `--json` で見る形）と削除段の失敗分岐は、権限操作を
要して実機では割に合わないので [prune_test.go — prune の確認ポリシー・root パス一覧の提示・ストリーム排他と --json の形状](../cli-json/prune-cli-test-go.md)（`cmd/layat/prune_test.go`）と [prune_test.go — 孤児系列の判定と系列ごとの削除・dryrun / 確認 / lock / 削除段の失敗](prune-test-go.md) の
担当。このシナリオは持たない。

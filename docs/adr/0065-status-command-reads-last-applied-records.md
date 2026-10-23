---
id: "ADR-0065"
type: adr
name: "status は last-applied 記録を走査して表示する"
status: 採用
issues:
  - "#512"
  - "#198"
origin: "Issue #198（`layat status [name]`）の実装前に、2026-07-19 の grilling 以後の実装の変化を差分 grilling で反映した（2026-10-23）"
justifies:
  - "REQ-e8d911ff-8ee2-4d68-87d0-26f6d6817173"
  - "REQ-fa181aa6-29a2-48c3-ae07-cc1b9a3b0303"
  - "DSG-85439cb5-a149-422e-8596-fc4e860bc3a5"
references:
  - "ADR-0036"
  - "ADR-0043"
  - "ADR-0058"
  - "ADR-0062"
---
# ADR-0065: status は last-applied 記録を走査して表示する

- ステータス: 採用
- 日付: 2026-10-23
- 関連: ADR-0036, ADR-0043, ADR-0058, ADR-0062, GitHub Issue #198, #512
- 起点: Issue #198（`layat status [name]`）の実装前に、2026-07-19 の grilling 以後の実装の変化を差分 grilling で反映した（2026-10-23）

## 背景

「今なにが apply されているか」を確かめる手段が無い。`list-generations` は世代の一覧しか出さず、
`apply --dryrun` は nix eval / build が要るうえ、示すのは新しい manifest との差分である。

Issue #198 は 2026-07-19 の grilling で、profile 基底 `<state>/nix/profiles/layat` を走査して
current 世代の manifest を表示する形に決まっていた。その後、apply は最後に配置した manifest を
last-applied 記録に書くようになり（ADR-0058）、system mode の記録は `/var/lib/layat` に置く
ことになった（ADR-0062）。記録は manifest を埋め込み、`appliedAt` と `linkFarm` を持つので、
store を辿らずに配置内容を読める。

## 決定

### 1. データ源は last-applied 記録にする

- `layat status [name]` は user の記録基底 `<state>/layat` と system の記録基底
  `/var/lib/layat` を、この順に走査し、記録を持つ config を列挙する。profile 基底は読まない。
  世代番号は `list-generations` の担当とする。
- 記録を読む関数は engine の非公開関数なので、engine に読み取り専用の `Status` を足す。記録の
  検査は apply が baseline を読むときと同じものを使い、status 用に緩めない。
- 基底の中では `<name>` 直キーを先に、`<roothash>` 系列を後に、それぞれ名前の昇順で並べる。
  `<roothash>` 系列の列挙は prune と同じ `ListRootHashSeries` を使う。
- 無い基底は空として扱う。system 基底が読めなければ warning を出して飛ばす
  （`list-generations --all` と同じ規則）。

### 2. subject 名と `[name]`

- 直キーは `<mode>:<name>`（mode は基底に応じて `home` / `system`）、`<roothash>` 系列は
  `<mode>:<name>@<root>`（mode は記録の manifest の `rootKind`、root は backref `.root` の値）。
  記録が読めず `rootKind` が取れない系列の mode は、基底に応じた `home` / `system` で代用する。
- `[name]` は `<name>` 部分の完全一致で絞り込み、mode / root を問わない。一致が 0 件なら
  stdout に一覧を出さず、stderr に 1 行通知して exit 0 にする。

### 3. 出力

- テキストは subject ごとに `# <subject> (applied <appliedAt>)` の見出し行と、entry ごとの
  `<target>\t<method>\t<dest>`（dest は planner が計算する symlink の指す先）で構成し、
  subject の間を空行で区切る。
- `--json` は outturn エンベロープで出す（ADR-0043）。`list-generations` と同じく
  `results[i].result.info` に `{mode, root, appliedAt, linkFarm, entries}` を載せ、`items: []`
  とする。`entries[]` は `{target, method, srcKind, src, subpath}`。
- 不正な記録は、その subject を error にして他の subject の列挙を続ける。envelope の status は
  error、exit 1。

### 4. 読み取り専用

status は flock を取らず、ディレクトリもファイルも作らず、記録を書き換えない。euid も検査
しない。記録は mode 0644 で書かれるので、非 root でも system 基底の記録を読める。

## 根拠

- **記録を読む**: 記録は配置した manifest をそのまま持つので、eval / build も store も要らない。
  profile の current は commit の失敗で記録とずれることがあり（ADR-0058）、baseline として
  信頼されているのは記録の側である。
- **2 基底の走査**: system mode の config も apply 済みの配置を持つ。基底ごとに別のコマンドへ
  分けると、全体を一度に見るという目的を満たせない。
- **厳格な検査**: 記録の不正は apply を止める。status が警告なしに飛ばすと、利用者は問題に apply の
  時点で初めて気付く。
- **0 件は exit 0**: 名前を指定して一致しないのは「その名前の記録が無い」という結果で、
  コマンドの失敗ではない。

## 影響

- 要求: status の要求（REQ-e8d911ff-8ee2-4d68-87d0-26f6d6817173）を起こし、読み取り系の
  JSON ペイロード（REQ-fa181aa6-29a2-48c3-ae07-cc1b9a3b0303）に status の info の形を足す。
  `--generator` を無視するコマンド、`--json` で止める行指向の出力、euid を検査しない経路の
  各列挙に status を加える。
- 設計: engine の `Status` と CLI の分担（DSG-85439cb5-a149-422e-8596-fc4e860bc3a5）。

## 棄却した代替案

- **profile 基底の current 世代を読む（Issue #198 の当初案）**: link-farm の store path を辿る
  必要があり、記録とずれた profile を表示しうる。
- **記録の検査を緩めて読めるところまで表示する**: 壊れた記録を見えなくし、apply の baseline と
  status の表示が食い違う。
- **mode / root で絞るフラグを持つ**: 列挙の件数は少なく、`[name]` の完全一致で足りる。

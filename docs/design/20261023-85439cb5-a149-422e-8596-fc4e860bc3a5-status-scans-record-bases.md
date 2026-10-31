---
id: "DSG-85439cb5-a149-422e-8596-fc4e860bc3a5"
type: design
name: "status を engine の記録基底走査と CLI の絞り込み・出力に分ける"
satisfies:
  - "REQ-e8d911ff-8ee2-4d68-87d0-26f6d6817173"
---
# DSG-85439cb5-a149-422e-8596-fc4e860bc3a5: status を engine の記録基底走査と CLI の絞り込み・出力に分ける

## 設計

`layat status` を 3 つに分担する。既存コマンドと同じ層境界（→ DSG-7d354fe0-a333-495b-9f4b-14bba316dc47）に
乗せる。

```
internal/paths/    ← <roothash> 系列の列挙。既存の ListRootHashSeries をそのまま使う
internal/engine/   ← 記録基底の走査と記録の読み取り。Status / StatusOptions / StatusSeries
cmd/layat/         ← [name] の絞り込み、subject 名の組み立て、テキスト / --json の出力
```

**`internal/engine`** が走査と読み取りを持つ。記録を読む `readRecord` は engine の非公開関数で、
apply が baseline を読むときと同じ厳格な検査（壊れた JSON・フィールドの欠落や型違い・未知の
フィールド・`recordVersion` が 1 でも 2 でもない記録を error にする）を持つ。status 用に検査を緩めた読み取りを
別に作らず、engine の中で同じ関数を呼ぶ。

**`internal/paths`** には関数を足さない。`<roothash>` 系列の列挙は prune が使う
`ListRootHashSeries` を記録基底に対して呼ぶ。この関数は「`.root` を持つディレクトリ」を系列と
みなすだけで、基底が profile 基底か記録基底かを問わない。記録基底も `<roothash>` 階層に
backref `.root` を置く同型のレイアウトなので（→ REQ-9d622539-a1e5-4719-af51-e0cd85b66f35）、
そのまま使える。`<name>` 直キーの列挙は status だけが使うので、`internal/engine/status.go` の中に
閉じる。

**`cmd/layat`** は engine の結果を絞り込んで出力するだけで、FS を読まない。

### engine API

```go
type StatusOptions struct {
    StateDir       string                    // <state>（空なら paths.StateDir()）
    SystemStateDir string                    // system の記録基底の親（空なら /var/lib）
    Warnf          func(format string, args ...any) // nil なら stderr
}

type StatusSeries struct {
    Mode       string             // home / project / fixed / system
    Name       string             // <name>。系列全体で 1 件として返す系列では空
    Root       string             // <roothash> 系列の backref .root の値。直キー・backref が読めない系列・系列全体で 1 件として返す系列では空
    RootHash   string             // <roothash> 系列のディレクトリ名。直キーでは空
    RecordFile string             // last-applied.json の絶対パス
    AppliedAt  string
    LinkFarm   string
    Manifest   *manifest.Manifest
    Err        error              // 記録・backref が読めない、または <name> を列挙できない・見つからないとき。他のフィールドより先に見る
}

func Status(opts StatusOptions) ([]StatusSeries, error)
```

走査する基底は `<StateDir>/layat`、`<SystemStateDir>/layat` の順。`SystemStateDir` は
apply が記録を書くときと同じ値で、CLI は `LAYAT_SYSTEM_STATE_DIR`（隔離・テスト用）があれば
その値を渡す。

### 基底内の列挙

基底ごとに次の 2 段で集め、連結する。

1. **`<name>` 直キー**: 基底直下のディレクトリのうち、段 2 の `ListRootHashSeries` が系列として
   返さなかったもので、`last-applied.json` を直下に持つもの。
   Mode は基底に応じて `home`（user 基底）/ `system`（system 基底）、Root は空。
2. **`<roothash>` 系列**: `ListRootHashSeries(base)` が返す系列の各 `<name>` のうち、
   `last-applied.json` を直下に持つもの。Mode は記録の manifest の `rootKind`、Root は系列の
   `Root`（backref `.root` の値）。

どちらの段も、`last-applied.json` を持たない `<name>`（reset が全 entry を撤去して記録を消した
後など）は列挙しない。記録を持つかどうかは `last-applied.json` の存在で決め、存在するのに
読めない・不正な記録は次々節の `Err` に回す。`last-applied.json` の stat が `fs.ErrNotExist`
以外で失敗した `<name>` も、記録の有無が分からないので落とさず、`Err` 付きで返す。

2 つの段は、段 1 が段 2 の系列を除いた残りから拾うことで排他になる。`ListRootHashSeries` は
`.root` の lstat が `fs.ErrNotExist` 以外で失敗したディレクトリも `BackrefErr` 付きの系列として
返すので、`.root` の有無を判定できないディレクトリは段 2 だけが扱う。どちらの段も名前の
昇順に並べる（段 2 は `<roothash>`、その中の `<name>` の順）。並び順は `ListRootHashSeries` が返す
順に依存させず、engine 側で並べる（同関数は順序を契約に含めていない）。

### 基底の不在と読めない基底

`ListRootHashSeries` は基底の `ReadDir` が失敗すると、`fs.ErrNotExist` も含めて error を返す。
そのため engine は基底ごとに先に `os.Stat` で存在を確かめ、無ければその基底を空として次へ
進む。`ListRootHashSeries` には存在する基底だけを渡す。

存在するのに読めない基底は、user 基底なら `Status` の error として返し、system 基底なら
`Warnf` に基底を名指しして 1 行出し、その基底を飛ばす。非 root 実行で system 基底が読めない
状況を、user 側の列挙まで止めずに知らせるためである（`list-generations --all` と同じ規則）。

### 系列単位の error

記録が読めない・不正、または `ListRootHashSeries` が `BackrefErr` / `NamesErr` を載せた系列は、
`Err` を立てた `StatusSeries` として他の系列と同じ並びに置き、列挙を続ける。`Err` を持つ
系列は記録の manifest から `rootKind` を取れないので、Mode には基底に応じた `home` / `system`
を入れる。

- **`BackrefErr`**: root が決まらないので Root は空のまま、記録を持つ `<name>` ごとに `Err` を
  立てて返す。RootHash は通常どおり入れる。記録を持つ `<name>` が 1 つも見つからない系列は、
  `NamesErr` と同じく Name と Root が空の 1 件として返す。検索権限の無い `<name>` 直キーの
  ディレクトリも `.root` の lstat が失敗してここに来るので、落とすと記録が黙って消える
- **`NamesErr`**: `<name>` が分からないので、系列全体を Name と Root が空の 1 件として返す。
  RootHash と `Err` を入れる。`BackrefErr` と同時に立ったときもこの 1 件にまとめる

直キーと `<roothash>` 系列は Root ではなく RootHash の有無で見分ける。`BackrefErr` の系列は
Root が空なので、Root で見分けると直キーと取り違える。

### 読み取り専用

`Status` は `os.Stat` / `os.ReadDir` / ファイルの読み取りだけを使う。apply・reset が使う
`recordFor`・profileDir の作成・flock の取得・`loadBaseline` を呼ばない。これらは
ディレクトリ・`.pending`・`.root` を作るか、profile 基底を読むためである。

### CLI 側

- `[name]` は `StatusSeries.Name` との完全一致で絞る。ただし Name が空の 1 件（系列全体で 1 件として
  返す系列）は、求めた名前が入っている可能性があるので落とさない。0 件なら stdout に一覧を出さず、
  stderr に 1 行出して exit 0。
- subject 名は次の表で決める。行は上から順に評価し、最初に当てはまった行を採る。

  | 系列 | 条件 | subject 名 |
  |---|---|---|
  | 直キー | RootHash が空 | `<Mode>:<Name>` |
  | 系列全体で 1 件として返す | Name が空 | `<Mode>:@<RootHash>` |
  | backref が読めない | Root が空 | `<Mode>:<Name>@<RootHash>` |
  | `<roothash>` 系列 | 上記以外 | `<Mode>:<Name>@<Root>` |

  `[name]` を指定して一致する config が無くても、系列全体で 1 件として返す系列の error subject が
  あれば、それを出力して exit 1 にする（一致 0 件の扱いにはしない）。

- テキストは subject ごとに見出し行と、manifest の entry ごとに `<target>\t<method>\t<dest>`
  （dest は `planner.LinkDest`）を出す。`Err` を持つ subject は stderr に error を出す。
- `--json` は `results[i].result.info` に `{mode, root, appliedAt, linkFarm, entries}` を載せ
  （→ REQ-fa181aa6-29a2-48c3-ae07-cc1b9a3b0303）、`Err` は該当 subject の `errors[]` に置く。

### この設計が引き受ける代償

`ListRootHashSeries` を profile 基底と記録基底の両方で共有するので、その列挙条件が変わると
prune と status の両方の列挙結果が変わる。status 側の取りこぼしは RISK-e92487a5-fff9-40db-844d-6590f1f36dec が
扱う。

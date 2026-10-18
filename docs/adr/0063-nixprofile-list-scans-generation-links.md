---
id: "ADR-0063"
type: adr
name: "nixprofile の List は nix-env を経由せず世代リンクを走査する"
status: 採用
issues:
  - "#129"
origin: "system mode の e2e で、root 所有の profileDir を非 root が `list-generations` すると nix-env の lock 取得で失敗することが分かった（2026-10-18）"
justifies:
  - "REQ-833205ca-adfc-4c86-bf66-ae4d8fe4885b"
revises:
  - "ADR-0057"
references:
  - "ADR-0062"
---
# ADR-0063: nixprofile の List は nix-env を経由せず世代リンクを走査する

- ステータス: 採用
- 日付: 2026-10-18
- 関連: ADR-0062, GitHub Issue #129
- 改訂対象: ADR-0057 §2 の「nixprofile は nix-env 3 コマンドと世代一覧のパースを所有する」のうち、List を `nix-env --list-generations` のパースから profileDir の世代リンクの読み取り専用の走査へ改める（Commit / Switch の nix-env・契約の 5 操作・Fake の位置づけは不変）
- 起点: system mode の e2e で、root 所有の profileDir を非 root が `list-generations` すると nix-env の lock 取得で失敗することが分かった（2026-10-18）

## 背景

ADR-0057 §2 は nixprofile の List を `nix-env --profile <p> --list-generations` の出力のパースで
実装した。nix-env は世代を列挙するだけでも profileDir に `profile.lock` を作って lock を取る。
書けない profileDir では列挙が失敗するので、root 所有の system mode の profile を非 root が
`list-generations` すると exit 1 になる。`list-generations` は読み取り専用のコマンドで、
system mode でも非 root に世代を見せる前提（ADR-0062）と食い違う。

一方、テストダブルの Fake は同じ List を `profile-<N>-link` の兄弟の走査で実装しており、
本番と Fake で列挙のロジックが別だった。

## 決定

### 1. List は世代リンクを読み取り専用で走査する

- nixprofile の List は nix-env を起動せず、`internal/generation` の `ListLinks` を返す。
- `ListLinks` は `<profileDir>/<profile>-<N>-link` の兄弟を世代番号の昇順に列挙する。日時は各リンク
  自身の lstat の mtime を `2006-01-02 15:04:05` でローカル時刻に整形する。current は `profile` の
  readlink が指す世代とする。
- profileDir が無ければ世代 0 件で error にしない。それ以外の ReadDir の失敗は error を返す。
  世代の名前の形をしていない兄弟（`profile.lock` 等）は無視する。
- Fake の List も同じ `ListLinks` に委譲し、本番と Fake の列挙ロジックを 1 つにする。
- Commit / Switch は従来どおり `nix-env --set` / `--switch-generation` で行う。

### 2. current の判定の補助規則は Fake 由来

`profile` の readlink が世代の名前でないときに EvalSymlinks の一致で current を探す規則は、Fake から
引き継いだもので nix-env には無い。`nix-env --set` が作る profile は readlink が常に世代の名前なので、
実運用ではこの分岐に入らない。

## 根拠

- **読み取り専用の走査**: nix 本体の世代列挙も `profile-<N>-link` の走査と lstat の mtime の整形で、
  出力は同じになる。lock を取らない分、書けない profileDir でも読める。
- **Fake と本番の一致**: 列挙ロジックを共用すれば、engine / cmd のテストが Fake で通した列挙が
  本番でも同じ結果になる。
- **ADR-0057 §3 との整合**: profileDir のレイアウトは契約の固定部分で、engine の読みはすでに
  レイアウトを直接読んでいる。List の走査はその読みをバックエンドの側にも認めるだけで、契約の形は
  変えない。

## 影響

- **`internal/generation`**: 走査を `ListLinks` と日時の書式 `DateLayout` として公開し、Fake と
  nixprofile で共用する。nixprofile の出力パースは無くなる。
- 要求: 世代バックエンドの契約（REQ-833205ca-adfc-4c86-bf66-ae4d8fe4885b）の List の実装方法を
  走査に改め、世代操作と GC（REQ-706de717-4e47-471a-a1c0-448635be159c）の nix-env の列挙から
  `--list-generations` を外す。
- **ADR-0057**: §2 に改訂注記を足す。

## 棄却した代替案

- **lock の取得に失敗したときだけ走査へフォールバックする**: 列挙のロジックが 2 つ残り、どちらが
  動いたかで結果が変わりうる。失敗の判別も nix-env のエラー文言に頼ることになる。
- **nix-env を維持する**: 非 root による system mode の `list-generations` が失敗したまま残る。

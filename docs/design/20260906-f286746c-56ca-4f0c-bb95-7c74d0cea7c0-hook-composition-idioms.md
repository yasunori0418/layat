---
id: "DSG-f286746c-56ca-4f0c-bb95-7c74d0cea7c0"
type: design
name: "hook 機構を engine に持たず、配線層ごとの合成 idiom で hook 需要を満たす"
satisfies:
  - "REQ-c1b3ca5f-d2f7-443c-bc4b-b18413ca97b9"
  - "REQ-6c4e174a-4d16-477a-96ff-17cb4eb5b564"
---
# DSG-f286746c-56ca-4f0c-bb95-7c74d0cea7c0: hook 機構を engine に持たず、配線層ごとの合成 idiom で hook 需要を満たす

## 設計

「配置の前後で任意の処理を走らせたい」（配置後のサービスリロード・通知・後処理）需要に対し、
**engine はイベント hook 機構を提供しない**（→ ADR-0041）。代わりに、需要を満たす手段を
**配線層ごとの合成 idiom** としてここに示す。

engine 側に置かない理由は 2 つの要求に対応する。

- **REQ-c1b3ca5f-d2f7-443c-bc4b-b18413ca97b9（モジュールと devShell は engine をキックする
  だけの配線）**: hook ポイントは既に配線層が持っている（devShell の `shellHook`・
  home-manager の activation DAG・standalone の shell 合成）。engine 内 hook はこれらと重複し、
  しかも層をバイパスして engine に配線責務を持ち込む
- **REQ-6c4e174a-4d16-477a-96ff-17cb4eb5b564（engine が叩く外部コマンドは nix と git のみ）**:
  任意スクリプトの実行は、サブプロセスの失敗・出力・タイムアウトという新しい失敗モード群を
  engine の意味論へ持ち込む。外部コマンドを 2 つに閉じる制約と両立しない

後処理コマンドの中身（`systemctl` / 通知コマンド等）は OS の機構であり nput の関心外
（→ REQ-c1b3ca5f-d2f7-443c-bc4b-b18413ca97b9）。以下の例は後処理を `./post.sh` と置く。

### 1. devShell — `shellHook` で前後に並べる

`shellHook` は shell script なので、`nput apply` の前後に任意のコマンドをそのまま並べられる
（devShell が engine を `shellHook` からキックすること自体は
REQ-a0bdf6db-6c0c-476c-916a-61ee4e4510d9）。

```nix
devShells.default = pkgs.mkShell {
  shellHook = ''
    nput apply skills --no-wait && ./post.sh
  '';
};
```

- **名指し apply か `--all --project-root` を使う**。素の `--all` は home mode config も
  `$HOME` へ配置する（→ REQ-d95b814f-aa7a-470e-9320-c14f9c14da7b）
- `--no-wait` は lock 競合時に待たずスキップする shellHook 向けのフラグ
  （→ REQ-1c1526b1-59e3-4264-bb7c-65a10a4aa461）。**スキップは exit 0** なので、上の `&&` は
  「配置をスキップした」ときにも `./post.sh` を走らせる。配置が実際に変化したときだけ
  後処理したいなら idiom 4 を使う

### 2. home-manager — activation DAG で `nput` の後に並べる

home-manager モジュールは `home.activation.nput` から engine をキックする
（`lib.hm.dag.entryAfter [ "writeBoundary" ]` → REQ-8085f194-c903-4ecb-abd8-c719fe7b3292）。
後処理は同じ DAG に `nput` の後続として置く。

```nix
home.activation.postNput = lib.hm.dag.entryAfter [ "nput" ] ''
  run ./post.sh
'';
```

- 依存先は `"nput"` — nput モジュール自身が使う activation 名。ここを `"writeBoundary"` に
  すると nput との順序が決まらない
- `run` で包むのは home-manager の activation ヘルパで、`home-manager switch --dry-run` を
  尊重させるため（nput モジュール自身も同じ形で engine を起動している）

### 3. standalone — shell 合成 + 終了コードで分岐する

CLI は UNIX の合成に乗る。終了コード表は
REQ-2c5a10d8-112b-4f96-947a-aba7164779c4（0 = 成功 / no-op / `--no-wait` の try-lock skip、
1 = 一般エラー、2 = `apply --dryrun` の conflict 検出）。

```sh
# 成否での分岐（0 / 1 の 2 値。apply に exit 2 は無い）
nput apply myconfig && ./post.sh
```

```sh
# CI の事前 gate。conflict（2）とエラー（1）を区別する
nput apply myconfig --dryrun
case $? in
  0) nput apply myconfig && ./post.sh ;;
  2) echo 'conflict detected — aborting' >&2; exit 2 ;;
  *) exit 1 ;;
esac
```

**exit 2 が返るのは `apply --dryrun` の conflict 検出だけ**で、通常の `apply` は 0 か 1 しか
返さない。3 値の分岐は dryrun を挟んだときにだけ意味を持つ。

### 4. 機械可読な結果が要るとき — `--json` を消費する

「何が変わったときだけ後処理する」「変更内容を通知に載せる」需要は、終了コードでは表せない
（no-op も成功も exit 0）。この場合は `--json` を消費して外部がオーケストレーションする。

```sh
result=$(nput apply myconfig --json)
if printf '%s' "$result" | jq -e '
      .status == "success"
      and ([.results[].result.changes // [] | length] | add > 0)
    ' >/dev/null; then
  ./post.sh
fi
```

- **`status` を先に見る**。途中失敗した run は undo ジャーナルで巻き戻されるが、`changes` は
  失敗時点までに生じた差分の記録として残る（subject に `W_NPUT_UNWOUND` が付く）。
  `changes` だけで分岐すると、**ディスク上に何も残っていない run でも後処理が走る**
- **`results[]` は全 subject を畳んでから判定する**。`.results[].result.changes | length > 0` は
  subject ごとに真偽値を 1 個ずつ出すため、`jq -e` の終了コードが最後の 1 個で決まる
  （`--all` で「先頭は変化あり・末尾は no-op」だと変化なしと判定される）

`--json` は stdout に niface エンベロープを 1 文書だけ出す opt-in の第 2 契約で、
`items` / `changes` / `info` は各 `results[i].result` 配下に入る
（→ REQ-a5053191-1c6a-449b-9c5e-5ff49dc5aead）。`status` は終了コード表に連動する
（exit 0 → `success`、exit 1・2 → `error`）。この経路は engine から見れば
「実行して結果を返しただけ」で、後続の判断は全て呼び出し側にある。

### 4 idiom の使い分け

| 需要 | idiom |
|---|---|
| シェル入室のたびに前後の処理を並べたい | 1（`shellHook`）|
| `home-manager switch` の一連の流れに組み込みたい | 2（activation DAG）|
| 成否・conflict で分岐したい | 3（shell 合成 + 終了コード）|
| 変化の有無・内容で分岐したい | 4（`--json` 消費）|

いずれの層でも、後処理を走らせる判断は**呼び出し側が持つ**。engine が持つのは配置そのものだけで、
この分担が REQ-c1b3ca5f-d2f7-443c-bc4b-b18413ca97b9 の「配線と配置コアの分離」を idiom の面で
実現している。

## 出典

ADR-0041「engine イベント hook スクリプト機構を採用しない（合成 idiom を docs に置く）」の
「代替: 合成 idiom を docs に 1 節置く」節。同 ADR は hook が欲しくなる場面ごとに層別 idiom
（`shellHook` / activation DAG / shell 合成 + 終了コード / `--json` 消費）を示す節を docs へ
追加することを帰結として定めており、本 item がその実体。

ADR 本文は置き場所を「spec.md（または usage ガイド側）」と書いているが、docs 縮退（→ epic
#203）後の `docs/spec.md` は requirement への索引で散文を足す場所ではないため、design item と
して起こした。

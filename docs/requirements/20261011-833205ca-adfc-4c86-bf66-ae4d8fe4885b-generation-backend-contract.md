---
id: "REQ-833205ca-adfc-4c86-bf66-ae4d8fe4885b"
type: requirement
name: "engine は世代操作を世代バックエンドの契約（Commit / List / Switch / UserBase / SystemBase）越しに行う"
derives_from:
  - "UC-0b6f60cb-3e98-4ee7-8929-4d94a29f0af6"
specification: |
  The engine SHALL perform every generation operation — committing a link-farm as a new
  generation, listing generations and switching to a generation — only through a
  generation backend contract, and SHALL NOT invoke a generation tool directly. The
  contract SHALL provide five operations whose inputs and outputs are paths and integers
  only, with no Go callbacks: Commit, which takes the profile link and a link-farm path;
  List, which takes the profile link and returns the generations (number, date, whether
  current); Switch, which takes the profile link and a generation number; UserBase, which
  maps a state directory to the base directory of user profiles; and SystemBase, which
  returns the base directory of system profiles. Every engine entry point that resolves a
  profileDir SHALL require a backend and SHALL fail with an error when none is injected;
  there SHALL be no implicit fallback to `nix-env`. The default backend SHALL be
  nixprofile, which implements Commit and Switch with `nix-env --profile <profileLink>`
  (`--set`, `--switch-generation`), implements List by scanning the generation links in
  profileDir read-only without invoking `nix-env`, and owns the two bases (`<state>/nix/profiles/layat` and
  `/nix/var/nix/profiles/layat`). The layout inside profileDir (`profile`,
  `profile-<N>-link`, `.pending`) SHALL be a fixed part of the contract: the engine MAY
  keep reading it directly, and the removal of `.pending` after a successful commit SHALL
  remain a plain filesystem operation of the engine. While nixprofile is the only backend,
  no selection mechanism (flag, environment variable or configuration key) and no
  configuration key for the base directories SHALL be provided.
specification_ja: |
  engine は世代操作（link-farm の新世代としてのコミット・世代一覧・世代の切替）を世代
  バックエンドの契約越しにのみ行わなければならず、世代ツールを直接呼び出してはならない。
  契約は入出力がパスと整数だけ（Go コールバックを含まない）の 5 操作を持たなければならない:
  profile リンクと link-farm のパスを取る Commit、profile リンクを取り世代（番号・日付・
  current か）を返す List、profile リンクと世代番号を取る Switch、state ディレクトリから
  ユーザー profile の基底を導く UserBase、system profile の基底を返す SystemBase。profileDir
  を解決する engine の全入口はバックエンドを要求しなければならず、注入が無ければ error で
  失敗しなければならない。`nix-env` への暗黙のフォールバックを持ってはならない。既定の
  バックエンドは nixprofile でなければならず、Commit と Switch を `nix-env --profile
  <profileLink>`（`--set` / `--switch-generation`）で実装し、List は `nix-env` を呼び出さず
  profileDir の世代リンクを読み取り専用で走査して実装し、2 つの基底（`<state>/nix/profiles/layat` と `/nix/var/nix/profiles/layat`）を所有する。
  profileDir 内部のレイアウト（`profile` / `profile-<N>-link` / `.pending`）は契約の固定部分
  でなければならず、engine はこれを直接読んでもよく、コミット成功後の `.pending` の除去は
  engine の素の FS 操作のままでなければならない。nixprofile が唯一のバックエンドである間、
  選択機構（フラグ・環境変数・設定キー）と基底を上書きする設定キーを設けてはならない。
---
# REQ-833205ca-adfc-4c86-bf66-ae4d8fe4885b: engine は世代操作を世代バックエンドの契約（Commit / List / Switch / UserBase / SystemBase）越しに行う

## 仕様

| 操作 | 入力 | 出力・事後条件 |
|---|---|---|
| Commit | profile リンク + link-farm のパス | link-farm を新世代としてコミットする（nixprofile: `nix-env --profile <profileLink> --set <linkFarm>`） |
| List | profile リンク | 世代の一覧（番号・日付・current）を返す（nixprofile: profileDir の `profile-<N>-link` を読み取り専用で走査する） |
| Switch | profile リンク + 世代番号 | その世代へ profile ポインタを移す（nixprofile: `--switch-generation N`） |
| UserBase | state ディレクトリ | ユーザー profile の基底（nixprofile: `<state>/nix/profiles/layat`） |
| SystemBase | なし | system profile の基底（nixprofile: `/nix/var/nix/profiles/layat`） |

- 入出力はパスと整数だけ。Go コールバックを契約に含めないのは、同じ境界をそのまま
  プロセス境界（将来の世代管理 CLI）へ載せられる形に保つため。
- profileDir を解決する全入口（apply / rollback / reset / prune / profile 解決）が
  バックエンドを要求し、nil は error。`nix-env` への暗黙のフォールバックは持たない。
- 基底の置き場所は世代バックエンドの責務で、layat の設定インターフェイスでは上書きできない。
  `UserBase` / `SystemBase` がその第一歩で、世代管理が layat から出たときは基底の所有ごと出る。
- profileDir 内部のレイアウトは契約の固定部分。engine の manifest 読み（前世代の
  `manifest.json`・世代リンクの解決）は契約を経由せず直接読む。外部バックエンドが実在した
  時点で読み操作を契約へ移す。
- `.pending` の後片付けは build 出力の一時リンクの片付けであり、engine の `os.Remove` のまま。
  pending の有界性は REQ-840b3641-6e76-46da-82e9-680cabd65abe の担当。

engine が叩く外部コマンドが `git` のみになることは REQ-6c4e174a-4d16-477a-96ff-17cb4eb5b564、
世代操作と GC の委譲先は REQ-706de717-4e47-471a-a1c0-448635be159c、パッケージ構成は
DSG-f8d9bba5-cfd5-4a19-b1fd-fd8ecc8a1814 の担当。

## 出典

ADR-0057「世代バックエンドの契約」§1〜§6。

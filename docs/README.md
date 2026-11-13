# layat コンセプト

layat が何であり何を解決するかの全体像を述べる散文。item の一覧は、生成した [index.md](index.md) が持つ。

解決したい課題・解決策の核心・想定する使われ方は **`docs/solution/` と `docs/use-cases/` の
item が持つ**。本文書は通読の起点として全体像を散文で述べる。
「設計の哲学」「既存ツールとの比較」「north-star」は item を立てず本文書に書き下す。

> **この文書の書き方（規約）**
>
> - 散文は **見出し 1 つ（h2 / h3 の最も内側）につき 10 行以内**。表・リンク列挙・
>   コードブロックはこの制限に含めない
> - item 化された内容は **リンクで示し、本文に書き下さない**
> - **item を立てず散文で書き下してよいのは 3 節（哲学・比較・north-star）に限る**。
>   索引の節（課題と核心・想定する使われ方・設計の変遷）は本文書に置くが、内容は item / ADR が
>   持つ。この 3 節以外で書き下したい内容が出てきたら、item を立てて本文書からはリンクへ置き換える
> - item を横断して検索・追跡するには `sara query`（→ `CLAUDE.md` の「item を辿る（`sara query`）」節）を使う
> - item の `## 出典` の読み方は `docs/model.md` の「item の `## 出典` の読み方」の節に従う

---

## 解決したい課題とコンセプトの核心

Nix で外部リポジトリを取得する手段はあるが、任意パスへ配置する既製手段は無い。home-manager は
配置できるが「全体管理」モデルであり、1 つの変更が全体に影響する。さらにモジュール抽象は
「何がどこにどう置かれるか」を内部実装に隠し、ユーザーから配置の制御を奪う。

layat はフレームワークではなく **配置プリミティブ**である。「nix store のパスを root 相対の
任意パスへ symlink または copy で置く」という単一の責務を、テスト可能な純粋関数として提供し、
設定は生成しない。

---

## 想定する使われ方

中心は project mode（プロジェクト内配置）。`$HOME` 配置（home mode）・system 配置は明示マーカーで
opt-in する例外として位置づける（→ [project mode を第一の使い方とする](adr/project-first.md)）。

---

## 設計の哲学

**取得と配置の分離**: 取得は Nix の評価フェーズ（`src` = ストアパス）、配置は実行フェーズ。
取得手段（npins / flake inputs / `fetchFromGitHub` 等）をツール側が抱えず「フェッチ済みの
ストアパスを受け取る」設計にすることで、取得方法の変化から独立する。

**配置ロジックはコアが所有し、モジュールは配線に徹する**: 配置の実体は全層で layat 自身の
固定 Go エンジンが実行し（→ [配置は全層で engine が行い、ネイティブ機構へ翻訳しない](adr/engine-owns-placement-no-native-translation.md), [engine を単一の固定 Go バイナリにする](adr/engine-fixed-go-binary.md)）、`home.file` / `systemd.tmpfiles` などの
ネイティブ機構には委譲しない。振る舞いが単一コアに集約され、テストと一貫性を担保できる。

**home-manager に依存しない**: どの環境でも同じ設定定義で動き、統合層はコアの薄いラッパー。

**冪等性と粒度の柔軟性**: 同じ設定を何度実行しても同じ結果になる。リポジトリ全体・
サブディレクトリ・単一ファイルを同一インターフェースで扱い、呼び出し側は型を意識しなくてよい。

---

## north-star: 配置プリミティブから組むミニマル distro

長期的な狙いは、nixpkgs のパッケージ群（＝ストアパス）を活かしつつ配置だけをユーザーに操作させ、
Arch / Gentoo 的なミニマル Linux ディストリビューションの基盤を作ること（→ [コアを store パスを root 相対の target へ置くプリミティブにする](adr/core-as-placement-primitive.md)）。
このためコアの中心抽象は root を `$HOME` に固定せず一般化する。

layat は他のツール群と stdout / stdin の JSON パイプで合成するエコシステムの一員でもある。
機械可読出力（`--json`）は **すべての機能で** outturn specVersion 1 規約に準拠する（→ [--json は outturn 規約に常時準拠する](adr/outturn-conformance-principle.md)）。

**スコープの線引き**: CLI 単体 + project mode をコアとし、home mode と system mode
（`systemRoot` で `/` 配下へ root 権限で配置。CLI を root で実行するか、NixOS / nix-darwin module の
activation から起動する。→ [system mode の変更系操作の前に engine が euid 0 を確認する](adr/system-mode-euid-gate.md), [NixOS / nix-darwin モジュールは root を systemRoot に固定する](adr/system-modules-pin-systemroot.md)）も実装済み。「関数ベースのパッケージ導入・PATH 追加」に
専用機構は持たない。`pkgs.buildEnv` の合成環境を entry として配置し、環境変数の配線は利用者側で行う（→ [layat.lib.mkEnv で buildEnv 合成環境を提供する](adr/mkenv-buildenv.md)）。
ブート / init / FS / パーティションの層は本構想でも空白であり、layat が担う領域ではない。

---

## 既存ツールとの比較

比較軸は「機能の有無」ではなく **「モジュール抽象で隠すか、純粋関数でユーザーに握らせるか」**。

| ツール | 役割 | アプローチ | layat との違い |
|---|---|---|---|
| npins / niv | ソースのバージョン固定 | — | 配置は行わない（layat と直交・併用可）|
| home-manager `home.file` | ファイル配置 + 世代 | モジュール（生成・宣言）| HM 必須。全体管理。file モジュールを standalone 切り出し不能 |
| `mkOutOfStoreSymlink`（HM）| out-of-store symlink | モジュール内ヘルパ | HM 文脈限定。layat は同等を非依存の明示関数で提供 |
| nixpkgs `linkFarm` / `symlinkJoin` | store 内 symlink ツリー生成 | 純粋関数 | 出力が**ストア内に閉じる**。store 外の任意パスへは置かない（layat は内部で利用可）|
| `nix profile` | 世代管理機構 | — | 配置先が `~/.nix-profile` 固定。任意パス配置はしない（layat が乗る対象）|
| `systemd.tmpfiles`（`L`）| 任意パスへの宣言的 symlink | モジュール（NixOS）| 低レベル・NixOS 専用・copy/世代/取得抽象なし |
| numtide/system-manager | 非NixOS の `/etc` + systemd + パッケージ | モジュール（`lib.evalModules`）| **ドメインは重なるがアプローチが逆**。任意パス配置・HOME dotfiles・サブディレクトリ取り出しはしない |
| `git clone`（シェル）| クローンと配置 | 命令的 | 再現性・Nix との統合がない |
| **本ツール** | 取得済みソースの独立配置 + 世代 + 明示 out-of-store | **純粋関数・ユーザー管理** | — |

layat とほぼ同一のツールは存在しない。構成要素（symlink farm / nix profile / out-of-store /
任意パス symlink）はすべて既存だが、それらを「取得手段非依存 + 生成しない + エントリ個別適用 +
HM 非依存の純粋関数コア + クロスプラットフォーム共通スキーマ + 任意パス配置 × 世代管理」として
束ねたものは無い。特に system-manager とはドメインこそ重なるが、思想レベルでアプローチが
異なるため競合しない。

### home-manager `home.file` との配置意味論の差

`home.file` と layat はどちらも前世代との diff で古くなった配置を除去するが、layat は前世代の
**自己記録の manifest** を一次情報に持ち、HM は on-disk の readlink パターンマッチで判定する
（HM 現行実装〔2026-07 時点〕との比較・→ [祖先 symlink は自己記録の古くなったものだけを移行する](adr/ancestor-symlink-recorded-stale-only.md), [配置前除去は drift を見つけたら error で止める](adr/preremove-drift-errors.md)）。差は次の 5 点に表れる。

1. **同名 leaf を含む per-file → dir symlink 遷移の自動移行** — HM は旧 leaf の残存を誤認して失敗しうるが、layat は manifest 記録との一致判定（recorded ∧ stale）で安全に移行する（→ [実 dir の target は全 leaf が除去可能なときだけ全体を配置前除去して symlink 化する](requirements/real-dir-target-migration.md)）
2. **所有判定の厳密さ** — HM は readlink の glob マッチ、layat は記録済みの配置先のうち readlink が記録 dest と完全一致するものだけを所有とみなす（→ [古くなった配置の除去は前世代の記録通りを指す symlink のみに限り、copy は消さず孤児を警告する](requirements/stale-removal-invariants.md)）
3. **配置を塞ぐ空 dir の除去** — HM は collision で停止するが、layat は rmdir が空 dir にしか成功しないことを利用し、由来を問わず除去する（→ [実 dir の target は全 leaf が除去可能なときだけ全体を配置前除去して symlink 化する](requirements/real-dir-target-migration.md)）
4. **祖先 symlink の安全性** — HM は確認せずに辿るが、layat は foreign な祖先 symlink を conflict で停止し、自己記録の古くなった配置だけを移行する（→ [祖先 symlink は自己記録の古くなったものだけを配置前除去し、それ以外はエラーで停止する](requirements/ancestor-symlink-nest-migration.md)）
5. **rename 可用性 + fail-fast drift** — 配置を塞ぐ依存除去だけを前段化し、その除去が drift を検出したら skip せず error で停止する（→ [実 dir の target は全 leaf が除去可能なときだけ全体を配置前除去して symlink 化する](requirements/real-dir-target-migration.md)・[rollback は FS を先に収束させてから profile ポインタを最後に移す](requirements/rollback-refit-then-pointer.md)）

空親ディレクトリ剪定・conflict 全件報告は同等。copy は HM に存在しないため比較の対象外。

---

## 設計の変遷

主要な設計判断の記録。全件は `docs/adr/` を参照。

- [store link を既定にし、out-of-store は明示関数でのみ選ばせる](adr/store-link-default-out-of-store-opt-in.md) — store link をデフォルトに統一し、out-of-store は明示関数へ降格
- [世代は link farm と nix profile に乗せる](adr/generations-on-link-farm-and-nix-profile.md), [CLI 単体では世代を常時有効にする](adr/standalone-generations-always-on.md), [copy は世代の外に place-once で置く](adr/copy-outside-generations-place-once.md) — nix profile に乗せた CLI 単体の世代管理。copy は世代外
- [配置は全層で engine が行い、ネイティブ機構へ翻訳しない](adr/engine-owns-placement-no-native-translation.md), [モジュールは root と起動の時機だけを供給して engine を起動する](adr/modules-wire-engine-kick-only.md) — 配置ロジックは全層でエンジンが所有し、モジュールは配線に徹する
- [コアを store パスを root 相対の target へ置くプリミティブにする](adr/core-as-placement-primitive.md), [root は 3 つの marker で表す](adr/three-root-markers.md), [root の明示を mkManifest 層で必須にする](adr/root-explicit-at-mkmanifest.md) — root を一般化し配置プリミティブにする
- [root は 3 つの marker で表す](adr/three-root-markers.md) — git toplevel 相対の project mode
- [engine を単一の固定 Go バイナリにする](adr/engine-fixed-go-binary.md), [manifest.json を schemaVersion 1 の唯一の契約として固定する](adr/manifest-json-contract-schemaversion-v1.md) — 配置ロジックを固定 Go エンジンに集約し、契約は manifest.json
- [CLI が一次 UX で devShell に pin 版を同梱する](adr/cli-primary-ux-devshell-pinned.md), [root の明示を mkManifest 層で必須にする](adr/root-explicit-at-mkmanifest.md) — 汎用 `layat` CLI を一次 UX にし、root は明示必須
- [配置元の中のパスを全層で subpath と呼ぶ](adr/subpath-naming-across-layers.md), [リポジトリ全体は subpath の省略で表す](adr/whole-repo-by-omitting-subpath.md) — `subpath` で取り出し、全体選択は省略で表す
- [entries を target をキーにした attrset にする](adr/entries-keyed-by-target.md) — `entries` は target キーの attrset
- [apply --recopy で全 copy target を src から無条件に上書きする](adr/apply-recopy-overwrites-copies.md), [reset で profile を触らずに配置物を撤去する](adr/reset-fs-only-teardown.md) — `method = "copy"`（世代外・place-once）・`--recopy`・`reset`（関連: [絶対パスと .. エスケープは eval 時に拒否する](adr/path-safety-eval-rejection.md), [symlink farm を GC アンカー専用にする](adr/farm-as-gc-anchor-only.md), [project mode では copy と out-of-store も ephemeral として扱う](adr/copy-and-out-of-store-ephemeral-in-project.md), [空の entries を正当な全クリアとして許可する](adr/empty-entries-allowed.md), [reset は config の名指しを必須にする](adr/reset-requires-config-name.md), [reset --dryrun は副作用なしで撤去対象を示す](adr/reset-dryrun-side-effect-free.md), [flock は明示実行で待ち、hook 実行でスキップする](adr/flock-mode-explicit-wait-hook-skip.md), [apply の修飾フラグを --all と合成できるようにする](adr/apply-modifiers-compose-with-all.md)）
- [apply --manifest でビルド済み link-farm を engine へ直接渡す](adr/apply-manifest-prebuilt-link-farm.md), [モジュールの activation は apply --all --manifest を 1 回だけ起動する](adr/module-activation-single-apply-all.md) — モジュールはビルド済み link-farm を `apply --manifest` で直接適用する
- [指定が無いときは成功時に何も出力しない](adr/silent-on-success.md), [-v で配置レポートを opt-in で出す](adr/verbose-opts-into-placement-report.md) — 成功時はデフォルト沈黙。`-v` で opt-in 表示
- [世代操作はパスと整数だけを受け渡す 5 操作の契約越しに行う](adr/generation-backend-contract-paths-and-ints.md), [engine が直接叩く外部コマンドを git に限る](adr/engine-external-command-git-only.md) — 世代操作を世代バックエンドの契約越しにし、engine から nix-env と profile 基底の固定を外す
- [last-applied 記録は manifest を内容ごと持つ](adr/record-embeds-manifest.md), [古くなった配置の除去は、比較の基準を記録から優先して読む](adr/baseline-prefers-record.md) — 最後に配置した manifest を last-applied 記録に持ち、古くなった配置の除去で使う比較の基準を世代の概念から FS の事実へ移す
- [root 相対マーカーを別関数 mkRootRelativeSymlink で作る](adr/mkrootrelativesymlink-separate-function.md), [root 相対マーカーのリンクを target 起点の相対パスで張る](adr/root-relative-link-as-relative-path.md) — 配置 root 相対の src を `mkRootRelativeSymlink` で宣言し、engine が target 起点の相対 symlink として張る

### 名前の由来

旧名 `nput` の「n」は **nix** を指していたが、engine の契約は `manifest.json` 1 本で生成者を
問わない。nix に依存するのは既定の生成器（nix）と世代バックエンド（nixprofile）で、engine ではない
ため、「n」が指すものは曖昧だった。現名 **layat** は
"**lay** \<src\> **at** \<target\>" の圧縮造語で、manifest の通りに src を target へ置く動作を
一語で表す（`chroot`・`getopt` と同じ動詞句の圧縮）。

- [ツール名を layat にする](adr/tool-named-layat.md) — `nput` を layat へ改名し、2 週間の改名予告期間を挟む（命名条件・棄却した候補・移行方針）

---

## 関連文書

- `README.md` / `README.ja.md` — 3 層構造の最上段（導入と使い方）
- `docs/requirements/README.md` — 仕様の全体像（`requirement`）
- `docs/design/README.md` — 設計の全体像（`design`）
- `docs/adr/` — 意思決定の記録
- `docs/model.yaml` — sara の型定義（item の型・関係・ID 形式）

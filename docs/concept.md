# layat コンセプト

layat が何であり何を解決するかの全体像と、solution / use_case item への索引。

解決したい課題・解決策の核心・想定する使われ方は **`docs/solution/` と `docs/use-cases/` の
item が持つ**。本文書は通読の入口として全体像を述べ、詳細は item へのリンクで示す
（README → 本文書 → item の 3 層構造）。「設計の哲学」「既存ツールとの比較」「north-star」は
item を立てず本文書に書き下す。

> **この文書の書き方（規約）**
>
> - 散文は **見出し 1 つ（h2 / h3 の最も内側）につき 10 行以内**。表・リンク列挙・
>   コードブロックはこの制限に含めない
> - item 化された内容は **リンクで示し、本文に書き下さない**
> - **item を立てず散文で書き下してよいのは 3 節（哲学・比較・north-star）に限る**。
>   索引の節（課題と核心・想定する使われ方・設計の変遷）は本文書に置くが、内容は item / ADR が
>   持つ。この 3 節以外で書き下したい内容が出てきたら、item を立てて本文書からはリンクへ置き換える
> - item を横断して検索・追跡するには `sara query`（→ `docs/agents/domain.md`）を使う
> - item の `## 出典` は本文書の現行章立てを指さない。原文の読み方を含め `docs/spec.md` の
>   同項が正（同項は 3 文書共通の手順として書いてある）

---

## 解決したい課題とコンセプトの核心

Nix で外部リポジトリを取得する手段はあるが、任意パスへ配置する既製手段は無い。home-manager は
配置できるが「全体管理」モデルであり、1 つの変更が全体に影響する。さらにモジュール抽象は
「何がどこにどう置かれるか」を内部実装に隠し、ユーザーから配置の制御を奪う。

layat はフレームワークではなく **配置プリミティブ**である。「nix store のパスを root 相対の
任意パスへ symlink または copy で置く」という単一の責務を、テスト可能な純粋関数として提供し、
設定は生成しない。

- [SOL-9fcd1d6e-6204-42e6-92bb-1faf966f0b3e](solution/20260802-9fcd1d6e-6204-42e6-92bb-1faf966f0b3e-layat-placement-primitive.md) — layat は nix store の物を任意パスへ置く配置プリミティブであり、設定を生成せずユーザーが配置を明示的に握る

---

## 想定する使われ方

中心は project mode（プロジェクト内配置）。`$HOME` 配置（home mode）・system 配置は明示マーカーで
opt-in する例外として位置づける（→ ADR-0007）。

- [UC-19a90989-0ae3-438f-8a75-4e1e2637f81c](use-cases/20260802-19a90989-0ae3-438f-8a75-4e1e2637f81c-project-mode-in-repo-placement.md) — プロジェクト repo 内へ nix store の物を devShell 入室のたびに配置してチームで共有する
- [UC-f2436d68-91ff-4c48-b1df-47acefe4f464](use-cases/20260802-f2436d68-91ff-4c48-b1df-47acefe4f464-home-mode-pinned-repo-placement.md) — home mode で外部リポジトリの中身をバージョン固定して $HOME 配下の任意パスへ配置する
- [UC-1c280dce-7c72-44c0-95ea-d06344f62a47](use-cases/20260802-1c280dce-7c72-44c0-95ea-d06344f62a47-independent-update-cycles.md) — 役割ごとに config を分けて更新を独立させ、1 つの更新を他の役割へ波及させない
- [UC-0b6f60cb-3e98-4ee7-8929-4d94a29f0af6](use-cases/20260802-0b6f60cb-3e98-4ee7-8929-4d94a29f0af6-generation-rollback-standalone.md) — 配置に失敗・後悔したとき standalone で前の世代へロールバックして元の状態へ戻す
- [UC-403fbe32-b146-401b-8b53-fe67c1e169c5](use-cases/20260802-403fbe32-b146-401b-8b53-fe67c1e169c5-copy-place-once-user-managed.md) — リポジトリの内容を copy で初回だけ配置し、その後はユーザーが手で編集して育てる
- [UC-01b896b4-04b9-40d0-bf9e-966eaf64c3d4](use-cases/20260802-01b896b4-04b9-40d0-bf9e-966eaf64c3d4-out-of-store-live-editing.md) — 開発中の手元 dotfiles を out-of-store symlink で参照し、編集と同時に反映しながら育てる
- [UC-d39c1994-f9a5-4860-80ba-f6e584adaf14](use-cases/20260802-d39c1994-f9a5-4860-80ba-f6e584adaf14-module-integration.md) — 既に home-manager / NixOS / nix-darwin を使っている環境へ layat をモジュールとして組み込む

---

## 設計の哲学

**取得と配置の分離**: 取得は Nix の評価フェーズ（`src` = ストアパス）、配置は実行フェーズ。
取得手段（npins / flake inputs / `fetchFromGitHub` 等）をツール側が抱えず「フェッチ済みの
ストアパスを受け取る」設計にすることで、取得方法の変化から独立する。

**配置ロジックはコアが所有し、モジュールは配線に徹する**: 配置の実体は全層で layat 自身の
固定 Go エンジンが実行し（→ ADR-0003, ADR-0006）、`home.file` / `systemd.tmpfiles` などの
ネイティブ機構には委譲しない。振る舞いが単一コアに集約され、テストと一貫性を担保できる。

**home-manager に依存しない**: どの環境でも同じ設定定義で動き、統合層はコアの薄いラッパー。

**冪等性と粒度の柔軟性**: 同じ設定を何度実行しても同じ結果になる。リポジトリ全体・
サブディレクトリ・単一ファイルを同一インターフェースで扱い、呼び出し側は型を意識しなくてよい。

---

## north-star: 配置プリミティブから組むミニマル distro

長期的な狙いは、nixpkgs のパッケージ群（＝ストアパス）を活かしつつ配置だけをユーザーに操作させ、
Arch / Gentoo 的なミニマル Linux ディストリビューションの基盤を作ること（→ ADR-0004）。
このためコアの中心抽象は root を `$HOME` に固定せず一般化する。

layat は他のツール群と stdout / stdin の JSON パイプで合成するエコシステムの一員でもある。
機械可読出力（`--json`）は **すべての機能で** outturn specVersion 1 規約に準拠する（→ ADR-0043）。

**スコープの線引き**: 実装スコープは standalone CLI + project mode をコアとし home mode も
対象、system 配置は将来拡張（→ ADR-0007）。「関数ベースのパッケージ導入・PATH 追加」の具体機構は
未定義。ブート / init / FS / パーティションの層は本構想でも空白であり、layat が担う領域ではない。

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

`home.file` と layat はどちらも前世代との diff で stale を除去するが、layat は前世代の
**自己記録の manifest** を一次情報に持ち、HM は on-disk の readlink パターンマッチで判定する
（HM 現行実装〔2026-07 時点〕との比較・→ ADR-0046, ADR-0047）。差は次の 5 点に表れる。

1. **per-file → dir symlink 遷移の自動移行** — HM は旧 leaf の残存を誤認して失敗しうるが、layat は manifest 記録との一致判定（recorded ∧ stale）で安全に移行する
2. **所有判定の厳密さ** — HM は readlink の glob マッチ、layat は記録済みの配置先のうち readlink が記録 dest と完全一致するものだけを所有とみなす
3. **配置を塞ぐ空 dir の除去** — HM は collision で停止するが、layat は rmdir が空 dir にしか成功しないことを利用し、由来を問わず除去する
4. **祖先 symlink の安全性** — HM は無検査で辿るが、layat は foreign な祖先 symlink を conflict で停止し、自己記録の stale だけを移行する
5. **fail-fast drift** — 前段化した依存除去が drift を検出したら skip せず error で停止する

空親ディレクトリ剪定・conflict 全件報告は同等。copy は HM に存在しないため比較の対象外。

---

## 設計の変遷

主要な設計判断の記録。全件は `docs/adr/` を参照。

- [ADR-0001](adr/0001-out-of-store-as-explicit-function.md) — store link をデフォルトに統一し、out-of-store は明示関数へ降格
- [ADR-0002](adr/0002-generations-on-nix-profile.md) — nix profile に乗せた standalone 世代管理。copy は世代外
- [ADR-0003](adr/0003-engine-owns-placement-modules-are-wiring.md) — 配置ロジックは全層でエンジンが所有し、モジュールは配線に徹する
- [ADR-0004](adr/0004-root-generalization-and-distro-positioning.md) — root を一般化し配置プリミティブにする
- [ADR-0005](adr/0005-project-mode-and-ephemeral-placement.md) — git toplevel 相対の project mode
- [ADR-0006](adr/0006-engine-as-go-binary-lib-produces-data.md) — 配置ロジックを固定 Go エンジンに集約し、契約は manifest.json
- [ADR-0007](adr/0007-cli-as-primary-ux-and-entrypoint-discovery.md) — 汎用 `layat` CLI を一次 UX にし、root は明示必須
- [ADR-0008](adr/0008-source-to-subpath-and-whole-repo-by-omission.md) — `subpath` で取り出し、全体選択は省略で表す
- [ADR-0014](adr/0014-entries-as-target-keyed-attrset.md) — `entries` は target キーの attrset
- [ADR-0020](adr/0020-copy-recopy-and-entry-reset.md) — `method = "copy"`（世代外・place-once）・`--recopy`・`reset`（関連: ADR-0019, ADR-0021）
- [ADR-0026](adr/0026-module-activation-applies-prebuilt-manifest.md) — モジュールはビルド済み link-farm を `apply --manifest` で直接適用する
- [ADR-0031](adr/0031-silent-on-success-output-discipline.md) — 成功時はデフォルト沈黙。`-v` で opt-in 表示

### 名前の由来

旧名 **nput** の「n」は **nix** を指していたが、engine の契約は `manifest.json` 1 本で生成者を
問わず、store を介さない運用もあるため、「n」が指すものは曖昧だった。現名 **layat** は
"**lay** \<src\> **at** \<target\>" の圧縮造語で、manifest の通りに src を target へ置く動作を
一語で表す（`chroot`・`getopt` と同じ動詞句の圧縮）。

- [ADR-0054](adr/0054-rename-nput-to-layat.md) — nput を layat へ改名し、2 週間の改名予告期間を挟む（命名条件・棄却した候補・移行方針）

---

## 関連文書

- `README.md` / `README.ja.md` — 3 層構造の最上段（導入と使い方）
- `docs/spec.md` — 仕様（requirement item への索引）
- `docs/design.md` — 設計（design item への索引）
- `docs/adr/` — 意思決定の記録
- `docs/model.yaml` — sara の型定義（item の型・関係・ID 形式）

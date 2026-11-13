# layat 仕様書

layat が「何を満たすべきか」の全体像を述べる散文。`requirement` item の一覧は、生成した
[index.md](index.md) が持つ。

規範的な仕様は **すべて `docs/requirements/` の item が持つ**。本文書は通読の起点として全体像だけを
散文で述べ、item の一覧は持たない。

> **この文書の書き方（規約）**
>
> - 散文は **見出し 1 つ（h2 / h3 の最も内側）につき 10 行以内**。表・リンク列挙・コード
>   ブロックはこの制限に含めない
> - **詳細は必ず item へリンクし、本文に書き下さない**。仕様の断片をここへ写すと item と
>   二重管理になり、改訂時に片方だけ古くなる
> - 本文書へ仕様を書き足さない
> - item を横断して検索・追跡するには `sara query`（→ `CLAUDE.md` の「item を辿る（`sara query`）」節）を使う
> - item の `## 出典` の読み方は `docs/model.md` の「item の `## 出典` の読み方」の節に従う

---

## アーキテクチャ概要

layat は **CLI とエンジンの 2 層**で構成する（→ [engine を単一の固定 Go バイナリにする](../adr/engine-fixed-go-binary.md), [engine は cmd/ と internal/ に置きライブラリ化しない](../adr/engine-internal-cmd-layout.md)）。CLI は manifest 生成器
（→ [manifest の取得をプロセス内の Go インターフェイスの生成器越しにする](../adr/manifest-generator-in-process-interface.md)、既定 `nix`）を通じて entrypoint を発見し、生成器の Build をエンジンへ渡す。エンジンは
ロック内で Build を呼び、生成器が build した link-farm のパス（既定の nixprofile では store path）を
受け取る。エンジンは配置内容（entry）を `manifest.json` から読み、ネイティブ FS 操作で配置と古くなった配置の除去を行う。
世代のコミットと切替は世代バックエンド（→ [世代操作はパスと整数だけを受け渡す 5 操作の契約越しに行う](../adr/generation-backend-contract-paths-and-ints.md)、既定 nixprofile）に任せる。
配置内容をエンジンへ渡すのは `manifest.json` だけ。

---

## lib API

`lib`（`mkManifest` / `mkApplyCommand` / マーカー群）は nixpkgs.lib のみに依存する純データ生成器で、配置ロジックは
持たない（→ [lib は純データ生成に徹する](../adr/lib-produces-data-only.md)）。入力の確認は `mkManifest` 自身が `evalModules` + `normalizeManifest` で
行い、1 か所に集約する。

---

## CLI 仕様（一次 UX）

`layat` CLI は PATH に常駐する一次 UX（→ [CLI が一次 UX で devShell に pin 版を同梱する](../adr/cli-primary-ux-devshell-pinned.md)）。entrypoint を発見し、rootKind を先取り
eval して root を解決し、flock を取ってから `nix build` をロック内で回してエンジンを駆動する。

---

## entries スキーマ仕様

`entries` は **属性キー = target の attrset**で、キーが識別子になる（→ [entries を target をキーにした attrset にする](../adr/entries-keyed-by-target.md)）。

---

## manifest.json スキーマ（v1・Nix↔Go 契約）

`manifest.json` は Nix とエンジンの唯一の安定契約（→ [manifest.json へ marker を `clean` enum で写す](../adr/wire-format-clean-enum.md), [manifest.json を schemaVersion 1 の唯一の契約として固定する](../adr/manifest-json-contract-schemaversion-v1.md)）。

---

## 配置動作仕様

symlink は配置前除去 → 配置 → 古くなった配置の除去の順で進み、途中失敗はインメモリ undo ジャーナルで
巻き戻す（→ [途中失敗した apply をインメモリ undo ジャーナルで巻き戻す](../adr/in-memory-undo-journal.md), [配置前除去を planner が計算するデータとして表す](../adr/preremove-as-planner-data.md)）。copy は place-once でユーザー管理に委ねる。

---

## 世代管理仕様

世代は link farm derivation を layat 自前の profile へコミットして積む（→ [世代は link farm と nix profile に乗せる](../adr/generations-on-link-farm-and-nix-profile.md), [profile は専用ディレクトリレイアウトに置く](../adr/profile-directory-layout.md)）。
古くなった配置の除去は前世代 manifest の記録通りを指す symlink だけに限る保守的な操作。

---

## root の解決

root は評価時にパスへ展開せず、マーカーが運ぶ root の種別をエンジンが実行時に解決する
（→ [root は 3 つの marker で表す](../adr/three-root-markers.md)）。

---

## モジュールオプション仕様・モジュール別動作仕様

全モジュール（HM / NixOS / nix-darwin）と devShell は、root と activation タイミングを供給して
エンジンを起動する薄い配線に徹する（→ [モジュールは root と起動の時機だけを供給して engine を起動する](../adr/modules-wire-engine-kick-only.md)）。ネイティブ機構へは翻訳しない。

---

## エラー仕様

設定の誤りは評価時に、実体の不整合はエンジン実行時に検出する層分けを守る。要求された操作が
成立しないときは暗黙のフォールバックを採らず、エラーで停止する。

---

## 設定の書き方（本文書の対象外）

実際に動く設定例は本文書では扱わない。`templates/`（`layat init` が展開する実物）と README の
コード例を参照する。本文書が扱うのは「何を満たすべきか」であり、書き方の例ではない。

---

## 関連文書

- `README.md` / `README.ja.md` — 3 層構造の最上段（導入と使い方）
- `docs/README.md` — コンセプト（`solution` / `use_case` の全体像）
- `docs/design/README.md` — 設計の全体像（`design`）
- `docs/infrastructure/` — 品質方針を実現する技術基盤（CI / リリース / 配信）
- `docs/adr/` — 意思決定の記録。`requirement` / `design` / `quality` / `test_plan` などを `justifies` で裏づける
- `docs/model.yaml` — sara の型定義（item の型・関係・ID 形式）

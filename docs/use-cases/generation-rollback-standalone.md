---
id: "UC-0b6f60cb-3e98-4ee7-8929-4d94a29f0af6"
type: use_case
name: "配置に失敗・後悔したとき CLI 単体で前の世代へロールバックして元の状態へ戻す"
status: stable
refines:
  - "SOL-9fcd1d6e-6204-42e6-92bb-1faf966f0b3e"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T22:03:12+09:00"
---
# 配置に失敗・後悔したとき CLI 単体で前の世代へロールバックして元の状態へ戻す

## 使われ方

`src` を更新して適用したところ、新しいバージョンのプラグインに不具合があった・意図しない内容が
配置された、といった場面で、ユーザーは前の世代へ戻して復旧する。

```bash
layat apply <name>            # 新世代を作って適用（name 省略時は layat.default）
layat rollback <name>         # 前世代へ戻す
layat list-generations <name> # 世代一覧
```

世代は nix profile に乗る（→ [世代は link farm と nix profile に乗せる](../adr/generations-on-link-farm-and-nix-profile.md)）。純粋関数 `lib.mkManifest` が link farm derivation
（`manifest.json` + ストア内の symlink ツリー）を生成し、engine が nix profile に登録する
ことで、世代番号・GC root・ロールバックを Nix 標準機構から得る。粒度は `layat.<name>`
単位 = 1 profile なので、役割ごとに独立したロールバック系列を持つ。

```
純粋関数:  lib.mkManifest が link farm derivation（manifest.json + symlink farm）を生成（副作用なし）
実行時:    固定 Go エンジンが配置し profile を swap（副作用はここだけ）
```

lib が副作用を持たず、engine だけが実行時に副作用を起こすことは [lib は純データ生成に徹する](../adr/lib-produces-data-only.md) が定める。

世代管理の対象になるのは symlink 配置。out-of-store symlink はリンク先のみが世代管理され
（指す先の内容は設計上ライブ）、copy は世代外である。

| 配置種別 | 世代管理 | 備考 |
|---|---|---|
| symlink（store）| あり（profile）| ロールバック可能 |
| out-of-store symlink | あり（リンク先のみ）| 指す先の内容は設計上ライブ。版管理しない |
| copy | **なし（世代外）** | place-once・ユーザー管理の副作用 |

> **上表は分割時点の `docs/concept.md`（原文）からの写し**。use_case は規範を持たないため、
> 配置種別ごとの世代管理の有無の規範は
> [method は配置方法を選び symlink は世代管理下・copy は世代管理外になる](../requirements/method-src-matrix.md)（method と世代管理の対応）/ [copy は target 不在のときだけマテリアライズする place-once で世代管理の対象外とする](../requirements/copy-place-once.md)（copy が
> 世代外であること）/ [out-of-store symlink は marker の絶対パスまたは root 相対パスを指し、版管理はリンク先マッピングのみとする](../requirements/out-of-store-link-mapping-only.md)（out-of-store の版管理範囲）にある。

この使われ方は CLI 単体（home mode）が対象。project mode は配置物が ephemeral なため
rollback を公開せず、モジュール経由のときはユーザー向け rollback を host に一本化する
（→ [プロジェクト repo 内へ nix store の物を devShell 入室のたびに配置してチームで共有する](project-mode-in-repo-placement.md) / [既に home-manager / NixOS / nix-darwin を使っている環境へ layat をモジュールとして組み込む](module-integration.md)）。

なお、下に挙げる要求のうち古くなった配置の除去は本 use_case に固有のものではなく、全モードの apply が
持つ冪等性の不変条件である。この使われ方に固有なのは「戻した世代の記録が除去の基準になる」
点に留まる。

## この使われ方が要求すること

- 適用のたびに新世代が積まれ、profile の差し替えで atomic に切り替わること
- 前世代へ戻すコマンドと、世代を一覧するコマンドがあること
- ロールバックが FS の状態まで含めて収束すること
- 適用やロールバックが途中で失敗しても中途半端な状態を残さないこと
- 世代を戻した先の記録に無い配置物が保守的に除去され、前世代の配置状態へ収束すること
- 世代を積み続けたときの GC・不要世代の間引き手段があること

## 出典

`docs/concept.md`「世代管理（standalone）」。

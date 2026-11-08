---
id: "DSG-107c19ea-bc90-4914-aaaa-5f2da4f3e082"
type: design
name: "配置プリミティブを store link・copy・out-of-store symlink の 3 種に定め、世代・reset・gitignore での扱いを種別ごとに固定する"
status: stable
satisfies:
  - "REQ-622787dc-4512-4ce9-9c7d-7b32bbb70557"
  - "REQ-d2277c7a-7992-49af-a9dc-4cc73843a6f9"
  - "REQ-a8a923ad-07fb-4582-b90a-07a6e0c41baa"
  - "REQ-31f2882e-d2e3-4e3b-b783-feb627d73ac6"
  - "REQ-60787ed2-4176-4bdd-800f-1600c0315551"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-07T02:02:33+09:00"
---
# DSG-107c19ea-bc90-4914-aaaa-5f2da4f3e082: 配置プリミティブを store link・copy・out-of-store symlink の 3 種に定め、世代・reset・gitignore での扱いを種別ごとに固定する

## 設計

entry が置く物の種別（配置プリミティブ）は 3 つに限る。種別は entry の `method`（`symlink` /
`copy`）と、`src` が out-of-store marker かどうかの 2 軸で決まる。2 × 2 のうち
「`copy` × out-of-store marker」は意図が矛盾するので評価時に拒否する
（→ [意図が矛盾する組み合わせをクロスフィールドの確認で評価時に拒否する](../requirements/normalize-manifest-cross-field-checks.md)）。残る 3 マスが次の 3 種になる。

| 種別 | `method` | `src` | 置かれる物 |
|---|---|---|---|
| store link | `symlink`（既定）| store パス | store パスを指す symlink |
| out-of-store symlink | `symlink` | `mkOutOfStoreSymlink` / `mkRootRelativeSymlink` の marker | ローカルの絶対パス、または target の親から root 相対の src への相対パスを指す symlink |
| copy | `copy` | store パス | store パスの内容を複製した実ファイル・実ディレクトリ |

store link と out-of-store symlink はどちらも symlink として実現されるため、「symlink」とだけ
書くと種別を特定できない（→ [store link](../ubiquitous-language/store-link.md)・[out-of-store symlink](../ubiquitous-language/out-of-store-symlink.md)）。この item と下流の
item では必ず 2 つを書き分ける。既定は store link で、out-of-store symlink は明示 marker で
選ぶ退避路に留める（→ [layat は nix store の物を任意パスへ置く配置プリミティブであり、設定を生成せずユーザーが配置を明示的に握る](../solution/layat-placement-primitive.md)）。

### 種別ごとの扱い

| | 世代（apply / rollback）| reset | gitignore |
|---|---|---|---|
| store link | リンク先の store パスごと世代が版管理する。rollback で戻り先世代の store パスを指し直す。entry が消えたら除去する | 記録どおりを指すものだけを除去する | 列挙する |
| out-of-store symlink | 世代が版管理するのは「どのパスを指すか」だけで、指す先の内容は版管理しない。entry が消えたら除去する | 記録どおりを指すものだけを除去する | 列挙する |
| copy | 世代管理の対象外。apply も rollback も target が不在のときだけ place-once で置く。entry が消えても消さず、孤児を警告する | 削除する（copy を消す唯一の明示手段）| 列挙する |

- **世代** — store link は内容が不変の store パスを指すので、マッピングを戻せば内容も戻る。
  out-of-store symlink は指す先がライブなので、戻るのはマッピングだけ
  （→ [out-of-store symlink は marker の絶対パスまたは root 相対パスを指し、版管理はリンク先マッピングのみとする](../requirements/out-of-store-link-mapping-only.md)）。copy はユーザー所有のデータとして扱い、
  世代を移っても内容を書き換えない。store の更新を反映するのは `apply --recopy` か、
  `reset` してから再 apply するかの 2 つだけ（→ [copy は target 不在のときだけマテリアライズする place-once で世代管理の対象外とする](../requirements/copy-place-once.md)）。
  entry が消えたときに何を除去してよいかの保守的な不変条件は
  [古くなった配置の除去は前世代の記録通りを指す symlink のみに限り、copy は消さず孤児を警告する](../requirements/stale-removal-invariants.md) が持つ
- **reset** — profile と世代に触れない FS-only の撤去で、2 種の symlink は entry が消えたときと
  同じ不変条件で除去し、copy は削除する。いずれの撤去後も空の親ディレクトリを剪定する
  （→ [reset は profile を触らない FS-only teardown で配置物を無い状態へ戻す](../requirements/reset-fs-only-teardown.md)）
- **gitignore** — 種別を区別せず全 target を列挙する。project mode の copy も clone ごとに
  再マテリアライズされる ephemeral な配置物で、コミット対象にしない
  （→ [gitignore は method を区別せず copy target も含めて全 target を列挙する](../requirements/gitignore-includes-copy-targets.md)。project mode 限定は
  [gitignore は project mode 限定で非 project config を指定したらエラーで停止する](../requirements/gitignore-project-mode-only.md)）

`method` の変更による移行は片方向だけで、`symlink` → `copy` は配置前除去で移行し、
`copy` → `symlink` は移行しない（→ [method 変更は symlink→copy のみ配置前除去で移行し、copy→symlink は移行しない](../requirements/method-change-symlink-to-copy-migration.md)）。

## 出典

[store link](../ubiquitous-language/store-link.md)・[out-of-store symlink](../ubiquitous-language/out-of-store-symlink.md) と [layat は nix store の物を任意パスへ置く配置プリミティブであり、設定を生成せずユーザーが配置を明示的に握る](../solution/layat-placement-primitive.md) が述べる
種別の定義を、上掲の REQ が個別に定める世代・reset・gitignore の振る舞いと 1 枚の表に束ねた。

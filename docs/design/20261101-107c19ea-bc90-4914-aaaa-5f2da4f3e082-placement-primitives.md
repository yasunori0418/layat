---
id: "DSG-107c19ea-bc90-4914-aaaa-5f2da4f3e082"
type: design
name: "配置プリミティブを store link・copy・out-of-store symlink の 3 種に定め、世代・reset・gitignore での扱いを種別ごとに固定する"
satisfies:
  - "REQ-622787dc-4512-4ce9-9c7d-7b32bbb70557"
  - "REQ-d2277c7a-7992-49af-a9dc-4cc73843a6f9"
  - "REQ-a8a923ad-07fb-4582-b90a-07a6e0c41baa"
  - "REQ-31f2882e-d2e3-4e3b-b783-feb627d73ac6"
  - "REQ-60787ed2-4176-4bdd-800f-1600c0315551"
---
# DSG-107c19ea-bc90-4914-aaaa-5f2da4f3e082: 配置プリミティブを store link・copy・out-of-store symlink の 3 種に定め、世代・reset・gitignore での扱いを種別ごとに固定する

## 設計

entry が置く物の種別（配置プリミティブ）は 3 つに限る。種別は entry の `method`（`symlink` /
`copy`）と、`src` が out-of-store marker かどうかの 2 軸で決まる。2 × 2 のうち
「`copy` × out-of-store marker」は意図が矛盾するので評価時に拒否する
（→ REQ-16faf428-77f3-492f-b858-222c5274cbf7）。残る 3 マスが次の 3 種になる。

| 種別 | `method` | `src` | 置かれる物 |
|---|---|---|---|
| store link | `symlink`（既定）| store パス | store パスを指す symlink |
| out-of-store symlink | `symlink` | `mkOutOfStoreSymlink` / `mkRootRelativeSymlink` の marker | ローカルの絶対パス、または target の親から root 相対の src への相対パスを指す symlink |
| copy | `copy` | store パス | store パスの内容を複製した実ファイル・実ディレクトリ |

store link と out-of-store symlink はどちらも symlink として実現されるため、「symlink」とだけ
書くと種別を特定できない（→ `docs/glossary.md`「Flagged ambiguities」）。この item と下流の
item では必ず 2 つを書き分ける。既定は store link で、out-of-store symlink は明示 marker で
選ぶ退避路に留める（→ SOL-9fcd1d6e-6204-42e6-92bb-1faf966f0b3e）。

### 種別ごとの扱い

| | 世代（apply / rollback）| reset | gitignore |
|---|---|---|---|
| store link | リンク先の store パスごと世代が版管理する。rollback で戻り先世代の store パスを指し直す。entry が消えたら除去する | 記録どおりを指すものだけを除去する | 列挙する |
| out-of-store symlink | 世代が版管理するのは「どのパスを指すか」だけで、指す先の内容は版管理しない。entry が消えたら除去する | 記録どおりを指すものだけを除去する | 列挙する |
| copy | 世代管理の対象外。apply も rollback も target が不在のときだけ place-once で置く。entry が消えても消さず、孤児を警告する | 削除する（copy を消す唯一の明示手段）| 列挙する |

- **世代** — store link は内容が不変の store パスを指すので、マッピングを戻せば内容も戻る。
  out-of-store symlink は指す先がライブなので、戻るのはマッピングだけ
  （→ REQ-a8a923ad-07fb-4582-b90a-07a6e0c41baa）。copy はユーザー所有のデータとして扱い、
  世代を移っても内容を書き換えない。store の更新を反映するのは `apply --recopy` か、
  `reset` してから再 apply するかの 2 つだけ（→ REQ-d2277c7a-7992-49af-a9dc-4cc73843a6f9）。
  entry が消えたときに何を除去してよいかの保守的な不変条件は
  REQ-16aef46b-7bb8-4ca1-b962-e9f3ed1fd1d2 が持つ
- **reset** — profile と世代に触れない FS-only の撤去で、2 種の symlink は entry が消えたときと
  同じ不変条件で除去し、copy は削除する。いずれの撤去後も空の親ディレクトリを剪定する
  （→ REQ-31f2882e-d2e3-4e3b-b783-feb627d73ac6）
- **gitignore** — 種別を区別せず全 target を列挙する。project mode の copy も clone ごとに
  再マテリアライズされる ephemeral な配置物で、コミット対象にしない
  （→ REQ-60787ed2-4176-4bdd-800f-1600c0315551。project mode 限定は
  REQ-eaa8c0df-af44-4f52-9603-cd2bc22a67e9）

`method` の変更による移行は片方向だけで、`symlink` → `copy` は配置前除去で移行し、
`copy` → `symlink` は移行しない（→ REQ-2b48620a-abaa-43df-a106-954bbba3de56）。

## 出典

`docs/glossary.md`「Placement kinds」と SOL-9fcd1d6e-6204-42e6-92bb-1faf966f0b3e が述べる
種別の定義を、上掲の REQ が個別に定める世代・reset・gitignore の振る舞いと 1 枚の表に束ねた。

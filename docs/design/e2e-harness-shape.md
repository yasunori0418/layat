---
id: "DSG-2947b4a5-ba5c-4218-8478-624cba156116"
type: design
name: "E2E は tests/e2e の bash ハーネスに置き、シナリオの詳細は同ディレクトリの README へ委譲する"
status: stable
satisfies:
  - "TP-229b69c0-cf5e-4fb6-a353-27e5064d93e9"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:36:11+09:00"
---
# E2E は tests/e2e の bash ハーネスに置き、シナリオの詳細は同ディレクトリの README へ委譲する

## 設計

[非 NixOS で動く主張を実 nix の一気通貫 E2E で検証する](../test-plan/e2e-harness-scope.md) が要求する一気通貫 E2E の実装形を次のように定める。

| 項目 | 選択 |
|---|---|
| 配置 | `tests/e2e/` |
| 実装言語 | bash |
| シナリオ詳細 | `tests/e2e/README.md` へ委譲 |

隔離した一時 `$HOME` / `$XDG_STATE_HOME` の下で動かすこと、偽 src をどう供給するかは
[非 NixOS で動く主張を実 nix の一気通貫 E2E で検証する](../test-plan/e2e-harness-scope.md) が規範として定めており、ハーネスはそれを前提に組む。

**bash を選ぶ**のは、E2E が検証するのが「ユーザーが端末で `layat` を叩いたときに
何が起きるか」であり、テスト自体もその叩き方をそのまま書ける形が最も直接的なため。
Go のテストから CLI を起動する形にすると、engine のテスト（[engine は nix を使わない tmpdir 統合テストで検証し、古くなった配置の除去の安全不変条件を table-driven に重点配分する](engine-test-go-unit-tmpdir.md)）と
同じ言語・同じテストランナーに乗ってしまい、「外から叩いている」ことが
構造として見えなくなる。実 nix を叩き `$HOME` を隔離するという性質上、
Go のテストフレームワークが提供する利点も薄い。

**`tests/e2e/` へ分離する**のは、このテスト群だけが実 nix と隔離 `$HOME` を要求し、
`nix flake check` の集約対象に載せられない（サンドボックス外で実行する必要がある）ため。
場所で分けておけば、実行系統の違いがディレクトリ境界に一致する。

**シナリオ詳細を `README.md` へ委譲する**のは、各シナリオが検証する仕様そのものは
テスト計画側（[非 NixOS で動く主張を実 nix の一気通貫 E2E で検証する](../test-plan/e2e-harness-scope.md) とその参照先）が持ち、ハーネス側に必要なのは
「どう起動しどう組むか」の手順だからである。仕様をハーネス側にも書くと
二重管理になる。

## 出典

`docs/design.md`「テスト戦略」の E2E 段落。

---
id: "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
type: use_case
name: "プロジェクト repo 内へ nix store の物を devShell 入室のたびに配置してチームで共有する"
status: stable
refines:
  - "SOL-9fcd1d6e-6204-42e6-92bb-1faf966f0b3e"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:30:05+09:00"
---
# プロジェクト repo 内へ nix store の物を devShell 入室のたびに配置してチームで共有する

## 使われ方

layat の**中心的な配置モード**。任意のプロジェクトに layat を組み込み、root をプロジェクト
ルートに解決して repo 内の任意パスへ nix store の物を配置する（→ [root は 3 つの marker で表す](../adr/three-root-markers.md) / [project mode を第一の使い方とする](../adr/project-first.md)）。
新規プロジェクトは `layat init project`、既に `flake.nix` がある既存 repo へは手動で組み込む
（→ [既存 flake への組込手順を docs に置く](../adr/existing-flake-integration-docs.md)）。

具体例:

- repo 内の `.claude/skills/` を nix store から配置してチームで共有する
- project-local な tool 設定・hook をバージョン固定で repo 内へ配置する
- 社内共有の設定リポジトリから特定ディレクトリだけを取り出してプロジェクトへ配置する

主トリガは devShell。`devShells.<name>` の `shellHook` から layat をキックし、`nix develop` /
direnv でプロジェクトに入ったときに配置される。CLI 本体は devShell の `packages` に pin 版
`layat` を同梱するのが canonical（→ [CLI が一次 UX で devShell に pin 版を同梱する](../adr/cli-primary-ux-devshell-pinned.md)）。

```nix
# entrypoint(flake.nix)が manifest を公開し、devShell で layat apply する
layat.${system}.skills = layat.lib.mkManifest {
  inherit pkgs;
  root = layat.lib.projectRoot;
  entries = {
    ".claude/skills/nix" = { src = inputs.claude-skills; subpath = "skills/nix"; };
  };
};
devShells.${system}.default = pkgs.mkShell {
  packages  = [ layat.packages.${system}.layat ];   # pin 版 layat を PATH へ
  shellHook = "layat apply skills --no-wait";
};
```

> **上は分割時点の `docs/concept.md`（原文）からの写し**（コメントの出典注記は省いた）。
> use_case は使われ方を述べる層で規範を持たないため、この例が示す各要素の規範は
> `requirement` 側にある
> （devShell 同梱 → [layat CLI は PATH 常駐の一次 UX で、project mode は devShell 同梱を canonical とする](../requirements/cli-primary-ux-installation.md)、`layat.<name>` のアドレッシング → [entrypoint は `layat.<name>` に named manifest を公開し CLI は形ごとの attr path で build する](../requirements/named-manifest-addressing.md)、
> devShell からの engine 起動 → [devShell は shellHook から engine を起動する配線で、シェル入室のたびに project mode で配置する](../requirements/devshell-wiring.md)）。

配置物は per-clone で再生成される前提の **ephemeral** であり、プロジェクトにはコミットされ
ない。したがって activation は git 状態に干渉せず、`.gitignore` に入れるべき target は専用
コマンド `layat gitignore` で列挙してプロジェクト管理者が一度登録する。

## この使われ方が要求すること

- root がプロジェクトルートとして解決されること（git toplevel 相対・`--root` で上書き可）
- 配置物が ephemeral として扱われ、activation が git 状態に干渉しないこと
- `.gitignore` へ登録すべき target を列挙する手段があること
- devShell の `shellHook` からキックする配線が成立すること。入室のたびに実行されるため、変化が
  無いときの再配置コストが抑えられていること
- 世代は内部機構に留め、ephemeral な配置に意味の薄い rollback を公開しないこと
- 配置をやめたくなったときに、配置物を無い状態へ戻す手段があること（`reset`。rollback と違い
  モードを限定せず project mode でも使える）

## 出典

`docs/concept.md`「プロジェクトに閉じた配置（project mode）」と「想定ユースケース」の
project mode 節。

あわせて「設計の哲学」の「home-manager に依存しない」節も、[home mode で外部リポジトリの中身をバージョン固定して $HOME 配下の任意パスへ配置する](home-mode-pinned-repo-placement.md) とともに本 item が
引き受ける。同節が述べる「HM 不在の環境でも同じ設定定義で動く」ことを体現するのは統合ではなく
CLI 単体の使われ方であり、本 item では devShell の `shellHook` から engine を起動する配線が
HM を介さずに成立することが同節に対応する。

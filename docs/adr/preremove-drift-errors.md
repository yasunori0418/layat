---
id: "ADR-2d4fdda4-b2e1-4161-8b71-3cff5b265f2f"
type: adr
name: "配置前除去は drift を見つけたら error で止める"
status: stable
sources:
  - id: "issue-170"
    resource: "https://github.com/yasunori0418/layat/issues/170"
    title: "#170"
  - id: "issue-175"
    resource: "https://github.com/yasunori0418/layat/issues/175"
    title: "#175"
  - id: "issue-172"
    resource: "https://github.com/yasunori0418/layat/issues/172"
    title: "#172"
origin: "grilling セッション（2026-07-11・2026-07-12）で確定。ADR-0046 §3（配置前除去の drift の扱い）と ADR-0047 §3 を書き直した"
justifies:
  - "REQ-7cee95dd-bc5a-4e86-bebc-6080ef78fe26"
  - "DSG-8b96869c-842e-4f78-8ff7-df1f1d6c1a68"
  - "REQ-0e341430-17f0-498b-9439-65491652163a"
references:
  - "ADR-88231ffc-4836-4252-be34-894944fd28c3"
  - "ADR-095220c9-5a3d-4fd4-94bd-8fa1a02dd36d"
  - "ADR-9755ba76-4c02-4b4e-a8c0-031a24a3a5b9"
  - "ADR-62540f8f-a8ba-4441-97b0-36b50d7862d8"
  - "ADR-1cbacca1-1930-46ca-bf82-30310f82216e"
  - "ADR-4c8d459a-c371-4f0d-b009-16e2ca096e49"
  - "ADR-eb1aa25f-a0ca-4efe-bcd9-4cbe11368764"
  - "ADR-ffe1afcb-0b51-4b87-b698-0a6dea307bdd"
  - "ADR-8b8c9109-a479-4571-8b10-5a6079127441"
  - "ADR-1908099c-87f7-4df1-bda5-7731df7d9ed7"
  - "ADR-efa7685c-6fa4-45df-b7a3-8f3f58e4118c"
  - "ADR-764ffff1-6b5a-4ed7-89d0-5700eb3ee883"
  - "ADR-9723dbb3-6659-4a69-ac4f-411c98bfa818"
  - "ADR-8357b519-cb23-4984-ba56-254046779f47"
  - "ADR-4f08aa18-6a38-49fd-a784-450677d80e1a"
  - "ADR-d97154cd-a2e5-4853-a682-800ff9435c89"
  - "ADR-3e283fe2-0177-4caf-8b60-577c5ee3b10a"
  - "ADR-b06dd8ab-317e-41ee-9c80-dd70dcd7f2fd"
  - "ADR-40dd7945-5792-4b80-bdfa-43f95f482554"
  - "ADR-74423bb8-1453-4689-a25e-30bf46536afc"
  - "ADR-9f2cc068-8edb-4e3d-958b-0dff8f6ed90c"
  - "ADR-3b8175b9-5ebf-4123-8855-f698dc843d6b"
  - "ADR-dfe6654c-8fcd-4cf6-a053-c21ad6dc373c"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:27:41+09:00"
---
# ADR-2d4fdda4-b2e1-4161-8b71-3cff5b265f2f: 配置前除去は drift を見つけたら error で止める

- ステータス: stable
- 日付: 2026-11-03
- 関連: [CLI が一次 UX で devShell に pin 版を同梱する](cli-primary-ux-devshell-pinned.md), [cross-config の同一 target は後勝ちにし、foreign symlink の上書きで警告する](cross-config-last-wins-with-foreign-warning.md), [世代バックエンドを nix 実装 1 つに留める](single-nix-generation-backend.md), [世代操作はパスと整数だけを受け渡す 5 操作の契約越しに行う](generation-backend-contract-paths-and-ints.md), [世代の一覧を nix-env を経由せず世代リンクの走査で得る](generation-list-scans-links.md), [世代バックエンドのテストダブルを FS ベースの Fake に統一する](fs-based-generation-fake.md), [世代バックエンドの注入を必須にし nix-env への暗黙のフォールバックを持たない](generation-backend-required-no-fallback.md), [profileDir の内部レイアウトを世代バックエンド契約の固定部分にする](profiledir-layout-fixed-in-contract.md), [前世代の読み取りは契約に載せず engine が直接読む](generation-reads-outside-contract.md), [engine が直接叩く外部コマンドを git に限る](engine-external-command-git-only.md), [Build の事後条件を世代バックエンドがコミットできる link-farm のパスに置く](build-postcondition-committable-link-farm.md), [state の基底を設定で上書きさせない](state-bases-not-configurable.md), [.pending の後片付けを gcroot を知らない engine の FS 操作として行う](pending-cleanup-without-gcroot.md), [配置前除去による移行を warning にせず -v のときだけ報告する](migration-reported-silently.md), [配置前除去の除去も undo の対象にする](undo-covers-preremove.md), [rollback は FS を収束させてからポインタを動かす](rollback-baseline-and-pointer-last.md), [entry の配置種別のフィールドを method と名付ける](mode-renamed-to-method.md), [manifest.json を schemaVersion 1 の唯一の契約として固定する](manifest-json-contract-schemaversion-v1.md), [flake check の警告は無害として許容する](flake-check-warnings-tolerated.md), [新規配置と張替えを先に、世代のコミットを最後に行う](placement-order-commit-last.md), [古くなった配置の除去は記録どおりの symlink だけに限る](conservative-removal-invariant.md), [配置前除去を planner が計算するデータとして表す](preremove-as-planner-data.md), [祖先 symlink は自己記録の古くなったものだけを移行する](ancestor-symlink-recorded-stale-only.md), GitHub Issue #170, #175, #172
- 起点: grilling セッション（2026-07-11・2026-07-12）で確定。ADR-0046 §3（配置前除去の drift の扱い）と ADR-0047 §3 を書き直した

## 背景

配置前除去（[配置前除去を planner が計算するデータとして表す](preremove-as-planner-data.md)）は、plan を計算した時点の FS をもとに
除去するものを決める。実行するまでに、対象の symlink が別の dest に張り替えられたり、空だった
ディレクトリに中身が入ったりすることがある（drift）。最終段の古くなった配置の除去は、記録どおりの
symlink だけを消し（[古くなった配置の除去は記録どおりの symlink だけに限る](conservative-removal-invariant.md)）、drift したものは warning を
出して残す。配置前除去でも同じ扱いでよいかを決める必要があった。

## 決定

配置前除去は、除去の直前に記録どおりかを確かめ直し、drift を見つけたら skip せず error で
止める。symlink が記録の dest と食い違うか消えていれば drift とする。ディレクトリは事前に空かを
調べず、rmdir の結果で判定し、空でなければ drift とする。これは配置前除去の全ての発生源（祖先の
symlink、実ディレクトリの target の配下、method の変更）に共通する。最終段の古くなった配置の除去の扱いは
変えない。

除去の手順は [実 dir の target は全 leaf が除去可能なときだけ全体を配置前除去して symlink 化する](../requirements/real-dir-target-migration.md) と
[配置ロジックを plan を計算する planner と FS 意味論を実行する engine に分け、dryrun と apply が同一の plan を共有する](../design/planner-engine-split.md) に、rollback の配置前除去も drift で error に
することは [rollback は FS を先に収束させてから profile ポインタを最後に移す](../requirements/rollback-refit-then-pointer.md) に置く。

## 根拠

後続の配置は、配置前除去で対象が消えていることを前提に、無条件で子を置く。drift を skip して
配置を続けると、drift した後の実体（foreign な symlink や書き込める実ディレクトリ）の配下へ
書き込んでしまう。祖先の symlink で止めて防いでいた store の汚染（[祖先 symlink は自己記録の古くなったものだけを移行する](ancestor-symlink-recorded-stale-only.md)）を、再び許す
ことになる。最終段の古くなった配置の除去は後に何も続かないので、残して warning にとどめてよい。

rmdir は空のディレクトリでしか成功しないので、その結果がそのまま確かめ直しになる。事前に空かを
調べると、調べてから消すまでの間に中身が入る隙が残る。

error で止まっても、冪等な再実行で現在の FS に対して plan を計算し直し、収束する
（[新規配置と張替えを先に、世代のコミットを最後に行う](placement-order-commit-last.md)）。

## 棄却した代替案

### drift を warning にして skip する（home-manager と同じ扱い）

drift した後の実体を経由して、store や foreign な場所を汚染する隙が開く。

### drift を見つけたら plan を計算し直して再試行する

仕組みが複雑になるうえ、冪等な再実行による収束と役割が重なる。

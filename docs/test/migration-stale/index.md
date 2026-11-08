# migration-stale

## テスト条件（`test_condition`）

* [空親ディレクトリの剪定が多階層を辿り、root・symlink・非空で停止すること](./ancestor-pruning-bounds.md) - target 除去後の親ディレクトリ剪定が、空である限り多階層を遡ること、および停止すべき境界で止まることを検証する。境界は 3 つある: root 自体（root は消さない）、symlink である祖先（その先へは進まない）、非空のディレクトリ（foreign な実ファイルを含む場合を含む）。
* [plan から unlink までの drift が除去直前に再検証され、処理ごとに正しく分岐すること](./drift-reverification.md) - plan 時に除去対象と判定された target が、実際の unlink までの間に不変条件から外れた場合（リンク先が別の場所を指すようになった / symlink でなくなった / 消えた）を検出し、記録された除去をそのまま実行しないことを検証する。drift の同値類ごとに確認する。
* [移行が部分適用されず、中断後の再実行で収束すること](./migration-atomicity-convergence.md) - ディレクトリ移行の途中に移行不能な実体（foreign な実ファイル・非空サブディレクトリ）が見つかったとき、移行可能だった兄弟が一つも除去されないことを検証する。「1 つでも塞がれていればツリー全体を移行しない」という規則が、plan 段階での停止として現れることを含む。
* [配置前除去による移行の可否判定が、自己記録の古くなった配置と method 変更の規則どおりであること](./migration-eligibility.md) - 配置 target を占有する実体を配置前除去して移行してよいのは、それが自己記録の古くなったものである（あるいは空ディレクトリである）場合に限られることを検証する。実ディレクトリが占有する場合は、ツリー全体の leaf が全て「記録済みで古くなっている、または空」であるときにのみ移行できる。
* [古くなった配置の除去の対象が前世代に記録され記録先を指す symlink に限られること](./stale-removal-targets.md) - 新世代の manifest から消えた entry のうち、前世代 manifest に記録されており実 FS 上でも記録先を指したままである symlink だけが除去されることを検証する。記録に無い foreign symlink・通常ファイル・実ディレクトリ・copy entry は、除去の対象から外れる。

## テストケース（`test_case`）

* [tests/e2e/scenarios/03-stale.sh — 実 nix での古くなった配置の除去](./e2e-03-stale.md) - `tests/e2e/scenarios/03-stale.sh`
* [internal/engine/preremove_generalization_test.go — レイアウト移行と method 変更](./preremove-generalization-test-go.md) - `internal/engine/preremove_generalization_test.go`
* [internal/engine/staleremove_test.go — drift 再検証と空親剪定](./staleremove-test-go.md) - `internal/engine/staleremove_test.go`

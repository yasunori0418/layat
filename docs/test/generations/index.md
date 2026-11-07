# generations

## テスト条件（`test_condition`）

* [reset が保守的不変条件の範囲だけを消し、dryrun は副作用ゼロ、非同意では中断する](./20260809-06052178-ed49-4800-beb2-1d7c5d696ea8-reset-teardown-safety.md) - `Reset`（profile を触らない FS-only teardown → REQ-31f2882e-d2e3-4e3b-b783-feb627d73ac6）を、一度 apply して世代を コミットした状態から駆動して検証する。
* [実行結果のインベントリ・世代観測・到達状態・構造化 warning が実態を表す](./20260809-1d19aebc-e1e6-4d6f-9440-3efbd69b18a8-result-observation-fields.md) - `Apply` / `Rollback` が返す `Result`（`Rollback` は `RollbackResult` として embed する） と、`Reset` が返す別型の `ResetResult` の各フィールドが、その run で実際に起きたことと 一致することを検証する（→ issue #130。outturn envelope の素材）。commit は `nix-env --set` の世代簿記を模した double へ差し替え、世代番号を読み戻せる形にする。
* [世代スキップは project / system mode で derivation 同一のときだけ発火し、それ以外は新世代を積む](./20260809-1ee7a729-5629-4c92-b02a-4f6e11c0b57d-generation-skip-commit-policy.md) - 世代を積むか積まないかの判定を、mode と derivation の同一性の掛け合わせで検証する。 commit の呼び出し記録を観測点にして「新世代が積まれたか」を判定する。
* [rollback が FS を前世代へ再収束させてから最後にポインタを移す](./20260809-36ea3609-d52e-42d4-975c-40fb89b23919-rollback-reconvergence.md) - `Rollback` を tmpdir 上の実 FS で駆動し、現世代 N から N-1 への FS 再収束を検証する。 world は「gen1 と gen2 の link farm を書き、`profile-N-link` 世代リンクを張り、profile を gen2 へ向け、FS を gen2 の配置状態にする」形で組む。
* [state 基底・profileDir のキーイング・世代リンク名が root と mode ごとに正しく決まる](./20260809-40b762fc-3b48-471c-b495-c5ffe1f584b5-profile-path-keying.md) - profile の on-disk レイアウトを決めるパス解決層を、環境変数と入力の組み合わせに対する 関数として検証する。
* [実 nix を通した home mode で世代がコミットされ、rollback が前世代の配置へ戻る](./20260809-527b5034-715a-4df1-871a-072dc9062704-home-generation-e2e.md) - Go テストが double へ差し替えている層（`nix build` / `nix-env --set` / 世代一覧）を 実物に戻し、home mode の世代機構が一気通貫で動くことを検証する。仮 `$HOME` と `$XDG_STATE_HOME` で隔離した環境に fixture flake を書き、entry を入れ替えながら apply を重ねる。
* [profileDir の世代リンクを番号順に列挙して current を判定し、読めない profileDir は黙認しない](./20260809-746cb5b9-fe27-4d09-a51d-eb03d56a93a7-generation-listing-parse.md) - profileDir に並ぶ `profile-<N>-link` を読み取り専用で走査して世代一覧を作る層を、tmpdir に 張った世代リンクに対して検証する。
* [lstat ドリフト修復が破損した entry だけを直し、ユーザーの編集には触れない](./20260809-765858b6-d596-4428-b4da-5d432944a714-drift-repair-scope.md) - 世代スキップ時に実行される `lstat` ベースのドリフト修復について、「何を直し、何を直さないか」を 検証する。同一の link farm を 2 回 apply して `generationUnchanged` を成立させ、2 回目の run が修復だけを行う状況を作る。
* [rollback は途中失敗を巻き戻し、ポインタ移動の失敗は巻き戻さず、前提不成立では停止する](./20260809-fa7911c6-7347-47ff-b121-bec53562c063-rollback-partial-and-abort.md) - TC-36ea3609-d52e-42d4-975c-40fb89b23919 の正常系に対し、`Rollback` の失敗時の処理を検証する。
* [prune の dryrun が副作用ゼロで、非同意は中断し、lock 中の系列は skip され、削除段の失敗は進捗と理由で振り分けられる](./20260823-a9857bf7-f7f9-41f9-b42c-9993fd16a5e9-prune-safety-gates.md) - `Prune`（→ DSG-096dc893-21f4-45e3-9347-986e9275b4d1）の安全機構を、孤児系列を含む state dir の tmpdir で駆動して検証する。何を対象と判定するかは TC-bc653f83-70ce-4920-b7cd-9b70a4cd2cae の担当で、ここは「対象と判定したものを実際に消してよいか」 だけを見る。
* [prune が孤児系列だけを系列ごと消し、生存・判定不能な系列は理由付きで残す](./20260823-bc653f83-70ce-4920-b7cd-9b70a4cd2cae-prune-target-accuracy.md) - `Prune`（→ DSG-096dc893-21f4-45e3-9347-986e9275b4d1）を、複数の系列を並べた state dir の tmpdir に対して駆動して検証する。判定の入力は backref `.root` と root の実在だけなので、 fixture は state dir の形と root ディレクトリの有無で作る。
* [比較の基準を last-applied 記録 → profile → 空の順で読み、各操作が記録を FS の事実どおりに保つ](./20261011-3c28934c-d098-40a3-b598-1b08117f9cf1-last-applied-baseline-precedence.md) - 古くなった配置の除去・ドリフト修復・reset・rollback が使う比較の基準（この root に layat が最後に 置いたもの）について、読みの優先順位と、各操作が記録を書き換える時機と内容を検証する。
* [実 nix を通した system mode で世代がコミット・往復し、非 root の変更系は FS に触れずに止まる](./20261018-3f8d24a7-4e43-4998-b47f-a9e03cbcbfee-system-generation-e2e.md) - Go テストが double へ差し替えている層（euid・`nix build` / `nix-env --set`・世代一覧）を 実物に戻し、system mode の世代機構と権限の条件が一気通貫で動くことを検証する。root は user namespace で取り、system の profile 基底と記録基底は隔離先へ向ける。
* [status が記録基底の全系列を漏れなく列挙し、記録の内容をそのまま反映し、状態を変えない](./20261023-076bd7b4-6427-4a35-9bab-86daaaada504-status-inventory-fidelity.md) - engine の `Status` が記録基底を走査した結果について、網羅・内容・並び順と、読み取り専用で あることを検証する。
* [記録は版 2 で出所を持ち、読みは版 1 と 2 を受け付け、出所の食い違いを警告し、書き直しは出所を引き継ぐ](./20261031-f0958a4c-b585-49c5-96a9-d7beff571e71-origin-mismatch-detection.md) - last-applied 記録の出所（`origin`）の書き・読み・比較・引き継ぎを検証する。

## テストケース（`test_case`）

* [result_extensions_test.go — 実行結果のインベントリ・世代観測・到達状態・構造化 warning](./20260809-2008a909-b45e-49e1-adf1-39191a8ddd95-result-extensions-test-go.md) - `internal/engine/result_extensions_test.go`
* [drift_test.go — 世代スキップの発火条件と lstat ドリフト修復の対象範囲](./20260809-2de3a3d8-3b67-485a-8ddf-a97dd182af23-drift-test-go.md) - `internal/engine/drift_test.go`
* [e2e 02-home — 実 nix での home mode 世代コミットと rollback / reset の往復](./20260809-2e7b9ea2-4e4d-48ca-9e1e-0bfbb6223ff2-e2e-02-home.md) - `tests/e2e/scenarios/02-home.sh`
* [generations_test.go — rollback の再収束と失敗時の処理](./20260809-364ebb9d-b9f8-4c79-8509-8bbf31e73698-generations-test-go.md) - `internal/engine/generations_test.go`
* [reset_test.go — FS-only teardown の除去範囲・フィルタ・dryrun・同意](./20260809-503c9021-a43f-4c06-8ef6-8630a7d4aeac-reset-test-go.md) - `internal/engine/reset_test.go`
* [paths_test.go — state 基底の解決・roothash の性質・profileDir キーイング・世代リンク名と兄弟パス](./20260809-ff4a842e-dd25-4c75-82ef-185507781d02-paths-test-go.md) - `internal/paths/paths_test.go`
* [prune_test.go — 孤児系列の判定と系列ごとの削除・dryrun / 確認 / lock / 削除段の失敗](./20260823-49ac6d68-ed29-41aa-b939-dcfdc8711082-prune-test-go.md) - `internal/engine/prune_test.go`
* [e2e 08-prune — 実 apply が作った系列を root 削除で孤児にし、prune が系列ごと消して gcroot が外れる](./20260823-b19188cb-e3f0-4c52-8c92-6db96bbbb45a-e2e-08-prune.md) - `tests/e2e/scenarios/08-prune.sh`
* [record_test.go — last-applied 記録の読み書き・比較の基準の優先順位・各操作での更新](./20261011-233855f7-912c-4477-af32-9a54c82353c3-record-test-go.md) - `internal/engine/record_test.go`
* [nixprofile_test.go — 世代リンクの走査と 2 基底](./20261011-80a327c3-5fdd-48d4-9446-4b3e0e262105-nixprofile-test-go.md) - `internal/generation/nixprofile/nixprofile_test.go`
* [e2e 12-system — user namespace の root で system mode の配置・世代・権限の条件を往復する](./20261018-99803cb0-6fd0-4004-af78-83e301320ce0-e2e-12-system.md) - `tests/e2e/scenarios/12-system.sh`
* [status_test.go — 記録基底の列挙順・基底の扱い・読めない系列・読み取り専用](./20261023-152e6f7b-cc49-4ac0-ad8e-0aba62dc1371-engine-status-test-go.md) - `internal/engine/status_test.go`
* [e2e 13-status — 実 apply が書いた記録を status が列挙し、\[name\] で絞り、system 基底の記録も拾う](./20261023-b2a424d5-991d-4534-b52f-62c9704a0b8b-e2e-13-status.md) - `tests/e2e/scenarios/13-status.sh`

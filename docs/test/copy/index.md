# copy

## テスト条件（`test_condition`）

* [コピー中の syscall 失敗が伝播し、失敗した entry が結果へ記録されないこと](./20260809-4f315cfc-446c-4c3f-8dc9-c36b10073e9d-copy-failure-propagation.md) - コピー処理の各 syscall（mkdir / open / open-file / read-write / readlink / lstat）の失敗がそれぞれエラーとして呼び出し元へ伝わることを検証する。エラーメッセージによる段の特定は、診断上の要となる 2 つの処理（copy 配置時の親ディレクトリ作成・recopy の lstat）で確認し、残る段は伝播そのものまでを見る。
* [copy が target 不在時のみマテリアライズし、既存 target を触らないこと](./20260809-596d697f-4ba6-4ec1-b71e-8b5375806c08-place-once-idempotence.md) - target が不在のときにだけコピーが行われ、既に存在するときは内容・属性ともに一切変更されないことを検証する。「再 apply しても破損しない」だけでは不足で、配置後にユーザーが編集した内容がそのまま残ることまでを条件とする。
* [コピーがツリー構造・mode・symlink を src どおりに再現し、owner-write を付与すること](./20260809-b1b8c163-9d37-47ee-9838-7168569df03a-tree-fidelity.md) - 再帰コピーが src のツリーを構造ごと再現することを検証する。ネストしたディレクトリ、空ディレクトリ（多階層に渡る場合を含む）、ツリー内部の symlink、トップレベルが symlink である場合、のいずれもが元の形のまま複製されることを含む。symlink は deref せず symlink として複製されなければならない。
* [out-of-store marker が live symlink として配置され、src 側の変更が透過すること](./20260809-cf1b44ec-2ec4-4ee1-9a37-378e41ccb01e-out-of-store-live.md) - `mkOutOfStoreSymlink` で作った marker が、コピーではなく marker の絶対パスを指す symlinkとして配置されることを検証する。配置後に src 側のファイルを書き換えたとき、target 越しにその変更がそのまま見えることまでを条件とする（コピーとして配置されていれば見えない）。
* [apply --recopy が全 copy target を無条件に src から更新すること](./20260809-d1eb1814-ac5e-4576-b092-7db4929fba43-recopy-overwrite.md) - `apply --recopy` が config 内の copy target を（place-once の判定を経ずに）src から無条件に再コピーすることを検証する。foreign な実ファイルが占有している target も対象に含まれるため、通常 apply では skip される対象が recopy では更新されることを併せて確認する。

## テストケース（`test_case`）

* [tests/e2e/scenarios/04-copy.sh — 実 nix での copy place-once と out-of-store](./20260809-601db08c-2323-48eb-99f5-8774bba2fa6e-e2e-04-copy.md) - `tests/e2e/scenarios/04-copy.sh`
* [internal/engine/copytree_test.go — 再帰コピーの構造・mode・symlink 再現](./20260809-82fdfb27-0bea-4833-b5e9-5f9e142a08df-copytree-test-go.md) - `internal/engine/copytree_test.go`
* [internal/engine/copy_test.go — コピー処理の syscall 失敗注入](./20260809-dcd5e7e4-9190-4b13-9521-9b8801ec85fb-copy-test-go.md) - `internal/engine/copy_test.go`

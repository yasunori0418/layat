# nix-unit アグリゲータのマージロジック。readDir から切り離した純関数で、衝突検査をテストできる。
# `tests/nix-unit/` 配下に置くとテストファイルとして import されるため、ここに置く。
{ lib }:
{
  # modules: `{ file, tests }` のリスト（file = 定義元ファイル名、tests = テスト attrset）。
  # 返り値: 全 tests を `//` マージした attrset。テスト名がファイル横断で衝突していれば
  # マージせず throw する。
  mergeTests =
    modules:
    let
      # テスト名 -> それを定義しているファイル名のリスト。名前だけを見るので、
      # 衝突が無ければ各テストの値（expr / expected）は評価しない。
      ownersByTest = lib.zipAttrs (map (m: lib.mapAttrs (_testName: _: m.file) m.tests) modules);

      collisions = lib.filterAttrs (_testName: owners: lib.length owners > 1) ownersByTest;

      collisionReport = lib.concatStringsSep "\n" (
        lib.mapAttrsToList (
          testName: owners: "  - ${testName}: ${lib.concatStringsSep ", " owners}"
        ) collisions
      );

      merged = lib.foldl' (acc: m: acc // m.tests) { } modules;
    in
    if collisions == { } then
      merged
    else
      throw ''
        tests/nix-unit: テスト名がファイル横断で衝突しています（`//` の後勝ちマージで
        片方のアサートが実行されないまま緑になります）。ファイル固有の接頭辞を付けて
        一意にしてください（→ TP-36e90d5d）。
        ${collisionReport}'';
}

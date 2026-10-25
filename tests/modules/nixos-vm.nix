# NixOS モジュールの実 activate を VM で検証する。起動するのは machine（構成 a）だけで、
# 構成 b（node b）は起動せず toplevel を machine の nix DB へ登録して switch 先に使う。
# 初回 boot の activation は nix DB のロードより先に走るため、boot 時の結果はアサートしない。
{ pkgs, layatPackage }:
let
  src = pkgs.runCommandLocal "layat-vm-src" { } ''
    mkdir -p "$out/a" "$out/b"
    echo AAA >"$out/a/file"
    echo BBB >"$out/b/file"
  '';
  # 構成 a / b の土台。layat の entry を差し替え、hostName は node 名によらず揃える。
  mkNode = target: subpath: {
    imports = [ ../../modules/nixos.nix ];
    _module.args.layatPackage = layatPackage;
    networking.hostName = "machine";
    system.switch.enable = true;
    # VM のディスクには bootloader を入れられないため、switch での install を止める。
    boot.loader.grub.enable = false;
    layat = {
      enable = true;
      configs.vm.entries.${target} = { inherit src subpath; };
    };
  };
in
pkgs.testers.runNixOSTest {
  name = "layat-nixos-vm";
  nodes = {
    machine =
      { nodes, ... }:
      {
        imports = [ (mkNode "etc/layat-a" "a") ];
        # engine の世代コミット（nix-env）は link-farm が nix DB に登録済みであることを要する。
        virtualisation.additionalPaths = [ nodes.b.system.build.toplevel ];
      };
    b = mkNode "etc/layat-b" "b";
  };
  testScript =
    { nodes, ... }:
    ''
      a = "${nodes.machine.system.build.toplevel}"
      b = "${nodes.b.system.build.toplevel}"
      profiles = "/nix/var/nix/profiles/layat/vm"
      record = "/var/lib/layat/vm/last-applied.json"

      def gens():
          return int(machine.succeed(f"ls {profiles} | grep -c '^profile-[0-9]*-link$' || true").strip())

      # target は配置元の store パスを直接指す symlink になる。
      def assert_placed(target, subpath, content):
          machine.succeed(
              f"test -L {target}",
              f'test "$(readlink {target})" = "${src}/{subpath}"',
              f'test "$(cat {target}/file)" = {content}',
          )

      machine.start()
      machine.wait_for_unit("multi-user.target")

      with subtest("(1) 構成 a の switch で /etc 配下へ配置し、system 基底に世代と記録が立つ"):
          machine.succeed(f"{a}/bin/switch-to-configuration switch")
          machine.succeed(f"test -L {profiles}/profile-1-link", f"test -f {record}")
          assert_placed("/etc/layat-a", "a", "AAA")
          assert gens() == 1, f"世代が 1 件ではありません: {gens()}"

      with subtest("(2) 同一構成の再 activate で世代が増えない"):
          machine.succeed("/run/current-system/activate")
          assert_placed("/etc/layat-a", "a", "AAA")
          assert gens() == 1, f"再 activate が世代を積みました: {gens()}"

      with subtest("(3) 構成 b の switch で target が入れ替わり、旧 target が stale 除去される"):
          machine.succeed(f"{b}/bin/switch-to-configuration switch")
          assert_placed("/etc/layat-b", "b", "BBB")
          machine.fail("test -e /etc/layat-a || test -L /etc/layat-a")
          assert gens() == 2, f"世代が 2 件ではありません: {gens()}"
          machine.succeed("/run/current-system/activate")
          assert gens() == 2, f"b の再 activate が世代を積みました: {gens()}"

      with subtest("(4) 構成 a の switch 再実行で配置が a に収束する"):
          machine.succeed(f"{a}/bin/switch-to-configuration switch")
          assert_placed("/etc/layat-a", "a", "AAA")
          machine.fail("test -e /etc/layat-b || test -L /etc/layat-b")
    '';
}

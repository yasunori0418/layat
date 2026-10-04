# nix-unit: symlink farm のアンカー対象抽出・アンカー名・アンカー配置シェルの生成と、
# それらを mkManifest がビルドスクリプトへ埋める配線をアサートする。
# src は toString が安定する fake な flake-input 相当（`{ outPath = …; }`）を使う。
{ lib, layat }:
let
  fakeSrc = {
    outPath = "/nix/store/00000000000000000000000000000000-fake-src";
  };
  norm = root: entries: layat.normalizeManifest { inherit lib root entries; };

  # store×symlink（採用）/ store×copy（除外）/ out-of-store×symlink（除外）が混在する manifest。
  # 抽出テストと配線テストが同じ入力を見るよう、entries を 1 箇所で宣言する。
  mixedEntries = {
    ".config/copy" = {
      src = fakeSrc;
      method = "copy";
    };
    ".config/out" = {
      src = layat.mkOutOfStoreSymlink "/home/me/dotfiles/x";
    };
    ".config/sym" = {
      src = fakeSrc;
    };
    ".config/sym2" = {
      src = fakeSrc;
    };
  };

  # normalizeManifest は target 辞書順で配列化するため mixed.entries の順は
  # [".config/copy", ".config/out", ".config/sym", ".config/sym2"]。
  mixed = norm layat.projectRoot mixedEntries;

  farm = layat.__internal.farmEntries lib mixed.entries;

  # copy しか無い manifest（アンカー対象が皆無になる入力）。
  copyOnlyEntries = {
    ".config/copy" = {
      src = fakeSrc;
      method = "copy";
    };
  };

  # 配線検証用の fake pkgs。writeText / runCommandLocal を引数を持ち帰る double に差し替え、
  # derivation を組まずにビルドスクリプト本文（`buildCommand`）を純評価で取り出す。
  fakePkgs = {
    inherit lib;
    writeText = name: _text: "/nix/store/fake-${name}";
    runCommandLocal = _name: _attrs: script: { buildCommand = script; };
  };

  buildCommandOf =
    entries:
    (layat.mkManifest {
      pkgs = fakePkgs;
      root = layat.projectRoot;
      inherit entries;
    }).buildCommand;
in
{
  # farmEntries は store×symlink のみを採用し、copy / out-of-store を除外する。
  testFarmEntriesIncludesOnlyStoreSymlink = {
    expr = map (e: e.target) farm;
    expected = [
      ".config/sym"
      ".config/sym2"
    ];
  };

  # store×symlink が皆無なら farmEntries は空（copy / out-of-store だけではアンカーを持たない）。
  testFarmEntriesEmptyWhenNoStoreSymlink = {
    expr =
      layat.__internal.farmEntries lib
        (norm layat.projectRoot {
          ".config/copy" = {
            src = fakeSrc;
            method = "copy";
          };
          ".config/out" = {
            src = layat.mkOutOfStoreSymlink "/home/me/dotfiles/x";
          };
        }).entries;
    expected = [ ];
  };

  # GC アンカー名は target の sha256 短縮 hex（32 文字）。
  testAnchorNameSha256ShortHex = {
    expr = layat.__internal.anchorName lib ".config/sym";
    expected = "029f105e76667554409c2422b0f61f1c";
  };

  # ---- anchorLines の単体（内容の正しさをここで固定する）------------------------------
  # 手組み入力に適用し、リテラルの期待値で押さえる。
  # 1 行の形は `ln -s <escapeShellArg src> "$out/<anchorName target>"`。
  testAnchorLinesSingleEntry = {
    expr = layat.__internal.anchorLines lib [
      {
        src = "/nix/store/00000000000000000000000000000000-fake-src";
        target = ".config/sym";
      }
    ];
    expected = ''ln -s /nix/store/00000000000000000000000000000000-fake-src "$out/029f105e76667554409c2422b0f61f1c"'';
  };

  # 複数エントリは改行連結（末尾に改行は付かない）。target ごとに anchor 名が変わる。
  testAnchorLinesJoinsWithNewline = {
    expr = layat.__internal.anchorLines lib [
      {
        src = "/nix/store/00000000000000000000000000000000-fake-src";
        target = ".config/sym";
      }
      {
        src = "/nix/store/00000000000000000000000000000000-fake-src";
        target = ".config/sym2";
      }
    ];
    expected = ''
      ln -s /nix/store/00000000000000000000000000000000-fake-src "$out/029f105e76667554409c2422b0f61f1c"
      ln -s /nix/store/00000000000000000000000000000000-fake-src "$out/1fa2d3541e7cab32b4961dfbdb6f1095"'';
  };

  # src 側は escapeShellArg を通る。空白・記号を含むパスが shell へ素通りしないこと。
  testAnchorLinesEscapesSrc = {
    expr = layat.__internal.anchorLines lib [
      {
        src = "/nix/store/x y & z";
        target = ".config/sym";
      }
    ];
    expected = ''ln -s '/nix/store/x y & z' "$out/029f105e76667554409c2422b0f61f1c"'';
  };

  # アンカー対象が皆無なら空文字列（埋め込み後の整形がどうなるかはここでは見ない）。
  testAnchorLinesEmptyWhenNoEntries = {
    expr = layat.__internal.anchorLines lib [ ];
    expected = "";
  };

  # ---- farm derivation への配線（ビルドスクリプトに何が埋まるかを見る）--------------------
  # mkManifest が埋めるアンカー行が、生成式へ farm 対象だけを通した結果であることを
  # ビルドスクリプト全体の一致で見る。
  testBuildCommandEmbedsAnchorLinesForFarmEntriesOnly = {
    expr = buildCommandOf mixedEntries;
    expected = ''
      mkdir -p "$out"
      cp /nix/store/fake-manifest.json "$out/manifest.json"
      ${layat.__internal.anchorLines lib farm}
    '';
  };

  # アンカー対象が皆無なら `ln -s` は 1 行も現れない（行の有無だけを見る）。
  testBuildCommandHasNoAnchorLinesWhenNoFarmEntries = {
    expr = lib.filter (l: lib.hasPrefix "ln -s " l) (
      lib.splitString "\n" (buildCommandOf copyOnlyEntries)
    );
    expected = [ ];
  };

  # アンカー対象が皆無でも manifest.json のコピーは行う。
  testBuildCommandStillCopiesManifestWhenNoFarmEntries = {
    expr = lib.filter (l: l != "") (lib.splitString "\n" (buildCommandOf copyOnlyEntries));
    expected = [
      ''mkdir -p "$out"''
      ''cp /nix/store/fake-manifest.json "$out/manifest.json"''
    ];
  };
}

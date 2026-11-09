{
  description = "layat development environment";

  inputs = {
    root.url = "path:../";
    nixpkgs.follows = "root/nixpkgs";
    flake-parts.follows = "root/flake-parts";

    # Claude Code 用スキル集（mattpocock/skills）。project mode の layat apply で配置する。
    matt-skills = {
      url = "github:mattpocock/skills";
      flake = false;
    };

    # outturn エンベロープの E2E 適合検証キット（outturn-validate CLI + id-vectors）。
    outturn = {
      url = "github:yasunori0418/outturn";
      inputs.nixpkgs.follows = "nixpkgs";
      inputs.flake-parts.follows = "flake-parts";
    };

    # 個人 NUR。devShell に sara を載せるために引く。nixpkgs を follows で寄せず、
    # yasunori0418.cachix.org にあるビルド済み store path をそのまま使う。
    nur = {
      url = "github:yasunori0418/nur-packages";
    };
  };

  outputs =
    inputs@{ flake-parts, ... }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "aarch64-darwin"
        "x86_64-darwin"
      ];
    in
    flake-parts.lib.mkFlake { inherit inputs; } {
      inherit systems;
      imports = [
        inputs.root.flakeModules.default
        # layat dogfood config（perSystem.layat.skills）を flake-parts module として切り出す。
        ./layat.nix
      ];
      perSystem =
        { inputs', pkgs, ... }:
        let
          # dev/tests/term-avoid.sh が avoid の綴りの検出に使う。
          textlint = pkgs.textlint.withPackages [ pkgs.textlint-rule-prh ];

          # outturn の Go 参照実装は subPackages 由来の bin/validate を生むため、
          # 規格上の CLI 名 outturn-validate で PATH に載せる薄い wrapper を挟む。
          outturn-validate = pkgs.writeShellScriptBin "outturn-validate" ''
            exec ${inputs'.outturn.packages.validate}/bin/validate "$@"
          '';

          # sara init を包む item 起票ラッパー。
          #
          #   sara-new <型> <slug> <配置ディレクトリ> -- --description <要約> [<sara init のオプション>...]
          #
          # 実体の dev/scripts/sara-new.sh を PATH に載せる薄い wrapper。
          sara-new = pkgs.writeShellApplication {
            name = "sara-new";
            runtimeInputs = [
              inputs'.nur.packages.sara
              # date / mkdir / mv / rm / tr。
              pkgs.coreutils
              # ID 行の抽出に使う。
              pkgs.gnused
              # frontmatter の閉じの直前へ generated を足すのに使う。
              pkgs.gawk
              # 下の text が exec するインタプリタ自身。
              pkgs.bash
            ];
            text = ''
              exec bash ${./scripts/sara-new.sh} "$@"
            '';
          };

          # sara ドキュメントグラフの未カバーを列挙する決定論コマンド。
          #
          #   sara-gap [--json]
          #
          # 実体の dev/scripts/sara-gap.sh を PATH に載せる薄い wrapper。
          sara-gap = pkgs.writeShellApplication {
            name = "sara-gap";
            runtimeInputs = [
              inputs'.nur.packages.sara
              pkgs.jq
              pkgs.coreutils
              # sara.toml のある走査基点をリポジトリルート基準で解決するため。
              pkgs.git
              # 下の text が exec するインタプリタ自身。
              pkgs.bash
            ];
            text = ''
              exec bash ${./scripts/sara-gap.sh} "$@"
            '';
          };
        in
        {
          devShells.default = pkgs.mkShell {
            packages = with pkgs; [
              statix
              nixd
              inputs'.root.formatter
              inputs'.root.packages.layat
              # 設計文書・要求をナレッジグラフとして扱う CLI（NUR 由来）。
              inputs'.nur.packages.sara
              # item 起票ラッパー（sara init + 規約どおりの rename）。
              sara-new
              # グラフ未カバーと到達率の列挙。
              sara-gap
              # CASE frontmatter を読む yq-go（mikefarah/yq v4）。python-yq とは別実装。
              yq-go
              # test-doc-matrix.sh が sara report matrix --format json を整形するのに使う。
              jq
              textlint
              go
              gopls
              # ローカルのカバレッジ計測。func サマリを出し、HTML 表示のコマンドを案内する。
              (writeShellScriptBin "layat-coverage" ''
                set -euo pipefail
                profile="cover.out"
                go test -coverprofile="$profile" ./...
                go tool cover -func="$profile"
                echo "HTML レポート: go tool cover -html=$profile"
              '')
            ];
            shellHook = ''
              export REPO_ROOT=$(git rev-parse --show-superproject-working-tree --show-toplevel)
              # mattpocock/skills と dev/skills/ を .claude/skills/ に dogfood 配置する。競合時は待たず skip する。
              layat apply skills -f "$REPO_ROOT/dev" --no-wait
            '';
          };

          # テストコード ⇔ CASE 対応の契約テスト（dev/tests/test-doc-map.sh）。CI はこの派生ではなく devShells.sara で走らせる。
          # ルート flake の store path を書き込み可能な場所へ複製し、dev/ は dev flake 側から重ねて走らせる。
          checks.test-doc-map =
            pkgs.runCommandLocal "test-doc-map"
              {
                nativeBuildInputs = [
                  pkgs.yq-go
                  pkgs.git
                  pkgs.coreutils
                  pkgs.gnused
                  pkgs.gnugrep
                  pkgs.diffutils
                  pkgs.findutils
                ];
              }
              ''
                cp -r ${inputs.root} repo
                chmod -R u+w repo
                rm -rf repo/dev/scripts repo/dev/tests
                mkdir -p repo/dev
                cp -r ${./scripts} repo/dev/scripts
                cp -r ${./tests} repo/dev/tests
                chmod -R u+w repo/dev
                cd repo
                # テスト側の走査基点はカレントへフォールバックする（git 管理外のため）。
                bash dev/tests/test-doc-map.sh
                touch "$out"
              '';

          # OKF 互換の入れ子キーの形の契約テスト（dev/tests/okf-frontmatter.sh）。CI はこの派生ではなく devShells.sara で走らせる。
          # ルート flake の store path を書き込み可能な場所へ複製し、dev/ は dev flake 側から重ねて走らせる。
          checks.okf-frontmatter =
            pkgs.runCommandLocal "okf-frontmatter"
              {
                nativeBuildInputs = [
                  pkgs.yq-go
                  pkgs.jq
                  pkgs.git
                  pkgs.coreutils
                  pkgs.gawk
                  pkgs.gnugrep
                  pkgs.findutils
                ];
              }
              ''
                cp -r ${inputs.root} repo
                chmod -R u+w repo
                rm -rf repo/dev/scripts repo/dev/tests
                mkdir -p repo/dev
                cp -r ${./scripts} repo/dev/scripts
                cp -r ${./tests} repo/dev/tests
                chmod -R u+w repo/dev
                cd repo
                # テスト側の走査基点はカレントへフォールバックする（git 管理外のため）。
                bash dev/tests/okf-frontmatter.sh
                touch "$out"
              '';

          # 用語索引 ⇔ 用語 item の 1:1 の契約テスト（dev/tests/glossary-index.sh）。CI はこの派生ではなく devShells.sara で走らせる。
          # ルート flake の store path を書き込み可能な場所へ複製し、dev/ は dev flake 側から重ねて走らせる。
          checks.glossary-index =
            pkgs.runCommandLocal "glossary-index"
              {
                nativeBuildInputs = [
                  inputs'.nur.packages.sara
                  pkgs.jq
                  pkgs.git
                  pkgs.coreutils
                  pkgs.gnused
                  pkgs.gnugrep
                  pkgs.gawk
                ];
              }
              ''
                cp -r ${inputs.root} repo
                chmod -R u+w repo
                rm -rf repo/dev/scripts repo/dev/tests
                mkdir -p repo/dev
                cp -r ${./scripts} repo/dev/scripts
                cp -r ${./tests} repo/dev/tests
                chmod -R u+w repo/dev
                cd repo
                # テスト側の走査基点はカレントへフォールバックする（git 管理外のため）。
                bash dev/tests/glossary-index.sh
                touch "$out"
              '';

          # OKF の index.md と生成結果の差分ゼロの契約テスト（dev/tests/okf-index.sh）。CI はこの派生ではなく devShells.sara で走らせる。
          # ルート flake の store path を書き込み可能な場所へ複製し、dev/ は dev flake 側から重ねて走らせる。
          checks.okf-index =
            pkgs.runCommandLocal "okf-index"
              {
                nativeBuildInputs = [
                  inputs'.nur.packages.sara
                  pkgs.jq
                  pkgs.git
                  pkgs.coreutils
                  pkgs.gnused
                  pkgs.gnugrep
                  pkgs.gawk
                  pkgs.diffutils
                  pkgs.findutils
                ];
              }
              ''
                cp -r ${inputs.root} repo
                chmod -R u+w repo
                rm -rf repo/dev/scripts repo/dev/tests
                mkdir -p repo/dev
                cp -r ${./scripts} repo/dev/scripts
                cp -r ${./tests} repo/dev/tests
                chmod -R u+w repo/dev
                cd repo
                # テスト側の走査基点はカレントへフォールバックする（git 管理外のため）。
                bash dev/tests/okf-index.sh
                touch "$out"
              '';

          # avoid の綴りの検出の契約テスト（dev/tests/term-avoid.sh）。CI はこの派生ではなく devShells.sara で走らせる。
          # ルート flake の store path を書き込み可能な場所へ複製し、dev/ は dev flake 側から重ねて走らせる。
          checks.term-avoid =
            pkgs.runCommandLocal "term-avoid"
              {
                nativeBuildInputs = [
                  inputs'.nur.packages.sara
                  textlint
                  pkgs.jq
                  pkgs.git
                  pkgs.coreutils
                  pkgs.findutils
                  pkgs.gnugrep
                ];
              }
              ''
                cp -r ${inputs.root} repo
                chmod -R u+w repo
                rm -rf repo/dev/scripts repo/dev/tests
                mkdir -p repo/dev
                cp -r ${./scripts} repo/dev/scripts
                cp -r ${./tests} repo/dev/tests
                chmod -R u+w repo/dev
                cd repo
                # テスト側の走査基点はカレントへフォールバックする（git 管理外のため）。
                bash dev/tests/term-avoid.sh
                touch "$out"
              '';

          # リンク文字列の裸のタグと name の繰り返し、許可外の ID の契約テスト（dev/tests/link-text.sh）。CI はこの派生ではなく devShells.sara で走らせる。
          # ルート flake の store path を書き込み可能な場所へ複製し、dev/ は dev flake 側から重ねて走らせる。
          checks.link-text =
            pkgs.runCommandLocal "link-text"
              {
                nativeBuildInputs = [
                  pkgs.git
                  pkgs.coreutils
                  pkgs.gawk
                  pkgs.gnugrep
                  pkgs.findutils
                  pkgs.diffutils
                ];
              }
              ''
                cp -r ${inputs.root} repo
                chmod -R u+w repo
                rm -rf repo/dev/scripts repo/dev/tests
                mkdir -p repo/dev
                cp -r ${./scripts} repo/dev/scripts
                cp -r ${./tests} repo/dev/tests
                chmod -R u+w repo/dev
                cd repo
                # テスト側の走査基点はカレントへフォールバックする（git 管理外のため）。
                bash dev/tests/link-text.sh
                touch "$out"
              '';

          # risk の level 導出マトリクス整合の契約テスト（dev/tests/risk-matrix.sh）。CI はこの派生ではなく devShells.sara で走らせる。
          checks.risk-matrix =
            pkgs.runCommandLocal "risk-matrix"
              {
                # サンドボックスには作業ツリーが無いため、risk item の在り処を store path で渡す。
                RISK_DOCS_DIR = ../docs/risks;
                nativeBuildInputs = [
                  pkgs.coreutils
                  pkgs.findutils
                  # 走査基点の解決に使う（無いと `git rev-parse` が command not found になる）。
                  pkgs.git
                  # frontmatter の読み取りに使う（mikefarah/yq v4）。
                  pkgs.yq-go
                  # lib-testdoc.sh が使う（read_tsv のコメント除去・require_yq_go の
                  # yq --version 判定）。
                  pkgs.gnugrep
                ];
              }
              ''
                # テストは lib-testdoc.sh を相対パスで source するため、dev/ の木を作ってから走らせる。
                mkdir -p dev
                cp -r ${./scripts} dev/scripts
                cp -r ${./tests} dev/tests
                chmod -R u+w dev
                bash dev/tests/risk-matrix.sh
                touch "$out"
              '';

          # sara-gap の検出契約を固定するテスト（dev/tests/sara-gap.sh）。CI はこの派生ではなく devShells.sara で走らせる。
          # テストは fixture を相対パスで解決するため、dev/ の木を作ってから走らせる。
          checks.sara-gap =
            pkgs.runCommandLocal "sara-gap-test"
              {
                # テストが fixture へ重ねる実物の docs/model.yaml を store path で渡す。
                SARA_GAP_MODEL_YAML = ../docs/model.yaml;
                nativeBuildInputs = [
                  sara-gap
                  # テスト自身のアサーション用。
                  pkgs.jq
                  pkgs.coreutils
                  pkgs.gnugrep
                  # ルート解決経路（SARA_GAP_ROOT 無し）の検証で一時 repo を git init する。
                  pkgs.git
                ];
              }
              ''
                mkdir -p dev
                cp -r ${./tests} dev/tests
                chmod -R u+w dev
                bash dev/tests/sara-gap.sh
                touch "$out"
              '';

          # sara-new の起票契約を固定するテスト（dev/tests/sara-new.sh）。CI はこの派生ではなく devShells.sara で走らせる。
          checks.sara-new =
            pkgs.runCommandLocal "sara-new-test"
              {
                # テストが fixture リポジトリへ重ねる実物のモデルを store path で渡す。
                SARA_NEW_MODEL_YAML = ../docs/model.yaml;
                nativeBuildInputs = [
                  sara-new
                  # テストが fixture の検証にも実 sara を使う。
                  inputs'.nur.packages.sara
                  pkgs.coreutils
                  pkgs.findutils
                  pkgs.gnugrep
                  # テストが sara-new の出力から id / file を抜くのに使う。
                  pkgs.gnused
                  # 走査基点の解決に使う（無いと `git rev-parse` が command not found になる）。
                  pkgs.git
                ];
              }
              ''
                bash ${./tests/sara-new.sh}
                touch "$out"
              '';

          # CI の sara check と dev/tests/ の各テスト専用シェル。
          # layat のビルドと dogfood の shellHook を伴わない。
          devShells.sara = pkgs.mkShell {
            packages = [
              inputs'.nur.packages.sara
              sara-new
              sara-gap
              # 以下は dev/tests/ の各テストが使う。対応する checks 派生と揃えて明示する。
              pkgs.git
              pkgs.gnused
              pkgs.coreutils
              # dev/tests/sara-new.sh が fixture の残存ファイルを数えるのに使う。
              pkgs.findutils
              pkgs.gnugrep
              # dev/tests/test-doc-map.sh が CASE frontmatter を読む yq-go（mikefarah/yq v4）。
              pkgs.yq-go
              # dev/tests/sara-gap.sh が --json 出力のアサーションに使う。
              pkgs.jq
              # dev/tests/term-avoid.sh が avoid の綴りを検出する。
              textlint
            ];
            env.TERM = "dumb";
          };

          # 非 NixOS E2E ハーネス（tests/e2e/run.sh）専用の最小 CI シェル。
          # nix は ambient のものを使い、TERM=dumb で対話 UI を抑える。
          devShells.ci = pkgs.mkShell {
            packages = with pkgs; [
              inputs'.root.packages.layat
              bash
              git
              jq
              coreutils
              # system mode の e2e が root を user namespace（unshare -Ur）で取る。
              util-linux
              # --json エンベロープの適合検証（schema + lint）。
              outturn-validate
            ];
            env.TERM = "dumb";
            # E2E の id-vectors 整合チェックが参照する適合ベクタ（outturn testdata の正本）。
            env.OUTTURN_ID_VECTORS = "${inputs.outturn}/testdata/v1/id-vectors.json";
          };
        };
    };
}

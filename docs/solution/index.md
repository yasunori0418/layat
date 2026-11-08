# 解決策（`solution`）

* [layat は nix store の物を任意パスへ置く配置プリミティブであり、設定を生成せずユーザーが配置を明示的に握る](./layat-placement-primitive.md) - `fetchFromGitHub` / `fetchGit` でリポジトリは取得できるが、ファイルシステムへの配置は別の話になる。home-manager の `home.file` を使えば配置できるが home-manager への依存が前提になり、シェルスクリプトで `git clone` すればバージョン固定・再現性・Nix との統合が失われる。

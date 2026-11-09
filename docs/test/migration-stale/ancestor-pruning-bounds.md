---
id: "TC-810d661d-6d3d-4199-828f-e44adcebad5a"
type: test_condition
name: "空親ディレクトリの剪定が多階層を辿り、root・symlink・非空で停止すること"
status: stable
mitigates:
  - "RISK-e659f9dd-fbc9-48aa-8ae5-ecc4faf2a85c"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:29:20+09:00"
---
# 空親剪定の境界

target 除去後の親ディレクトリ剪定が、空である限り多階層を遡ること、および停止すべき境界で
止まることを検証する。境界は 3 つある: root 自体（root は消さない）、symlink である祖先
（その先へは進まない）、非空のディレクトリ（foreign な実ファイルを含む場合を含む）。

剪定の失敗が警告に留まり apply 全体を失敗させないこと、および剪定を実行する処理（古くなった配置の除去後・
reset 後）と実行しない処理（配置前除去。直後に同じ場所へ配置し直すため）が区別されていることを
条件に含む。同じディレクトリが二重に剪定報告されないことも含む。

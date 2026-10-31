---
id: "RISK-94be367e-9e2b-43ed-a351-6f16d41113b8"
type: risk
name: "出所の異なる apply が同名 config の記録を共有し、互いの entry を黙って stale 除去し合う"
likelihood: medium
impact: high
level: high
threatens:
  - "REQ-ee5085a5-077b-4f5a-aa6a-841bc5515282"
  - "REQ-9d622539-a1e5-4719-af51-e0cd85b66f35"
---
# RISK-94be367e-9e2b-43ed-a351-6f16d41113b8: 出所の異なる apply が同名 config の記録を共有し、互いの entry を黙って stale 除去し合う

## リスク

home mode の profileDir と last-applied 記録の鍵は config 名だけで決まる。home-manager
モジュール（`apply --all --manifest`・prebuilt 生成器）と standalone の CLI（entrypoint 経由・
nix 生成器）が同じ名前を使うと、片方の apply がもう片方の置いた entry を stale として除去する。

- **黙った除去** — target の衝突ではないので foreign symlink の warning に掛からない。出所の
  比較が効かないと、交互の apply で entry が消えたり戻ったりしても何も出ない
- **版の読み違い** — 版 1 の記録を拒むと既存利用者の apply が止まる。逆に版 1 を出所ありとして
  比べると、版を上げた直後の apply で全員に偽の警告が出る
- **書き直しでの出所の喪失** — reset / rollback が `origin` を落とすと、次の apply で比較が
  効かず、食い違いを見逃す

## 影響

impact を high とするのは、利用者の `$HOME` から設定ファイルの symlink が予告なく消え、どちらの
経路が消したかも分からないため。likelihood を medium とするのは、両経路の併用は移行期や
試用に限られる一方、`default` という名前を両方で使う構成は珍しくないため。likelihood medium ×
impact high から level は high。

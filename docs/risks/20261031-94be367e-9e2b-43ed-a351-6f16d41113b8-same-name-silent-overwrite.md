---
id: "RISK-94be367e-9e2b-43ed-a351-6f16d41113b8"
type: risk
name: "出所の異なる apply が同名 config の記録を共有し、互いの entry を黙って stale 除去し合う"
likelihood: medium
impact: medium
level: medium
threatens:
  - "REQ-ee5085a5-077b-4f5a-aa6a-841bc5515282"
  - "REQ-9d622539-a1e5-4719-af51-e0cd85b66f35"
---
# RISK-94be367e-9e2b-43ed-a351-6f16d41113b8: 出所の異なる apply が同名 config の記録を共有し、互いの entry を黙って stale 除去し合う

## リスク

home mode の profileDir と last-applied 記録の鍵は config 名だけで決まる。home-manager
モジュール（`apply --all --manifest`・prebuilt 生成器）と standalone の CLI（entrypoint 経由・
nix 生成器）が同じ名前を使うと、片方の apply がもう片方の置いた entry を stale として除去する。

- **黙った除去** — target の衝突ではないので foreign symlink の warning に掛からない。出所を
  比べなければ、交互の apply で entry が消えたり戻ったりしても warning は出ない
- **版の読み違い** — 版 1 の記録を拒むと既存利用者の apply が止まる。逆に版 1 の記録を今回の出所と
  比べると、版を上げた直後の apply で全員に誤った警告が出る
- **書き直しでの出所の喪失** — reset / rollback が `origin` を省くと、次の apply は出所を
  比べられず、食い違いを見逃す

## 影響

impact を medium とするのは、除去されるのが stale として扱われる symlink で、消えた entry は
もう片方の経路で apply を再実行すれば戻るため。気付かれないまま続くことは impact では評価せず、
上のリスクとして書いた。「版の読み違い」と「書き直しでの出所の喪失」も、apply が止まるか
警告が過不足するだけで、版や記録を直して再実行すれば戻るので medium を超えない。
likelihood を medium とするのは、出所の比較が engine・生成器・CLI にまたがり、`--dryrun` と
generation skip を含む経路ごとに入れる必要があるため。どれか 1 経路で比較が抜けても、既存の
テストでは見つからない。likelihood medium × impact medium から level は medium。

---
id: "ADR-efa7685c-6fa4-45df-b7a3-8f3f58e4118c"
type: adr
name: "Build の事後条件を世代バックエンドがコミットできる link-farm のパスに置く"
status: 採用
issues:
  - "#406"
origin: "Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0057 §5・ADR-0055 §4 を書き直した"
justifies:
  - "REQ-194e4209-d804-4a4b-a2b8-3d39c6c33729"
references:
  - "ADR-51863e94-6a9d-4029-95e7-a7ad2bfa93c1"
---
# ADR-efa7685c-6fa4-45df-b7a3-8f3f58e4118c: Build の事後条件を世代バックエンドがコミットできる link-farm のパスに置く

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-51863e94-6a9d-4029-95e7-a7ad2bfa93c1
- 起点: Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0057 §5・ADR-0055 §4 を書き直した

## 背景

manifest 生成器の Build は link-farm を build し、pending に張ってそのパスを返す。返したパスは世代として
コミットされる。nix profile への世代のコミット（`nix-env --set`）は store 外のパスを拒否する。

pending に張るのは、build からコミットまでの間に GC が走っても link-farm が dangling にならないようにするため
（→ ADR-51863e94-6a9d-4029-95e7-a7ad2bfa93c1）。

世代操作は世代バックエンドの契約越しになったが、Build の事後条件を「store path」と書くと、生成器の契約が
nix profile に結び付いたまま残る。

## 決定

Build は、選択中の世代バックエンドがコミットできる link-farm のパスを pending に張って返す。nixprofile が
受け入れるのは store path なので、いまは store path を返すことになる。

達成方法は生成器に任せる。非 nix の生成器向けに store へ入れるヘルパーや、store 外のパスを受け入れる実装は
作らない。

生成器の契約は REQ-194e4209-d804-4a4b-a2b8-3d39c6c33729 に置く。

## 根拠

事後条件をバックエンドに委ねる書き方にしておけば、バックエンドが増えても生成器の契約を書き換えずに済む。
いま要るのは nixprofile が受け入れる store path だけで、それ以外のパスを受け入れる実装には使い道が無い。

## 棄却した代替案

### 事後条件を「世代コミット可能な store path」に固定する

世代バックエンドを差し替えられるようになっても、生成器の条件だけが nix profile に縛られたまま残る。

### store 外のパスを受け入れる実装や、store へ入れるヘルパーをいま作る

受け入れる世代バックエンドが無く、使われない実装になる。

---
id: "ADR-0058"
type: adr
name: "last-applied 記録"
status: 採用
issues:
  - "#444"
  - "#406"
origin: "Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。世代管理 CLI への分離計画（Issue #447）への橋"
justifies:
  - "REQ-9d622539-a1e5-4719-af51-e0cd85b66f35"
  - "REQ-1be4d678-959c-44d7-a346-44bfd95af56e"
  - "DSG-3c156bbe-e72b-43dd-8cb0-2cc105314ad5"
revises:
  - "ADR-0002"
  - "ADR-0015"
references:
  - "ADR-0017"
  - "ADR-0020"
  - "ADR-0025"
  - "ADR-0057"
---
# ADR-0058: last-applied 記録

- ステータス: 採用
- 日付: 2026-10-11
- 関連: ADR-0002（stale 除去の baseline は世代由来の store manifest・却下案「store 外の可変 JSON」）, ADR-0015（rollback は baseline = 離れる世代・ポインタ移動は最後）, ADR-0017（世代スキップと lstat ドリフト修復）, ADR-0020（`reset` の FS-only teardown）, ADR-0025（profileDir の鍵と backref `.root`）, ADR-0057（世代バックエンドの契約・終着点で layat 側に残る baseline の取得）
- 改訂対象: ADR-0002 の「可変 JSON ファイルは持たない」「記録は世代に捕捉され rollback で記録も巻き戻る」を、last-applied 記録に限って覆す（stale 除去の baseline を世代の概念から FS の事実へ移す）。ADR-0015 §5 の rollback の baseline（離れる世代 N の manifest）の供給元を、世代リンクから記録優先へ変える（baseline = 離れる世代・ポインタ移動は最後という順序は不変）
- 起点: Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。世代管理 CLI への分離計画（Issue #447）への橋

## 背景

stale 除去・drift 検出・reset・rollback は「前回 layat がこの root に何を置いたか」を baseline として要る。現状その baseline は nix profile の current 世代（`<profileDir>/profile` が指す link-farm の `manifest.json`）から読む。ADR-0002 はこれを「世代由来の store manifest」と定め、store 外の可変 JSON を「GC root を作らず参照中 store パスが消えると rollback が壊れる・standalone では profile と二重」として却下した。

世代管理が layat から出る終着点（→ ADR-0057 北極星）では、profile は layat の所有物でなくなる。baseline を profile から読み続けると、layat は配置のために世代管理 CLI の内部レイアウトへ依存し続ける。「この root に layat が最後に置いたもの」は世代の概念ではなく FS の事実であり、layat 自身が記録すれば世代側に訊く必要が無い。

## 決定

### 1. 記録の置き場と鍵

apply は「最後に配置した manifest」を layat 自前の記録 `<state>/layat/<鍵>/last-applied.json` に書く。鍵は profileDir と同じ規則: home は `<name>` 直、他は `<roothash>/<name>` + backref `.root`。

### 2. 記録の内容

```json
{ "recordVersion": 1, "linkFarm": "<store path>", "appliedAt": "<RFC3339>", "manifest": <manifest.json を verbatim 埋め込み> }
```

manifest を内容ごと持つ。記録の有効性は store の GC と無関係で、`linkFarm` は profile との突合に使う。rollback で書き直すときの `linkFarm` は N-1 の link-farm、reset で書き直すときは保持する（残った entry の出所。記録が無く baseline を profile から取ったときはその current 世代の link-farm）。`appliedAt` はその操作の時刻に更新する。

### 3. baseline の優先順位

1. 記録があれば baseline = 記録の `manifest`
2. 無ければ従来どおり nix profile の current（profile リンク）
3. 両方あって `linkFarm` が profile の指す先と食い違えば記録を採り、warning を 1 行出す
4. どちらも無ければ baseline は空（初回）。「last-applied 記録なし」を warning 1 行として出す（warning は沈黙規律の対象外で stderr へ出る）

対象は apply の stale 除去と drift 修復、reset の撤去対象、rollback の baseline（離れる世代 N）。**generation skip の判定（新 link-farm がコミット済みのものと同一か）は記録ではなく profile の current 世代と比べる**。記録は commit の前に書かれるので、記録と比べると commit 失敗後の同じ設定での再実行が skip に入り、未コミットのまま冪等な再実行で収束する保証が崩れるため。**rollback の target（N-1）は引き続き世代リンクから読む**（記録は「最後に置いたもの」しか持たない）。

### 4. 記録の書き込み

- apply では全 FS 操作の成功直後・世代 commit の前に temp + rename で atomic に書く。記録は FS の事実なので commit の成否と独立。書けなければ apply は失敗し、undo journal が FS を巻き戻す。
- generation skip 経路でも更新する。失敗時の扱いは apply 本経路と同じ（apply は失敗し、undo journal がドリフト修復を巻き戻す）。
- rollback では FS の収束後・profile ポインタ移動の前に N-1 の manifest で書き直す（apply と同じ「FS 成功直後・世代操作の前」）。書けなければ rollback は失敗し、undo journal が FS を巻き戻す。
- reset では実際に撤去した entry を除いた manifest で書き直す（foreign として残した entry は記録に残る）。全撤去なら記録ファイルを削除する。書けなければ reset は失敗する。reset は undo journal を持たないので FS は reset が残した状態のまま。
- 読みは JSON として不正な記録・必須フィールドの欠落や型違い・`recordVersion` が 1 でない記録を error にする（「記録なし」には倒さない。誤った baseline で stale 除去するより止める）。
- `--json` の契約・flock の鍵（profileDir）・prune の対象は変えない。

## 根拠

ADR-0002 の却下理由は、記録に限っては成立しない。

- **GC と無関係**: 却下案は manifest を store パスの参照で持ち、参照先が GC で消えると壊れた。記録は manifest を内容ごと持つので、参照する store パスが消えても baseline は読める。GC root は引き続き世代（profile）が担い、記録は baseline の読み出しだけを担う。記録があることで GC が安全になるわけではない。
- **二重にならない**: 却下理由の「standalone では profile と二重」は profile が layat の所有物である間の話。profile が layat から出れば、記録が唯一の baseline になる。移行期の二重は §3 の優先順位と warning で吸収する。
- **rollback で巻き戻す対象ではない**: 記録は「最後に置いたもの」であって世代ではない。rollback は FS を N-1 に収束させた後に記録を N-1 の manifest で書き直すので、記録が世代に捕捉されている必要が無い。

## 影響

- 実装は Issue #448（`internal/paths` の鍵解決・engine の読み書き・各経路の更新・e2e の検証）。
- ADR-0002 / ADR-0015 に改訂注記を足す。
- 要求: 記録（REQ-9d622539-a1e5-4719-af51-e0cd85b66f35）を起票し、世代機構（REQ-1be4d678-959c-44d7-a346-44bfd95af56e）の「store 外に可変 JSON を持たない」を「last-applied 記録だけが store 外の可変 JSON」へ改訂する。レイアウトは DSG-3c156bbe-e72b-43dd-8cb0-2cc105314ad5。

## 棄却した代替案

- **パイプラインが `--previous <link-farm>` で前回の状態を渡す**: 終着点のトポロジーでは成り立つが、呼び出し側が前回の状態識別子を保持する責務を負う。layat が自分で記録すれば呼び出し側は状態を持たずに済む。
- **layat が世代管理 CLI に訊く**: 配置のために世代側の API と内部レイアウトへ依存し続ける。切り出す目的と矛盾する。
- **記録を世代に捕捉させる（link-farm 内に置く）**: ADR-0002 の現状そのもので、profile が layat から出ると読めなくなる。
- **記録を flock の鍵にする**: flock・prune・`--json` の payload は profileDir 由来の責務として分離計画（#447）の未決に残す。本 ADR の範囲で動かさない。

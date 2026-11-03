---
id: "ADR-02ed93fa-c59f-4790-805d-759501017d2b"
type: adr
name: "method の copy から symlink への変更は移行しない"
status: 採用
issues:
  - "#175"
  - "#172"
origin: "epic #172 の grilling セッション（2026-07-12）で確定。ADR-0047 §2 の copy から symlink への変更を書き直した"
justifies:
  - "REQ-2b48620a-abaa-43df-a106-954bbba3de56"
references:
  - "ADR-c4ad9987-850e-4834-bbea-d8de3b073548"
  - "ADR-c42a51ff-fc51-4d19-b8c6-2b79704dd30f"
  - "ADR-fc28d09c-9d52-4539-8838-e9ae5a414594"
---
# ADR-02ed93fa-c59f-4790-805d-759501017d2b: method の copy から symlink への変更は移行しない

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-c4ad9987-850e-4834-bbea-d8de3b073548, ADR-c42a51ff-fc51-4d19-b8c6-2b79704dd30f, ADR-fc28d09c-9d52-4539-8838-e9ae5a414594, GitHub Issue #175, #172
- 起点: epic #172 の grilling セッション（2026-07-12）で確定。ADR-0047 §2 の copy から symlink への変更を書き直した

## 背景

method の symlink から copy への変更は自動で移行する
（ADR-fc28d09c-9d52-4539-8838-e9ae5a414594）。逆向きの copy から symlink への変更でも、前世代が
置いた copy が target を占めている。こちらも自動で移行するかを決める必要があった。

## 決定

copy から symlink への変更は自動で移行せず、target に既存の実体があるとして conflict のまま
にする。利用者が `--backup` を明示したときだけ、copy を退避して symlink を置ける
（ADR-c42a51ff-fc51-4d19-b8c6-2b79704dd30f）。

判定の規範は REQ-2b48620a-abaa-43df-a106-954bbba3de56 に置く。

## 根拠

copy は place-once で置いた後は利用者が管理するもので（ADR-c4ad9987-850e-4834-bbea-d8de3b073548）、
編集されている可能性がある。symlink にするために黙って消せば、利用者の編集を失う。

向きによって扱いが違うのは、データを失う危険の非対称を映したものである。symlink はデータを
持たないが、copy は利用者の編集を含みうる。

## 棄却した代替案

### copy から symlink への変更も自動で移行する

利用者が編集した copy のデータを黙って失う可能性があり、copy を利用者の管理に委ねる
place-once の考え方と衝突する。

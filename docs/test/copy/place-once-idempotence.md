---
id: "TC-596d697f-4ba6-4ec1-b71e-8b5375806c08"
type: test_condition
name: "copy が target 不在時のみマテリアライズし、既存 target を触らないこと"
status: stable
mitigates:
  - "RISK-01cf9f47-95bd-419a-863d-0d7c1f8188ba"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:37:16+09:00"
---
# TC-596d697f-4ba6-4ec1-b71e-8b5375806c08: place-once の冪等性

target が不在のときにだけコピーが行われ、既に存在するときは内容・属性ともに一切変更されない
ことを検証する。「再 apply しても破損しない」だけでは不足で、配置後にユーザーが編集した内容が
そのまま残ることまでを条件とする。

構造不一致（subpath がディレクトリのとき target が通常ファイル、またはその逆）が利用者に知らせずに
書き込まれることなく conflict として停止することも同じ条件の下に置く。

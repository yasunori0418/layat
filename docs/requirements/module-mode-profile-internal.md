---
id: "REQ-844ee375-919f-4341-81e1-a5f89fd32840"
type: requirement
name: "module 時は rollback を host へ一本化し、layat profile は前進のみで追従する"
status: stable
derives_from:
  - "UC-d39c1994-f9a5-4860-80ba-f6e584adaf14"
specification: |
  Under a module, `layat rollback` SHALL NOT be exposed and user-facing rollback SHALL be
  unified on the host (`home-manager --rollback` and its equivalents). A rollback on the
  host SHALL be followed automatically by re-activating the old config and kicking layat
  again, which converges the filesystem; the layat profile SHALL therefore only ever move
  forward, stacking a new generation holding the old content. What the host is required to
  provide to layat SHALL be no more than the kick on switch, and no plumbing of the host's
  `oldGenPath` SHALL be required.
specification_ja: |
  module では `layat rollback` を公開してはならず、ユーザー向けロールバックは host
  （`home-manager --rollback` 等）に一本化しなければならない。host rollback は旧 config を
  再 activate して layat を再 kick することで FS を自動収束させなければならない。したがって
  layat profile は前進のみ（旧内容の新世代を積む）でなければならない。host から layat へ
  要求するのは switch 時の kick だけとしなければならず、ホストの `oldGenPath` 配管を要求
  してはならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T23:23:46+09:00"
---
# REQ-844ee375-919f-4341-81e1-a5f89fd32840: module 時は rollback を host へ一本化し、layat profile は前進のみで追従する

## 仕様

layat は自前 profile を**持つ**が、module ではそれを前世代マニフェスト + stale 追跡のための
内部機構に留める。ユーザー向けロールバックは host に一本化し、host rollback は旧 config を
再 activate して layat を再 kick することで FS を自動収束させる（layat profile は前進のみ＝
旧内容の新世代を積む）。host から layat へ要求するのは「switch 時の kick」だけで、ホストの
oldGenPath 配管は不要。

> **上は原文の写しで、規範は frontmatter が正**。全モードで自前 profile を持つこと・module では
> それを内部機構に留めること・前世代を layat 自身の profile から読みホストの oldGenPath に
> 依存しないことは [世代は link farm derivation を layat 自前 profile へコミットして積み、前世代 manifest から古くなった配置を除去する](generation-mechanism-link-farm-manifest.md) の担当で、本 item はユーザー向け rollback の一本化先と、
> それが成立する追従の仕組み（前進のみ）を規定する。home mode 側の rollback 手順は
> [rollback は FS を先に収束させてから profile ポインタを最後に移す](rollback-refit-then-pointer.md)、`rollback` が home / system mode 限定であることは [rollback と list-generations は home / system mode 限定にする](rollback-list-generations-home-only.md) の担当。

## 出典

`docs/spec.md`「世代管理仕様」→「モジュール時」節と、「ロールバック」節の module の箇条書き。
「モジュール別動作仕様」→「home-manager モジュール」「NixOS モジュール（将来拡張）」
「nix-darwin モジュール（将来拡張）」各節が述べる同趣旨（layat は自前 profile を内部機構として
持ち、ユーザー向け rollback は host 世代へ一本化する）も本 item の担当で、独立 item を立てず
本 item に畳んだ。

決定の実体は [module と project では世代操作を公開しない](../adr/module-and-project-hide-generation-ops.md)・[1 config = 1 profile を全モードで自前に持つ](../adr/one-config-one-profile-all-modes.md)。

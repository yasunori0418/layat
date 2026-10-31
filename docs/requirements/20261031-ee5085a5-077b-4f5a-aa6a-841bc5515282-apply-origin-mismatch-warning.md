---
id: "REQ-ee5085a5-077b-4f5a-aa6a-841bc5515282"
type: requirement
name: "apply は last-applied 記録に出所を持たせ、記録と今回の出所が食い違えば警告して続行する"
derives_from:
  - "UC-d39c1994-f9a5-4860-80ba-f6e584adaf14"
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
specification: |
  The last-applied record SHALL carry `origin = {generator, entrypoint}` identifying the
  source of the most recent apply that wrote the record (reset and rollback are not applies
  and do not change it). `generator` SHALL be `nix` when the link-farm was built by
  the nix generator and `prebuilt` when it was given through `--manifest`. `entrypoint`
  SHALL be the absolute path of the discovered entrypoint when `generator` is `nix`, and
  SHALL be omitted when `generator` is `prebuilt`. apply SHALL write the record with
  `recordVersion` 2 and the origin of the current run.
  Reading SHALL accept `recordVersion` 1 and 2; any other version and any unknown field
  SHALL remain an error; an `origin` that is neither an object nor `null` SHALL be an error
  (`null` is read as no origin), and inside `origin`, a missing `generator`, a field of the wrong type
  or an unknown field SHALL likewise be an error. Other values of `generator` and
  `entrypoint` SHALL be read as given and used in the comparison. A version 1 record SHALL
  be read as having no origin.
  apply, including `--dryrun` and the generation-skip path, SHALL compare the origin in the
  record with the origin of the current run before placing. When both exist and differ in
  `generator` or `entrypoint`, apply SHALL write one warning line to stderr and SHALL emit
  the subject warning `W_LAYAT_ORIGIN_MISMATCH` with `detail = {recorded, current}` (both
  origin objects) in `subjectResult.warnings`, and SHALL continue; the later apply SHALL
  win as for any other shared profileDir. When the record has no origin, apply SHALL NOT
  compare and SHALL NOT warn.
  When reset or rollback rewrites the record, it SHALL carry over the recorded origin; when
  there is no recorded origin (including when no record exists and the baseline was taken
  from the profile), it SHALL write `recordVersion` 2 with `origin` omitted.
specification_ja: |
  last-applied 記録は、直近に記録を書いた apply の出所（reset と rollback は apply ではなく、
  出所を変えない）を表す
  `origin = {generator, entrypoint}` を持たなければならない。`generator` は、link-farm を nix 生成器が組んだときは `nix`、
  `--manifest` で渡されたときは `prebuilt` としなければならない。`entrypoint` は `generator`
  が `nix` のときは発見した entrypoint の絶対パスとし、`prebuilt` のときは省略しなければ
  ならない。apply は `recordVersion` 2 と今回の実行の origin で記録を書かなければならない。
  読みは `recordVersion` 1 と 2 を受け付けなければならず、それ以外の版と未知のフィールドは
  従来どおり error としなければならない。`origin` がオブジェクトでも `null` でもなければ error
  としなければならない（`null` は origin 無しとして読む）。`origin` の中でも、`generator` の欠落・各フィールドの
  型違い・未知のフィールドは同じく error としなければならない。それ以外の `generator` と `entrypoint` の
  値は読んだまま受け付け、比較に使わなければならない。版 1 の記録は origin 無しとして
  読まなければならない。
  apply（`--dryrun` と generation skip 経路を含む）は、配置の前に記録の origin と今回の
  origin を比べなければならない。両方あって `generator` か `entrypoint` が食い違えば、stderr へ
  warning を 1 行出し、`subjectResult.warnings` に subject 警告 `W_LAYAT_ORIGIN_MISMATCH`
  （`detail = {recorded, current}`。どちらも origin のオブジェクト）を載せて続行しなければ
  ならない。profileDir を共有する他の場合と同じく後勝ちとしなければならない。記録に origin
  が無いときは比べてはならず、警告してはならない。
  reset と rollback が記録を書き直すときは記録済みの origin を引き継がなければならない。
  記録済みの origin が無ければ（記録が無く profile から baseline を取った場合を含む）、
  `origin` を省いて `recordVersion` 2 で書かなければならない。
---
# REQ-ee5085a5-077b-4f5a-aa6a-841bc5515282: apply は last-applied 記録に出所を持たせ、記録と今回の出所が食い違えば警告して続行する

## 仕様

- **出所（origin）**: 直近に記録を書いた apply の出所 `{generator, entrypoint}`。
  reset / rollback は apply ではないので、出所を変えない。`generator` は `nix`（nix 生成器が link-farm を
  組んだ）か `prebuilt`（`--manifest` で渡された）。`entrypoint` は `nix` のとき発見した
  entrypoint の絶対パスで、`prebuilt` では省く。
- **版**: apply は `recordVersion` 2 で書く。読みは 1 と 2 を受け付け、1 は origin 無しとして
  読む。それ以外の版と未知のフィールドは従来どおり error。`origin` がオブジェクトでも `null`
  でもなければ error（`null` は origin 無し）。`origin` の中の `generator` の
  欠落・各フィールドの型違い・未知のフィールドも error。それ以外の値は読んだまま比較に使う。
- **食い違いの警告**: apply（`--dryrun` と generation skip 経路を含む）は配置の前に記録の
  origin と今回の origin を比べる。両方あって食い違えば stderr に 1 行、`--json` では
  `subjectResult.warnings` に `W_LAYAT_ORIGIN_MISMATCH`（`detail = {recorded, current}`）を出して
  続行する。止めず、後勝ちのまま。記録に origin が無ければ比べない。
- **書き直しでの引き継ぎ**: reset / rollback は記録済みの origin を引き継ぐ。無ければ
  （記録が無く profile から baseline を取った場合を含む）`origin` を省いて版 2 で書く。
- **採らないもの**: `status` への origin の露出と、食い違いで止めるモード。

想定する食い違いは、home-manager モジュール（`apply --all --manifest`・prebuilt 生成器）と
standalone の CLI（entrypoint 経由・nix 生成器）が同じ config 名を使い、profileDir と
last-applied 記録を共有する場合。互いの entry を stale として除去し合うこと自体は止めず、
それが起きていることを warning で知らせる。

記録の置き場・残りの内容・書き込みの規則は REQ-9d622539-a1e5-4719-af51-e0cd85b66f35、
警告コードの一覧と payload 上の置き場は REQ-2ea19863-eaa2-466b-b1ed-3f56f6417c62、
レイアウトは DSG-3c156bbe-e72b-43dd-8cb0-2cc105314ad5 の担当。

## 出典

ADR-0070「記録の出所と同名 config の警告」§1〜§4・§6。

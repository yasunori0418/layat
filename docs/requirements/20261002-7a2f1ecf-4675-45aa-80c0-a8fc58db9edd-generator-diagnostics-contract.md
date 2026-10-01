---
id: "REQ-7a2f1ecf-4675-45aa-80c0-a8fc58db9edd"
type: requirement
name: "生成器は診断を CLI が渡す writer へ素通しし、失敗を構造化して返し、表示は CLI が決める"
derives_from:
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  A manifest generator SHALL write its diagnostics to the `io.Writer` passed by the CLI and
  SHALL NOT write to stdout. On success, the diagnostics of both the evaluation and the
  build paths SHALL be passed through to that writer rather than discarded; the
  silent-on-success discipline SHALL NOT be a reason to suppress them. A failure SHALL be
  returned as `generator.Error` carrying Generator, Stage (`discover`, `roots` or
  `build`; a DryBuild failure is `build`), Kind (`PrerequisiteMissing`, `NotFound` or `Failed`), Message, Guidance and
  the captured Stderr. Generator-specific string matching and guidance text SHALL stay
  inside that generator's implementation. The CLI SHALL decide the presentation: the
  human-readable form SHALL show the summary and the guidance without repeating the raw
  diagnostics, and the `--json` form SHALL include both the summary and the captured
  diagnostics in `errors[].message`. Under `--debug`, the generator SHALL write each
  internal command it runs, with its name, to the writer, and the CLI SHALL include the
  generator name and the failed internal command in the summary of a failure. Failures at
  the `roots` and `build` stages SHALL be classified as `E_LAYAT_BUILD` regardless of
  Kind; failures at the `discover` stage and failures to read the prebuilt input SHALL
  keep the classification they had before the generator contract. The envelope SHALL NOT
  name the generator.
specification_ja: |
  manifest 生成器は診断を CLI から渡された `io.Writer` へ書かなければならず、stdout へ
  書いてはならない。成功時も eval・build の両経路の診断を捨てずにその writer へ素通し
  しなければならず、成功時沈黙の規律をその抑制の根拠にしてはならない。失敗は
  Generator・Stage（`discover` / `roots` / `build`。DryBuild の失敗は `build`）・Kind（`PrerequisiteMissing` /
  `NotFound` / `Failed`）・Message・Guidance・キャプチャした Stderr を持つ
  `generator.Error` で返さなければならない。生成器固有の文字列判定と案内文はその生成器の
  実装の中に閉じなければならない。表示は CLI が決めなければならない: 人間向けは要約と
  案内を示し生の診断を再掲せず、`--json` は `errors[].message` に要約とキャプチャした
  診断の両方を含めなければならない。`--debug` のとき、生成器は実行する内部コマンドを
  生成器名付きで writer へ書き、CLI は失敗の要約に生成器名と失敗した内部コマンドを
  含めなければならない。Stage `roots` / `build` の失敗は Kind によらず `E_LAYAT_BUILD` に
  分類しなければならず、Stage `discover` の失敗と prebuilt の入力を読めない失敗は
  生成器契約の前の分類を変えてはならない。エンベロープに生成器名を出してはならない。
---
# REQ-7a2f1ecf-4675-45aa-80c0-a8fc58db9edd: 生成器は診断を CLI が渡す writer へ素通しし、失敗を構造化して返し、表示は CLI が決める

## 仕様

| 担い手 | 責務 |
|---|---|
| 生成器 | 診断を渡された writer へ書く（成功時も素通し・stdout には書かない）。`--debug` のとき実行する内部コマンドを生成器名付きで writer へ逐次書く。失敗を `generator.Error` に分類する。固有の文字列判定と案内文を持つ |
| CLI | 表示を決める。人間向け = 要約 + 案内、`--json` = `errors[].message` に要約 + キャプチャ、`--debug` = 失敗の要約に生成器名と失敗した内部コマンドを含める |

`generator.Error` の `Kind`:

- `PrerequisiteMissing` — 生成器の前提条件が満たされていない（nix 生成器では
  experimental-features の未有効）
- `NotFound` — config 名が entrypoint に無い
- `Failed` — それ以外

`Stage` は `discover` / `roots` / `build` の 3 値で、DryBuild の失敗も `build`。

`Kind` は `--json` のコードに影響させない（`NotFound` も `E_NOTFOUND` に写さない）。
`E_LAYAT_BUILD` に分類するのは Stage `roots` / `build` の失敗で、Stage `discover` の失敗
（entrypoint の不在・`-f` のパスの不在）と prebuilt の入力を読めない失敗は従来の分類
（`E_LAYAT_FAILED` / `E_NOTFOUND` 等）のまま。

nix 生成器の前提条件と案内の中身は REQ-f9920c87-8551-4aa3-bf03-26fdf4191ed6、`--debug` の分離そのものは
REQ-0a123b89-0399-4f76-b988-56a5f7e0becf、`--json` のエラーの置き場（トップ `errors[]` / `results[].errors[]`）は
REQ-9341fa5d-836e-4023-af53-cc7d273438d1 の担当。

## 出典

ADR-0055「manifest 生成器の契約」§6〜§8。起点は Issue #396（成功時に評価器の stderr を捨てる
現状と nix 固有のエラー分類）。

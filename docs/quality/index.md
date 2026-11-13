# 品質（`quality`）

* [設計判断は ADR に記録し、改訂の書き戻しと item への接続を同じ変更の中で完了させる](./adr-decisions-recorded-and-connected.md) - 決定を記録しない、あるいは記録したが後続の改訂で一部だけが古くなった ADR を有効なまま残す状態は、その ADR だけを読んだ人に「上書きされた内容」を最新仕様と誤読させる。本 item はその防止を規範として固定する。
* [ADR は 1 本に 1 つの決定を 4 節で書き、justifies は規範の形成に寄与した item へ張る](./adr-one-decision-four-sections-justifies-by-contribution.md) - ADR は単独で覆せる決定ごとに 1 本を 4 節で書き、justifies は規範の形成に寄与した item へ張る
* [自動化が取り込む実行物は不変な識別子で固定し、権限は最小に絞り、不正入力では成果物を作る前に失敗する](./automation-supply-chain-safety.md) - CI と自動化は、このリポジトリの変更なしに振る舞いが変わり得る面を 3 つ持つ。外部から取り込む実行物・与えた権限・受け取る入力である。3 つとも、問題が起きたときに影響が及ぶ先が「マージ前の検証」と「リリースされる成果物」なので、規範として固定する。
* [コメントと利用者向け文字列は現在の動作だけを書き、決定の記録を参照しない](./comments-state-current-behaviour-only.md) - コメントは 3 行以内で現在の動作だけを書き、経緯と ADR・Issue・PR への参照を書かない。利用者向け文字列に ADR 番号を出さない
* [マージ前の自動検証を必須にし、プラットフォーム差が出る層は全プラットフォームで通す](./cross-platform-verification-before-merge.md) - layat は複数の OS / system で動くことを主張するツールであり、その主張は「手元の 1 環境で通った」では裏づけられない。マージ前の自動検証を必須にし、プラットフォーム差が影響する層は全プラットフォームで通す。
* [欠陥はトラッカーが持ちドキュメントグラフは規範のみを持つ。起票は分類語彙を経由する](./defects-in-tracker-graph-holds-norms.md) - グラフが持つのは「そうあるべき姿」であり、欠陥は「起きた出来事」になる。両者を同じ場所に置くと、規範の集合に時系列の記録が紛れ、規範だけを読みたい参照が汚れる。所在をトラッカーとグラフで分ける。
* [ドキュメントグラフのトレーサビリティは機械的に検証する](./document-graph-mechanically-verified.md) - 文書間の参照が指す先を失う、というずれ方は「気づいた人が直す」運用では検出できない。存在しない ID への参照・ID の重複・循環は、人が読んで気づく類の誤りではないため、機械検証に置く。
* [完成の定義は単一の機械パース可能な文書で持ち、項目数を上限で縛る](./dod-single-machine-parsable-document.md) - 完成の定義が複数箇所に散ると、どれを満たせば完成なのかが読み手ごとに変わる。単一の文書を固定パスに置き、そこだけを見れば判定できる状態を保つ。
* [コード整形と静的解析を自動検証に載せ、同じ判定を手元でも得られるようにする](./formatting-and-static-analysis-automated.md) - 整形の揺れと静的解析が見つける類の誤りは、レビューで指摘できるが指摘に人の注意を払わせるだけの価値が無い。マージ前の自動検証の側へ回して、人のレビューを意味判断に集中させる。
* [フル ID は frontmatter だけに書き、散文では item にリンクで触れる](./full-ids-in-frontmatter-links-in-prose.md) - ID はフル UUIDv4 で frontmatter だけに書き、ファイル名は slug とし、散文では name を文字列にしたリンクで item に触れる
* [用語の正名は用語 item が固定し、執筆はそれに従う](./glossary-fixes-canonical-terms.md) - 同じ概念が箇所ごとに違う語で呼ばれると、読み手は同じものを指しているのか判断できず、検索も当たらなくなる。正名を 1 箇所で固定し、執筆はそこに従う。
* [level が high の risk は人が確かめた記録を持つ](./high-risk-verified-by-human.md) - level が high の risk は、generated.at 以降に人が内容を確かめた verified を持つ
* [傾向の計測は報告に留め、マージの条件にしない](./measurement-reports-without-gating.md) - [マージ前の自動検証を必須にし、プラットフォーム差が出る層は全プラットフォームで通す](cross-platform-verification-before-merge.md) が必須にする側の規範であるのに対し、本 item は**必須化しない側の境界**を定める。両方が無いと、必須化の規範が「検証と名の付くものはすべて塞ぐ」と読めてしまう。
* [規範の型は、それが誰を縛るかで決める](./norm-type-by-who-it-binds.md) - 規範の型は、それが誰を縛るかで requirement・quality・test_plan から選ぶ
* [規範は 1 ファイル 1 item のグラフが持ち、item の一覧は生成した index.md だけが持つ](./normative-items-placement-and-authoring.md) - 規範を手書きの文書と item の両方に書くと二重管理になり、改訂で片方だけが古くなる。規範の所在をグラフ側に一本化し、型ごとの item の一覧はグラフから生成した index.md が持つ。各ディレクトリの手書きの README.md は、item から導けない全体像の散文だけを持つ。この分担は文書の分量ではなく**どちらを正とするか**の宣言であり、README.md へ仕様や item の一覧を書き足すことは規範の複製にあたる。
* [最新 main のビルド成果物を再ビルドなしに消費できる状態を保つ](./prebuilt-artifact-availability.md) - layat は flake input として消費されるツールであり、消費側の CI やローカル環境が layat を使うたびにソースからビルドするのは実用上のコストになる。最新 main の成果物を公開キャッシュから取得できる状態を保つ。
* [公開ドキュメントは英語を canonical とし、日本語版を対で保守する](./public-docs-english-canonical.md) - 公開されたリポジトリの読者は日本語話者に限らない。利用者に向けた面は英語を canonical に置き、日本語版はその対として保守する。どちらが原本かを固定しないと、片方だけ更新されたときにどちらが正しいかを読者が判断できなくなる。
* [リファレンスドキュメントは記述対象のソースから生成し、生成物を持たない](./reference-docs-generated-from-source.md) - 記述対象を変えたのに記述を直し忘れる、というずれ方は「気づいた人が直す」運用では検出できない。ソースの doc-comment から毎回生成し、生成物をコミットしないことで、乖離が起こり得る状態そのものを作らない。
* [リリースはバージョンの一次情報の変更で駆動し、手作業の工程を挟まない](./release-driven-by-source-of-truth.md) - バージョンは flake の `version` と Go バイナリに埋め込む文字列の両方が述べるが、それぞれを手で書き換える運用は片方だけが古くなる。一次情報を 1 箇所に置き、他はそこから導出する。
* [risk はまず requirement に張り、設計を差し替えると消える懸念だけを design に張る](./risk-attaches-to-requirement-by-default.md) - risk はまず requirement に張り、設計を差し替えると消える懸念だけを design に張る
* [risk の level は likelihood と impact から導き、手で決めない](./risk-level-derived-from-likelihood-and-impact.md) - risk の likelihood と impact を決まった尺度で採点し、level はその 2 つから導く
* [specification は SHALL 系で書き、日本語の規範文は同じ強度の規範助動詞で終える](./specification-shall-family-and-matching-auxiliaries.md) - specification は SHALL 系で書き、specification_ja は同じ強度の規範助動詞で終える
* [上流より古い item と generated の更新漏れは機械的に列挙する](./stale-items-enumerated-mechanically.md) - 上流より基準時刻が古い下流の item と、本文を変えたのに generated を更新していない item を機械的に列挙し、結果は報告に留める
* [テスト系 item は、テスト資産・検証テーマ・脅威の 3 段の粒度で起こす](./test-items-three-tier-granularity.md) - テスト系 item は CASE をテスト資産・TC を検証テーマ・RISK を脅威の単位で起こす

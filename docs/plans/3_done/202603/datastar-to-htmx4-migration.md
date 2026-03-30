# Datastar から htmx 4 への移行 作業計画書

<!--
このテンプレートの使い方:
1. このファイルを `docs/plans/2_todo/` ディレクトリにコピー
   例: cp docs/plans/template.md docs/plans/2_todo/new-feature.md
2. [機能名] などのプレースホルダーを実際の内容に置き換え
3. 各セクションのガイドラインに従って記述
4. コメント（ `\<!-- ... --\>` ）はガイドラインとして残してください

**作業計画書の性質**:
- 作業計画書は「何をどう変えるか」という変更内容を記述するドキュメントです
- 新しい機能の場合は、概要・要件・設計もこのドキュメントに記述します
- 現在のシステムの状態は `docs/specs/` の仕様書に記述されています
- タスク完了後は、仕様書を新しい状態に更新してください（設計判断や採用しなかった方針も含める）

**仕様書との関係**:
- 新しい機能の場合: タスク完了後に `docs/specs/` に仕様書を作成する
- 既存機能の変更の場合: 「仕様書」セクションに対応する仕様書へのリンクを記載し、タスク完了後に仕様書を更新する

**公開時の注意事項**:
- 開発用ドメイン名を記載する場合は `example.dev` を使用してください（実際のドメイン名は記載しない）
- 環境変数の値はサンプル値のみ記載し、実際の値は含めないでください
-->

## 仕様書

<!--
- 既存機能を変更する場合: 変更対象の仕様書へのリンクを記載してください
- 新しい機能の場合: タスク完了後に作成予定の仕様書のパスを記載してください
-->

- 該当なし（インフラ・ライブラリの変更のため、機能仕様への影響なし）

## 概要

<!--
ガイドライン:
- この機能が「何であるか」「なぜ必要か」を簡潔に説明
- 2-3段落程度で簡潔に
- 既存機能の変更の場合は、変更の背景と目的を記述
-->

Go 版 Annict のハイパーメディアフレームワークを Datastar v1.0.0-RC.6（CDN 経由）から htmx 4 に移行する。

移行の動機:

- **コミュニティの規模**: htmx のほうがコミュニティが大きく、情報や事例が豊富
- **過剰な依存**: 現在の Datastar の使用箇所はフォーム二重送信防止（`$isSubmitting` シグナル）のみであり、Datastar のリアクティブシグナル機能は過剰
- **依存関係のシンプルさ**: htmx 単体で現在の全ユースケースをカバーできる
- **将来の拡張性**: htmx はフラグメント更新やOOBスワップなどサーバードリブン UI の標準的な機能を備えており、将来的にページネーションやリアルタイム更新が必要になった場合にも対応できる

現在の Annict Go 版では Datastar を**クライアントサイドのみ**で使用しており、Go SDK（`datastar-go`）への依存はない。そのため移行の影響範囲はテンプレートファイルと JS 読み込み部分に限定される。

## 要件

<!--
ガイドライン:
- 機能要件: 「何ができるべきか」を記述
- 非機能要件: 「どのように動くべきか」を必要に応じて記述
-->

### 機能要件

<!--
「ユーザーは〇〇できる」「システムは〇〇する」という形式で記述
箇条書きで簡潔に
-->

- 既存の全 Datastar ベースの UI インタラクションが htmx 4 で同等に動作する
  - フォーム二重送信防止（9 ファイル、合計 13 フォーム）
- Datastar の依存を完全に除去する（CDN の script タグ）

### 非機能要件

<!--
必要に応じて以下のような項目を追加してください：
- セキュリティ（認証、認可、暗号化、監査ログなど）
- パフォーマンス（応答時間、スループット、リソース使用量など）
- ユーザビリティ（UX）（使いやすさ、わかりやすさ、アクセシビリティなど）
- 可用性・信頼性（稼働率、障害時の挙動、エラーハンドリングなど）
- 保守性（テストのしやすさ、コードの読みやすさ、ドキュメントなど）

不要な場合はこのセクション全体を削除してください。
-->

- ユーザーから見た挙動は変更前と同一であること（リグレッションなし）
- CSRF 保護が引き続き機能すること

## 実装ガイドラインの参照

<!--
**重要**: 作業計画書を作成する前に、対象プラットフォームのガイドラインを必ず確認してください。
特に以下の点に注意してください：
- ディレクトリ構造・ファイル名の命名規則
- コーディング規約
- アーキテクチャパターン

ガイドラインに沿わない設計は、実装時にそのまま実装されてしまうため、
作業計画書作成の段階でガイドラインに準拠していることを確認してください。
-->

### Go 版の実装の場合

以下のガイドラインに従って設計・実装を行ってください：

- [@go/CLAUDE.md](/workspace/go/CLAUDE.md) - 全体的なコーディング規約
- [@go/docs/architecture-guide.md](/workspace/go/docs/architecture-guide.md) - アーキテクチャガイド
- [@go/docs/handler-guide.md](/workspace/go/docs/handler-guide.md) - HTTP ハンドラーガイドライン（**ファイル名は標準の 9 種類のみ**）
- [@go/docs/i18n-guide.md](/workspace/go/docs/i18n-guide.md) - 国際化ガイド
- [@go/docs/security-guide.md](/workspace/go/docs/security-guide.md) - セキュリティガイドライン
- [@go/docs/templ-guide.md](/workspace/go/docs/templ-guide.md) - templ テンプレートガイド
- [@go/docs/validation-guide.md](/workspace/go/docs/validation-guide.md) - バリデーションガイド

## 設計

<!--
ガイドライン:
- 技術的な実装の設計を記述
- 必要に応じて以下のようなサブセクションを追加してください：
  - データベース設計（テーブル定義、インデックス、制約など）
  - API設計（エンドポイント、リクエスト/レスポンス形式など）
  - UI設計（画面構成、ユーザーフローなど）
  - セキュリティ設計（認証・認可、トークン管理など）
  - コード設計（パッケージ構成、主要な構造体など）

**重要: 設計は実装中に更新する**:
- 作業計画書内の設計は初期の方針であり、完璧ではない
- 実装中により良いアプローチが見つかった場合は、設計を積極的に更新する
- 設計に固執して実装の質を下げるよりも、実装で得た知見を設計に反映する方が重要
- 変更した場合は「採用しなかった方針」セクションに変更前の方針と変更理由を記録する
-->

### 移行対象ファイルの一覧

#### templ テンプレート（Datastar 属性使用: 9 ファイル）

| ファイル                                               | フォーム数 | 使用パターン             | 移行方針                |
| ------------------------------------------------------ | ---------- | ------------------------ | ----------------------- |
| `internal/templates/pages/sign_in/new.templ`           | 1          | `$isSubmitting` シグナル | `hx-on:submit` で無効化 |
| `internal/templates/pages/sign_in_code/show.templ`     | 2          | `$isSubmitting` シグナル | `hx-on:submit` で無効化 |
| `internal/templates/pages/sign_in_password/show.templ` | 1          | `$isSubmitting` シグナル | `hx-on:submit` で無効化 |
| `internal/templates/pages/sign_up/new.templ`           | 1          | `$isSubmitting` シグナル | `hx-on:submit` で無効化 |
| `internal/templates/pages/sign_up_code/new.templ`      | 2          | `$isSubmitting` シグナル | `hx-on:submit` で無効化 |
| `internal/templates/pages/sign_up_username/new.templ`  | 1          | `$isSubmitting` シグナル | `hx-on:submit` で無効化 |
| `internal/templates/pages/password/edit.templ`         | 1          | `$isSubmitting` シグナル | `hx-on:submit` で無効化 |
| `internal/templates/pages/password/reset.templ`        | 1          | `$isSubmitting` シグナル | `hx-on:submit` で無効化 |
| `internal/templates/pages/supporters/show.templ`       | 3          | `$isSubmitting` シグナル | `hx-on:submit` で無効化 |

#### その他（1 ファイル）

| ファイル                                   | 用途                 | 移行方針           |
| ------------------------------------------ | -------------------- | ------------------ |
| `internal/templates/components/head.templ` | Datastar JS 読み込み | htmx JS に差し替え |

### パターン: フォーム二重送信防止（9 ファイル・13 フォーム）

Annict では Datastar の使用パターンがフォーム二重送信防止のみであるため、移行パターンは 1 種類のみ。

**現在（Datastar）**:

```templ
<form data-on:submit__passive="$isSubmitting = true" method="POST" action="/path">
  <button data-attr:disabled="$isSubmitting == true" type="submit">送信</button>
</form>
```

Datastar の `$isSubmitting` リアクティブシグナルでフォーム送信状態を管理し、送信ボタンを自動的に disabled にしている。

**移行後（htmx 4）**:

```templ
<form hx-on:submit="disableSubmitButtons(this)" method="POST" action="/path">
  <button type="submit">送信</button>
</form>
```

`disableSubmitButtons` は `web/main.js` に定義するグローバル関数で、フォーム内の全送信ボタンを disabled にする。htmx 4 の `hx-on:submit` は通常の DOM イベントリスナーとして動作するため、htmx 経由でないフォーム送信でも機能する。通常の HTML フォーム送信（`method="POST" action="/path"`）を維持したまま、テンプレートを見るだけで送信時の挙動がわかる。

### JS の差し替え

- `head.templ` から Datastar CDN の `<script>` タグを削除
- htmx 4 の JS ファイルを `go/static/js/vendor/` に配置し、`head.templ` で読み込み
- `web/main.js` に `disableSubmitButtons` 関数を追加

### CSRF 対策

フォーム送信方式は変更しない（通常の HTML フォーム送信を維持）ため、既存の CSRF ミドルウェアがそのまま機能する。

## 採用しなかった方針

<!--
ガイドライン:
- 検討したが採用しなかった設計や機能を、理由とともに記述
- 将来の開発者が同じ検討を繰り返さないための判断記録
- タスク完了後、この内容は `specs/` の仕様書にも転記する
- 該当がない場合は「なし」と記載
-->

### htmx 2.x を使用する

htmx 2.x は安定版だが、fetch API ベースの新しいアーキテクチャを持つ htmx 4 のほうが将来的に有利。htmx 4 はまだ alpha だが、Annict の開発ペースでは安定版リリースまでに十分な時間がある。Wikino でも htmx 4 への移行を完了しており、スキルファイルを共有できる。

### フォーム送信を `hx-post` に変換して `hx-disable` を使用する

htmx 4 の `hx-disable` はリクエスト中に要素を自動で disabled にするが、htmx 経由（`hx-post` 等）でフォーム送信される場合のみ機能する。フォームを `hx-post` 化すれば `hx-disable` が使えるが、以下の理由で不採用とした：

- ハンドラー側の変更が必要（成功時は `HX-Redirect` ヘッダー、エラー時はフォーム HTML の再レンダリング）で、テンプレート変更だけでは完結しない
- 対象ハンドラーが多岐にわたり（sign_in, sign_in_code, sign_in_password, sign_up, sign_up_code, sign_up_username, password, supporters）、影響範囲が大きい

代わりに `hx-on:submit` による DOM イベントリスナー方式を採用した。通常の HTML フォーム送信を維持したまま、テンプレート変更のみで二重送信防止を実現できる。Wikino でも同じ方式を採用しており、実績がある。

### Datastar を使い続ける

現在のフォーム二重送信防止のみの用途には Datastar は過剰であり、CDN 依存（ローカルにベンダーファイルがない）という点でもリスクがある。htmx に統一することで、プロジェクト間（Annict/Wikino）での知識共有も容易になる。

## タスクリスト

<!--
ガイドライン:
- フェーズごとに段階的な実装計画を記述
- チェックボックスで進捗を管理
- **重要**: 1タスク = 1 Pull Request の粒度で作成してください
- **重要**: 各タスクには想定ファイル数と想定行数を明記してください（PRサイズの見積もりのため）
- 想定ファイル数は「実装」と「テスト」に分けて記載してください
- 想定行数も「実装」と「テスト」に分けて記載してください
- 依存関係を明確に
- Pull Requestのガイドラインは CLAUDE.md を参照（変更ファイル数20以下、変更行数300行以下）

フィーチャーフラグ:
- 新機能の開発ではフィーチャーフラグによる制御を検討してください
- フラグで制御することで、実装途中でも develop ブランチにマージできます
- フラグのセットアップ（フラグ名の定義、ルーティングパターンの追加など）をフェーズ1に含めてください
- 機能が安定した後のフラグ削除タスクも計画に含めてください
- 詳細は CLAUDE.md の「フィーチャーフラグによる開発」セクションを参照

タスク番号の付け方:
- 各タスクには階層的な番号を付与します（例: 1-1, 1-2, 2-1, 2-2）
- フォーマット: **フェーズ番号-タスク番号**: タスク名
- **フェーズ番号は半角英数字とハイフンのみで表記**してください（ブランチ名に使用するため）
  - 例: フェーズ 1, フェーズ 2, フェーズ 5a（フェーズ 5 と 6 の間に追加する場合）
  - NG: フェーズ 5.5（ドットは使用不可）
- タスクの前に別のタスクを追加する場合は、サブ番号を使用します
  - 例: タスク 2-1 の前にタスクを追加する場合 → 2-0
  - 例: タスク 2-0 の前にタスクを追加する場合 → 2-0-1
- この番号はブランチ名の一部として使用されます（例: feature-1-1, feature-2-0）

プラットフォームプレフィックス:
- Go版またはRails版の修正を行うタスクには、タスク名の先頭にプラットフォームを示すプレフィックスを付けてください
- フォーマット: **フェーズ番号-タスク番号**: [Go] タスク名 または **フェーズ番号-タスク番号**: [Rails] タスク名
- Go版とRails版の両方を修正する場合は、別々のタスクに分けてください
- 例:
  - `- [ ] **1-1**: [Go] マイグレーション作成`
  - `- [ ] **1-2**: [Rails] モデルへのコールバック追加`
-->

### フェーズ 1: 準備

- [x] **1-1**: [Go] htmx JS ファイルの配置と読み込み設定
  - htmx 4 の JS ファイルを `go/static/js/vendor/` に配置
  - `head.templ` に htmx の `<script>` タグを追加（Datastar と並行して読み込み、段階的移行を可能にする）
  - `web/main.js` に `disableSubmitButtons` グローバル関数を追加
  - **想定ファイル数**: 約 3 ファイル（実装 3 + テスト 0）
  - **想定行数**: 約 20 行（実装 20 行 + テスト 0 行）

### フェーズ 2: フォーム二重送信防止の移行

<!--
$isSubmittingシグナルをhtmxのフォーム送信制御に置き換える。
9ファイル・13フォームを移行。
-->

- [x] **2-1**: [Go] 認証フォームの二重送信防止の移行（sign_in, sign_in_code, sign_in_password）
  - `/htmx4` スキルを使用して実装する
  - 以下 3 ファイルの `data-on:submit__passive` + `data-attr:disabled` を `hx-on:submit` による送信ボタン無効化に置き換え
    - `internal/templates/pages/sign_in/new.templ`（1 フォーム）
    - `internal/templates/pages/sign_in_code/show.templ`（2 フォーム）
    - `internal/templates/pages/sign_in_password/show.templ`（1 フォーム）
  - **想定ファイル数**: 約 3 ファイル（実装 3 + テスト 0）
  - **想定行数**: 約 30 行（実装 30 行 + テスト 0 行）

- [x] **2-2**: [Go] ユーザー登録フォームの二重送信防止の移行（sign_up, sign_up_code, sign_up_username）
  - `/htmx4` スキルを使用して実装する
  - 以下 3 ファイルの `data-on:submit__passive` + `data-attr:disabled` を `hx-on:submit` による送信ボタン無効化に置き換え
    - `internal/templates/pages/sign_up/new.templ`（1 フォーム）
    - `internal/templates/pages/sign_up_code/new.templ`（2 フォーム）
    - `internal/templates/pages/sign_up_username/new.templ`（1 フォーム）
  - **想定ファイル数**: 約 3 ファイル（実装 3 + テスト 0）
  - **想定行数**: 約 30 行（実装 30 行 + テスト 0 行）

- [x] **2-3**: [Go] パスワード・サポーターフォームの二重送信防止の移行
  - `/htmx4` スキルを使用して実装する
  - 以下 3 ファイルの `data-on:submit__passive` + `data-attr:disabled` を `hx-on:submit` による送信ボタン無効化に置き換え
    - `internal/templates/pages/password/edit.templ`（1 フォーム）
    - `internal/templates/pages/password/reset.templ`（1 フォーム）
    - `internal/templates/pages/supporters/show.templ`（3 フォーム）
  - **想定ファイル数**: 約 3 ファイル（実装 3 + テスト 0）
  - **想定行数**: 約 30 行（実装 30 行 + テスト 0 行）

### フェーズ 3: クリーンアップ

- [x] **3-1**: [Go] Datastar 依存の完全除去
  - `/htmx4` スキルを使用して実装する
  - `head.templ` から Datastar CDN の `<script>` タグを削除
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 5 行（実装 5 行 + テスト 0 行）

### 実装しない機能（スコープ外）

<!--
今回は実装しないが、将来的に検討する機能を明記
-->

以下の機能は今回の実装では**実装しません**：

- **フォーム送信の htmx 化（`hx-post` 等）**: フォーム二重送信防止は `hx-on:submit` による DOM イベントリスナー方式を採用したため、フォーム送信の htmx 化は不要。将来的にフォームのプログレッシブエンハンスメントが必要になった場合に検討する
- **htmx 4 のフラグメント更新・OOB スワップの導入**: 現時点では SSE やフラグメント更新が必要なユースケースがないため見送り。将来的にページネーションやリアルタイム更新が必要になった場合に検討する

## 参考資料

<!--
参考にしたドキュメント、記事、OSSプロジェクトなど
-->

- htmx 4 ソースコード・ドキュメント: `.claude/skills/htmx4/src/htmx-4.0.0-alpha8/`
- Wikino の Datastar→htmx4 移行作業計画書: `/wikino/docs/plans/3_done/202603/datastar-to-htmx4-migration.md`

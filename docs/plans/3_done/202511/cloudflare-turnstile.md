# Cloudflare Turnstile によるBot対策 設計書

<!--
このテンプレートの使い方:
1. このファイルを `.claude/designs/2_todo/` ディレクトリにコピー
   例: cp .claude/designs/template.md .claude/designs/2_todo/new-feature.md
2. [機能名] などのプレースホルダーを実際の内容に置き換え
3. 各セクションのガイドラインに従って記述
4. コメント（ `\<!-- ... --\>` ）はガイドラインとして残しておくことを推奨
-->

## 概要

<!--
ガイドライン:
- この機能が「何を」実現するのかを簡潔に説明
- ユーザーにとっての価値や背景を記述
- 2-3段落程度で簡潔に
-->

Cloudflare Turnstile を使用したBot対策機能を導入します。これにより、ログインフォームやサインアップフォームなどの重要なエンドポイントを自動化されたBot攻撃から保護します。

Turnstile は、従来のCAPTCHAと異なり、ほとんどの場合ユーザーに追加の操作を求めません（invisible challenge）。これにより、セキュリティを強化しつつ、ユーザーエクスペリエンスを損なわない実装が可能です。

**目的**:

- Bot による自動ログイン試行（ブルートフォース攻撃）を防ぐ
- Bot による自動アカウント作成（スパムアカウント）を防ぐ
- 正規ユーザーのUXを損なわずにセキュリティを強化する
- Cloudflare エコシステムとの統合を活かす（Annict は既に Cloudflare R2 を使用）

**背景**:

- Web アプリケーションはBot攻撃のリスクに常にさらされている
- 従来のCAPTCHA（画像選択など）はユーザーにストレスを与える
- Cloudflare Turnstile は無料で無制限、かつユーザーフレンドリー
- 設計書（@.claude/designs/1_doing/go.md:783）にも「Cloudflare Turnstile によるボット対策」として記載済み

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

- ログインフォーム送信時に Turnstile チャレンジが自動的に実行される
- サインアップフォーム送信時に Turnstile チャレンジが自動的に実行される
- パスワードリセット申請フォーム送信時に Turnstile チャレンジが自動的に実行される
- Turnstile 検証が失敗した場合、適切なエラーメッセージが表示される
- サーバーサイドで Turnstile トークンを検証し、不正なリクエストを拒否する

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

#### セキュリティ

- **トークン検証**: すべてのフォーム送信時にサーバーサイドで Turnstile トークンを検証する
- **シークレット管理**: Turnstile Secret Key は環境変数で管理し、コードに含めない
- **リトライ制限**: Turnstile 検証失敗時は適切にエラーを返し、無限リトライを防ぐ
- **ログ記録**: Turnstile 検証の成功・失敗をログに記録（監査用）

#### ユーザビリティ（UX）

- **透明性**: ほとんどの場合、ユーザーは何もする必要がない（invisible challenge）
- **フォールバック**: Turnstile が読み込めない場合も適切なエラーメッセージを表示
- **アクセシビリティ**: Turnstile ウィジェットは ARIA ラベルをサポート
- **多言語対応**: エラーメッセージは i18n で国際化対応（日本語・英語）

#### パフォーマンス

- **非同期読み込み**: Turnstile JavaScript は非同期で読み込む（ページ表示を遅延させない）
- **軽量**: Turnstile は reCAPTCHA よりも軽量で、ページ読み込みへの影響が少ない

#### 可用性・信頼性

- **エラーハンドリング**: Cloudflare API がダウンした場合も適切にエラーを処理
- **タイムアウト**: Turnstile API 検証にタイムアウト（5秒）を設定

## 設計

<!--
ガイドライン:
- 技術的な実装の詳細を記述
- 必要に応じて以下のようなサブセクションを追加してください：
  - 技術スタック（使用するライブラリ、フレームワーク、ツールなど）
  - アーキテクチャ（システム全体の構成、コンポーネント間の関係など）
  - データベース設計（テーブル定義、インデックス、制約など）
  - API設計（エンドポイント、リクエスト/レスポンス形式など）
  - セキュリティ設計（認証・認可、トークン管理、Rate Limitingなど）
  - コード設計（パッケージ構成、主要な構造体、インターフェースなど）
  - テスト戦略（単体テスト、統合テスト、E2Eテストの方針）
  - マイグレーション管理（データベースマイグレーションの方針）
  - 実装方針（特記事項、既存システムとの関係、制約など）

不要な場合はこのセクション全体を削除してください。
-->

### 技術スタック

- **Cloudflare Turnstile**: Bot対策サービス（完全無料、無制限）
- **Turnstile JavaScript API**: フロントエンドウィジェット
- **Cloudflare Siteverify API**: サーバーサイドトークン検証

### アーキテクチャ

```
┌─────────────────────────────────────────────────────────────────┐
│                        ブラウザ                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  ログインフォーム                                         │  │
│  │  ┌─────────────────────────────────────┐                 │  │
│  │  │  email: [          ]                │                 │  │
│  │  │  password: [       ]                │                 │  │
│  │  │  ┌─────────────────────────────┐   │                 │  │
│  │  │  │ Turnstile Widget (invisible)│   │                 │  │
│  │  │  └─────────────────────────────┘   │                 │  │
│  │  │  [ログイン]                         │                 │  │
│  │  └─────────────────────────────────────┘                 │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ POST /sign_in
                              │ cf-turnstile-response: <token>
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Go Application                              │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Handler (internal/handler/sign_in/create.go)            │  │
│  │    1. Turnstileトークンを取得                             │  │
│  │    2. VerifyTurnstile()を呼び出し                        │  │
│  │    3. 検証成功 → ログイン処理続行                        │  │
│  │    4. 検証失敗 → エラーレスポンス                        │  │
│  └──────────────────────────────────────────────────────────┘  │
│                              │                                   │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Turnstile Service (internal/turnstile/verify.go)        │  │
│  │    1. トークンをCloudflare APIに送信                     │  │
│  │    2. レスポンスを検証                                    │  │
│  │    3. 成功/失敗を返す                                     │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ POST https://challenges.cloudflare.com/turnstile/v0/siteverify
                              │ secret=<secret_key>&response=<token>
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│           Cloudflare Turnstile Siteverify API                    │
│  検証結果を返す: {"success": true/false}                         │
└─────────────────────────────────────────────────────────────────┘
```

### セキュリティ設計

#### トークンフロー

1. **フロントエンド**: ユーザーがフォームを送信すると、Turnstile がチャレンジを実行
2. **トークン生成**: Turnstile が検証トークン（`cf-turnstile-response`）を生成
3. **トークン送信**: フォーム送信時にトークンを hidden input として送信
4. **サーバーサイド検証**: Go アプリケーションが Cloudflare Siteverify API にトークンを送信
5. **検証結果**: Cloudflare が検証結果（success: true/false）を返す
6. **アクセス制御**: 検証成功 → 処理続行、検証失敗 → エラーレスポンス

#### 環境変数

```bash
# Turnstile Site Key (公開可能、フロントエンドで使用)
ANNICT_TURNSTILE_SITE_KEY=1x00000000000000000000AA

# Turnstile Secret Key (秘密情報、サーバーサイドで使用)
ANNICT_TURNSTILE_SECRET_KEY=1x0000000000000000000000000000000AA
```

### コード設計

#### パッケージ構成

```
internal/
├── turnstile/              # Turnstile関連のロジック
│   ├── verify.go           # トークン検証ロジック
│   └── verify_test.go      # テスト
├── handler/
│   ├── sign_in/
│   │   ├── new.go          # RESTful命名（GET /sign_in）
│   │   ├── create.go       # Turnstile検証を追加（POST /sign_in）
│   │   └── create_test.go
│   ├── sign_up/
│   │   ├── create.go       # Turnstile検証を追加（将来）
│   │   └── create_test.go
│   └── password/
│       ├── create.go       # Turnstile検証を追加
│       └── create_test.go
├── templates/
│   ├── components/
│   │   └── turnstile.templ # Turnstileウィジェットコンポーネント
│   └── pages/
│       ├── sign_in/
│       │   └── new.templ   # Turnstileコンポーネントを追加
│       └── password/
│           └── new.templ   # Turnstileコンポーネントを追加
└── config/
    └── config.go           # Turnstile設定を追加
```

#### 主要な構造体

```go
// internal/turnstile/verify.go
package turnstile

// Client は Turnstile API クライアント
type Client struct {
    siteKey   string
    secretKey string
    httpClient *http.Client
}

// VerifyResponse は Cloudflare Siteverify API のレスポンス
type VerifyResponse struct {
    Success     bool     `json:"success"`
    ChallengeTS string   `json:"challenge_ts"`
    Hostname    string   `json:"hostname"`
    ErrorCodes  []string `json:"error-codes"`
}

// Verify はトークンを検証する
func (c *Client) Verify(ctx context.Context, token string) (bool, error)
```

### テスト戦略

#### 単体テスト

- `internal/turnstile/verify_test.go`: Turnstile API 検証のモックテスト
  - 成功ケース: `{"success": true}` のレスポンス
  - 失敗ケース: `{"success": false}` のレスポンス
  - タイムアウトケース: API がタイムアウトする場合
  - ネットワークエラーケース: API が到達不能な場合

#### 統合テスト

- `internal/handler/sign_in/create_test.go`: ログインフォーム送信のテスト
  - Turnstile トークンが含まれない場合 → エラー
  - Turnstile トークンが無効な場合 → エラー
  - Turnstile トークンが有効な場合 → ログイン成功

#### E2Eテスト（手動テスト）

- 実際のブラウザで Turnstile ウィジェットが表示されることを確認
- フォーム送信時に Turnstile チャレンジが実行されることを確認
- 検証失敗時に適切なエラーメッセージが表示されることを確認

### 実装方針

#### 段階的な導入

1. **フェーズ1**: ログインフォームにのみ Turnstile を導入（影響範囲を最小化）
2. **フェーズ2**: パスワードリセット申請フォームに導入
3. **フェーズ3**: サインアップフォームに導入（実装完了後）

#### 既存システムとの関係

- **CSRF 対策**: Turnstile は CSRF トークンと併用する（二重のセキュリティ）
- **Rate Limiting**: 将来的に Redis ベースの Rate Limiting と組み合わせる
- **セッション管理**: Turnstile 検証はセッション作成前に実行

#### 制約

- Cloudflare アカウントとドメインの設定が必要
- Site Key と Secret Key の取得が必要
- JavaScript が無効なブラウザでは動作しない（フォールバック不要、JavaScript 必須の方針）

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

タスク番号の付け方:
- 各タスクには階層的な番号を付与します（例: 1-1, 1-2, 2-1, 2-2）
- フォーマット: **フェーズ番号-タスク番号**: タスク名
- タスクの前に別のタスクを追加する場合は、サブ番号を使用します
  - 例: タスク 2-1 の前にタスクを追加する場合 → 2-0
  - 例: タスク 2-0 の前にタスクを追加する場合 → 2-0-1
- この番号はブランチ名の一部として使用されます（例: feature-1-1, feature-2-0）
-->

### フェーズ 1: インフラ準備

<!--
例: インフラ準備、基本機能実装、セキュリティ機能など
各タスクは1つのPull Requestで完結する粒度で記述してください
各タスクには想定サイズを明記してください
-->

- [x] **1-1**: Cloudflare Turnstile アカウント設定とドメイン登録
  - Cloudflare ダッシュボードで Turnstile を有効化
  - ドメイン（`annict.com`, `example.dev`）を登録
  - Site Key と Secret Key を取得
  - `.env.example` にサンプル設定を追加
  - ドキュメント更新（`go/CLAUDE.md` に Turnstile の説明を追加）
  - **想定ファイル数**: 約 2 ファイル（実装 2 + テスト 0）
  - **想定行数**: 約 20 行（実装 20 行 + テスト 0 行）

- [x] **1-2**: 環境変数とConfigの追加
  - `internal/config/config.go` に Turnstile 設定を追加
  - 環境変数（`ANNICT_TURNSTILE_SITE_KEY`, `ANNICT_TURNSTILE_SECRET_KEY`）を読み込む
  - テスト環境用のモック設定を追加
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 50 行（実装 30 行 + テスト 20 行）

### フェーズ 2: Turnstile API クライアント実装

- [x] **2-1**: Turnstile API クライアントの実装
  - `internal/turnstile/verify.go` を作成
  - `Client` 構造体を定義
  - `Verify()` メソッドを実装（Cloudflare Siteverify API を呼び出し）
  - タイムアウト設定（5秒）
  - エラーハンドリング（ネットワークエラー、APIエラー）
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 150 行（実装 80 行 + テスト 70 行）

### フェーズ 3: フロントエンド実装（ログインフォーム）

- [x] **3-1**: Turnstile ウィジェットコンポーネントの作成
  - `internal/templates/components/turnstile.templ` を作成
  - Turnstile JavaScript の読み込み（非同期）
  - ウィジェットの表示（invisible mode）
  - Site Key を props で受け取る
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 80 行（実装 40 行 + テスト 40 行）

- [x] **3-1-1**: ログインハンドラーとテンプレートをRESTful命名に統一
  - `internal/handler/sign_in/show.go` → `internal/handler/sign_in/new.go` にリネーム
  - メソッド名 `Show` → `New` に変更
  - `internal/templates/pages/sign_in.templ` → `internal/templates/pages/sign_in/new.templ` に移動
  - テンプレート関数名 `SignInShow` → `SignInNew` に変更
  - RESTful命名規則に準拠（GET /sign_in = new、POST /sign_in = create）
  - プロジェクトのガイドライン（機能別にディレクトリを分割）に準拠
  - ハンドラーでのimportパスとルーティングを更新
  - テンプレートテストを更新
  - **想定ファイル数**: 約 6 ファイル（実装 4 + テスト 2）
  - **想定行数**: 約 40 行（実装 20 行 + テスト 20 行）
  - **Note**: 既存ファイルの移動とリネーム、importパス変更が主な作業

- [x] **3-2**: ログインフォームに Turnstile ウィジェットを追加
  - `internal/templates/pages/sign_in/new.templ` を更新
  - Turnstile コンポーネントを追加
  - テンプレートテストを更新
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 60 行（実装 30 行 + テスト 30 行）

### フェーズ 4: バックエンド実装（ログインフォーム）

- [x] **4-1**: ログインハンドラーに Turnstile 検証を追加
  - `internal/handler/sign_in/create.go` を更新
  - フォームから `cf-turnstile-response` トークンを取得
  - `turnstile.Client.Verify()` を呼び出し
  - 検証失敗時にエラーレスポンスを返す
  - ログ記録（検証成功・失敗）
  - **想定ファイル数**: 約 2 ファイル（実装 1 + テスト 1）
  - **想定行数**: 約 120 行（実装 60 行 + テスト 60 行）

- [x] **4-2**: エラーメッセージの国際化対応
  - `internal/i18n/locales/ja.toml` と `en.toml` に Turnstile エラーメッセージを追加
  - エラーメッセージキー: `errors.turnstile_verification_failed`
  - 日本語: "Bot対策の検証に失敗しました。もう一度お試しください。"
  - 英語: "Bot verification failed. Please try again."
  - **想定ファイル数**: 約 2 ファイル（実装 2 + テスト 0）
  - **想定行数**: 約 10 行（実装 10 行 + テスト 0 行）

### フェーズ 5: パスワードリセット申請フォームへの展開

- [x] **5-1**: パスワードリセット申請フォームに Turnstile ウィジェットを追加
  - `internal/templates/pages/password/reset.templ` を更新
  - Turnstile コンポーネントを追加
  - `internal/handler/password_reset/new.go` を更新して `h.cfg.TurnstileSiteKey` を渡す
  - テンプレートテストを更新（既存のテストが通ることを確認）
  - **想定ファイル数**: 約 2 ファイル（実装 2）
  - **想定行数**: 約 30 行（実装 30 行）

- [x] **5-2**: パスワードリセット申請ハンドラーに Turnstile 検証を追加
  - `internal/handler/password_reset/create.go` を更新
  - フォームから `cf-turnstile-response` トークンを取得
  - `turnstile.Client.Verify()` を呼び出し
  - 検証失敗時にエラーレスポンスを返す
  - ログ記録（検証成功・失敗）
  - **想定ファイル数**: 約 4 ファイル（実装 3 + テスト 1）
  - **想定行数**: 約 150 行（実装 60 行 + テスト 90 行）
  - **Note**: インターフェース `turnstile.Verifier` を追加してテスタビリティを向上

### フェーズ 6: ドキュメントと動作確認

- [ ] **6-1**: ドキュメント更新と動作確認
  - `go/docs/security-guide.md` に Turnstile の説明を追加
  - 開発環境での動作確認手順をドキュメント化
  - 本番環境へのデプロイ前のチェックリストを作成
  - 手動E2Eテストを実施（実際のブラウザで確認）
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 100 行（実装 100 行 + テスト 0 行）

### 実装しない機能（スコープ外）

<!--
今回は実装しないが、将来的に検討する機能を明記
-->

以下の機能は今回の実装では**実装しません**：

- **サインアップフォームへの適用**: サインアップ機能がまだ実装されていないため、実装完了後に別途対応
- **Rate Limiting との統合**: Redis ベースの Rate Limiting は別の機能として実装予定
- **管理画面での統計確認**: Cloudflare ダッシュボードで統計を確認するため、Go アプリ内での統計表示は不要
- **Turnstile のカスタマイズ**: デフォルト設定（invisible mode）で十分なため、カスタマイズは将来的に検討

## 参考資料

<!--
参考にしたドキュメント、記事、OSSプロジェクトなど
-->

- [Cloudflare Turnstile 公式ドキュメント](https://developers.cloudflare.com/turnstile/)
- [Turnstile Client-side Integration](https://developers.cloudflare.com/turnstile/get-started/client-side-rendering/)
- [Turnstile Server-side Validation](https://developers.cloudflare.com/turnstile/get-started/server-side-validation/)
- [Turnstile vs reCAPTCHA 比較](https://blog.cloudflare.com/turnstile-private-captcha-alternative/)

---

## テンプレート使用例

実際の使用例は以下を参照してください：

- [パスワードリセット機能](doing/password-reset.md)

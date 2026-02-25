# 新規登録機能 設計書

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

新規ユーザーがメールアドレスで登録し、新規登録の確認コードを入力してアカウントを作成できる機能を実装します。パスワードは設定せず、メールアドレスログインのみに対応します。

**目的**:

- ユーザーがメールアドレスだけで簡単にアカウントを作成できるようにする
- パスワードレス認証により、パスワード管理の負担を軽減する
- セキュリティを最優先に、OWASP推奨のベストプラクティスに従った実装を行う

**背景**:

- Rails版の Devise ベースの実装とは独立した、Go版独自の新規登録機能を構築する
- メールアドレスでログイン機能が既に実装されているため、パスワード設定は不要
- Cloudflare Turnstile によるBot対策を導入し、スパム登録を防ぐ

**用語の統一**:

- **UI表示**: 「新規登録」（日本語）、「Sign Up」（英語）
- **コード**: `sign_up`（内部実装）
- **確認コード**: 「新規登録の確認コード」（新規登録時）、「ログインの確認コード」（ログイン時）

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

- ユーザーはメールアドレスを入力して新規登録を申請できる
- ユーザーは利用規約とプライバシーポリシーに同意する必要がある
- システムはユーザーのメールアドレスに6桁の新規登録の確認コードを送信する
- ユーザーはメールで受け取った新規登録の確認コードを入力して検証できる
- 新規登録の確認コードは15分の有効期限を持つ
- ユーザーは確認コード検証後、ユーザー名を設定してアカウントを作成できる
- ユーザー名は一意である必要がある（重複チェック）
- アカウント作成後、ユーザーは自動的にログイン状態になる
- メールは日本語と英語の両方に対応し、ユーザーのロケール設定に応じて送信される
- ユーザーはコードの再送信をリクエストできる

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

- **新規登録の確認コード生成**: 暗号学的に安全な乱数生成器（crypto/rand）を使用し、6桁の数字コードを生成
- **新規登録の確認コード保存**: bcryptでハッシュ化してデータベース（sign_up_codesテーブル）に保存し、15分の有効期限を設定
- **試行回数制限**: データベースに試行回数（attempts）を記録し、5回を超えたら検証を拒否
- **Rate Limiting**: ブルートフォース攻撃を防ぐため、IP単位・メール単位での厳格な制限を実装（Redis使用）
- **Bot対策**: Cloudflare Turnstile でスパム登録を防ぐ
- **情報漏洩対策**: エラーメッセージでユーザーの存在やコードの状態を明かさない
- **監査ログ**: セキュリティインシデントの調査のため、すべての重要なイベントをログに記録
- **ユーザー名ポリシー**: 20文字以内、半角英数字とアンダースコアのみ（最小文字数制限なし）

#### パフォーマンス

- 新規登録の確認コード検証はデータベースインデックスを活用した高速検索
- Rate LimitingにはRedisを使用し、高速な制限チェックを実現
- バックグラウンドジョブ（River）により、メール送信がHTTPリクエストをブロックしない

#### ユーザビリティ（UX）

- わかりやすいエラーメッセージとヘルプテキストを提供
- ユーザー名要件を明確に表示（20文字以内、半角英数字とアンダースコア）
- メールが届かない場合のサポート情報を提供
- コード再送信機能でスムーズな体験を実現
- アカウント作成後の自動ログインでスムーズな体験を実現

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

- **メール送信**: resend-go/v2（Resend API）
- **バックグラウンドジョブ**: riverqueue/river
- **Rate Limiting**: Redis + go-redis/v9
- **新規登録の確認コード生成**: crypto/rand
- **Bot対策**: Cloudflare Turnstile
- **ユーザー名検証**: sqlc（既存のusersテーブルを使用）

### データベース設計

#### 既存のusersテーブルを使用

既存の `users` テーブルを使用します。

```sql
-- 既存のusersテーブル（参考）
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    encrypted_password VARCHAR(255) NOT NULL DEFAULT '',  -- パスワードレス登録の場合は空文字列
    locale VARCHAR(10) DEFAULT 'ja',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    ...
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
```

**テーブル設計のポイント**:

- **username**: ユーザー名（一意制約あり）
- **email**: メールアドレス（一意制約あり）
- **encrypted_password**: パスワードレス登録の場合は空文字列（NOT NULL 制約）
- **locale**: ユーザーのロケール（メール送信時に使用）

#### 新規登録の確認コードテーブル（sign_up_codes）

新規登録の確認コードはデータベースに保存します。既存のログイン機能と同じパターンで、別テーブルとして実装します。

```sql
-- 新規登録の確認コードテーブル
CREATE TABLE sign_up_codes (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    code_digest VARCHAR(255) NOT NULL,  -- bcryptでハッシュ化された確認コード
    attempts INTEGER NOT NULL DEFAULT 0,  -- 検証試行回数
    used_at TIMESTAMP WITH TIME ZONE,  -- 使用日時（NULL = 未使用）
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,  -- 有効期限（15分）
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sign_up_codes_email ON sign_up_codes(email);
CREATE INDEX idx_sign_up_codes_expires_at ON sign_up_codes(expires_at);
```

**テーブル設計のポイント**:

- **email**: メールアドレス（ユニーク制約なし。複数回の申請を許可）
- **code_digest**: bcryptでハッシュ化された6桁の確認コード
- **attempts**: 検証試行回数（ブルートフォース攻撃対策）
- **used_at**: 使用日時（NULL = 未使用、監査ログとして記録）
- **expires_at**: 有効期限（作成時刻 + 15分）
- **インデックス**: メールアドレスと有効期限での検索を高速化

**セキュリティと監査ログ**:

- 確認コードは bcrypt でハッシュ化して保存（平文保存は禁止）
- ログイン機能と同じパターンで実装（既存の `email_login_codes` テーブルと同様）
- 既存のコードは DELETE せず無効化（`used_at` を更新）することで監査ログとして残す
  - 「いつコードを送信したか」の履歴が残る
  - デバッグ時に「何回コードを送信したか」がわかる
  - ユーザー行動の分析に使える

**将来の命名統一**:

現在、ログイン機能では `email_login_codes` テーブルを使用していますが、将来的に `sign_in_codes` にリネームすることで、以下の統一されたパターンになります：

- `sign_up_codes` - 新規登録の確認コード
- `sign_in_codes` - ログインの確認コード（将来的にリネーム）

この命名パターンは `{action}_codes` という一貫性のある形式で、`email_` プレフィックスを削除することでシンプルになります。

#### 関連テーブル（profiles, settings, email_notifications）

ユーザー作成時には、以下の関連テーブルにもレコードを作成します。これはRails版の`build_relations`メソッドと同じ動作です。

**profilesテーブル**:

```sql
-- プロフィールテーブル（既存）
CREATE TABLE profiles (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    name VARCHAR(510) NOT NULL DEFAULT '',
    description VARCHAR(510) NOT NULL DEFAULT '',
    url VARCHAR,
    image_data TEXT,
    background_image_data TEXT,
    background_image_animated BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_profiles_user_id ON profiles(user_id);
```

**settingsテーブル**:

```sql
-- 設定テーブル（既存）
CREATE TABLE settings (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    privacy_policy_agreed BOOLEAN NOT NULL DEFAULT false,
    hide_record_body BOOLEAN NOT NULL DEFAULT true,
    hide_supporter_badge BOOLEAN NOT NULL DEFAULT false,
    -- その他の設定フィールド...
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_settings_user_id ON settings(user_id);
```

**email_notificationsテーブル**:

```sql
-- メール通知設定テーブル（既存）
CREATE TABLE email_notifications (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    unsubscription_key VARCHAR NOT NULL,
    event_followed_user BOOLEAN NOT NULL DEFAULT true,
    event_liked_episode_record BOOLEAN NOT NULL DEFAULT true,
    event_friends_joined BOOLEAN NOT NULL DEFAULT true,
    event_next_season_came BOOLEAN NOT NULL DEFAULT true,
    event_favorite_works_added BOOLEAN NOT NULL DEFAULT true,
    event_related_works_added BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_email_notifications_user_id ON email_notifications(user_id);
```

**ユーザー作成時のデフォルト値**:

- **profiles**: `name`にユーザー名を設定、`description`は空文字列
- **settings**: `privacy_policy_agreed`を`true`に設定、その他はデフォルト値
- **email_notifications**: `unsubscription_key`にUUIDを生成して設定、その他はデフォルト値（全てtrue）

**トランザクション管理**:

これらのレコードはすべてトランザクション内で作成され、いずれか1つでも失敗した場合は全てロールバックされます。

### API設計

#### 1. GET /sign_up - 新規登録フォーム

新規ユーザー登録フォームを表示

**レスポンス**: 200 OK（HTML）

#### 2. POST /sign_up - 新規登録の確認コード送信

メールアドレスを受け取り、新規登録の確認コードをメールで送信

**リクエスト**:

```
POST /sign_up
Content-Type: application/x-www-form-urlencoded

email=user@example.com
terms_agreed=true
cf-turnstile-response=xxx
csrf_token=xxx
```

**処理フロー**:

1. CSRF トークン検証
2. Cloudflare Turnstile 検証
3. Rate Limiting チェック（IP、メールアドレス）
4. メールアドレスの重複チェック（usersテーブル）
5. 新規登録の確認コードを生成（6桁の数字）
6. 確認コードをbcryptでハッシュ化
7. データベースに保存（sign_up_codesテーブル、15分の有効期限）
8. メール送信ジョブをエンキュー（平文の確認コードを渡す）
9. 新規登録の確認コード入力画面にリダイレクト

**レスポンス**: 302 Found（リダイレクト） or 400 Bad Request（HTML）

#### 3. GET /sign_up/code - 新規登録の確認コード入力フォーム

新規登録の確認コード入力フォームを表示

**セッション要件**:

- `sign_up_email`: メールアドレス（確認コードを送信したメールアドレス）がセッションに保存されている必要がある

**レスポンス**: 200 OK（HTML）

#### 4. POST /sign_up/code - 新規登録の確認コード検証

新規登録の確認コードを検証し、正しければユーザー名設定画面にリダイレクト

**リクエスト**:

```
POST /sign_up/code
Content-Type: application/x-www-form-urlencoded

email=user@example.com
code=123456
csrf_token=xxx
```

**処理フロー**:

1. CSRF トークン検証
2. Rate Limiting チェック（IP単位、コード検証）
3. データベースから確認コードを取得（sign_up_codesテーブル）
4. 有効期限チェック（expires_at）
5. 使用済みチェック（used_at IS NULL）
6. 試行回数チェック（attemptsが上限を超えていないか）
7. bcrypt.CompareHashAndPasswordでコードを検証
8. 検証成功: used_atを更新（監査ログとして記録）
9. 検証成功: 一時トークンを生成してRedisに保存（次のステップで使用）
10. 検証失敗: attemptsをインクリメント、エラーメッセージを表示
11. ユーザー名設定画面にリダイレクト

**レスポンス**: 302 Found（リダイレクト） or 400 Bad Request（HTML）

#### 5. GET /sign_up/username - ユーザー名設定フォーム

ユーザー名設定フォームを表示

**クエリパラメータ**:

- `token`: 一時トークン（確認コード検証後に生成）

**レスポンス**: 200 OK（HTML）

#### 6. POST /sign_up/username - ユーザー登録完了

ユーザー名を設定してアカウントを作成

**リクエスト**:

```
POST /sign_up/username
Content-Type: application/x-www-form-urlencoded

token=abc123xyz
username=johndoe
csrf_token=xxx
```

**処理フロー**:

1. CSRF トークン検証
2. 一時トークンの検証（Redisから取得）
3. ユーザー名のバリデーション（長さ、文字種、一意性）
4. トランザクション開始
5. ユーザーをusersテーブルに作成（encrypted_passwordは空文字列 `''`）
6. プロフィールをprofilesテーブルに作成（name: ユーザー名、description: 空文字列）
7. 設定をsettingsテーブルに作成（privacy_policy_agreed: true、その他はデフォルト値）
8. メール通知設定をemail_notificationsテーブルに作成（unsubscription_key: UUID）
9. セッションを作成してログイン状態にする
10. 一時トークンを削除（Redis）
11. トランザクションコミット
12. ホームページにリダイレクト

**レスポンス**: 302 Found（リダイレクト） or 400 Bad Request（HTML）

#### 7. PATCH /sign_up/code - 新規登録の確認コード再送信

新規登録の確認コードを再送信

**リクエスト**:

```
PATCH /sign_up/code
Content-Type: application/x-www-form-urlencoded

csrf_token=xxx
```

**処理フロー**:

1. CSRF トークン検証
2. セッションから`sign_up_email`を取得（なければ `/sign_up` にリダイレクト）
3. Rate Limiting チェック（IP、メールアドレス）
4. トランザクション開始
5. 既存の未使用コードを無効化（`used_at`を更新）
6. 新規登録の確認コードを生成（6桁の数字）
7. 確認コードをbcryptでハッシュ化
8. データベースに保存（sign_up_codesテーブル）
9. トランザクションコミット
10. メール送信ジョブをエンキュー（平文の確認コードを渡す）
11. フラッシュメッセージを設定（「確認コードを再送信しました」）
12. `/sign_up/code` にリダイレクト

**レスポンス**: 302 Found（リダイレクト）

### セキュリティ設計

#### 1. 新規登録の確認コード生成

**要件**:

- 暗号学的に安全な乱数生成器（`crypto/rand`）を使用
- 6桁の数字コード（100000～999999）

**実装例**:

```go
import (
    "crypto/rand"
    "fmt"
    "math/big"
)

// generateVerificationCode は6桁の確認コードを生成します
// 新規登録とログインの両方で共通して使用します
func generateVerificationCode() (string, error) {
    // 100000～999999の範囲の乱数を生成
    n, err := rand.Int(rand.Reader, big.NewInt(900000))
    if err != nil {
        return "", err
    }
    code := n.Int64() + 100000
    return fmt.Sprintf("%06d", code), nil
}
```

#### 2. 新規登録の確認コード保存（データベース）

**要件**:

- データベースに保存し、15分の有効期限を設定
- bcryptでハッシュ化して保存（平文保存は禁止）
- 試行回数を記録してブルートフォース攻撃を防ぐ

**実装例（sqlc）**:

```sql
-- name: CreateSignUpCode :one
INSERT INTO sign_up_codes (email, code_digest, expires_at)
VALUES ($1, $2, NOW() + INTERVAL '15 minutes')
RETURNING *;

-- name: GetValidSignUpCode :one
SELECT * FROM sign_up_codes
WHERE email = $1
  AND expires_at > NOW()
  AND used_at IS NULL
  AND attempts < 5
ORDER BY created_at DESC
LIMIT 1;

-- name: IncrementSignUpCodeAttempts :exec
UPDATE sign_up_codes
SET attempts = attempts + 1,
    updated_at = NOW()
WHERE id = $1;

-- name: MarkSignUpCodeAsUsed :exec
UPDATE sign_up_codes
SET used_at = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: InvalidateSignUpCodesByEmail :exec
UPDATE sign_up_codes
SET used_at = NOW(),
    updated_at = NOW()
WHERE email = $1
  AND used_at IS NULL;
```

**使用例**:

```go
import (
    "context"
    "time"

    "golang.org/x/crypto/bcrypt"
)

// 確認コードを保存
func saveVerificationCode(ctx context.Context, queries *repository.Queries, email, code string) error {
    // bcryptでハッシュ化
    hashedCode, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
    if err != nil {
        return err
    }

    // データベースに保存
    _, err = queries.CreateSignUpCode(ctx, repository.CreateSignUpCodeParams{
        Email:      email,
        CodeDigest: string(hashedCode),
    })
    return err
}

// 確認コードを検証
func verifyCode(ctx context.Context, queries *repository.Queries, email, code string) (int64, error) {
    // 有効な確認コードを取得（used_at IS NULL かつ有効期限内）
    record, err := queries.GetValidSignUpCode(ctx, email)
    if err != nil {
        return 0, err
    }

    // bcryptで検証
    err = bcrypt.CompareHashAndPassword([]byte(record.CodeDigest), []byte(code))
    if err != nil {
        // 検証失敗: 試行回数をインクリメント
        _ = queries.IncrementSignUpCodeAttempts(ctx, record.ID)
        return 0, fmt.Errorf("invalid code")
    }

    // 検証成功: used_atを更新（監査ログとして記録）
    err = queries.MarkSignUpCodeAsUsed(ctx, record.ID)
    if err != nil {
        return 0, err
    }

    return record.ID, nil
}

// 確認コードを再送信（ユースケース例）
func resendVerificationCode(ctx context.Context, db *sql.DB, queries *repository.Queries, email string) (string, error) {
    // トランザクション開始
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return "", fmt.Errorf("トランザクション開始に失敗: %w", err)
    }
    defer tx.Rollback()

    queriesWithTx := queries.WithTx(tx)

    // 既存の未使用コードを無効化（監査ログとして残す）
    if err := queriesWithTx.InvalidateSignUpCodesByEmail(ctx, email); err != nil {
        return "", fmt.Errorf("古いコードの無効化に失敗: %w", err)
    }

    // 新しい確認コードを生成
    code, err := generateVerificationCode()
    if err != nil {
        return "", fmt.Errorf("コード生成に失敗: %w", err)
    }

    // bcryptでハッシュ化
    hashedCode, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
    if err != nil {
        return "", fmt.Errorf("コードのハッシュ化に失敗: %w", err)
    }

    // データベースに保存
    _, err = queriesWithTx.CreateSignUpCode(ctx, repository.CreateSignUpCodeParams{
        Email:      email,
        CodeDigest: string(hashedCode),
    })
    if err != nil {
        return "", fmt.Errorf("確認コードの作成に失敗: %w", err)
    }

    // トランザクションコミット
    if err := tx.Commit(); err != nil {
        return "", fmt.Errorf("トランザクションのコミットに失敗: %w", err)
    }

    return code, nil
}
```

#### 3. Rate Limiting（レート制限）

ブルートフォース攻撃を防ぐため、以下の制限を実装します。

**制限ルール**:

| 制限対象           | 制限値        | 理由                             |
| ------------------ | ------------- | -------------------------------- |
| IP アドレス単位    | 5 回/時間     | 同一 IP からの登録申請を制限     |
| メールアドレス単位 | 3 回/時間     | 同一メールアドレスへの送信を制限 |
| コード検証         | 10 回/時間/IP | コードを総当たりで試す攻撃を防ぐ |

**実装方針**:

Redis を使用して Rate Limiting を実装します。

- インフラ: Rails が既に使用している Redis を活用（追加コスト不要）
- クライアント: github.com/redis/go-redis/v9
- namespace: `rate_limit:*` で Rails のキャッシュと分離
- アルゴリズム: Sliding Window Counter（Lua スクリプトでアトミック操作）
- TTL: 自動期限切れで古いデータを削除（クリーンアップ不要）

#### 4. Cloudflare Turnstile（Bot対策）

**要件**:

- 新規登録フォームにCloudflare Turnstileウィジェットを配置
- フォーム送信時にTurnstileトークンを検証
- 検証失敗時はエラーメッセージを表示

**実装方針**:

既存の `internal/turnstile` パッケージを使用します。

```go
// Turnstile検証
verifyResult, err := h.turnstile.Verify(ctx, turnstileToken, clientIP)
if err != nil || !verifyResult.Success {
    // Bot検出: エラーメッセージを表示
    return
}
```

#### 5. ユーザー名バリデーション

**要件**:

新しいユーザー名は以下の条件を満たす必要があります：

- **最大文字数**: 20 文字以内
- **使用可能文字**: 半角英数字（a-z、A-Z、0-9）とアンダースコア（\_）のみ
- **一意性**: 既存ユーザーと重複しない
- **必須**: 空文字列は不可

実装例:

```go
import (
    "errors"
    "regexp"
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{1,20}$`)

// ValidateUsername はユーザー名の形式をチェックします
func ValidateUsername(username string) error {
    if username == "" {
        return errors.New("ユーザー名を入力してください")
    }
    if !usernameRegex.MatchString(username) {
        return errors.New("ユーザー名は20文字以内の半角英数字とアンダースコアのみ使用できます")
    }
    return nil
}
```

#### 6. 情報漏洩対策

**セキュアなエラーメッセージ**:

システムの内部情報を漏らさないため、常に一般的なメッセージを返します。

| シナリオ           | NG（情報漏洩）                               | OK（セキュア）                         |
| ------------------ | -------------------------------------------- | -------------------------------------- |
| メールアドレス重複 | 「このメールアドレスは既に登録されています」 | 「このメールアドレスは使用できません」 |
| ユーザー名重複     | 「このユーザー名は既に使用されています」     | 「このユーザー名は使用できません」     |
| コード期限切れ     | 「コードの有効期限が切れています」           | 「無効なコードです」                   |
| コード不正         | 「コードが正しくありません」                 | 「無効なコードです」                   |

### バックグラウンドジョブ設計

#### メール送信ジョブ（River）

**ジョブ引数**:

```go
type SendVerificationCodeEmailArgs struct {
    Email string `json:"email"`
    Code  string `json:"code"` // 6桁の数字コード
}

func (SendVerificationCodeEmailArgs) Kind() string {
    return "send_verification_code_email"
}
```

**ワーカー実装**:

```go
import (
    "context"
    "fmt"
    "github.com/resend/resend-go/v2"
)

type SendVerificationCodeEmailWorker struct {
    river.WorkerDefaults[SendVerificationCodeEmailArgs]
    resendClient *resend.Client
    fromEmail    string
}

func (w *SendVerificationCodeEmailWorker) Work(ctx context.Context, job *river.Job[SendVerificationCodeEmailArgs]) error {
    // メール送信（Resend API）
    params := &resend.SendEmailRequest{
        From:    w.fromEmail,
        To:      []string{job.Args.Email},
        Subject: "Annict アカウント確認コード",
        Text: fmt.Sprintf(`Annictへようこそ！

アカウント登録を完了するため、以下の確認コードを入力してください：

%s

このコードは15分間有効です。

このメールに心当たりがない場合は、無視してください。
`, job.Args.Code),
    }

    _, err := w.resendClient.Emails.Send(params)
    if err != nil {
        return fmt.Errorf("failed to send email: %w", err)
    }

    return nil
}
```

### メール送信設計

#### メールテンプレート

メールは日本語と英語の両方に対応し、ユーザーのロケール設定に応じて送信します。

**テキストメール（日本語）**:

```
件名: Annict アカウント確認コード

本文:
Annictへようこそ！

アカウント登録を完了するため、以下の確認コードを入力してください：

123456

このコードは15分間有効です。

このメールに心当たりがない場合は、無視してください。
```

**テキストメール（英語）**:

```
Subject: Annict Account Verification Code

Body:
Welcome to Annict!

Please enter the following verification code to complete your account registration:

123456

This code is valid for 15 minutes.

If you did not request this, please ignore this email.
```

**HTML メール（日本語）**:

```html
<!DOCTYPE html>
<html lang="ja">
  <head>
    <meta charset="UTF-8" />
    <title>Annict アカウント確認コード</title>
  </head>
  <body>
    <h2>Annictへようこそ！</h2>
    <p>アカウント登録を完了するため、以下の確認コードを入力してください：</p>
    <p style="font-size: 24px; font-weight: bold; letter-spacing: 4px;">123456</p>
    <p>このコードは15分間有効です。</p>
    <p>このメールに心当たりがない場合は、無視してください。</p>
  </body>
</html>
```

**HTML メール（英語）**:

```html
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <title>Annict Account Verification Code</title>
  </head>
  <body>
    <h2>Welcome to Annict!</h2>
    <p>Please enter the following verification code to complete your account registration:</p>
    <p style="font-size: 24px; font-weight: bold; letter-spacing: 4px;">123456</p>
    <p>This code is valid for 15 minutes.</p>
    <p>If you did not request this, please ignore this email.</p>
  </body>
</html>
```

### 監査ログ設計

セキュリティインシデントの調査のため、以下をログに記録します。

**ログ項目**:

| イベント           | ログレベル | 記録内容                                             |
| ------------------ | ---------- | ---------------------------------------------------- |
| 新規登録申請       | Info       | メールアドレス、IP アドレス、成功/失敗               |
| Rate Limiting 発動 | Warning    | IP アドレス、メールアドレス、制限種別                |
| 確認コード検証失敗 | Warning    | メールアドレス、IP アドレス、失敗理由                |
| ユーザー登録成功   | Info       | ユーザー ID、ユーザー名、メールアドレス、IP アドレス |
| ユーザー登録失敗   | Warning    | メールアドレス、IP アドレス、失敗理由                |

**実装例（log/slog）**:

```go
import "log/slog"

// 新規登録申請
slog.InfoContext(ctx, "sign up requested",
    "email", email,
    "ip_address", getClientIP(r),
)

// 確認コード検証失敗
slog.WarnContext(ctx, "invalid verification code",
    "email", email,
    "ip_address", getClientIP(r),
    "reason", "expired",
)

// ユーザー登録成功
slog.InfoContext(ctx, "user registered successfully",
    "user_id", user.ID,
    "username", user.Username,
    "email", user.Email,
    "ip_address", getClientIP(r),
)
```

### UX 設計

#### 1. わかりやすいメッセージ

- 新規登録の確認コード送信後: 「メールを確認してください。6桁の新規登録の確認コードを送信しました。」
- メールが届かない場合のヘルプ: 「メールが届かない場合は、迷惑メールフォルダを確認してください。」
- 再送信リンク: 「メールが届かない場合は、[こちら]から再送信できます。」
- ユーザー名要件の表示: 「20文字以内、半角英数字とアンダースコアのみ使用可能です。」

#### 2. 成功後の自動ログイン

ユーザー登録後、自動的にログイン状態にしてホームページにリダイレクト：

```go
// セッションを作成してログイン状態にする
sessionID, _ := h.sessionStore.CreateSession(ctx, user.ID)
http.SetCookie(w, &http.Cookie{
    Name:     "_annict_session_v201904",
    Value:    sessionID,
    Path:     "/",
    Domain:   ".example.dev",
    Secure:   true,
    HttpOnly: true,
    SameSite: http.SameSiteLaxMode,
})

http.Redirect(w, r, "/", http.StatusSeeOther)
```

### テスト戦略

#### テスト用データベース

テストでは本番環境のデータベースとは分離された**テスト用データベース**を使用します。

**テストヘルパー（internal/testutil）**:

既存の `testutil.SetupTestDB(t)` と `testutil.SetupTestRedis(t)` を使用します。

- **SetupTestDB**: テスト用データベース接続とトランザクションをセットアップ（テスト終了時に自動ロールバック）
- **SetupTestRedis**: テスト用 Redis インスタンスをセットアップ（Rate Limiting用）

#### 統合テスト

新規登録フローのテスト:

```go
func TestSignUpFlow(t *testing.T) {
    db, tx := testutil.SetupTestDB(t)
    rdb := testutil.SetupTestRedis(t)

    queries := repository.New(db).WithTx(tx)
    handler := &Handler{
        queries: queries,
        redis:   rdb,
    }

    // 1. 新規登録申請
    req := httptest.NewRequest("POST", "/sign_up", strings.NewReader("email=test@example.com&terms_agreed=true"))
    rr := httptest.NewRecorder()
    handler.Create(rr, req)

    // 2. 新規登録の確認コードがDBに保存されているか確認
    signUpCode, err := queries.GetValidSignUpCode(ctx, "test@example.com")
    if err != nil {
        t.Fatalf("code not found: %v", err)
    }

    // 3. 確認コードを検証（テスト用に平文コードを取得）
    // NOTE: 実際のテストでは平文コードをテストデータとして保持しておく必要がある
    code := "123456" // テスト用の平文コード
    req = httptest.NewRequest("POST", "/sign_up/verify", strings.NewReader(
        fmt.Sprintf("email=test@example.com&code=%s", code),
    ))
    rr = httptest.NewRecorder()
    handler.Verify(rr, req)

    // 4. 一時トークンが生成されているか確認（Redis）
    token, _ := rdb.Get(ctx, "signup_token:test@example.com").Result()

    // 5. ユーザー登録
    req = httptest.NewRequest("POST", "/sign_up/complete", strings.NewReader(
        fmt.Sprintf("token=%s&username=testuser", token),
    ))
    rr = httptest.NewRecorder()
    handler.Complete(rr, req)

    // 6. ユーザーがDBに作成されているか確認
    user, err := queries.GetUserByUsername(ctx, "testuser")
    if err != nil {
        t.Fatalf("user not created: %v", err)
    }
    if user.Email != "test@example.com" {
        t.Errorf("wrong email: got %s, want test@example.com", user.Email)
    }
}
```

#### セキュリティテスト

**Rate Limiting のテスト**:

```go
func TestRateLimiting(t *testing.T) {
    db, tx := testutil.SetupTestDB(t)
    rdb := testutil.SetupTestRedis(t)  // テスト用 Redis（Rate Limiting用）

    queries := repository.New(db).WithTx(tx)
    handler := &Handler{
        queries: queries,
        redis:   rdb,
    }

    // 同一IPから6回新規登録申請（制限: 5回/時間）
    for i := 0; i < 6; i++ {
        req := httptest.NewRequest("POST", "/sign_up", strings.NewReader("email=test@example.com&terms_agreed=true"))
        req.RemoteAddr = "203.0.113.1:12345"
        rr := httptest.NewRecorder()
        handler.Create(rr, req)

        if i < 5 {
            // 最初の5回は成功
            if rr.Code != http.StatusOK && rr.Code != http.StatusSeeOther {
                t.Errorf("request %d should succeed", i+1)
            }
        } else {
            // 6回目は失敗（Rate Limit）
            if rr.Code != http.StatusTooManyRequests {
                t.Errorf("request %d should be rate limited", i+1)
            }
        }
    }
}
```

**確認コード試行回数制限のテスト**:

```go
func TestVerificationCodeAttemptsLimit(t *testing.T) {
    db, tx := testutil.SetupTestDB(t)
    queries := repository.New(db).WithTx(tx)

    // テスト用の確認コードを作成
    email := "test@example.com"
    correctCode := "123456"
    hashedCode, _ := bcrypt.GenerateFromPassword([]byte(correctCode), bcrypt.DefaultCost)

    signUpCode, err := queries.CreateSignUpCode(ctx, repository.CreateSignUpCodeParams{
        Email:      email,
        CodeDigest: string(hashedCode),
    })
    if err != nil {
        t.Fatalf("failed to create sign up code: %v", err)
    }

    // 5回間違ったコードで検証を試行
    for i := 0; i < 5; i++ {
        req := httptest.NewRequest("POST", "/sign_up/verify", strings.NewReader(
            fmt.Sprintf("email=%s&code=999999", email),
        ))
        rr := httptest.NewRecorder()
        handler.Verify(rr, req)

        // すべて失敗するはず
        if rr.Code == http.StatusSeeOther {
            t.Errorf("request %d should fail with wrong code", i+1)
        }
    }

    // 6回目は正しいコードでも拒否されるはず
    req := httptest.NewRequest("POST", "/sign_up/verify", strings.NewReader(
        fmt.Sprintf("email=%s&code=%s", email, correctCode),
    ))
    rr := httptest.NewRecorder()
    handler.Verify(rr, req)

    if rr.Code == http.StatusSeeOther {
        t.Errorf("request should be rejected due to too many attempts")
    }
}
```

### コード設計

#### パッケージ構成

- **ハンドラー**:
  - `internal/handler/sign_up/`
    - `handler.go`: Handler構造体と依存性（メールアドレス入力）
    - `new.go`: 新規登録フォーム表示（GET /sign_up）
    - `create.go`: 新規登録の確認コード送信（POST /sign_up）
    - `*_test.go`: 各ハンドラーのテスト
  - `internal/handler/sign_up_code/`
    - `handler.go`: Handler構造体と依存性（確認コード検証）
    - `show.go`: 新規登録の確認コード入力フォーム表示（GET /sign_up/code）
    - `create.go`: 新規登録の確認コード検証（POST /sign_up/code）
    - `update.go`: 新規登録の確認コード再送信（PATCH /sign_up/code）
    - `*_test.go`: 各ハンドラーのテスト
  - `internal/handler/sign_up_username/`
    - `handler.go`: Handler構造体と依存性（ユーザー登録）
    - `show.go`: ユーザー名設定フォーム表示（GET /sign_up/username）
    - `create.go`: ユーザー登録完了（POST /sign_up/username）
    - `*_test.go`: 各ハンドラーのテスト
- **テンプレート**:
  - `internal/templates/pages/sign_up/`
    - `new.templ`: 新規登録フォーム
  - `internal/templates/pages/sign_up_code/`
    - `show.templ`: 新規登録の確認コード入力フォーム
  - `internal/templates/pages/sign_up_username/`
    - `show.templ`: ユーザー名設定フォーム
- **ワーカー**: `internal/worker/`
  - `send_verification_code_email.go`: メール送信ジョブ
- **Repository**: sqlcで生成
  - `queries/users.sql`: ユーザー関連のSQLクエリ
  - `queries/sign_up_codes.sql`: 新規登録確認コード関連のSQLクエリ

#### 主要な構造体

```go
// Handler は新規登録ハンドラーの依存性を管理します
type Handler struct {
    queries   *repository.Queries
    redis     *redis.Client
    turnstile *turnstile.Client
    river     *river.Client[pgx.Tx]
    cfg       *config.Config
}

// CreateRequest は新規登録フォームのリクエストを表します
type CreateRequest struct {
    Email              string `form:"email"`
    TermsAgreed        bool   `form:"terms_agreed"`
    TurnstileResponse  string `form:"cf-turnstile-response"`
}

// VerifyRequest は新規登録の確認コード検証のリクエストを表します
type VerifyRequest struct {
    Email string `form:"email"`
    Code  string `form:"code"`
}

// CompleteRequest はユーザー登録完了のリクエストを表します
type CompleteRequest struct {
    Token    string `form:"token"`
    Username string `form:"username"`
}
```

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

### フェーズ 1: 基盤構築

<!--
例: インフラ準備、基本機能実装、セキュリティ機能など
各タスクは1つのPull Requestで完結する粒度で記述してください
各タスクには想定サイズを明記してください
-->

- [x] **1-1**: sign_up_codesテーブルの作成
  - データベースマイグレーションファイルの作成（`db/migrations/YYYYMMDDHHMMSS_create_sign_up_codes.sql`）
  - sqlcクエリファイルの作成（`internal/repository/queries/sign_up_codes.sql`）
    - `CreateSignUpCode`: 確認コードを作成
    - `GetValidSignUpCode`: 有効な確認コードを取得（used_at IS NULL）
    - `IncrementSignUpCodeAttempts`: 試行回数をインクリメント
    - `MarkSignUpCodeAsUsed`: used_atを更新（監査ログとして記録）
    - `InvalidateSignUpCodesByEmail`: 既存の未使用コードを無効化
  - 確認コード生成関数の実装（`internal/auth/verification_code.go`）
  - sqlcコード生成（`make sqlc-generate`）
  - 単体テストの実装
  - **想定ファイル数**: 約4ファイル（実装3 + テスト1）
  - **想定行数**: 約250行（実装180行 + テスト70行）

- [x] **1-2**: メール送信ジョブの実装
  - SendVerificationCodeEmailArgs 構造体の定義
  - SendVerificationCodeEmailWorker の実装
  - ワーカーの登録（internal/worker パッケージ）
  - 日本語メールテンプレートの作成（テキスト・HTML）
  - 英語メールテンプレートの作成（テキスト・HTML）
  - ユーザーロケールに応じたメール送信ロジック
  - 単体テストの実装
  - **想定ファイル数**: 約6ファイル（実装5 + テスト1）
  - **想定行数**: 約300行（実装220行 + テスト80行）

### フェーズ 2: サインアップフロー実装

- [x] **2-1**: メールアドレス入力とコード送信
  - ハンドラーの実装（`internal/handler/sign_up/handler.go`, `new.go`, `create.go`）
  - テンプレートの実装（`internal/templates/pages/sign_up/new.templ`）
  - バリデーションロジック（メールアドレス重複チェック、利用規約同意チェック）
  - Cloudflare Turnstile 検証
  - 確認コード生成・bcryptハッシュ化・データベース保存
  - メールアドレスをセッションに保存（`sign_up_email`）
  - メール送信ジョブのエンキュー
  - `/sign_up/code` にリダイレクト
  - 国際化メッセージの追加（ja.toml, en.toml）
  - 統合テストの実装
  - **想定ファイル数**: 約8ファイル（実装5 + テスト3）
  - **想定行数**: 約400行（実装280行 + テスト120行）

- [x] **2-2**: 確認コード検証
  - ハンドラーの実装（`internal/handler/sign_up_code/handler.go`, `show.go`, `create.go`, `update.go`）
  - テンプレートの実装（`internal/templates/pages/sign_up_code/show.templ`）
  - セッションから`sign_up_email`を取得（なければ `/sign_up` にリダイレクト）
  - 確認コード検証ロジック（データベースから取得、used_atチェック、bcrypt検証、試行回数チェック）
  - 検証成功時のused_at更新（監査ログとして記録）
  - 一時トークンを生成してRedisに保存（次のステップで使用）
  - 再送信機能の実装（`PATCH /sign_up/code`、既存の未使用コードを無効化して新規作成）
  - 国際化メッセージの追加
  - 統合テストの実装
  - **想定ファイル数**: 約7ファイル（実装5 + テスト2）
  - **想定行数**: 約400行（実装280行 + テスト120行）

- [x] **2-3**: ユーザー名設定とユーザー登録
  - ハンドラーの実装（`internal/handler/sign_up_username/handler.go`, `show.go`, `create.go`）
  - テンプレートの実装（`internal/templates/pages/sign_up_username/show.templ`）
  - 一時トークンの検証（Redisから取得）
  - ユーザー名バリデーション（形式チェック、一意性チェック）
  - ユーザー作成（usersテーブルへの挿入）
  - セッション作成（自動ログイン）
  - 一時トークン削除（Redis）
  - セッションから`sign_up_email`を削除
  - 国際化メッセージの追加
  - 統合テストの実装
  - **想定ファイル数**: 約7ファイル（実装5 + テスト2）
  - **想定行数**: 約400行（実装280行 + テスト120行）

- [x] **2-3-1**: 関連レコード作成（profiles, settings, email_notifications）
  - **目的**: Rails版の`build_relations`メソッドと同様に、ユーザー作成時に必要な関連レコードを作成する
  - **背景**: 現在の実装ではusersテーブルのレコードのみを作成しているが、Rails版では以下の関連レコードも作成している
    - `profiles`: ユーザープロフィール（名前、説明など）
    - `settings`: ユーザー設定（プライバシーポリシー同意など）
    - `email_notifications`: メール通知設定（配信停止キーなど）
  - **実装内容**:
    - sqlcクエリファイルの作成（`internal/repository/queries/profiles.sql`, `settings.sql`, `email_notifications.sql`）
      - `CreateProfile`: プロフィールレコードを作成（name: ユーザー名、description: 空文字列）
      - `CreateSetting`: 設定レコードを作成（privacy_policy_agreed: true、その他はデフォルト値）
      - `CreateEmailNotification`: メール通知設定レコードを作成（unsubscription_key: UUID）
    - ユーザー登録ハンドラーの更新（`internal/handler/sign_up_username/create.go`）
      - トランザクション内でuser、profile、setting、email_notificationを順次作成
      - エラー時は全てロールバック
    - UUIDライブラリの使用（`github.com/google/uuid`）
    - 統合テストの実装（関連レコードが正しく作成されているか確認）
  - **参考**: Rails版の実装
    - `app/models/user.rb` の `build_relations` メソッド（149-174行目）
    - `app/controllers/api/internal/registrations_controller.rb` の `create` アクション
  - **想定ファイル数**: 約5ファイル（実装3 + テスト2）
  - **想定行数**: 約300行（実装200行 + テスト100行）

- [x] **2-4**: `sign_up_code` ハンドラーの命名規則修正
  - **目的**: フォーム表示のハンドラーを `show.go` から `new.go` にリネーム（ハンドラー命名規則の統一）
  - ファイルのリネーム:
    - `internal/handler/sign_up_code/show.go` → `new.go`
    - `internal/handler/sign_up_code/show_test.go` → `new_test.go`（存在する場合）
    - `internal/templates/pages/sign_up_code/show.templ` → `new.templ`
  - ハンドラー関数名の変更: `Show` → `New`
  - テスト関数名の変更: `TestShow` → `TestNew`（存在する場合）
  - テンプレート関数名の変更: `Show` → `New`
  - `cmd/server/main.go` のルーティング更新: `signUpCodeHandler.Show` → `signUpCodeHandler.New`
  - templ コード再生成（`make templ-generate`）
  - テストの実行
  - **想定ファイル数**: 約4ファイル（実装4 + テスト0）
  - **想定行数**: 約50行（実装50行 + テスト0行）※リネームと関数名変更のみ

### フェーズ 3: セキュリティとUX改善

- [x] **3-1**: Rate Limiting実装
  - Rate Limiting ロジックの実装（既存の`internal/ratelimit`パッケージを使用）
  - IP単位の制限（5回/時間）
  - メールアドレス単位の制限（3回/時間）
  - コード検証の制限（10回/時間/IP）
  - 統合テストの実装
  - **想定ファイル数**: 約3ファイル（実装2 + テスト1）
  - **想定行数**: 約200行（実装120行 + テスト80行）

- [x] **3-2**: エラーハンドリングとバリデーション強化
  - セキュアなエラーメッセージの実装（情報漏洩対策）
  - バリデーションエラーの改善
  - 監査ログの実装（log/slog）
  - 統合テストの実装
  - **想定ファイル数**: 約4ファイル（実装3 + テスト1）
  - **想定行数**: 約250行（実装180行 + テスト70行）

- [x] **3-3**: 統合テストと本番準備
  - 新規登録フロー全体の統合テスト
  - セキュリティテスト（Rate Limiting、Turnstile、情報漏洩対策）
  - E2Eテストの実装（可能であれば）
  - リバースプロキシミドルウェアの更新（`/sign_up`, `/sign_up/code`, `/sign_up/username` パスを Go版で処理）
  - ドキュメントの更新（README、CLAUDE.md）
  - **想定ファイル数**: 約5ファイル（実装2 + テスト3）
  - **想定行数**: 約300行（実装100行 + テスト200行）

### フェーズ 4: 用語の統一

- [x] **4-1**: UI表示の用語統一
  - **目的**: ユーザー向けの表示を「新規登録」「ログインの確認コード」「新規登録の確認コード」に統一
  - テンプレートの更新（`internal/templates/pages/sign_up/*.templ`, `sign_up_code/*.templ`, `sign_up_username/*.templ`）
    - 「サインアップ」→「新規登録」
    - 「確認コード」→「新規登録の確認コード」
  - テンプレートの更新（`internal/templates/pages/sign_in_code/*.templ`）
    - 「ログインコード」→「ログインの確認コード」
  - 国際化メッセージの更新（`internal/i18n/locales/ja.toml`, `en.toml`）
    - `sign_up.title`: 「サインアップ」→「新規登録」
    - `sign_up_code.verification_code`: 「確認コード」→「新規登録の確認コード」
    - `sign_in_code.verification_code`: 「ログインコード」→「ログインの確認コード」
  - メールテンプレートの更新
    - 件名: 「Annict アカウント確認コード」→「Annict 新規登録の確認コード」
    - 本文: 「確認コード」→「新規登録の確認コード」
  - **想定ファイル数**: 約10ファイル（実装10 + テスト0）
  - **想定行数**: 約150行（実装150行 + テスト0行）

### 実装しない機能（スコープ外）

<!--
今回は実装しないが、将来的に検討する機能を明記
-->

以下の機能は今回の実装では**実装しません**：

- **OAuth認証（GitHub、Twitterなど）**: 実装の複雑さとセキュリティリスクが高いため見送り
- **パスワード設定**: メールアドレスログインのみに対応（将来的に検討）
- **電話番号による登録**: 実装の優先度が低いため見送り
- **プロフィール画像のアップロード**: サインアップ時ではなく、後から設定できるようにする
- **メールアドレス確認後の自動ログイン**: セキュリティ上、ユーザー名設定後に自動ログインする
- **ユーザー名のサジェスト機能**: 実装の優先度が低いため見送り

## 参考資料

<!--
参考にしたドキュメント、記事、OSSプロジェクトなど
-->

- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
- [Resend Documentation](https://resend.com/docs)
- [Cloudflare Turnstile Documentation](https://developers.cloudflare.com/turnstile/)
- [River Documentation](https://riverqueue.com/docs)

---

## テンプレート使用例

実際の使用例は以下を参照してください：

- [パスワードリセット機能](../3_done/202510/password-reset.md)

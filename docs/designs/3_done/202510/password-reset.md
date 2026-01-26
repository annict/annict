# パスワードリセット機能 設計書

## 概要

パスワードを忘れたユーザーがメールアドレスを入力し、受信したメール内のリンクから新しいパスワードを設定できる機能を実装します。

**目的**:

- ユーザーがパスワードを忘れた場合に、安全に新しいパスワードを設定できるようにする
- セキュリティを最優先に、OWASP 推奨のベストプラクティスに従った実装を行う

**背景**:

- Rails 版の Devise ベースの実装とは独立した、Go 版独自のパスワードリセット機能を構築する
- 既存の PostgreSQL データベースを共有しながら、Go 標準ライブラリを中心としたシンプルな実装を目指す

## 要件

### 機能要件

- ユーザーはメールアドレスを入力してパスワードリセットを申請できる
- システムはユーザーのメールアドレスにリセット用の安全なリンクを送信する
- ユーザーはメール内のリンクから新しいパスワードを設定できる
- リセットトークンは 1 時間の有効期限を持ち、使用後は無効化される
- パスワード変更後、ユーザーは自動的にログイン状態になる
- メールは日本語と英語の両方に対応し、ユーザーのロケール設定に応じて送信される

### 非機能要件

#### セキュリティ

- **トークン生成**: 暗号学的に安全な乱数生成器（crypto/rand）を使用し、最低 256 ビットの強度を確保
- **トークン保存**: SHA-256 でハッシュ化してデータベースに保存し、平文は保存しない
- **Rate Limiting**: ブルートフォース攻撃を防ぐため、IP 単位・メール単位での厳格な制限を実装
- **情報漏洩対策**: エラーメッセージでユーザーの存在やトークンの状態を明かさない
- **監査ログ**: セキュリティインシデントの調査のため、すべての重要なイベントをログに記録
- **パスワードポリシー**: NIST SP 800-63B-4 準拠（文字種の複雑性要件を廃止し、長さを重視）

#### パフォーマンス

- トークン検証はインデックスを使用した高速な検索を実現
- Rate Limiting にはインメモリストアの Redis を使用し、高速な制限チェックを実現
- バックグラウンドジョブ（River）により、メール送信が HTTP リクエストをブロックしない

#### ユーザビリティ（UX）

- わかりやすいエラーメッセージとヘルプテキストを提供
- パスワード要件を明確に表示（8文字以上、128文字以内、印字可能ASCII文字のみ）
- パスワードマネージャーの使用を推奨
- メールが届かない場合のサポート情報を提供
- パスワード変更後の自動ログインでスムーズな体験を実現

## 設計

### 技術スタック

- **メール送信**: resend-go/v2（Resend API）
- **バックグラウンドジョブ**: riverqueue/river
- **Rate Limiting**: Redis + go-redis/v9
- **トークン生成**: crypto/rand
- **トークンハッシュ化**: crypto/sha256
- **パスワードハッシュ化**: bcrypt（Devise 互換）

### データベース設計

#### password_reset_tokens テーブル

```sql
CREATE TABLE password_reset_tokens (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_digest VARCHAR(64) NOT NULL,  -- SHA-256ハッシュ（hex表現で64文字）
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used_at TIMESTAMP WITH TIME ZONE,   -- NULL = 未使用
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT unique_active_token_per_user UNIQUE (user_id, token_digest)
);

CREATE INDEX idx_password_reset_tokens_user_id ON password_reset_tokens(user_id);
CREATE INDEX idx_password_reset_tokens_token_digest ON password_reset_tokens(token_digest);
CREATE INDEX idx_password_reset_tokens_expires_at ON password_reset_tokens(expires_at);
```

**テーブル設計のポイント**:

- **user_id**: 外部キー制約でユーザー削除時にトークンも削除
- **token_digest**: トークンの SHA-256 ハッシュ（64 文字の hex 文字列）
- **expires_at**: トークンの有効期限（作成時刻 + 1 時間）
- **used_at**: トークン使用時刻（NULL = 未使用）
- **インデックス**: token_digest での検索を高速化

### API 設計

#### 1. GET /password/reset - リセット申請フォーム

パスワードリセット申請フォームを表示

**レスポンス**: 200 OK（HTML）

#### 2. POST /password/reset - リセットメール送信

メールアドレスを受け取り、パスワードリセットメールを送信

**リクエスト**:

```
POST /password/reset
Content-Type: application/x-www-form-urlencoded

email=user@example.com
csrf_token=xxx
```

**処理フロー**:

1. CSRF トークン検証
2. Rate Limiting チェック（IP、メールアドレス）
3. メールアドレスでユーザー検索
4. ユーザーが存在する場合:
   - トランザクション開始
   - 既存の未使用トークンを無効化
   - 新しいトークンを生成・保存
   - メール送信ジョブをエンキュー
   - トランザクションコミット
5. 常に「メールを送信しました」メッセージを表示（ユーザーの存在を明かさない）

**レスポンス**: 200 OK（HTML）

#### 3. GET /password/edit?token=xxx - 新パスワード入力フォーム

トークンを検証し、新しいパスワード入力フォームを表示

**リクエスト**:

```
GET /password/edit?token=abc123xyz
```

**処理フロー**:

1. トークンをハッシュ化
2. データベースからトークン情報を取得
3. トークンの検証（存在、有効期限、未使用）
4. 検証成功: パスワード入力フォームを表示
5. 検証失敗: 「無効なリンクです」エラーを表示

**レスポンス**: 200 OK（HTML）または 400 Bad Request（HTML）

#### 4. PUT /password - パスワード更新

新しいパスワードを設定し、トークンを無効化

**リクエスト**:

```
PUT /password
Content-Type: application/x-www-form-urlencoded

token=abc123xyz
password=NewPassword123!
password_confirmation=NewPassword123!
csrf_token=xxx
```

**処理フロー**:

1. CSRF トークン検証
2. トークンをハッシュ化
3. データベースからトークン情報を取得
4. トークンの検証（存在、有効期限、未使用）
5. パスワードのバリデーション（一致確認、長さチェック、文字種チェック）
   - NIST SP 800-63B-4 準拠: 8文字以上、128文字以内をチェック
   - 印字可能ASCII文字のみを許可（スペース・Unicode文字は不可）
   - 文字種の複雑性要件（大文字・小文字・数字・記号の組み合わせ）はチェックしない
6. トランザクション開始
7. パスワードを bcrypt でハッシュ化して users テーブルを更新
8. トークンを無効化（used_at を設定）
9. 新しいセッションを作成してログイン状態にする
10. トランザクションコミット
11. ホームページにリダイレクト

**レスポンス**: 302 Found（リダイレクト）または 400 Bad Request（HTML）

### セキュリティ設計

#### 1. トークン生成

**要件**:

- 暗号学的に安全な乱数生成器（`crypto/rand`）を使用
- トークン長: 最低 32 バイト（256 ビット）
- エンコード: URL セーフな Base64（base64.RawURLEncoding）

**実装例**:

```go
import (
    "crypto/rand"
    "encoding/base64"
)

func generateResetToken() (string, error) {
    b := make([]byte, 32) // 256 bits
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    return base64.RawURLEncoding.EncodeToString(b), nil
}
```

#### 2. トークン保存

**要件**:

- SHA-256 でハッシュ化してデータベースに保存
- トークンの平文はメール送信時のみ使用し、保存しない

**実装例**:

```go
import (
    "crypto/sha256"
    "encoding/hex"
)

func hashToken(token string) string {
    hash := sha256.Sum256([]byte(token))
    return hex.EncodeToString(hash[:])
}
```

#### 3. 有効期限とライフサイクル

**要件**:

- 有効期限: 1 時間（OWASP 推奨: "rarely more than an hour"）
- 1 回限り使用: トークン使用後は `used_at` を記録して無効化（レコードは削除しない）
- 古いトークンの無効化: 同一ユーザーが新しいリセットを申請した場合、未使用の古いトークンをすべて無効化

**トークンの状態管理**:

```go
// トークンの状態チェック
func (t *PasswordResetToken) IsValid() bool {
    // 使用済みチェック
    if t.UsedAt != nil {
        return false
    }

    // 有効期限チェック
    if time.Now().After(t.ExpiresAt) {
        return false
    }

    return true
}
```

**クリーンアップ処理**:

有効期限切れおよび使用済みトークンは、River のスケジュールジョブで定期的に削除します。

実装例:

```go
// トークンクリーンアップジョブの引数
type CleanupExpiredTokensArgs struct{}

func (CleanupExpiredTokensArgs) Kind() string {
    return "cleanup_expired_tokens"
}

// トークンクリーンアップワーカー
type CleanupExpiredTokensWorker struct {
    river.WorkerDefaults[CleanupExpiredTokensArgs]
    queries *repository.Queries
}

func (w *CleanupExpiredTokensWorker) Work(ctx context.Context, job *river.Job[CleanupExpiredTokensArgs]) error {
    // 24時間以上前に期限切れまたは使用済みになったトークンを削除
    cutoff := time.Now().Add(-24 * time.Hour)
    return w.queries.DeleteExpiredPasswordResetTokens(ctx, cutoff)
}

// スケジュールジョブの登録（main.go）
_, err := riverClient.PeriodicJobs().Add(&river.PeriodicJobArgs{
    ScheduleFunc: func() time.Time {
        // 毎日深夜2時に実行
        now := time.Now()
        next := time.Date(now.Year(), now.Month(), now.Day()+1, 2, 0, 0, 0, now.Location())
        if now.Hour() >= 2 {
            // 今日の2時を過ぎていたら明日の2時
            next = next.Add(24 * time.Hour)
        }
        return next
    },
    ConstructorFunc: func() (river.JobArgs, *river.InsertOpts, error) {
        return CleanupExpiredTokensArgs{}, nil, nil
    },
})
```

**SQL クエリ**:

```sql
-- queries/password_reset_tokens.sql
-- name: DeleteExpiredPasswordResetTokens :exec
DELETE FROM password_reset_tokens
WHERE (expires_at < $1 OR used_at < $1);
```

#### 4. Rate Limiting（レート制限）

ブルートフォース攻撃を防ぐため、以下の制限を実装します。

**制限ルール**:

| 制限対象           | 制限値        | 理由                                     |
| ------------------ | ------------- | ---------------------------------------- |
| IP アドレス単位    | 5 回/時間     | 同一 IP からのリセット申請を制限         |
| メールアドレス単位 | 3 回/時間     | 同一メールアドレスへの送信を制限         |
| トークン検証       | 10 回/時間/IP | URL のトークンを総当たりで試す攻撃を防ぐ |

**実装方針**:

Redis を使用して Rate Limiting を実装します。

- インフラ: Rails が既に使用している Redis を活用（追加コスト不要）
- クライアント: github.com/redis/go-redis/v9
- namespace: `rate_limit:*` で Rails のキャッシュと分離
- アルゴリズム: Sliding Window Counter（Lua スクリプトでアトミック操作）
- TTL: 自動期限切れで古いデータを削除（クリーンアップ不要）

**Redis を選択した理由**:

1. 既存インフラの活用: Rails が既に Redis を使っているため、追加インフラ不要
2. 複数プロセス対応: Dokku で web プロセスを複数起動しても正確な制限
3. パフォーマンス: インメモリストアで高速（PostgreSQL より優位）
4. 将来の拡張性: Web API の Rate Limiting でも同じ仕組みを使える
5. 業界標準: Web API の Rate Limiting では Redis が最も一般的

実装例:

```go
import (
    "context"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
)

// Redis クライアント初期化
func NewRedisClient() *redis.Client {
    return redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
        DB:   0, // Rails と同じ DB を使用
    })
}

// Rate Limiting チェック（Lua スクリプトでアトミック操作）
func checkRateLimit(ctx context.Context, rdb *redis.Client, key string, limit int, window time.Duration) (bool, error) {
    script := `
        local current = redis.call('INCR', KEYS[1])
        if current == 1 then
            redis.call('EXPIRE', KEYS[1], ARGV[1])
        end
        return current
    `

    fullKey := fmt.Sprintf("rate_limit:%s", key)
    count, err := rdb.Eval(ctx, script, []string{fullKey}, int(window.Seconds())).Int()
    if err != nil {
        return false, err
    }

    return count <= limit, nil
}

// 使用例: IP アドレス単位の制限
func (h *Handler) ProcessPasswordReset(w http.ResponseWriter, r *http.Request) {
    ip := getClientIP(r)
    key := fmt.Sprintf("password_reset:ip:%s", ip)

    allowed, err := checkRateLimit(r.Context(), h.redis, key, 5, 1*time.Hour)
    if err != nil {
        // エラーハンドリング
        return
    }

    if !allowed {
        // 429 Too Many Requests
        http.Error(w, "リクエストが多すぎます。しばらくしてから再度お試しください。", http.StatusTooManyRequests)
        return
    }

    // パスワードリセット処理
}
```

**Rate Limiting 超過時のレスポンス**:

- HTTP ステータスコード: 429 Too Many Requests
- エラーメッセージ: 「リクエストが多すぎます。しばらくしてから再度お試しください。」（シンプルで一般的なメッセージ）
- Retry-After ヘッダー: 実装しない（セキュリティ上、次に試行可能な時刻は明かさない）

#### 5. パスワード強度チェック

**要件（NIST SP 800-63B-4 準拠）**:

新しいパスワードは以下の条件を満たす必要があります：

- **最小文字数**: 8 文字以上
  - NIST SP 800-63B-4 では多要素認証時: 8文字、単一要素認証時: 15文字を推奨
  - 本実装では、パスワードリセット後に自動ログインするため8文字を採用
- **最大文字数**: 128 文字まで（NIST は64文字以上を推奨）
- **使用可能文字**: 印字可能ASCII文字のみ（0x21～0x7E）
  - 英小文字（a-z）、英大文字（A-Z）、数字（0-9）、記号（!@#$%^&*() など）
  - **スペースとUnicode文字は許可しない**（実装の複雑さを避けるため）
- **文字種の組み合わせ要求**: **禁止** （NIST SP 800-63B-4 で "SHALL NOT" と明記）
  - 従来の「大文字・小文字・数字・記号を必ず含む」という要件は有害とされ廃止

**重要**: NIST SP 800-63B-4 では、文字種の複雑性要件が明確に禁止されています。代わりに「長さ」を重視し、ユーザーがパスワードマネージャーで生成した長く一意なパスワードを使いやすくすることが推奨されています。

実装例:

```go
import (
    "errors"
    "strings"
)

// ValidatePasswordStrength はパスワードの強度をチェックします
// NIST SP 800-63B-4 準拠
func ValidatePasswordStrength(password string) error {
    // 最小文字数チェック
    if len(password) < 8 {
        return errors.New("パスワードは8文字以上である必要があります")
    }

    // 最大文字数チェック
    if len(password) > 128 {
        return errors.New("パスワードは128文字以内である必要があります")
    }

    // 印字可能ASCII文字のみを許可（0x21～0x7E）
    for _, char := range password {
        if char < 0x21 || char > 0x7E {
            return errors.New("パスワードは印字可能なASCII文字のみ使用できます")
        }
    }

    return nil
}
```

**将来の拡張検討事項**:

- 漏洩パスワードのブロックリスト実装（Have I Been Pwned API など）
- 辞書語、サービス名のブロックリスト
- ユーザー名と同一のパスワード禁止

**参考資料**:

- [NIST SP 800-63B-4 解説（崎村夏彦氏）](https://www.sakimura.org/2025/10/7710/)
- [NIST SP 800-63B-4 公式ドキュメント](https://pages.nist.gov/800-63-4/)

#### 6. 情報漏洩対策

**セキュアなエラーメッセージ**:

システムの内部情報を漏らさないため、常に一般的なメッセージを返します。

| シナリオ             | NG（情報漏洩）                             | OK（セキュア）                             |
| -------------------- | ------------------------------------------ | ------------------------------------------ |
| メールアドレス未登録 | 「このメールアドレスは登録されていません」 | 「パスワードリセットメールを送信しました」 |
| トークン期限切れ     | 「トークンの有効期限が切れています」       | 「無効なリンクです」                       |
| トークン使用済み     | 「このトークンは既に使用されています」     | 「無効なリンクです」                       |
| トークン不正         | 「トークンが見つかりません」               | 「無効なリンクです」                       |

実装例:

```go
// リセット申請（常に成功メッセージを返す）
func (h *Handler) ProcessPasswordReset(w http.ResponseWriter, r *http.Request) {
    email := r.FormValue("email")

    // ユーザー検索（存在しなくてもエラーを返さない）
    user, _ := h.queries.GetUserByEmail(ctx, email)

    if user != nil {
        // トークン生成とメール送信ジョブをエンキュー
        // ...
    }

    // 常に同じメッセージを返す（ユーザーの存在を明かさない）
    h.renderTemplate(w, "password_reset_sent.html", data)
}
```

### バックグラウンドジョブ設計

#### メール送信ジョブ（River）

**ジョブ引数**:

```go
type SendPasswordResetEmailArgs struct {
    UserID int64  `json:"user_id"`
    Token  string `json:"token"` // 平文トークン（メール本文に含めるため）
}

func (SendPasswordResetEmailArgs) Kind() string {
    return "send_password_reset_email"
}
```

**ワーカー実装**:

```go
import (
    "context"
    "fmt"
    "github.com/resend/resend-go/v2"
)

type SendPasswordResetEmailWorker struct {
    river.WorkerDefaults[SendPasswordResetEmailArgs]
    resendClient *resend.Client
    fromEmail    string
}

func (w *SendPasswordResetEmailWorker) Work(ctx context.Context, job *river.Job[SendPasswordResetEmailArgs]) error {
    // ユーザー情報を取得
    user, err := w.queries.GetUserByID(ctx, job.Args.UserID)
    if err != nil {
        return fmt.Errorf("failed to get user: %w", err)
    }

    // リセットリンクを生成
    resetURL := fmt.Sprintf("https://go.example.dev/password/edit?token=%s", job.Args.Token)

    // メール送信（Resend API）
    params := &resend.SendEmailRequest{
        From:    w.fromEmail,
        To:      []string{user.Email},
        Subject: "パスワードリセットのご案内",
        Text: fmt.Sprintf(`パスワードリセットのリクエストを受け付けました。

以下のリンクをクリックして、新しいパスワードを設定してください：
%s

このリンクは1時間有効です。

このメールに心当たりがない場合は、無視してください。
`, resetURL),
    }

    _, err = w.resendClient.Emails.Send(params)
    if err != nil {
        return fmt.Errorf("failed to send email: %w", err)
    }

    return nil
}
```

**トランザクション内でのエンキュー**:

```go
tx, _ := db.Begin(ctx)
defer tx.Rollback()

// トークン生成
token, _ := generateResetToken()
tokenDigest := hashToken(token)

// トークンをDBに保存
_, err := queries.CreatePasswordResetToken(ctx, tx, repository.CreatePasswordResetTokenParams{
    UserID:      user.ID,
    TokenDigest: tokenDigest,
    ExpiresAt:   time.Now().Add(1 * time.Hour),
})

// メール送信ジョブをエンキュー（トランザクション成功時のみ実行）
_, err = riverClient.InsertTx(ctx, tx, SendPasswordResetEmailArgs{
    UserID: user.ID,
    Token:  token, // 平文トークン
}, nil)

tx.Commit() // ここで成功すればジョブが確実にエンキューされる
```

### メール送信設計

#### Resend API 設定（環境変数）

秘密情報は `.env.*.local` ファイルで管理します。

```bash
# .env.development.local（開発環境）
ANNICT_RESEND_API_KEY=re_xxxxx
ANNICT_RESEND_FROM_EMAIL=noreply@annict.com

# .env.test.local（テスト環境）
ANNICT_RESEND_API_KEY=re_test_dummy_key
ANNICT_RESEND_FROM_EMAIL=test@annict.com
```

**セットアップ手順**:

1. サンプルファイルをコピー：

   ```bash
   cp .env.development.local.example .env.development.local
   cp .env.test.local.example .env.test.local
   ```

2. `.env.development.local` を開き、実際の Resend API キーを設定

**注意**: `.env.*.local` ファイルは `.gitignore` に登録されており、バージョン管理に含まれません。

**Resend の特徴**:

- **シンプルな API**: RESTful API で簡単にメール送信
- **開発者フレンドリー**: 無料プランで月 100 通まで送信可能
- **信頼性**: 高い到達率とパフォーマンス
- **SMTP 不要**: API 経由で直接送信（SMTP サーバーの管理が不要）

#### メールテンプレート

メールは日本語と英語の両方に対応し、ユーザーのロケール設定に応じて送信します。

**テキストメール（日本語）**:

```
件名: パスワードリセットのご案内

本文:
パスワードリセットのリクエストを受け付けました。

以下のリンクをクリックして、新しいパスワードを設定してください：
https://go.example.dev/password/edit?token=abc123xyz

このリンクは1時間有効です。

このメールに心当たりがない場合は、無視してください。
```

**テキストメール（英語）**:

```
Subject: Password Reset Request

Body:
We have received a request to reset your password.

Please click the link below to set a new password:
https://go.example.dev/password/edit?token=abc123xyz

This link is valid for 1 hour.

If you did not request this, please ignore this email.
```

**HTML メール（日本語）**:

```html
<!DOCTYPE html>
<html lang="ja">
  <head>
    <meta charset="UTF-8" />
    <title>パスワードリセットのご案内</title>
  </head>
  <body>
    <h2>パスワードリセットのご案内</h2>
    <p>パスワードリセットのリクエストを受け付けました。</p>
    <p>以下のボタンをクリックして、新しいパスワードを設定してください：</p>
    <p>
      <a
        href="https://go.example.dev/password/edit?token=abc123xyz"
        style="background-color: #007bff; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px; display: inline-block;"
        >パスワードをリセット</a
      >
    </p>
    <p>このリンクは1時間有効です。</p>
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
    <title>Password Reset Request</title>
  </head>
  <body>
    <h2>Password Reset Request</h2>
    <p>We have received a request to reset your password.</p>
    <p>Please click the button below to set a new password:</p>
    <p>
      <a
        href="https://go.example.dev/password/edit?token=abc123xyz"
        style="background-color: #007bff; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px; display: inline-block;"
        >Reset Password</a
      >
    </p>
    <p>This link is valid for 1 hour.</p>
    <p>If you did not request this, please ignore this email.</p>
  </body>
</html>
```

**多言語対応の実装**:

```go
// ユーザーのロケールに応じてメールを送信
func (w *SendPasswordResetEmailWorker) Work(ctx context.Context, job *river.Job[SendPasswordResetEmailArgs]) error {
    user, _ := w.queries.GetUserByID(ctx, job.Args.UserID)

    // ユーザーのロケール（"ja" または "en"）
    locale := user.Locale
    if locale == "" {
        locale = "ja" // デフォルトは日本語
    }

    // ロケールに応じた件名と本文を取得
    subject := i18n.T(ctx, "password_reset_email_subject", locale)
    textBody := renderEmailTemplate(locale, "password_reset_text", map[string]string{
        "ResetURL": resetURL,
    })

    // メール送信（Resend API）
    params := &resend.SendEmailRequest{
        From:    w.fromEmail,
        To:      []string{user.Email},
        Subject: subject,
        Text:    textBody,
    }

    _, err := w.resendClient.Emails.Send(params)
    return err
}
```

### 監査ログ設計

セキュリティインシデントの調査のため、以下をログに記録します。

**ログ項目**:

| イベント               | ログレベル | 記録内容                                                   |
| ---------------------- | ---------- | ---------------------------------------------------------- |
| パスワードリセット申請 | Info       | ユーザー ID、メールアドレス、IP アドレス、成功/失敗        |
| Rate Limiting 発動     | Warning    | IP アドレス、メールアドレス、制限種別                      |
| トークン検証失敗       | Warning    | トークン（ハッシュ化）、失敗理由（期限切れ/使用済み/不正） |
| パスワード変更成功     | Info       | ユーザー ID、IP アドレス                                   |
| パスワード変更失敗     | Warning    | ユーザー ID、IP アドレス、失敗理由                         |

**実装例（log/slog）**:

```go
import "log/slog"

// パスワードリセット申請
slog.InfoContext(ctx, "password reset requested",
    "user_id", user.ID,
    "email", user.Email,
    "ip_address", getClientIP(r),
)

// トークン検証失敗
slog.WarnContext(ctx, "invalid password reset token",
    "token_digest", tokenDigest,
    "reason", "expired",
    "ip_address", getClientIP(r),
)

// パスワード変更成功
slog.InfoContext(ctx, "password changed successfully",
    "user_id", user.ID,
    "ip_address", getClientIP(r),
)
```

### UX 設計

#### 1. わかりやすいメッセージ

- リセット申請後: 「メールを確認してください。パスワードリセット用のリンクを送信しました。」
- メールが届かない場合のヘルプ: 「メールが届かない場合は、迷惑メールフォルダを確認してください。」
- 再送信リンク: 「メールが届かない場合は、[こちら]から再送信できます。」
- パスワード要件の表示: 「8文字以上、128文字以内、印字可能なASCII文字のみ使用可能です。パスワードマネージャーの使用を推奨します。」
  - **重要**: 従来の「大文字・小文字・数字・記号を含む」という表示は削除（NIST SP 800-63B-4 準拠）

#### 2. 成功後の自動ログイン

パスワード変更後、自動的にログイン状態にしてホームページにリダイレクト：

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

#### テスト用 Redis インスタンス

テストでは本番環境の Redis とは分離された**テスト用 Redis インスタンス**を使用します。

**docker-compose.yml に追加**:

```yaml
services:
  redis-test:
    image: redis:7-alpine
    ports:
      - "16379:6379"
    command: redis-server --save "" --appendonly no # 永続化無効（高速化）
```

**テストヘルパー（internal/testutil/redis.go）**:

```go
package testutil

import (
    "context"
    "testing"

    "github.com/redis/go-redis/v9"
)

// SetupTestRedis はテスト用 Redis クライアントをセットアップし、
// テスト終了時に自動的にクリーンアップします
func SetupTestRedis(t *testing.T) *redis.Client {
    t.Helper()

    // テスト用 Redis に接続（ポート 16379）
    rdb := redis.NewClient(&redis.Options{
        Addr: "localhost:16379",
        DB:   0,
    })

    // 接続確認
    ctx := context.Background()
    if err := rdb.Ping(ctx).Err(); err != nil {
        t.Fatalf("failed to connect to test Redis: %v", err)
    }

    // テスト終了時にデータをクリーンアップ
    t.Cleanup(func() {
        // テストで使用したキーをすべて削除
        rdb.FlushDB(ctx)
        rdb.Close()
    })

    return rdb
}
```

環境変数（`.env.test`）:

```bash
ANNICT_REDIS_URL=redis://localhost:16379/0
```

#### 単体テスト

トークン生成・ハッシュ化のテスト:

```go
func TestGenerateResetToken(t *testing.T) {
    token, err := generateResetToken()
    if err != nil {
        t.Fatalf("failed to generate token: %v", err)
    }

    // トークン長の検証（Base64エンコード後の長さ）
    if len(token) < 43 { // 32バイト → 43文字（Base64 RawURL）
        t.Errorf("token too short: %d", len(token))
    }
}

func TestHashToken(t *testing.T) {
    token := "test-token"
    hash1 := hashToken(token)
    hash2 := hashToken(token)

    // 同じトークンは同じハッシュを生成
    if hash1 != hash2 {
        t.Errorf("hash mismatch: %s != %s", hash1, hash2)
    }

    // ハッシュ長の検証（SHA-256 → 64文字のhex）
    if len(hash1) != 64 {
        t.Errorf("hash length should be 64, got %d", len(hash1))
    }
}
```

#### 統合テスト

パスワードリセットフローのテスト:

```go
func TestPasswordResetFlow(t *testing.T) {
    db, tx := testutil.SetupTestDB(t)
    rdb := testutil.SetupTestRedis(t)

    // テストユーザー作成
    userID := testutil.NewUserBuilder(t, tx).
        WithEmail("test@example.com").
        Build()

    // 1. リセット申請
    req := httptest.NewRequest("POST", "/password/reset", strings.NewReader("email=test@example.com"))
    rr := httptest.NewRecorder()
    handler.ProcessPasswordReset(rr, req)

    // 2. トークンがDBに保存されているか確認
    token, err := queries.GetLatestPasswordResetToken(ctx, tx, userID)
    if err != nil {
        t.Fatalf("token not found: %v", err)
    }

    // 3. トークンが有効か確認
    if !token.IsValid() {
        t.Error("token should be valid")
    }

    // 4. パスワード更新
    req = httptest.NewRequest("PUT", "/password", strings.NewReader(
        fmt.Sprintf("token=%s&password=NewPassword123!&password_confirmation=NewPassword123!", token.Token),
    ))
    rr = httptest.NewRecorder()
    handler.UpdatePassword(rr, req)

    // 5. トークンが使用済みになっているか確認
    token, _ = queries.GetPasswordResetToken(ctx, tx, token.TokenDigest)
    if token.UsedAt == nil {
        t.Error("token should be marked as used")
    }
}
```

#### セキュリティテスト

Rate Limiting のテスト:

```go
func TestRateLimiting(t *testing.T) {
    db, tx := testutil.SetupTestDB(t)
    rdb := testutil.SetupTestRedis(t)  // テスト用 Redis

    queries := repository.New(db).WithTx(tx)
    handler := &Handler{
        queries: queries,
        redis:   rdb,  // テスト用 Redis を使用
    }

    // 同一IPから6回リセット申請（制限: 5回/時間）
    for i := 0; i < 6; i++ {
        req := httptest.NewRequest("POST", "/password/reset", strings.NewReader("email=test@example.com"))
        req.RemoteAddr = "203.0.113.1:12345"
        rr := httptest.NewRecorder()
        handler.ProcessPasswordReset(rr, req)

        if i < 5 {
            // 最初の5回は成功
            if rr.Code != http.StatusOK {
                t.Errorf("request %d should succeed", i+1)
            }
        } else {
            // 6回目は失敗（Rate Limit）
            if rr.Code != http.StatusTooManyRequests {
                t.Errorf("request %d should be rate limited", i+1)
            }
        }
    }

    // テスト終了時に Redis のデータは自動的にクリーンアップされる
}
```

テストのポイント:

- `testutil.SetupTestRedis(t)` でテスト用 Redis を初期化
- 本番環境の Redis（ポート 6379）とは分離された環境（ポート 16379）
- テスト終了時に自動的に `FlushDB` でデータをクリーンアップ
- 並列テスト（`t.Parallel()`）も安全に実行可能

### マイグレーション管理

#### dbmate の採用

Go プロジェクトでのマイグレーション管理には **dbmate** を採用します。

選定理由:

- シンプル: Rails に近い思想（マイグレーションファイル + structure.sql）
- language-agnostic: Go に依存しない設計
- `structure.sql` 管理: Rails と同様に DB スキーマをダンプできる
- Go らしいシンプルさ: 過度な抽象化がない

基本的な使い方:

```bash
# 新しいマイグレーションを作成
dbmate new create_password_reset_tokens

# マイグレーションを実行
dbmate up

# ロールバック
dbmate down

# structure.sql をダンプ
dbmate dump
```

### 実装方針

#### Go 版独自のテーブルを作成

Rails 側の Devise とは独立した `password_reset_tokens` テーブルを作成します。

理由:

- Rails との干渉なし: Rails 側のパスワードリセット機能と競合しない
- 実装の自由度: Go の標準的な実装方法で自由に設計できる
- 保守性: Go 版が本流になった後、Rails のコードを削除しやすい
- シンプルさ: Devise の実装に依存せず、Go の標準ライブラリで完結

注意点:

- Go 版と Rails 版でパスワードリセット機能が並存する期間がありますが、それぞれ独立して動作するため問題ありません
- ユーザーは Go 版のリセットフォーム（`go.example.dev/password/reset`）を使用します

## タスクリスト

### フェーズ 1: インフラ準備

- [x] dbmate 導入
  - **想定ファイル数**: 約3ファイル（実装3 + テスト0）
  - **想定行数**: 約50行（実装50行 + テスト0行）

- [x] Resend 導入
  - **想定ファイル数**: 約2ファイル（実装2 + テスト0）
  - **想定行数**: 約30行（実装30行 + テスト0行）

### フェーズ 2: データベーステーブル作成

- [x] password_reset_tokensテーブルのマイグレーション作成・実行
  - **想定ファイル数**: 約2ファイル（実装2 + テスト0）
  - **想定行数**: 約50行（実装50行 + テスト0行）

- [x] sqlcクエリの定義と生成
  - **想定ファイル数**: 約2ファイル（実装2 + テスト0）
  - **想定行数**: 約100行（実装100行 + テスト0行）

### フェーズ 3: 基本機能実装（メール送信なし）

- [x] トークン生成・ハッシュ化の実装
  - **想定ファイル数**: 約2ファイル（実装2 + テスト0）
  - **想定行数**: 約80行（実装80行 + テスト0行）

- [x] リセット申請フォームとリセット申請処理（GET/POST /password/reset）
  - **想定ファイル数**: 約5ファイル（実装5 + テスト0）
  - **想定行数**: 約200行（実装200行 + テスト0行）

- [x] 新パスワード入力フォームとパスワード更新処理（GET /password/edit、PUT /password）
  - **想定ファイル数**: 約5ファイル（実装5 + テスト0）
  - **想定行数**: 約250行（実装250行 + テスト0行）

### フェーズ 4: セキュリティ機能

- [x] Rate Limiting実装（Redis + テストヘルパー）
  - **想定ファイル数**: 約4ファイル（実装4 + テスト0）
  - **想定行数**: 約180行（実装180行 + テスト0行）

- [x] 監査ログ実装（log/slog）
  - **想定ファイル数**: 約2ファイル（実装2 + テスト0）
  - **想定行数**: 約40行（実装40行 + テスト0行）

- [x] パスワード強度チェックとセキュアなエラーメッセージ
  - **想定ファイル数**: 約3ファイル（実装3 + テスト0）
  - **想定行数**: 約120行（実装120行 + テスト0行）

### フェーズ 5: River セットアップとメール送信ジョブ

- [x] River ライブラリ導入とマイグレーション
  - riverqueue/river ライブラリの導入（go.mod、go.sum）
  - River テーブルのマイグレーション実行（river_job, river_leader など）
  - River クライアントの初期化（internal/worker パッケージ）
  - cmd/server/main.go での River クライアント初期化
  - 環境変数設定（.env.development、.env.test）
  - **想定ファイル数**: 約5ファイル（実装5 + テスト0）
  - **想定行数**: 約200行（実装200行 + テスト0行）

- [x] メール送信ジョブ実装（日本語のみ）
  - SendPasswordResetEmailArgs 構造体の定義
  - SendPasswordResetEmailWorker の実装
  - ワーカーの登録（internal/worker パッケージ）
  - 日本語メールテンプレートの作成（テキスト形式）
  - リセット申請処理にトランザクション内ジョブエンキューを追加
  - テスト実装（ジョブエンキューの確認）
  - **想定ファイル数**: 約8ファイル（実装5 + テスト3）
  - **想定行数**: 約350行（実装250行 + テスト100行）

- [x] 英語メール対応とHTMLメール実装
  - 英語メールテンプレートの作成（テキスト・HTML）
  - 日本語HTMLメールテンプレートの追加
  - ユーザーロケールに応じたメール送信ロジック
  - マルチパートメール（テキスト + HTML）の実装
  - **想定ファイル数**: 約6ファイル（実装5 + テスト1）
  - **想定行数**: 約200行（実装150行 + テスト50行）

### フェーズ 6: トークンクリーンアップとテスト

- [x] トークンクリーンアップジョブ実装
  - CleanupExpiredTokensArgs 構造体の定義
  - CleanupExpiredTokensWorker の実装
  - 定期実行スケジュールの設定（毎日深夜2時）
  - DeleteExpiredPasswordResetTokens SQLクエリの実装
  - **想定ファイル数**: 約5ファイル（実装3 + テスト2）
  - **想定行数**: 約170行（実装120行 + テスト50行）

- [x] 統合テストとセキュリティテスト
  - パスワードリセットフローの統合テスト
  - トークン有効期限と使用後の無効化テスト
  - メール送信ジョブのテスト
  - **想定ファイル数**: 約3ファイル（実装0 + テスト3）
  - **想定行数**: 約200行（実装0行 + テスト200行）

- [x] UX改善
  - 「メールが届かない場合」ヘルプメッセージの追加
  - テンプレートの改善（わかりやすいメッセージ）
  - 国際化対応の確認
  - **想定ファイル数**: 約4ファイル（実装3 + テスト1）
  - **想定行数**: 約100行（実装80行 + テスト20行）

### フェーズ 7: パスワードポリシーの見直し（NIST SP 800-63B-4 準拠）

- [x] パスワード強度チェックロジックの変更
  - 文字種の複雑性要件を削除（大文字・小文字・数字・記号の組み合わせ要求を廃止）
  - 最小文字数: 8文字（変更なし）
  - 最大文字数: 128文字に設定
  - 印字可能ASCII文字のみを許可（0x21～0x7E）
    - スペースとUnicode文字は許可しない（実装の複雑さを避けるため）
  - `internal/auth/password.go`のValidatePasswordStrength関数を簡素化
  - **想定ファイル数**: 約2ファイル（実装1 + テスト1）
  - **想定行数**: 約50行（実装30行 + テスト20行）

- [x] バリデーションエラーメッセージとテンプレートの更新
  - バリデーションエラーメッセージの更新（文字種要求の削除）
  - テンプレートのヘルプテキスト更新（パスワードマネージャー使用推奨を追加）
  - 国際化メッセージの更新（ja.toml, en.toml）
  - パスワード入力フォームのプレースホルダーや説明文の更新
  - **想定ファイル数**: 約5ファイル（実装4 + テスト1）
  - **想定行数**: 約80行（実装60行 + テスト20行）

- [x] テストケースの更新
  - パスワード強度チェックのテストケース更新（複雑性要件のテストを削除）
  - 新しいバリデーションルールのテストケース追加（最大文字数128文字など）
  - 統合テストの更新（エラーメッセージの変更に対応）
  - **想定ファイル数**: 約2ファイル（実装0 + テスト2）
  - **想定行数**: 約30行（実装0行 + テスト30行）

- [x] パスワード強度チェック関数の国際化と統合
  - **目的**: 将来的なユーザー登録機能でも使いやすくするため、`internal/auth/password.go` に統一
  - `internal/auth/password.go` の `ValidatePasswordStrength` を context を受け取るように変更
    - シグネチャ: `func ValidatePasswordStrength(ctx context.Context, password string) error`
    - エラーメッセージを i18n.T(ctx, "翻訳キー") に変更
    - 翻訳キー: `password_strength_min_length`, `password_strength_max_length`, `password_strength_invalid_chars`
  - `internal/password_reset/validation.go` の `ValidatePasswordStrength` を削除
  - `internal/handler/password_reset.go` での呼び出しを `auth.ValidatePasswordStrength(ctx, password)` に変更
  - `internal/auth/password_test.go` のテストを更新
    - context を渡すように修正
    - 国際化メッセージのテスト（日本語・英語）を追加
  - `internal/password_reset/validation_test.go` を削除（テストは `internal/auth/password_test.go` に統合）
  - **想定ファイル数**: 約4ファイル（実装3 + テスト1）
  - **想定行数**: 約100行（実装60行 + テスト40行）

**参考資料**:
- [NIST SP 800-63B-4 解説（崎村夏彦氏）](https://www.sakimura.org/2025/10/7710/)
- [NIST SP 800-63B-4 公式ドキュメント](https://pages.nist.gov/800-63-4/)

**注意事項**:
- この変更により、既存ユーザーが設定済みのパスワード（文字種要件を満たすもの）はそのまま有効
- 新しく設定するパスワード、またはリセット後のパスワードのみが新ポリシーに従う
- パスワードマネージャーで生成した長く複雑なパスワードが使いやすくなる

### 実装しない機能（スコープ外）

以下の機能は今回のパスワードリセット実装では実装しません：

- **パスワード強度インジケーター**: フロントエンドの実装が複雑になるため見送り
  - 代わりにサーバーサイドでのパスワード強度チェックとエラーメッセージで対応
- **全セッション無効化**: 他のデバイスでログアウトされることによる UX 悪化や、警告メッセージの実装など考慮すべき点が多いため見送り
  - パスワード変更後は新しいセッションを作成してログイン状態にする
- **パスワード変更後のメール通知**: 実装の優先度が低いため見送り（将来的に検討）
- **CAPTCHA/Turnstile**: 今回は見送り（フェーズ 7 で Turnstile 導入後に検討）

## 参考資料

### セキュリティ・認証

- [OWASP Forgot Password Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Forgot_Password_Cheat_Sheet.html)
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
- [NIST SP 800-63B-4: Digital Identity Guidelines - Authentication and Lifecycle Management](https://pages.nist.gov/800-63-4/)
- [NIST SP 800-63B-4 解説（崎村夏彦氏）](https://www.sakimura.org/2025/10/7710/)

### ライブラリ・ツール

- [riverqueue/river GitHub](https://github.com/riverqueue/river)
- [River Documentation](https://riverqueue.com/docs)
- [Resend Documentation](https://resend.com/docs)

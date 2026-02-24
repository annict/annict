# ログインページ

## 概要

メールアドレス（またはユーザー名）とパスワードによるログイン機能の実装。
Rails の Devise と互換性のある bcrypt パスワードハッシュを使用し、
PostgreSQL の sessions テーブルでセッション管理を行う。

## 実装内容

### フェーズ 5: ログイン機能

- [x] **GET /sign_in エンドポイント**
  - [x] ログインフォーム表示
  - [x] CSRF トークンの埋め込み

- [x] **POST /sign_in 認証処理**
  - [x] メールアドレスまたはユーザー名でユーザー検索
  - [x] bcrypt によるパスワード検証（Devise 互換）
  - [x] 認証失敗時のエラーメッセージ表示

- [x] **セッション作成（Rails 互換）**
  - [x] sessions テーブルへの INSERT
  - [x] セッション Cookie の発行
  - [x] Rails と同じセッションキー（`_annict_session_v201904`）使用

- [x] **ログイン後のリダイレクト**
  - [x] トップページへのリダイレクト
  - [x] ログイン前にアクセスしようとしたページへのリダイレクト（オプション）

## 認証方式

### パスワード認証

- **Rails との互換性**: Devise の `encrypted_password` カラム（bcrypt）を使用
- **ハッシュ化**: bcrypt（コスト: 10、Rails のデフォルト）
- **検証**: `internal/auth` パッケージの `CheckPassword` 関数

### セッション管理

- **ストレージ**: PostgreSQL の `sessions` テーブル
- **セッション ID**: UUID v4（ランダム生成）
- **有効期限**: 30 日間（Rails と同じ）
- **Cookie 設定**:
  - Name: `_annict_session_v201904`
  - Domain: `.example.dev`（開発）、`.annict.com`（本番）
  - Secure: true（HTTPS のみ）
  - HttpOnly: true（JavaScript からアクセス不可）
  - SameSite: Lax（CSRF 対策）

## 実装詳細

### ログインフォーム（GET /sign_in）

```html
<!-- internal/templates/auth/sign_in.html -->
<form method="POST" action="/sign_in">
    <input type="hidden" name="csrf_token" value="{{.CSRFToken}}" />

    <div>
        <label for="email_username">{{call .T "sign_in_email_username_label"}}</label>
        <input type="text" id="email_username" name="email_username" required />
    </div>

    <div>
        <label for="password">{{call .T "sign_in_password_label"}}</label>
        <input type="password" id="password" name="password" required />
    </div>

    <button type="submit">{{call .T "sign_in_submit"}}</button>

    <a href="/password/reset">{{call .T "sign_in_forgot_password"}}</a>
</form>
```

### 認証処理（POST /sign_in）

```go
// internal/handler/sign_in.go
func (h *Handler) ProcessSignIn(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // リクエストのパース
    req := &SignInRequest{
        EmailOrUsername: r.FormValue("email_username"),
        Password:        r.FormValue("password"),
    }

    // 形式バリデーション
    if formErrors := req.Validate(ctx); formErrors != nil {
        h.renderSignInWithErrors(w, r, formErrors)
        return
    }

    // ユーザー検索
    user, err := h.queries.GetUserByEmailOrUsername(ctx, req.EmailOrUsername)
    if err != nil {
        h.renderSignInWithError(w, r, i18n.T(ctx, "sign_in_error_invalid_credentials"))
        return
    }

    // パスワード検証
    if err := auth.CheckPassword(user.EncryptedPassword, req.Password); err != nil {
        h.renderSignInWithError(w, r, i18n.T(ctx, "sign_in_error_invalid_credentials"))
        return
    }

    // セッション作成
    uc := usecase.NewCreateSessionUsecase(h.queries)
    result, err := uc.Execute(ctx, user.ID, user.EncryptedPassword)
    if err != nil {
        h.renderSignInWithError(w, r, i18n.T(ctx, "sign_in_error_session_creation_failed"))
        return
    }

    // セッションCookieの発行
    http.SetCookie(w, &http.Cookie{
        Name:     "_annict_session_v201904",
        Value:    result.PublicID,
        Path:     "/",
        Domain:   h.cfg.Domain,
        Secure:   true,
        HttpOnly: true,
        SameSite: http.SameSiteLaxMode,
        MaxAge:   86400 * 30, // 30日間
    })

    // トップページにリダイレクト
    http.Redirect(w, r, "/", http.StatusSeeOther)
}
```

### セッション作成 Usecase

```go
// internal/usecase/create_session.go
type CreateSessionUsecase struct {
    queries *repository.Queries
}

type SessionResult struct {
    PublicID string
    UserID   int64
}

func (uc *CreateSessionUsecase) Execute(ctx context.Context, userID int64, encryptedPassword string) (*SessionResult, error) {
    // セッションIDの生成（UUID v4）
    publicID := uuid.New().String()

    // セッションデータの作成（Rails互換のMarshal形式）
    sessionData := createRailsCompatibleSessionData(userID, encryptedPassword)

    // sessionsテーブルへのINSERT
    err := uc.queries.CreateSession(ctx, repository.CreateSessionParams{
        SessionID: publicID,
        Data:      sessionData,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create session: %w", err)
    }

    return &SessionResult{
        PublicID: publicID,
        UserID:   userID,
    }, nil
}
```

### パスワード検証

```go
// internal/auth/password.go
func CheckPassword(hashedPassword string, plainPassword string) error {
    return bcrypt.CompareHashAndPassword(
        []byte(hashedPassword),
        []byte(plainPassword),
    )
}
```

## セキュリティ考慮事項

### タイミング攻撃対策

ユーザーの存在を推測されないよう、常に同じエラーメッセージを返す：

```go
// ❌ NG: ユーザーの存在を明かす
if user == nil {
    return "このメールアドレスは登録されていません"
}
if !passwordMatch {
    return "パスワードが間違っています"
}

// ✅ OK: 常に同じメッセージ
if user == nil || !passwordMatch {
    return "メールアドレスまたはパスワードが間違っています"
}
```

### CSRF 対策

- すべてのフォーム送信に CSRF トークンを含める
- `justinas/nosurf` ミドルウェアで検証

### ブルートフォース攻撃対策

将来的に Rate Limiting を実装予定（パスワードリセット機能で実装した Redis ベースの仕組みを流用）。

## テスト

### 認証処理のテスト

```go
func TestProcessSignIn(t *testing.T) {
    db, tx := testutil.SetupTestDB(t)

    // テストユーザー作成
    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
    userID := testutil.NewUserBuilder(t, tx).
        WithEmail("test@example.com").
        WithEncryptedPassword(string(hashedPassword)).
        Build()

    // ログインリクエスト
    form := url.Values{}
    form.Add("email_username", "test@example.com")
    form.Add("password", "password123")

    req := httptest.NewRequest("POST", "/sign_in", strings.NewReader(form.Encode()))
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    rr := httptest.NewRecorder()
    handler := &Handler{queries: repository.New(db).WithTx(tx)}
    handler.ProcessSignIn(rr, req)

    // アサーション
    assert.Equal(t, http.StatusSeeOther, rr.Code) // リダイレクト
    assert.Equal(t, "/", rr.Header().Get("Location"))

    // Cookieの確認
    cookies := rr.Result().Cookies()
    assert.Len(t, cookies, 1)
    assert.Equal(t, "_annict_session_v201904", cookies[0].Name)
}
```

### パスワード検証のテスト

```go
func TestCheckPassword(t *testing.T) {
    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

    // 正しいパスワード
    err := auth.CheckPassword(string(hashedPassword), "password123")
    assert.Nil(t, err)

    // 間違ったパスワード
    err = auth.CheckPassword(string(hashedPassword), "wrongpassword")
    assert.NotNil(t, err)
}
```

## 国際化

### 翻訳キー

```toml
# internal/i18n/locales/ja.toml
[sign_in_title]
other = "ログイン"

[sign_in_email_username_label]
other = "メールアドレスまたはユーザー名"

[sign_in_password_label]
other = "パスワード"

[sign_in_submit]
other = "ログイン"

[sign_in_forgot_password]
other = "パスワードをお忘れですか？"

[sign_in_error_email_username_required]
other = "メールアドレスまたはユーザー名を入力してください"

[sign_in_error_password_required]
other = "パスワードを入力してください"

[sign_in_error_invalid_credentials]
other = "メールアドレスまたはパスワードが間違っています"

[sign_in_error_session_creation_failed]
other = "ログインに失敗しました。もう一度お試しください。"
```

## 関連機能

- **パスワードリセット機能**: [フェーズ 6](../202510/password-reset.md) で実装中
- **ログアウト機能**: [フェーズ 7](../202510/sign-out.md) で実装予定

## 成果

- **Rails 互換のログイン**: Devise と同じ bcrypt パスワードハッシュを使用
- **セッション共有**: Rails 側のセッションと互換性のあるセッション管理
- **セキュアな実装**: CSRF 対策、タイミング攻撃対策を実装
- **国際化対応**: 日本語・英語のログインページ
- **テスト基盤**: 認証処理のテストパターンを確立

## 関連ドキュメント

- [プロジェクト全体の設計書](./go.md)
- [認証とセッション管理](./authentication-session.md)
- [パスワードリセット機能（フェーズ 6）](../202510/password-reset.md)
- [ログアウト機能（フェーズ 7）](../202510/sign-out.md)

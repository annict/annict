package sign_in_password

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/ratelimit"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/validator"
)

const rateLimitExceededMessage = "リクエストが多すぎます。しばらくしてから再度お試しください。"

// newRateLimitTestHandlerは限定器を持つテスト用のHandlerとセッションマネージャーを作成します
func newRateLimitTestHandler(t *testing.T, db *sql.DB, tx *sql.Tx, limiter *ratelimit.Limiter) (*Handler, *session.Manager) {
	t.Helper()

	queries := testutil.NewQueriesWithTx(db, tx)

	cfg := &config.Config{
		CookieDomain:  ".examle.com",
		SessionSecure: "false",
	}
	sessionMgr := session.NewManager(repository.NewSessionRepository(queries), cfg)

	signInPasswordValidator := validator.NewSignInPasswordCreateValidator(repository.NewUserRepository(queries))
	createSessionUC := usecase.NewCreateSessionUsecase(repository.NewSessionRepository(queries))
	authenticateByPasswordUC := usecase.NewAuthenticateByPasswordUsecase(createSessionUC, signInPasswordValidator)

	return NewHandler(cfg, sessionMgr, testutil.NewTestFlashManager(), limiter, authenticateByPasswordUC), sessionMgr
}

// postSignInPasswordはPOST /sign_in/passwordを指定したIPアドレスから送信します
func postSignInPassword(t *testing.T, handler *Handler, sessionCookie *http.Cookie, clientIP string, password string) *httptest.ResponseRecorder {
	t.Helper()

	form := url.Values{}
	form.Set("password", password)

	req := httptest.NewRequest("POST", "/sign_in/password", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.RemoteAddr = clientIP + ":12345"
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()

	testutil.ApplyI18nMiddleware(t, handler.Create)(rr, req)

	return rr
}

// TestCreate_RateLimiting_Emailはメールアドレス単位のRate Limitingテスト
// 上限を超えたあとは、正しいパスワードでもサインインできないことを確認する
func TestCreate_RateLimiting_Email(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	limiter := ratelimit.NewLimiter(repository.NewRateLimitRepository(testutil.NewQueriesWithTx(db, tx)))
	handler, sessionMgr := newRateLimitTestHandler(t, db, tx, limiter)

	// 並列に走る他テストとrate_limitsの一意制約で待ち合わないよう、
	// 本テスト固有のキー構成値を組み立てる。
	prefix := testutil.UniqueRateLimitPrefix(t)
	email := strings.ToLower(prefix) + "@example.com"
	ipFor := func(i int) string { return prefix + "-ip" + strconv.Itoa(i) }

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("パスワードハッシュ生成エラー: %v", err)
	}
	testutil.NewUserBuilder(t, tx).
		WithUsername(strings.ToLower(prefix)).
		WithEmail(email).
		WithEncryptedPassword(string(hashedPassword)).
		Build()

	sessionCookie := setupSessionWithEmail(t, sessionMgr, email)

	// 成功したサインインも1回に数える
	rr := postSignInPassword(t, handler, sessionCookie, ipFor(0), "password123")
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("成功したサインインのステータスコード = %v、期待値 = %v", rr.Code, http.StatusSeeOther)
	}
	sessionCookie = setupSessionWithEmail(t, sessionMgr, email)

	// 異なるIPアドレスから間違ったパスワードで9回試す (制限: 10回/時間)
	for i := 1; i < 10; i++ {
		rr := postSignInPassword(t, handler, sessionCookie, ipFor(i), "wrongpassword")
		if rr.Code != http.StatusUnprocessableEntity {
			t.Errorf("%d件目のリクエストのステータスコード = %v、期待値 = %v", i+1, rr.Code, http.StatusUnprocessableEntity)
		}
		if strings.Contains(rr.Body.String(), rateLimitExceededMessage) {
			t.Errorf("%d件目のリクエストがRate Limitingにより制限されました", i+1)
		}
	}

	// 11回目は正しいパスワードでも制限される
	rr = postSignInPassword(t, handler, sessionCookie, ipFor(10), "password123")
	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("11件目のリクエストのステータスコード = %v、期待値 = %v (レート制限)", rr.Code, http.StatusUnprocessableEntity)
	}
	if !strings.Contains(rr.Body.String(), rateLimitExceededMessage) {
		t.Error("11件目のリクエストでRate Limitingのエラーメッセージが表示されていません")
	}
	for _, cookie := range rr.Result().Cookies() {
		if cookie.Name == session.SessionKey && cookie.Value != sessionCookie.Value {
			t.Error("制限されたリクエストで新しいセッションCookieが設定されました")
		}
	}
}

// TestCreate_RateLimiting_IPはIP単位のRate Limitingテスト
func TestCreate_RateLimiting_IP(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	limiter := ratelimit.NewLimiter(repository.NewRateLimitRepository(testutil.NewQueriesWithTx(db, tx)))
	handler, sessionMgr := newRateLimitTestHandler(t, db, tx, limiter)

	// 並列に走る他テストとrate_limitsの一意制約で待ち合わないよう、
	// 本テスト固有のキー構成値を組み立てる。
	prefix := testutil.UniqueRateLimitPrefix(t)
	clientIP := prefix + "-ip"
	emailFor := func(i int) string { return strings.ToLower(prefix) + "-" + strconv.Itoa(i) + "@example.com" }

	// メールアドレス単位の制限 (10回/時間) に掛からないよう、
	// 10回ごとに別のメールアドレスのセッションに切り替えて、同じIPアドレスから31回試す (制限: 30回/時間)
	sessionCookies := make([]*http.Cookie, 4)
	for i := range sessionCookies {
		sessionCookies[i] = setupSessionWithEmail(t, sessionMgr, emailFor(i))
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("パスワードハッシュ生成エラー: %v", err)
	}
	testutil.NewUserBuilder(t, tx).
		WithUsername(strings.ToLower(prefix)).
		WithEmail(emailFor(0)).
		WithEncryptedPassword(string(hashedPassword)).
		Build()

	// 成功したサインインもIP単位の1回に数える
	rr := postSignInPassword(t, handler, sessionCookies[0], clientIP, "password123")
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("成功したサインインのステータスコード = %v、期待値 = %v", rr.Code, http.StatusSeeOther)
	}
	sessionCookies[0] = setupSessionWithEmail(t, sessionMgr, emailFor(0))

	for i := 1; i < 31; i++ {
		rr := postSignInPassword(t, handler, sessionCookies[i/10], clientIP, "wrongpassword")
		if rr.Code != http.StatusUnprocessableEntity {
			t.Errorf("%d件目のリクエストのステータスコード = %v、期待値 = %v", i+1, rr.Code, http.StatusUnprocessableEntity)
		}

		limited := strings.Contains(rr.Body.String(), rateLimitExceededMessage)
		if i < 30 && limited {
			t.Errorf("%d件目のリクエストがRate Limitingにより制限されました", i+1)
		}
		if i == 30 && !limited {
			t.Error("31件目のリクエストでRate Limitingのエラーメッセージが表示されていません")
		}
	}
}

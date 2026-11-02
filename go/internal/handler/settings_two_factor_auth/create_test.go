package settings_two_factor_auth

import (
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/testutil"
)

// TestCreateは、認証アプリのコードが合えば二要素認証を有効にし、リカバリーコードをこの応答で一度だけ表示して
// キャッシュさせないことを検証する。
func TestCreate(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	handler := newTestHandler(t, db)
	user := createCommittedUser(t, db)
	serve(handler, newRequest(t, http.MethodGet, "/settings/two_factor_auth/new", nil, user, "ja"))
	code := testutil.TOTPCode(t, pendingSecretForTest(t, handler, user), time.Now())

	rr := serve(handler, newRequest(t, http.MethodPost, "/settings/two_factor_auth", url.Values{"code": {code}}, user, "ja"))

	if rr.Code != http.StatusOK {
		t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusOK)
	}
	assertNoStore(t, rr)
	body := rr.Body.String()
	assertContains(t, body, "<title>リカバリーコード | Annict</title>", "二要素認証を有効にしました。", "このコードは今しか表示されません", `href="/settings/two_factor_auth"`)
	if got := len(regexp.MustCompile(`<li>[a-z2-9]{4}-[a-z2-9]{4}</li>`).FindAllString(body, -1)); got != 10 {
		t.Errorf("表示したリカバリーコードの数 = %d、期待値 = 10", got)
	}

	var enabled bool
	if err := db.QueryRow("SELECT enabled_at IS NOT NULL FROM user_two_factor_auths WHERE user_id = $1", int64(user.ID)).Scan(&enabled); err != nil {
		t.Fatalf("二要素認証の設定の取得に失敗: %v", err)
	}
	if !enabled {
		t.Error("二要素認証が有効になっていない")
	}
}

// TestCreate_IncorrectCodeは、コードが合わないときに同じ秘密鍵のまま登録画面を422で描き直し、
// 入力したコードとエラーを表示することを検証する。
func TestCreate_IncorrectCode(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	handler := newTestHandler(t, db)
	user := createCommittedUser(t, db)
	serve(handler, newRequest(t, http.MethodGet, "/settings/two_factor_auth/new", nil, user, "ja"))
	secret := pendingSecretForTest(t, handler, user)
	wrongCode := testutil.NonMatchingTOTPCode(t, secret, time.Now())

	rr := serve(handler, newRequest(t, http.MethodPost, "/settings/two_factor_auth", url.Values{"code": {wrongCode}}, user, "ja"))

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusUnprocessableEntity)
	}
	assertNoStore(t, rr)
	assertContains(t, rr.Body.String(),
		"コードが正しくありません",
		`aria-invalid="true"`,
		`aria-describedby="code-error-1"`,
		`value="`+wrongCode+`"`,
		"secret="+secret,
	)
	assertFocusedErrorSummary(t, rr.Body.String(), "コードが正しくありません", `href="#code"`)
	if got := pendingSecretForTest(t, handler, user); got != secret {
		t.Errorf("描き直したあとの秘密鍵 = %q、期待値 = %q (作り直さない)", got, secret)
	}
}

// TestCreate_NotPendingは、登録画面を開いていない (登録の途中の設定が無い) ときに、
// 設定画面へ送って状態が変わったことを知らせることを検証する。
func TestCreate_NotPending(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	handler := newTestHandler(t, db)
	user := createCommittedUser(t, db)

	rr := serve(handler, newRequest(t, http.MethodPost, "/settings/two_factor_auth", url.Values{"code": {"123456"}}, user, "ja"))

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusSeeOther)
	}
	if location := rr.Header().Get("Location"); location != "/settings/two_factor_auth" {
		t.Errorf("Location = %q、期待値 = %q", location, "/settings/two_factor_auth")
	}
	var hasFlash bool
	for _, cookie := range rr.Result().Cookies() {
		if cookie.Name == session.FlashCookieName && cookie.Value != "" {
			hasFlash = true
		}
	}
	if !hasFlash {
		t.Error("フラッシュメッセージのCookieが設定されていない")
	}
}

// TestCreate_NotPendingInvalidFormatは、登録の途中の設定が無いときに形式の誤ったコードを送ると、
// 登録画面を描き直さずに設定画面へ送ることを検証する。
func TestCreate_NotPendingInvalidFormat(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	handler := newTestHandler(t, db)
	user := createCommittedUser(t, db)
	enableForTest(t, db, user.ID, 10)

	rr := serve(handler, newRequest(t, http.MethodPost, "/settings/two_factor_auth", url.Values{"code": {"12345"}}, user, "ja"))

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusSeeOther)
	}
	if location := rr.Header().Get("Location"); location != "/settings/two_factor_auth" {
		t.Errorf("Location = %q、期待値 = %q", location, "/settings/two_factor_auth")
	}
}

// TestCreate_RequiresAuthは、未ログインのリクエストを受け付けずログイン画面へ送ることを検証する。
func TestCreate_RequiresAuth(t *testing.T) {
	t.Parallel()

	handler := newTestHandler(t, testutil.GetTestDB())

	rr := serveUnauthenticated(handler, http.MethodPost, "/settings/two_factor_auth")

	assertRedirectToSignIn(t, rr, "/settings/two_factor_auth")
}

// TestCreate_RateLimitedは、1人のユーザーの有効化のコードの照合が15分で5回を超えると、
// 照合せずに登録画面を422で描き直して回数の超過を伝えることを検証する。
func TestCreate_RateLimited(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	handler := newTestHandler(t, db)
	user := createCommittedUser(t, db)
	serve(handler, newRequest(t, http.MethodGet, "/settings/two_factor_auth/new", nil, user, "ja"))

	for i := range enableRateLimit {
		rr := serve(handler, newRequest(t, http.MethodPost, "/settings/two_factor_auth", url.Values{"code": {"12345"}}, user, "ja"))
		if strings.Contains(rr.Body.String(), "リクエストが多すぎます") {
			t.Fatalf("%d回目で回数の超過になった、%d回までは照合することを期待", i+1, enableRateLimit)
		}
	}

	// 上限を超えたあとは、正しいコードでも有効にしない。
	code := testutil.TOTPCode(t, pendingSecretForTest(t, handler, user), time.Now())
	rr := serve(handler, newRequest(t, http.MethodPost, "/settings/two_factor_auth", url.Values{"code": {code}}, user, "ja"))

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusUnprocessableEntity)
	}
	assertContains(t, rr.Body.String(), "リクエストが多すぎます")
	assertFocusedErrorSummary(t, rr.Body.String(), "リクエストが多すぎます")
	var enabled bool
	if err := db.QueryRow("SELECT enabled_at IS NOT NULL FROM user_two_factor_auths WHERE user_id = $1", int64(user.ID)).Scan(&enabled); err != nil {
		t.Fatalf("二要素認証の設定の取得に失敗: %v", err)
	}
	if enabled {
		t.Error("回数の超過後に二要素認証が有効になった")
	}
}

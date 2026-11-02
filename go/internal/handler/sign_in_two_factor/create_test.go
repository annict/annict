package sign_in_two_factor

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/ratelimit"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/testutil"
)

// TestCreateは、認証アプリのコードが合えば新しいセッションを発行してログイン途中の状態を消し、戻り先へ送ることを検証する。
func TestCreate(t *testing.T) {
	t.Parallel()

	e := newTestEnv(t)
	userID := testutil.NewUserBuilder(t, e.tx).Build()
	secret := testutil.EnableTwoFactorAuthForTest(t, e.tx, userID)
	cookie := e.startSignInTwoFactor(t, userID, time.Now())

	form := url.Values{"code": {testutil.TOTPCode(t, secret, time.Now())}, "back": {"/works/1"}}
	rr := e.serve(newRequest(http.MethodPost, "/sign_in/two_factor", form, cookie))

	assertRedirect(t, rr, "/works/1")
	assertNoStore(t, rr)
	newCookie := sessionCookie(t, rr)
	if newCookie.Value == cookie.Value || newCookie.Value == "" {
		t.Errorf("セッションのCookie = %q、ログイン前と違う新しいセッションを期待", newCookie.Value)
	}

	signedIn := newRequest(http.MethodGet, "/", nil, newCookie)
	user, err := e.sessionMgr.GetCurrentUser(signedIn.Context(), signedIn)
	if err != nil || user == nil || user.ID != userID {
		t.Errorf("GetCurrentUser() = (%+v, %v)、ログインしたユーザーを期待", user, err)
	}

	pendingReq := newRequest(http.MethodGet, "/sign_in/two_factor", nil, cookie)
	if _, ok, err := e.sessionMgr.SignInTwoFactorUserID(pendingReq.Context(), pendingReq, time.Now()); err != nil || ok {
		t.Errorf("SignInTwoFactorUserID() = (ok:%t, %v)、ログイン途中の状態が消えていることを期待", ok, err)
	}
}

// TestCreate_IncorrectCodeは、コードが合わないときに入力を戻して画面を422で描き直し、
// エラーの要約とコードの欄に理由を出してセッションを発行しないことを検証する。
func TestCreate_IncorrectCode(t *testing.T) {
	t.Parallel()

	e := newTestEnv(t)
	userID := testutil.NewUserBuilder(t, e.tx).Build()
	secret := testutil.EnableTwoFactorAuthForTest(t, e.tx, userID)
	cookie := e.startSignInTwoFactor(t, userID, time.Now())
	wrongCode := testutil.NonMatchingTOTPCode(t, secret, time.Now())

	rr := e.serve(newRequest(http.MethodPost, "/sign_in/two_factor", url.Values{"code": {wrongCode}, "back": {"/works/1"}}, cookie))

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusUnprocessableEntity)
	}
	assertNoStore(t, rr)
	assertContains(t, rr.Body.String(),
		"コードが正しくありません",
		`aria-invalid="true"`,
		`value="`+wrongCode+`"`,
		`name="back" value="/works/1"`,
		// 文の長さで広がらないよう、ログイン画面と同じ幅に固定する
		`min-w-full md:min-w-md md:max-w-md`,
	)
	for _, c := range rr.Result().Cookies() {
		if c.Name == session.SessionKey && c.Value != cookie.Value {
			t.Error("コードが合わないのに新しいセッションのCookieが設定された")
		}
	}
}

// TestCreate_NoPendingは、ログイン途中の状態が無いときに、照合せずにやり直しを伝えて
// メールアドレスの入力からやり直させることを検証する。
func TestCreate_NoPending(t *testing.T) {
	t.Parallel()

	e := newTestEnv(t)

	rr := e.serve(newRequest(http.MethodPost, "/sign_in/two_factor", url.Values{"code": {"123456"}, "back": {"/works/1"}}, nil))

	assertRedirect(t, rr, "/sign_in?back=%2Fworks%2F1")
	assertNoStore(t, rr)
	e.assertRestartFlash(t, rr)
}

// TestCreate_TwoFactorAuthDisabledは、第1の要素を確かめたあとに二要素認証を無効にしたとき、
// ログイン途中の状態を消し、やり直しを伝えてメールアドレスの入力からやり直させることを検証する。
func TestCreate_TwoFactorAuthDisabled(t *testing.T) {
	t.Parallel()

	e := newTestEnv(t)
	userID := testutil.NewUserBuilder(t, e.tx).Build()
	cookie := e.startSignInTwoFactor(t, userID, time.Now())

	rr := e.serve(newRequest(http.MethodPost, "/sign_in/two_factor", url.Values{"code": {"123456"}}, cookie))

	assertRedirect(t, rr, "/sign_in")
	assertNoStore(t, rr)
	e.assertRestartFlash(t, rr)
	pendingReq := newRequest(http.MethodGet, "/sign_in/two_factor", nil, cookie)
	if _, ok, err := e.sessionMgr.SignInTwoFactorUserID(pendingReq.Context(), pendingReq, time.Now()); err != nil || ok {
		t.Errorf("SignInTwoFactorUserID() = (ok:%t, %v)、ログイン途中の状態が消えていることを期待", ok, err)
	}
}

// TestCreate_RateLimitedは、1人のユーザーの試行が15分で5回を超えると、照合せずに画面を422で描き直して
// 回数の超過を伝え、正しいコードでもログインさせないことを検証する。
func TestCreate_RateLimited(t *testing.T) {
	t.Parallel()

	e := newTestEnv(t)
	userID := testutil.NewUserBuilder(t, e.tx).Build()
	secret := testutil.EnableTwoFactorAuthForTest(t, e.tx, userID)
	cookie := e.startSignInTwoFactor(t, userID, time.Now())

	for i := range ratelimit.SignInTwoFactorUserLimit {
		rr := e.serve(newRequest(http.MethodPost, "/sign_in/two_factor", url.Values{"code": {"12345"}}, cookie))
		if strings.Contains(rr.Body.String(), "リクエストが多すぎます") {
			t.Fatalf("%d回目で回数の超過になった、%d回までは照合することを期待", i+1, ratelimit.SignInTwoFactorUserLimit)
		}
	}

	rr := e.serve(newRequest(http.MethodPost, "/sign_in/two_factor", url.Values{"code": {testutil.TOTPCode(t, secret, time.Now())}}, cookie))

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusUnprocessableEntity)
	}
	assertContains(t, rr.Body.String(), "リクエストが多すぎます")
}

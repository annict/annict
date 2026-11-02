package sign_in_two_factor_recovery

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/ratelimit"
	"github.com/annict/annict/go/internal/repository"
)

// TestCreateは、リカバリーコードが合えば新しいセッションを発行してログイン途中の状態を消し、戻り先へ送ることを検証する。
func TestCreate(t *testing.T) {
	t.Parallel()

	e := newTestEnv(t)
	userID := e.createUser(t, true)
	cookie := e.startSignInTwoFactor(t, userID)

	rr := e.serve(newRequest(http.MethodPost, "/sign_in/two_factor/recovery", url.Values{"code": {testRecoveryCodes[0]}, "back": {"/works/1"}}, cookie))

	assertRedirect(t, rr, "/works/1")
	assertNoStore(t, rr)
	newCookie := sessionCookie(t, rr)
	if newCookie.Value == cookie.Value {
		t.Error("ログイン前と同じセッションのCookieが設定された、新しいセッションを期待")
	}

	signedIn := newRequest(http.MethodGet, "/", nil, newCookie)
	user, err := e.sessionMgr.GetCurrentUser(signedIn.Context(), signedIn)
	if err != nil || user == nil || user.ID != userID {
		t.Errorf("GetCurrentUser() = (%+v, %v)、ログインしたユーザーを期待", user, err)
	}

	e.assertPendingCleared(t, cookie)
}

// TestCreate_IncorrectCodeは、リカバリーコードが合わないときに入力を戻して画面を422で描き直し、
// セッションを発行しないことを検証する。
func TestCreate_IncorrectCode(t *testing.T) {
	t.Parallel()

	e := newTestEnv(t)
	cookie := e.startSignInTwoFactor(t, e.createUser(t, true))

	rr := e.serve(newRequest(http.MethodPost, "/sign_in/two_factor/recovery", url.Values{"code": {"zzzz-zzzz"}, "back": {"/works/1"}}, cookie))

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusUnprocessableEntity)
	}
	assertNoStore(t, rr)
	assertContains(t, rr.Body.String(),
		"リカバリーコードが正しくないか、すでに使われています",
		`aria-invalid="true"`,
		`value="zzzz-zzzz"`,
		`name="back" value="/works/1"`,
	)
	assertNoNewSessionCookie(t, rr, cookie)
}

// TestCreate_NoPendingは、ログイン途中の状態が無いときに、照合せずにやり直しを伝えて
// メールアドレスの入力からやり直させることを検証する。
func TestCreate_NoPending(t *testing.T) {
	t.Parallel()

	e := newTestEnv(t)

	rr := e.serve(newRequest(http.MethodPost, "/sign_in/two_factor/recovery", url.Values{"code": {testRecoveryCodes[0]}, "back": {"/works/1"}}, nil))

	assertRedirect(t, rr, "/sign_in?back=%2Fworks%2F1")
	assertNoStore(t, rr)
	e.assertRestartFlash(t, rr)
}

// TestCreate_TwoFactorAuthDisabledは、第1の要素を確かめたあとに二要素認証を無効にしたとき、
// ログイン途中の状態を消し、やり直しを伝えて戻り先を引き継いだメールアドレスの入力へ送ることを検証する。
func TestCreate_TwoFactorAuthDisabled(t *testing.T) {
	t.Parallel()

	e := newTestEnv(t)
	cookie := e.startSignInTwoFactor(t, e.createUser(t, false))

	rr := e.serve(newRequest(http.MethodPost, "/sign_in/two_factor/recovery", url.Values{"code": {testRecoveryCodes[0]}, "back": {"/works/1"}}, cookie))

	assertRedirect(t, rr, "/sign_in?back=%2Fworks%2F1")
	assertNoStore(t, rr)
	e.assertPendingCleared(t, cookie)
	e.assertRestartFlash(t, rr)
}

// TestCreate_RateLimitedSharedは、認証アプリのコードの画面で数えた試行がリカバリーコードの画面でも数えられ、
// 上限を超えたあとは正しいリカバリーコードでもログインさせず、コードも消費しないことを検証する。
func TestCreate_RateLimitedShared(t *testing.T) {
	t.Parallel()

	e := newTestEnv(t)
	userID := e.createUser(t, true)
	cookie := e.startSignInTwoFactor(t, userID)

	// 認証アプリのコードの画面と同じカウンターで、ユーザーの単位の上限まで試したことにする。
	// IPアドレスはこのリクエストと別の値にし、ユーザーのカウンターが2つの画面で共有されることだけを確かめる。
	for i := range ratelimit.SignInTwoFactorUserLimit {
		if _, err := e.limiter.CheckSignInTwoFactor(context.Background(), "rate-limited-shared-test-"+userID.String(), userID.String()); err != nil {
			t.Fatalf("%d回目のCheckSignInTwoFactor()のエラー = %v", i+1, err)
		}
	}

	rr := e.serve(newRequest(http.MethodPost, "/sign_in/two_factor/recovery", url.Values{"code": {testRecoveryCodes[0]}}, cookie))

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusUnprocessableEntity)
	}
	assertContains(t, rr.Body.String(), "リクエストが多すぎます")
	assertNoNewSessionCookie(t, rr, cookie)
	count, err := repository.NewUserTwoFactorRecoveryCodeRepository(query.New(e.db)).CountUnused(context.Background(), userID)
	if err != nil {
		t.Fatalf("CountUnused()のエラー = %v", err)
	}
	if count != len(testRecoveryCodes) {
		t.Errorf("未使用のリカバリーコードの数 = %d、期待値 = %d (回数の超過後は消費しない)", count, len(testRecoveryCodes))
	}
}

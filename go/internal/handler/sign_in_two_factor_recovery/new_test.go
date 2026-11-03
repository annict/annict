package sign_in_two_factor_recovery

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestNewは、ログイン途中の状態があるときに、戻り先を載せたリカバリーコードの入力画面を表示し、
// 認証アプリのコードの画面へ戻り先を引き継いで戻れること、キャッシュさせないことを検証する。
func TestNew(t *testing.T) {
	t.Parallel()

	e := newTestEnv(t)
	cookie := e.startSignInTwoFactor(t, e.createUser(t, true))

	rr := e.serve(newRequest(http.MethodGet, "/sign_in/two_factor/recovery?back=%2Fworks%2F1", nil, cookie))

	if rr.Code != http.StatusOK {
		t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusOK)
	}
	assertNoStore(t, rr)
	assertContains(t, rr.Body.String(),
		"<title>リカバリーコードの入力 | Annict</title>",
		`action="/sign_in/two_factor/recovery"`,
		`name="csrf_token"`,
		`name="back" value="/works/1"`,
		`name="code"`,
		`autocomplete="off"`,
		`href="/sign_in/two_factor?back=%2Fworks%2F1"`,
		// 文の長さで広がらないよう、ログイン画面と同じ幅に固定する
		`min-w-full md:min-w-md md:max-w-md`,
	)
}

// TestNew_NoPendingは、ログイン途中の状態が無いときに、戻り先を引き継いでメールアドレスの入力へ戻すことを検証する。
func TestNew_NoPending(t *testing.T) {
	t.Parallel()

	e := newTestEnv(t)

	rr := e.serve(newRequest(http.MethodGet, "/sign_in/two_factor/recovery?back=%2Fworks%2F1", nil, nil))

	assertRedirect(t, rr, "/sign_in?back=%2Fworks%2F1")
	assertNoStore(t, rr)
}

// TestNew_TwoFactorAuthDisabledは、第1の要素を確かめたあとに二要素認証を無効にしたとき、
// ログイン途中の状態を消し、やり直しを伝えてメールアドレスの入力からやり直させることを検証する。
func TestNew_TwoFactorAuthDisabled(t *testing.T) {
	t.Parallel()

	e := newTestEnv(t)
	cookie := e.startSignInTwoFactor(t, e.createUser(t, false))

	rr := e.serve(newRequest(http.MethodGet, "/sign_in/two_factor/recovery", nil, cookie))

	assertRedirect(t, rr, "/sign_in")
	assertNoStore(t, rr)
	e.assertPendingCleared(t, cookie)
	e.assertRestartFlash(t, rr)
}

// TestNew_SignedInは、ログイン済みのときは入力画面を出さずにホームへ送ることを検証する。
func TestNew_SignedIn(t *testing.T) {
	t.Parallel()

	e := newTestEnv(t)
	userID := e.createUser(t, false)
	signInRR := httptest.NewRecorder()
	signInReq := newRequest(http.MethodGet, "/", nil, nil)
	if err := e.sessionMgr.CreateSession(signInReq.Context(), signInRR, signInReq, userID); err != nil {
		t.Fatalf("CreateSession()のエラー = %v", err)
	}

	rr := e.serve(newRequest(http.MethodGet, "/sign_in/two_factor/recovery?back=%2Fworks%2F1", nil, sessionCookie(t, signInRR)))

	assertRedirect(t, rr, "/")
	assertNoStore(t, rr)
}

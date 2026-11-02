package sign_in_two_factor

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/testutil"
)

// TestNewは、ログイン途中の状態があるときに、戻り先を載せた認証アプリのコードの入力画面を表示し、キャッシュさせないことを検証する。
func TestNew(t *testing.T) {
	t.Parallel()

	e := newTestEnv(t)
	userID := testutil.NewUserBuilder(t, e.tx).Build()
	testutil.EnableTwoFactorAuthForTest(t, e.tx, userID)
	cookie := e.startSignInTwoFactor(t, userID, time.Now())

	rr := e.serve(newRequest(http.MethodGet, "/sign_in/two_factor?back=%2Fworks%2F1", nil, cookie))

	if rr.Code != http.StatusOK {
		t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusOK)
	}
	assertNoStore(t, rr)
	assertContains(t, rr.Body.String(),
		"<title>認証アプリのコードの入力 | Annict</title>",
		`action="/sign_in/two_factor"`,
		`name="csrf_token"`,
		`name="back" value="/works/1"`,
		`name="code"`,
		`autocomplete="one-time-code"`,
		`inputmode="numeric"`,
		`href="/sign_in/two_factor/recovery?back=%2Fworks%2F1"`,
	)
}

// TestNew_Unavailableは、ログイン途中に無効化・再登録の開始・退会があったときに、
// 途中の状態を消して、戻り先とキャッシュ抑止の指示を保ったままログインをやり直させることを検証する。
func TestNew_Unavailable(t *testing.T) {
	t.Parallel()

	for _, state := range []string{"無効化", "登録の途中", "退会"} {
		t.Run(state, func(t *testing.T) {
			t.Parallel()

			e := newTestEnv(t)
			userID := testutil.NewUserBuilder(t, e.tx).Build()
			testutil.EnableTwoFactorAuthForTest(t, e.tx, userID)
			cookie := e.startSignInTwoFactor(t, userID, time.Now())
			if state == "退会" {
				if _, err := e.tx.Exec("UPDATE users SET deleted_at = NOW() WHERE id = $1", int64(userID)); err != nil {
					t.Fatalf("ユーザーの論理削除に失敗: %v", err)
				}
			} else {
				if _, err := e.tx.Exec("DELETE FROM user_two_factor_auths WHERE user_id = $1", int64(userID)); err != nil {
					t.Fatalf("設定の削除に失敗: %v", err)
				}
				if state == "登録の途中" {
					testutil.NewUserTwoFactorAuthBuilder(t, e.tx, userID).Build()
				}
			}

			rr := e.serve(newRequest(http.MethodGet, "/sign_in/two_factor?back=%2Fworks%2F1", nil, cookie))

			assertRedirect(t, rr, "/sign_in?back=%2Fworks%2F1")
			assertNoStore(t, rr)
			pendingReq := newRequest(http.MethodGet, "/sign_in/two_factor", nil, cookie)
			if _, ok, err := e.sessionMgr.SignInTwoFactorUserID(pendingReq.Context(), pendingReq, time.Now()); err != nil || ok {
				t.Errorf("SignInTwoFactorUserID() = (ok:%t, %v)、ログイン途中の状態が消えていることを期待", ok, err)
			}
			e.assertRestartFlash(t, rr)
		})
	}
}

// TestNew_SignedInは、ログイン済みのときは入力画面を出さずにホームへ送ることを検証する。
func TestNew_SignedIn(t *testing.T) {
	t.Parallel()

	e := newTestEnv(t)
	userID := testutil.NewUserBuilder(t, e.tx).Build()
	signInRR := httptest.NewRecorder()
	signInReq := newRequest(http.MethodGet, "/", nil, nil)
	if err := e.sessionMgr.CreateSession(signInReq.Context(), signInRR, signInReq, userID); err != nil {
		t.Fatalf("CreateSession()のエラー = %v", err)
	}

	rr := e.serve(newRequest(http.MethodGet, "/sign_in/two_factor?back=%2Fworks%2F1", nil, sessionCookie(t, signInRR)))

	assertRedirect(t, rr, "/")
	assertNoStore(t, rr)
}

// TestNew_NoPendingは、ログイン途中の状態が無い・期限が切れたときに、戻り先を引き継いでメールアドレスの入力へ戻すことを検証する。
func TestNew_NoPending(t *testing.T) {
	t.Parallel()

	e := newTestEnv(t)
	userID := testutil.NewUserBuilder(t, e.tx).Build()
	expired := e.startSignInTwoFactor(t, userID, time.Now().Add(-11*time.Minute))

	for name, req := range map[string]*http.Request{
		"Cookieが無い": newRequest(http.MethodGet, "/sign_in/two_factor?back=%2Fworks%2F1", nil, nil),
		"期限切れ":      newRequest(http.MethodGet, "/sign_in/two_factor?back=%2Fworks%2F1", nil, expired),
	} {
		rr := e.serve(req)
		t.Run(name, func(t *testing.T) {
			assertRedirect(t, rr, "/sign_in?back=%2Fworks%2F1")
			assertNoStore(t, rr)
		})
	}
}

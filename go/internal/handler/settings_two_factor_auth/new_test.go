package settings_two_factor_auth

import (
	"net/http"
	"testing"

	"github.com/annict/annict/go/internal/testutil"
)

// TestNewは、登録画面がQRコード・otpauthのリンク・手入力用のキーとコードの入力欄を表示し、
// 登録の途中の設定を作ってキャッシュさせないことを検証する。
func TestNew(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	handler := newTestHandler(t, db)
	user := createCommittedUser(t, db)

	rr := serve(handler, newRequest(t, http.MethodGet, "/settings/two_factor_auth/new", nil, user, "ja"))

	if rr.Code != http.StatusOK {
		t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusOK)
	}
	assertNoStore(t, rr)

	secret := pendingSecretForTest(t, handler, user)
	assertContains(t, rr.Body.String(),
		"<title>認証アプリの登録 | Annict</title>",
		`role="img"`,
		`aria-label="認証アプリに登録するためのQRコード"`,
		`href="otpauth://totp/Annict:`+user.Username+`?`,
		"secret="+secret,
		`value="`+secret[:4]+" "+secret[4:8]+" ",
		`action="/settings/two_factor_auth"`,
		`name="csrf_token"`,
		`autocomplete="one-time-code"`,
		`inputmode="numeric"`,
		`name="code"`,
	)
}

// TestNew_RequiresAuthは、未ログインのリクエストを登録画面に通さずログイン画面へ送ることを検証する。
func TestNew_RequiresAuth(t *testing.T) {
	t.Parallel()

	handler := newTestHandler(t, testutil.GetTestDB())

	rr := serveUnauthenticated(handler, http.MethodGet, "/settings/two_factor_auth/new")

	assertRedirectToSignIn(t, rr, "/settings/two_factor_auth/new")
}

// TestNew_Enabledは、既に有効にしているときに秘密鍵を作らず設定画面へ送ることを検証する。
func TestNew_Enabled(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	handler := newTestHandler(t, db)
	user := createCommittedUser(t, db)
	enableForTest(t, db, user.ID, 10)

	rr := serve(handler, newRequest(t, http.MethodGet, "/settings/two_factor_auth/new", nil, user, "ja"))

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusSeeOther)
	}
	if location := rr.Header().Get("Location"); location != "/settings/two_factor_auth" {
		t.Errorf("Location = %q、期待値 = %q", location, "/settings/two_factor_auth")
	}
}

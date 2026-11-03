package settings_two_factor_auth

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/testutil"
)

// deleteTestPasswordは無効化のハンドラーテストのユーザーに設定するパスワード。
const deleteTestPassword = "delete-password-123"

// twoFactorAuthRowsは、ユーザーの二要素認証の設定とリカバリーコードの行数を返す。
func twoFactorAuthRows(t *testing.T, userID model.UserID) (int, int) {
	t.Helper()

	db := testutil.GetTestDB()
	var settings, codes int
	if err := db.QueryRow("SELECT COUNT(*) FROM user_two_factor_auths WHERE user_id = $1", int64(userID)).Scan(&settings); err != nil {
		t.Fatalf("二要素認証の設定の件数の取得に失敗: %v", err)
	}
	if err := db.QueryRow("SELECT COUNT(*) FROM user_two_factor_recovery_codes WHERE user_id = $1", int64(userID)).Scan(&codes); err != nil {
		t.Fatalf("リカバリーコードの件数の取得に失敗: %v", err)
	}
	return settings, codes
}

// hasFlashCookieは応答がフラッシュメッセージのCookieを設定したかを返す。
func hasFlashCookie(rr interface{ Result() *http.Response }) bool {
	for _, cookie := range rr.Result().Cookies() {
		if cookie.Name == session.FlashCookieName && cookie.Value != "" {
			return true
		}
	}
	return false
}

// newDeleteRequestは、設定画面の無効化のフォームと同じく、POSTに_method=DELETEを載せたリクエストを作る。
func newDeleteRequest(t *testing.T, credential string, user *model.User) *http.Request {
	t.Helper()

	form := url.Values{"_method": {"DELETE"}, "credential": {credential}}
	return newRequest(t, http.MethodPost, "/settings/two_factor_auth", form, user, "ja")
}

// TestDeleteは、本人確認が通れば設定とリカバリーコードを削除し、設定画面へ送って無効にしたことを知らせることを検証する。
func TestDelete(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	handler := newTestHandler(t, db)
	user := createCommittedUserWithPassword(t, db, deleteTestPassword)
	enableForTest(t, db, user.ID, 10)

	rr := serve(handler, newDeleteRequest(t, deleteTestPassword, user))

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusSeeOther)
	}
	assertNoStore(t, rr)
	if location := rr.Header().Get("Location"); location != "/settings/two_factor_auth" {
		t.Errorf("Location = %q、期待値 = %q", location, "/settings/two_factor_auth")
	}
	if !hasFlashCookie(rr) {
		t.Error("フラッシュメッセージのCookieが設定されていない")
	}
	if settings, codes := twoFactorAuthRows(t, user.ID); settings != 0 || codes != 0 {
		t.Errorf("残った行数 = (設定 %d、リカバリーコード %d)、期待値 = (0、0)", settings, codes)
	}
}

// TestDelete_IncorrectCredentialは、本人確認が一致しないときに設定画面を422で描き直してエラーの要約へフォーカスを移し、
// 設定を残すことを検証する。
func TestDelete_IncorrectCredential(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	handler := newTestHandler(t, db)
	user := createCommittedUserWithPassword(t, db, deleteTestPassword)
	enableForTest(t, db, user.ID, 10)

	rr := serve(handler, newDeleteRequest(t, "wrong-password", user))

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusUnprocessableEntity)
	}
	assertNoStore(t, rr)
	body := rr.Body.String()
	assertFocusedErrorSummary(t, body, "パスワードまたはコードが正しくありません")
	assertContains(t, body, "二要素認証は有効です", `aria-invalid="true"`)
	if strings.Contains(body, "wrong-password") {
		t.Error("入力したパスワードを描き直した画面に含めている")
	}
	if settings, codes := twoFactorAuthRows(t, user.ID); settings != 1 || codes != 10 {
		t.Errorf("残った行数 = (設定 %d、リカバリーコード %d)、期待値 = (1、10)", settings, codes)
	}
}

// TestDelete_NotEnabledは、二要素認証が有効でないときに設定画面へ送り、状態が変わったことを知らせることを検証する。
func TestDelete_NotEnabled(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	handler := newTestHandler(t, db)
	user := createCommittedUserWithPassword(t, db, deleteTestPassword)

	rr := serve(handler, newDeleteRequest(t, deleteTestPassword, user))

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusSeeOther)
	}
	if location := rr.Header().Get("Location"); location != "/settings/two_factor_auth" {
		t.Errorf("Location = %q、期待値 = %q", location, "/settings/two_factor_auth")
	}
	if !hasFlashCookie(rr) {
		t.Error("フラッシュメッセージのCookieが設定されていない")
	}
}

// TestDelete_RateLimitedは、1人のユーザーの無効化の本人確認が15分で5回を超えると、
// 照合せずに設定画面を422で描き直して回数の超過を伝えることを検証する。
func TestDelete_RateLimited(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	handler := newTestHandler(t, db)
	user := createCommittedUserWithPassword(t, db, deleteTestPassword)
	enableForTest(t, db, user.ID, 10)

	for i := range disableRateLimit {
		rr := serve(handler, newDeleteRequest(t, "wrong-password", user))
		if strings.Contains(rr.Body.String(), "リクエストが多すぎます") {
			t.Fatalf("%d回目で回数の超過になった、%d回までは照合することを期待", i+1, disableRateLimit)
		}
	}

	// 上限を超えたあとは、正しいパスワードでも無効にしない。
	rr := serve(handler, newDeleteRequest(t, deleteTestPassword, user))

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusUnprocessableEntity)
	}
	assertFocusedErrorSummary(t, rr.Body.String(), "リクエストが多すぎます")
	if settings, _ := twoFactorAuthRows(t, user.ID); settings != 1 {
		t.Errorf("設定の行数 = %d、期待値 = 1 (回数の超過後は無効にしない)", settings)
	}
}

// TestDelete_Unauthenticatedは、未ログインの無効化をログイン画面へ送ることを検証する。
func TestDelete_Unauthenticated(t *testing.T) {
	t.Parallel()

	handler := newTestHandler(t, testutil.GetTestDB())

	rr := serveUnauthenticated(handler, http.MethodDelete, "/settings/two_factor_auth")

	assertRedirectToSignIn(t, rr, "/settings/two_factor_auth")
}

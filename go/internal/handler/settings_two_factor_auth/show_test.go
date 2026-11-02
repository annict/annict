package settings_two_factor_auth

import (
	"net/http"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/testutil"
)

// TestShow_Disabledは、二要素認証が無効なときに状態を警告のAlertで示し、アイコン付きの戻るリンク・登録画面への導線・
// ヘルプへのリンクを表示し、キャッシュさせないことを検証する。
func TestShow_Disabled(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	handler := newTestHandler(t, db)
	user := createCommittedUser(t, db)

	rr := serve(handler, newRequest(t, http.MethodGet, "/settings/two_factor_auth", nil, user, "ja"))

	if rr.Code != http.StatusOK {
		t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusOK)
	}
	assertNoStore(t, rr)
	body := rr.Body.String()
	assertContains(t, body,
		"<title>二要素認証 | Annict</title>",
		`<div class="alert" data-variant="warning"><h2>二要素認証は無効です</h2><section>有効にするには`,
		`href="/settings/two_factor_auth/new"`,
		// 戻るリンクの中の装飾のアイコンは支援技術から隠し、arrow-bend-up-left-regularのパスで描く
		`href="/settings/account" class="inline-flex w-fit items-center gap-1 text-sm underline"><svg aria-hidden="true" focusable="false"`,
		`<path d="M232,200a8,8`,
		`href="https://wikino.app/s/annict/pages/425"`,
		"二要素認証の設定のヘルプ",
	)
	if strings.Contains(body, "未使用のリカバリーコード") {
		t.Error("無効なのにリカバリーコードの数を表示している")
	}
}

// TestShow_Enabledは、二要素認証が有効なときに状態を成功のAlertで示し、未使用のリカバリーコードの数を表示することを検証する。
func TestShow_Enabled(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	handler := newTestHandler(t, db)
	user := createCommittedUser(t, db)
	enableForTest(t, db, user.ID, 7)

	rr := serve(handler, newRequest(t, http.MethodGet, "/settings/two_factor_auth", nil, user, "ja"))

	if rr.Code != http.StatusOK {
		t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusOK)
	}
	body := rr.Body.String()
	assertContains(t, body,
		`<div class="alert" data-variant="success"><h2>二要素認証は有効です</h2></div>`,
		"未使用のリカバリーコード", "7個",
		"二要素認証を無効にする",
		`name="_method" value="DELETE"`,
		`name="csrf_token"`,
		`id="credential"`,
		`type="password"`,
		`autocomplete="current-password"`,
	)
	if strings.Contains(body, `href="/settings/two_factor_auth/new"`) {
		t.Error("有効なのに登録画面への導線を表示している")
	}
}

// TestShow_RequiresAuthは、未ログインのリクエストを設定画面に通さずログイン画面へ送ることを検証する。
func TestShow_RequiresAuth(t *testing.T) {
	t.Parallel()

	handler := newTestHandler(t, testutil.GetTestDB())

	rr := serveUnauthenticated(handler, http.MethodGet, "/settings/two_factor_auth")

	assertRedirectToSignIn(t, rr, "/settings/two_factor_auth")
}

// TestShow_Englishは英語のロケールで文言が英語になることを検証する。
func TestShow_English(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	handler := newTestHandler(t, db)
	user := createCommittedUser(t, db)

	rr := serve(handler, newRequest(t, http.MethodGet, "/settings/two_factor_auth", nil, user, "en"))

	assertContains(t, rr.Body.String(),
		"<title>Two-Factor Authentication | Annict</title>",
		"Two-factor authentication is off",
		"Two-factor authentication setup help",
	)
}

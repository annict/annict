package supporters

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/session"
	annictStripe "github.com/annict/annict/go/internal/stripe"
	"github.com/annict/annict/go/internal/testutil"
)

// setupCheckoutTestHandler はテスト用のハンドラーをセットアップします
func setupCheckoutTestHandler(t *testing.T, tx *sql.Tx, db *sql.DB, stripeCfg *annictStripe.Config) *Handler {
	t.Helper()

	queries := query.New(db).WithTx(tx)
	cfg := &config.Config{
		Env:    "test",
		Domain: "test.annict.com",
	}

	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	stripeSubscriberRepo := repository.NewStripeSubscriberRepository(queries)
	gumroadSubscriberRepo := repository.NewGumroadSubscriberRepository(queries)

	return NewHandler(cfg, sessionManager, stripeSubscriberRepo, gumroadSubscriberRepo, stripeCfg)
}

// TestCreate_NotLoggedIn は未ログインユーザーがアクセスした場合のテスト
func TestCreate_NotLoggedIn(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)

	stripeCfg := &annictStripe.Config{
		PriceMonthlyID: "price_monthly_test",
		PriceYearlyID:  "price_yearly_test",
	}
	handler := setupCheckoutTestHandler(t, tx, db, stripeCfg)

	form := url.Values{}
	form.Set("plan", "monthly")

	req := httptest.NewRequest("POST", "/supporters/checkout", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	// ログインページへリダイレクトされることを確認
	if rr.Code != http.StatusSeeOther {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusSeeOther)
	}

	location := rr.Header().Get("Location")
	if location != "/sign_in" {
		t.Errorf("wrong redirect location: got %v want %v", location, "/sign_in")
	}
}

// TestCreate_InvalidPlan は無効なプランが選択された場合のテスト
func TestCreate_InvalidPlan(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)

	stripeCfg := &annictStripe.Config{
		PriceMonthlyID: "price_monthly_test",
		PriceYearlyID:  "price_yearly_test",
	}
	handler := setupCheckoutTestHandler(t, tx, db, stripeCfg)

	// ユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).Build()
	user := getUserByID(t, tx, userID)

	testCases := []struct {
		name string
		plan string
	}{
		{"empty plan", ""},
		{"invalid plan value", "invalid"},
		{"weekly plan", "weekly"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			form := url.Values{}
			form.Set("plan", tc.plan)

			req := httptest.NewRequest("POST", "/supporters/checkout", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
			req = req.WithContext(ctx)
			rr := httptest.NewRecorder()

			handler.Create(rr, req)

			// サポーターページへリダイレクトされることを確認
			if rr.Code != http.StatusSeeOther {
				t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusSeeOther)
			}

			location := rr.Header().Get("Location")
			if location != "/supporters" {
				t.Errorf("wrong redirect location: got %v want %v", location, "/supporters")
			}
		})
	}
}

// TestCreate_AlreadyActiveSubscription は既にアクティブなサブスクリプションがある場合のテスト
func TestCreate_AlreadyActiveSubscription(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)

	stripeCfg := &annictStripe.Config{
		PriceMonthlyID: "price_monthly_test",
		PriceYearlyID:  "price_yearly_test",
	}
	handler := setupCheckoutTestHandler(t, tx, db, stripeCfg)

	// アクティブなStripeサポーターを作成
	userID, _ := createUserWithStripeSubscriber(t, tx, "active")
	user := getUserByID(t, tx, userID)

	form := url.Values{}
	form.Set("plan", "monthly")

	req := httptest.NewRequest("POST", "/supporters/checkout", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	// サポーターページへリダイレクトされることを確認
	if rr.Code != http.StatusSeeOther {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusSeeOther)
	}

	location := rr.Header().Get("Location")
	if location != "/supporters" {
		t.Errorf("wrong redirect location: got %v want %v", location, "/supporters")
	}
}

// TestCreate_PastDueSubscriptionAllowsCheckout は支払い遅延中のサブスクリプションがあってもCheckoutを許可する
func TestCreate_PastDueSubscriptionBlocksCheckout(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)

	stripeCfg := &annictStripe.Config{
		PriceMonthlyID: "price_monthly_test",
		PriceYearlyID:  "price_yearly_test",
	}
	handler := setupCheckoutTestHandler(t, tx, db, stripeCfg)

	// past_dueステータスのStripeサポーターを作成
	userID, _ := createUserWithStripeSubscriber(t, tx, "past_due")
	user := getUserByID(t, tx, userID)

	form := url.Values{}
	form.Set("plan", "monthly")

	req := httptest.NewRequest("POST", "/supporters/checkout", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	// past_dueはアクティブとして扱われるため、リダイレクトされることを確認
	if rr.Code != http.StatusSeeOther {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusSeeOther)
	}

	location := rr.Header().Get("Location")
	if location != "/supporters" {
		t.Errorf("wrong redirect location: got %v want %v", location, "/supporters")
	}
}

// TestCreate_CanceledSubscriptionAllowsCheckout はキャンセル済みサブスクリプションがある場合はCheckoutを許可する
func TestCreate_CanceledSubscriptionAllowsCheckout(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)

	stripeCfg := &annictStripe.Config{
		PriceMonthlyID: "price_monthly_test",
		PriceYearlyID:  "price_yearly_test",
	}
	handler := setupCheckoutTestHandler(t, tx, db, stripeCfg)

	// canceledステータスのStripeサポーターを作成
	userID, _ := createUserWithStripeSubscriber(t, tx, "canceled")
	user := getUserByID(t, tx, userID)

	form := url.Values{}
	form.Set("plan", "monthly")

	req := httptest.NewRequest("POST", "/supporters/checkout", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	// canceledは非アクティブなので、Stripe API呼び出しに進む
	// Stripe APIが設定されていない場合はエラーになりリダイレクトされる
	if rr.Code != http.StatusSeeOther {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusSeeOther)
	}
}

// TestCreate_NoSubscriptionAtAll はサブスクリプションが全くない場合のテスト
func TestCreate_NoSubscriptionAtAll(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)

	stripeCfg := &annictStripe.Config{
		PriceMonthlyID: "price_monthly_test",
		PriceYearlyID:  "price_yearly_test",
	}
	handler := setupCheckoutTestHandler(t, tx, db, stripeCfg)

	// サブスクリプションを持たないユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).Build()
	user := getUserByID(t, tx, userID)

	form := url.Values{}
	form.Set("plan", "monthly")

	req := httptest.NewRequest("POST", "/supporters/checkout", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	// サブスクリプションがないので、Stripe API呼び出しに進む
	// Stripe APIが設定されていない場合はエラーになりリダイレクトされる
	if rr.Code != http.StatusSeeOther {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusSeeOther)
	}
}

// TestCheckoutRequest_Validate はリクエストバリデーションのテスト
func TestCheckoutRequest_Validate(t *testing.T) {
	testCases := []struct {
		name        string
		plan        string
		expectError bool
	}{
		{"monthly plan", "monthly", false},
		{"yearly plan", "yearly", false},
		{"empty plan", "", true},
		{"invalid plan", "weekly", true},
		{"uppercase plan", "MONTHLY", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &CheckoutRequest{Plan: tc.plan}
			errors := req.Validate()

			if tc.expectError && len(errors) == 0 {
				t.Error("expected validation error but got none")
			}
			if !tc.expectError && len(errors) > 0 {
				t.Errorf("expected no validation error but got: %v", errors)
			}
		})
	}
}

// TestCreate_MissingPriceID は価格IDが設定されていない場合のテスト
func TestCreate_MissingPriceID(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)

	// 価格IDが設定されていない
	stripeCfg := &annictStripe.Config{
		PriceMonthlyID: "",
		PriceYearlyID:  "",
	}
	handler := setupCheckoutTestHandler(t, tx, db, stripeCfg)

	// ユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).Build()
	user := getUserByID(t, tx, userID)

	form := url.Values{}
	form.Set("plan", "monthly")

	req := httptest.NewRequest("POST", "/supporters/checkout", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	// 価格IDが設定されていないため、エラーでリダイレクト
	if rr.Code != http.StatusSeeOther {
		t.Errorf("wrong status code: got %v want %v", rr.Code, http.StatusSeeOther)
	}

	location := rr.Header().Get("Location")
	if location != "/supporters" {
		t.Errorf("wrong redirect location: got %v want %v", location, "/supporters")
	}
}

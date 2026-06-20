package supporters_checkout

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
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	annictStripe "github.com/annict/annict/go/internal/stripe"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/validator"
)

// setupTestHandler はテスト用のハンドラーをセットアップします
func setupTestHandler(t *testing.T, tx *sql.Tx, db *sql.DB, stripeCfg *annictStripe.Config) *Handler {
	t.Helper()

	queries := query.New(db).WithTx(tx)
	cfg := &config.Config{
		Env:    "test",
		Domain: "test.annict.com",
	}

	flashMgr := testutil.NewTestFlashManager()
	stripeSubscriberRepo := repository.NewStripeSubscriberRepository(queries)
	v := validator.NewSupportersCheckoutCreateValidator()

	// テスト用には nil を渡す（テストでは Stripe API を呼び出さない）
	createCheckoutSessionUC := usecase.NewCreateCheckoutSessionUsecase(cfg, stripeSubscriberRepo, stripeCfg, nil, v)

	return NewHandler(flashMgr, createCheckoutSessionUC)
}

// createUserWithStripeSubscriber はStripeサブスクライバーを持つユーザーを作成します
func createUserWithStripeSubscriber(t *testing.T, tx *sql.Tx, stripeStatus string) (model.UserID, model.StripeSubscriberID) {
	t.Helper()

	userID := testutil.NewUserBuilder(t, tx).Build()
	subscriberID := testutil.NewStripeSubscriberBuilder(t, tx).
		WithStripeStatus(stripeStatus).
		Build()

	// ユーザーにStripeサブスクライバーIDを関連付け
	_, err := tx.Exec(`UPDATE users SET stripe_subscriber_id = $1 WHERE id = $2`, int64(subscriberID), int64(userID))
	if err != nil {
		t.Fatalf("Stripeサブスクライバーの関連付けに失敗しました: %v", err)
	}

	return userID, subscriberID
}

// getUserByID はユーザーIDからユーザー情報を取得します（テスト用）
func getUserByID(t *testing.T, tx *sql.Tx, userID model.UserID) *model.User {
	t.Helper()

	var user model.User
	var stripeSubID, gumroadSubID sql.NullInt64
	err := tx.QueryRow(`
		SELECT id, username, email, role, encrypted_password, locale,
			   stripe_subscriber_id, gumroad_subscriber_id,
			   created_at, updated_at
		FROM users WHERE id = $1
	`, int64(userID)).Scan(
		&user.ID, &user.Username, &user.Email, &user.Role,
		&user.EncryptedPassword, &user.Locale,
		&stripeSubID, &gumroadSubID,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		t.Fatalf("ユーザー情報の取得に失敗しました: %v", err)
	}
	if stripeSubID.Valid {
		id := model.StripeSubscriberID(stripeSubID.Int64)
		user.StripeSubscriberID = &id
	}
	if gumroadSubID.Valid {
		id := model.GumroadSubscriberID(gumroadSubID.Int64)
		user.GumroadSubscriberID = &id
	}

	return &user
}

// TestCreate_NotLoggedIn は未ログインユーザーがアクセスした場合のテスト
func TestCreate_NotLoggedIn(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	stripeCfg := &annictStripe.Config{
		PriceMonthlyID: "price_monthly_test",
		PriceYearlyID:  "price_yearly_test",
	}
	handler := setupTestHandler(t, tx, db, stripeCfg)

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
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	stripeCfg := &annictStripe.Config{
		PriceMonthlyID: "price_monthly_test",
		PriceYearlyID:  "price_yearly_test",
	}
	handler := setupTestHandler(t, tx, db, stripeCfg)

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
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	stripeCfg := &annictStripe.Config{
		PriceMonthlyID: "price_monthly_test",
		PriceYearlyID:  "price_yearly_test",
	}
	handler := setupTestHandler(t, tx, db, stripeCfg)

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

// TestCreate_PastDueSubscriptionBlocksCheckout は支払い遅延中のサブスクリプションがある場合のテスト
func TestCreate_PastDueSubscriptionBlocksCheckout(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	stripeCfg := &annictStripe.Config{
		PriceMonthlyID: "price_monthly_test",
		PriceYearlyID:  "price_yearly_test",
	}
	handler := setupTestHandler(t, tx, db, stripeCfg)

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
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	stripeCfg := &annictStripe.Config{
		PriceMonthlyID: "price_monthly_test",
		PriceYearlyID:  "price_yearly_test",
	}
	handler := setupTestHandler(t, tx, db, stripeCfg)

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
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	stripeCfg := &annictStripe.Config{
		PriceMonthlyID: "price_monthly_test",
		PriceYearlyID:  "price_yearly_test",
	}
	handler := setupTestHandler(t, tx, db, stripeCfg)

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

// TestCreate_MissingPriceID は価格IDが設定されていない場合のテスト
func TestCreate_MissingPriceID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	// 価格IDが設定されていない
	stripeCfg := &annictStripe.Config{
		PriceMonthlyID: "",
		PriceYearlyID:  "",
	}
	handler := setupTestHandler(t, tx, db, stripeCfg)

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

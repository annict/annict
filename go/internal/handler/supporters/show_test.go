package supporters

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/image"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/session"
	annictStripe "github.com/annict/annict/go/internal/stripe"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/usecase"
)

// setupTestHandlerはテスト用のハンドラーをセットアップします
func setupTestHandler(t *testing.T, tx *sql.Tx, db *sql.DB) *Handler {
	t.Helper()

	queries := query.New(db).WithTx(tx)
	cfg := &config.Config{
		Env:    "test",
		Domain: "test.annict.com",
	}

	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	imageHelper := image.NewHelper(cfg)
	stripeSubscriberRepo := repository.NewStripeSubscriberRepository(queries)
	gumroadSubscriberRepo := repository.NewGumroadSubscriberRepository(queries)

	// テスト用のStripe設定 (テストではStripe APIを呼び出さないため空でOK)
	stripeCfg := &annictStripe.Config{}

	getSupporterStatusUC := usecase.NewGetSupporterStatusUsecase(stripeSubscriberRepo, gumroadSubscriberRepo)

	// テスト用にはnilを渡す (テストではStripe APIを呼び出さない)
	return NewHandler(cfg, sessionManager, imageHelper, getSupporterStatusUC, stripeCfg, nil)
}

// createUserWithStripeSubscriberはStripeサブスクライバーを持つユーザーを作成します
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

// createUserWithGumroadSubscriberはGumroadサブスクライバーを持つユーザーを作成します
func createUserWithGumroadSubscriber(t *testing.T, tx *sql.Tx, ended bool) (model.UserID, model.GumroadSubscriberID) {
	t.Helper()

	userID := testutil.NewUserBuilder(t, tx).Build()

	builder := testutil.NewGumroadSubscriberBuilder(t, tx)
	if ended {
		// 過去の日時を設定 (終了済み)
		builder = builder.WithGumroadEndedAt(time.Now().AddDate(-1, 0, 0))
	}
	subscriberID := builder.Build()

	// ユーザーにGumroadサブスクライバーIDを関連付け
	_, err := tx.Exec(`UPDATE users SET gumroad_subscriber_id = $1 WHERE id = $2`, int64(subscriberID), int64(userID))
	if err != nil {
		t.Fatalf("Gumroadサブスクライバーの関連付けに失敗しました: %v", err)
	}

	return userID, subscriberID
}

// getUserByIDはユーザーIDからユーザー情報を取得します (テスト用)
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

// TestShow_NotLoggedInは未ログインユーザーの場合のテスト
func TestShow_NotLoggedIn(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := setupTestHandler(t, tx, db)

	req := httptest.NewRequest("GET", "/supporters", nil)
	rr := httptest.NewRecorder()

	handler.Show(rr, req)

	// ステータスコードを確認
	if rr.Code != http.StatusOK {
		t.Errorf("ステータスコード = %v、期待値 = %v", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()

	// ログイン促進セクションが表示されることを確認
	expectedContents := []string{
		"サポーター", // ページタイトル
		"ログイン",  // ログインボタン
		// canonicalはページ自身を指し、og:urlも同じURLを宣言する。
		`<link rel="canonical" href="https://test.annict.com/supporters">`,
		`<meta property="og:url" content="https://test.annict.com/supporters">`,
	}

	for _, expected := range expectedContents {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに含まれていない文字列 = %q", expected)
		}
	}

	// Content-Typeを確認
	if ct := rr.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %v、期待値 = %v", ct, "text/html; charset=utf-8")
	}
}

// TestShow_LoggedIn_NotSupporterはログイン済み・非サポーターユーザーの場合のテスト
func TestShow_LoggedIn_NotSupporter(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := setupTestHandler(t, tx, db)

	// ユーザーを作成 (サブスクリプションなし)
	userID := testutil.NewUserBuilder(t, tx).Build()
	user := getUserByID(t, tx, userID)

	req := httptest.NewRequest("GET", "/supporters", nil)
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.Show(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("ステータスコード = %v、期待値 = %v", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()

	// Checkoutセクションが表示されることを確認 (非サポーター向け)
	expectedContents := []string{
		"サポーター",                           // ページタイトル
		"¥290",                            // 月額プラン
		"¥2,900",                          // 年額プラン
		"action=\"/supporters/checkout\"", // Checkoutフォーム
	}

	for _, expected := range expectedContents {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに含まれていない文字列 = %q", expected)
		}
	}
}

// TestShow_LoggedIn_NotSupporter_Conditionsは非サポーターの申込み画面に契約条件が表示されることを確認するテスト
func TestShow_LoggedIn_NotSupporter_Conditions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		locale   string
		expected []string
	}{
		{
			name:   "日本語",
			locale: i18n.LangJa,
			expected: []string{
				"1年間続けた場合の総額の目安: ¥3,480",
				"お申し込みの前にご確認ください",
				"解約するまで1か月 (月額プラン) または1年 (年額プラン) ごとに自動で更新・請求されます",
				"クレジットカード (JCBを含む)、Apple Pay、Linkでお支払いいただけます (Stripe決済)",
				"お支払いが完了すると、すぐに特典をご利用いただけます",
				"解約後もお支払い済みの期間が終わるまで特典をご利用いただけます",
				"お支払い済みの料金は返金いたしません (二重に請求した場合など、運営者の責任による場合を除きます)",
				"退会すると、その時点で解約され、残りの期間の料金は返金されません",
				"支払時期・解約・返金などの詳しい条件は、次のページでご確認いただけます",
				// ページと同じ言語の版を指すため、hreflangを付けない
				`<a href="/legal" class="link">`,
				"特定商取引法に基づく表記",
				`<a href="/terms#第11条-サポーター" class="link">`,
				"利用規約 (第11条 サポーター)",
			},
		},
		{
			name:   "英語",
			locale: i18n.LangEn,
			expected: []string{
				"Estimated total for one year: ¥3,480",
				"Before you subscribe",
				"renews automatically every month (Monthly Plan) or every year (Yearly Plan)",
				"Pay by credit card (including JCB), Apple Pay, or Link (powered by Stripe).",
				"Benefits are available as soon as your payment is complete.",
				"Benefits remain available until the end of the period you have already paid for.",
				"Payments already made are non-refundable, except in cases caused by us, such as a duplicate charge.",
				"If you delete your account, your subscription is canceled immediately and the remaining period is not refunded.",
				"Full details on billing, cancellation, and refunds are available on the following pages.",
				// 特定商取引法に基づく表記は日本語版だけのため、日本語版を指してhreflangと言語を添える
				`<a href="/legal" class="link" hreflang="ja">`,
				"Legal Notice (Specified Commercial Transactions Act, in Japanese)",
				// 利用規約は英語版のサポーターの条を指す
				`<a href="/en/terms#article-11-annict-supporters" class="link">`,
				"Terms of Service (Article 11: Annict Supporters)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db, tx := testutil.SetupTx(t)
			handler := setupTestHandler(t, tx, db)

			userID := testutil.NewUserBuilder(t, tx).Build()
			user := getUserByID(t, tx, userID)

			req := httptest.NewRequest("GET", "/supporters", nil)
			ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
			ctx = i18n.SetLocale(ctx, tt.locale)
			req = req.WithContext(ctx)
			rr := httptest.NewRecorder()

			handler.Show(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("ステータスコード = %v、期待値 = %v", rr.Code, http.StatusOK)
			}

			body := rr.Body.String()
			for _, expected := range tt.expected {
				if !strings.Contains(body, expected) {
					t.Errorf("レスポンスに含まれていない文字列 = %q", expected)
				}
			}
		})
	}
}

// TestShow_StripeSupporterはStripeサポーターの場合のテスト
func TestShow_StripeSupporter(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := setupTestHandler(t, tx, db)

	// アクティブなStripeサポーターを作成
	userID, _ := createUserWithStripeSubscriber(t, tx, "active")
	user := getUserByID(t, tx, userID)

	req := httptest.NewRequest("GET", "/supporters", nil)
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.Show(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("ステータスコード = %v、期待値 = %v", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()

	// Stripeサポーターセクションが表示されることを確認
	expectedContents := []string{
		"サポーター",                         // ページタイトル
		"action=\"/supporters/portal\"", // 管理ポータルフォーム
	}

	for _, expected := range expectedContents {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに含まれていない文字列 = %q", expected)
		}
	}

	// 非サポーター向けのCheckoutセクションが表示されないことを確認
	if strings.Contains(body, "action=\"/supporters/checkout\"") {
		t.Error("Stripeのサポーターに対してレスポンスが決済フォームを含んでいる")
	}
	if strings.Contains(body, "お申し込みの前にご確認ください") {
		t.Error("Stripeのサポーターに対してレスポンスが申込み前の契約条件を含んでいる")
	}
}

// TestShow_GumroadSupporterはGumroadサポーターの場合のテスト
func TestShow_GumroadSupporter(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := setupTestHandler(t, tx, db)

	// アクティブなGumroadサポーターを作成
	userID, _ := createUserWithGumroadSubscriber(t, tx, false)
	user := getUserByID(t, tx, userID)

	req := httptest.NewRequest("GET", "/supporters", nil)
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.Show(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("ステータスコード = %v、期待値 = %v", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()

	// Gumroadサポーターセクションが表示されることを確認
	// (移行案内メッセージが含まれる緑のセクション)
	if !strings.Contains(body, "bg-green-50") {
		t.Error("レスポンスにGumroadのサポーター向けセクションが含まれていない (緑の背景)")
	}

	// 非サポーター向けのCheckoutセクションが表示されないことを確認
	if strings.Contains(body, "action=\"/supporters/checkout\"") {
		t.Error("Gumroadのサポーターに対してレスポンスが決済フォームを含んでいる")
	}
}

// TestShow_BothActiveは両方アクティブな場合のテスト
func TestShow_BothActive(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := setupTestHandler(t, tx, db)

	// StripeとGumroad両方のサブスクリプションを持つユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).Build()

	stripeSubscriberID := testutil.NewStripeSubscriberBuilder(t, tx).
		WithStripeStatus("active").
		Build()
	gumroadSubscriberID := testutil.NewGumroadSubscriberBuilder(t, tx).Build()

	// 両方を関連付け
	_, err := tx.Exec(`UPDATE users SET stripe_subscriber_id = $1, gumroad_subscriber_id = $2 WHERE id = $3`,
		stripeSubscriberID, gumroadSubscriberID, userID)
	if err != nil {
		t.Fatalf("サブスクライバーの関連付けに失敗しました: %v", err)
	}

	user := getUserByID(t, tx, userID)

	req := httptest.NewRequest("GET", "/supporters", nil)
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.Show(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("ステータスコード = %v、期待値 = %v", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()

	// 両方アクティブの場合、Stripeセクションが優先表示される
	expectedContents := []string{
		"action=\"/supporters/portal\"", // Stripe管理ポータル
	}

	for _, expected := range expectedContents {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに含まれていない文字列 = %q", expected)
		}
	}
}

// TestShow_SuccessQueryParamはsuccessクエリパラメータがある場合のテスト
func TestShow_SuccessQueryParam(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := setupTestHandler(t, tx, db)

	req := httptest.NewRequest("GET", "/supporters?success=true", nil)
	rr := httptest.NewRecorder()

	handler.Show(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("ステータスコード = %v、期待値 = %v", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()

	// 成功メッセージが表示されることを確認
	if !strings.Contains(body, `<div class="alert" data-variant="success">`) {
		t.Error("レスポンスに成功メッセージが含まれていない (successのalert)")
	}
}

// TestShow_CanceledQueryParamはcanceledクエリパラメータがある場合のテスト
func TestShow_CanceledQueryParam(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := setupTestHandler(t, tx, db)

	req := httptest.NewRequest("GET", "/supporters?canceled=true", nil)
	rr := httptest.NewRecorder()

	handler.Show(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("ステータスコード = %v、期待値 = %v", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()

	// キャンセルメッセージが表示されることを確認
	if !strings.Contains(body, `<div class="alert" data-variant="warning">`) {
		t.Error("レスポンスに決済キャンセルのメッセージが含まれていない (warningのalert)")
	}
}

// TestShow_InactiveStripeSubscriptionは非アクティブなStripeサブスクリプションの場合のテスト
func TestShow_InactiveStripeSubscription(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := setupTestHandler(t, tx, db)

	// キャンセル済みのStripeサポーターを作成
	userID, _ := createUserWithStripeSubscriber(t, tx, "canceled")
	user := getUserByID(t, tx, userID)

	req := httptest.NewRequest("GET", "/supporters", nil)
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.Show(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("ステータスコード = %v、期待値 = %v", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()

	// 非アクティブなので、Checkoutセクションが表示される
	if !strings.Contains(body, "action=\"/supporters/checkout\"") {
		t.Error("無効なサブスクリプションでレスポンスに決済フォームが含まれていない")
	}
}

// TestShow_EndedGumroadSubscriptionは終了したGumroadサブスクリプションの場合のテスト
func TestShow_EndedGumroadSubscription(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler := setupTestHandler(t, tx, db)

	// 終了したGumroadサポーターを作成
	userID, _ := createUserWithGumroadSubscriber(t, tx, true)
	user := getUserByID(t, tx, userID)

	req := httptest.NewRequest("GET", "/supporters", nil)
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.Show(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("ステータスコード = %v、期待値 = %v", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()

	// 終了済みなので、Checkoutセクションが表示される
	if !strings.Contains(body, "action=\"/supporters/checkout\"") {
		t.Error("終了済みのサブスクリプションでレスポンスに決済フォームが含まれていない")
	}
}

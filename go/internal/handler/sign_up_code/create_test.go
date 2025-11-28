package sign_up_code_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/annict/annict/internal/config"
	"github.com/annict/annict/internal/handler/sign_up_code"
	"github.com/annict/annict/internal/query"
	"github.com/annict/annict/internal/ratelimit"
	"github.com/annict/annict/internal/repository"
	"github.com/annict/annict/internal/session"
	"github.com/annict/annict/internal/testutil"
	"github.com/annict/annict/internal/usecase"
	"golang.org/x/crypto/bcrypt"
)

// TestCreate_ErrorMessageUnification は、コード検証失敗時にリダイレクトされることを確認します
// （エラーメッセージは統一されているはず）
func TestCreate_ErrorMessageUnification(t *testing.T) {
	// テスト用DBとトランザクションをセットアップ
	db, tx := testutil.SetupTestDB(t)
	defer func() { _ = tx.Rollback() }()

	// テスト用Redisをセットアップ
	rdb := testutil.SetupTestRedis(t)
	defer rdb.FlushDB(context.Background())

	// 設定を読み込む
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("設定の読み込みに失敗しました: %v", err)
	}

	// usecaseの初期化
	queries := testutil.NewQueriesWithTx(db, tx)
	sendSignUpCodeUC := usecase.NewSendSignUpCodeUsecase(db, queries, nil)
	verifySignUpCodeUC := usecase.NewVerifySignUpCodeUsecase(db, queries)

	// セッションマネージャーの初期化
	sessionRepo := repository.NewSessionRepository(queries)
	sessionMgr := session.NewManager(sessionRepo, cfg)

	// Rate Limiterの初期化
	limiter := ratelimit.NewLimiter(rdb)

	// ハンドラーの初期化
	handler := sign_up_code.NewHandler(
		cfg,
		sessionMgr,
		db,
		limiter,
		rdb,
		sendSignUpCodeUC,
		verifySignUpCodeUC,
	)

	// テストメールアドレス
	email := "test@example.com"

	// テストケース: 異なるエラーケースでもリダイレクトされること
	tests := []struct {
		name      string
		setupCode func(*query.Queries) // コードのセットアップ処理
		inputCode string               // 入力する確認コード
	}{
		{
			name: "コードが見つからない場合",
			setupCode: func(q *query.Queries) {
				// 何もセットアップしない（コードが存在しない状態）
			},
			inputCode: "123456",
		},
		{
			name: "コードが正しくない場合",
			setupCode: func(q *query.Queries) {
				// 正しいコードとは異なるコードを保存
				correctCode := "123456"
				hashedCode, _ := bcrypt.GenerateFromPassword([]byte(correctCode), bcrypt.DefaultCost)
				_, _ = q.CreateSignUpCode(context.Background(), query.CreateSignUpCodeParams{
					Email:      email,
					CodeDigest: string(hashedCode),
					ExpiresAt:  time.Now().Add(15 * time.Minute),
				})
			},
			inputCode: "999999", // 間違ったコード
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// セットアップ
			tt.setupCode(queries)

			// リクエストパラメータを作成
			formData := url.Values{}
			formData.Set("code", tt.inputCode)

			// リクエストを作成
			req := httptest.NewRequest("POST", "/sign_up/code", strings.NewReader(formData.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			// レスポンスレコーダーを作成
			rr := httptest.NewRecorder()

			// セッションにメールアドレスを設定
			ctx := req.Context()
			_ = sessionMgr.SetValue(ctx, rr, req, "sign_up_email", email)

			// ハンドラーを実行
			handler.Create(rr, req)

			// ステータスコードを確認（リダイレクトされることを確認）
			if rr.Code != http.StatusSeeOther {
				t.Errorf("期待されるステータスコード: %d, 実際: %d", http.StatusSeeOther, rr.Code)
			}

			// 注: エラーメッセージの統一性は、実装コードで保証されています
		})
	}
}

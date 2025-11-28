package sign_in

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/annict/annict/internal/config"
	"github.com/annict/annict/internal/query"
	"github.com/annict/annict/internal/repository"
	"github.com/annict/annict/internal/session"
	"github.com/annict/annict/internal/testutil"
	"github.com/annict/annict/internal/turnstile"
	"github.com/annict/annict/internal/usecase"
)

func TestHandler_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		email             string
		userExists        bool
		hasPassword       bool
		wantStatus        int
		wantLocation      string
		wantSessionEmail  bool
		wantSessionUserID bool
		wantFormError     bool
		description       string
	}{
		{
			name:              "正常系 - パスワードありユーザー",
			email:             "user-with-password@example.com",
			userExists:        true,
			hasPassword:       true,
			wantStatus:        http.StatusSeeOther,
			wantLocation:      "/sign_in/password",
			wantSessionEmail:  true,
			wantSessionUserID: true,
			description:       "パスワードが存在する場合、/sign_in/password へリダイレクト",
		},
		{
			name:              "正常系 - パスワードなしユーザー",
			email:             "user-without-password@example.com",
			userExists:        true,
			hasPassword:       false,
			wantStatus:        http.StatusSeeOther,
			wantLocation:      "/sign_in/code",
			wantSessionEmail:  true,
			wantSessionUserID: true,
			description:       "パスワードが存在しない場合、/sign_in/code へリダイレクト",
		},
		{
			name:          "異常系 - メールアドレスが空",
			email:         "",
			userExists:    false,
			wantStatus:    http.StatusSeeOther,
			wantLocation:  "/sign_in",
			wantFormError: true,
			description:   "メールアドレスが空の場合、バリデーションエラーで /sign_in へリダイレクト",
		},
		{
			name:          "異常系 - ユーザーが存在しない",
			email:         "notfound@example.com",
			userExists:    false,
			wantStatus:    http.StatusSeeOther,
			wantLocation:  "/sign_in",
			wantFormError: true,
			description:   "ユーザーが存在しない場合、エラーメッセージを表示して /sign_in へリダイレクト",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// パスワードなしユーザーの場合、トランザクションをコミットするため、
			// 並列実行を無効にして競合を避ける
			if tt.hasPassword || !tt.userExists {
				t.Parallel()
			}

			// テスト用DBとトランザクションをセットアップ
			db, tx := testutil.SetupTestDB(t)

			// テスト用ユーザーを作成
			var testEmail string
			if tt.userExists {
				// テストごとにユニークなメールアドレスを生成（タイムスタンプを追加）
				testEmail = fmt.Sprintf("%d-%s", time.Now().UnixNano(), tt.email)
				// テストケースごとにユニークなユーザー名を生成
				username := strings.ReplaceAll(testEmail, "@", "_at_")
				username = strings.ReplaceAll(username, ".", "_")
				builder := testutil.NewUserBuilder(t, tx).
					WithEmail(testEmail).
					WithUsername(username)
				if !tt.hasPassword {
					builder = builder.WithEncryptedPassword("")
				}
				builder.Build()

				// パスワードなしユーザーの場合、ユースケースが新しいトランザクションを開始するため、
				// ユーザーをDBに永続化する必要があります
				if !tt.hasPassword {
					if err := tx.Commit(); err != nil {
						t.Fatalf("トランザクションのコミットに失敗しました: %v", err)
					}
					// 新しいトランザクションを開始
					newTx, err := db.Begin()
					if err != nil {
						t.Fatalf("新しいトランザクションの開始に失敗しました: %v", err)
					}
					tx = newTx
					// テスト終了時に新しいトランザクションをロールバック（元のt.Cleanupは既に設定済み）
					t.Cleanup(func() {
						if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
							t.Errorf("トランザクションのロールバックに失敗しました: %v", err)
						}
					})
				}
			} else {
				testEmail = tt.email
			}

			// sqlcリポジトリを作成
			queries := query.New(db).WithTx(tx)

			// 設定とセッションマネージャーを作成
			cfg := &config.Config{
				CookieDomain:  ".example.com",
				SessionSecure: "false",
			}
			sessionRepo := repository.NewSessionRepository(queries)
			sessionMgr := session.NewManager(sessionRepo, cfg)

			// UserRepositoryを作成
			userRepo := repository.NewUserRepository(queries)

			// ログインコード送信ユースケースを作成（riverClient は nil でメール送信をスキップ）
			sendSignInCodeUC := usecase.NewSendSignInCodeUsecase(db, queries, nil)

			// Turnstile クライアントを作成（テスト環境用: 空のSecretKeyで検証をスキップ）
			turnstileClient := turnstile.NewClient("", "")

			// ハンドラーを作成
			handler := NewHandler(cfg, sessionMgr, userRepo, sendSignInCodeUC, turnstileClient)

			// テスト用HTTPリクエストを作成
			form := url.Values{}
			form.Add("email", testEmail)
			req := httptest.NewRequest("POST", "/sign_in", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			// ResponseRecorderを作成
			rr := httptest.NewRecorder()

			// ハンドラーを実行
			handler.Create(rr, req)

			// ステータスコードを検証
			if rr.Code != tt.wantStatus {
				t.Errorf("wrong status code: got %v want %v", rr.Code, tt.wantStatus)
			}

			// リダイレクト先を検証
			if tt.wantLocation != "" {
				location := rr.Header().Get("Location")
				if location != tt.wantLocation {
					t.Errorf("wrong location: got %v want %v", location, tt.wantLocation)
				}
			}

			// TODO: セッション検証はセッション実装が完成したら追加する
		})
	}
}

package testutil

import (
	"database/sql"
	"testing"

	"github.com/annict/annict/internal/config"
	"github.com/annict/annict/internal/image"
	"github.com/annict/annict/internal/query"
	"github.com/annict/annict/internal/repository"
	"github.com/annict/annict/internal/session"
)

// NewTestSessionManager はテスト用のセッションマネージャーを作成します
func NewTestSessionManager(t *testing.T) *session.Manager {
	t.Helper()

	// テスト用のDBセットアップ
	db, _ := SetupTestDB(t)

	// sqlcリポジトリを作成
	queries := query.New(db)

	// SessionRepositoryを作成
	sessionRepo := repository.NewSessionRepository(queries)

	// テスト用のConfig
	cfg := &config.Config{
		CookieDomain:    ".test.example.com",
		SessionSecure:   "false",
		SessionHTTPOnly: "true",
	}

	return session.NewManager(sessionRepo, cfg)
}

// NewTestSessionManagerWithDB はDB接続を受け取ってセッションマネージャーを作成します
func NewTestSessionManagerWithDB(db *sql.DB) *session.Manager {
	queries := query.New(db)

	// SessionRepositoryを作成
	sessionRepo := repository.NewSessionRepository(queries)

	// テスト用のConfig
	cfg := &config.Config{
		CookieDomain:    ".test.example.com",
		SessionSecure:   "false",
		SessionHTTPOnly: "true",
	}

	return session.NewManager(sessionRepo, cfg)
}

// NewTestImageHelper はテスト用の画像ヘルパーを作成します
func NewTestImageHelper() *image.Helper {
	cfg := &config.Config{
		Env:              "test",
		ImgproxyEndpoint: "http://localhost:18080",
		ImgproxyKey:      "test-key",
		ImgproxySalt:     "test-salt",
		S3BucketName:     "test-bucket",
	}

	return image.NewHelper(cfg)
}

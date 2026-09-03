package config

import (
	"os"
	"reflect"
	"testing"
)

// TestLoad は環境変数から直接設定を読み込むテスト
func TestLoad(t *testing.T) {
	// 元のワーキングディレクトリを保存
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(originalWd) }() // テスト終了後に元のディレクトリに戻す

	// テストの前にワーキングディレクトリをプロジェクトルート（go/）に変更
	if err := os.Chdir("../.."); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// 必須の環境変数を設定
	_ = os.Setenv("APP_ENV", "dev")
	_ = os.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test")
	_ = os.Setenv("ANNICT_PORT", "4004")
	_ = os.Setenv("ANNICT_DOMAIN", "test.example.com")
	_ = os.Setenv("ANNICT_COOKIE_DOMAIN", ".test.example.com")
	_ = os.Setenv("ANNICT_SESSION_SECURE", "false")
	_ = os.Setenv("ANNICT_SESSION_HTTPONLY", "true")
	_ = os.Setenv("ANNICT_S3_BUCKET_NAME", "test-bucket")
	_ = os.Setenv("ANNICT_IMGPROXY_ENDPOINT", "http://test:8080")
	_ = os.Setenv("ANNICT_IMGPROXY_KEY", "test-key")
	_ = os.Setenv("ANNICT_IMGPROXY_SALT", "test-salt")

	// クリーンアップ処理をdeferで確実に実行
	defer func() {
		_ = os.Unsetenv("APP_ENV")
		_ = os.Unsetenv("DATABASE_URL")
		_ = os.Unsetenv("ANNICT_PORT")
		_ = os.Unsetenv("ANNICT_DOMAIN")
		_ = os.Unsetenv("ANNICT_COOKIE_DOMAIN")
		_ = os.Unsetenv("ANNICT_SESSION_SECURE")
		_ = os.Unsetenv("ANNICT_SESSION_HTTPONLY")
		_ = os.Unsetenv("ANNICT_S3_BUCKET_NAME")
		_ = os.Unsetenv("ANNICT_IMGPROXY_ENDPOINT")
		_ = os.Unsetenv("ANNICT_IMGPROXY_KEY")
		_ = os.Unsetenv("ANNICT_IMGPROXY_SALT")
		_ = os.Unsetenv("ANNICT_TURNSTILE_SITE_KEY")
		_ = os.Unsetenv("ANNICT_TURNSTILE_SECRET_KEY")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// 基本的な設定が読み込まれていることを確認
	if cfg.DatabaseURL == "" {
		t.Error("DatabaseURL should not be empty")
	}
	if cfg.Port == "" {
		t.Error("Port should not be empty")
	}
	if cfg.Env != "dev" {
		t.Errorf("Env = %v, want dev", cfg.Env)
	}

	t.Logf("Config loaded successfully:")
	t.Logf("  Env: %s", cfg.Env)
	t.Logf("  DatabaseURL: %s", cfg.DatabaseURL)
	t.Logf("  Port: %s", cfg.Port)
	t.Logf("  RedisURL: %s", cfg.RedisURL)
	t.Logf("  ResendAPIKey: %s", cfg.ResendAPIKey)
	t.Logf("  ResendFromEmail: %s", cfg.ResendFromEmail)
}

func TestDatabaseDSN(t *testing.T) {
	cfg := &Config{
		DatabaseURL: "postgres://user:pass@localhost:5432/testdb?sslmode=disable",
	}

	dsn := cfg.DatabaseDSN()
	expected := "postgres://user:pass@localhost:5432/testdb?sslmode=disable"

	if dsn != expected {
		t.Errorf("DatabaseDSN() = %v, want %v", dsn, expected)
	}
}

func TestIsDev(t *testing.T) {
	tests := []struct {
		env  string
		want bool
	}{
		{"dev", true},
		{"test", false},
		{"prod", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.env, func(t *testing.T) {
			cfg := &Config{Env: tt.env}
			if got := cfg.IsDev(); got != tt.want {
				t.Errorf("IsDev() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoad_TurnstileConfig(t *testing.T) {
	// 元のワーキングディレクトリを保存
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(originalWd) }() // テスト終了後に元のディレクトリに戻す

	// テストの前にワーキングディレクトリをプロジェクトルート（go/）に変更
	if err := os.Chdir("../.."); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// 既存の環境変数を保存（テスト後に復元）
	originalSiteKey := os.Getenv("ANNICT_TURNSTILE_SITE_KEY")
	originalSecretKey := os.Getenv("ANNICT_TURNSTILE_SECRET_KEY")
	// ANNICT_TURNSTILE_DISABLE is also saved and restored. When the dev .env
	// sets this flag to true, the Turnstile keys are cleared in non-production,
	// which would break this test's key assertions, so it is explicitly unset
	// during this test.
	//
	// [Ja] ANNICT_TURNSTILE_DISABLE も保存・復元する。dev の実 .env でこのフラグが
	// true に設定されていると、非 prod で Turnstile キーが空に落とされ、本テストの
	// キー設定検証が壊れるため、本テスト中は明示的に未設定にする。
	originalDisableTurnstile := os.Getenv("ANNICT_TURNSTILE_DISABLE")
	defer func() {
		if originalSiteKey != "" {
			_ = os.Setenv("ANNICT_TURNSTILE_SITE_KEY", originalSiteKey)
		} else {
			_ = os.Unsetenv("ANNICT_TURNSTILE_SITE_KEY")
		}
		if originalSecretKey != "" {
			_ = os.Setenv("ANNICT_TURNSTILE_SECRET_KEY", originalSecretKey)
		} else {
			_ = os.Unsetenv("ANNICT_TURNSTILE_SECRET_KEY")
		}
		if originalDisableTurnstile != "" {
			_ = os.Setenv("ANNICT_TURNSTILE_DISABLE", originalDisableTurnstile)
		} else {
			_ = os.Unsetenv("ANNICT_TURNSTILE_DISABLE")
		}
	}()

	tests := []struct {
		name               string
		siteKey            string
		secretKey          string
		wantSiteKey        string
		wantSecretKey      string
		shouldSetSiteKey   bool
		shouldSetSecretKey bool
	}{
		{
			name:               "Turnstile環境変数が設定されている場合",
			siteKey:            "1x00000000000000000000AA",
			secretKey:          "1x0000000000000000000000000000000AA",
			wantSiteKey:        "1x00000000000000000000AA",
			wantSecretKey:      "1x0000000000000000000000000000000AA",
			shouldSetSiteKey:   true,
			shouldSetSecretKey: true,
		},
		{
			name:               "Turnstile環境変数が設定されていない場合（テスト環境のモック設定）",
			siteKey:            "",
			secretKey:          "",
			wantSiteKey:        "",
			wantSecretKey:      "",
			shouldSetSiteKey:   false,
			shouldSetSecretKey: false,
		},
		{
			name:               "Site Keyのみ設定されている場合",
			siteKey:            "1x00000000000000000000AA",
			secretKey:          "",
			wantSiteKey:        "1x00000000000000000000AA",
			wantSecretKey:      "",
			shouldSetSiteKey:   true,
			shouldSetSecretKey: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 必須の環境変数を設定
			_ = os.Setenv("APP_ENV", "test")
			_ = os.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test")
			_ = os.Setenv("ANNICT_PORT", "4004")
			_ = os.Setenv("ANNICT_DOMAIN", "test.example.com")
			_ = os.Setenv("ANNICT_COOKIE_DOMAIN", ".test.example.com")
			_ = os.Setenv("ANNICT_SESSION_SECURE", "false")
			_ = os.Setenv("ANNICT_SESSION_HTTPONLY", "true")
			_ = os.Setenv("ANNICT_IMGPROXY_ENDPOINT", "http://test:8080")
			_ = os.Setenv("ANNICT_IMGPROXY_KEY", "test-key")
			_ = os.Setenv("ANNICT_IMGPROXY_SALT", "test-salt")

			// Unset ANNICT_TURNSTILE_DISABLE so the key assertions are not
			// affected by the disable flag; this test only checks that the
			// site/secret keys are reflected as-is.
			//
			// [Ja] ANNICT_TURNSTILE_DISABLE を未設定にして、キー設定の検証が
			// 無効化フラグの影響を受けないようにする。本テストは site/secret
			// キーがそのまま反映されることだけを検証する。
			_ = os.Unsetenv("ANNICT_TURNSTILE_DISABLE")

			// Turnstile環境変数を設定
			if tt.shouldSetSiteKey {
				_ = os.Setenv("ANNICT_TURNSTILE_SITE_KEY", tt.siteKey)
			} else {
				_ = os.Unsetenv("ANNICT_TURNSTILE_SITE_KEY")
			}
			if tt.shouldSetSecretKey {
				_ = os.Setenv("ANNICT_TURNSTILE_SECRET_KEY", tt.secretKey)
			} else {
				_ = os.Unsetenv("ANNICT_TURNSTILE_SECRET_KEY")
			}

			// Config を読み込み
			cfg, err := Load()
			if err != nil {
				t.Fatalf("Load() failed: %v", err)
			}

			// Turnstile設定を検証
			if cfg.TurnstileSiteKey != tt.wantSiteKey {
				t.Errorf("TurnstileSiteKey = %q, want %q", cfg.TurnstileSiteKey, tt.wantSiteKey)
			}
			if cfg.TurnstileSecretKey != tt.wantSecretKey {
				t.Errorf("TurnstileSecretKey = %q, want %q", cfg.TurnstileSecretKey, tt.wantSecretKey)
			}
		})
	}
}

// TestLoad_DisableTurnstile verifies the ANNICT_TURNSTILE_DISABLE dev override.
// In non-production it clears both Turnstile keys (so the existing "empty key"
// path disables Turnstile); in production it is ignored (fail-closed) so bot
// protection can never be silently switched off by a stray env var.
//
// [Ja] ANNICT_TURNSTILE_DISABLE の dev 用オーバーライドを検証する。非本番では
// Turnstile の 2 キーを空に落とし (既存の「キー空」経路で Turnstile が無効化される)、
// 本番では無視する (fail-closed) ため、環境変数の漏れで Bot 対策が黙って無効化される
// ことはない。
func TestLoad_DisableTurnstile(t *testing.T) {
	// 元のワーキングディレクトリを保存
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(originalWd) }() // テスト終了後に元のディレクトリに戻す

	// テストの前にワーキングディレクトリをプロジェクトルート（go/）に変更
	if err := os.Chdir("../.."); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Restore ANNICT_TURNSTILE_DISABLE after the test. The unset subtest below
	// clears it without an automatic restore, and the dev .env sets it to true,
	// so restore it here to avoid leaking into later tests.
	//
	// [Ja] テスト後に ANNICT_TURNSTILE_DISABLE を復元する。下の未設定ケースは自動
	// 復元なしでこの変数を消し、dev の実 .env では true が入っているため、後続の
	// テストへ漏れないようここで復元する。
	originalDisableTurnstile := os.Getenv("ANNICT_TURNSTILE_DISABLE")
	defer func() {
		if originalDisableTurnstile != "" {
			_ = os.Setenv("ANNICT_TURNSTILE_DISABLE", originalDisableTurnstile)
		} else {
			_ = os.Unsetenv("ANNICT_TURNSTILE_DISABLE")
		}
	}()

	// Turnstile キーがセットされている前提を作るためのサンプル値
	const (
		sampleSiteKey   = "1x00000000000000000000AA"
		sampleSecretKey = "1x0000000000000000000000000000000AA"
	)

	tests := []struct {
		name          string
		env           string
		disableFlag   string // 空文字列は「未設定」を表す
		wantSiteKey   string
		wantSecretKey string
	}{
		{
			name:          "dev + フラグ true で両キーが空に落ちる",
			env:           "dev",
			disableFlag:   "true",
			wantSiteKey:   "",
			wantSecretKey: "",
		},
		{
			name:          "test + フラグ true で両キーが空に落ちる",
			env:           "test",
			disableFlag:   "true",
			wantSiteKey:   "",
			wantSecretKey: "",
		},
		{
			name:          "prod + フラグ true でもキーは維持される（fail-closed）",
			env:           "prod",
			disableFlag:   "true",
			wantSiteKey:   sampleSiteKey,
			wantSecretKey: sampleSecretKey,
		},
		{
			name:          "dev + フラグ未設定ならキーは維持される",
			env:           "dev",
			disableFlag:   "",
			wantSiteKey:   sampleSiteKey,
			wantSecretKey: sampleSecretKey,
		},
		{
			name:          "dev + フラグ false ならキーは維持される",
			env:           "dev",
			disableFlag:   "false",
			wantSiteKey:   sampleSiteKey,
			wantSecretKey: sampleSecretKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 必須の環境変数を設定
			t.Setenv("APP_ENV", tt.env)
			t.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test")
			t.Setenv("ANNICT_PORT", "4004")
			t.Setenv("ANNICT_DOMAIN", "test.example.com")
			t.Setenv("ANNICT_COOKIE_DOMAIN", ".test.example.com")
			t.Setenv("ANNICT_SESSION_SECURE", "false")
			t.Setenv("ANNICT_SESSION_HTTPONLY", "true")
			t.Setenv("ANNICT_IMGPROXY_ENDPOINT", "http://test:8080")
			t.Setenv("ANNICT_IMGPROXY_KEY", "test-key")
			t.Setenv("ANNICT_IMGPROXY_SALT", "test-salt")

			// Turnstile キーは常にセットしておき、無効化フラグの効果だけを見る
			t.Setenv("ANNICT_TURNSTILE_SITE_KEY", sampleSiteKey)
			t.Setenv("ANNICT_TURNSTILE_SECRET_KEY", sampleSecretKey)

			if tt.disableFlag != "" {
				t.Setenv("ANNICT_TURNSTILE_DISABLE", tt.disableFlag)
			} else {
				_ = os.Unsetenv("ANNICT_TURNSTILE_DISABLE")
			}

			cfg, err := Load()
			if err != nil {
				t.Fatalf("Load() failed: %v", err)
			}

			if cfg.TurnstileSiteKey != tt.wantSiteKey {
				t.Errorf("TurnstileSiteKey = %q, want %q", cfg.TurnstileSiteKey, tt.wantSiteKey)
			}
			if cfg.TurnstileSecretKey != tt.wantSecretKey {
				t.Errorf("TurnstileSecretKey = %q, want %q", cfg.TurnstileSecretKey, tt.wantSecretKey)
			}
		})
	}
}

func TestParseAdminIPs(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "単一IP",
			input: "192.168.1.1",
			want:  []string{"192.168.1.1"},
		},
		{
			name:  "複数IP",
			input: "192.168.1.1,10.0.0.1",
			want:  []string{"192.168.1.1", "10.0.0.1"},
		},
		{
			name:  "複数IPスペースあり",
			input: "192.168.1.1, 10.0.0.1, 172.16.0.1",
			want:  []string{"192.168.1.1", "10.0.0.1", "172.16.0.1"},
		},
		{
			name:  "空白のみの要素を除去",
			input: "192.168.1.1,  ,10.0.0.1",
			want:  []string{"192.168.1.1", "10.0.0.1"},
		},
		{
			name:  "空文字列",
			input: "",
			want:  []string{},
		},
		{
			name:  "先頭と末尾の空白を除去",
			input: "  192.168.1.1  ",
			want:  []string{"192.168.1.1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseAdminIPs(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseAdminIPs(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseSentryTracesSampleRate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{
			name:     "空文字列はデフォルト値0.5",
			input:    "",
			expected: 0.5,
		},
		{
			name:     "有効な値0.0",
			input:    "0.0",
			expected: 0.0,
		},
		{
			name:     "有効な値1.0",
			input:    "1.0",
			expected: 1.0,
		},
		{
			name:     "有効な値0.25",
			input:    "0.25",
			expected: 0.25,
		},
		{
			name:     "有効な値0.75",
			input:    "0.75",
			expected: 0.75,
		},
		{
			name:     "無効な値（文字列）はデフォルト値0.5",
			input:    "invalid",
			expected: 0.5,
		},
		{
			name:     "範囲外の値（負数）はデフォルト値0.5",
			input:    "-0.1",
			expected: 0.5,
		},
		{
			name:     "範囲外の値（1より大きい）はデフォルト値0.5",
			input:    "1.5",
			expected: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseSentryTracesSampleRate(tt.input)
			if result != tt.expected {
				t.Errorf("parseSentryTracesSampleRate(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLoad_MaintenanceMode(t *testing.T) {
	// 元のワーキングディレクトリを保存
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(originalWd) }()

	// テストの前にワーキングディレクトリをプロジェクトルート（go/）に変更
	if err := os.Chdir("../.."); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// 既存の環境変数を保存
	originalMaintenanceMode := os.Getenv("ANNICT_MAINTENANCE_MODE")
	originalAdminIP := os.Getenv("ANNICT_ADMIN_IP")
	defer func() {
		if originalMaintenanceMode != "" {
			_ = os.Setenv("ANNICT_MAINTENANCE_MODE", originalMaintenanceMode)
		} else {
			_ = os.Unsetenv("ANNICT_MAINTENANCE_MODE")
		}
		if originalAdminIP != "" {
			_ = os.Setenv("ANNICT_ADMIN_IP", originalAdminIP)
		} else {
			_ = os.Unsetenv("ANNICT_ADMIN_IP")
		}
	}()

	tests := []struct {
		name                string
		maintenanceMode     string
		adminIP             string
		wantMaintenanceMode bool
		wantAdminIPs        []string
	}{
		{
			name:                "メンテナンスモードON、単一IP",
			maintenanceMode:     "on",
			adminIP:             "192.168.1.1",
			wantMaintenanceMode: true,
			wantAdminIPs:        []string{"192.168.1.1"},
		},
		{
			name:                "メンテナンスモードON、複数IP",
			maintenanceMode:     "on",
			adminIP:             "192.168.1.1,10.0.0.1",
			wantMaintenanceMode: true,
			wantAdminIPs:        []string{"192.168.1.1", "10.0.0.1"},
		},
		{
			name:                "メンテナンスモードOFF",
			maintenanceMode:     "off",
			adminIP:             "192.168.1.1",
			wantMaintenanceMode: false,
			wantAdminIPs:        []string{"192.168.1.1"},
		},
		{
			name:                "メンテナンスモード未設定",
			maintenanceMode:     "",
			adminIP:             "",
			wantMaintenanceMode: false,
			wantAdminIPs:        nil,
		},
		{
			name:                "メンテナンスモードON、管理者IP未設定",
			maintenanceMode:     "on",
			adminIP:             "",
			wantMaintenanceMode: true,
			wantAdminIPs:        nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 必須の環境変数を設定
			_ = os.Setenv("APP_ENV", "test")
			_ = os.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test")
			_ = os.Setenv("ANNICT_PORT", "4004")
			_ = os.Setenv("ANNICT_DOMAIN", "test.example.com")
			_ = os.Setenv("ANNICT_COOKIE_DOMAIN", ".test.example.com")
			_ = os.Setenv("ANNICT_SESSION_SECURE", "false")
			_ = os.Setenv("ANNICT_SESSION_HTTPONLY", "true")
			_ = os.Setenv("ANNICT_IMGPROXY_ENDPOINT", "http://test:8080")
			_ = os.Setenv("ANNICT_IMGPROXY_KEY", "test-key")
			_ = os.Setenv("ANNICT_IMGPROXY_SALT", "test-salt")

			// メンテナンスモード環境変数を設定
			if tt.maintenanceMode != "" {
				_ = os.Setenv("ANNICT_MAINTENANCE_MODE", tt.maintenanceMode)
			} else {
				_ = os.Unsetenv("ANNICT_MAINTENANCE_MODE")
			}
			if tt.adminIP != "" {
				_ = os.Setenv("ANNICT_ADMIN_IP", tt.adminIP)
			} else {
				_ = os.Unsetenv("ANNICT_ADMIN_IP")
			}

			// Config を読み込み
			cfg, err := Load()
			if err != nil {
				t.Fatalf("Load() failed: %v", err)
			}

			// メンテナンスモード設定を検証
			if cfg.MaintenanceMode != tt.wantMaintenanceMode {
				t.Errorf("MaintenanceMode = %v, want %v", cfg.MaintenanceMode, tt.wantMaintenanceMode)
			}
			if !reflect.DeepEqual(cfg.AdminIPs, tt.wantAdminIPs) {
				t.Errorf("AdminIPs = %v, want %v", cfg.AdminIPs, tt.wantAdminIPs)
			}
		})
	}
}

// TestGetGitCommitHash verifies that the GIT_REV environment variable takes
// precedence and is shortened to 7 characters.
//
// A Dokku deploy target has no .git directory, so the git command fails;
// whether GIT_REV is usable therefore decides the Sentry release (avoiding a
// fallback to "dev").
//
// [Ja] GIT_REV 環境変数が最優先され、7 文字に短縮されることを検証する。
//
// Dokku のデプロイ先には .git が無く git コマンドが失敗するため、GIT_REV を
// 使えるかどうかが Sentry の release ("dev" 化の回避) を左右する。
func TestGetGitCommitHash(t *testing.T) {
	tests := []struct {
		name   string
		gitRev string
		want   string
	}{
		{
			name:   "フルSHAは7文字に短縮される",
			gitRev: "1234567890abcdef1234567890abcdef12345678",
			want:   "1234567",
		},
		{
			name:   "7文字以下ならそのまま返す",
			gitRev: "abc123",
			want:   "abc123",
		},
		{
			name:   "前後の空白は除去される",
			gitRev: "  1234567890abcdef  ",
			want:   "1234567",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("GIT_REV", tt.gitRev)

			got := getGitCommitHash()
			if got != tt.want {
				t.Errorf("getGitCommitHash() = %q, want %q", got, tt.want)
			}
		})
	}
}

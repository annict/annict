package manifest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/annict/annict/internal/config"
	"github.com/annict/annict/internal/i18n"
)

func TestShow(t *testing.T) {
	tests := []struct {
		name            string
		env             string
		wantName        string
		wantShortName   string
		wantStatusCode  int
		wantContentType string
	}{
		{
			name:            "開発環境でのmanifest生成",
			env:             "dev",
			wantName:        "Annict (Dev)",
			wantShortName:   "Annict (Dev)",
			wantStatusCode:  http.StatusOK,
			wantContentType: "application/manifest+json",
		},
		{
			name:            "本番環境でのmanifest生成",
			env:             "prod",
			wantName:        "Annict",
			wantShortName:   "Annict",
			wantStatusCode:  http.StatusOK,
			wantContentType: "application/manifest+json",
		},
		{
			name:            "テスト環境でのmanifest生成",
			env:             "test",
			wantName:        "Annict",
			wantShortName:   "Annict",
			wantStatusCode:  http.StatusOK,
			wantContentType: "application/manifest+json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト用の設定を作成
			cfg := &config.Config{
				Env: tt.env,
			}

			// ハンドラーを作成
			handler := NewHandler(cfg)

			// リクエストを作成
			req := httptest.NewRequest(http.MethodGet, "/manifest.json", nil)

			// i18n.T が動作するように locale をコンテキストに設定
			// GetLocalizer が自動的に localizer を作成してくれる
			ctx := context.Background()
			ctx = i18n.SetLocale(ctx, i18n.LangJa)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			// ハンドラーを実行
			handler.Show(rr, req)

			// ステータスコードを確認
			if rr.Code != tt.wantStatusCode {
				t.Errorf("ステータスコード = %v, want %v", rr.Code, tt.wantStatusCode)
			}

			// Content-Typeを確認
			contentType := rr.Header().Get("Content-Type")
			if contentType != tt.wantContentType {
				t.Errorf("Content-Type = %v, want %v", contentType, tt.wantContentType)
			}

			// JSONをパース
			var manifest Manifest
			if err := json.NewDecoder(rr.Body).Decode(&manifest); err != nil {
				t.Fatalf("JSONのパースに失敗しました: %v", err)
			}

			// name を確認
			if manifest.Name != tt.wantName {
				t.Errorf("Name = %v, want %v", manifest.Name, tt.wantName)
			}

			// short_name を確認
			if manifest.ShortName != tt.wantShortName {
				t.Errorf("ShortName = %v, want %v", manifest.ShortName, tt.wantShortName)
			}

			// その他のフィールドを確認
			if manifest.BackgroundColor != "#f85b73" {
				t.Errorf("BackgroundColor = %v, want #f85b73", manifest.BackgroundColor)
			}
			if manifest.ThemeColor != "#f85b73" {
				t.Errorf("ThemeColor = %v, want #f85b73", manifest.ThemeColor)
			}
			if manifest.Display != "standalone" {
				t.Errorf("Display = %v, want standalone", manifest.Display)
			}
			if manifest.Scope != "/" {
				t.Errorf("Scope = %v, want /", manifest.Scope)
			}
			if manifest.StartURL != "/" {
				t.Errorf("StartURL = %v, want /", manifest.StartURL)
			}

			// description を確認（空でないこと）
			if manifest.Description == "" {
				t.Error("Description が空です")
			}

			// icons を確認
			if len(manifest.Icons) != 2 {
				t.Errorf("Icons の数 = %v, want 2", len(manifest.Icons))
			}

			// 192x192 のアイコンを確認
			if len(manifest.Icons) > 0 {
				icon := manifest.Icons[0]
				if icon.Sizes != "192x192" {
					t.Errorf("Icons[0].Sizes = %v, want 192x192", icon.Sizes)
				}
				if icon.Src != "/static/images/icon-192.png" {
					t.Errorf("Icons[0].Src = %v, want /static/images/icon-192.png", icon.Src)
				}
				if icon.Type != "image/png" {
					t.Errorf("Icons[0].Type = %v, want image/png", icon.Type)
				}
				if icon.Purpose != "any maskable" {
					t.Errorf("Icons[0].Purpose = %v, want any maskable", icon.Purpose)
				}
			}

			// 512x512 のアイコンを確認
			if len(manifest.Icons) > 1 {
				icon := manifest.Icons[1]
				if icon.Sizes != "512x512" {
					t.Errorf("Icons[1].Sizes = %v, want 512x512", icon.Sizes)
				}
				if icon.Src != "/static/images/icon-512.png" {
					t.Errorf("Icons[1].Src = %v, want /static/images/icon-512.png", icon.Src)
				}
				if icon.Type != "image/png" {
					t.Errorf("Icons[1].Type = %v, want image/png", icon.Type)
				}
				if icon.Purpose != "any maskable" {
					t.Errorf("Icons[1].Purpose = %v, want any maskable", icon.Purpose)
				}
			}
		})
	}
}

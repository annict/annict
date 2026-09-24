package manifest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/i18n"
)

func TestShow(t *testing.T) {
	t.Parallel()

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

			// i18n.Tが動作するようにlocaleをコンテキストに設定
			// GetLocalizerが自動的にlocalizerを作成してくれる
			ctx := context.Background()
			ctx = i18n.SetLocale(ctx, i18n.LangJa)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			// ハンドラーを実行
			handler.Show(rr, req)

			// ステータスコードを確認
			if rr.Code != tt.wantStatusCode {
				t.Errorf("ステータスコード = %v、期待値 = %v", rr.Code, tt.wantStatusCode)
			}

			// Content-Typeを確認
			contentType := rr.Header().Get("Content-Type")
			if contentType != tt.wantContentType {
				t.Errorf("Content-Type = %v、期待値 = %v", contentType, tt.wantContentType)
			}

			// JSONをパース
			var manifest Manifest
			if err := json.NewDecoder(rr.Body).Decode(&manifest); err != nil {
				t.Fatalf("JSONのパースに失敗しました: %v", err)
			}

			// nameを確認
			if manifest.Name != tt.wantName {
				t.Errorf("Name = %v、期待値 = %v", manifest.Name, tt.wantName)
			}

			// short_nameを確認
			if manifest.ShortName != tt.wantShortName {
				t.Errorf("ShortName = %v、期待値 = %v", manifest.ShortName, tt.wantShortName)
			}

			// その他のフィールドを確認
			if manifest.BackgroundColor != "#f85b73" {
				t.Errorf("BackgroundColor = %v、期待値 = #f85b73", manifest.BackgroundColor)
			}
			if manifest.ThemeColor != "#f85b73" {
				t.Errorf("ThemeColor = %v、期待値 = #f85b73", manifest.ThemeColor)
			}
			if manifest.Display != "standalone" {
				t.Errorf("Display = %v、期待値 = standalone", manifest.Display)
			}
			if manifest.Scope != "/" {
				t.Errorf("Scope = %v、期待値 = /", manifest.Scope)
			}
			if manifest.StartURL != "/" {
				t.Errorf("StartURL = %v、期待値 = /", manifest.StartURL)
			}

			// descriptionを確認 (空でないこと)
			if manifest.Description == "" {
				t.Error("Descriptionが空です")
			}

			// iconsを確認
			if len(manifest.Icons) != 2 {
				t.Errorf("Iconsの数 = %v、期待値 = 2", len(manifest.Icons))
			}

			// 192x192のアイコンを確認
			if len(manifest.Icons) > 0 {
				icon := manifest.Icons[0]
				if icon.Sizes != "192x192" {
					t.Errorf("Icons[0].Sizes = %v、期待値 = 192x192", icon.Sizes)
				}
				if icon.Src != "/static/images/icon-192.png" {
					t.Errorf("Icons[0].Src = %v、期待値 = /static/images/icon-192.png", icon.Src)
				}
				if icon.Type != "image/png" {
					t.Errorf("Icons[0].Type = %v、期待値 = image/png", icon.Type)
				}
				if icon.Purpose != "any maskable" {
					t.Errorf("Icons[0].Purpose = %v、期待値 = any maskable", icon.Purpose)
				}
			}

			// 512x512のアイコンを確認
			if len(manifest.Icons) > 1 {
				icon := manifest.Icons[1]
				if icon.Sizes != "512x512" {
					t.Errorf("Icons[1].Sizes = %v、期待値 = 512x512", icon.Sizes)
				}
				if icon.Src != "/static/images/icon-512.png" {
					t.Errorf("Icons[1].Src = %v、期待値 = /static/images/icon-512.png", icon.Src)
				}
				if icon.Type != "image/png" {
					t.Errorf("Icons[1].Type = %v、期待値 = image/png", icon.Type)
				}
				if icon.Purpose != "any maskable" {
					t.Errorf("Icons[1].Purpose = %v、期待値 = any maskable", icon.Purpose)
				}
			}
		})
	}
}

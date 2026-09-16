// Package manifestはWeb App Manifestのハンドラーを提供します
package manifest

import (
	"github.com/annict/annict/go/internal/config"
)

// HandlerはWeb App Manifest関連のHTTPハンドラーです
type Handler struct {
	cfg *config.Config
}

// NewHandlerは新しいHandlerを作成します
func NewHandler(cfg *config.Config) *Handler {
	return &Handler{
		cfg: cfg,
	}
}

// ManifestIconはWeb App ManifestのIconを表します
type ManifestIcon struct {
	Purpose string `json:"purpose"`
	Sizes   string `json:"sizes"`
	Src     string `json:"src"`
	Type    string `json:"type"`
}

// ManifestはWeb App Manifestの構造を表します
type Manifest struct {
	BackgroundColor string         `json:"background_color"`
	Description     string         `json:"description"`
	Display         string         `json:"display"`
	Icons           []ManifestIcon `json:"icons"`
	Name            string         `json:"name"`
	Scope           string         `json:"scope"`
	ShortName       string         `json:"short_name"`
	StartURL        string         `json:"start_url"`
	ThemeColor      string         `json:"theme_color"`
}

package home

import (
	"net/http"
)

// Show GET / - ホームページ
// Rails版のトップページにリダイレクトします
func (h *Handler) Show(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, h.cfg.RailsAppURL, http.StatusFound)
}

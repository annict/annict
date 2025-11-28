package health

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

// Show GET /health - ヘルスチェック
func (h *Handler) Show(w http.ResponseWriter, r *http.Request) {
	// データベース接続確認
	var dbStatus string
	ctx := r.Context()
	_, err := h.workRepo.GetByID(ctx, 1)
	if err != nil && err != sql.ErrNoRows {
		dbStatus = fmt.Sprintf("unhealthy: %v", err)
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		dbStatus = "healthy"
	}

	health := map[string]any{
		"status":   "ok",
		"database": dbStatus,
		"env":      h.cfg.Env,
	}

	w.Header().Set("Content-Type", "application/json")
	if dbStatus != "healthy" {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	if err := json.NewEncoder(w).Encode(health); err != nil {
		slog.ErrorContext(ctx, "ヘルスチェックレスポンスのエンコードに失敗", "error", err)
	}
}

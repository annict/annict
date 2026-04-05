package health

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// Show GET /health - ヘルスチェック
func (h *Handler) Show(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	result, err := h.checkHealthUC.Execute(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "ヘルスチェックの実行に失敗", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		if encErr := json.NewEncoder(w).Encode(map[string]any{
			"status":   "ok",
			"database": "unhealthy: internal error",
			"env":      h.cfg.Env,
		}); encErr != nil {
			slog.ErrorContext(ctx, "ヘルスチェックレスポンスのエンコードに失敗", "error", encErr)
		}
		return
	}

	var dbStatus string
	if result.DBHealthy {
		dbStatus = "healthy"
	} else {
		dbStatus = result.DBError
	}

	health := map[string]any{
		"status":   "ok",
		"database": dbStatus,
		"env":      h.cfg.Env,
	}

	w.Header().Set("Content-Type", "application/json")
	if !result.DBHealthy {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	if err := json.NewEncoder(w).Encode(health); err != nil {
		slog.ErrorContext(ctx, "ヘルスチェックレスポンスのエンコードに失敗", "error", err)
	}
}

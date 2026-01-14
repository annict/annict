package sign_out

import (
	"log/slog"
	"net/http"
)

// Delete はログアウト処理を行います
// DBからセッションレコードを削除し、Cookieをクリアしてホームページにリダイレクトします
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// セッションを削除（DBレコード削除 + Cookie削除）
	if err := h.sessionMgr.DestroySession(ctx, w, r); err != nil {
		slog.ErrorContext(ctx, "セッションの削除に失敗しました", "error", err)
		// エラーが発生してもホームにリダイレクトする
	}

	// ホームページにリダイレクト
	http.Redirect(w, r, "/", http.StatusFound)
}

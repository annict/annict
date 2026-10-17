package middleware

import (
	"log/slog"
	"net/http"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/annict/annict/go/internal/clientip"
)

// AccessLogは1リクエストにつき1行、メソッド・ホスト・パス・ステータス・応答のバイト数・
// 処理時間・クライアントのIPアドレスをslogで出力する。
//
// URLはクエリ文字列を除いたパスだけを出力する。パスワードリセットのトークン・新規登録の
// 一時トークン・APIのアクセストークンなど、秘密の値がクエリ文字列で渡るため。
// IPアドレスは不正なアクセスやレート制限の調査に使うため出力する。
//
// Rails版へ中継するリクエストも記録するため、リバースプロキシミドルウェアより前に登録する。
func AccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
		start := time.Now()

		// ハンドラーがpanicしてもリクエストの記録を残すため、deferで出力する。
		defer func() {
			slog.InfoContext(r.Context(), "リクエストを処理しました",
				"method", r.Method,
				"host", r.Host,
				"path", r.URL.Path,
				"status", ww.Status(),
				"bytes", ww.BytesWritten(),
				"duration", time.Since(start),
				"ip_address", clientip.GetClientIP(r),
			)
		}()

		next.ServeHTTP(ww, r)
	})
}

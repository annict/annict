package middleware

import (
	"net/http"
	"runtime/debug"

	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/annict/annict/go/internal/apierror"
)

// Recovererはパニックから復帰して500を返すミドルウェア。
// APIのパスにはProblem Detailsを返し、それ以外はchiのRecovererに任せる
type Recoverer struct {
	problems *apierror.Writer
}

// NewRecovererは新しいRecovererを作成する
func NewRecoverer(problems *apierror.Writer) *Recoverer {
	return &Recoverer{problems: problems}
}

// MiddlewareはHTTPミドルウェアを返す。
// APIのパスでの復帰の処理 (ログ出力・http.ErrAbortHandlerの扱い) はchiのRecovererに揃える
func (m *Recoverer) Middleware(next http.Handler) http.Handler {
	htmlRecoverer := chimiddleware.Recoverer(next)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !IsAPIPath(r.URL.Path) {
			htmlRecoverer.ServeHTTP(w, r)
			return
		}

		defer func() {
			rvr := recover()
			if rvr == nil {
				return
			}
			if rvr == http.ErrAbortHandler {
				// クライアントへの応答を中断するためのパニックなので、復帰せずに投げ直す
				panic(rvr)
			}

			if logEntry := chimiddleware.GetLogEntry(r); logEntry != nil {
				logEntry.Panic(rvr, debug.Stack())
			} else {
				chimiddleware.PrintPrettyStack(rvr)
			}

			m.problems.InternalServerError(w, r)
		}()

		next.ServeHTTP(w, r)
	})
}

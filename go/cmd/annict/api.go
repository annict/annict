package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/annict/annict/go/internal/apierror"
	"github.com/annict/annict/go/internal/apigen"
	"github.com/annict/annict/go/internal/apipagination"
	"github.com/annict/annict/go/internal/middleware"
)

// allowCandidateMethodsは405の `Allow` に載せうるメソッド
var allowCandidateMethods = []string{
	http.MethodGet,
	http.MethodHead,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
	http.MethodOptions,
}

// newAPIRouterはREST API V2 (middleware.APIPathPrefix配下) のルーターを作成する。
// ルートの不一致・メソッドの不一致・生成コードでのリクエストの読み取りの失敗・ハンドラーのエラーを
// すべてProblem Detailsで返す
func newAPIRouter(
	server apigen.StrictServerInterface,
	problems *apierror.Writer,
	tokenAuth *middleware.APITokenAuth,
	featureFlag *middleware.APIFeatureFlag,
	requestValidator *middleware.APIRequestValidator,
) *chi.Mux {
	r := chi.NewRouter()

	r.NotFound(problems.NotFound)
	r.MethodNotAllowed(func(w http.ResponseWriter, req *http.Request) {
		problems.MethodNotAllowed(w, req, allowedMethods(r, req))
	})

	// フィーチャーフラグは照合で得た呼び出し主体の利用者で判定するため、トークンの照合の後に掛ける。
	// フラグが無効な間はV2の存在を明かさないよう、securityの検証 (401・403) より前で404にする
	r.Use(tokenAuth.Middleware)
	r.Use(featureFlag.Middleware)
	r.Use(requestValidator.Middleware)

	// リクエストはミドルウェアでスキーマに照らして検証済みのため、ここでの読み取りの失敗は
	// 検証をすり抜けた形式の問題として400で返す
	readErrorHandler := func(w http.ResponseWriter, req *http.Request, err error) {
		problems.ValidationFailed(w, req, []apierror.FieldError{{Detail: err.Error()}})
	}
	strictHandler := apigen.NewStrictHandlerWithOptions(server, []apigen.StrictMiddlewareFunc{withRequestURL}, apigen.StrictHTTPServerOptions{
		RequestErrorHandlerFunc:  readErrorHandler,
		ResponseErrorHandlerFunc: problems.WriteError,
	})
	apigen.HandlerWithOptions(strictHandler, apigen.ChiServerOptions{
		BaseRouter:       r,
		ErrorHandlerFunc: readErrorHandler,
	})

	return r
}

// withRequestURLは、一覧のハンドラーが次のページの `Link` ヘッダーを組み立てられるよう、
// リクエストのURLをcontextに入れる。strict serverのハンドラーはリクエストを受け取らないため、ここで渡す
func withRequestURL(f apigen.StrictHandlerFunc, _ string) apigen.StrictHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (any, error) {
		return f(apipagination.WithRequestURL(ctx, r.URL), w, r, request)
	}
}

// allowedMethodsはリクエストのパスでルーターが受け付けるメソッドを返す。
// chiは独自の405ハンドラーに受け付けるメソッドを渡さないため、ルーターに問い合わせて求める
func allowedMethods(router *chi.Mux, r *http.Request) []string {
	// chiと同じく、エスケープされたパスがあればそれでルートを探す
	path := r.URL.RawPath
	if path == "" {
		path = r.URL.Path
	}
	path = strings.TrimPrefix(path, middleware.APIPathPrefix)

	var methods []string
	for _, method := range allowCandidateMethods {
		if router.Match(chi.NewRouteContext(), method, path) {
			methods = append(methods, method)
		}
	}
	return methods
}

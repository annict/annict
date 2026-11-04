package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/annict/annict/go/internal/apierror"
	"github.com/annict/annict/go/internal/model"
)

const (
	// apiPrincipalContextKeyはコンテキストにREST API V2の呼び出し主体を保存するためのキー
	apiPrincipalContextKey contextKey = "api_principal"
	// apiTokenRejectedContextKeyは、送られてきたトークンを受け付けられなかったことをコンテキストに残すためのキー
	apiTokenRejectedContextKey contextKey = "api_token_rejected"
)

// apiPrincipalGetterは、トークンの文字列を照合して呼び出し主体を返すインターフェース。
// usecase.GetAPIPrincipalUsecaseがこのインターフェースを満たす。
// トークンを受け付けられない場合は (nil, nil) を返す
type apiPrincipalGetter interface {
	Execute(ctx context.Context, accessToken string) (*model.APIPrincipal, error)
}

// APITokenAuthは `Authorization: Bearer` のトークンを照合し、呼び出し主体をコンテキストに入れるミドルウェア。
// トークンの無いリクエストと、受け付けられないトークンのリクエストもそのまま通す。
// フィーチャーフラグが有効になるまでV2の存在を隠すため、ここでは401を返さず、
// フラグの判定 (APIFeatureFlag) とoperationのsecurityの検証 (APIRequestValidator) に任せる
type APITokenAuth struct {
	getter   apiPrincipalGetter
	problems *apierror.Writer
}

// NewAPITokenAuthは新しいAPITokenAuthを作成する
func NewAPITokenAuth(getter apiPrincipalGetter, problems *apierror.Writer) *APITokenAuth {
	return &APITokenAuth{
		getter:   getter,
		problems: problems,
	}
}

// MiddlewareはHTTPミドルウェアを返す
func (a *APITokenAuth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			next.ServeHTTP(w, r)
			return
		}

		ctx := r.Context()
		token, ok := bearerToken(header)
		if !ok {
			next.ServeHTTP(w, r.WithContext(context.WithValue(ctx, apiTokenRejectedContextKey, true)))
			return
		}

		principal, err := a.getter.Execute(ctx, token)
		if err != nil {
			slog.ErrorContext(ctx, "APIのトークンの照合に失敗しました", "error", err)
			a.problems.InternalServerError(w, r)
			return
		}
		if principal == nil {
			next.ServeHTTP(w, r.WithContext(context.WithValue(ctx, apiTokenRejectedContextKey, true)))
			return
		}

		next.ServeHTTP(w, r.WithContext(SetAPIPrincipalToContext(ctx, principal)))
	})
}

// bearerTokenは `Authorization` ヘッダーの値からBearerトークンを取り出す。
// 認証スキームの名前は大文字小文字を区別しない (RFC 9110 §11.1)
func bearerToken(header string) (string, bool) {
	scheme, token, ok := strings.Cut(header, " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") {
		return "", false
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return "", false
	}
	return token, true
}

// APIPrincipalFromContextはコンテキストからREST API V2の呼び出し主体を取得する。
// トークンの無いリクエストと、受け付けられないトークンのリクエストではnilを返す
func APIPrincipalFromContext(ctx context.Context) *model.APIPrincipal {
	principal, ok := ctx.Value(apiPrincipalContextKey).(*model.APIPrincipal)
	if !ok {
		return nil
	}
	return principal
}

// SetAPIPrincipalToContextはコンテキストにREST API V2の呼び出し主体を設定する。
// テストでトークン認証済みの状態を再現するためにも使う
func SetAPIPrincipalToContext(ctx context.Context, principal *model.APIPrincipal) context.Context {
	return context.WithValue(ctx, apiPrincipalContextKey, principal)
}

// apiTokenRejectedFromContextは、送られてきたトークンを受け付けられなかったかを返す
func apiTokenRejectedFromContext(ctx context.Context) bool {
	rejected, _ := ctx.Value(apiTokenRejectedContextKey).(bool)
	return rejected
}

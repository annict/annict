package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/annict/annict/go/internal/apierror"
)

// APIFeatureFlagは、フィーチャーフラグ `rest_api_v2` が有効な利用者のトークンのリクエストにだけ
// REST API V2を応答し、それ以外には存在しないものとして404を返すミドルウェア。
// トークンの無いリクエストと受け付けられないトークンのリクエストも404にし、フラグが無効な間はV2の存在を明かさない。
// securityが空のoperation (`getOpenAPIDescription`) のパスは、トークンもフラグも要らないため通す。
// 呼び出し主体をコンテキストから読むため、APITokenAuthの後に掛ける
type APIFeatureFlag struct {
	problems    *apierror.Writer
	publicPaths map[string]bool
}

// NewAPIFeatureFlagは新しいAPIFeatureFlagを作成する。
// specにはOpenAPI記述 (`apigen.GetSpec()`) を渡し、トークンもフラグも要らないパスをそこから求める
func NewAPIFeatureFlag(spec *openapi3.T, problems *apierror.Writer) (*APIFeatureFlag, error) {
	publicPaths := make(map[string]bool)
	for path, item := range spec.Paths.Map() {
		if !isPublicPathItem(item) {
			continue
		}
		// パスパラメーターを含むパスはリクエストのパスと文字列で照合できないため、扱わない
		if strings.Contains(path, "{") {
			return nil, fmt.Errorf("securityが空のoperationのパスにパスパラメーターがあります: %s", path)
		}
		publicPaths[APIPathPrefix+path] = true
	}

	return &APIFeatureFlag{
		problems:    problems,
		publicPaths: publicPaths,
	}, nil
}

// isPublicPathItemは、パスのすべてのoperationのsecurityが空 (`security: []`) かを返す
func isPublicPathItem(item *openapi3.PathItem) bool {
	operations := item.Operations()
	if len(operations) == 0 {
		return false
	}
	for _, op := range operations {
		if op.Security == nil || len(*op.Security) != 0 {
			return false
		}
	}
	return true
}

// MiddlewareはHTTPミドルウェアを返す
func (f *APIFeatureFlag) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if f.publicPaths[r.URL.Path] {
			next.ServeHTTP(w, r)
			return
		}

		principal := APIPrincipalFromContext(r.Context())
		if principal == nil || !principal.RESTAPIV2Enabled {
			f.problems.NotFound(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

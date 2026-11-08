package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/annict/annict/go/internal/apierror"
	"github.com/annict/annict/go/internal/model"
)

// testPrincipalGetterは、トークンの文字列に対応する呼び出し主体を返す。errを設定すると照合に失敗する
type testPrincipalGetter struct {
	principals map[string]*model.APIPrincipal
	err        error
}

func (g testPrincipalGetter) Execute(_ context.Context, accessToken string) (*model.APIPrincipal, error) {
	if g.err != nil {
		return nil, g.err
	}
	return g.principals[accessToken], nil
}

func TestAPITokenAuth(t *testing.T) {
	t.Parallel()

	principal := &model.APIPrincipal{UserID: 1, TokenKind: model.APITokenKindPersonalAccessToken, Scopes: []string{"read"}}
	getter := testPrincipalGetter{principals: map[string]*model.APIPrincipal{"valid-token": principal}}

	tests := []struct {
		name          string
		authorization string
		wantPrincipal bool
		wantRejected  bool
	}{
		{name: "トークンの無いリクエスト"},
		{name: "照合できたトークン", authorization: "Bearer valid-token", wantPrincipal: true},
		{name: "スキームの大文字小文字を区別しない", authorization: "bearer valid-token", wantPrincipal: true},
		{name: "照合できないトークン", authorization: "Bearer unknown-token", wantRejected: true},
		{name: "Bearerでない認証", authorization: "Basic dXNlcjpwYXNz", wantRejected: true},
		{name: "空のBearerトークン", authorization: "Bearer ", wantRejected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var gotPrincipal *model.APIPrincipal
			var gotRejected bool
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPrincipal = APIPrincipalFromContext(r.Context())
				gotRejected = apiTokenRejectedFromContext(r.Context())
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/api/v2/anime/1", nil)
			if tt.authorization != "" {
				req.Header.Set("Authorization", tt.authorization)
			}
			rr := httptest.NewRecorder()
			NewAPITokenAuth(getter, apierror.NewWriter("https://example.com")).Middleware(next).ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("ステータスコード = %d、期待値 = 次のハンドラーの応答", rr.Code)
			}
			if (gotPrincipal != nil) != tt.wantPrincipal {
				t.Errorf("呼び出し主体 = %+v、あるべきか = %v", gotPrincipal, tt.wantPrincipal)
			}
			if gotRejected != tt.wantRejected {
				t.Errorf("受け付けられなかったか = %v、期待値 = %v", gotRejected, tt.wantRejected)
			}
		})
	}
}

func TestAPITokenAuth_GetterError(t *testing.T) {
	t.Parallel()

	getter := testPrincipalGetter{err: errors.New("DBに接続できません")}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("照合に失敗したリクエストで次のハンドラーが呼ばれた")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v2/anime/1", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()
	NewAPITokenAuth(getter, apierror.NewWriter("https://example.com")).Middleware(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusInternalServerError)
	}
}

// featureFlagTestSpecは、フィーチャーフラグの判定のテスト用のOpenAPI記述。
// %sに、securityが空のoperationのパスを入れる
const featureFlagTestSpec = `
openapi: 3.1.1
info:
  title: test
  version: 1.0.0
servers:
  - url: /api/v2
security:
  - personalAccessToken: [read]
paths:
  %s:
    get:
      operationId: getPublic
      security: []
      responses:
        "200":
          description: ok
  /secret:
    get:
      operationId: getSecret
      responses:
        "200":
          description: ok
components:
  securitySchemes:
    personalAccessToken:
      type: http
      scheme: bearer
`

func loadFeatureFlagTestSpec(t *testing.T, publicPath string) *openapi3.T {
	t.Helper()

	spec, err := openapi3.NewLoader().LoadFromData([]byte(fmt.Sprintf(featureFlagTestSpec, publicPath)))
	if err != nil {
		t.Fatalf("テスト用のOpenAPI記述を読み込めない: %v", err)
	}
	return spec
}

func TestAPIFeatureFlag(t *testing.T) {
	t.Parallel()

	f, err := NewAPIFeatureFlag(loadFeatureFlagTestSpec(t, "/public"), apierror.NewWriter("https://example.com"))
	if err != nil {
		t.Fatalf("NewAPIFeatureFlag()のエラー = %v", err)
	}

	tests := []struct {
		name       string
		path       string
		principal  *model.APIPrincipal
		wantStatus int
	}{
		{name: "securityが空のoperationのパスはトークンが無くても通す", path: "/api/v2/public", wantStatus: http.StatusOK},
		{name: "トークンが無ければ404", path: "/api/v2/secret", wantStatus: http.StatusNotFound},
		{name: "フラグの無い利用者は404", path: "/api/v2/secret", principal: &model.APIPrincipal{UserID: 1}, wantStatus: http.StatusNotFound},
		{name: "フラグの有効な利用者は通す", path: "/api/v2/secret", principal: &model.APIPrincipal{UserID: 1, RESTAPIV2Enabled: true}, wantStatus: http.StatusOK},
		{name: "記述に無いパスもフラグで判定する", path: "/api/v2/unknown", wantStatus: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.principal != nil {
				req = req.WithContext(SetAPIPrincipalToContext(req.Context(), tt.principal))
			}
			rr := httptest.NewRecorder()
			f.Middleware(testHandler()).ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, tt.wantStatus)
			}
		})
	}
}

func TestNewAPIFeatureFlag_PublicPathWithParameter(t *testing.T) {
	t.Parallel()

	// パスパラメーターを含むパスはリクエストのパスと文字列で照合できないため、作成に失敗する
	if _, err := NewAPIFeatureFlag(loadFeatureFlagTestSpec(t, "/public/{id}"), apierror.NewWriter("https://example.com")); err == nil {
		t.Error("エラーを期待したが、nilだった")
	}
}

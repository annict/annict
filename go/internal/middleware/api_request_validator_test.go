package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/annict/annict/go/internal/apierror"
	"github.com/annict/annict/go/internal/model"
)

// validatorTestSpecはリクエスト検証のテスト用のOpenAPI記述。
// パラメーターとsecurityの検証を、実際のoperationが揃う前に確かめるために使う
const validatorTestSpec = `
openapi: 3.1.1
info:
  title: test
  version: 1.0.0
servers:
  - url: /api/v2
security:
  - oauth2: [read]
  - personalAccessToken: [read]
paths:
  /things:
    get:
      operationId: listThings
      security: []
      parameters:
        - name: limit
          in: query
          schema:
            type: integer
            minimum: 1
            maximum: 100
        - name: sort
          in: query
          schema:
            type: string
            enum: [id, -id]
      responses:
        "200":
          description: ok
  /things/{thing_id}:
    get:
      operationId: getThing
      security: []
      parameters:
        - name: thing_id
          in: path
          required: true
          schema:
            type: integer
      responses:
        "200":
          description: ok
  /secret:
    get:
      operationId: getSecret
      responses:
        "200":
          description: ok
  /writer:
    get:
      operationId: getWriter
      security:
        - oauth2: [write]
        - personalAccessToken: [write]
      responses:
        "200":
          description: ok
components:
  securitySchemes:
    oauth2:
      type: oauth2
      flows:
        authorizationCode:
          authorizationUrl: /oauth/authorize
          tokenUrl: /oauth/token
          scopes:
            read: 読む
            write: 書き込む
    personalAccessToken:
      type: http
      scheme: bearer
`

func newTestAPIRequestValidator(t *testing.T) *APIRequestValidator {
	t.Helper()

	spec, err := openapi3.NewLoader().LoadFromData([]byte(validatorTestSpec))
	if err != nil {
		t.Fatalf("テスト用のOpenAPI記述を読み込めない: %v", err)
	}
	v, err := NewAPIRequestValidator(spec, apierror.NewWriter("https://example.com"))
	if err != nil {
		t.Fatalf("NewAPIRequestValidator()のエラー = %v", err)
	}
	return v
}

type validatorProblem struct {
	Type   string `json:"type"`
	Status int    `json:"status"`
	Errors []struct {
		Detail    string  `json:"detail"`
		Pointer   *string `json:"pointer"`
		Parameter *string `json:"parameter"`
	} `json:"errors"`
}

func TestAPIRequestValidator_PassesValidRequests(t *testing.T) {
	t.Parallel()

	v := newTestAPIRequestValidator(t)

	tests := []struct {
		name   string
		method string
		target string
	}{
		{name: "パラメーターの無いGET", method: http.MethodGet, target: "/api/v2/things"},
		{name: "正しいクエリパラメーター", method: http.MethodGet, target: "/api/v2/things?limit=100&sort=-id"},
		{name: "正しいパスパラメーター", method: http.MethodGet, target: "/api/v2/things/1"},
		{name: "記述に無いパスはルーターに任せる", method: http.MethodGet, target: "/api/v2/unknown"},
		{name: "記述に無いメソッドはルーターに任せる", method: http.MethodDelete, target: "/api/v2/things"},
		{name: "絶対形式のリクエストもパスで照合する", method: http.MethodGet, target: "https://example.com/api/v2/things?limit=1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			called := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			})

			rr := httptest.NewRecorder()
			v.Middleware(next).ServeHTTP(rr, httptest.NewRequest(tt.method, tt.target, nil))

			if !called {
				t.Fatalf("次のハンドラーが呼ばれない: ステータス = %d、本文 = %s", rr.Code, rr.Body.String())
			}
		})
	}
}

func TestAPIRequestValidator_RejectsInvalidParameters(t *testing.T) {
	t.Parallel()

	v := newTestAPIRequestValidator(t)

	tests := []struct {
		name           string
		target         string
		wantParameters []string
	}{
		{name: "範囲外のクエリパラメーター", target: "/api/v2/things?limit=0", wantParameters: []string{"limit"}},
		{name: "型の違うクエリパラメーター", target: "/api/v2/things?limit=abc", wantParameters: []string{"limit"}},
		{name: "列挙値に無いクエリパラメーター", target: "/api/v2/things?sort=title", wantParameters: []string{"sort"}},
		{name: "複数の問題はまとめて返す", target: "/api/v2/things?limit=0&sort=title", wantParameters: []string{"limit", "sort"}},
		{name: "型の違うパスパラメーター", target: "/api/v2/things/abc", wantParameters: []string{"thing_id"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Error("検証に失敗したリクエストで次のハンドラーが呼ばれた")
			})

			rr := httptest.NewRecorder()
			v.Middleware(next).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, tt.target, nil))

			if rr.Code != http.StatusBadRequest {
				t.Fatalf("ステータスコード = %d、期待値 = %d (本文: %s)", rr.Code, http.StatusBadRequest, rr.Body.String())
			}
			var problem validatorProblem
			if err := json.Unmarshal(rr.Body.Bytes(), &problem); err != nil {
				t.Fatalf("本文をJSONとして読めない: %v", err)
			}
			if want := "https://example.com/api/problems/validation-failed"; problem.Type != want {
				t.Errorf("type = %q、期待値 = %q", problem.Type, want)
			}

			got := make(map[string]bool, len(problem.Errors))
			for _, fe := range problem.Errors {
				if fe.Parameter == nil {
					t.Errorf("parameterの無い問題 = %+v", fe)
					continue
				}
				if fe.Pointer != nil {
					t.Errorf("パラメーターの問題にpointerがある: %q", *fe.Pointer)
				}
				if fe.Detail == "" {
					t.Errorf("%sの問題の説明が空", *fe.Parameter)
				}
				got[*fe.Parameter] = true
			}
			if len(got) != len(tt.wantParameters) {
				t.Errorf("問題のあったパラメーター = %v、期待値 = %v", got, tt.wantParameters)
			}
			for _, p := range tt.wantParameters {
				if !got[p] {
					t.Errorf("%sの問題が無い (問題: %v)", p, got)
				}
			}
		})
	}
}

func TestAPIRequestValidator_RejectsRequestsWithoutToken(t *testing.T) {
	t.Parallel()

	v := newTestAPIRequestValidator(t)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("securityを満たさないリクエストで次のハンドラーが呼ばれた")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v2/secret", nil)
	req.Header.Set("Authorization", "Bearer token")
	rr := httptest.NewRecorder()
	v.Middleware(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("ステータスコード = %d、期待値 = %d (本文: %s)", rr.Code, http.StatusUnauthorized, rr.Body.String())
	}
	if got := rr.Header().Get("WWW-Authenticate"); got != "Bearer" {
		t.Errorf("WWW-Authenticate = %q、期待値 = %q", got, "Bearer")
	}
}

func TestAPIRequestValidator_ChecksTokenScopes(t *testing.T) {
	t.Parallel()

	v := newTestAPIRequestValidator(t)
	personalRead := &model.APIPrincipal{UserID: 1, TokenKind: model.APITokenKindPersonalAccessToken, Scopes: []string{"read"}}
	oauthReadWrite := &model.APIPrincipal{UserID: 1, TokenKind: model.APITokenKindOAuthAccessToken, Scopes: []string{"read", "write"}}
	noScope := &model.APIPrincipal{UserID: 1, TokenKind: model.APITokenKindPersonalAccessToken}

	tests := []struct {
		name          string
		path          string
		principal     *model.APIPrincipal
		rejected      bool
		wantStatus    int
		wantChallenge string
	}{
		{name: "readの個人用アクセストークン", path: "/api/v2/secret", principal: personalRead, wantStatus: http.StatusOK},
		{name: "read・writeのOAuthのトークン", path: "/api/v2/writer", principal: oauthReadWrite, wantStatus: http.StatusOK},
		{name: "スコープの足りないトークンは403", path: "/api/v2/writer", principal: personalRead, wantStatus: http.StatusForbidden, wantChallenge: `Bearer error="insufficient_scope", scope="write"`},
		{name: "スコープの無いトークンは403", path: "/api/v2/secret", principal: noScope, wantStatus: http.StatusForbidden, wantChallenge: `Bearer error="insufficient_scope", scope="read"`},
		{name: "受け付けられなかったトークンは401にinvalid_tokenを付ける", path: "/api/v2/secret", rejected: true, wantStatus: http.StatusUnauthorized, wantChallenge: `Bearer error="invalid_token"`},
		{name: "トークンの無いリクエストは401にエラーコードを付けない", path: "/api/v2/secret", wantStatus: http.StatusUnauthorized, wantChallenge: "Bearer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			ctx := req.Context()
			if tt.principal != nil {
				ctx = SetAPIPrincipalToContext(ctx, tt.principal)
			}
			if tt.rejected {
				ctx = context.WithValue(ctx, apiTokenRejectedContextKey, true)
			}
			rr := httptest.NewRecorder()
			v.Middleware(testHandler()).ServeHTTP(rr, req.WithContext(ctx))

			if rr.Code != tt.wantStatus {
				t.Fatalf("ステータスコード = %d、期待値 = %d (本文: %s)", rr.Code, tt.wantStatus, rr.Body.String())
			}
			if got := rr.Header().Get("WWW-Authenticate"); got != tt.wantChallenge {
				t.Errorf("WWW-Authenticate = %q、期待値 = %q", got, tt.wantChallenge)
			}
		})
	}
}

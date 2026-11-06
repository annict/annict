package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/annict/annict/go/internal/middleware"
)

// newMethodOverrideRouterは、MethodOverrideを掛けたルーターを返す。
// 各ルートは、届いたルートのメソッドをレスポンスの本文に書く。
// mountedがtrueなら、本番のWeb画面と同じく外側のルーターにMountで載せる。
func newMethodOverrideRouter(mounted bool) http.Handler {
	sub := chi.NewRouter()
	sub.Use(middleware.MethodOverride)

	respond := func(name string) http.HandlerFunc {
		return func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(name))
		}
	}
	sub.Get("/resource", respond(http.MethodGet))
	sub.Post("/resource", respond(http.MethodPost))
	sub.Put("/resource", respond(http.MethodPut))
	sub.Patch("/resource", respond(http.MethodPatch))
	sub.Delete("/resource", respond(http.MethodDelete))
	// 同じパスにPOSTとDELETEのルートがある形 (二要素認証の有効化と無効化)
	sub.Post("/post_and_delete", respond(http.MethodPost))
	sub.Delete("/post_and_delete", respond(http.MethodDelete))

	if !mounted {
		return sub
	}

	root := chi.NewRouter()
	root.Mount("/", sub)
	return root
}

func TestMethodOverride(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		method    string
		path      string
		form      url.Values
		wantRoute string
	}{
		{
			name:      "_method=PUTのPOSTはPUTのルートに届く",
			method:    http.MethodPost,
			path:      "/resource",
			form:      url.Values{"_method": {"PUT"}},
			wantRoute: http.MethodPut,
		},
		{
			name:      "_method=PATCHのPOSTはPATCHのルートに届く",
			method:    http.MethodPost,
			path:      "/resource",
			form:      url.Values{"_method": {"PATCH"}},
			wantRoute: http.MethodPatch,
		},
		{
			name:      "_method=DELETEのPOSTはDELETEのルートに届く",
			method:    http.MethodPost,
			path:      "/resource",
			form:      url.Values{"_method": {"DELETE"}},
			wantRoute: http.MethodDelete,
		},
		{
			name:      "小文字の_methodも上書きする",
			method:    http.MethodPost,
			path:      "/resource",
			form:      url.Values{"_method": {"delete"}},
			wantRoute: http.MethodDelete,
		},
		{
			name:      "同じパスにPOSTのルートがあっても上書きしたメソッドのルートに届く",
			method:    http.MethodPost,
			path:      "/post_and_delete",
			form:      url.Values{"_method": {"DELETE"}},
			wantRoute: http.MethodDelete,
		},
		{
			name:      "_methodの無いPOSTはPOSTのルートに届く",
			method:    http.MethodPost,
			path:      "/post_and_delete",
			form:      url.Values{"name": {"value"}},
			wantRoute: http.MethodPost,
		},
		{
			name:      "POST以外の要求に付けた_methodは無視する",
			method:    http.MethodGet,
			path:      "/resource?_method=DELETE",
			wantRoute: http.MethodGet,
		},
		{
			name:      "上書き先がPUT・PATCH・DELETE以外の_methodは無視する",
			method:    http.MethodPost,
			path:      "/resource",
			form:      url.Values{"_method": {"GET"}},
			wantRoute: http.MethodPost,
		},
	}

	routers := []struct {
		name    string
		mounted bool
	}{
		{name: "Mountを挟む構成", mounted: true},
		{name: "Mountを挟まない構成", mounted: false},
	}

	for _, router := range routers {
		for _, tt := range tests {
			t.Run(router.name+"/"+tt.name, func(t *testing.T) {
				t.Parallel()

				handler := newMethodOverrideRouter(router.mounted)

				var req *http.Request
				if tt.form != nil {
					req = httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.form.Encode()))
					req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				} else {
					req = httptest.NewRequest(tt.method, tt.path, nil)
				}
				rr := httptest.NewRecorder()

				handler.ServeHTTP(rr, req)

				if rr.Code != http.StatusOK {
					t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusOK)
				}
				if got := rr.Body.String(); got != tt.wantRoute {
					t.Errorf("届いたルートのメソッド = %s、期待値 = %s", got, tt.wantRoute)
				}
			})
		}
	}
}

func TestMethodOverride_WithoutRouteContext(t *testing.T) {
	t.Parallel()

	for _, method := range []string{http.MethodPut, http.MethodPatch, http.MethodDelete} {
		t.Run("chiのルーターを介さず_method="+method+"を上書きする", func(t *testing.T) {
			t.Parallel()

			var gotMethod string
			handler := middleware.MethodOverride(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				gotMethod = r.Method
			}))
			form := url.Values{"_method": {method}}
			req := httptest.NewRequest(http.MethodPost, "/resource", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if gotMethod != method {
				t.Errorf("後続ハンドラーのメソッド = %q、期待値 = %q", gotMethod, method)
			}
		})
	}
}

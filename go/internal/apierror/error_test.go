package apierror_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/apierror"
	"github.com/annict/annict/go/internal/model"
)

// problemResponseはレスポンス本文を、キーの有無まで確かめられるように読み込んだもの
type problemResponse map[string]any

func decodeProblem(t *testing.T, rr *httptest.ResponseRecorder) problemResponse {
	t.Helper()

	if got := rr.Header().Get("Content-Type"); got != apierror.ContentType {
		t.Errorf("Content-Type = %q、期待値 = %q", got, apierror.ContentType)
	}
	var body problemResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("本文をJSONとして読めない: %v (本文: %s)", err, rr.Body.String())
	}
	return body
}

func TestWriter_StatusProblems(t *testing.T) {
	t.Parallel()

	pw := apierror.NewWriter("https://example.com")

	tests := []struct {
		name       string
		write      func(w http.ResponseWriter, r *http.Request)
		wantStatus int
		wantHeader map[string]string
	}{
		{
			name: "トークンの無いリクエストの401",
			write: func(w http.ResponseWriter, r *http.Request) {
				pw.Unauthorized(w, r, apierror.BearerChallenge{})
			},
			wantStatus: http.StatusUnauthorized,
			wantHeader: map[string]string{"WWW-Authenticate": "Bearer"},
		},
		{
			name: "無効なトークンの401",
			write: func(w http.ResponseWriter, r *http.Request) {
				pw.Unauthorized(w, r, apierror.BearerChallenge{Error: "invalid_token"})
			},
			wantStatus: http.StatusUnauthorized,
			wantHeader: map[string]string{"WWW-Authenticate": `Bearer error="invalid_token"`},
		},
		{
			name: "スコープ不足の403",
			write: func(w http.ResponseWriter, r *http.Request) {
				pw.Forbidden(w, r, apierror.BearerChallenge{Error: "insufficient_scope", Scope: []string{"read", "write"}})
			},
			wantStatus: http.StatusForbidden,
			wantHeader: map[string]string{"WWW-Authenticate": `Bearer error="insufficient_scope", scope="read write"`},
		},
		{
			name:       "404",
			write:      pw.NotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name: "405は使えるメソッドをAllowで示す",
			write: func(w http.ResponseWriter, r *http.Request) {
				pw.MethodNotAllowed(w, r, []string{http.MethodGet, http.MethodHead})
			},
			wantStatus: http.StatusMethodNotAllowed,
			wantHeader: map[string]string{"Allow": "GET, HEAD"},
		},
		{
			name:       "413",
			write:      pw.ContentTooLarge,
			wantStatus: http.StatusRequestEntityTooLarge,
		},
		{
			name: "429は1秒未満の待ち時間を1秒に切り上げる",
			write: func(w http.ResponseWriter, r *http.Request) {
				pw.TooManyRequests(w, r, 300*time.Millisecond)
			},
			wantStatus: http.StatusTooManyRequests,
			wantHeader: map[string]string{"Retry-After": "1"},
		},
		{
			name:       "500",
			write:      pw.InternalServerError,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "503は再開の見込みまでの秒数をRetry-Afterで示す",
			write: func(w http.ResponseWriter, r *http.Request) {
				pw.ServiceUnavailable(w, r, time.Hour)
			},
			wantStatus: http.StatusServiceUnavailable,
			wantHeader: map[string]string{"Retry-After": "3600"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rr := httptest.NewRecorder()
			tt.write(rr, httptest.NewRequest(http.MethodGet, "/api/v2/anime/1", nil))

			if rr.Code != tt.wantStatus {
				t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, tt.wantStatus)
			}
			for name, want := range tt.wantHeader {
				if got := rr.Header().Get(name); got != want {
					t.Errorf("%s = %q、期待値 = %q", name, got, want)
				}
			}

			body := decodeProblem(t, rr)
			if body["type"] != "about:blank" {
				t.Errorf("type = %v、期待値 = about:blank", body["type"])
			}
			if body["title"] != http.StatusText(tt.wantStatus) {
				t.Errorf("title = %v、期待値 = %q", body["title"], http.StatusText(tt.wantStatus))
			}
			if body["status"] != float64(tt.wantStatus) {
				t.Errorf("status = %v、期待値 = %d", body["status"], tt.wantStatus)
			}
			// 値が無くてもキーを省かずnullにする
			if detail, ok := body["detail"]; !ok || detail != nil {
				t.Errorf("detail = %v (キーの有無 = %v)、期待値 = null", detail, ok)
			}
			if _, ok := body["errors"]; ok {
				t.Error("about:blankの問題にerrorsが含まれている")
			}
		})
	}
}

func TestWriter_ValidationFailed(t *testing.T) {
	t.Parallel()

	pw := apierror.NewWriter("https://example.com/")
	parameter := "limit"

	rr := httptest.NewRecorder()
	pw.ValidationFailed(rr, httptest.NewRequest(http.MethodGet, "/api/v2/anime/works", nil), []apierror.FieldError{
		{Detail: "1以上を指定してください", Parameter: &parameter},
	})

	if rr.Code != http.StatusBadRequest {
		t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusBadRequest)
	}
	body := decodeProblem(t, rr)
	if want := "https://example.com/api/problems/validation-failed"; body["type"] != want {
		t.Errorf("type = %v、期待値 = %q", body["type"], want)
	}
	if body["status"] != float64(http.StatusBadRequest) {
		t.Errorf("status = %v、期待値 = 400", body["status"])
	}
	errs, ok := body["errors"].([]any)
	if !ok || len(errs) != 1 {
		t.Fatalf("errors = %v、期待値 = 1件", body["errors"])
	}
	fe := errs[0].(map[string]any)
	if fe["detail"] != "1以上を指定してください" || fe["parameter"] != "limit" {
		t.Errorf("errors[0] = %v", fe)
	}
	if pointer, ok := fe["pointer"]; !ok || pointer != nil {
		t.Errorf("errors[0].pointer = %v (キーの有無 = %v)、期待値 = null", pointer, ok)
	}
}

func TestWriter_ValidationFailed_空の一覧(t *testing.T) {
	t.Parallel()

	pw := apierror.NewWriter("https://example.com")
	rr := httptest.NewRecorder()
	pw.ValidationFailed(rr, httptest.NewRequest(http.MethodGet, "/api/v2/legacy_ids", nil), nil)

	body := decodeProblem(t, rr)
	errs, ok := body["errors"].([]any)
	if !ok || len(errs) != 0 {
		t.Errorf("errors = %v、期待値 = 空の配列", body["errors"])
	}
}

func TestWriter_WriteError(t *testing.T) {
	t.Parallel()

	pw := apierror.NewWriter("https://example.com")

	validationErr := model.NewValidationError()
	validationErr.AddGlobal("旧IDを1つ以上指定してください")
	validationErr.AddField("work_ids", "100件以内で指定してください")
	validationErr.AddField("episode_ids", "整数で指定してください")

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantErrors []map[string]any
	}{
		{
			name:       "ParameterErrorは400でパラメーターを示す",
			err:        fmt.Errorf("包んだエラー: %w", &apierror.ParameterError{Parameter: "cursor", Detail: "カーソルを読めません"}),
			wantStatus: http.StatusBadRequest,
			wantErrors: []map[string]any{
				{"detail": "カーソルを読めません", "pointer": nil, "parameter": "cursor"},
			},
		},
		{
			name:       "ValidationErrorは400で、全体の問題のあとにパラメーター名の順に並べる",
			err:        validationErr,
			wantStatus: http.StatusBadRequest,
			wantErrors: []map[string]any{
				{"detail": "旧IDを1つ以上指定してください", "pointer": nil, "parameter": nil},
				{"detail": "整数で指定してください", "pointer": nil, "parameter": "episode_ids"},
				{"detail": "100件以内で指定してください", "pointer": nil, "parameter": "work_ids"},
			},
		},
		{
			name:       "未存在のAppErrorは404",
			err:        &model.AppError{Code: model.AppErrCodeResourceNotFound, UserMsg: "見つかりません"},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "権限不足のAppErrorはリソースの存在を秘匿して404",
			err:        &model.AppError{Code: model.AppErrCodeForbidden, UserMsg: "権限がありません"},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "そのほかのAppErrorは500",
			err:        &model.AppError{Code: model.AppErrCodeInternal, UserMsg: "内部エラー"},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "素のエラーは500",
			err:        errors.New("予期しないエラー"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rr := httptest.NewRecorder()
			pw.WriteError(rr, httptest.NewRequest(http.MethodGet, "/api/v2/anime/1", nil), tt.err)

			if rr.Code != tt.wantStatus {
				t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, tt.wantStatus)
			}
			body := decodeProblem(t, rr)
			if tt.wantErrors == nil {
				if _, ok := body["errors"]; ok {
					t.Errorf("errors = %v、期待値 = キーが無い", body["errors"])
				}
				return
			}

			got, err := json.Marshal(body["errors"])
			if err != nil {
				t.Fatalf("errorsのエンコードのエラー = %v", err)
			}
			want, err := json.Marshal(tt.wantErrors)
			if err != nil {
				t.Fatalf("期待値のエンコードのエラー = %v", err)
			}
			if string(got) != string(want) {
				t.Errorf("errors = %s、期待値 = %s", got, want)
			}
		})
	}
}

func TestBearerChallenge_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		challenge apierror.BearerChallenge
		want      string
	}{
		{name: "パラメーターが無い", challenge: apierror.BearerChallenge{}, want: "Bearer"},
		{name: "エラーコードだけ", challenge: apierror.BearerChallenge{Error: "invalid_token"}, want: `Bearer error="invalid_token"`},
		{
			name:      "エラーコードとスコープ",
			challenge: apierror.BearerChallenge{Error: "insufficient_scope", Scope: []string{"read"}},
			want:      `Bearer error="insufficient_scope", scope="read"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.challenge.String(); got != tt.want {
				t.Errorf("String() = %q、期待値 = %q", got, tt.want)
			}
		})
	}
}

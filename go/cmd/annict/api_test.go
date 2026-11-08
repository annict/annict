package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/annict/annict/go/api"
	"github.com/annict/annict/go/internal/apierror"
	"github.com/annict/annict/go/internal/apigen"
	"github.com/annict/annict/go/internal/apihandler"
	apianime "github.com/annict/annict/go/internal/apihandler/anime"
	apiepisode "github.com/annict/annict/go/internal/apihandler/episode"
	apilegacyid "github.com/annict/annict/go/internal/apihandler/legacy_id"
	"github.com/annict/annict/go/internal/apihandler/openapi_description"
	apiwork "github.com/annict/annict/go/internal/apihandler/work"
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/ratelimit"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/validator"
)

// testAPIPrincipalsは、テストのトークンの文字列と、照合して得る呼び出し主体の対応
var testAPIPrincipals = map[string]*model.APIPrincipal{
	"flagged-token":   {UserID: 1, TokenKind: model.APITokenKindPersonalAccessToken, Scopes: []string{"read"}, RESTAPIV2Enabled: true},
	"unflagged-token": {UserID: 2, TokenKind: model.APITokenKindPersonalAccessToken, Scopes: []string{"read"}},
}

// testAPIPrincipalGetterはtestAPIPrincipalsでトークンを照合する
type testAPIPrincipalGetter struct{}

func (testAPIPrincipalGetter) Execute(_ context.Context, accessToken string) (*model.APIPrincipal, error) {
	return testAPIPrincipals[accessToken], nil
}

// testAPIRateLimiterは、数えたキーを残し、allowedに従って判定するレート制限
type testAPIRateLimiter struct {
	allowed bool
	panics  bool
	keys    []string
}

func (l *testAPIRateLimiter) Check(_ context.Context, input ratelimit.CheckInput) (*ratelimit.CheckResult, error) {
	l.keys = append(l.keys, input.Key)
	if l.panics {
		panic("計数中のパニック")
	}
	remaining := input.Limit - 1
	if !l.allowed {
		remaining = 0
	}
	return &ratelimit.CheckResult{
		Allowed:   l.allowed,
		Count:     input.Limit - remaining,
		Remaining: remaining,
		ResetAt:   time.Now().Add(30 * time.Second),
	}, nil
}

// newTestAPIHandlerは、serve.goと同じくmiddleware.APIPathPrefixにAPIのルーターをマウントしたハンドラーを返す。
// レート制限はすべてのリクエストを受け付ける
func newTestAPIHandler(t *testing.T) http.Handler {
	t.Helper()

	return newTestAPIHandlerWith(t, &testAPIRateLimiter{allowed: true}, &config.Config{Domain: "example.com"})
}

// newTestAPIHandlerWithは、レート制限を差し替えてnewTestAPIHandlerと同じハンドラーを返す
func newTestAPIHandlerWith(t *testing.T, limiter *testAPIRateLimiter, cfg *config.Config) http.Handler {
	t.Helper()

	// ルーターとミドルウェアだけのテストでは、作品・エピソードと旧IDのoperationを呼ばない
	server := apihandler.NewServer(
		apianime.NewHandler(nil),
		apiepisode.NewHandler(nil),
		apilegacyid.NewHandler(nil),
		openapi_description.NewHandler(api.OpenAPIDescription),
		apiwork.NewHandler(nil),
	)
	return newTestAPIHandlerWithServer(t, limiter, cfg, server)
}

// newTestAPIHandlerWithServerは、実際のUseCaseを注入したAPIも共通のミドルウェアを通して検証できるようにする
func newTestAPIHandlerWithServer(t *testing.T, limiter *testAPIRateLimiter, cfg *config.Config, server apigen.StrictServerInterface) http.Handler {
	t.Helper()

	problems := apierror.NewWriter("https://example.com")
	spec, err := apigen.GetSpec()
	if err != nil {
		t.Fatalf("GetSpec()のエラー = %v", err)
	}
	validator, err := middleware.NewAPIRequestValidator(spec, problems)
	if err != nil {
		t.Fatalf("NewAPIRequestValidator()のエラー = %v", err)
	}
	featureFlag, err := middleware.NewAPIFeatureFlag(spec, problems)
	if err != nil {
		t.Fatalf("NewAPIFeatureFlag()のエラー = %v", err)
	}
	rateLimit := middleware.NewAPIRateLimit(cfg, limiter, problems)
	tokenAuth := middleware.NewAPITokenAuth(testAPIPrincipalGetter{}, problems)
	r := chi.NewRouter()
	r.Use(middleware.APIResponseHeader)
	r.Use(middleware.APICORS)
	r.Use(rateLimit.CountMiddleware)
	r.Use(middleware.NewRequestBodyLimitMiddleware(10*1024*1024, problems).Middleware)
	r.Use(middleware.NewMaintenanceMiddleware(cfg, problems).Middleware)
	r.Use(rateLimit.EnforceMiddleware)
	r.Mount(middleware.APIPathPrefix, newAPIRouter(server, problems, tokenAuth, featureFlag, validator))
	return r
}

func TestAPIRouter_GetOpenAPIDescription(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	newTestAPIHandler(t).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v2/openapi.yaml", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("ステータスコード = %d、期待値 = %d (本文: %s)", rr.Code, http.StatusOK, rr.Body.String())
	}
	if got := rr.Header().Get("Content-Type"); got != "application/yaml" {
		t.Errorf("Content-Type = %q、期待値 = application/yaml", got)
	}
	body, err := io.ReadAll(rr.Body)
	if err != nil {
		t.Fatalf("本文を読めない: %v", err)
	}
	if !bytes.Equal(body, api.OpenAPIDescription) {
		t.Error("本文がapi/openapi.yamlと一致しない")
	}
}

func TestAPIRouter_Problems(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		method     string
		target     string
		wantStatus int
		wantAllow  string
	}{
		{name: "トークンの無いリクエストは記述にあるかを問わず404", method: http.MethodGet, target: "/api/v2/unknown", wantStatus: http.StatusNotFound},
		{name: "V2の接頭辞だけのパスは404", method: http.MethodGet, target: "/api/v2", wantStatus: http.StatusNotFound},
		{name: "受け付けないメソッドは405で使えるメソッドを示す", method: http.MethodPost, target: "/api/v2/openapi.yaml", wantStatus: http.StatusMethodNotAllowed, wantAllow: "GET"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rr := httptest.NewRecorder()
			newTestAPIHandler(t).ServeHTTP(rr, httptest.NewRequest(tt.method, tt.target, nil))

			if rr.Code != tt.wantStatus {
				t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, tt.wantStatus)
			}
			if got := rr.Header().Get("Content-Type"); got != apierror.ContentType {
				t.Errorf("Content-Type = %q、期待値 = %q", got, apierror.ContentType)
			}
			if got := rr.Header().Get("Allow"); got != tt.wantAllow {
				t.Errorf("Allow = %q、期待値 = %q", got, tt.wantAllow)
			}
			var problem apierror.Problem
			if err := json.Unmarshal(rr.Body.Bytes(), &problem); err != nil {
				t.Fatalf("本文をJSONとして読めない: %v (本文: %s)", err, rr.Body.String())
			}
			if problem.Status != tt.wantStatus {
				t.Errorf("status = %d、期待値 = %d", problem.Status, tt.wantStatus)
			}
		})
	}
}

func TestAPIRouter_RequestBodyTooLarge(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/api/v2/openapi.yaml", strings.NewReader("本文"))
	req.ContentLength = 10*1024*1024 + 1
	req.Header.Set("Origin", "https://client.example.com")
	rr := httptest.NewRecorder()
	newTestAPIHandler(t).ServeHTTP(rr, req)

	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusRequestEntityTooLarge)
	}
	wantHeaders := map[string]string{
		"Content-Type":                apierror.ContentType,
		"Cache-Control":               "no-store",
		"X-Content-Type-Options":      "nosniff",
		"X-Robots-Tag":                "noindex",
		"Content-Security-Policy":     "frame-ancestors 'none'",
		"Access-Control-Allow-Origin": "*",
	}
	for name, want := range wantHeaders {
		if got := rr.Header().Get(name); got != want {
			t.Errorf("%s = %q、期待値 = %q", name, got, want)
		}
	}
	var problem apierror.Problem
	if err := json.Unmarshal(rr.Body.Bytes(), &problem); err != nil {
		t.Fatalf("本文をProblem Detailsとして読めない: %v (本文: %s)", err, rr.Body.String())
	}
	if problem.Status != http.StatusRequestEntityTooLarge || problem.Type != "about:blank" {
		t.Errorf("Problem Details = %+v、期待値 = about:blankの413", problem)
	}
}

func TestAPIRouter_FeatureFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		target        string
		authorization string
		wantStatus    int
	}{
		{name: "OpenAPI記述はトークンが無くても返す", target: "/api/v2/openapi.yaml", wantStatus: http.StatusOK},
		{name: "OpenAPI記述は受け付けられないトークンでも返す", target: "/api/v2/openapi.yaml", authorization: "Bearer unknown-token", wantStatus: http.StatusOK},
		{name: "受け付けられないトークンは404", target: "/api/v2/unknown", authorization: "Bearer unknown-token", wantStatus: http.StatusNotFound},
		{name: "Bearerでない認証は404", target: "/api/v2/unknown", authorization: "Basic dXNlcjpwYXNz", wantStatus: http.StatusNotFound},
		{name: "フラグの無い利用者のトークンは404", target: "/api/v2/unknown", authorization: "Bearer unflagged-token", wantStatus: http.StatusNotFound},
		{name: "フラグの有効な利用者のトークンはルーターまで届く", target: "/api/v2/openapi.yaml", authorization: "Bearer flagged-token", wantStatus: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, tt.target, nil)
			if tt.authorization != "" {
				req.Header.Set("Authorization", tt.authorization)
			}
			rr := httptest.NewRecorder()
			newTestAPIHandler(t).ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("ステータスコード = %d、期待値 = %d (本文: %s)", rr.Code, tt.wantStatus, rr.Body.String())
			}
			if got := rr.Header().Get("WWW-Authenticate"); got != "" {
				t.Errorf("WWW-Authenticate = %q、フラグの判定では401を返さない", got)
			}
		})
	}
}

func TestAPIRouter_RateLimit(t *testing.T) {
	t.Parallel()

	t.Run("すべての応答に残量を付ける", func(t *testing.T) {
		t.Parallel()

		rr := httptest.NewRecorder()
		newTestAPIHandler(t).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v2/unknown", nil))

		if rr.Code != http.StatusNotFound {
			t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusNotFound)
		}
		if got := rr.Header().Get("RateLimit-Limit"); got != "60" {
			t.Errorf("RateLimit-Limit = %q、期待値 = 60", got)
		}
		if got := rr.Header().Get("RateLimit-Remaining"); got != "59" {
			t.Errorf("RateLimit-Remaining = %q、期待値 = 59", got)
		}
	})

	t.Run("上限を超えたら、トークンを照合する前に429を返す", func(t *testing.T) {
		t.Parallel()

		limiter := &testAPIRateLimiter{allowed: false}
		req := httptest.NewRequest(http.MethodGet, "/api/v2/openapi.yaml", nil)
		req.Header.Set("Authorization", "Bearer unknown-token")
		req.Header.Set("Origin", "https://client.example.com")
		rr := httptest.NewRecorder()
		newTestAPIHandlerWith(t, limiter, &config.Config{Domain: "example.com"}).ServeHTTP(rr, req)

		if rr.Code != http.StatusTooManyRequests {
			t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusTooManyRequests)
		}
		if got := rr.Header().Get("Retry-After"); got != rr.Header().Get("RateLimit-Reset") || got == "" {
			t.Errorf("Retry-After = %q、RateLimit-Reset = %q、期待値 = 同じ秒数", got, rr.Header().Get("RateLimit-Reset"))
		}
		if got := rr.Header().Get("Content-Type"); got != apierror.ContentType {
			t.Errorf("Content-Type = %q、期待値 = %q", got, apierror.ContentType)
		}
		if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "*" {
			t.Errorf("Access-Control-Allow-Origin = %q、ブラウザのスクリプトが429を読めるよう * を期待", got)
		}
		if len(limiter.keys) != 1 || !strings.HasPrefix(limiter.keys[0], "api:token:") {
			t.Errorf("数えたキー = %v、期待値 = api:token: で始まる1件", limiter.keys)
		}
	})
}

func TestAPIRouter_RateLimitEarlyResponses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		target      string
		host        string
		oversize    bool
		maintenance bool
		preflight   bool
		denied      bool
		panics      bool
		wantStatus  int
		wantCount   int
		wantHeaders bool
	}{
		{name: "通常の応答は1回だけ数える", wantStatus: http.StatusOK, wantCount: 1, wantHeaders: true},
		{name: "サイズ超過の413も数える", oversize: true, wantStatus: http.StatusRequestEntityTooLarge, wantCount: 1, wantHeaders: true},
		{name: "メンテナンスの503も数える", maintenance: true, wantStatus: http.StatusServiceUnavailable, wantCount: 1, wantHeaders: true},
		{name: "プリフライトは数えずに204で応える", preflight: true, wantStatus: http.StatusNoContent},
		{name: "上限を超えていてもプリフライトは204で応える", preflight: true, denied: true, wantStatus: http.StatusNoContent},
		{name: "メンテナンス中もプリフライトは204で応える", preflight: true, maintenance: true, wantStatus: http.StatusNoContent},
		{name: "制限の超過よりメンテナンスの503を優先する", maintenance: true, denied: true, wantStatus: http.StatusServiceUnavailable, wantCount: 1, wantHeaders: true},
		{name: "制限の超過よりサイズ超過の413を優先する", oversize: true, denied: true, wantStatus: http.StatusRequestEntityTooLarge, wantCount: 1, wantHeaders: true},
		{name: "Web画面はAPIのカウンターで数えない", target: "/works", wantStatus: http.StatusNotFound},
		{name: "Rails版のAPIホストはGo版で重ねて数えない", host: "api.example.com", wantStatus: http.StatusOK},
		{name: "計数中のパニックもProblem Detailsで返す", panics: true, wantStatus: http.StatusInternalServerError, wantCount: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			limiter := &testAPIRateLimiter{allowed: !tt.denied, panics: tt.panics}
			cfg := &config.Config{Domain: "example.com", MaintenanceMode: tt.maintenance}
			target := tt.target
			if target == "" {
				target = "/api/v2/openapi.yaml"
			}
			req := httptest.NewRequest(http.MethodGet, target, nil)
			req.Header.Set("Origin", "https://client.example.com")
			if tt.host != "" {
				req.Host = tt.host
			}
			if tt.oversize {
				req.Body = io.NopCloser(strings.NewReader("本文"))
				req.ContentLength = 10*1024*1024 + 1
			}
			if tt.preflight {
				req.Method = http.MethodOptions
				req.Header.Set("Access-Control-Request-Method", http.MethodGet)
				req.Header.Set("Access-Control-Request-Headers", "Authorization")
			}
			rr := httptest.NewRecorder()
			newTestAPIHandlerWith(t, limiter, cfg).ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("ステータスコード = %d、期待値 = %d (本文: %s)", rr.Code, tt.wantStatus, rr.Body.String())
			}
			if len(limiter.keys) != tt.wantCount {
				t.Errorf("計数の回数 = %d、期待値 = %d", len(limiter.keys), tt.wantCount)
			}
			for _, name := range []string{"RateLimit-Limit", "RateLimit-Remaining", "RateLimit-Reset"} {
				if got := rr.Header().Get(name); (got != "") != tt.wantHeaders {
					t.Errorf("%s = %q、ヘッダーがあるべきか = %v", name, got, tt.wantHeaders)
				}
			}
			if !middleware.IsAPIPath(target) {
				return
			}
			if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "*" {
				t.Errorf("Access-Control-Allow-Origin = %q、期待値 = *", got)
			}
			if got := rr.Header().Get("Cache-Control"); got != "no-store" {
				t.Errorf("Cache-Control = %q、期待値 = no-store", got)
			}
			if got := rr.Header().Get("Access-Control-Allow-Headers"); tt.preflight && got != "Authorization" {
				t.Errorf("Access-Control-Allow-Headers = %q、期待値 = Authorization (ブラウザが本要求を送れるように)", got)
			}
			if rr.Code == http.StatusServiceUnavailable && rr.Header().Get("Retry-After") != "3600" {
				t.Errorf("Retry-After = %q、期待値 = 3600", rr.Header().Get("Retry-After"))
			}
			if rr.Code >= 400 {
				if got := rr.Header().Get("Content-Type"); got != apierror.ContentType {
					t.Errorf("Content-Type = %q、期待値 = %q", got, apierror.ContentType)
				}
				var problem apierror.Problem
				if err := json.Unmarshal(rr.Body.Bytes(), &problem); err != nil {
					t.Fatalf("本文をProblem Detailsとして読めない: %v", err)
				}
				if problem.Status != rr.Code {
					t.Errorf("status = %d、期待値 = %d", problem.Status, rr.Code)
				}
			}
		})
	}
}

func TestAPIRouter_GetAnime_EpisodeParentStatus(t *testing.T) {
	t.Parallel()

	for _, status := range []model.AnimeStatus{model.AnimeStatusPublished, model.AnimeStatusArchived, model.AnimeStatusDeleted} {
		t.Run(string(status), func(t *testing.T) {
			t.Parallel()

			db, tx := testutil.SetupTx(t)
			ctx := context.Background()
			queries := query.New(db).WithTx(tx)
			animeRepo := repository.NewAnimeRepository(queries)
			classificationRepo := repository.NewAnimeClassificationRepository(queries)
			parent, err := animeRepo.Create(ctx, repository.CreateAnimeParams{
				Title:  sql.NullString{String: "親の作品のタイトル", Valid: true},
				Status: status,
			})
			if err != nil {
				t.Fatalf("親の作品の作成に失敗しました: %v", err)
			}
			if _, err := classificationRepo.Create(ctx, repository.CreateAnimeClassificationParams{
				AnimeID: parent.ID, Kind: model.AnimeClassificationKindWork,
			}); err != nil {
				t.Fatalf("親の分類の作成に失敗しました: %v", err)
			}
			episode, err := animeRepo.Create(ctx, repository.CreateAnimeParams{Status: model.AnimeStatusPublished})
			if err != nil {
				t.Fatalf("エピソードの作成に失敗しました: %v", err)
			}
			if _, err := classificationRepo.Create(ctx, repository.CreateAnimeClassificationParams{
				AnimeID: episode.ID, Kind: model.AnimeClassificationKindEpisode,
				ParentAnimeID: &parent.ID, SortNumber: sql.NullInt32{Int32: 1, Valid: true},
			}); err != nil {
				t.Fatalf("エピソードの分類の作成に失敗しました: %v", err)
			}

			server := newTestAPIServer(queries)
			handler := newTestAPIHandlerWithServer(t, &testAPIRateLimiter{allowed: true}, &config.Config{Domain: "example.com"}, server)
			req := httptest.NewRequest(http.MethodGet, "/api/v2/anime/"+episode.ID.String(), nil)
			req.Header.Set("Authorization", "Bearer flagged-token")
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if status != model.AnimeStatusPublished {
				if rr.Code != http.StatusNotFound || rr.Header().Get("Content-Type") != apierror.ContentType {
					t.Fatalf("応答 = %d %q %s、期待値 = Problem Detailsの404", rr.Code, rr.Header().Get("Content-Type"), rr.Body.String())
				}
				var problem apierror.Problem
				if err := json.Unmarshal(rr.Body.Bytes(), &problem); err != nil || problem.Status != http.StatusNotFound {
					t.Errorf("本文 = %s、エラー = %v、期待値 = statusが404のProblem Details", rr.Body.String(), err)
				}
				if strings.Contains(rr.Body.String(), parent.Title.String) {
					t.Error("非公開・削除済みの親のタイトルが返っている")
				}
				return
			}
			if rr.Code != http.StatusOK || rr.Header().Get("Content-Type") != "application/json" {
				t.Fatalf("応答 = %d %q %s、期待値 = JSONの200", rr.Code, rr.Header().Get("Content-Type"), rr.Body.String())
			}
			var body apigen.EpisodeAnime
			if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
				t.Fatalf("エピソードのJSONを読めない: %v", err)
			}
			if body.Id != int64(episode.ID) || body.Episode.Work.Id != int64(parent.ID) || body.Episode.Work.Title == nil || *body.Episode.Work.Title != parent.Title.String {
				t.Errorf("応答 = %s、期待値 = 公開中の親を持つエピソード", rr.Body.String())
			}
		})
	}
}

// newTestAPIServerは、serve.goと同じUseCaseを注入したAPIのハンドラーを返す
func newTestAPIServer(queries *query.Queries) *apihandler.Server {
	animeRepo := repository.NewAnimeRepository(queries)
	classificationRepo := repository.NewAnimeClassificationRepository(queries)
	legacyIDRepo := repository.NewAnimeLegacyIDRepository(queries)
	satelliteRepos := usecase.WorkSatelliteRepos{
		ExternalID:      repository.NewAnimeExternalIDRepository(queries),
		Link:            repository.NewAnimeLinkRepository(queries),
		OfficialAccount: repository.NewAnimeOfficialAccountRepository(queries),
		Hashtag:         repository.NewAnimeHashtagRepository(queries),
		Season:          repository.NewAnimeSeasonRepository(queries),
		Event:           repository.NewAnimeEventRepository(queries),
	}
	redirectRepo := repository.NewAnimeRedirectRepository(queries)
	return apihandler.NewServer(
		apianime.NewHandler(usecase.NewGetAPIAnimeUsecase(animeRepo, classificationRepo, redirectRepo, legacyIDRepo, satelliteRepos)),
		apiepisode.NewHandler(usecase.NewGetAPIEpisodesUsecase(animeRepo, classificationRepo, legacyIDRepo, satelliteRepos)),
		apilegacyid.NewHandler(usecase.NewGetAPILegacyIDsUsecase(validator.NewAPILegacyIDIndexValidator(), animeRepo, classificationRepo, redirectRepo, legacyIDRepo)),
		openapi_description.NewHandler(api.OpenAPIDescription),
		apiwork.NewHandler(usecase.NewGetAPIWorksUsecase(animeRepo, classificationRepo, legacyIDRepo, satelliteRepos)),
	)
}

// createTestAPIAnimeは、公開中の作品かエピソードのanimeと分類を作る。parentIDがあればエピソードにする
func createTestAPIAnime(t *testing.T, queries *query.Queries, title string, parentID *model.AnimeID, sortNumber int32) model.AnimeID {
	t.Helper()
	ctx := context.Background()
	anime, err := repository.NewAnimeRepository(queries).Create(ctx, repository.CreateAnimeParams{
		Title: sql.NullString{String: title, Valid: true},
	})
	if err != nil {
		t.Fatalf("animeの作成に失敗しました: %v", err)
	}
	params := repository.CreateAnimeClassificationParams{AnimeID: anime.ID, Kind: model.AnimeClassificationKindWork}
	if parentID != nil {
		params = repository.CreateAnimeClassificationParams{
			AnimeID: anime.ID, Kind: model.AnimeClassificationKindEpisode,
			ParentAnimeID: parentID, SortNumber: sql.NullInt32{Int32: sortNumber, Valid: true},
		}
	}
	if _, err := repository.NewAnimeClassificationRepository(queries).Create(ctx, params); err != nil {
		t.Fatalf("分類の作成に失敗しました: %v", err)
	}
	return anime.ID
}

// getTestAPIは、フラグの有効な利用者のトークンでAPIを呼ぶ
func getTestAPI(t *testing.T, handler http.Handler, target string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, target, nil)
	req.Header.Set("Authorization", "Bearer flagged-token")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

func TestAPIRouter_ListWorks(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	handler := newTestAPIHandlerWithServer(t, &testAPIRateLimiter{allowed: true}, &config.Config{Domain: "example.com"}, newTestAPIServer(queries))

	spring := model.SeasonNameSpring
	first := createTestAPIAnime(t, queries, "1つめの作品", nil, 0)
	second := createTestAPIAnime(t, queries, "2つめの作品", nil, 0)
	third := createTestAPIAnime(t, queries, "3つめの作品", nil, 0)
	createTestAPIAnime(t, queries, "エピソード", &first, 1)
	if _, err := repository.NewAnimeSeasonRepository(queries).Create(context.Background(), repository.CreateAnimeSeasonParams{
		AnimeID: second, Year: 3001, Name: &spring, IsPrimary: true,
	}); err != nil {
		t.Fatalf("季節の作成に失敗しました: %v", err)
	}
	ids := fmt.Sprintf("%d,%d,%d", first, second, third)

	listIDs := func(t *testing.T, rr *httptest.ResponseRecorder) (got []model.AnimeID, nextCursor *string) {
		t.Helper()
		if rr.Code != http.StatusOK {
			t.Fatalf("応答 = %d %s、期待値 = 200", rr.Code, rr.Body.String())
		}
		var body apigen.AnimeList
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatalf("一覧のJSONを読めない: %v", err)
		}
		for _, item := range body.Items {
			w, err := item.AsWorkAnime()
			if err != nil || w.Kind != apigen.Work {
				t.Fatalf("要素 = %v、エラー = %v、期待値 = 作品", item, err)
			}
			got = append(got, model.AnimeID(w.Id))
		}
		return got, body.NextCursor
	}

	t.Run("カンマ区切りのidsで絞り、Linkヘッダーで次のページを示す", func(t *testing.T) {
		rr := getTestAPI(t, handler, "/api/v2/anime/works?ids="+ids+"&limit=2")
		got, next := listIDs(t, rr)
		if fmt.Sprint(got) != fmt.Sprint([]model.AnimeID{first, second}) || next == nil {
			t.Fatalf("1ページめ = %v、next_cursor = %v、期待値 = [%d %d] と続きのカーソル", got, next, first, second)
		}
		link := rr.Header().Get("Link")
		if !strings.HasPrefix(link, "</api/v2/anime/works?") || !strings.HasSuffix(link, `>; rel="next"`) {
			t.Fatalf("Link = %q、期待値 = 次のページの相対URL", link)
		}

		got, next = listIDs(t, getTestAPI(t, handler, strings.TrimSuffix(strings.TrimPrefix(link, "<"), `>; rel="next"`)))
		if fmt.Sprint(got) != fmt.Sprint([]model.AnimeID{third}) || next != nil {
			t.Errorf("2ページめ = %v、next_cursor = %v、期待値 = [%d] で続き無し", got, next, third)
		}
	})

	t.Run("季節で絞る", func(t *testing.T) {
		got, _ := listIDs(t, getTestAPI(t, handler, "/api/v2/anime/works?ids="+ids+"&season=3001-spring,3001-summer&sort=-season"))
		if fmt.Sprint(got) != fmt.Sprint([]model.AnimeID{second}) {
			t.Errorf("作品 = %v、期待値 = [%d]", got, second)
		}
	})

	invalidTests := []struct {
		name      string
		target    string
		parameter string
	}{
		{name: "読めないカーソルは400", target: "/api/v2/anime/works?cursor=invalid", parameter: "cursor"},
		{name: "季節の形でないseasonは400", target: "/api/v2/anime/works?season=2026-autumn", parameter: "season"},
		{name: "idsの上限を超えると400", target: "/api/v2/anime/works?ids=" + strings.TrimSuffix(strings.Repeat("1,", 51), ","), parameter: "ids"},
	}
	for _, tt := range invalidTests {
		t.Run(tt.name, func(t *testing.T) {
			rr := getTestAPI(t, handler, tt.target)
			if rr.Code != http.StatusBadRequest || !strings.Contains(rr.Body.String(), `"parameter":"`+tt.parameter+`"`) {
				t.Errorf("応答 = %d %s、期待値 = %sの400", rr.Code, rr.Body.String(), tt.parameter)
			}
		})
	}

	t.Run("別の並び順で作ったカーソルは400", func(t *testing.T) {
		_, next := listIDs(t, getTestAPI(t, handler, "/api/v2/anime/works?ids="+ids+"&limit=1"))
		if next == nil {
			t.Fatal("続きのカーソルが無い")
		}
		rr := getTestAPI(t, handler, "/api/v2/anime/works?ids="+ids+"&limit=1&sort=season&cursor="+*next)
		if rr.Code != http.StatusBadRequest || !strings.Contains(rr.Body.String(), `"parameter":"cursor"`) {
			t.Errorf("応答 = %d %s、期待値 = cursorの400", rr.Code, rr.Body.String())
		}
	})
}

func TestAPIRouter_ListEpisodes(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	handler := newTestAPIHandlerWithServer(t, &testAPIRateLimiter{allowed: true}, &config.Config{Domain: "example.com"}, newTestAPIServer(queries))

	work := createTestAPIAnime(t, queries, "作品", nil, 0)
	second := createTestAPIAnime(t, queries, "第2話", &work, 2)
	first := createTestAPIAnime(t, queries, "第1話", &work, 1)

	t.Run("エピソードをsort_numberの順に返し、前後のエピソードを付ける", func(t *testing.T) {
		rr := getTestAPI(t, handler, "/api/v2/anime/"+work.String()+"/episodes")
		if rr.Code != http.StatusOK {
			t.Fatalf("応答 = %d %s、期待値 = 200", rr.Code, rr.Body.String())
		}
		var body apigen.AnimeList
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatalf("一覧のJSONを読めない: %v", err)
		}
		if len(body.Items) != 2 || body.NextCursor != nil {
			t.Fatalf("本文 = %s、期待値 = 続きの無い2件", rr.Body.String())
		}
		e1, err1 := body.Items[0].AsEpisodeAnime()
		e2, err2 := body.Items[1].AsEpisodeAnime()
		if err1 != nil || err2 != nil || e1.Id != int64(first) || e2.Id != int64(second) {
			t.Fatalf("本文 = %s、期待値 = 第1話・第2話の順", rr.Body.String())
		}
		if e1.Episode.NextEpisode == nil || e1.Episode.NextEpisode.Id != int64(second) || e2.Episode.PrevEpisode == nil || e2.Episode.PrevEpisode.Id != int64(first) {
			t.Errorf("本文 = %s、期待値 = 互いを前後に持つ", rr.Body.String())
		}
	})

	t.Run("エピソードのanime_idを指定すると404", func(t *testing.T) {
		rr := getTestAPI(t, handler, "/api/v2/anime/"+first.String()+"/episodes")
		if rr.Code != http.StatusNotFound || rr.Header().Get("Content-Type") != apierror.ContentType {
			t.Errorf("応答 = %d %q %s、期待値 = Problem Detailsの404", rr.Code, rr.Header().Get("Content-Type"), rr.Body.String())
		}
	})
}

func TestAPIRouter_ListCursorTrailingContent(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	handler := newTestAPIHandlerWithServer(t, &testAPIRateLimiter{allowed: true}, &config.Config{Domain: "example.com"}, newTestAPIServer(queries))
	work := createTestAPIAnime(t, queries, "作品1", nil, 0)
	other := createTestAPIAnime(t, queries, "作品2", nil, 0)
	createTestAPIAnime(t, queries, "第1話", &work, 1)
	createTestAPIAnime(t, queries, "第2話", &work, 2)

	for _, endpoint := range []struct {
		name   string
		target string
	}{
		{name: "作品の一覧", target: "/api/v2/anime/works?ids=" + work.String() + "," + other.String() + "&limit=1"},
		{name: "エピソードの一覧", target: "/api/v2/anime/" + work.String() + "/episodes?limit=1"},
	} {
		t.Run(endpoint.name, func(t *testing.T) {
			first := getTestAPI(t, handler, endpoint.target)
			var body apigen.AnimeList
			if err := json.Unmarshal(first.Body.Bytes(), &body); err != nil || first.Code != http.StatusOK || body.NextCursor == nil {
				t.Fatalf("カーソルを取得できません: 応答 = %d %s、エラー = %v", first.Code, first.Body.String(), err)
			}
			raw, err := base64.RawURLEncoding.DecodeString(*body.NextCursor)
			if err != nil {
				t.Fatalf("発行されたカーソルを読めません: %v", err)
			}
			for _, tt := range []struct {
				name       string
				suffix     string
				wantStatus int
			}{
				{name: "余分な閉じ角括弧は400", suffix: "]", wantStatus: http.StatusBadRequest},
				{name: "余分な閉じ波括弧は400", suffix: "}", wantStatus: http.StatusBadRequest},
				{name: "閉じ括弧の後ろに文字列があれば400", suffix: "]garbage", wantStatus: http.StatusBadRequest},
				{name: "末尾が空白だけなら200", suffix: " \n\t", wantStatus: http.StatusOK},
			} {
				t.Run(tt.name, func(t *testing.T) {
					// 応答で発行したJSONの後ろへ足し、カーソルの型や必要なキーが原因の400と区別する
					cursor := base64.RawURLEncoding.EncodeToString(append(append([]byte{}, raw...), tt.suffix...))
					rr := getTestAPI(t, handler, endpoint.target+"&cursor="+cursor)
					if rr.Code != tt.wantStatus {
						t.Fatalf("応答 = %d %s、期待値 = %d", rr.Code, rr.Body.String(), tt.wantStatus)
					}
					if tt.wantStatus == http.StatusBadRequest {
						var problem apierror.Problem
						if err := json.Unmarshal(rr.Body.Bytes(), &problem); err != nil || problem.Status != http.StatusBadRequest {
							t.Errorf("本文 = %s、エラー = %v、期待値 = statusが400のProblem Details", rr.Body.String(), err)
						}
						if rr.Header().Get("Content-Type") != apierror.ContentType || !strings.Contains(rr.Body.String(), `"parameter":"cursor"`) {
							t.Errorf("応答 = %q %s、期待値 = cursorのProblem Details", rr.Header().Get("Content-Type"), rr.Body.String())
						}
					}
				})
			}
		})
	}
}

func TestAPIRouter_ListLegacyIds(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	handler := newTestAPIHandlerWithServer(t, &testAPIRateLimiter{allowed: true}, &config.Config{Domain: "example.com"}, newTestAPIServer(queries))

	work := createTestAPIAnime(t, queries, "作品", nil, 0)
	legacyWorkID := testutil.NewWorkBuilder(t, tx).WithTitle("旧作品").Build()
	if err := repository.NewWorkRepository(queries).UpdateAnimeID(context.Background(), legacyWorkID, work); err != nil {
		t.Fatalf("旧作品のanime_idの設定に失敗しました: %v", err)
	}

	t.Run("旧IDごとのanime_idを返し、指定しなかった種類は空の配列にする", func(t *testing.T) {
		rr := getTestAPI(t, handler, "/api/v2/legacy_ids?work_ids="+legacyWorkID.String()+",999999999")
		if rr.Code != http.StatusOK || rr.Header().Get("Content-Type") != "application/json" {
			t.Fatalf("応答 = %d %q %s、期待値 = JSONの200", rr.Code, rr.Header().Get("Content-Type"), rr.Body.String())
		}
		want := `{"episode_ids":[],"work_ids":[{"anime_id":` + work.String() + `,"legacy_work_id":` + legacyWorkID.String() + `},{"anime_id":null,"legacy_work_id":999999999}]}`
		var got, wantJSON any
		if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
			t.Fatalf("JSONを読めない: %v", err)
		}
		_ = json.Unmarshal([]byte(want), &wantJSON)
		gotBytes, _ := json.Marshal(got)
		wantBytes, _ := json.Marshal(wantJSON)
		if string(gotBytes) != string(wantBytes) {
			t.Errorf("本文 = %s、期待値 = %s", gotBytes, wantBytes)
		}
	})

	manyIDs := strings.TrimSuffix(strings.Repeat("1,", 60), ",")
	badRequestTests := []struct {
		name   string
		target string
	}{
		{name: "旧IDのパラメーターが無いと400", target: "/api/v2/legacy_ids?work_id=1"},
		{name: "1つの種類で100件を超えると400", target: "/api/v2/legacy_ids?work_ids=" + strings.TrimSuffix(strings.Repeat("1,", 101), ",")},
		{name: "種類をまたいで合計100件を超えると400", target: "/api/v2/legacy_ids?work_ids=" + manyIDs + "&episode_ids=" + manyIDs},
		{name: "整数でない旧IDは400", target: "/api/v2/legacy_ids?work_ids=abc"},
	}
	for _, tt := range badRequestTests {
		t.Run(tt.name, func(t *testing.T) {
			rr := getTestAPI(t, handler, tt.target)
			if rr.Code != http.StatusBadRequest || rr.Header().Get("Content-Type") != apierror.ContentType {
				t.Errorf("応答 = %d %q %s、期待値 = Problem Detailsの400", rr.Code, rr.Header().Get("Content-Type"), rr.Body.String())
			}
		})
	}
}

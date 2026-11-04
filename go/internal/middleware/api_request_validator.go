package middleware

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"slices"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/legacy"

	"github.com/annict/annict/go/internal/apierror"
	"github.com/annict/annict/go/internal/model"
)

// securitySchemeTokenKindsは、OpenAPI記述のsecuritySchemesの名前と、そのスキームで受け付けるトークンの種類の対応
var securitySchemeTokenKinds = map[string]model.APITokenKind{
	"oauth2":              model.APITokenKindOAuthAccessToken,
	"personalAccessToken": model.APITokenKindPersonalAccessToken,
}

// errAPITokenMissingは、トークンの要るoperationを、照合済みのトークンの無いリクエストで呼んだことを表すエラー
var errAPITokenMissing = errors.New("トークンがありません")

// errAPITokenSchemeMismatchは、トークンの種類がsecuritySchemeと合わないことを表すエラー。
// securityの別の選択肢 (個人用アクセストークンとOAuthなど) が通れば、リクエストは受け付けられる
var errAPITokenSchemeMismatch = errors.New("トークンの種類がsecuritySchemeと合いません")

// insufficientScopeErrorは、トークンがoperationの要るスコープを持たないことを表すエラー
type insufficientScopeError struct {
	scopes []string
}

func (e *insufficientScopeError) Error() string {
	return "トークンのスコープが足りません: " + strings.Join(e.scopes, " ")
}

// APIRequestValidatorはREST API V2のリクエストを、OpenAPI記述のパラメーター・リクエスト本文・securityで検証するミドルウェア。
// 形式の検証 (必須・型・範囲・列挙値) はここで行い、ハンドラーやUseCaseでは繰り返さない
type APIRequestValidator struct {
	router   routers.Router
	problems *apierror.Writer
	options  *openapi3filter.Options
}

// NewAPIRequestValidatorは新しいAPIRequestValidatorを作成する。
// specにはOpenAPI記述 (`apigen.GetSpec()`) を渡す
func NewAPIRequestValidator(spec *openapi3.T, problems *apierror.Writer) (*APIRequestValidator, error) {
	router, err := legacy.NewRouter(spec)
	if err != nil {
		return nil, fmt.Errorf("OpenAPI記述からルーターを作成できません: %w", err)
	}

	return &APIRequestValidator{
		router:   router,
		problems: problems,
		options: &openapi3filter.Options{
			// パラメーター単位の問題をまとめて返すため、最初の問題で止めない
			MultiError:         true,
			AuthenticationFunc: authenticateAPIRequest,
		},
	}, nil
}

// authenticateAPIRequestは、operationのsecurityに書かれたスコープを、APITokenAuthがコンテキストに
// 入れた呼び出し主体のトークンが持つかを確かめる。
// スコープの要求はOpenAPI記述を唯一の正本とし、ハンドラーやUseCaseに二重に書かない。
// securityが空のoperation (`getOpenAPIDescription` など) では呼ばれない
func authenticateAPIRequest(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
	principal := APIPrincipalFromContext(ctx)
	if principal == nil {
		return errAPITokenMissing
	}
	if kind, ok := securitySchemeTokenKinds[input.SecuritySchemeName]; !ok || kind != principal.TokenKind {
		return errAPITokenSchemeMismatch
	}
	for _, scope := range input.Scopes {
		if !principal.HasScope(scope) {
			return &insufficientScopeError{scopes: input.Scopes}
		}
	}
	return nil
}

// MiddlewareはHTTPミドルウェアを返す
func (v *APIRequestValidator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route, pathParams, err := v.router.FindRoute(routeLookupRequest(r))
		if err != nil {
			// 記述に無いパスやメソッドは、ルーターの404・405に任せる
			next.ServeHTTP(w, r)
			return
		}

		input := &openapi3filter.RequestValidationInput{
			Request:    r,
			PathParams: pathParams,
			Route:      route,
			Options:    v.options,
		}
		if err := openapi3filter.ValidateRequest(r.Context(), input); err != nil {
			v.writeValidationError(w, r, err)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// writeValidationErrorは検証の失敗をProblem Detailsで書き込む。
// 認証の失敗は内容の問題より先に401で返す
func (v *APIRequestValidator) writeValidationError(w http.ResponseWriter, r *http.Request, err error) {
	errs := flattenErrors(err)

	for _, e := range errs {
		var securityErr *openapi3filter.SecurityRequirementsError
		if errors.As(e, &securityErr) {
			v.writeSecurityError(w, r, securityErr)
			return
		}
	}

	var fieldErrs []apierror.FieldError
	for _, e := range errs {
		var requestErr *openapi3filter.RequestError
		if !errors.As(e, &requestErr) {
			slog.ErrorContext(r.Context(), "APIのリクエストの検証で予期しないエラーが発生しました", "error", e)
			v.problems.InternalServerError(w, r)
			return
		}
		fieldErrs = append(fieldErrs, requestFieldErrors(requestErr)...)
	}

	v.problems.ValidationFailed(w, r, fieldErrs)
}

// writeSecurityErrorはsecurityの検証の失敗を書き込む。
// トークンはあるがスコープが足りない場合は403、それ以外は401を返す
func (v *APIRequestValidator) writeSecurityError(w http.ResponseWriter, r *http.Request, securityErr *openapi3filter.SecurityRequirementsError) {
	for _, e := range securityErr.Errors {
		var scopeErr *insufficientScopeError
		if errors.As(e, &scopeErr) {
			v.problems.Forbidden(w, r, apierror.BearerChallenge{Error: "insufficient_scope", Scope: scopeErr.scopes})
			return
		}
	}

	challenge := apierror.BearerChallenge{}
	// トークンを送ってきたリクエストにだけエラーコードを付ける (RFC 6750 §3.1)
	if APIPrincipalFromContext(r.Context()) != nil || apiTokenRejectedFromContext(r.Context()) {
		challenge.Error = "invalid_token"
	}
	v.problems.Unauthorized(w, r, challenge)
}

// requestFieldErrorsはkin-openapiのRequestErrorを、問題の箇所ごとのFieldErrorに変換する
func requestFieldErrors(requestErr *openapi3filter.RequestError) []apierror.FieldError {
	newFieldError := func(detail string, pointer string) apierror.FieldError {
		fe := apierror.FieldError{Detail: detail}
		if p := requestErr.Parameter; p != nil {
			name := p.Name
			fe.Parameter = &name
		} else {
			fe.Pointer = &pointer
		}
		return fe
	}

	schemaErrs := leafSchemaErrors(requestErr.Err)
	if len(schemaErrs) == 0 {
		return []apierror.FieldError{newFieldError(requestErrorDetail(requestErr), jsonPointer())}
	}

	fieldErrs := make([]apierror.FieldError, 0, len(schemaErrs))
	for _, schemaErr := range schemaErrs {
		pointer, detail := schemaErrorLocation(schemaErr)
		fieldErrs = append(fieldErrs, newFieldError(detail, pointer))
	}
	slices.SortStableFunc(fieldErrs, func(a, b apierror.FieldError) int {
		return strings.Compare(derefOrEmpty(a.Pointer), derefOrEmpty(b.Pointer))
	})
	return fieldErrs
}

// leafSchemaErrorsはerrに含まれるSchemaErrorを、入れ子の末端まで辿って返す。
// OpenAPI 3.1の記述ではJSON Schema 2020-12の検証器が使われ、複数の問題は
// 「全体のSchemaError (Originに個々のSchemaErrorのMultiError)」の形で返るため
func leafSchemaErrors(err error) []*openapi3.SchemaError {
	var leaves []*openapi3.SchemaError
	for _, e := range flattenErrors(err) {
		var schemaErr *openapi3.SchemaError
		if !errors.As(e, &schemaErr) {
			continue
		}
		var causes openapi3.MultiError
		if schemaErr.Origin != nil && errors.As(schemaErr.Origin, &causes) {
			if nested := leafSchemaErrors(causes); len(nested) > 0 {
				leaves = append(leaves, nested...)
				continue
			}
		}
		leaves = append(leaves, schemaErr)
	}
	return leaves
}

// jsonSchemaErrorReasonPatternは、JSON Schema 2020-12の検証器のエラーをkin-openapiが変換したSchemaErrorの理由。
// 例: `error at "/title": at '/title': minLength: got 0, want 1`
// 位置はSchemaError.JSONPointer() で取得し、理由からは位置の接頭辞を取り除く
var jsonSchemaErrorReasonPattern = regexp.MustCompile(`(?s)^(?:error at "[^"]*": )?at '[^']*': (.*)$`)

// schemaErrorLocationはSchemaErrorの位置 (URIフラグメント形式のJSON Pointer) と、位置を除いた説明を返す
func schemaErrorLocation(schemaErr *openapi3.SchemaError) (string, string) {
	detail := schemaErr.Reason
	if m := jsonSchemaErrorReasonPattern.FindStringSubmatch(schemaErr.Reason); m != nil {
		detail = m[1]
	}
	return jsonPointer(schemaErr.JSONPointer()...), detail
}

// jsonPointerはリクエスト本文の中の位置を、URIフラグメント形式のJSON Pointer (RFC 6901 §6) で表す
func jsonPointer(tokens ...string) string {
	var b strings.Builder
	b.WriteString("#")
	for _, token := range tokens {
		b.WriteString("/")
		token = strings.ReplaceAll(token, "~", "~0")
		token = strings.ReplaceAll(token, "/", "~1")
		b.WriteString(token)
	}
	return b.String()
}

func derefOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// requestErrorDetailはRequestErrorの理由を、どのパラメーターかを除いた文で返す
// (パラメーター名や本文の位置はFieldErrorの別の項目で示す)
func requestErrorDetail(requestErr *openapi3filter.RequestError) string {
	reason := requestErr.Reason
	if e := requestErr.Err; e != nil {
		if reason == "" || reason == e.Error() {
			return e.Error()
		}
		return reason + ": " + e.Error()
	}
	return reason
}

// flattenErrorsはopenapi3.MultiErrorを入れ子まで展開した一覧を返す。
// RequestErrorなどがUnwrapで返すMultiErrorは展開しない (どのパラメーターの問題かが失われるため)
func flattenErrors(err error) []error {
	if err == nil {
		return nil
	}
	multi, ok := err.(openapi3.MultiError)
	if !ok {
		return []error{err}
	}
	var errs []error
	for _, e := range multi {
		errs = append(errs, flattenErrors(e)...)
	}
	return errs
}

// routeLookupRequestは、OpenAPI記述のルートを探すためのリクエストを返す。
// 記述のserversは相対URL (`/api/v2`) のため、プロキシ向けの絶対形式のリクエスト (RFC 9112 §3.2.2) でも
// パスだけで照合できるよう、URLからスキームとホストを外す
func routeLookupRequest(r *http.Request) *http.Request {
	if r.URL.Scheme == "" && r.URL.Host == "" {
		return r
	}
	u := *r.URL
	u.Scheme = ""
	u.Host = ""
	lookup := r.WithContext(r.Context())
	lookup.URL = &u
	return lookup
}

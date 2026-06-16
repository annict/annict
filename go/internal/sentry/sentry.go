// Package sentry はSentryエラー追跡サービスとの連携機能を提供します
package sentry

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
)

// Config holds the Sentry settings.
//
// [Ja] Sentry の設定を保持する。
type Config struct {
	DSN              string
	Environment      string
	Release          string
	TracesSampleRate float64
	Debug            bool
}

const maskedValue = "[FILTERED]"

// sensitiveHeaders はマスクすべきHTTPヘッダー名のリスト（小文字）
var sensitiveHeaders = []string{
	"authorization",
	"cookie",
	"x-csrf-token",
}

// sensitiveBodyKeys はマスクすべきリクエストボディのキー（部分一致、小文字）
var sensitiveBodyKeys = []string{
	"password",
	"token",
	"secret",
}

// sensitiveQueryKeys はマスクすべきクエリパラメータのキー（部分一致、小文字）
var sensitiveQueryKeys = []string{
	"token",
	"key",
}

// sensitiveTagKeys lists tag keys to mask (partial match, lowercase). Sentry
// tags are populated from slog attributes by sentryslog, so PII logged as a
// structured attribute (e.g. "email" on email-send failure logs) would reach
// Sentry unmasked without this filter. The stderr log keeps the original
// value, so debugging is still possible there.
//
// [Ja] マスクすべきタグのキー (部分一致、小文字)。Sentry のタグには sentryslog
// 経由で slog 属性がそのまま乗るため、構造化属性としてログに載せた PII (例:
// メール送信失敗ログの "email") がマスクされないまま Sentry に届いてしまう。
// 標準エラー出力側のログには元の値が残るため、デバッグはそちらで行える。
var sensitiveTagKeys = []string{
	"email",
	"password",
	"secret",
	"token",
}

// ignoredErrorPatterns lists message-level patterns that skip Sentry capture (regular expression).
// These filter out client-disconnect noise and Go runtime's normal aborts.
//
// [Ja] メッセージレベルで Sentry 送信をスキップするパターン (正規表現)。
// クライアント切断由来のノイズや Go runtime の正常な中断をフィルタする。
var ignoredErrorPatterns = []string{
	"context canceled",
	"net/http: abort Handler",
}

// Init initializes Sentry. If the DSN is empty, initialization is skipped
// and nil is returned (used when Sentry is not used in development environments).
//
// [Ja] Sentry を初期化する。DSN が空の場合は初期化をスキップし nil を返す
// (開発環境で Sentry を使用しない場合)。
func Init(cfg Config) error {
	if cfg.DSN == "" {
		slog.Info("Sentry DSNが設定されていないため、Sentryは無効化されています")
		return nil
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.DSN,
		Environment:      cfg.Environment,
		Release:          cfg.Release,
		TracesSampleRate: cfg.TracesSampleRate,
		EnableTracing:    true,
		Debug:            cfg.Debug,
		BeforeSend:       beforeSend,
		IgnoreErrors:     ignoredErrorPatterns,
	})
	if err != nil {
		return err
	}

	slog.Info("Sentryを初期化しました",
		"environment", cfg.Environment,
		"release", cfg.Release,
		"traces_sample_rate", cfg.TracesSampleRate,
	)
	return nil
}

// beforeSend filters events before sending them to Sentry.
// Errors caused by client disconnects or normal aborts are dropped,
// reverse-proxy noise is filtered out via the source tag, and the rest
// has its sensitive data masked.
//
// [Ja] Sentry にイベントを送信する前にフィルタリングを行う。
// クライアント切断や正常な中断由来のエラーは破棄し、リバースプロキシ経由の
// ノイズは source タグで識別して捨てる。残りはセンシティブデータをマスクする。
func beforeSend(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
	if hint != nil && shouldDropError(hint.OriginalException) {
		return nil
	}

	// Drop events tagged as reverse-proxy failures: those failures belong to
	// the Rails Sentry project, not the Go one. sentryslog stamps the slog
	// attribute SourceAttrKey onto event.Tags, so a simple tag check suffices.
	//
	// [Ja] リバースプロキシ由来とタグ付けされたイベントは捨てる。Rails 側の
	// 障害は Rails の Sentry プロジェクトで扱うべきため。sentryslog は slog
	// 属性 SourceAttrKey を event.Tags にそのまま乗せるので、タグ照合で判別できる。
	if event.Tags[SourceAttrKey] == ReverseProxySource {
		return nil
	}

	filterTags(event)

	if event.Request != nil {
		filterRequestHeaders(event.Request)
		filterRequestData(event.Request)
		filterQueryString(event.Request)
	}
	return event
}

// shouldDropError reports whether the error is caused by a client disconnect
// or runtime abort. If true, the event should not be sent to Sentry.
//
// [Ja] クライアント切断・runtime 中断由来のエラーかを判定する。
// 該当する場合は Sentry に送らない。
func shouldDropError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) {
		return true
	}
	if errors.Is(err, http.ErrAbortHandler) {
		return true
	}
	return false
}

// filterTags masks sensitive tag values such as email addresses. Unlike the
// request filters below, this also covers events generated from slog records,
// whose attributes sentryslog stamps onto event.Tags.
//
// [Ja] センシティブなタグ (メールアドレス等) をマスクする。下のリクエスト系
// フィルタと異なり、sentryslog が slog 属性を event.Tags に乗せて生成した
// イベントもカバーする。
func filterTags(event *sentry.Event) {
	for key := range event.Tags {
		lowerKey := strings.ToLower(key)
		for _, sensitive := range sensitiveTagKeys {
			if strings.Contains(lowerKey, sensitive) {
				event.Tags[key] = maskedValue
				break
			}
		}
	}
}

// filterRequestHeaders はセンシティブなHTTPヘッダーをマスクする
func filterRequestHeaders(req *sentry.Request) {
	if req.Headers == nil {
		return
	}

	for headerName := range req.Headers {
		lowerName := strings.ToLower(headerName)
		for _, sensitive := range sensitiveHeaders {
			if lowerName == sensitive {
				req.Headers[headerName] = maskedValue
				break
			}
		}
	}
}

// filterRequestData はセンシティブなリクエストボディのフィールドをマスクする
func filterRequestData(req *sentry.Request) {
	if req.Data == "" {
		return
	}

	// フォームデータ（application/x-www-form-urlencoded）をパースしてフィルタリング
	values, err := url.ParseQuery(req.Data)
	if err != nil {
		// パースできない場合は安全のためデータ全体を削除
		req.Data = maskedValue
		return
	}

	filtered := false
	for key := range values {
		lowerKey := strings.ToLower(key)
		for _, sensitive := range sensitiveBodyKeys {
			if strings.Contains(lowerKey, sensitive) {
				values.Set(key, maskedValue)
				filtered = true
				break
			}
		}
	}

	if filtered {
		req.Data = values.Encode()
	}
}

// filterQueryString はセンシティブなクエリパラメータをマスクする
func filterQueryString(req *sentry.Request) {
	if req.QueryString == "" {
		return
	}

	values, err := url.ParseQuery(req.QueryString)
	if err != nil {
		// パースできない場合は安全のためクエリ全体を削除
		req.QueryString = maskedValue
		return
	}

	filtered := false
	for key := range values {
		lowerKey := strings.ToLower(key)
		for _, sensitive := range sensitiveQueryKeys {
			if strings.Contains(lowerKey, sensitive) {
				values.Set(key, maskedValue)
				filtered = true
				break
			}
		}
	}

	if filtered {
		req.QueryString = values.Encode()
	}
}

// Flush はバッファリングされたイベントをSentryに送信する
// アプリケーション終了時に呼び出す
func Flush(timeout time.Duration) {
	sentry.Flush(timeout)
}

// CaptureError はエラーをSentryに送信する
func CaptureError(ctx context.Context, err error) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		hub.CaptureException(err)
	} else {
		sentry.CaptureException(err)
	}
}

// CaptureMessage はメッセージをSentryに送信する
func CaptureMessage(ctx context.Context, message string) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		hub.CaptureMessage(message)
	} else {
		sentry.CaptureMessage(message)
	}
}

// Package sentry はSentryエラー追跡サービスとの連携機能を提供します
package sentry

import (
	"context"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
)

// Config はSentryの設定を保持する
type Config struct {
	DSN              string
	Environment      string
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

// Init はSentryを初期化する
// DSNが空の場合は初期化をスキップしtrueを返す（開発環境でSentryを使用しない場合）
func Init(cfg Config) error {
	if cfg.DSN == "" {
		slog.Info("Sentry DSNが設定されていないため、Sentryは無効化されています")
		return nil
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.DSN,
		Environment:      cfg.Environment,
		TracesSampleRate: cfg.TracesSampleRate,
		Debug:            cfg.Debug,
		BeforeSend:       beforeSend,
	})
	if err != nil {
		return err
	}

	slog.Info("Sentryを初期化しました", "environment", cfg.Environment, "traces_sample_rate", cfg.TracesSampleRate)
	return nil
}

// beforeSend はSentryにイベントを送信する前にセンシティブデータをフィルタリングする
func beforeSend(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
	if event.Request != nil {
		filterRequestHeaders(event.Request)
		filterRequestData(event.Request)
		filterQueryString(event.Request)
	}
	return event
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

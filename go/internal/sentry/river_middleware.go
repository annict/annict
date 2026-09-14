package sentry

import (
	"context"
	"strconv"

	"github.com/getsentry/sentry-go"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

// RiverWorkerMiddlewareはWorkerが最終的に返したエラーをSentryイベントとして捕捉するriverの
// WorkerMiddlewareを返す。各ジョブはCloneしたHubの上で動くため、ジョブ単位の
// タグ (job.kind / job.attempt) が他ジョブに漏れることはない。また、Cloneした
// Hubをctxにbindするので、ジョブ内のslog.ErrorContext (NewSlogHandler経由)
// も同じスコープにイベントを乗せる。
//
// shouldDropErrorに該当するエラー (context.Canceled / http.ErrAbortHandler)
// はriverのリトライ判断を変えないようにそのままreturnするが、Sentryには
// 送らない。シャットダウン由来のノイズやruntime中断で誤通知させないため。
func RiverWorkerMiddleware() rivertype.WorkerMiddleware {
	return river.WorkerMiddlewareFunc(func(ctx context.Context, job *rivertype.JobRow, doInner func(ctx context.Context) error) error {
		hub := cloneHubForJob(ctx)
		hub.Scope().SetTag("job.kind", job.Kind)
		hub.Scope().SetTag("job.attempt", strconv.Itoa(job.Attempt))
		ctx = sentry.SetHubOnContext(ctx, hub)

		err := doInner(ctx)
		if err != nil && !shouldDropError(err) {
			hub.CaptureException(err)
		}
		return err
	})
}

// ジョブ実行用にHubをCloneして返す。ctxに既にHubが乗っている場合
// (例: テストが事前に注入したケース) はそれをCloneしてスコープ情報を保つ。
// 無ければグローバルのCurrentHubをCloneする。いずれにしてもジョブ単位の
// スコープ書き換えが元のHubに波及しない。
func cloneHubForJob(ctx context.Context) *sentry.Hub {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		return hub.Clone()
	}
	return sentry.CurrentHub().Clone()
}

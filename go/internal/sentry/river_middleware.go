package sentry

import (
	"context"
	"strconv"

	"github.com/getsentry/sentry-go"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

// RiverWorkerMiddleware returns a river WorkerMiddleware that captures the
// final error returned by a worker as a Sentry event. Each job runs against a
// cloned Hub so per-job tags (job.kind / job.attempt) do not leak across jobs,
// and the cloned Hub is bound to ctx so anything captured inside the worker
// (notably slog.ErrorContext via sentryslog) shares the same scope.
//
// Errors that match shouldDropError (context.Canceled / http.ErrAbortHandler)
// are still returned to river so its retry logic stays unchanged, but they are
// not sent to Sentry: shutdown noise and runtime aborts should never page on
// their own.
//
// [Ja] Worker が最終的に返したエラーを Sentry イベントとして捕捉する river の
// WorkerMiddleware を返す。各ジョブは Clone した Hub の上で動くため、ジョブ単位の
// タグ (job.kind / job.attempt) が他ジョブに漏れることはない。また、Clone した
// Hub を ctx に bind するので、ジョブ内の slog.ErrorContext (sentryslog 経由)
// も同じスコープにイベントを乗せる。
//
// shouldDropError に該当するエラー (context.Canceled / http.ErrAbortHandler)
// は river のリトライ判断を変えないようにそのまま return するが、Sentry には
// 送らない。シャットダウン由来のノイズや runtime 中断で誤通知させないため。
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

// cloneHubForJob returns a fresh Hub clone for the running job. If ctx already
// carries a Hub (in case a caller pre-populated one, e.g. tests), that Hub is
// cloned so its scope data is preserved; otherwise the global CurrentHub is
// cloned. Cloning either way isolates per-job scope mutations from the source.
//
// [Ja] ジョブ実行用に Hub を Clone して返す。ctx に既に Hub が乗っている場合
// (例: テストが事前に注入したケース) はそれを Clone してスコープ情報を保つ。
// 無ければグローバルの CurrentHub を Clone する。いずれにしてもジョブ単位の
// スコープ書き換えが元の Hub に波及しない。
func cloneHubForJob(ctx context.Context) *sentry.Hub {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		return hub.Clone()
	}
	return sentry.CurrentHub().Clone()
}

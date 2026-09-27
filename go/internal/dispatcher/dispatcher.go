// Package dispatcherはジョブキューへの投入を抽象化する。
// Repositoryがデータベースアクセスを抽象化するのと同じ発想で、
// Dispatcherがジョブキューアクセスを抽象化する。
package dispatcher

import (
	"context"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

// --- ジョブ引数型 ---

// SendSignInCodeEmailArgsはログインコード送信メールジョブの引数
type SendSignInCodeEmailArgs struct {
	Email  string `json:"email"`
	Code   string `json:"code"`
	Locale string `json:"locale"`
}

// Kindはジョブの種類を返す
func (SendSignInCodeEmailArgs) Kind() string { return "send_sign_in_code_email" }

// InsertOptsはジョブのInsertオプションを返す
func (SendSignInCodeEmailArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{Queue: river.QueueDefault, MaxAttempts: 5}
}

// SendSignUpCodeEmailArgsは新規登録確認コード送信メールジョブの引数
type SendSignUpCodeEmailArgs struct {
	Email  string `json:"email"`
	Code   string `json:"code"`
	Locale string `json:"locale"`
}

// Kindはジョブの種類を返す
func (SendSignUpCodeEmailArgs) Kind() string { return "send_sign_up_code_email" }

// InsertOptsはジョブのInsertオプションを返す
func (SendSignUpCodeEmailArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{Queue: river.QueueDefault, MaxAttempts: 5}
}

// SendPasswordResetEmailArgsはパスワードリセットメール送信ジョブの引数
type SendPasswordResetEmailArgs struct {
	Email    string `json:"email"`
	ResetURL string `json:"reset_url"`
	Locale   string `json:"locale"`
}

// Kindはジョブの種類を返す
func (SendPasswordResetEmailArgs) Kind() string { return "send_password_reset_email" }

// InsertOptsはジョブのInsertオプションを返す
func (SendPasswordResetEmailArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{Queue: river.QueueDefault, MaxAttempts: 5}
}

// --- 定期ジョブの引数型 ---
//
// 以下のジョブはUseCaseから投入するのではなく、Riverの定期ジョブとして
// cmd/annict/serve.goで登録し、Riverがスケジュールに従って投入する。
// そのためEnqueue*メソッドは持たない。Kind()の文字列はriver_jobに保存され、
// Workerを引くキーになるため変えないこと。

// CleanupExpiredTokensArgsはトークンクリーンアップジョブの引数。
type CleanupExpiredTokensArgs struct{}

// Kindはジョブの種類を返す。
func (CleanupExpiredTokensArgs) Kind() string { return "cleanup_expired_tokens" }

// InsertOptsはジョブ挿入時のデフォルトオプションを返す。
func (CleanupExpiredTokensArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{Queue: river.QueueDefault, MaxAttempts: 3}
}

// CleanupExpiredSignInCodesArgsは期限切れログインコードのクリーンアップジョブの引数。
type CleanupExpiredSignInCodesArgs struct{}

// Kindはジョブの種類を返す。
func (CleanupExpiredSignInCodesArgs) Kind() string { return "cleanup_expired_sign_in_codes" }

// InsertOptsはジョブ挿入時のデフォルトオプションを返す。
func (CleanupExpiredSignInCodesArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{Queue: river.QueueDefault, MaxAttempts: 3}
}

// CleanupExpiredSessionsArgsは期限切れセッションのクリーンアップジョブの引数。
type CleanupExpiredSessionsArgs struct{}

// Kindはジョブの種類を返す。
func (CleanupExpiredSessionsArgs) Kind() string { return "cleanup_expired_sessions" }

// InsertOptsはジョブ挿入時のデフォルトオプションを返す。
func (CleanupExpiredSessionsArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{Queue: river.QueueDefault, MaxAttempts: 3}
}

// SyncAnimesArgsはフェーズ2のフル・リコンシリエーションバッチジョブの引数型。
// ペイロードは持たない。ジョブはworks / episodesテーブル全体をリコンサイルするため、
// 実行ごとにパラメータ化するものがない。
type SyncAnimesArgs struct{}

// Kindはジョブの種類を返す。
func (SyncAnimesArgs) Kind() string { return "sync_animes" }

// InsertOptsはジョブ挿入時のデフォルトオプションを返す。リコンサイルは冪等なので、
// 一時的な失敗は数回まで安全に再試行でき、取りこぼしは次回の定期実行でも拾われる。
func (SyncAnimesArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{Queue: river.QueueDefault, MaxAttempts: 3}
}

// --- Dispatcher ---

// JobInserterはジョブをキューに追加するインターフェース
type JobInserter interface {
	Insert(ctx context.Context, args river.JobArgs, opts *river.InsertOpts) (*rivertype.JobInsertResult, error)
}

// Dispatcherはバックグラウンドジョブの投入を担当する
type Dispatcher struct {
	client JobInserter
}

// NewDispatcherは新しいDispatcherを生成する
func NewDispatcher(client JobInserter) *Dispatcher {
	return &Dispatcher{client: client}
}

// EnqueueSignInCodeEmailはログインコード送信メールジョブをキューに追加する
func (d *Dispatcher) EnqueueSignInCodeEmail(ctx context.Context, email, code, locale string) error {
	args := SendSignInCodeEmailArgs{Email: email, Code: code, Locale: locale}
	opts := args.InsertOpts()
	_, err := d.client.Insert(ctx, args, &opts)
	return err
}

// EnqueueSignUpCodeEmailは新規登録確認コード送信メールジョブをキューに追加する
func (d *Dispatcher) EnqueueSignUpCodeEmail(ctx context.Context, email, code, locale string) error {
	args := SendSignUpCodeEmailArgs{Email: email, Code: code, Locale: locale}
	opts := args.InsertOpts()
	_, err := d.client.Insert(ctx, args, &opts)
	return err
}

// EnqueuePasswordResetEmailはパスワードリセットメール送信ジョブをキューに追加する
func (d *Dispatcher) EnqueuePasswordResetEmail(ctx context.Context, email, resetURL, locale string) error {
	args := SendPasswordResetEmailArgs{Email: email, ResetURL: resetURL, Locale: locale}
	opts := args.InsertOpts()
	_, err := d.client.Insert(ctx, args, &opts)
	return err
}

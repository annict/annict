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

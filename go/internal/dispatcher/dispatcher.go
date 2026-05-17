// Package dispatcher はジョブキューへの投入を抽象化する。
// Repository がデータベースアクセスを抽象化するのと同じ発想で、
// Dispatcher がジョブキューアクセスを抽象化する。
package dispatcher

import (
	"context"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

// --- ジョブ引数型 ---

// SendSignInCodeEmailArgs はログインコード送信メールジョブの引数
type SendSignInCodeEmailArgs struct {
	Email  string `json:"email"`
	Code   string `json:"code"`
	Locale string `json:"locale"`
}

// Kind はジョブの種類を返す
func (SendSignInCodeEmailArgs) Kind() string { return "send_sign_in_code_email" }

// InsertOpts はジョブの Insert オプションを返す
func (SendSignInCodeEmailArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{Queue: river.QueueDefault, MaxAttempts: 5}
}

// SendSignUpCodeEmailArgs は新規登録確認コード送信メールジョブの引数
type SendSignUpCodeEmailArgs struct {
	Email  string `json:"email"`
	Code   string `json:"code"`
	Locale string `json:"locale"`
}

// Kind はジョブの種類を返す
func (SendSignUpCodeEmailArgs) Kind() string { return "send_sign_up_code_email" }

// InsertOpts はジョブの Insert オプションを返す
func (SendSignUpCodeEmailArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{Queue: river.QueueDefault, MaxAttempts: 5}
}

// SendPasswordResetEmailArgs はパスワードリセットメール送信ジョブの引数
type SendPasswordResetEmailArgs struct {
	Email    string `json:"email"`
	ResetURL string `json:"reset_url"`
	Locale   string `json:"locale"`
}

// Kind はジョブの種類を返す
func (SendPasswordResetEmailArgs) Kind() string { return "send_password_reset_email" }

// InsertOpts はジョブの Insert オプションを返す
func (SendPasswordResetEmailArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{Queue: river.QueueDefault, MaxAttempts: 5}
}

// --- Dispatcher ---

// JobInserter はジョブをキューに追加するインターフェース
type JobInserter interface {
	Insert(ctx context.Context, args river.JobArgs, opts *river.InsertOpts) (*rivertype.JobInsertResult, error)
}

// Dispatcher はバックグラウンドジョブの投入を担当する
type Dispatcher struct {
	client JobInserter
}

// NewDispatcher は新しい Dispatcher を生成する
func NewDispatcher(client JobInserter) *Dispatcher {
	return &Dispatcher{client: client}
}

// EnqueueSignInCodeEmail はログインコード送信メールジョブをキューに追加する
func (d *Dispatcher) EnqueueSignInCodeEmail(ctx context.Context, email, code, locale string) error {
	args := SendSignInCodeEmailArgs{Email: email, Code: code, Locale: locale}
	opts := args.InsertOpts()
	_, err := d.client.Insert(ctx, args, &opts)
	return err
}

// EnqueueSignUpCodeEmail は新規登録確認コード送信メールジョブをキューに追加する
func (d *Dispatcher) EnqueueSignUpCodeEmail(ctx context.Context, email, code, locale string) error {
	args := SendSignUpCodeEmailArgs{Email: email, Code: code, Locale: locale}
	opts := args.InsertOpts()
	_, err := d.client.Insert(ctx, args, &opts)
	return err
}

// EnqueuePasswordResetEmail はパスワードリセットメール送信ジョブをキューに追加する
func (d *Dispatcher) EnqueuePasswordResetEmail(ctx context.Context, email, resetURL, locale string) error {
	args := SendPasswordResetEmailArgs{Email: email, ResetURL: resetURL, Locale: locale}
	opts := args.InsertOpts()
	_, err := d.client.Insert(ctx, args, &opts)
	return err
}

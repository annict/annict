// Package dispatcher はバックグラウンドジョブの投入を担当します
package dispatcher

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"

	"github.com/annict/annict/go/internal/worker"
)

// Dispatcher はバックグラウンドジョブの投入を担当する
type Dispatcher struct {
	riverClient *river.Client[pgx.Tx]
}

// NewDispatcher は新しい Dispatcher を作成する
func NewDispatcher(riverClient *river.Client[pgx.Tx]) *Dispatcher {
	return &Dispatcher{riverClient: riverClient}
}

// InsertSignInCodeEmail はログインコード送信メールジョブを投入する
func (d *Dispatcher) InsertSignInCodeEmail(ctx context.Context, email, code, locale string) error {
	args, err := worker.BuildSignInCodeEmail(ctx, email, code, locale)
	if err != nil {
		return err
	}
	_, err = d.riverClient.Insert(ctx, *args, nil)
	return err
}

// InsertSignUpCodeEmail は新規登録確認コード送信メールジョブを投入する
func (d *Dispatcher) InsertSignUpCodeEmail(ctx context.Context, email, code, locale string) error {
	args, err := worker.BuildSignUpCodeEmail(ctx, email, code, locale)
	if err != nil {
		return err
	}
	_, err = d.riverClient.Insert(ctx, *args, nil)
	return err
}

// InsertPasswordResetEmail はパスワードリセットメール送信ジョブを投入する
func (d *Dispatcher) InsertPasswordResetEmail(ctx context.Context, email, resetURL, locale string) error {
	args, err := worker.BuildPasswordResetEmail(ctx, email, resetURL, locale)
	if err != nil {
		return err
	}
	_, err = d.riverClient.Insert(ctx, *args, nil)
	return err
}

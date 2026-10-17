package dispatcher

import (
	"context"
	"database/sql"
	"testing"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

// mockJobInserterはテスト用のモック
type mockJobInserter struct {
	called bool
	args   river.JobArgs
	opts   *river.InsertOpts
}

func (m *mockJobInserter) Insert(_ context.Context, args river.JobArgs, opts *river.InsertOpts) (*rivertype.JobInsertResult, error) {
	m.called = true
	m.args = args
	m.opts = opts
	return &rivertype.JobInsertResult{}, nil
}

// mockTxJobInserterはトランザクションの中での投入を記録するテスト用のモック
type mockTxJobInserter struct {
	called bool
	tx     *sql.Tx
	args   river.JobArgs
	opts   *river.InsertOpts
}

func (m *mockTxJobInserter) InsertTx(_ context.Context, tx *sql.Tx, args river.JobArgs, opts *river.InsertOpts) (*rivertype.JobInsertResult, error) {
	m.called = true
	m.tx = tx
	m.args = args
	m.opts = opts
	return &rivertype.JobInsertResult{}, nil
}

// assertEnqueueInvocationはInsertが期待通りのoptsで呼ばれたことを検証する
func assertEnqueueInvocation(t *testing.T, mock *mockJobInserter) {
	t.Helper()
	if !mock.called {
		t.Fatal("Insertが呼ばれていません")
	}
	if mock.opts == nil {
		t.Fatal("InsertOptsがnilです")
	}
	if mock.opts.MaxAttempts != 5 {
		t.Errorf("MaxAttempts = %d、期待値 = 5", mock.opts.MaxAttempts)
	}
	if mock.opts.Queue != river.QueueDefault {
		t.Errorf("Queue = %s、期待値 = %s", mock.opts.Queue, river.QueueDefault)
	}
}

func TestEnqueueSignInCodeEmail(t *testing.T) {
	t.Parallel()

	mock := &mockJobInserter{}
	d := NewDispatcher(mock, nil)

	err := d.EnqueueSignInCodeEmail(context.Background(), "test@example.com", "123456", "ja")
	if err != nil {
		t.Fatalf("EnqueueSignInCodeEmail()のエラー = %v", err)
	}

	assertEnqueueInvocation(t, mock)

	args, ok := mock.args.(SendSignInCodeEmailArgs)
	if !ok {
		t.Fatalf("argsの型がSendSignInCodeEmailArgsではありません: %T", mock.args)
	}
	if args.Email != "test@example.com" {
		t.Errorf("Email = %s、期待値 = test@example.com", args.Email)
	}
	if args.Code != "123456" {
		t.Errorf("Code = %s、期待値 = 123456", args.Code)
	}
	if args.Locale != "ja" {
		t.Errorf("Locale = %s、期待値 = ja", args.Locale)
	}
}

func TestEnqueueSignUpCodeEmail(t *testing.T) {
	t.Parallel()

	mock := &mockJobInserter{}
	d := NewDispatcher(mock, nil)

	err := d.EnqueueSignUpCodeEmail(context.Background(), "test@example.com", "654321", "en")
	if err != nil {
		t.Fatalf("EnqueueSignUpCodeEmail()のエラー = %v", err)
	}

	assertEnqueueInvocation(t, mock)

	args, ok := mock.args.(SendSignUpCodeEmailArgs)
	if !ok {
		t.Fatalf("argsの型がSendSignUpCodeEmailArgsではありません: %T", mock.args)
	}
	if args.Email != "test@example.com" {
		t.Errorf("Email = %s、期待値 = test@example.com", args.Email)
	}
	if args.Code != "654321" {
		t.Errorf("Code = %s、期待値 = 654321", args.Code)
	}
	if args.Locale != "en" {
		t.Errorf("Locale = %s、期待値 = en", args.Locale)
	}
}

func TestEnqueuePasswordResetEmail(t *testing.T) {
	t.Parallel()

	mock := &mockJobInserter{}
	d := NewDispatcher(mock, nil)

	err := d.EnqueuePasswordResetEmail(context.Background(), "test@example.com", "https://example.com/password/edit?token=abc", "ja")
	if err != nil {
		t.Fatalf("EnqueuePasswordResetEmail()のエラー = %v", err)
	}

	assertEnqueueInvocation(t, mock)

	args, ok := mock.args.(SendPasswordResetEmailArgs)
	if !ok {
		t.Fatalf("argsの型がSendPasswordResetEmailArgsではありません: %T", mock.args)
	}
	if args.Email != "test@example.com" {
		t.Errorf("Email = %s、期待値 = test@example.com", args.Email)
	}
	if args.ResetURL != "https://example.com/password/edit?token=abc" {
		t.Errorf("ResetURL = %s、期待値 = https://example.com/password/edit?token=abc", args.ResetURL)
	}
	if args.Locale != "ja" {
		t.Errorf("Locale = %s、期待値 = ja", args.Locale)
	}
}

func TestArgs_Kind(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		gotKind string
		want    string
	}{
		{"SendSignInCodeEmailArgs", (SendSignInCodeEmailArgs{}).Kind(), "send_sign_in_code_email"},
		{"SendSignUpCodeEmailArgs", (SendSignUpCodeEmailArgs{}).Kind(), "send_sign_up_code_email"},
		{"SendPasswordResetEmailArgs", (SendPasswordResetEmailArgs{}).Kind(), "send_password_reset_email"},
		{"SendAccountDeletionEmailArgs", (SendAccountDeletionEmailArgs{}).Kind(), "send_account_deletion_email"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.gotKind != tt.want {
				t.Errorf("Kind() = %s、期待値 = %s", tt.gotKind, tt.want)
			}
		})
	}
}

func TestArgs_InsertOpts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		opts            river.InsertOpts
		wantQueue       string
		wantMaxAttempts int
	}{
		{
			name:            "SendSignInCodeEmailArgs",
			opts:            SendSignInCodeEmailArgs{}.InsertOpts(),
			wantQueue:       river.QueueDefault,
			wantMaxAttempts: 5,
		},
		{
			name:            "SendSignUpCodeEmailArgs",
			opts:            SendSignUpCodeEmailArgs{}.InsertOpts(),
			wantQueue:       river.QueueDefault,
			wantMaxAttempts: 5,
		},
		{
			name:            "SendPasswordResetEmailArgs",
			opts:            SendPasswordResetEmailArgs{}.InsertOpts(),
			wantQueue:       river.QueueDefault,
			wantMaxAttempts: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.opts.Queue != tt.wantQueue {
				t.Errorf("Queue = %q、期待値 = %q", tt.opts.Queue, tt.wantQueue)
			}
			if tt.opts.MaxAttempts != tt.wantMaxAttempts {
				t.Errorf("MaxAttempts = %d、期待値 = %d", tt.opts.MaxAttempts, tt.wantMaxAttempts)
			}
		})
	}
}

// TestPeriodicJobArgsは定期ジョブのArgsのKindとInsertOptsを検証する。
// Kindの文字列はriver_jobに保存され、Workerを引くキーになるため固定する。
func TestPeriodicJobArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args interface {
			river.JobArgs
			InsertOpts() river.InsertOpts
		}
		wantKind        string
		wantQueue       string
		wantMaxAttempts int
	}{
		{
			name:            "CleanupExpiredTokensArgs",
			args:            CleanupExpiredTokensArgs{},
			wantKind:        "cleanup_expired_tokens",
			wantQueue:       river.QueueDefault,
			wantMaxAttempts: 3,
		},
		{
			name:            "CleanupExpiredSignInCodesArgs",
			args:            CleanupExpiredSignInCodesArgs{},
			wantKind:        "cleanup_expired_sign_in_codes",
			wantQueue:       river.QueueDefault,
			wantMaxAttempts: 3,
		},
		{
			name:            "CleanupExpiredRateLimitsArgs",
			args:            CleanupExpiredRateLimitsArgs{},
			wantKind:        "cleanup_expired_rate_limits",
			wantQueue:       river.QueueDefault,
			wantMaxAttempts: 3,
		},
		{
			name:            "CleanupExpiredSessionsArgs",
			args:            CleanupExpiredSessionsArgs{},
			wantKind:        "cleanup_expired_sessions",
			wantQueue:       river.QueueDefault,
			wantMaxAttempts: 3,
		},
		{
			name:            "CleanupAnonymousSessionsArgs",
			args:            CleanupAnonymousSessionsArgs{},
			wantKind:        "cleanup_anonymous_sessions",
			wantQueue:       river.QueueDefault,
			wantMaxAttempts: 3,
		},
		{
			name:            "SyncAnimesArgs",
			args:            SyncAnimesArgs{},
			wantKind:        "sync_animes",
			wantQueue:       river.QueueDefault,
			wantMaxAttempts: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.args.Kind(); got != tt.wantKind {
				t.Errorf("Kind() = %q、期待値 = %q", got, tt.wantKind)
			}
			opts := tt.args.InsertOpts()
			if opts.Queue != tt.wantQueue {
				t.Errorf("Queue = %q、期待値 = %q", opts.Queue, tt.wantQueue)
			}
			if opts.MaxAttempts != tt.wantMaxAttempts {
				t.Errorf("MaxAttempts = %d、期待値 = %d", opts.MaxAttempts, tt.wantMaxAttempts)
			}
		})
	}
}

func TestDeleteUserArgs(t *testing.T) {
	t.Parallel()

	args := DeleteUserArgs{UserID: 42}
	if got := args.Kind(); got != "delete_user" {
		t.Errorf("Kind() = %q、期待値 = %q", got, "delete_user")
	}

	opts := args.InsertOpts()
	if opts.Queue != river.QueueDefault {
		t.Errorf("Queue = %q、期待値 = %q", opts.Queue, river.QueueDefault)
	}
	if opts.MaxAttempts != 10 {
		t.Errorf("MaxAttempts = %d、期待値 = 10", opts.MaxAttempts)
	}
	// 同じユーザーの削除ジョブが並行して進まないよう、引数 (ユーザーID) で一意にする。
	if !opts.UniqueOpts.ByArgs {
		t.Error("UniqueOpts.ByArgs = false、期待値 = true")
	}
}

// TestEnqueueDeleteUserTxは削除ジョブを、渡したトランザクションの中で
// ユーザーIDで一意なジョブとして投入することを検証する。
func TestEnqueueDeleteUserTx(t *testing.T) {
	t.Parallel()

	mock := &mockTxJobInserter{}
	d := NewDispatcher(&mockJobInserter{}, mock)
	tx := &sql.Tx{}

	if err := d.EnqueueDeleteUserTx(context.Background(), tx, 42); err != nil {
		t.Fatalf("EnqueueDeleteUserTx()のエラー = %v", err)
	}

	if !mock.called {
		t.Fatal("InsertTxが呼ばれていません")
	}
	if mock.tx != tx {
		t.Error("渡したトランザクションでInsertTxが呼ばれていません")
	}
	args, ok := mock.args.(DeleteUserArgs)
	if !ok {
		t.Fatalf("argsの型がDeleteUserArgsではありません: %T", mock.args)
	}
	if args.UserID != 42 {
		t.Errorf("UserID = %d、期待値 = 42", args.UserID)
	}
	if mock.opts == nil {
		t.Fatal("InsertOptsがnilです")
	}
	if !mock.opts.UniqueOpts.ByArgs {
		t.Error("UniqueOpts.ByArgs = false、期待値 = true")
	}
	if mock.opts.MaxAttempts != 10 {
		t.Errorf("MaxAttempts = %d、期待値 = 10", mock.opts.MaxAttempts)
	}
}

// TestEnqueueAccountDeletionEmailTxは退会完了メールのジョブを、渡したトランザクションの中で
// 投入することを検証する。
func TestEnqueueAccountDeletionEmailTx(t *testing.T) {
	t.Parallel()

	mock := &mockTxJobInserter{}
	d := NewDispatcher(&mockJobInserter{}, mock)
	tx := &sql.Tx{}

	if err := d.EnqueueAccountDeletionEmailTx(context.Background(), tx, "user@example.com", "en", true); err != nil {
		t.Fatalf("EnqueueAccountDeletionEmailTx()のエラー = %v", err)
	}

	if !mock.called {
		t.Fatal("InsertTxが呼ばれていません")
	}
	if mock.tx != tx {
		t.Error("渡したトランザクションでInsertTxが呼ばれていません")
	}
	args, ok := mock.args.(SendAccountDeletionEmailArgs)
	if !ok {
		t.Fatalf("argsの型がSendAccountDeletionEmailArgsではありません: %T", mock.args)
	}
	want := SendAccountDeletionEmailArgs{Email: "user@example.com", Locale: "en", IsStripeSupporter: true}
	if args != want {
		t.Errorf("args = %+v、期待値 = %+v", args, want)
	}
	if mock.opts == nil {
		t.Fatal("InsertOptsがnilです")
	}
	if mock.opts.MaxAttempts != 5 {
		t.Errorf("MaxAttempts = %d、期待値 = 5", mock.opts.MaxAttempts)
	}
}

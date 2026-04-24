package dispatcher

import (
	"context"
	"testing"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

// mockJobInserter はテスト用のモック
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

// assertEnqueueInvocation は Insert が期待通りの opts で呼ばれたことを検証する
func assertEnqueueInvocation(t *testing.T, mock *mockJobInserter) {
	t.Helper()
	if !mock.called {
		t.Fatal("Insert が呼ばれていません")
	}
	if mock.opts == nil {
		t.Fatal("InsertOpts が nil です")
	}
	if mock.opts.MaxAttempts != 5 {
		t.Errorf("MaxAttempts = %d, want 5", mock.opts.MaxAttempts)
	}
	if mock.opts.Queue != river.QueueDefault {
		t.Errorf("Queue = %s, want %s", mock.opts.Queue, river.QueueDefault)
	}
}

func TestEnqueueSignInCodeEmail(t *testing.T) {
	t.Parallel()

	mock := &mockJobInserter{}
	d := NewDispatcher(mock)

	err := d.EnqueueSignInCodeEmail(context.Background(), "test@example.com", "123456", "ja")
	if err != nil {
		t.Fatalf("EnqueueSignInCodeEmail() error = %v", err)
	}

	assertEnqueueInvocation(t, mock)

	args, ok := mock.args.(SendSignInCodeEmailArgs)
	if !ok {
		t.Fatalf("args の型が SendSignInCodeEmailArgs ではありません: %T", mock.args)
	}
	if args.Email != "test@example.com" {
		t.Errorf("Email = %s, want test@example.com", args.Email)
	}
	if args.Code != "123456" {
		t.Errorf("Code = %s, want 123456", args.Code)
	}
	if args.Locale != "ja" {
		t.Errorf("Locale = %s, want ja", args.Locale)
	}
}

func TestEnqueueSignUpCodeEmail(t *testing.T) {
	t.Parallel()

	mock := &mockJobInserter{}
	d := NewDispatcher(mock)

	err := d.EnqueueSignUpCodeEmail(context.Background(), "test@example.com", "654321", "en")
	if err != nil {
		t.Fatalf("EnqueueSignUpCodeEmail() error = %v", err)
	}

	assertEnqueueInvocation(t, mock)

	args, ok := mock.args.(SendSignUpCodeEmailArgs)
	if !ok {
		t.Fatalf("args の型が SendSignUpCodeEmailArgs ではありません: %T", mock.args)
	}
	if args.Email != "test@example.com" {
		t.Errorf("Email = %s, want test@example.com", args.Email)
	}
	if args.Code != "654321" {
		t.Errorf("Code = %s, want 654321", args.Code)
	}
	if args.Locale != "en" {
		t.Errorf("Locale = %s, want en", args.Locale)
	}
}

func TestEnqueuePasswordResetEmail(t *testing.T) {
	t.Parallel()

	mock := &mockJobInserter{}
	d := NewDispatcher(mock)

	err := d.EnqueuePasswordResetEmail(context.Background(), "test@example.com", "https://example.com/password/edit?token=abc", "ja")
	if err != nil {
		t.Fatalf("EnqueuePasswordResetEmail() error = %v", err)
	}

	assertEnqueueInvocation(t, mock)

	args, ok := mock.args.(SendPasswordResetEmailArgs)
	if !ok {
		t.Fatalf("args の型が SendPasswordResetEmailArgs ではありません: %T", mock.args)
	}
	if args.Email != "test@example.com" {
		t.Errorf("Email = %s, want test@example.com", args.Email)
	}
	if args.ResetURL != "https://example.com/password/edit?token=abc" {
		t.Errorf("ResetURL = %s, want https://example.com/password/edit?token=abc", args.ResetURL)
	}
	if args.Locale != "ja" {
		t.Errorf("Locale = %s, want ja", args.Locale)
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.gotKind != tt.want {
				t.Errorf("Kind() = %s, want %s", tt.gotKind, tt.want)
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
				t.Errorf("Queue = %q, want %q", tt.opts.Queue, tt.wantQueue)
			}
			if tt.opts.MaxAttempts != tt.wantMaxAttempts {
				t.Errorf("MaxAttempts = %d, want %d", tt.opts.MaxAttempts, tt.wantMaxAttempts)
			}
		})
	}
}

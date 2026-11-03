package usecase

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/auth"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/password_reset"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/validator"
)

// passwordResetRaceFixtureは、パスワードリセットと競合する操作を挟むテストで使うユーザーとトークンをまとめる。
// UseCaseが開くトランザクションと競合させるため、準備した行はコミットしておく。
type passwordResetRaceFixture struct {
	db           *sql.DB
	uc           *UpdatePasswordResetUsecase
	userID       model.UserID
	originalHash string
	token        string
	resetTokenID model.PasswordResetTokenID
}

func newPasswordResetRaceFixture(t *testing.T, ctx context.Context) *passwordResetRaceFixture {
	t.Helper()

	db := testutil.GetTestDB()
	originalHash, err := auth.HashPassword("OldPassword123!")
	if err != nil {
		t.Fatalf("元のパスワードのハッシュ化に失敗: %v", err)
	}
	userID, _ := createCommittedUserWithBuilderForTest(t, db, func(tx *sql.Tx) *testutil.UserBuilder {
		return testutil.NewUserBuilder(t, tx).WithEncryptedPassword(originalHash)
	}, false)
	queries := query.New(db)
	tokenRepo := repository.NewPasswordResetTokenRepository(queries)
	token, err := password_reset.GenerateToken()
	if err != nil {
		t.Fatalf("トークンの生成に失敗: %v", err)
	}
	resetToken, err := tokenRepo.Create(ctx, userID, password_reset.HashToken(token), time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("トークンの作成に失敗: %v", err)
	}

	return &passwordResetRaceFixture{
		db:           db,
		uc:           NewUpdatePasswordResetUsecase(db, tokenRepo, repository.NewUserRepository(queries), validator.NewPasswordUpdateValidator()),
		userID:       userID,
		originalHash: originalHash,
		token:        token,
		resetTokenID: resetToken.ID,
	}
}

// beginBlockerは、競合する操作を行うトランザクションを開き、そのバックエンドのIDと一緒に返す。
func (f *passwordResetRaceFixture) beginBlocker(t *testing.T, ctx context.Context) (*sql.Tx, int) {
	t.Helper()

	blocker, err := f.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("競合する操作のトランザクションの開始に失敗: %v", err)
	}
	t.Cleanup(func() { _ = blocker.Rollback() })
	var blockerPID int
	if err := blocker.QueryRowContext(ctx, "SELECT pg_backend_pid()").Scan(&blockerPID); err != nil {
		t.Fatalf("競合する操作のバックエンドIDの取得に失敗: %v", err)
	}
	return blocker, blockerPID
}

// executeWhileBlockedは、リセットがblockerの行ロックを待つことを確かめてからblockerをコミットし、リセットの結果を返す。
func (f *passwordResetRaceFixture) executeWhileBlocked(t *testing.T, ctx context.Context, blocker *sql.Tx, blockerPID int) (*UpdatePasswordResetOutput, error) {
	t.Helper()

	type result struct {
		output *UpdatePasswordResetOutput
		err    error
	}
	results := make(chan result, 1)
	go func() {
		output, err := f.uc.Execute(ctx, UpdatePasswordResetInput{Token: f.token, Password: "NewPassword456!", PasswordConfirmation: "NewPassword456!"})
		results <- result{output: output, err: err}
	}()
	awaitBlockedByBackend(t, ctx, f.db, blockerPID, results)
	if err := blocker.Commit(); err != nil {
		t.Fatalf("競合する操作のコミットに失敗: %v", err)
	}
	select {
	case got := <-results:
		return got.output, got.err
	case <-ctx.Done():
		t.Fatalf("競合する操作の確定後もリセットが終わらなかった: %v", ctx.Err())
		return nil, nil
	}
}

// assertPasswordUnchangedは、ユーザーのパスワードが元のままであることを確かめる。
func (f *passwordResetRaceFixture) assertPasswordUnchanged(t *testing.T, ctx context.Context) {
	t.Helper()

	var savedHash string
	if err := f.db.QueryRowContext(ctx, "SELECT encrypted_password FROM users WHERE id = $1", int64(f.userID)).Scan(&savedHash); err != nil {
		t.Fatalf("パスワードの取得に失敗: %v", err)
	}
	if savedHash != f.originalHash {
		t.Error("リセットを拒んだのにパスワードが変更された")
	}
}

// tokenUsedは、トークンが使用済みかどうかを返す。
func (f *passwordResetRaceFixture) tokenUsed(t *testing.T, ctx context.Context) bool {
	t.Helper()

	var usedAt sql.NullTime
	if err := f.db.QueryRowContext(ctx, "SELECT used_at FROM password_reset_tokens WHERE id = $1", int64(f.resetTokenID)).Scan(&usedAt); err != nil {
		t.Fatalf("トークンの状態の取得に失敗: %v", err)
	}
	return usedAt.Valid
}

// TestUpdatePasswordResetUsecase_Execute_UserDeactivatedDuringResetは、トークン検証の後に利用停止が
// 確定した場合、パスワードを変えず、トークンも消費せずに無効なトークンのエラーを返すことを検証する。
func TestUpdatePasswordResetUsecase_Execute_UserDeactivatedDuringReset(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	f := newPasswordResetRaceFixture(t, ctx)

	// 利用停止を未コミットにしておくと、トークンの検証は更新前のユーザーを読み、
	// パスワードのUPDATEはこの行のロックを待つ。待機を確認してから利用停止を確定させる。
	blocker, blockerPID := f.beginBlocker(t, ctx)
	if _, err := blocker.ExecContext(ctx, "UPDATE users SET deleted_at = NOW() WHERE id = $1", int64(f.userID)); err != nil {
		t.Fatalf("利用停止の設定に失敗: %v", err)
	}

	output, err := f.executeWhileBlocked(t, ctx, blocker, blockerPID)

	if output != nil || !errors.Is(err, ErrInvalidPasswordResetToken) {
		t.Errorf("利用停止後のリセット = (%+v, %v)、結果なしと無効なトークンのエラーを期待", output, err)
	}
	f.assertPasswordUnchanged(t, ctx)
	if f.tokenUsed(t, ctx) {
		t.Error("利用停止したユーザーのトークンが使用済みになった")
	}
}

// TestUpdatePasswordResetUsecase_Execute_TokenUsedDuringResetは、トークン検証の後に同じトークンを使った
// 別のリセットが確定した場合、パスワードの更新を戻して無効なトークンのエラーを返すことを検証する。
func TestUpdatePasswordResetUsecase_Execute_TokenUsedDuringReset(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	f := newPasswordResetRaceFixture(t, ctx)

	// 別のリセットがトークンを使用済みにしたまま未コミットにしておくと、トークンの検証は未使用のトークンを読み、
	// トークンの使用済み化はこの行のロックを待つ。待機を確認してから別のリセットを確定させる。
	blocker, blockerPID := f.beginBlocker(t, ctx)
	if _, err := blocker.ExecContext(ctx, "UPDATE password_reset_tokens SET used_at = NOW() WHERE id = $1", int64(f.resetTokenID)); err != nil {
		t.Fatalf("別のリセットによるトークンの使用済み化に失敗: %v", err)
	}

	output, err := f.executeWhileBlocked(t, ctx, blocker, blockerPID)

	if output != nil || !errors.Is(err, ErrInvalidPasswordResetToken) {
		t.Errorf("トークンが使われた後のリセット = (%+v, %v)、結果なしと無効なトークンのエラーを期待", output, err)
	}
	f.assertPasswordUnchanged(t, ctx)
}

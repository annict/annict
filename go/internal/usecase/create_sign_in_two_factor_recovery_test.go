package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/lib/pq"

	"github.com/annict/annict/go/internal/auth"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/validator"
)

// recoveryTestCodesは、リカバリーコードでのログインのテストで発行したことにするコード。
var recoveryTestCodes = []string{"abcd-2345", "efgh-6789"}

// newSignInTwoFactorRecoveryUsecaseForTestは共有テストDBに対してUseCaseを組み立てる。
// UseCaseが開くトランザクションから見えるよう、テストは準備した行をコミットしておく。
func newSignInTwoFactorRecoveryUsecaseForTest(t *testing.T, db *sql.DB) *CreateSignInTwoFactorRecoveryUsecase {
	t.Helper()

	queries := query.New(db)
	return NewCreateSignInTwoFactorRecoveryUsecase(
		db,
		testutil.NewTestTwoFactorKey(t),
		validator.NewSignInTwoFactorRecoveryCreateValidator(),
		repository.NewUserRepository(queries),
		repository.NewUserTwoFactorAuthRepository(queries),
		repository.NewUserTwoFactorRecoveryCodeRepository(queries),
		NewCreateSessionUsecase(repository.NewSessionRepository(queries)),
	)
}

// createRecoveryTestUserは、二要素認証を有効にしてrecoveryTestCodesを発行したユーザーをコミットして作る。
// enableがfalseのときは、二要素認証の設定を作らない。
func createRecoveryTestUser(t *testing.T, db *sql.DB, enable bool) model.UserID {
	t.Helper()

	userID, _ := createCommittedUserForTest(t, db, false)
	if !enable {
		return userID
	}

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("準備用トランザクションの開始に失敗: %v", err)
	}
	defer func() { _ = tx.Rollback() }()
	testutil.EnableTwoFactorAuthForTest(t, tx, userID)
	key := testutil.NewTestTwoFactorKey(t)
	digests := make([]string, len(recoveryTestCodes))
	for i, code := range recoveryTestCodes {
		digests[i] = key.RecoveryCodeDigest(auth.NormalizeRecoveryCode(code))
	}
	if err := repository.NewUserTwoFactorRecoveryCodeRepository(query.New(db).WithTx(tx)).CreateAll(context.Background(), userID, digests); err != nil {
		t.Fatalf("CreateAll()のエラー = %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("準備用トランザクションのコミットに失敗: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM user_two_factor_recovery_codes WHERE user_id = $1", int64(userID))
		_, _ = db.Exec("DELETE FROM user_two_factor_auths WHERE user_id = $1", int64(userID))
	})
	return userID
}

// TestCreateSignInTwoFactorRecoveryUsecaseは、リカバリーコードが合えば消費してセッションを作り、
// 同じコードを再び使えないことを検証する。入力は大文字・区切りの無い形でも受け付ける。
func TestCreateSignInTwoFactorRecoveryUsecase(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newSignInTwoFactorRecoveryUsecaseForTest(t, db)
	ctx := i18n.SetLocale(context.Background(), "ja")
	userID := createRecoveryTestUser(t, db, true)

	output, err := uc.Execute(ctx, CreateSignInTwoFactorRecoveryInput{UserID: userID, Code: " " + strings.ToUpper(strings.ReplaceAll(recoveryTestCodes[0], "-", "")) + " "})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if output.PublicID == "" || output.UserID != userID {
		t.Errorf("Execute() = %+v、セッションのPublicIDとユーザーを期待", output)
	}

	if _, err := repository.NewSessionRepository(query.New(db)).GetSessionByID(ctx, output.PublicID); err != nil {
		t.Errorf("作ったセッションのGetSessionByID()のエラー = %v、コミットされたセッションを期待", err)
	}

	count, err := repository.NewUserTwoFactorRecoveryCodeRepository(query.New(db)).CountUnused(ctx, userID)
	if err != nil {
		t.Fatalf("CountUnused()のエラー = %v", err)
	}
	if count != len(recoveryTestCodes)-1 {
		t.Errorf("未使用のリカバリーコードの数 = %d、期待値 = %d", count, len(recoveryTestCodes)-1)
	}

	_, err = uc.Execute(ctx, CreateSignInTwoFactorRecoveryInput{UserID: userID, Code: recoveryTestCodes[0]})
	assertCodeFieldError(t, err, "リカバリーコードが正しくないか、すでに使われています")
}

// TestCreateSignInTwoFactorRecoveryUsecase_InvalidCodeは、形の違うコードと一致しないコードを
// コードの欄のValidationErrorで拒み、リカバリーコードを消費しないことを検証する。
func TestCreateSignInTwoFactorRecoveryUsecase_InvalidCode(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newSignInTwoFactorRecoveryUsecaseForTest(t, db)
	ctx := i18n.SetLocale(context.Background(), "ja")
	userID := createRecoveryTestUser(t, db, true)

	tests := []struct {
		name    string
		code    string
		wantMsg string
	}{
		{name: "空", code: "", wantMsg: "入力してください"},
		{name: "認証アプリのコードの形", code: "123456", wantMsg: "リカバリーコード (xxxx-xxxxの形) を入力してください"},
		{name: "一致しない", code: "zzzz-zzzz", wantMsg: "リカバリーコードが正しくないか、すでに使われています"},
	}
	for _, tt := range tests {
		_, err := uc.Execute(ctx, CreateSignInTwoFactorRecoveryInput{UserID: userID, Code: tt.code})
		assertCodeFieldError(t, err, tt.wantMsg)
	}

	count, err := repository.NewUserTwoFactorRecoveryCodeRepository(query.New(db)).CountUnused(ctx, userID)
	if err != nil {
		t.Fatalf("CountUnused()のエラー = %v", err)
	}
	if count != len(recoveryTestCodes) {
		t.Errorf("未使用のリカバリーコードの数 = %d、期待値 = %d (消費しない)", count, len(recoveryTestCodes))
	}
}

// TestCreateSignInTwoFactorRecoveryUsecase_Unavailableは、二要素認証の設定が無い・退会したユーザーに
// AppErrCodeConflictを返し、ログインをやり直させることを検証する。
func TestCreateSignInTwoFactorRecoveryUsecase_Unavailable(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newSignInTwoFactorRecoveryUsecaseForTest(t, db)
	ctx := i18n.SetLocale(context.Background(), "ja")

	noSettingUserID := createRecoveryTestUser(t, db, false)
	deletedUserID := createRecoveryTestUser(t, db, true)
	if _, err := db.ExecContext(ctx, "UPDATE users SET deleted_at = $1 WHERE id = $2", time.Now(), int64(deletedUserID)); err != nil {
		t.Fatalf("ユーザーの論理削除に失敗: %v", err)
	}

	for name, userID := range map[string]model.UserID{"設定が無い": noSettingUserID, "退会した": deletedUserID} {
		if _, err := uc.Execute(ctx, CreateSignInTwoFactorRecoveryInput{UserID: userID, Code: recoveryTestCodes[0]}); !isConflictForTest(err) {
			t.Errorf("%s: Execute()のエラー = %v、AppErrCodeConflictを期待", name, err)
		}
	}
}

// newRecoverySessionFailureDBは、セッションの保存だけが失敗するテスト用の接続を作る。
// 専用スキーマにsessions表を作り、この接続のsearch_pathだけを変えるため、共有のセッション表や他のテストには影響しない。
func newRecoverySessionFailureDB(t *testing.T, db *sql.DB, userID model.UserID) *sql.DB {
	t.Helper()
	ctx := context.Background()
	schemaName := fmt.Sprintf("test_recovery_session_%d", userID)
	schema := pq.QuoteIdentifier(schemaName)
	if _, err := db.ExecContext(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatalf("セッション保存の失敗を検証するスキーマの作成に失敗: %v", err)
	}
	t.Cleanup(func() {
		if _, err := db.ExecContext(ctx, "DROP SCHEMA "+schema+" CASCADE"); err != nil {
			t.Errorf("テスト専用スキーマの削除に失敗: %v", err)
		}
	})
	if _, err := db.ExecContext(ctx, "CREATE TABLE "+schema+".sessions (LIKE public.sessions INCLUDING ALL, CONSTRAINT reject_recovery_session CHECK (false))"); err != nil {
		t.Fatalf("保存を拒否するセッション表の作成に失敗: %v", err)
	}

	dsn := testutil.TestDatabaseURL()
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		u, err := url.Parse(dsn)
		if err != nil {
			t.Fatal("テスト用DBのURLを解析できません")
		}
		params := u.Query()
		params.Set("search_path", schemaName+",public")
		u.RawQuery = params.Encode()
		dsn = u.String()
	} else {
		// lib/pqのキーと値を並べる接続文字列にも、同じsearch_pathを指定する。
		dsn += " search_path=" + schemaName + ",public"
	}
	failureDB, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("保存の失敗を検証する接続の作成に失敗: %v", err)
	}
	t.Cleanup(func() {
		if err := failureDB.Close(); err != nil {
			t.Errorf("テスト専用の接続を閉じられません: %v", err)
		}
	})
	return failureDB
}

// TestCreateSignInTwoFactorRecoveryUsecase_SessionFailureRollsBackは、コード消費後にセッションの保存が
// 失敗してもコードが未使用に戻り、セッションを作らず、同じコードでログインをやり直せることを検証する。
func TestCreateSignInTwoFactorRecoveryUsecase_SessionFailureRollsBack(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	ctx := context.Background()
	userID := createRecoveryTestUser(t, db, true)
	failureDB := newRecoverySessionFailureDB(t, db, userID)
	uc := newSignInTwoFactorRecoveryUsecaseForTest(t, failureDB)

	output, err := uc.Execute(ctx, CreateSignInTwoFactorRecoveryInput{UserID: userID, Code: recoveryTestCodes[0]})
	if output != nil {
		t.Errorf("保存に失敗したExecute()の結果 = %+v、結果なしを期待", output)
	}
	var pgErr *pq.Error
	if !errors.As(err, &pgErr) || pgErr.Code != "23514" || pgErr.Constraint != "reject_recovery_session" {
		t.Fatalf("Execute()のエラー = %v、セッション表のCHECK制約による保存失敗を期待", err)
	}

	count, err := repository.NewUserTwoFactorRecoveryCodeRepository(query.New(db)).CountUnused(ctx, userID)
	if err != nil {
		t.Fatalf("CountUnused()のエラー = %v", err)
	}
	if count != len(recoveryTestCodes) {
		t.Errorf("保存に失敗した後の未使用数 = %d、期待値 = %d", count, len(recoveryTestCodes))
	}
	digest := testutil.NewTestTwoFactorKey(t).RecoveryCodeDigest(auth.NormalizeRecoveryCode(recoveryTestCodes[0]))
	var usedAt sql.NullTime
	if err := db.QueryRowContext(ctx, "SELECT used_at FROM user_two_factor_recovery_codes WHERE user_id = $1 AND code_digest = $2", int64(userID), digest).Scan(&usedAt); err != nil {
		t.Fatalf("使おうとしたコードの状態の取得に失敗: %v", err)
	}
	if usedAt.Valid {
		t.Error("セッションを保存できなかったのにコードが使用済みになった")
	}
	for name, sessionDB := range map[string]*sql.DB{"共有の表": db, "テスト専用の表": failureDB} {
		var sessions int
		if err := sessionDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM sessions WHERE data->'warden.user.user.key'->0->>0 = $1`, userID.String()).Scan(&sessions); err != nil {
			t.Fatalf("%sのセッション数の取得に失敗: %v", name, err)
		}
		if sessions != 0 {
			t.Errorf("保存に失敗した後の%sのセッション数 = %d、期待値 = 0", name, sessions)
		}
	}

	// 保存ができる接続から同じコードを再送すると、ログインできる。
	output, err = newSignInTwoFactorRecoveryUsecaseForTest(t, db).Execute(ctx, CreateSignInTwoFactorRecoveryInput{UserID: userID, Code: recoveryTestCodes[0]})
	if err != nil || output == nil || output.PublicID == "" {
		t.Fatalf("同じコードでの再試行 = (%+v, %v)、ログインの成功を期待", output, err)
	}
	t.Cleanup(func() {
		if err := repository.NewSessionRepository(query.New(db)).DeleteSession(ctx, output.PublicID); err != nil {
			t.Errorf("再試行で作ったセッションの削除に失敗: %v", err)
		}
	})
}

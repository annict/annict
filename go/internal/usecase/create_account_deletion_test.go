package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverdatabasesql"
	"github.com/riverqueue/river/rivertype"

	"github.com/annict/annict/go/internal/dispatcher"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/validator"
)

// failingTxJobInserterはジョブの投入に失敗するTxJobInserter
type failingTxJobInserter struct{}

func (failingTxJobInserter) InsertTx(context.Context, *sql.Tx, river.JobArgs, *river.InsertOpts) (*rivertype.JobInsertResult, error) {
	return nil, errors.New("投入に失敗")
}

// failingAccountDeletionEmailTxJobInserterは削除ジョブを投入し、退会完了メールのジョブだけ失敗させる。
type failingAccountDeletionEmailTxJobInserter struct {
	inserter dispatcher.TxJobInserter
}

func (f failingAccountDeletionEmailTxJobInserter) InsertTx(ctx context.Context, tx *sql.Tx, args river.JobArgs, opts *river.InsertOpts) (*rivertype.JobInsertResult, error) {
	if args.Kind() == (dispatcher.SendAccountDeletionEmailArgs{}).Kind() {
		return nil, errors.New("退会完了メールのジョブ投入に失敗")
	}
	return f.inserter.InsertTx(ctx, tx, args, opts)
}

// newCreateAccountDeletionUsecaseForTestは共有テストDBに対してCreateAccountDeletionUsecaseを組み立てる。
// UseCaseが開くトランザクションから見えるよう、テストは準備した行をコミットしておく。
func newCreateAccountDeletionUsecaseForTest(t *testing.T, db *sql.DB, txInserter dispatcher.TxJobInserter) *CreateAccountDeletionUsecase {
	t.Helper()

	queries := query.New(db)
	return NewCreateAccountDeletionUsecase(
		db,
		repository.NewUserRepository(queries),
		repository.NewStripeSubscriberRepository(queries),
		dispatcher.NewDispatcher(nil, txInserter),
		validator.NewSettingsAccountDeletionCreateValidator(repository.NewOAuthApplicationRepository(queries)),
	)
}

func newRiverTxInserterForTest(t *testing.T, db *sql.DB) *river.Client[*sql.Tx] {
	t.Helper()

	client, err := river.NewClient(riverdatabasesql.New(db), &river.Config{})
	if err != nil {
		t.Fatalf("Riverクライアントの作成に失敗: %v", err)
	}
	return client
}

// createCommittedUserForTestはユーザーを作成してコミットし、テストの終了時にユーザーとジョブを消す。
func createCommittedUserForTest(t *testing.T, db *sql.DB, withOAuthApplication bool) (model.UserID, string) {
	t.Helper()

	return createCommittedUserWithBuilderForTest(t, db, func(tx *sql.Tx) *testutil.UserBuilder {
		return testutil.NewUserBuilder(t, tx)
	}, withOAuthApplication)
}

// createCommittedUserWithBuilderForTestはnewBuilderが返すビルダーでユーザーを作成してコミットし、
// テストの終了時にユーザー・ジョブ・Stripeの購読情報を消す。
func createCommittedUserWithBuilderForTest(t *testing.T, db *sql.DB, newBuilder func(tx *sql.Tx) *testutil.UserBuilder, withOAuthApplication bool) (model.UserID, string) {
	t.Helper()

	setupTx, err := db.Begin()
	if err != nil {
		t.Fatalf("準備用トランザクションの開始に失敗: %v", err)
	}
	defer func() { _ = setupTx.Rollback() }()

	userID := newBuilder(setupTx).Build()
	if withOAuthApplication {
		testutil.CreateTestOAuthApplication(t, setupTx, userID, false)
	}
	var username, email string
	var stripeSubscriberID sql.NullInt64
	if err := setupTx.QueryRow("SELECT username, email, stripe_subscriber_id FROM users WHERE id = $1", int64(userID)).Scan(&username, &email, &stripeSubscriberID); err != nil {
		t.Fatalf("ユーザーの取得に失敗: %v", err)
	}
	if err := setupTx.Commit(); err != nil {
		t.Fatalf("準備用トランザクションのコミットに失敗: %v", err)
	}

	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM river_job WHERE kind = 'delete_user' AND (args->>'user_id')::bigint = $1", int64(userID))
		_, _ = db.Exec("DELETE FROM river_job WHERE kind = 'send_account_deletion_email' AND args->>'email' = $1", email)
		_, _ = db.Exec("DELETE FROM oauth_applications WHERE owner_id = $1", int64(userID))
		_, _ = db.Exec("DELETE FROM users WHERE id = $1", int64(userID))
		if stripeSubscriberID.Valid {
			_, _ = db.Exec("DELETE FROM stripe_subscribers WHERE id = $1", stripeSubscriberID.Int64)
		}
	})
	return userID, username
}

// getAccountDeletionEmailJobsForTestは、送信先がemailの退会完了メールのジョブの引数を返す。
func getAccountDeletionEmailJobsForTest(t *testing.T, db *sql.DB, email string) []dispatcher.SendAccountDeletionEmailArgs {
	t.Helper()

	rows, err := db.Query("SELECT args FROM river_job WHERE kind = 'send_account_deletion_email' AND args->>'email' = $1", email)
	if err != nil {
		t.Fatalf("退会完了メールのジョブの取得に失敗: %v", err)
	}
	defer func() { _ = rows.Close() }()

	var jobs []dispatcher.SendAccountDeletionEmailArgs
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			t.Fatalf("ジョブの引数の読み取りに失敗: %v", err)
		}
		var args dispatcher.SendAccountDeletionEmailArgs
		if err := json.Unmarshal(raw, &args); err != nil {
			t.Fatalf("ジョブの引数のデコードに失敗: %v", err)
		}
		jobs = append(jobs, args)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("退会完了メールのジョブの取得に失敗: %v", err)
	}
	return jobs
}

// getEmailForTestはユーザーの現在のメールアドレスを返す。
func getEmailForTest(t *testing.T, db *sql.DB, userID model.UserID) string {
	t.Helper()

	var email string
	if err := db.QueryRow("SELECT email FROM users WHERE id = $1", int64(userID)).Scan(&email); err != nil {
		t.Fatalf("メールアドレスの取得に失敗: %v", err)
	}
	return email
}

type userDeletionStateForTest struct {
	username   string
	email      string
	deleted    bool
	jobsQueued int
}

func getUserDeletionStateForTest(t *testing.T, db *sql.DB, userID model.UserID) userDeletionStateForTest {
	t.Helper()

	var state userDeletionStateForTest
	var deletedAt sql.NullTime
	if err := db.QueryRow("SELECT username, email, deleted_at FROM users WHERE id = $1", int64(userID)).Scan(&state.username, &state.email, &deletedAt); err != nil {
		t.Fatalf("ユーザーの取得に失敗: %v", err)
	}
	state.deleted = deletedAt.Valid
	if err := db.QueryRow("SELECT COUNT(*) FROM river_job WHERE kind = 'delete_user' AND (args->>'user_id')::bigint = $1", int64(userID)).Scan(&state.jobsQueued); err != nil {
		t.Fatalf("削除ジョブの件数の取得に失敗: %v", err)
	}
	return state
}

var anonymizedUsernamePattern = regexp.MustCompile(`^[0-9a-f]{8}_[0-9a-f]{4}_[0-9a-f]{4}_[0-9a-f]{4}_[0-9a-f]{12}$`)

// TestCreateAccountDeletionUsecase_Executeは、ユーザー名とメールアドレスを匿名化して論理削除し、
// 削除ジョブと匿名化前のメールアドレス宛ての退会完了メールのジョブを同じトランザクションで投入することを検証する。
func TestCreateAccountDeletionUsecase_Execute(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	userID, username := createCommittedUserForTest(t, db, false)
	originalEmail := getEmailForTest(t, db, userID)
	uc := newCreateAccountDeletionUsecaseForTest(t, db, newRiverTxInserterForTest(t, db))

	ctx := i18n.SetLocale(context.Background(), "ja")
	if err := uc.Execute(ctx, CreateAccountDeletionInput{
		User:     &model.User{ID: userID, Username: username},
		Username: username,
	}); err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	state := getUserDeletionStateForTest(t, db, userID)
	if !anonymizedUsernamePattern.MatchString(state.username) {
		t.Errorf("username = %q、UUIDのハイフンをアンダースコアにした値を期待", state.username)
	}
	if state.email != state.username+"@example.com" {
		t.Errorf("email = %q、期待値 = %q", state.email, state.username+"@example.com")
	}
	if !state.deleted {
		t.Error("deleted_atが立っていない")
	}
	if state.jobsQueued != 1 {
		t.Errorf("削除ジョブの件数 = %d、期待値 = 1", state.jobsQueued)
	}

	emailJobs := getAccountDeletionEmailJobsForTest(t, db, originalEmail)
	want := dispatcher.SendAccountDeletionEmailArgs{Email: originalEmail, Locale: "ja", IsStripeSupporter: false}
	if len(emailJobs) != 1 || emailJobs[0] != want {
		t.Errorf("退会完了メールのジョブ = %+v、期待値 = [%+v]", emailJobs, want)
	}
}

// TestCreateAccountDeletionUsecase_Execute_Emailは、退会完了メールのジョブにユーザーのロケールと、
// 有効なStripeのサブスクリプションがあるかを渡すことを検証する。
func TestCreateAccountDeletionUsecase_Execute_Email(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                  string
		locale                string
		stripeStatus          string
		wantIsStripeSupporter bool
	}{
		{name: "サポーターでない", locale: "en"},
		{name: "有効なStripeのサブスクリプションがある", locale: "ja", stripeStatus: "active", wantIsStripeSupporter: true},
		{name: "支払い遅延中のStripeのサブスクリプションがある", locale: "ja", stripeStatus: "past_due", wantIsStripeSupporter: true},
		{name: "解約済みのStripeのサブスクリプションがある", locale: "ja", stripeStatus: "canceled", wantIsStripeSupporter: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := testutil.GetTestDB()
			userID, username := createCommittedUserWithBuilderForTest(t, db, func(tx *sql.Tx) *testutil.UserBuilder {
				builder := testutil.NewUserBuilder(t, tx).WithLocale(tt.locale)
				if tt.stripeStatus != "" {
					stripeSubscriberID := testutil.CreateTestStripeSubscriberWithStatus(t, tx, tt.stripeStatus)
					builder = builder.WithStripeSubscriberID(&stripeSubscriberID)
				}
				return builder
			}, false)
			originalEmail := getEmailForTest(t, db, userID)
			uc := newCreateAccountDeletionUsecaseForTest(t, db, newRiverTxInserterForTest(t, db))

			if err := uc.Execute(i18n.SetLocale(context.Background(), "ja"), CreateAccountDeletionInput{
				User:     &model.User{ID: userID, Username: username},
				Username: username,
			}); err != nil {
				t.Fatalf("Execute()のエラー = %v", err)
			}

			emailJobs := getAccountDeletionEmailJobsForTest(t, db, originalEmail)
			want := dispatcher.SendAccountDeletionEmailArgs{Email: originalEmail, Locale: tt.locale, IsStripeSupporter: tt.wantIsStripeSupporter}
			if len(emailJobs) != 1 || emailJobs[0] != want {
				t.Errorf("退会完了メールのジョブ = %+v、期待値 = [%+v]", emailJobs, want)
			}
		})
	}
}

// TestCreateAccountDeletionUsecase_Execute_UsernameChangedは、退会リクエストの認証後に
// ユーザー名が変わった場合、古いユーザー名で退会を確定しないことを検証する。
func TestCreateAccountDeletionUsecase_Execute_UsernameChanged(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	userID, oldUsername := createCommittedUserForTest(t, db, false)
	changedUsername := oldUsername + "_changed"
	if _, err := db.Exec("UPDATE users SET username = $1 WHERE id = $2", changedUsername, int64(userID)); err != nil {
		t.Fatalf("ユーザー名の変更に失敗: %v", err)
	}
	uc := newCreateAccountDeletionUsecaseForTest(t, db, newRiverTxInserterForTest(t, db))
	err := uc.Execute(i18n.SetLocale(context.Background(), "ja"), CreateAccountDeletionInput{
		User:     &model.User{ID: userID, Username: oldUsername},
		Username: oldUsername,
	})
	ve := model.AsValidationError(err)
	if ve == nil || !ve.HasFieldError("username") {
		t.Fatalf("ユーザー名のバリデーションエラーを期待したが、%vだった", err)
	}
	state := getUserDeletionStateForTest(t, db, userID)
	if state.deleted || state.username != changedUsername || state.jobsQueued != 0 {
		t.Errorf("ユーザー名の変更後に退会が進んだ: %+v", state)
	}
}

// TestCreateAccountDeletionUsecase_Execute_OAuthApplicationCreatedConcurrentlyは、
// アプリケーションの作成がユーザー行のロックを先に取得した場合、退会がその確定を待ち、
// 有効なアプリケーションを検出して受け付けないことを検証する。
func TestCreateAccountDeletionUsecase_Execute_OAuthApplicationCreatedConcurrently(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	userID, username := createCommittedUserForTest(t, db, false)
	createTx, err := db.Begin()
	if err != nil {
		t.Fatalf("作成用トランザクションの開始に失敗: %v", err)
	}
	defer func() { _ = createTx.Rollback() }()
	if _, err := createTx.Exec("SELECT id FROM users WHERE id = $1 FOR UPDATE", int64(userID)); err != nil {
		t.Fatalf("ユーザー行のロックに失敗: %v", err)
	}

	uc := newCreateAccountDeletionUsecaseForTest(t, db, newRiverTxInserterForTest(t, db))
	result := make(chan error, 1)
	go func() {
		result <- uc.Execute(i18n.SetLocale(context.Background(), "ja"), CreateAccountDeletionInput{
			User:     &model.User{ID: userID, Username: username},
			Username: username,
		})
	}()
	select {
	case err := <-result:
		t.Fatalf("作成側のロックがあるのに退会が終了した: %v", err)
	case <-time.After(50 * time.Millisecond):
	}

	testutil.CreateTestOAuthApplication(t, createTx, userID, false)
	if err := createTx.Commit(); err != nil {
		t.Fatalf("アプリケーションの作成のコミットに失敗: %v", err)
	}
	select {
	case err := <-result:
		if model.AsValidationError(err) == nil {
			t.Fatalf("有効なアプリケーションのバリデーションエラーを期待したが、%vだった", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("退会の受付が終了しなかった")
	}
	state := getUserDeletionStateForTest(t, db, userID)
	if state.deleted || state.jobsQueued != 0 {
		t.Errorf("OAuthアプリケーションがあるのに退会が進んだ: %+v", state)
	}
}

// TestCreateAccountDeletionUsecase_Execute_AlreadyDeletedは、同時に送られた受付のように
// すでに論理削除されたユーザーでは、何も変えずに成功として扱うことを検証する。
func TestCreateAccountDeletionUsecase_Execute_AlreadyDeleted(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	userID, username := createCommittedUserForTest(t, db, false)
	if _, err := db.Exec("UPDATE users SET deleted_at = NOW() WHERE id = $1", int64(userID)); err != nil {
		t.Fatalf("ユーザーの論理削除に失敗: %v", err)
	}
	uc := newCreateAccountDeletionUsecaseForTest(t, db, newRiverTxInserterForTest(t, db))

	ctx := i18n.SetLocale(context.Background(), "ja")
	if err := uc.Execute(ctx, CreateAccountDeletionInput{
		User:     &model.User{ID: userID, Username: username},
		Username: username,
	}); err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	state := getUserDeletionStateForTest(t, db, userID)
	if state.username != username {
		t.Errorf("username = %q、期待値 = %q", state.username, username)
	}
	if state.jobsQueued != 0 {
		t.Errorf("削除ジョブの件数 = %d、期待値 = 0", state.jobsQueued)
	}
	if emailJobs := getAccountDeletionEmailJobsForTest(t, db, state.email); len(emailJobs) != 0 {
		t.Errorf("退会完了メールのジョブの件数 = %d、期待値 = 0", len(emailJobs))
	}
}

// TestCreateAccountDeletionUsecase_Execute_ValidationErrorは、受け付けない場合に
// ユーザーを変更せず、削除ジョブも投入しないことを検証する。
func TestCreateAccountDeletionUsecase_Execute_ValidationError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		withOAuthApplication bool
		inputUsername        func(username string) string
	}{
		{
			name:          "ユーザー名が一致しない",
			inputUsername: func(string) string { return "wrong_username" },
		},
		{
			name:                 "OAuthアプリケーションが残っている",
			withOAuthApplication: true,
			inputUsername:        func(username string) string { return username },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := testutil.GetTestDB()
			userID, username := createCommittedUserForTest(t, db, tt.withOAuthApplication)
			uc := newCreateAccountDeletionUsecaseForTest(t, db, newRiverTxInserterForTest(t, db))

			ctx := i18n.SetLocale(context.Background(), "ja")
			err := uc.Execute(ctx, CreateAccountDeletionInput{
				User:     &model.User{ID: userID, Username: username},
				Username: tt.inputUsername(username),
			})
			if model.AsValidationError(err) == nil {
				t.Fatalf("ValidationErrorを期待したが、%vだった", err)
			}

			state := getUserDeletionStateForTest(t, db, userID)
			if state.deleted {
				t.Error("受け付けなかったユーザーのdeleted_atが立っている")
			}
			if state.username != username {
				t.Errorf("username = %q、期待値 = %q", state.username, username)
			}
			if state.jobsQueued != 0 {
				t.Errorf("削除ジョブの件数 = %d、期待値 = 0", state.jobsQueued)
			}
			if emailJobs := getAccountDeletionEmailJobsForTest(t, db, state.email); len(emailJobs) != 0 {
				t.Errorf("退会完了メールのジョブの件数 = %d、期待値 = 0", len(emailJobs))
			}
		})
	}
}

// TestCreateAccountDeletionUsecase_Execute_EnqueueFailureは、削除ジョブを投入できなければ
// 論理削除もロールバックすることを検証する。
func TestCreateAccountDeletionUsecase_Execute_EnqueueFailure(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	userID, username := createCommittedUserForTest(t, db, false)
	uc := newCreateAccountDeletionUsecaseForTest(t, db, failingTxJobInserter{})

	ctx := i18n.SetLocale(context.Background(), "ja")
	err := uc.Execute(ctx, CreateAccountDeletionInput{
		User:     &model.User{ID: userID, Username: username},
		Username: username,
	})
	if err == nil {
		t.Fatal("エラーを期待したが、nilだった")
	}

	state := getUserDeletionStateForTest(t, db, userID)
	if state.deleted {
		t.Error("削除ジョブを投入できなかったユーザーのdeleted_atが立っている")
	}
	if state.username != username {
		t.Errorf("username = %q、期待値 = %q", state.username, username)
	}
}

// TestCreateAccountDeletionUsecase_Execute_EmailEnqueueFailureは、削除ジョブの投入後に
// 退会完了メールのジョブ投入が失敗すると、ユーザーの匿名化と削除ジョブをロールバックすることを検証する。
func TestCreateAccountDeletionUsecase_Execute_EmailEnqueueFailure(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	userID, username := createCommittedUserForTest(t, db, false)
	originalEmail := getEmailForTest(t, db, userID)
	uc := newCreateAccountDeletionUsecaseForTest(t, db, failingAccountDeletionEmailTxJobInserter{
		inserter: newRiverTxInserterForTest(t, db),
	})

	err := uc.Execute(i18n.SetLocale(context.Background(), "ja"), CreateAccountDeletionInput{
		User:     &model.User{ID: userID, Username: username},
		Username: username,
	})
	if err == nil {
		t.Fatal("エラーを期待したが、nilだった")
	}

	state := getUserDeletionStateForTest(t, db, userID)
	if state.deleted {
		t.Error("メールジョブを投入できなかったユーザーのdeleted_atが立っている")
	}
	if state.username != username {
		t.Errorf("username = %q、期待値 = %q", state.username, username)
	}
	if state.email != originalEmail {
		t.Errorf("email = %q、期待値 = %q", state.email, originalEmail)
	}
	if state.jobsQueued != 0 {
		t.Errorf("削除ジョブの件数 = %d、期待値 = 0", state.jobsQueued)
	}
	if emailJobs := getAccountDeletionEmailJobsForTest(t, db, originalEmail); len(emailJobs) != 0 {
		t.Errorf("退会完了メールのジョブの件数 = %d、期待値 = 0", len(emailJobs))
	}
}

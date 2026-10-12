package usecase

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	annictstripe "github.com/annict/annict/go/internal/stripe"
	"github.com/annict/annict/go/internal/testutil"
)

// subscriptionCancelerStubはCancelSubscriptionの呼び出しを記録し、指定した結果を返す。
type subscriptionCancelerStub struct {
	calledWith []string
	result     annictstripe.CancelSubscriptionResult
	err        error
}

func (s *subscriptionCancelerStub) CancelSubscription(_ context.Context, subscriptionID string) (annictstripe.CancelSubscriptionResult, error) {
	s.calledWith = append(s.calledWith, subscriptionID)
	return s.result, s.err
}

// objectDeleterStubはDeleteObjectsの呼び出しを記録し、指定したエラーを返す。
type objectDeleterStub struct {
	calledWith [][]string
	err        error
}

func (s *objectDeleterStub) DeleteObjects(_ context.Context, keys []string) error {
	s.calledWith = append(s.calledWith, keys)
	return s.err
}

// newDeleteUserUsecaseはqueriesを使うリポジトリでDeleteUserUsecaseを組み立てる。
func newDeleteUserUsecase(db *sql.DB, queries *query.Queries, canceler SubscriptionCanceler, timeLimit time.Duration) *DeleteUserUsecase {
	return NewDeleteUserUsecase(
		db,
		repository.NewUserRepository(queries),
		repository.NewStripeSubscriberRepository(queries),
		repository.NewActivityRepository(queries),
		repository.NewActivityGroupRepository(queries),
		repository.NewLibraryEntryRepository(queries),
		repository.NewStatusRepository(queries),
		repository.NewCommentRepository(queries),
		repository.NewEpisodeRecordRepository(queries),
		repository.NewWorkRecordRepository(queries),
		repository.NewMultipleEpisodeRecordRepository(queries),
		repository.NewRecordRepository(queries),
		repository.NewNotificationRepository(queries),
		repository.NewLikeRepository(queries),
		repository.NewFollowRepository(queries),
		repository.NewMuteUserRepository(queries),
		repository.NewCharacterFavoriteRepository(queries),
		repository.NewPersonFavoriteRepository(queries),
		repository.NewOrganizationFavoriteRepository(queries),
		repository.NewWorkTaggingRepository(queries),
		repository.NewWorkTaggableRepository(queries),
		repository.NewCollectionItemRepository(queries),
		repository.NewCollectionRepository(queries),
		repository.NewForumPostParticipantRepository(queries),
		repository.NewForumCommentRepository(queries),
		repository.NewForumPostRepository(queries),
		repository.NewDbActivityRepository(queries),
		repository.NewDbCommentRepository(queries),
		repository.NewWorkCommentRepository(queries),
		repository.NewUserlandProjectMemberRepository(queries),
		repository.NewOAuthAccessTokenRepository(queries),
		repository.NewOAuthAccessGrantRepository(queries),
		repository.NewOAuthApplicationRepository(queries),
		repository.NewProfileRepository(queries),
		canceler,
		&objectDeleterStub{},
		timeLimit,
	)
}

// newDeleteUserUsecaseForTestはテスト用のトランザクション上のリポジトリでDeleteUserUsecaseを組み立てる。
// ユーザー行の削除段階は自前のトランザクションを開くため、このトランザクションで準備した行は見えず、
// ユーザー行は削除されない。ユーザー行の削除まで確かめるテストはnewCommittedDeleteUserUsecaseForTestを使う。
func newDeleteUserUsecaseForTest(t *testing.T, canceler SubscriptionCanceler, timeLimit time.Duration) (*DeleteUserUsecase, *sql.Tx) {
	t.Helper()
	db, tx := testutil.SetupTx(t)
	return newDeleteUserUsecase(db, query.New(db).WithTx(tx), canceler, timeLimit), tx
}

// newCommittedDeleteUserUsecaseForTestは共有テストDBに対してDeleteUserUsecaseを組み立て、行の準備に使う
// トランザクションを返す。ユーザー行の削除段階が開くトランザクションから準備した行が見えるよう、
// テストは準備用のトランザクションをコミットしてからExecuteを呼ぶ。
func newCommittedDeleteUserUsecaseForTest(t *testing.T, canceler SubscriptionCanceler, timeLimit time.Duration) (*DeleteUserUsecase, *sql.DB, *sql.Tx) {
	t.Helper()
	db := testutil.GetTestDB()
	setupTx, err := db.Begin()
	if err != nil {
		t.Fatalf("準備用トランザクションの開始に失敗: %v", err)
	}
	t.Cleanup(func() { _ = setupTx.Rollback() })
	return newDeleteUserUsecase(db, query.New(db), canceler, timeLimit), db, setupTx
}

// beginReadTxForTestは、コミット済みの行を確かめるための読み取り用トランザクションを開く。
func beginReadTxForTest(t *testing.T, db *sql.DB) *sql.Tx {
	t.Helper()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("確認用トランザクションの開始に失敗: %v", err)
	}
	t.Cleanup(func() { _ = tx.Rollback() })
	return tx
}

func TestDeleteUserUsecase_Execute_CancelStripeSubscription(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		result annictstripe.CancelSubscriptionResult
	}{
		{name: "正常系: 解約した", result: annictstripe.CancelSubscriptionResultCanceled},
		{name: "正常系: 解約済みだった", result: annictstripe.CancelSubscriptionResultAlreadyEnded},
		{name: "正常系: Stripe上に存在しなかった", result: annictstripe.CancelSubscriptionResultNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			canceler := &subscriptionCancelerStub{result: tt.result}
			uc, tx := newDeleteUserUsecaseForTest(t, canceler, DeleteUserTimeLimit)

			subscriber := testutil.NewStripeSubscriberBuilder(t, tx).BuildWithResult()
			userID := testutil.NewUserBuilder(t, tx).
				WithStripeSubscriberID(&subscriber.ID).
				WithDeletedAt(time.Now()).
				Build()

			result, err := uc.Execute(context.Background(), DeleteUserInput{UserID: userID})
			if err != nil {
				t.Fatalf("Execute()のエラー = %v", err)
			}
			if !result.Completed {
				t.Error("Completed = false、期待値 = true")
			}
			if len(canceler.calledWith) != 1 || canceler.calledWith[0] != subscriber.StripeSubscriptionID {
				t.Errorf("CancelSubscription()の呼び出し = %v、期待値 = [%s]", canceler.calledWith, subscriber.StripeSubscriptionID)
			}
		})
	}
}

func TestDeleteUserUsecase_Execute_WithoutStripeSubscriber(t *testing.T) {
	t.Parallel()

	canceler := &subscriptionCancelerStub{}
	uc, tx := newDeleteUserUsecaseForTest(t, canceler, DeleteUserTimeLimit)

	userID := testutil.NewUserBuilder(t, tx).WithDeletedAt(time.Now()).Build()

	result, err := uc.Execute(context.Background(), DeleteUserInput{UserID: userID})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if !result.Completed {
		t.Error("Completed = false、期待値 = true")
	}
	if len(canceler.calledWith) != 0 {
		t.Errorf("CancelSubscription()の呼び出し = %v、期待値 = 呼び出しなし", canceler.calledWith)
	}
}

func TestDeleteUserUsecase_Execute_CancelFailure(t *testing.T) {
	t.Parallel()

	stripeErr := errors.New("Stripe APIの一時的な障害")
	canceler := &subscriptionCancelerStub{err: stripeErr}
	uc, tx := newDeleteUserUsecaseForTest(t, canceler, DeleteUserTimeLimit)

	// 解約できない間は後続の段階に進まないことを確かめるため、解約の後ろに段階を足す。
	nextStageCalled := false
	uc.stages = append(uc.stages, deleteUserStage{
		name: "next",
		run: func(context.Context, *model.UserDeletionTarget) (bool, error) {
			nextStageCalled = true
			return true, nil
		},
	})

	subscriberID := testutil.NewStripeSubscriberBuilder(t, tx).Build()
	userID := testutil.NewUserBuilder(t, tx).
		WithStripeSubscriberID(&subscriberID).
		WithDeletedAt(time.Now()).
		Build()

	for _, finalAttempt := range []bool{false, true} {
		result, err := uc.Execute(context.Background(), DeleteUserInput{UserID: userID, FinalAttempt: finalAttempt})
		if !errors.Is(err, stripeErr) {
			t.Errorf("FinalAttempt = %v のときのExecute()のエラー = %v、期待値 = %v", finalAttempt, err, stripeErr)
		}
		if result != nil {
			t.Errorf("FinalAttempt = %v のときの結果 = %+v、期待値 = nil", finalAttempt, result)
		}
	}
	if nextStageCalled {
		t.Error("解約に失敗したのに後続の段階が実行された")
	}
}

func TestDeleteUserUsecase_Execute_FinalAttemptLog(t *testing.T) {
	// slogの既定ロガーを差し替えるため、このテストは並列化しない。
	previousLogger := slog.Default()
	t.Cleanup(func() { slog.SetDefault(previousLogger) })

	uc, tx := newDeleteUserUsecaseForTest(t, &subscriptionCancelerStub{}, 0)
	userID := testutil.NewUserBuilder(t, tx).WithDeletedAt(time.Now()).Build()

	var logOutput bytes.Buffer
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logOutput, nil)))
	checkLog := func(t *testing.T, wantCount int, stage string) {
		t.Helper()
		got := logOutput.String()
		if count := strings.Count(got, "CLIから再実行してください"); count != wantCount {
			t.Errorf("再実行案内のログ件数 = %d、期待値 = %d: %s", count, wantCount, got)
		}
		if wantCount > 0 && !strings.Contains(got, `"stage":"`+stage+`"`) {
			t.Errorf("ログに段階 %q がありません: %s", stage, got)
		}
		logOutput.Reset()
	}

	// 対象の取得を失敗させるため、期限切れのcontextを渡す (キャンセルと違い、タイムアウトは失敗として返る)。
	expiredCtx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	if _, err := uc.Execute(expiredCtx, DeleteUserInput{UserID: userID}); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("対象取得失敗時のエラー = %v、期待値 = context.DeadlineExceeded", err)
	}
	checkLog(t, 0, "")

	if _, err := uc.Execute(expiredCtx, DeleteUserInput{UserID: userID, FinalAttempt: true}); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("最終試行の対象取得失敗時のエラー = %v、期待値 = context.DeadlineExceeded", err)
	}
	checkLog(t, 1, "find_deletion_target")

	// 論理削除されていないユーザーはWorkerがジョブを取り消すため、再実行の案内は出さない。
	activeUserID := testutil.NewUserBuilder(t, tx).Build()
	if _, err := uc.Execute(context.Background(), DeleteUserInput{UserID: activeUserID, FinalAttempt: true}); !errors.Is(err, ErrUserNotDeleted) {
		t.Fatalf("最終試行の論理削除チェック失敗時のエラー = %v、期待値 = %v", err, ErrUserNotDeleted)
	}
	checkLog(t, 0, "")

	// キャンセルによる中断は失敗ではないため、最終試行でも再実行の案内は出さない。
	canceledCtx, cancelStage := context.WithCancel(context.Background())
	uc.stages = []deleteUserStage{{
		name: "canceled_stage",
		run: func(ctx context.Context, _ *model.UserDeletionTarget) (bool, error) {
			cancelStage()
			return false, ctx.Err()
		},
	}}
	if result, err := uc.Execute(canceledCtx, DeleteUserInput{UserID: userID, FinalAttempt: true}); err != nil || result.Completed {
		t.Fatalf("最終試行のキャンセル時の結果 = %+v、エラー = %v、期待値 = 未完了でエラーなし", result, err)
	}
	checkLog(t, 0, "")

	stageErr := errors.New("段階の処理に失敗")
	uc.stages = []deleteUserStage{{
		name: "test_stage",
		run: func(context.Context, *model.UserDeletionTarget) (bool, error) {
			return false, stageErr
		},
	}}
	if _, err := uc.Execute(context.Background(), DeleteUserInput{UserID: userID, FinalAttempt: true}); !errors.Is(err, stageErr) {
		t.Fatalf("最終試行の段階処理失敗時のエラー = %v、期待値 = %v", err, stageErr)
	}
	checkLog(t, 1, "test_stage")
}

func TestDeleteUserUsecase_Execute_StageProgressLog(t *testing.T) {
	// slogの既定ロガーを差し替えるため、このテストは並列化しない。
	previousLogger := slog.Default()
	t.Cleanup(func() { slog.SetDefault(previousLogger) })

	uc, tx := newDeleteUserUsecaseForTest(t, &subscriptionCancelerStub{}, 0)
	userID := testutil.NewUserBuilder(t, tx).WithDeletedAt(time.Now()).Build()

	// 1つ目の段階は2単位かかり、終えたときだけ進捗を出すことを確かめる。
	firstStageRuns := 0
	uc.stages = []deleteUserStage{
		{
			name: "first_stage",
			run: func(context.Context, *model.UserDeletionTarget) (bool, error) {
				firstStageRuns++
				return firstStageRuns >= 2, nil
			},
		},
		{
			name: "second_stage",
			run: func(context.Context, *model.UserDeletionTarget) (bool, error) {
				return true, nil
			},
		},
	}

	var logOutput bytes.Buffer
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logOutput, nil)))

	result, err := uc.Execute(context.Background(), DeleteUserInput{UserID: userID})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if !result.Completed {
		t.Error("Completed = false、期待値 = true")
	}

	var stages []string
	for line := range strings.SplitSeq(strings.TrimSpace(logOutput.String()), "\n") {
		if !strings.Contains(line, "ユーザーの削除の段階を終えました") {
			continue
		}
		if !strings.Contains(line, fmt.Sprintf(`"user_id":%d`, userID)) {
			t.Errorf("進捗のログにユーザーIDがありません: %s", line)
		}
		for _, name := range []string{"first_stage", "second_stage"} {
			if strings.Contains(line, `"stage":"`+name+`"`) {
				stages = append(stages, name)
			}
		}
	}
	if want := []string{"first_stage", "second_stage"}; strings.Join(stages, ",") != strings.Join(want, ",") {
		t.Errorf("進捗のログに出た段階 = %v、期待値 = %v: %s", stages, want, logOutput.String())
	}
}

func TestDeleteUserUsecase_Execute_NotDeletedUser(t *testing.T) {
	t.Parallel()

	canceler := &subscriptionCancelerStub{}
	uc, tx := newDeleteUserUsecaseForTest(t, canceler, DeleteUserTimeLimit)

	subscriberID := testutil.NewStripeSubscriberBuilder(t, tx).Build()
	userID := testutil.NewUserBuilder(t, tx).WithStripeSubscriberID(&subscriberID).Build()

	result, err := uc.Execute(context.Background(), DeleteUserInput{UserID: userID})
	if !errors.Is(err, ErrUserNotDeleted) {
		t.Fatalf("Execute()のエラー = %v、期待値 = %v", err, ErrUserNotDeleted)
	}
	if result != nil {
		t.Errorf("結果 = %+v、期待値 = nil", result)
	}
	if len(canceler.calledWith) != 0 {
		t.Errorf("CancelSubscription()の呼び出し = %v、期待値 = 呼び出しなし", canceler.calledWith)
	}
}

func TestDeleteUserUsecase_Execute_UserNotFound(t *testing.T) {
	t.Parallel()

	canceler := &subscriptionCancelerStub{}
	uc, _ := newDeleteUserUsecaseForTest(t, canceler, DeleteUserTimeLimit)

	result, err := uc.Execute(context.Background(), DeleteUserInput{UserID: model.UserID(-1)})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if !result.Completed {
		t.Error("Completed = false、期待値 = true (削除を終えたユーザーとして扱う)")
	}
}

func TestDeleteUserUsecase_Execute_TimeLimit(t *testing.T) {
	t.Parallel()

	t.Run("上限に達したら段階の途中で中断する", func(t *testing.T) {
		t.Parallel()

		uc, tx := newDeleteUserUsecaseForTest(t, &subscriptionCancelerStub{}, time.Nanosecond)
		userID := testutil.NewUserBuilder(t, tx).WithDeletedAt(time.Now()).Build()

		calls := 0
		uc.stages = []deleteUserStage{{
			name: "never_done",
			run: func(context.Context, *model.UserDeletionTarget) (bool, error) {
				calls++
				return false, nil
			},
		}}

		result, err := uc.Execute(context.Background(), DeleteUserInput{UserID: userID})
		if err != nil {
			t.Fatalf("Execute()のエラー = %v", err)
		}
		if result.Completed {
			t.Error("Completed = true、期待値 = false")
		}
		// 中断する前に少なくとも1単位は進める。
		if calls != 1 {
			t.Errorf("段階の呼び出し回数 = %d、期待値 = 1", calls)
		}
	})

	t.Run("上限が0なら段階を終えるまで続ける", func(t *testing.T) {
		t.Parallel()

		uc, tx := newDeleteUserUsecaseForTest(t, &subscriptionCancelerStub{}, 0)
		userID := testutil.NewUserBuilder(t, tx).WithDeletedAt(time.Now()).Build()

		calls := 0
		uc.stages = []deleteUserStage{{
			name: "done_after_three_calls",
			run: func(context.Context, *model.UserDeletionTarget) (bool, error) {
				calls++
				return calls >= 3, nil
			},
		}}

		result, err := uc.Execute(context.Background(), DeleteUserInput{UserID: userID})
		if err != nil {
			t.Fatalf("Execute()のエラー = %v", err)
		}
		if !result.Completed {
			t.Error("Completed = false、期待値 = true")
		}
		if calls != 3 {
			t.Errorf("段階の呼び出し回数 = %d、期待値 = 3", calls)
		}
	})
}

func TestDeleteUserUsecase_Execute_Canceled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		// runは段階の1単位。cancelを呼んでcontextをキャンセルできる。
		run       func(ctx context.Context, cancel context.CancelFunc) (bool, error)
		cancelNow bool
		wantCalls int
	}{
		{
			name:      "対象の取得前にキャンセルされていたら段階に進まずに中断する",
			run:       func(context.Context, context.CancelFunc) (bool, error) { return true, nil },
			cancelNow: true,
			wantCalls: 0,
		},
		{
			name: "1単位を終えたあとにキャンセルされたら次の単位に進まずに中断する",
			run: func(_ context.Context, cancel context.CancelFunc) (bool, error) {
				cancel()
				return false, nil
			},
			wantCalls: 1,
		},
		{
			name: "1単位の途中でキャンセルされたら失敗ではなく中断として扱う",
			run: func(ctx context.Context, cancel context.CancelFunc) (bool, error) {
				cancel()
				return false, fmt.Errorf("削除のクエリに失敗: %w", ctx.Err())
			},
			wantCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// 上限なしで呼び、上限時間ではなくキャンセルで中断したことを確かめる。
			uc, tx := newDeleteUserUsecaseForTest(t, &subscriptionCancelerStub{}, 0)
			userID := testutil.NewUserBuilder(t, tx).WithDeletedAt(time.Now()).Build()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if tt.cancelNow {
				cancel()
			}

			calls := 0
			uc.stages = []deleteUserStage{{
				name: "cancelable",
				run: func(ctx context.Context, _ *model.UserDeletionTarget) (bool, error) {
					calls++
					return tt.run(ctx, cancel)
				},
			}}

			result, err := uc.Execute(ctx, DeleteUserInput{UserID: userID})
			if err != nil {
				t.Fatalf("Execute()のエラー = %v、期待値 = nil", err)
			}
			if result.Completed {
				t.Error("Completed = true、期待値 = false")
			}
			if calls != tt.wantCalls {
				t.Errorf("段階の呼び出し回数 = %d、期待値 = %d", calls, tt.wantCalls)
			}
		})
	}
}

func TestDeleteUserUsecase_Execute_DeleteStatusesAndActivities(t *testing.T) {
	t.Parallel()

	uc, tx := newDeleteUserUsecaseForTest(t, &subscriptionCancelerStub{}, 0)
	// 一定件数ずつの削除を繰り返して段階を終えることを確かめるため、上限を小さくする。
	uc.batchSize = 2

	userID := testutil.NewUserBuilder(t, tx).WithDeletedAt(time.Now()).Build()
	otherUserID := testutil.NewUserBuilder(t, tx).Build()

	watchingWorkID := testutil.NewWorkBuilder(t, tx).WithWatchersCount(10).Build()
	watchedWorkID := testutil.NewWorkBuilder(t, tx).WithWatchersCount(10).Build()
	stopWatchingWorkID := testutil.NewWorkBuilder(t, tx).WithWatchersCount(10).Build()
	for workID, status := range map[model.WorkID]string{
		watchingWorkID:     "watching",
		watchedWorkID:      "watched",
		stopWatchingWorkID: "stop_watching",
	} {
		testutil.NewLibraryEntryBuilder(t, tx).WithUserID(userID).WithWorkID(workID).WithStatus(status).Build()
	}

	// ステータスを参照するアクティビティを1つのグループにまとめ、外部キーの順序どおりに消せることを確かめる。
	statusID := testutil.NewStatusBuilder(t, tx, userID, watchingWorkID).Build()
	activityGroupID := testutil.CreateTestActivityGroup(t, tx, userID)
	for range 3 {
		testutil.NewActivityBuilder(t, tx, userID, watchingWorkID).
			WithActivityGroupID(activityGroupID).
			WithStatusID(statusID).
			Build()
	}
	likeID := testutil.CreateTestLike(t, tx, otherUserID, "Status", statusID)
	notificationID := testutil.CreateTestNotification(t, tx, userID, otherUserID, "Like", likeID)

	otherStatusID := testutil.NewStatusBuilder(t, tx, otherUserID, watchingWorkID).Build()
	otherActivityID := testutil.NewActivityBuilder(t, tx, otherUserID, watchingWorkID).WithStatusID(otherStatusID).Build()
	otherLibraryEntryID := testutil.NewLibraryEntryBuilder(t, tx).WithUserID(otherUserID).WithWorkID(watchingWorkID).Build()

	result, err := uc.Execute(context.Background(), DeleteUserInput{UserID: userID})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if !result.Completed {
		t.Error("Completed = false、期待値 = true")
	}

	for _, table := range []string{"activities", "activity_groups", "library_entries", "statuses"} {
		if got := testutil.CountRows(t, tx, table, "user_id = $1", int64(userID)); got != 0 {
			t.Errorf("ユーザーの%sの件数 = %d、期待値 = 0", table, got)
		}
	}

	rows := []struct {
		name  string
		table string
		id    int64
		want  int
	}{
		{name: "ステータスへのいいね", table: "likes", id: likeID, want: 0},
		{name: "ステータスへのいいねの通知", table: "notifications", id: notificationID, want: 0},
		{name: "他のユーザーのステータス", table: "statuses", id: otherStatusID, want: 1},
		{name: "他のユーザーのアクティビティ", table: "activities", id: otherActivityID, want: 1},
		{name: "他のユーザーのライブラリエントリ", table: "library_entries", id: otherLibraryEntryID, want: 1},
	}
	for _, row := range rows {
		if got := testutil.CountRows(t, tx, row.table, "id = $1", row.id); got != row.want {
			t.Errorf("%sの件数 = %d、期待値 = %d", row.name, got, row.want)
		}
	}

	for workID, want := range map[model.WorkID]int32{
		watchingWorkID:     9,
		watchedWorkID:      9,
		stopWatchingWorkID: 10,
	} {
		var watchersCount int32
		if err := tx.QueryRow("SELECT watchers_count FROM works WHERE id = $1", int64(workID)).Scan(&watchersCount); err != nil {
			t.Fatalf("作品の視聴者数の取得に失敗: %v", err)
		}
		if watchersCount != want {
			t.Errorf("作品 %d の視聴者数 = %d、期待値 = %d", workID, watchersCount, want)
		}
	}
}

func TestDeleteUserUsecase_Execute_DeleteRecords(t *testing.T) {
	t.Parallel()

	uc, tx := newDeleteUserUsecaseForTest(t, &subscriptionCancelerStub{}, 0)
	// 一定件数ずつの削除を繰り返して段階を終えることを確かめるため、上限を小さくする。
	uc.batchSize = 1

	userID := testutil.NewUserBuilder(t, tx).WithDeletedAt(time.Now()).Build()
	otherUserID := testutil.NewUserBuilder(t, tx).Build()
	workID := testutil.NewWorkBuilder(t, tx).Build()
	episodeID := testutil.NewEpisodeBuilder(t, tx, workID).Build()
	if _, err := tx.Exec(
		"UPDATE works SET records_count = 10, work_records_count = 10, work_records_with_body_count = 10 WHERE id = $1",
		int64(workID),
	); err != nil {
		t.Fatalf("作品の記録数の設定に失敗: %v", err)
	}
	if _, err := tx.Exec(
		"UPDATE episodes SET episode_records_count = 10, episode_record_bodies_count = 10 WHERE id = $1",
		int64(episodeID),
	); err != nil {
		t.Fatalf("エピソードの記録数の設定に失敗: %v", err)
	}

	// 複数エピソードの記録・アクティビティ・コメントから参照される記録を作り、外部キーの順序どおりに消せることを確かめる。
	multipleEpisodeRecordID := testutil.CreateTestMultipleEpisodeRecord(t, tx, userID, workID)
	episodeRecords := make([]testutil.EpisodeRecordResult, 2)
	for i := range episodeRecords {
		episodeRecords[i] = testutil.NewEpisodeRecordBuilder(t, tx, userID, workID, episodeID).
			WithBody("感想").
			WithMultipleEpisodeRecordID(multipleEpisodeRecordID).
			Build()
		testutil.NewActivityBuilder(t, tx, userID, workID).WithEpisodeRecordID(episodeRecords[i].EpisodeRecordID).Build()
	}
	workRecord := testutil.NewWorkRecordBuilder(t, tx, userID, workID).WithBody("感想").Build()
	commentID := testutil.CreateTestComment(t, tx, otherUserID, episodeRecords[0].EpisodeRecordID)
	likeID := testutil.CreateTestLike(t, tx, otherUserID, "EpisodeRecord", episodeRecords[0].EpisodeRecordID)

	otherEpisodeRecord := testutil.NewEpisodeRecordBuilder(t, tx, otherUserID, workID, episodeID).Build()
	otherWorkRecord := testutil.NewWorkRecordBuilder(t, tx, otherUserID, workID).Build()

	result, err := uc.Execute(context.Background(), DeleteUserInput{UserID: userID})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if !result.Completed {
		t.Error("Completed = false、期待値 = true")
	}

	for _, table := range []string{"records", "episode_records", "work_records", "multiple_episode_records"} {
		if got := testutil.CountRows(t, tx, table, "user_id = $1", int64(userID)); got != 0 {
			t.Errorf("ユーザーの%sの件数 = %d、期待値 = 0", table, got)
		}
	}

	rows := []struct {
		name  string
		table string
		id    int64
		want  int
	}{
		{name: "記録へのコメント", table: "comments", id: commentID, want: 0},
		{name: "記録へのいいね", table: "likes", id: likeID, want: 0},
		{name: "作品の記録", table: "work_records", id: workRecord.WorkRecordID, want: 0},
		{name: "他のユーザーの記録", table: "records", id: otherEpisodeRecord.RecordID, want: 1},
		{name: "他のユーザーのエピソードの記録", table: "episode_records", id: otherEpisodeRecord.EpisodeRecordID, want: 1},
		{name: "他のユーザーの作品の記録", table: "work_records", id: otherWorkRecord.WorkRecordID, want: 1},
	}
	for _, row := range rows {
		if got := testutil.CountRows(t, tx, row.table, "id = $1", row.id); got != row.want {
			t.Errorf("%sの件数 = %d、期待値 = %d", row.name, got, row.want)
		}
	}

	var recordsCount, workRecordsCount, workRecordsWithBodyCount int32
	if err := tx.QueryRow(
		"SELECT records_count, work_records_count, work_records_with_body_count FROM works WHERE id = $1",
		int64(workID),
	).Scan(&recordsCount, &workRecordsCount, &workRecordsWithBodyCount); err != nil {
		t.Fatalf("作品の記録数の取得に失敗: %v", err)
	}
	if recordsCount != 7 {
		t.Errorf("作品の記録数 = %d、期待値 = 7", recordsCount)
	}
	if workRecordsCount != 9 {
		t.Errorf("作品の作品の記録数 = %d、期待値 = 9", workRecordsCount)
	}
	if workRecordsWithBodyCount != 9 {
		t.Errorf("作品の本文のある作品の記録数 = %d、期待値 = 9", workRecordsWithBodyCount)
	}

	var episodeRecordsCount, episodeRecordBodiesCount int32
	if err := tx.QueryRow(
		"SELECT episode_records_count, episode_record_bodies_count FROM episodes WHERE id = $1",
		int64(episodeID),
	).Scan(&episodeRecordsCount, &episodeRecordBodiesCount); err != nil {
		t.Fatalf("エピソードの記録数の取得に失敗: %v", err)
	}
	if episodeRecordsCount != 8 {
		t.Errorf("エピソードの記録数 = %d、期待値 = 8", episodeRecordsCount)
	}
	if episodeRecordBodiesCount != 8 {
		t.Errorf("エピソードの本文のある記録数 = %d、期待値 = 8", episodeRecordBodiesCount)
	}
}

func TestDeleteUserUsecase_Execute_DeleteLikesCommentsAndNotifications(t *testing.T) {
	t.Parallel()

	uc, tx := newDeleteUserUsecaseForTest(t, &subscriptionCancelerStub{}, 0)
	// 一定件数ずつの削除を繰り返して段階を終えることを確かめるため、上限を小さくする。
	uc.batchSize = 1

	userID := testutil.NewUserBuilder(t, tx).WithDeletedAt(time.Now()).Build()
	otherUserID := testutil.NewUserBuilder(t, tx).Build()
	thirdUserID := testutil.NewUserBuilder(t, tx).Build()
	workID := testutil.NewWorkBuilder(t, tx).Build()
	episodeID := testutil.NewEpisodeBuilder(t, tx, workID).Build()

	otherEpisodeRecord := testutil.NewEpisodeRecordBuilder(t, tx, otherUserID, workID, episodeID).Build()
	if _, err := tx.Exec(
		"UPDATE episode_records SET likes_count = 5, comments_count = 5 WHERE id = $1",
		otherEpisodeRecord.EpisodeRecordID,
	); err != nil {
		t.Fatalf("記録のいいね数・コメント数の設定に失敗: %v", err)
	}
	if _, err := tx.Exec("UPDATE users SET notifications_count = 5 WHERE id = $1", int64(otherUserID)); err != nil {
		t.Fatalf("未読数の設定に失敗: %v", err)
	}

	// ユーザーが他人の記録に付けたいいね・書いたコメントと、それで他人に届いた未読の通知。
	likeID := testutil.CreateTestLike(t, tx, userID, "EpisodeRecord", otherEpisodeRecord.EpisodeRecordID)
	likeNotificationID := testutil.CreateTestNotification(t, tx, otherUserID, userID, "Like", likeID)
	commentID := testutil.CreateTestComment(t, tx, userID, otherEpisodeRecord.EpisodeRecordID)
	commentNotificationID := testutil.CreateTestNotification(t, tx, otherUserID, userID, "Comment", commentID)
	// フォローの通知のように、いいね・コメント以外で他人に届いた通知もこの段階で削除する。
	followNotificationID := testutil.CreateTestNotification(t, tx, otherUserID, userID, "Follow", 1)

	// ユーザーのコメントに他人が付けたいいねと、それでユーザーに届いた通知。
	likeToCommentID := testutil.CreateTestLike(t, tx, thirdUserID, "Comment", commentID)
	notificationToUserID := testutil.CreateTestNotification(t, tx, userID, thirdUserID, "Like", likeToCommentID)

	// 他人どうしのいいね・コメント・通知は残す。
	otherLikeID := testutil.CreateTestLike(t, tx, thirdUserID, "EpisodeRecord", otherEpisodeRecord.EpisodeRecordID)
	otherNotificationID := testutil.CreateTestNotification(t, tx, otherUserID, thirdUserID, "Like", otherLikeID)
	otherCommentID := testutil.CreateTestComment(t, tx, thirdUserID, otherEpisodeRecord.EpisodeRecordID)

	result, err := uc.Execute(context.Background(), DeleteUserInput{UserID: userID})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if !result.Completed {
		t.Error("Completed = false、期待値 = true")
	}

	rows := []struct {
		name  string
		table string
		id    int64
		want  int
	}{
		{name: "ユーザーのいいね", table: "likes", id: likeID, want: 0},
		{name: "ユーザーのいいねの通知", table: "notifications", id: likeNotificationID, want: 0},
		{name: "ユーザーのコメント", table: "comments", id: commentID, want: 0},
		{name: "ユーザーのコメントの通知", table: "notifications", id: commentNotificationID, want: 0},
		{name: "ユーザーのフォローの通知", table: "notifications", id: followNotificationID, want: 0},
		{name: "ユーザーのコメントへのいいね", table: "likes", id: likeToCommentID, want: 0},
		{name: "ユーザー宛ての通知", table: "notifications", id: notificationToUserID, want: 0},
		{name: "他人のいいね", table: "likes", id: otherLikeID, want: 1},
		{name: "他人のいいねの通知", table: "notifications", id: otherNotificationID, want: 1},
		{name: "他人のコメント", table: "comments", id: otherCommentID, want: 1},
	}
	for _, row := range rows {
		if got := testutil.CountRows(t, tx, row.table, "id = $1", row.id); got != row.want {
			t.Errorf("%sの件数 = %d、期待値 = %d", row.name, got, row.want)
		}
	}

	var likesCount, commentsCount int32
	if err := tx.QueryRow(
		"SELECT likes_count, comments_count FROM episode_records WHERE id = $1",
		otherEpisodeRecord.EpisodeRecordID,
	).Scan(&likesCount, &commentsCount); err != nil {
		t.Fatalf("記録のいいね数・コメント数の取得に失敗: %v", err)
	}
	if likesCount != 4 {
		t.Errorf("記録のいいね数 = %d、期待値 = 4", likesCount)
	}
	if commentsCount != 4 {
		t.Errorf("記録のコメント数 = %d、期待値 = 4", commentsCount)
	}

	var notificationsCount int32
	if err := tx.QueryRow("SELECT notifications_count FROM users WHERE id = $1", int64(otherUserID)).Scan(&notificationsCount); err != nil {
		t.Fatalf("未読数の取得に失敗: %v", err)
	}
	if notificationsCount != 2 {
		t.Errorf("記録の持ち主の未読数 = %d、期待値 = 2", notificationsCount)
	}
}

func TestDeleteUserUsecase_Execute_DeleteFollowsMutesFavoritesAndTags(t *testing.T) {
	t.Parallel()

	uc, tx := newDeleteUserUsecaseForTest(t, &subscriptionCancelerStub{}, 0)
	// 一定件数ずつの削除を繰り返して段階を終えることを確かめるため、上限を小さくする。
	uc.batchSize = 1

	userID := testutil.NewUserBuilder(t, tx).WithDeletedAt(time.Now()).Build()
	otherUserID := testutil.NewUserBuilder(t, tx).Build()
	thirdUserID := testutil.NewUserBuilder(t, tx).Build()
	workID := testutil.NewWorkBuilder(t, tx).Build()

	characterID := testutil.CreateTestCharacter(t, tx)
	personID := testutil.CreateTestPerson(t, tx)
	organizationID := testutil.CreateTestOrganization(t, tx)
	workTagID := testutil.CreateTestWorkTag(t, tx)
	counters := []struct {
		name  string
		query string
		args  []any
		want  int32
	}{
		{name: "フォロー先のフォロワー数", query: "SELECT followers_count FROM users WHERE id = $1", args: []any{int64(otherUserID)}, want: 4},
		{name: "フォローしていた人のフォロー数", query: "SELECT following_count FROM users WHERE id = $1", args: []any{int64(thirdUserID)}, want: 4},
		{name: "キャラクターのお気に入り数", query: "SELECT favorite_users_count FROM characters WHERE id = $1", args: []any{characterID}, want: 4},
		{name: "人物のお気に入り数", query: "SELECT favorite_users_count FROM people WHERE id = $1", args: []any{personID}, want: 4},
		{name: "団体のお気に入り数", query: "SELECT favorite_users_count FROM organizations WHERE id = $1", args: []any{organizationID}, want: 4},
		{name: "タグの作品数", query: "SELECT work_taggings_count FROM work_tags WHERE id = $1", args: []any{workTagID}, want: 4},
	}
	for _, stmt := range []struct {
		query string
		arg   any
	}{
		{query: "UPDATE users SET followers_count = 5, following_count = 5 WHERE id = $1", arg: int64(otherUserID)},
		{query: "UPDATE users SET followers_count = 5, following_count = 5 WHERE id = $1", arg: int64(thirdUserID)},
		{query: "UPDATE characters SET favorite_users_count = 5 WHERE id = $1", arg: characterID},
		{query: "UPDATE people SET favorite_users_count = 5 WHERE id = $1", arg: personID},
		{query: "UPDATE organizations SET favorite_users_count = 5 WHERE id = $1", arg: organizationID},
		{query: "UPDATE work_tags SET work_taggings_count = 5 WHERE id = $1", arg: workTagID},
	} {
		if _, err := tx.Exec(stmt.query, stmt.arg); err != nil {
			t.Fatalf("カウンタの設定に失敗: %v", err)
		}
	}

	// ユーザーがしたフォロー・されたフォローと、それぞれの通知。
	followID := testutil.CreateTestFollow(t, tx, userID, otherUserID)
	followNotificationID := testutil.CreateTestNotification(t, tx, otherUserID, userID, "Follow", followID)
	followerID := testutil.CreateTestFollow(t, tx, thirdUserID, userID)
	followerNotificationID := testutil.CreateTestNotification(t, tx, userID, thirdUserID, "Follow", followerID)

	rows := []struct {
		name  string
		table string
		id    int64
		want  int
	}{
		{name: "ユーザーのフォロー", table: "follows", id: followID, want: 0},
		{name: "ユーザーのフォローの通知", table: "notifications", id: followNotificationID, want: 0},
		{name: "ユーザーへのフォロー", table: "follows", id: followerID, want: 0},
		{name: "ユーザーへのフォローの通知", table: "notifications", id: followerNotificationID, want: 0},
		{name: "ユーザーがしたミュート", table: "mute_users", id: testutil.CreateTestMuteUser(t, tx, userID, otherUserID), want: 0},
		{name: "ユーザーがされたミュート", table: "mute_users", id: testutil.CreateTestMuteUser(t, tx, otherUserID, userID), want: 0},
		{name: "キャラクターのお気に入り", table: "character_favorites", id: testutil.CreateTestCharacterFavorite(t, tx, userID, characterID), want: 0},
		{name: "人物のお気に入り", table: "person_favorites", id: testutil.CreateTestPersonFavorite(t, tx, userID, personID), want: 0},
		{name: "団体のお気に入り", table: "organization_favorites", id: testutil.CreateTestOrganizationFavorite(t, tx, userID, organizationID), want: 0},
		{name: "ユーザーのタグ付け", table: "work_taggings", id: testutil.CreateTestWorkTagging(t, tx, userID, workID, workTagID), want: 0},
		{name: "ユーザーが使ったタグの記録", table: "work_taggables", id: testutil.CreateTestWorkTaggable(t, tx, userID, workTagID), want: 0},
		// 他人どうしの行と、他人と共有するタグそのものは残す。
		{name: "他人どうしのフォロー", table: "follows", id: testutil.CreateTestFollow(t, tx, otherUserID, thirdUserID), want: 1},
		{name: "他人どうしのミュート", table: "mute_users", id: testutil.CreateTestMuteUser(t, tx, otherUserID, thirdUserID), want: 1},
		{name: "他人のお気に入り", table: "character_favorites", id: testutil.CreateTestCharacterFavorite(t, tx, otherUserID, characterID), want: 1},
		{name: "他人のタグ付け", table: "work_taggings", id: testutil.CreateTestWorkTagging(t, tx, otherUserID, workID, workTagID), want: 1},
		{name: "タグ", table: "work_tags", id: workTagID, want: 1},
	}

	result, err := uc.Execute(context.Background(), DeleteUserInput{UserID: userID})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if !result.Completed {
		t.Error("Completed = false、期待値 = true")
	}

	for _, row := range rows {
		if got := testutil.CountRows(t, tx, row.table, "id = $1", row.id); got != row.want {
			t.Errorf("%sの件数 = %d、期待値 = %d", row.name, got, row.want)
		}
	}
	for _, c := range counters {
		var got int32
		if err := tx.QueryRow(c.query, c.args...).Scan(&got); err != nil {
			t.Fatalf("%sの取得に失敗: %v", c.name, err)
		}
		if got != c.want {
			t.Errorf("%s = %d、期待値 = %d", c.name, got, c.want)
		}
	}
}

func TestDeleteUserUsecase_Execute_DeleteCollectionsForumsAndAccount(t *testing.T) {
	t.Parallel()

	uc, tx := newDeleteUserUsecaseForTest(t, &subscriptionCancelerStub{}, 0)
	// 一定件数ずつの削除を繰り返して段階を終えることを確かめるため、上限を小さくする。
	uc.batchSize = 1

	userID := testutil.NewUserBuilder(t, tx).WithDeletedAt(time.Now()).Build()
	otherUserID := testutil.NewUserBuilder(t, tx).Build()
	workID := testutil.NewWorkBuilder(t, tx).Build()

	collectionID := testutil.CreateTestCollection(t, tx, userID)
	forumPostID := testutil.CreateTestForumPost(t, tx, userID)
	otherForumPostID := testutil.CreateTestForumPost(t, tx, otherUserID)
	detachedForumCommentID := testutil.CreateTestForumComment(t, tx, userID, otherForumPostID)
	deletedApplicationID := testutil.CreateTestOAuthApplication(t, tx, userID, true)
	otherStatusID := testutil.NewStatusBuilder(t, tx, otherUserID, workID).Build()
	otherWorkRecordID := testutil.NewWorkRecordBuilder(t, tx, otherUserID, workID).Build().WorkRecordID
	otherEpisodeRecordID := testutil.NewEpisodeRecordBuilder(t, tx, otherUserID, workID, testutil.NewEpisodeBuilder(t, tx, workID).Build()).Build().EpisodeRecordID
	for _, record := range []struct {
		table string
		id    int64
	}{
		{table: "statuses", id: otherStatusID},
		{table: "work_records", id: otherWorkRecordID},
		{table: "episode_records", id: otherEpisodeRecordID},
	} {
		if _, err := tx.Exec("UPDATE "+record.table+" SET oauth_application_id = $1 WHERE id = $2", deletedApplicationID, record.id); err != nil {
			t.Fatalf("%sの投稿元のアプリケーションの設定に失敗: %v", record.table, err)
		}
	}
	activeApplicationID := testutil.CreateTestOAuthApplication(t, tx, otherUserID, false)

	rows := []struct {
		name  string
		table string
		id    int64
		want  int
	}{
		{name: "コレクション", table: "collections", id: collectionID, want: 0},
		{name: "コレクション項目", table: "collection_items", id: testutil.CreateTestCollectionItem(t, tx, userID, collectionID, workID), want: 0},
		{name: "コレクションへの他人のいいね", table: "likes", id: testutil.CreateTestLike(t, tx, otherUserID, "Collection", collectionID), want: 0},
		{name: "スレッド", table: "forum_posts", id: forumPostID, want: 0},
		{name: "スレッドへの他人のコメント", table: "forum_comments", id: testutil.CreateTestForumComment(t, tx, otherUserID, forumPostID), want: 0},
		{name: "スレッドへの他人の参加", table: "forum_post_participants", id: testutil.CreateTestForumPostParticipant(t, tx, otherUserID, forumPostID), want: 0},
		{name: "他人のスレッドへの参加", table: "forum_post_participants", id: testutil.CreateTestForumPostParticipant(t, tx, userID, otherForumPostID), want: 0},
		{name: "他人のスレッドに書いたコメント", table: "forum_comments", id: detachedForumCommentID, want: 1},
		{name: "Annict DBの編集履歴", table: "db_activities", id: testutil.CreateTestDbActivity(t, tx, userID, workID), want: 0},
		{name: "Annict DBのコメント", table: "db_comments", id: testutil.CreateTestDbComment(t, tx, userID, workID), want: 0},
		{name: "作品へのコメント", table: "work_comments", id: testutil.CreateTestWorkComment(t, tx, userID, workID), want: 0},
		{name: "Userlandプロジェクトへの参加", table: "userland_project_members", id: testutil.CreateTestUserlandProjectMember(t, tx, userID), want: 0},
		{name: "アクセストークン", table: "oauth_access_tokens", id: testutil.CreateTestOAuthAccessToken(t, tx, userID, activeApplicationID), want: 0},
		{name: "認可", table: "oauth_access_grants", id: testutil.CreateTestOAuthAccessGrant(t, tx, userID, activeApplicationID), want: 0},
		{name: "論理削除済みのアプリケーションに残る他人のアクセストークン", table: "oauth_access_tokens", id: testutil.CreateTestOAuthAccessToken(t, tx, otherUserID, deletedApplicationID), want: 0},
		{name: "論理削除済みのアプリケーションに残る他人の認可", table: "oauth_access_grants", id: testutil.CreateTestOAuthAccessGrant(t, tx, otherUserID, deletedApplicationID), want: 0},
		{name: "論理削除済みのアプリケーション", table: "oauth_applications", id: deletedApplicationID, want: 0},
		{name: "他人のステータス", table: "statuses", id: otherStatusID, want: 1},
		{name: "他人の作品の記録", table: "work_records", id: otherWorkRecordID, want: 1},
		{name: "他人のエピソードの記録", table: "episode_records", id: otherEpisodeRecordID, want: 1},
	}

	result, err := uc.Execute(context.Background(), DeleteUserInput{UserID: userID})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if !result.Completed {
		t.Error("Completed = false、期待値 = true")
	}

	for _, row := range rows {
		if got := testutil.CountRows(t, tx, row.table, "id = $1", row.id); got != row.want {
			t.Errorf("%sの件数 = %d、期待値 = %d", row.name, got, row.want)
		}
	}
	for _, c := range []struct {
		name  string
		table string
		id    int64
		where string
	}{
		{name: "他人のスレッドに書いたコメントの投稿者", table: "forum_comments", id: detachedForumCommentID, where: "user_id IS NULL"},
		{name: "他人のステータスの投稿元", table: "statuses", id: otherStatusID, where: "oauth_application_id IS NULL"},
		{name: "他人の作品の記録の投稿元", table: "work_records", id: otherWorkRecordID, where: "oauth_application_id IS NULL"},
		{name: "他人のエピソードの記録の投稿元", table: "episode_records", id: otherEpisodeRecordID, where: "oauth_application_id IS NULL"},
	} {
		if got := testutil.CountRows(t, tx, c.table, "id = $1 AND "+c.where, c.id); got != 1 {
			t.Errorf("%sが外れていない", c.name)
		}
	}
}

func TestDeleteUserUsecase_Execute_DeleteUserAndSubscribers(t *testing.T) {
	t.Parallel()

	// ユーザー行の削除段階が開くトランザクションから見えるよう、ユーザー行とその従属データだけを
	// コミットして準備する。いずれも削除で消えるため後片付けは要らない。
	uc, db, tx := newCommittedDeleteUserUsecaseForTest(t, &subscriptionCancelerStub{result: annictstripe.CancelSubscriptionResultCanceled}, 0)

	stripeSubscriberID := testutil.NewStripeSubscriberBuilder(t, tx).Build()
	gumroadSubscriberID := testutil.NewGumroadSubscriberBuilder(t, tx).Build()
	userID := testutil.NewUserBuilder(t, tx).
		WithStripeSubscriberID(&stripeSubscriberID).
		WithDeletedAt(time.Now()).
		Build()
	if _, err := tx.Exec(`UPDATE users SET gumroad_subscriber_id = $1 WHERE id = $2`, int64(gumroadSubscriberID), int64(userID)); err != nil {
		t.Fatalf("Gumroadの購読情報の設定に失敗: %v", err)
	}
	emailConfirmationID := testutil.CreateTestEmailConfirmation(t, tx, userID)
	if err := tx.Commit(); err != nil {
		t.Fatalf("準備用トランザクションのコミットに失敗: %v", err)
	}

	result, err := uc.Execute(context.Background(), DeleteUserInput{UserID: userID})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if !result.Completed {
		t.Error("Completed = false、期待値 = true")
	}

	readTx := beginReadTxForTest(t, db)
	for _, row := range []struct {
		name  string
		table string
		id    int64
	}{
		{name: "ユーザー", table: "users", id: int64(userID)},
		{name: "メールアドレスの確認", table: "email_confirmations", id: emailConfirmationID},
		{name: "Stripeの購読情報", table: "stripe_subscribers", id: int64(stripeSubscriberID)},
		{name: "Gumroadの購読情報", table: "gumroad_subscribers", id: int64(gumroadSubscriberID)},
	} {
		if got := testutil.CountRows(t, readTx, row.table, "id = $1", row.id); got != 0 {
			t.Errorf("%sの件数 = %d、期待値 = 0", row.name, got)
		}
	}
}

func TestDeleteUserUsecase_Execute_DeleteProfileImages(t *testing.T) {
	t.Parallel()

	const imageData = `{"original":{"id":"profile/1/image/original-a.png","storage":"store"},"master":{"id":"profile/1/image/master-a.jpg","storage":"store"}}`
	const backgroundImageData = `{"original":{"id":"profile/1/background_image/original-b.png","storage":"store"}}`

	tests := []struct {
		name                string
		imageData           string
		backgroundImageData string
		deleteErr           error
		wantKeys            [][]string
		wantErr             bool
		wantUserRows        int
	}{
		{
			name:                "正常系: プロフィール画像と背景画像を削除してからユーザーを削除する",
			imageData:           imageData,
			backgroundImageData: backgroundImageData,
			wantKeys: [][]string{{
				"shrine/profile/1/image/original-a.png",
				"shrine/profile/1/image/master-a.jpg",
				"shrine/profile/1/background_image/original-b.png",
			}},
			wantUserRows: 0,
		},
		{
			name:         "正常系: 画像が無ければストレージを呼ばない",
			wantKeys:     nil,
			wantUserRows: 0,
		},
		{
			name:                "正常系: 解析できない添付データは省いて削除を進める",
			imageData:           `{`,
			backgroundImageData: backgroundImageData,
			wantKeys:            [][]string{{"shrine/profile/1/background_image/original-b.png"}},
			wantUserRows:        0,
		},
		{
			name:         "異常系: ストレージからの削除に失敗したらユーザーを削除しない",
			imageData:    imageData,
			deleteErr:    errors.New("ストレージの一時的な障害"),
			wantKeys:     [][]string{{"shrine/profile/1/image/original-a.png", "shrine/profile/1/image/master-a.jpg"}},
			wantErr:      true,
			wantUserRows: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc, db, tx := newCommittedDeleteUserUsecaseForTest(t, &subscriptionCancelerStub{}, 0)
			deleter := &objectDeleterStub{err: tt.deleteErr}
			uc.objectDeleter = deleter

			userID := testutil.NewUserBuilder(t, tx).WithDeletedAt(time.Now()).Build()
			nullIfEmpty := func(s string) sql.NullString { return sql.NullString{String: s, Valid: s != ""} }
			if _, err := tx.Exec(
				`UPDATE profiles SET image_data = $1, background_image_data = $2 WHERE user_id = $3`,
				nullIfEmpty(tt.imageData), nullIfEmpty(tt.backgroundImageData), int64(userID),
			); err != nil {
				t.Fatalf("プロフィールの画像の設定に失敗: %v", err)
			}
			if err := tx.Commit(); err != nil {
				t.Fatalf("準備用トランザクションのコミットに失敗: %v", err)
			}
			// ストレージからの削除に失敗したケースで残るユーザーを片付ける。
			t.Cleanup(func() {
				cleanupTx, err := db.Begin()
				if err != nil {
					t.Errorf("後片付けのトランザクションの開始に失敗: %v", err)
					return
				}
				defer func() { _ = cleanupTx.Rollback() }()
				if _, err := uc.userRepo.WithTx(cleanupTx).DeleteByID(context.Background(), userID); err != nil {
					t.Errorf("後片付けでユーザーを削除できない: %v", err)
					return
				}
				if err := cleanupTx.Commit(); err != nil {
					t.Errorf("後片付けのコミットに失敗: %v", err)
				}
			})

			_, err := uc.Execute(context.Background(), DeleteUserInput{UserID: userID})
			if (err != nil) != tt.wantErr {
				t.Fatalf("Execute()のエラー = %v、エラーの期待 = %v", err, tt.wantErr)
			}
			if fmt.Sprint(deleter.calledWith) != fmt.Sprint(tt.wantKeys) {
				t.Errorf("DeleteObjects()の呼び出し = %v、期待値 = %v", deleter.calledWith, tt.wantKeys)
			}
			if got := testutil.CountRows(t, beginReadTxForTest(t, db), "users", "id = $1", int64(userID)); got != tt.wantUserRows {
				t.Errorf("ユーザーの件数 = %d、期待値 = %d", got, tt.wantUserRows)
			}
		})
	}
}

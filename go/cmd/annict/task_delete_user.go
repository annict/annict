package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/storage"
	annictStripe "github.com/annict/annict/go/internal/stripe"
	"github.com/annict/annict/go/internal/usecase"
)

// newDeleteUserUsecaseは、論理削除したユーザーのデータを削除するUseCaseを組み立てる。
// `serve` (削除ジョブ) と `task delete-user` (同期実行) の双方が本ヘルパー経由で組み立てるため、
// 削除の段階が使うリポジトリの配線がずれない。StripeとストレージのアダプタはserveがほかのUseCase
// と共有するため、呼び出し側から受け取る。timeLimitは1回の実行時間の上限で、0なら最後まで削除する。
func newDeleteUserUsecase(
	db *sql.DB,
	queries *query.Queries,
	subscriptionCanceler usecase.SubscriptionCanceler,
	objectDeleter usecase.ObjectDeleter,
	timeLimit time.Duration,
) *usecase.DeleteUserUsecase {
	return usecase.NewDeleteUserUsecase(
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
		subscriptionCanceler,
		objectDeleter,
		timeLimit,
	)
}

// deleteUserはdelete-userタスクの本体で、指定したユーザーのデータを最後まで削除する。
// 削除ジョブと違いRiverを介さず、実行時間の上限を設けずに同じUseCaseを呼ぶ。モデレーションでの
// アカウント削除と、削除ジョブが再試行を打ち切ったあとの再実行に使う。論理削除されていない
// ユーザーはUseCaseが削除を始めずにエラーを返す。段階ごとの進捗はUseCaseがログに出す。
func deleteUser(ctx context.Context, cfg *config.Config, db *sql.DB, queries *query.Queries, args []string) error {
	userID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil || userID <= 0 {
		return fmt.Errorf("ユーザーIDには正の整数を指定してください: %q", args[0])
	}

	storageAdapter := storage.NewAdapter(
		storage.NewClient(cfg.S3Endpoint, cfg.S3AccessKeyID, cfg.S3SecretAccessKey, cfg.S3Region),
		cfg.S3BucketName,
	)
	uc := newDeleteUserUsecase(db, queries, annictStripe.NewAdapter(annictStripe.NewClient(cfg.StripeSecretKey)), storageAdapter, 0)

	result, err := uc.Execute(ctx, usecase.DeleteUserInput{UserID: model.UserID(userID)})
	if err != nil {
		return fmt.Errorf("ユーザーの削除に失敗しました: %w", err)
	}
	if !result.Completed {
		return errors.New("ユーザーの削除が中断されました。同じコマンドを再実行すると続きから削除します")
	}

	return nil
}

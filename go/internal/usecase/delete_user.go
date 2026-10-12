package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/annict/annict/go/internal/image"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
	annictstripe "github.com/annict/annict/go/internal/stripe"
)

// DeleteUserTimeLimitは削除ジョブ1回の実行時間の上限。記録の多いユーザーでも1回の実行が
// ワーカーを長く占有しないよう区切り、残りは次の実行に持ち越す。
const DeleteUserTimeLimit = 5 * time.Minute

// deleteUserBatchSizeは削除の段階が1回のDELETEで削除する行の上限。1回の削除を短く
// 保ち、ロックの保持時間とWALの量を抑える。
const deleteUserBatchSize = 1000

// ErrUserNotDeletedは論理削除されていないユーザーを削除しようとしたことを表す。
// 誤って投入されたジョブで有効なユーザーを消さないため、削除を始めずに返す。
var ErrUserNotDeleted = errors.New("論理削除されていないユーザーは削除できません")

// SubscriptionCancelerはStripeサブスクリプションの即時解約を抽象化する。
type SubscriptionCanceler interface {
	CancelSubscription(ctx context.Context, subscriptionID string) (annictstripe.CancelSubscriptionResult, error)
}

// ObjectDeleterはオブジェクトストレージからのオブジェクトの削除を抽象化する。
type ObjectDeleter interface {
	DeleteObjects(ctx context.Context, keys []string) error
}

// DeleteUserUsecaseは退会・利用停止で論理削除したユーザーのデータを段階的に削除する。
type DeleteUserUsecase struct {
	db                        *sql.DB
	userRepo                  *repository.UserRepository
	stripeSubscriberRepo      *repository.StripeSubscriberRepository
	activityRepo              *repository.ActivityRepository
	activityGroupRepo         *repository.ActivityGroupRepository
	libraryEntryRepo          *repository.LibraryEntryRepository
	statusRepo                *repository.StatusRepository
	commentRepo               *repository.CommentRepository
	episodeRecordRepo         *repository.EpisodeRecordRepository
	workRecordRepo            *repository.WorkRecordRepository
	multipleEpisodeRecordRepo *repository.MultipleEpisodeRecordRepository
	recordRepo                *repository.RecordRepository
	notificationRepo          *repository.NotificationRepository
	likeRepo                  *repository.LikeRepository
	followRepo                *repository.FollowRepository
	muteUserRepo              *repository.MuteUserRepository
	characterFavoriteRepo     *repository.CharacterFavoriteRepository
	personFavoriteRepo        *repository.PersonFavoriteRepository
	organizationFavoriteRepo  *repository.OrganizationFavoriteRepository
	workTaggingRepo           *repository.WorkTaggingRepository
	workTaggableRepo          *repository.WorkTaggableRepository
	collectionItemRepo        *repository.CollectionItemRepository
	collectionRepo            *repository.CollectionRepository
	forumPostParticipantRepo  *repository.ForumPostParticipantRepository
	forumCommentRepo          *repository.ForumCommentRepository
	forumPostRepo             *repository.ForumPostRepository
	dbActivityRepo            *repository.DbActivityRepository
	dbCommentRepo             *repository.DbCommentRepository
	workCommentRepo           *repository.WorkCommentRepository
	userlandProjectMemberRepo *repository.UserlandProjectMemberRepository
	oauthAccessTokenRepo      *repository.OAuthAccessTokenRepository
	oauthAccessGrantRepo      *repository.OAuthAccessGrantRepository
	oauthApplicationRepo      *repository.OAuthApplicationRepository
	profileRepo               *repository.ProfileRepository
	subscriptionCanceler      SubscriptionCanceler
	objectDeleter             ObjectDeleter
	timeLimit                 time.Duration
	batchSize                 int32
	// stagesは削除処理の段階を実行順に並べたもの。Stripeの解約を最初に置き、
	// 解約できるまでデータの削除に進まないようにする。
	stages []deleteUserStage
}

// NewDeleteUserUsecaseは新しいDeleteUserUsecaseを作成する。
// timeLimitは1回の実行時間の上限で、0なら上限を設けずに最後まで削除する。
func NewDeleteUserUsecase(
	db *sql.DB,
	userRepo *repository.UserRepository,
	stripeSubscriberRepo *repository.StripeSubscriberRepository,
	activityRepo *repository.ActivityRepository,
	activityGroupRepo *repository.ActivityGroupRepository,
	libraryEntryRepo *repository.LibraryEntryRepository,
	statusRepo *repository.StatusRepository,
	commentRepo *repository.CommentRepository,
	episodeRecordRepo *repository.EpisodeRecordRepository,
	workRecordRepo *repository.WorkRecordRepository,
	multipleEpisodeRecordRepo *repository.MultipleEpisodeRecordRepository,
	recordRepo *repository.RecordRepository,
	notificationRepo *repository.NotificationRepository,
	likeRepo *repository.LikeRepository,
	followRepo *repository.FollowRepository,
	muteUserRepo *repository.MuteUserRepository,
	characterFavoriteRepo *repository.CharacterFavoriteRepository,
	personFavoriteRepo *repository.PersonFavoriteRepository,
	organizationFavoriteRepo *repository.OrganizationFavoriteRepository,
	workTaggingRepo *repository.WorkTaggingRepository,
	workTaggableRepo *repository.WorkTaggableRepository,
	collectionItemRepo *repository.CollectionItemRepository,
	collectionRepo *repository.CollectionRepository,
	forumPostParticipantRepo *repository.ForumPostParticipantRepository,
	forumCommentRepo *repository.ForumCommentRepository,
	forumPostRepo *repository.ForumPostRepository,
	dbActivityRepo *repository.DbActivityRepository,
	dbCommentRepo *repository.DbCommentRepository,
	workCommentRepo *repository.WorkCommentRepository,
	userlandProjectMemberRepo *repository.UserlandProjectMemberRepository,
	oauthAccessTokenRepo *repository.OAuthAccessTokenRepository,
	oauthAccessGrantRepo *repository.OAuthAccessGrantRepository,
	oauthApplicationRepo *repository.OAuthApplicationRepository,
	profileRepo *repository.ProfileRepository,
	subscriptionCanceler SubscriptionCanceler,
	objectDeleter ObjectDeleter,
	timeLimit time.Duration,
) *DeleteUserUsecase {
	uc := &DeleteUserUsecase{
		db:                        db,
		userRepo:                  userRepo,
		stripeSubscriberRepo:      stripeSubscriberRepo,
		activityRepo:              activityRepo,
		activityGroupRepo:         activityGroupRepo,
		libraryEntryRepo:          libraryEntryRepo,
		statusRepo:                statusRepo,
		commentRepo:               commentRepo,
		episodeRecordRepo:         episodeRecordRepo,
		workRecordRepo:            workRecordRepo,
		multipleEpisodeRecordRepo: multipleEpisodeRecordRepo,
		recordRepo:                recordRepo,
		notificationRepo:          notificationRepo,
		likeRepo:                  likeRepo,
		followRepo:                followRepo,
		muteUserRepo:              muteUserRepo,
		characterFavoriteRepo:     characterFavoriteRepo,
		personFavoriteRepo:        personFavoriteRepo,
		organizationFavoriteRepo:  organizationFavoriteRepo,
		workTaggingRepo:           workTaggingRepo,
		workTaggableRepo:          workTaggableRepo,
		collectionItemRepo:        collectionItemRepo,
		collectionRepo:            collectionRepo,
		forumPostParticipantRepo:  forumPostParticipantRepo,
		forumCommentRepo:          forumCommentRepo,
		forumPostRepo:             forumPostRepo,
		dbActivityRepo:            dbActivityRepo,
		dbCommentRepo:             dbCommentRepo,
		workCommentRepo:           workCommentRepo,
		userlandProjectMemberRepo: userlandProjectMemberRepo,
		oauthAccessTokenRepo:      oauthAccessTokenRepo,
		oauthAccessGrantRepo:      oauthAccessGrantRepo,
		oauthApplicationRepo:      oauthApplicationRepo,
		profileRepo:               profileRepo,
		subscriptionCanceler:      subscriptionCanceler,
		objectDeleter:             objectDeleter,
		timeLimit:                 timeLimit,
		batchSize:                 deleteUserBatchSize,
	}
	// 削除の段階は、即時に検査される外部キーを満たす順に並べる。アクティビティは
	// アクティビティグループ・ステータス・各種の記録を、ライブラリエントリはステータスを、
	// エピソードの記録は作品の記録 (review_id)・複数エピソードの記録・記録を、作品の記録は
	// 記録を参照するため、参照する側を先に削除する。記録に付いたコメントは、コメントの
	// いいね・通知と一緒に消すため、エピソードの記録より先に削除する。
	// ユーザーが起こした通知は、他人の通知一覧に参照先の無い通知を残さないよう、通知の
	// きっかけになったいいね・コメントより先に削除する。フォローの通知も同じ段階で消えるため、
	// フォローはそのあとに削除する。
	// コレクション項目はコレクションを、フォーラムの参加者とコメントはスレッドを参照するため、
	// 参照する側を先に削除する。論理削除済みのOAuthアプリケーションは、参照するアクセストークン・
	// 認可を削除し、他人の記録・ステータスからの参照を外してから削除する。
	// プロフィール画像はユーザー行と一緒に消えるプロフィールから求めるため、ユーザー行より先に
	// ストレージから削除する。ユーザー行は最後に削除し、残りの従属データは外部キーのON DELETEに任せる。
	uc.stages = []deleteUserStage{
		{name: "cancel_stripe_subscription", run: uc.cancelStripeSubscription},
		{name: "delete_activities", run: uc.deleteInBatches(uc.activityRepo.DeleteByUserID)},
		{name: "delete_activity_groups", run: uc.deleteInBatches(uc.activityGroupRepo.DeleteByUserID)},
		{name: "delete_library_entries", run: uc.deleteInBatches(uc.libraryEntryRepo.DeleteByUserID)},
		{name: "delete_statuses", run: uc.deleteInBatches(uc.statusRepo.DeleteByUserID)},
		{name: "delete_comments_on_episode_records", run: uc.deleteInBatches(uc.commentRepo.DeleteOnEpisodeRecordsByUserID)},
		{name: "delete_episode_records", run: uc.deleteInBatches(uc.episodeRecordRepo.DeleteByUserID)},
		{name: "delete_work_records", run: uc.deleteInBatches(uc.workRecordRepo.DeleteByUserID)},
		{name: "delete_multiple_episode_records", run: uc.deleteInBatches(uc.multipleEpisodeRecordRepo.DeleteByUserID)},
		{name: "delete_records", run: uc.deleteInBatches(uc.recordRepo.DeleteByUserID)},
		{name: "delete_notifications_by_action_user", run: uc.deleteInBatches(uc.notificationRepo.DeleteByActionUserID)},
		{name: "delete_notifications", run: uc.deleteInBatches(uc.notificationRepo.DeleteByUserID)},
		{name: "delete_likes", run: uc.deleteInBatches(uc.likeRepo.DeleteByUserID)},
		{name: "delete_comments", run: uc.deleteInBatches(uc.commentRepo.DeleteByUserID)},
		{name: "delete_follows", run: uc.deleteInBatches(uc.followRepo.DeleteByUserID)},
		{name: "delete_followers", run: uc.deleteInBatches(uc.followRepo.DeleteByFollowingID)},
		{name: "delete_mute_users", run: uc.deleteInBatches(uc.muteUserRepo.DeleteByUserID)},
		{name: "delete_muted_users", run: uc.deleteInBatches(uc.muteUserRepo.DeleteByMutedUserID)},
		{name: "delete_character_favorites", run: uc.deleteInBatches(uc.characterFavoriteRepo.DeleteByUserID)},
		{name: "delete_person_favorites", run: uc.deleteInBatches(uc.personFavoriteRepo.DeleteByUserID)},
		{name: "delete_organization_favorites", run: uc.deleteInBatches(uc.organizationFavoriteRepo.DeleteByUserID)},
		{name: "delete_work_taggings", run: uc.deleteInBatches(uc.workTaggingRepo.DeleteByUserID)},
		{name: "delete_work_taggables", run: uc.deleteInBatches(uc.workTaggableRepo.DeleteByUserID)},
		{name: "delete_collection_items", run: uc.deleteInBatches(uc.collectionItemRepo.DeleteByUserID)},
		{name: "delete_collections", run: uc.deleteInBatches(uc.collectionRepo.DeleteByUserID)},
		{name: "delete_forum_post_participants", run: uc.deleteInBatches(uc.forumPostParticipantRepo.DeleteByUserID)},
		{name: "delete_forum_comments_on_forum_posts", run: uc.deleteInBatches(uc.forumCommentRepo.DeleteOnForumPostsByUserID)},
		{name: "detach_forum_comments", run: uc.deleteInBatches(uc.forumCommentRepo.DetachFromUser)},
		{name: "delete_forum_posts", run: uc.deleteInBatches(uc.forumPostRepo.DeleteByUserID)},
		{name: "delete_db_activities", run: uc.deleteInBatches(uc.dbActivityRepo.DeleteByUserID)},
		{name: "delete_db_comments", run: uc.deleteInBatches(uc.dbCommentRepo.DeleteByUserID)},
		{name: "delete_work_comments", run: uc.deleteInBatches(uc.workCommentRepo.DeleteByUserID)},
		{name: "delete_userland_project_members", run: uc.deleteInBatches(uc.userlandProjectMemberRepo.DeleteByUserID)},
		{name: "delete_oauth_access_tokens", run: uc.deleteInBatches(uc.oauthAccessTokenRepo.DeleteByUserID)},
		{name: "delete_oauth_access_grants", run: uc.deleteInBatches(uc.oauthAccessGrantRepo.DeleteByUserID)},
		{name: "detach_deleted_oauth_applications_from_statuses", run: uc.deleteInBatches(uc.statusRepo.DetachDeletedOAuthApplicationsByOwnerID)},
		{name: "detach_deleted_oauth_applications_from_episode_records", run: uc.deleteInBatches(uc.episodeRecordRepo.DetachDeletedOAuthApplicationsByOwnerID)},
		{name: "detach_deleted_oauth_applications_from_work_records", run: uc.deleteInBatches(uc.workRecordRepo.DetachDeletedOAuthApplicationsByOwnerID)},
		{name: "delete_deleted_oauth_applications", run: uc.deleteInBatches(uc.oauthApplicationRepo.DeleteDeletedByOwnerID)},
		{name: "delete_profile_images", run: uc.deleteProfileImages},
		{name: "delete_user", run: uc.deleteUser},
	}
	return uc
}

// DeleteUserInputはユーザーの削除の入力。
type DeleteUserInput struct {
	UserID model.UserID
	// FinalAttemptは、失敗してもこれ以上再試行されない実行かを表す。trueで失敗した場合は、
	// 運営者が手動で再実行できるようエラーログを出す。
	FinalAttempt bool
}

// DeleteUserResultはユーザーの削除の結果。
type DeleteUserResult struct {
	// Completedはすべての段階を終えたかを表す。実行時間の上限、またはcontextのキャンセルで
	// 打ち切った場合はfalseで、呼び出し側は同じ入力で再度実行して続きを進める。
	Completed bool
}

// deleteUserStageは削除処理の1段階。runは段階を1単位 (一定件数の削除など) 進め、
// 段階を終えたらtrueを返す。すでに終えた段階を再度呼んでもtrueを返すよう冪等に作る。
type deleteUserStage struct {
	name string
	run  func(ctx context.Context, target *model.UserDeletionTarget) (bool, error)
}

// Executeはユーザーの削除を段階順に進める。進捗はデータそのもの (残っている行) が表すため、
// 中断したあとに同じ入力で実行すれば続きから進む。ユーザーがすでに存在しない場合は
// 削除を終えたものとして扱う。
func (uc *DeleteUserUsecase) Execute(ctx context.Context, input DeleteUserInput) (result *DeleteUserResult, err error) {
	start := time.Now()
	stageName := "find_deletion_target"
	defer func() {
		// 論理削除されていないユーザーはWorkerがジョブを取り消すため、再実行の案内は出さない。
		if input.FinalAttempt && err != nil && !errors.Is(err, ErrUserNotDeleted) {
			slog.ErrorContext(ctx, "ユーザーの削除が最終試行で失敗しました。原因を確認し、必要ならCLIから再実行してください",
				"user_id", input.UserID,
				"stage", stageName,
				"error", err,
			)
		}
	}()

	target, err := uc.userRepo.FindDeletionTarget(ctx, input.UserID)
	if err != nil {
		if isCanceled(ctx, err) {
			return uc.interrupt(ctx, input, stageName), nil
		}
		return nil, fmt.Errorf("削除対象のユーザーの取得に失敗: %w", err)
	}
	if target == nil {
		slog.InfoContext(ctx, "ユーザーが存在しないため削除を終えたものとして扱います", "user_id", input.UserID)
		return &DeleteUserResult{Completed: true}, nil
	}
	stageName = "verify_deleted"
	if !target.Deleted {
		return nil, fmt.Errorf("%w (ユーザーID: %d)", ErrUserNotDeleted, input.UserID)
	}

	for _, stage := range uc.stages {
		stageName = stage.name
		for {
			if isCanceled(ctx, ctx.Err()) {
				return uc.interrupt(ctx, input, stage.name), nil
			}
			done, err := stage.run(ctx, target)
			if err != nil {
				// 1単位の途中でキャンセルされた場合も中断として扱う。1単位はトランザクションで閉じ、
				// Stripeの解約も冪等なため、次の実行で同じ単位からやり直せる。
				if isCanceled(ctx, err) {
					return uc.interrupt(ctx, input, stage.name), nil
				}
				return nil, fmt.Errorf("ユーザーの削除の段階 %s に失敗: %w", stage.name, err)
			}
			if done {
				// CLIで同期的に実行したときに、長い削除がどこまで進んだかを追えるようにする。
				slog.InfoContext(ctx, "ユーザーの削除の段階を終えました",
					"user_id", input.UserID,
					"stage", stage.name,
					"elapsed", time.Since(start),
				)
				break
			}

			if uc.timeLimit > 0 && time.Since(start) >= uc.timeLimit {
				slog.InfoContext(ctx, "実行時間の上限に達したためユーザーの削除を中断します",
					"user_id", input.UserID,
					"stage", stage.name,
					"elapsed", time.Since(start),
				)
				return &DeleteUserResult{Completed: false}, nil
			}
		}
	}

	slog.InfoContext(ctx, "ユーザーの削除を終えました", "user_id", input.UserID, "elapsed", time.Since(start))
	return &DeleteUserResult{Completed: true}, nil
}

// isCanceledは、呼び出し元のcontextのキャンセル (ワーカーの停止など) によってerrが起きたかを返す。
// タイムアウト (context.DeadlineExceeded) は中断ではなく失敗として扱うため含めない。
func isCanceled(ctx context.Context, err error) bool {
	return errors.Is(err, context.Canceled) && errors.Is(ctx.Err(), context.Canceled)
}

// interruptは、contextのキャンセルで削除を中断したことを記録し、未完了の結果を返す。
func (uc *DeleteUserUsecase) interrupt(ctx context.Context, input DeleteUserInput, stageName string) *DeleteUserResult {
	slog.InfoContext(ctx, "contextがキャンセルされたためユーザーの削除を中断します",
		"user_id", input.UserID,
		"stage", stageName,
	)
	return &DeleteUserResult{Completed: false}
}

// cancelStripeSubscriptionはユーザーのStripeサブスクリプションを即時解約する。
// 解約済み・Stripe上に存在しない場合も解約できたものとして扱う。
func (uc *DeleteUserUsecase) cancelStripeSubscription(ctx context.Context, target *model.UserDeletionTarget) (bool, error) {
	if target.StripeSubscriberID == nil {
		return true, nil
	}

	subscriber, err := uc.stripeSubscriberRepo.GetByID(ctx, *target.StripeSubscriberID)
	if err != nil {
		return false, fmt.Errorf("StripeSubscriberの取得に失敗: %w", err)
	}
	if subscriber == nil {
		return true, nil
	}

	result, err := uc.subscriptionCanceler.CancelSubscription(ctx, subscriber.StripeSubscriptionID)
	if err != nil {
		return false, fmt.Errorf("サブスクリプションのStripeでの解約に失敗: %w", err)
	}

	attrs := []any{
		"user_id", target.ID,
		"stripe_subscriber_id", subscriber.ID,
		"stripe_subscription_id", subscriber.StripeSubscriptionID,
	}
	switch result {
	case annictstripe.CancelSubscriptionResultCanceled:
		slog.InfoContext(ctx, "Stripeサブスクリプションを解約しました", attrs...)
	case annictstripe.CancelSubscriptionResultAlreadyEnded:
		slog.InfoContext(ctx, "Stripeサブスクリプションは解約済みでした", attrs...)
	case annictstripe.CancelSubscriptionResultNotFound:
		slog.InfoContext(ctx, "StripeサブスクリプションがStripe上に存在しませんでした", attrs...)
	default:
		return false, fmt.Errorf("サブスクリプションの解約の結果が不明です: %d", result)
	}
	return true, nil
}

// deleteProfileImagesは、ユーザーのプロフィール画像と背景画像をオブジェクトストレージから削除する。
// 存在しないオブジェクトの削除も成功するため、再実行しても失敗しない。
func (uc *DeleteUserUsecase) deleteProfileImages(ctx context.Context, target *model.UserDeletionTarget) (bool, error) {
	imageDataList, err := uc.profileRepo.ListImageDataByUserID(ctx, target.ID)
	if err != nil {
		return false, fmt.Errorf("プロフィール画像の添付データの取得に失敗: %w", err)
	}

	var keys []string
	for _, imageData := range imageDataList {
		imageKeys, err := image.ObjectKeys(imageData)
		if err != nil {
			// 解析できない添付データのために削除全体を止めないよう、記録して残りを進める。
			slog.WarnContext(ctx, "プロフィール画像の添付データを解析できないため、画像の削除を省きます",
				"user_id", target.ID,
				"error", err,
			)
			continue
		}
		keys = append(keys, imageKeys...)
	}
	if len(keys) == 0 {
		return true, nil
	}

	if err := uc.objectDeleter.DeleteObjects(ctx, keys); err != nil {
		return false, fmt.Errorf("プロフィール画像のストレージからの削除に失敗: %w", err)
	}
	slog.InfoContext(ctx, "プロフィール画像をストレージから削除しました", "user_id", target.ID, "count", len(keys))
	return true, nil
}

// deleteUserはユーザー行を削除する。すでに削除されていても成功として扱う。
// 購読情報を共有するユーザーの削除を直列化する行ロックを削除まで保持するため、トランザクションで囲む。
func (uc *DeleteUserUsecase) deleteUser(ctx context.Context, target *model.UserDeletionTarget) (bool, error) {
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("ユーザー削除トランザクションの開始に失敗: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := uc.userRepo.WithTx(tx).DeleteByID(ctx, target.ID); err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("ユーザー削除トランザクションのコミットに失敗: %w", err)
	}
	return true, nil
}

// deleteInBatchesは、ユーザーの行を一定件数ずつ削除する関数を段階に変換する。
// 削除した件数が上限に満たなければ、残りが無くなったものとして段階を終える。
func (uc *DeleteUserUsecase) deleteInBatches(
	deleteByUserID func(ctx context.Context, userID model.UserID, limit int32) (int64, error),
) func(ctx context.Context, target *model.UserDeletionTarget) (bool, error) {
	return func(ctx context.Context, target *model.UserDeletionTarget) (bool, error) {
		count, err := deleteByUserID(ctx, target.ID, uc.batchSize)
		if err != nil {
			return false, err
		}
		return count < int64(uc.batchSize), nil
	}
}

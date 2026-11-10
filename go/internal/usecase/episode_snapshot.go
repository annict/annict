package usecase

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// readEpisodeAnimeSnapshotは、エピソードのanimeと分類を読み、書き込む前のスナップショットを組み立てる。
// エピソードは別表の行を持たないため、別表は空のままにする。animeの行が無いときはnilを返す
// (書き込みもそのanimeに届かないため、記録しない)。
func readEpisodeAnimeSnapshot(
	ctx context.Context,
	animeRepo *repository.AnimeRepository,
	classificationRepo *repository.AnimeClassificationRepository,
	animeID model.AnimeID,
) (*model.AnimeSnapshot, error) {
	anime, err := animeRepo.GetByID(ctx, animeID)
	if err != nil {
		return nil, fmt.Errorf("animeの取得に失敗: %w", err)
	}
	if anime == nil {
		return nil, nil
	}
	classification, err := classificationRepo.GetByAnimeID(ctx, animeID)
	if err != nil {
		return nil, fmt.Errorf("anime_classificationの取得に失敗: %w", err)
	}
	snapshot := animeSnapshotFromRows(anime, classification, workSatelliteExisting{})
	return &snapshot, nil
}

// episodeAnimeSnapshotFromCreateParamsは、作成するエピソードのanimeと分類の値から作成の版のスナップショットを
// 組み立てる。
func episodeAnimeSnapshotFromCreateParams(anime repository.CreateAnimeParams, classification repository.CreateAnimeClassificationParams) model.AnimeSnapshot {
	return animeSnapshotFromCreateParams(anime, classification, workSatellitePlans{})
}

// planEpisodeStatusRecordは、エピソードのanimeのstatusだけを書き換える書き込み (非公開・再公開・削除) の
// 記録の計画を立てる。書き込んだ後の状態は、書き込む前の状態のstatusを差し替えたもの。エピソードが
// animeにマッピングされていないときと、animeの行が無いときはnilを返す。
func planEpisodeStatusRecord(
	ctx context.Context,
	animeRepo *repository.AnimeRepository,
	classificationRepo *repository.AnimeClassificationRepository,
	dbEditRecordRepos DBEditRecordRepos,
	user *model.User,
	action model.DBEditAction,
	animeID *model.AnimeID,
	status model.AnimeStatus,
) (*dbEditRecordPlan, error) {
	if animeID == nil {
		return nil, nil
	}
	before, err := readEpisodeAnimeSnapshot(ctx, animeRepo, classificationRepo, *animeID)
	if err != nil {
		return nil, err
	}
	if before == nil {
		return nil, nil
	}
	after := *before
	after.Status = status
	return planDBEditRecord(ctx, dbEditRecordRepos, DBEditRecordInput{
		Actor:  model.DBEditActorUser,
		UserID: &user.ID,
		Action: action,
		Animes: []AnimeRevisionInput{{AnimeID: *animeID, Before: before, After: after}},
	})
}

// episodeAnimeIDChangedは、書き込んだ行が返したanimeの写像が、記録の計画を立てたときの写像と食い違うかを返す。
// 食い違うときは、トランザクションの前に組み立てた版が書き込んだanimeを表さないため、書き込みをやめる。
func episodeAnimeIDChanged(planned, written *model.AnimeID) bool {
	if planned == nil || written == nil {
		return planned != written
	}
	return *planned != *written
}

// episodeConflictErrorは、トランザクションの前に読んだエピソードが書き込むまでの間に変わったため、
// 書き込みをやめたことを伝える。
func episodeConflictError(ctx context.Context, episodeID model.EpisodeID) *model.AppError {
	return &model.AppError{
		Code:     model.AppErrCodeConflict,
		UserMsg:  i18n.T(ctx, "validation_version_conflict"),
		Metadata: map[string]string{"episode_id": episodeID.String()},
	}
}

// writeCreatedEpisodesDBEditRecordは、トランザクションの中で作ったエピソードの版を、1つの作成の操作として書く。
// 作成するエピソードの並び順と親作品はトランザクションの中でロックを取ってから決まるため、計画も
// トランザクションの中で立てる。作成の版は起点の版を持たないため、計画は何も読まない。
// revisionsが空のとき (作品がanimeにマッピングされていないとき) は何も書かない。
func writeCreatedEpisodesDBEditRecord(
	ctx context.Context,
	dbEditRecordRepos DBEditRecordRepos,
	tx *sql.Tx,
	actor model.DBEditActor,
	userID *model.UserID,
	revisions []AnimeRevisionInput,
) error {
	if len(revisions) == 0 {
		return nil
	}
	repos := dbEditRecordRepos.WithTx(tx)
	plan, err := planDBEditRecord(ctx, repos, DBEditRecordInput{
		Actor:  actor,
		UserID: userID,
		Action: model.DBEditActionCreate,
		Animes: revisions,
	})
	if err != nil {
		return fmt.Errorf("編集履歴の記録の計画に失敗しました: %w", err)
	}
	return writeDBEditRecord(ctx, repos, plan)
}

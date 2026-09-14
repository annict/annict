package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// AnimeExternalIDRepositoryはanime_external_idsテーブル (Syobocal /
// MyAnimeListなど外部データベースにおけるanimeのID) へのデータアクセスを担う。
type AnimeExternalIDRepository struct {
	queries *query.Queries
}

// NewAnimeExternalIDRepositoryはAnimeExternalIDRepositoryを生成する。
func NewAnimeExternalIDRepository(queries *query.Queries) *AnimeExternalIDRepository {
	return &AnimeExternalIDRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいAnimeExternalIDRepositoryを返す。
func (r *AnimeExternalIDRepository) WithTx(tx *sql.Tx) *AnimeExternalIDRepository {
	return &AnimeExternalIDRepository{queries: r.queries.WithTx(tx)}
}

// CreateAnimeExternalIDParamsは外部ID作成時の属性を保持する。idと
// タイムスタンプはデータベースが採番する。(anime_id, service) がUNIQUEインデックスで
// 守られる自然キー。
type CreateAnimeExternalIDParams struct {
	AnimeID    model.AnimeID
	Service    model.AnimeExternalService
	ExternalID string
}

// Createは新しい外部IDを挿入し、作成された行を返す。
func (r *AnimeExternalIDRepository) Create(ctx context.Context, params CreateAnimeExternalIDParams) (*model.AnimeExternalID, error) {
	row, err := r.queries.CreateAnimeExternalID(ctx, query.CreateAnimeExternalIDParams{
		AnimeID:    int64(params.AnimeID),
		Service:    query.AnimeExternalService(params.Service),
		ExternalID: params.ExternalID,
	})
	if err != nil {
		return nil, err
	}
	externalID := toAnimeExternalIDModel(row)
	return &externalID, nil
}

// UpdateAnimeExternalIDParamsは主キーで特定した外部IDの更新時の属性を保持する。
// 可変なのはexternal_idのみで、自然キー (anime_id, service) は固定のため、値が変わった
// サービスは削除と再作成ではなくその場で更新する。
type UpdateAnimeExternalIDParams struct {
	ID         model.AnimeExternalIDID
	ExternalID string
}

// Updateは指定行のexternal_idを上書きする。
func (r *AnimeExternalIDRepository) Update(ctx context.Context, params UpdateAnimeExternalIDParams) error {
	return r.queries.UpdateAnimeExternalID(ctx, query.UpdateAnimeExternalIDParams{
		ID:         int64(params.ID),
		ExternalID: params.ExternalID,
	})
}

// Deleteは指定主キーの外部IDを削除する。
func (r *AnimeExternalIDRepository) Delete(ctx context.Context, id model.AnimeExternalIDID) error {
	return r.queries.DeleteAnimeExternalID(ctx, int64(id))
}

// ListByAnimeIDsは指定anime ID群の外部IDを (anime_id, service) 順でロードする。
// フェーズ2の別表リコンシリエーションが、anime解決済みworksの1ページぶんの既存行を
// N回のanime単位ルックアップではなく1クエリで一括取得するために使う。空入力では
// クエリせず空スライスを返す。
func (r *AnimeExternalIDRepository) ListByAnimeIDs(ctx context.Context, animeIDs []model.AnimeID) ([]*model.AnimeExternalID, error) {
	if len(animeIDs) == 0 {
		return []*model.AnimeExternalID{}, nil
	}

	rawIDs := make([]int64, len(animeIDs))
	for i, id := range animeIDs {
		rawIDs[i] = int64(id)
	}

	rows, err := r.queries.ListAnimeExternalIDsByAnimeIDs(ctx, rawIDs)
	if err != nil {
		return nil, err
	}

	externalIDs := make([]*model.AnimeExternalID, len(rows))
	for i, row := range rows {
		externalID := toAnimeExternalIDModel(row)
		externalIDs[i] = &externalID
	}
	return externalIDs, nil
}

// toAnimeExternalIDModelはqueryの行をドメインモデルに変換する。
func toAnimeExternalIDModel(row query.AnimeExternalID) model.AnimeExternalID {
	return model.AnimeExternalID{
		ID:         model.AnimeExternalIDID(row.ID),
		AnimeID:    model.AnimeID(row.AnimeID),
		Service:    model.AnimeExternalService(row.Service),
		ExternalID: row.ExternalID,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}
}

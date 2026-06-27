package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// AnimeExternalIDRepository handles data access for the anime_external_ids table
// (an anime's IDs in external databases such as Syobocal / MyAnimeList).
//
// [Ja] AnimeExternalIDRepository は anime_external_ids テーブル (Syobocal /
// MyAnimeList など外部データベースにおける anime の ID) へのデータアクセスを担う。
type AnimeExternalIDRepository struct {
	queries *query.Queries
}

// NewAnimeExternalIDRepository constructs an AnimeExternalIDRepository.
//
// [Ja] NewAnimeExternalIDRepository は AnimeExternalIDRepository を生成する。
func NewAnimeExternalIDRepository(queries *query.Queries) *AnimeExternalIDRepository {
	return &AnimeExternalIDRepository{queries: queries}
}

// WithTx returns a new AnimeExternalIDRepository bound to the given transaction.
//
// [Ja] WithTx はトランザクションを使用する新しい AnimeExternalIDRepository を返す。
func (r *AnimeExternalIDRepository) WithTx(tx *sql.Tx) *AnimeExternalIDRepository {
	return &AnimeExternalIDRepository{queries: r.queries.WithTx(tx)}
}

// CreateAnimeExternalIDParams holds the attributes for creating an external ID.
// The id and timestamps are assigned by the database; (anime_id, service) is the
// natural key enforced by a UNIQUE index.
//
// [Ja] CreateAnimeExternalIDParams は外部 ID 作成時の属性を保持する。id と
// タイムスタンプはデータベースが採番する。(anime_id, service) が UNIQUE インデックスで
// 守られる自然キー。
type CreateAnimeExternalIDParams struct {
	AnimeID    model.AnimeID
	Service    model.AnimeExternalService
	ExternalID string
}

// Create inserts a new external ID and returns the created row.
//
// [Ja] Create は新しい外部 ID を挿入し、作成された行を返す。
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

// UpdateAnimeExternalIDParams holds the attributes for updating an external ID,
// identified by its primary key. Only external_id is mutable; the natural key
// (anime_id, service) is fixed, so a service whose value changed is updated in
// place rather than deleted and re-created.
//
// [Ja] UpdateAnimeExternalIDParams は主キーで特定した外部 ID の更新時の属性を保持する。
// 可変なのは external_id のみで、自然キー (anime_id, service) は固定のため、値が変わった
// サービスは削除と再作成ではなくその場で更新する。
type UpdateAnimeExternalIDParams struct {
	ID         model.AnimeExternalIDID
	ExternalID string
}

// Update overwrites the external_id of the identified row.
//
// [Ja] Update は指定行の external_id を上書きする。
func (r *AnimeExternalIDRepository) Update(ctx context.Context, params UpdateAnimeExternalIDParams) error {
	return r.queries.UpdateAnimeExternalID(ctx, query.UpdateAnimeExternalIDParams{
		ID:         int64(params.ID),
		ExternalID: params.ExternalID,
	})
}

// Delete removes the external ID with the given primary key.
//
// [Ja] Delete は指定主キーの外部 ID を削除する。
func (r *AnimeExternalIDRepository) Delete(ctx context.Context, id model.AnimeExternalIDID) error {
	return r.queries.DeleteAnimeExternalID(ctx, int64(id))
}

// ListByAnimeIDs loads the external IDs for the given anime IDs, ordered by
// (anime_id, service). It is used by the phase 2 satellite reconciliation to
// batch-fetch the existing rows for a page of anime-resolved works in one query
// instead of N per-anime lookups. An empty input returns an empty slice without
// querying.
//
// [Ja] ListByAnimeIDs は指定 anime ID 群の外部 ID を (anime_id, service) 順でロードする。
// フェーズ 2 の別表リコンシリエーションが、anime 解決済み works の 1 ページぶんの既存行を
// N 回の anime 単位ルックアップではなく 1 クエリで一括取得するために使う。空入力では
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

// toAnimeExternalIDModel converts a query row into the domain model.
//
// [Ja] toAnimeExternalIDModel は query の行をドメインモデルに変換する。
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

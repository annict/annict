package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// AnimeClassificationRepositoryはanime_classificationsテーブル
// (第2層: カタログ分類) へのデータアクセスを担う。
type AnimeClassificationRepository struct {
	queries *query.Queries
}

// NewAnimeClassificationRepositoryはAnimeClassificationRepositoryを生成する。
func NewAnimeClassificationRepository(queries *query.Queries) *AnimeClassificationRepository {
	return &AnimeClassificationRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいAnimeClassificationRepositoryを返す。
func (r *AnimeClassificationRepository) WithTx(tx *sql.Tx) *AnimeClassificationRepository {
	return &AnimeClassificationRepository{queries: r.queries.WithTx(tx)}
}

// CreateAnimeClassificationParamsは分類作成時の属性を保持する。work限定の
// フィールド (workではParentAnimeIDがNULL、episodeでは生成設定がNULL) は
// CHECK制約を満たす必要があり、整合した形での指定は呼び出し元の責務とする。
type CreateAnimeClassificationParams struct {
	AnimeID               model.AnimeID
	Kind                  model.AnimeClassificationKind
	ParentAnimeID         *model.AnimeID
	Number                sql.NullString
	NumberText            sql.NullString
	SortNumber            sql.NullInt32
	Standalone            bool
	NumberFormatID        *model.NumberFormatID
	EpisodeStartNumber    sql.NullString
	ExpectedEpisodesCount sql.NullInt32
}

// Createは新しい分類を挿入し、作成された行を返す。
func (r *AnimeClassificationRepository) Create(ctx context.Context, params CreateAnimeClassificationParams) (*model.AnimeClassification, error) {
	row, err := r.queries.CreateAnimeClassification(ctx, query.CreateAnimeClassificationParams{
		AnimeID:               int64(params.AnimeID),
		Kind:                  query.AnimeClassificationKind(params.Kind),
		ParentAnimeID:         nullInt64FromAnimeID(params.ParentAnimeID),
		Number:                params.Number,
		NumberText:            params.NumberText,
		SortNumber:            params.SortNumber,
		Standalone:            params.Standalone,
		NumberFormatID:        nullInt64FromNumberFormatID(params.NumberFormatID),
		EpisodeStartNumber:    params.EpisodeStartNumber,
		ExpectedEpisodesCount: params.ExpectedEpisodesCount,
	})
	if err != nil {
		return nil, err
	}
	classification := toAnimeClassificationModel(row)
	return &classification, nil
}

// Upsertは欠損した分類を作成し、anime_idで特定される既存行があれば上書きする。
// 1文で行うことで、存在確認と書き込みの間に並行削除が入り込む隙間を作らない。
func (r *AnimeClassificationRepository) Upsert(ctx context.Context, params CreateAnimeClassificationParams) error {
	return r.queries.UpsertAnimeClassification(ctx, query.UpsertAnimeClassificationParams{
		AnimeID:               int64(params.AnimeID),
		Kind:                  query.AnimeClassificationKind(params.Kind),
		ParentAnimeID:         nullInt64FromAnimeID(params.ParentAnimeID),
		Number:                params.Number,
		NumberText:            params.NumberText,
		SortNumber:            params.SortNumber,
		Standalone:            params.Standalone,
		NumberFormatID:        nullInt64FromNumberFormatID(params.NumberFormatID),
		EpisodeStartNumber:    params.EpisodeStartNumber,
		ExpectedEpisodesCount: params.ExpectedEpisodesCount,
	})
}

// UpdateAnimeClassificationParamsはanime_id (UNIQUE) で特定した分類の
// 更新時の属性を保持する。
type UpdateAnimeClassificationParams struct {
	AnimeID               model.AnimeID
	Kind                  model.AnimeClassificationKind
	ParentAnimeID         *model.AnimeID
	Number                sql.NullString
	NumberText            sql.NullString
	SortNumber            sql.NullInt32
	Standalone            bool
	NumberFormatID        *model.NumberFormatID
	EpisodeStartNumber    sql.NullString
	ExpectedEpisodesCount sql.NullInt32
}

// UpdateByAnimeIDは指定アニメの分類を上書きする。anime_idが自然キー
// (UNIQUE) であり、フェーズ2の同期はanime_idで行を解決してその場で更新する。
func (r *AnimeClassificationRepository) UpdateByAnimeID(ctx context.Context, params UpdateAnimeClassificationParams) error {
	return r.queries.UpdateAnimeClassificationByAnimeID(ctx, query.UpdateAnimeClassificationByAnimeIDParams{
		AnimeID:               int64(params.AnimeID),
		Kind:                  query.AnimeClassificationKind(params.Kind),
		ParentAnimeID:         nullInt64FromAnimeID(params.ParentAnimeID),
		Number:                params.Number,
		NumberText:            params.NumberText,
		SortNumber:            params.SortNumber,
		Standalone:            params.Standalone,
		NumberFormatID:        nullInt64FromNumberFormatID(params.NumberFormatID),
		EpisodeStartNumber:    params.EpisodeStartNumber,
		ExpectedEpisodesCount: params.ExpectedEpisodesCount,
	})
}

// GetByAnimeIDは指定アニメの分類を検索する。該当行が無い場合は (nil, nil)
// を返し、sql.ErrNoRowsをRepositoryの外へ漏らさない。
func (r *AnimeClassificationRepository) GetByAnimeID(ctx context.Context, animeID model.AnimeID) (*model.AnimeClassification, error) {
	row, err := r.queries.GetAnimeClassificationByAnimeID(ctx, int64(animeID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	classification := toAnimeClassificationModel(row)
	return &classification, nil
}

// ListByAnimeIDsは指定anime ID群の分類をanime_id昇順でロードする。
// フェーズ2のリコンシリエーションが、マッピング済みworksの1ページぶんの既存分類を
// N回の行単位ルックアップではなく1クエリで一括取得するために使う。
// 空入力ではクエリせず空スライスを返す。
func (r *AnimeClassificationRepository) ListByAnimeIDs(ctx context.Context, animeIDs []model.AnimeID) ([]*model.AnimeClassification, error) {
	if len(animeIDs) == 0 {
		return []*model.AnimeClassification{}, nil
	}

	rawIDs := make([]int64, len(animeIDs))
	for i, id := range animeIDs {
		rawIDs[i] = int64(id)
	}

	rows, err := r.queries.ListAnimeClassificationsByAnimeIDs(ctx, rawIDs)
	if err != nil {
		return nil, err
	}

	classifications := make([]*model.AnimeClassification, len(rows))
	for i, row := range rows {
		classification := toAnimeClassificationModel(row)
		classifications[i] = &classification
	}
	return classifications, nil
}

// toAnimeClassificationModelはqueryの行をドメインモデルに変換する。
func toAnimeClassificationModel(row query.AnimeClassification) model.AnimeClassification {
	classification := model.AnimeClassification{
		ID:                    model.AnimeClassificationID(row.ID),
		AnimeID:               model.AnimeID(row.AnimeID),
		Kind:                  model.AnimeClassificationKind(row.Kind),
		Number:                row.Number,
		NumberText:            row.NumberText,
		SortNumber:            row.SortNumber,
		Standalone:            row.Standalone,
		EpisodeStartNumber:    row.EpisodeStartNumber,
		ExpectedEpisodesCount: row.ExpectedEpisodesCount,
		CreatedAt:             row.CreatedAt,
		UpdatedAt:             row.UpdatedAt,
	}
	if row.ParentAnimeID.Valid {
		parentAnimeID := model.AnimeID(row.ParentAnimeID.Int64)
		classification.ParentAnimeID = &parentAnimeID
	}
	if row.NumberFormatID.Valid {
		numberFormatID := model.NumberFormatID(row.NumberFormatID.Int64)
		classification.NumberFormatID = &numberFormatID
	}
	return classification
}

// nullInt64FromAnimeIDは任意のアニメID外部キーをsqlcのNULL許容intに
// 写像する。
func nullInt64FromAnimeID(id *model.AnimeID) sql.NullInt64 {
	if id == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(*id), Valid: true}
}

// nullInt64FromNumberFormatIDは任意のnumber_format ID外部キーをsqlcの
// NULL許容intに写像する。
func nullInt64FromNumberFormatID(id *model.NumberFormatID) sql.NullInt64 {
	if id == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(*id), Valid: true}
}

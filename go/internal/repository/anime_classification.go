package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// AnimeClassificationRepository handles data access for the
// anime_classifications table (layer 2: catalog classification).
//
// [Ja] AnimeClassificationRepository は anime_classifications テーブル
// (第 2 層: カタログ分類) へのデータアクセスを担う。
type AnimeClassificationRepository struct {
	queries *query.Queries
}

// NewAnimeClassificationRepository constructs an AnimeClassificationRepository.
//
// [Ja] NewAnimeClassificationRepository は AnimeClassificationRepository を生成する。
func NewAnimeClassificationRepository(queries *query.Queries) *AnimeClassificationRepository {
	return &AnimeClassificationRepository{queries: queries}
}

// WithTx returns a new AnimeClassificationRepository bound to the given
// transaction.
//
// [Ja] WithTx はトランザクションを使用する新しい AnimeClassificationRepository を返す。
func (r *AnimeClassificationRepository) WithTx(tx *sql.Tx) *AnimeClassificationRepository {
	return &AnimeClassificationRepository{queries: r.queries.WithTx(tx)}
}

// CreateAnimeClassificationParams holds the attributes for creating a
// classification. The work-only fields (ParentAnimeID is NULL for a work, the
// generation settings are NULL for an episode) must satisfy the CHECK
// constraints; the caller is responsible for supplying a consistent shape.
//
// [Ja] CreateAnimeClassificationParams は分類作成時の属性を保持する。work 限定の
// フィールド (work では ParentAnimeID が NULL、episode では生成設定が NULL) は
// CHECK 制約を満たす必要があり、整合した形での指定は呼び出し元の責務とする。
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

// Create inserts a new classification and returns the created row.
//
// [Ja] Create は新しい分類を挿入し、作成された行を返す。
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

// UpdateAnimeClassificationParams holds the attributes for updating a
// classification, identified by its anime_id (UNIQUE).
//
// [Ja] UpdateAnimeClassificationParams は anime_id (UNIQUE) で特定した分類の
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

// UpdateByAnimeID overwrites the classification of the given anime. The anime_id
// is the natural key (UNIQUE), so the phase 2 sync resolves a row by anime_id
// and updates it in place.
//
// [Ja] UpdateByAnimeID は指定アニメの分類を上書きする。anime_id が自然キー
// (UNIQUE) であり、フェーズ 2 の同期は anime_id で行を解決してその場で更新する。
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

// GetByAnimeID looks up the classification of the given anime. It returns
// (nil, nil) when no row matches, keeping sql.ErrNoRows from leaking out of the
// repository.
//
// [Ja] GetByAnimeID は指定アニメの分類を検索する。該当行が無い場合は (nil, nil)
// を返し、sql.ErrNoRows を Repository の外へ漏らさない。
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

// ListByAnimeIDs loads the classifications for the given anime IDs, ordered by
// anime_id. It is used by the phase 2 reconciliation to batch-fetch the existing
// classifications for a page of mapped works in one query instead of N per-row
// lookups. An empty input returns an empty slice without querying.
//
// [Ja] ListByAnimeIDs は指定 anime ID 群の分類を anime_id 昇順でロードする。
// フェーズ 2 のリコンシリエーションが、マッピング済み works の 1 ページぶんの既存分類を
// N 回の行単位ルックアップではなく 1 クエリで一括取得するために使う。
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

// toAnimeClassificationModel converts a query row into the domain model.
//
// [Ja] toAnimeClassificationModel は query の行をドメインモデルに変換する。
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

// nullInt64FromAnimeID maps an optional anime ID FK to sqlc's nullable int.
//
// [Ja] nullInt64FromAnimeID は任意のアニメ ID 外部キーを sqlc の NULL 許容 int に
// 写像する。
func nullInt64FromAnimeID(id *model.AnimeID) sql.NullInt64 {
	if id == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(*id), Valid: true}
}

// nullInt64FromNumberFormatID maps an optional number-format ID FK to sqlc's
// nullable int.
//
// [Ja] nullInt64FromNumberFormatID は任意の number_format ID 外部キーを sqlc の
// NULL 許容 int に写像する。
func nullInt64FromNumberFormatID(id *model.NumberFormatID) sql.NullInt64 {
	if id == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(*id), Valid: true}
}

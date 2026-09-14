package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// AnimeSeasonRepositoryはanime_seasonsテーブル (animeが掲載される季節。例: 2024
// spring) へのデータアクセスを担う。
type AnimeSeasonRepository struct {
	queries *query.Queries
}

// NewAnimeSeasonRepositoryはAnimeSeasonRepositoryを生成する。
func NewAnimeSeasonRepository(queries *query.Queries) *AnimeSeasonRepository {
	return &AnimeSeasonRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいAnimeSeasonRepositoryを返す。
func (r *AnimeSeasonRepository) WithTx(tx *sql.Tx) *AnimeSeasonRepository {
	return &AnimeSeasonRepository{queries: r.queries.WithTx(tx)}
}

// CreateAnimeSeasonParamsは季節作成時の属性を保持する。idとタイムスタンプは
// データベースが採番する。Nameは季節名が未定 (年のみ判明) のときnil。IsPrimaryは
// worksがsourceする行がis_primary=trueでなければならない一方で列の既定値がfalseの
// ため、明示的に渡す。自然キー (anime_id, year, name) はUNIQUEインデックス (NULLS NOT
// DISTINCT) で守られ、部分UNIQUEインデックスがanimeごとにis_primary行を高々1つに保つ。
type CreateAnimeSeasonParams struct {
	AnimeID   model.AnimeID
	Year      int32
	Name      *model.SeasonName
	IsPrimary bool
}

// Createは新しい季節を挿入し、作成された行を返す。
func (r *AnimeSeasonRepository) Create(ctx context.Context, params CreateAnimeSeasonParams) (*model.AnimeSeason, error) {
	row, err := r.queries.CreateAnimeSeason(ctx, query.CreateAnimeSeasonParams{
		AnimeID:   int64(params.AnimeID),
		Year:      params.Year,
		Name:      nullSeasonNameFromModel(params.Name),
		IsPrimary: params.IsPrimary,
	})
	if err != nil {
		return nil, err
	}
	season := toAnimeSeasonModel(row)
	return &season, nil
}

// Updateは意図的に設けない。worksはyearとname (どちらも自然キー) に加え
// is_primary (works管理スロットではtrue固定) をsourceするため、行をその場で書き換える
// ことがない。季節の変更は旧行の削除と新行の作成になる (SyncAnimeSeasonsUsecaseを参照)。

// Deleteは指定主キーの季節を削除する。
func (r *AnimeSeasonRepository) Delete(ctx context.Context, id model.AnimeSeasonID) error {
	return r.queries.DeleteAnimeSeason(ctx, int64(id))
}

// ListByAnimeIDsは指定anime ID群の季節を (anime_id, year, name) 順でロードする。
// フェーズ2の別表リコンシリエーションが、anime解決済みworksの1ページぶんの既存行を
// N回のanime単位ルックアップではなく1クエリで一括取得するために使う。空入力では
// クエリせず空スライスを返す。
func (r *AnimeSeasonRepository) ListByAnimeIDs(ctx context.Context, animeIDs []model.AnimeID) ([]*model.AnimeSeason, error) {
	if len(animeIDs) == 0 {
		return []*model.AnimeSeason{}, nil
	}

	rawIDs := make([]int64, len(animeIDs))
	for i, id := range animeIDs {
		rawIDs[i] = int64(id)
	}

	rows, err := r.queries.ListAnimeSeasonsByAnimeIDs(ctx, rawIDs)
	if err != nil {
		return nil, err
	}

	seasons := make([]*model.AnimeSeason, len(rows))
	for i, row := range rows {
		season := toAnimeSeasonModel(row)
		seasons[i] = &season
	}
	return seasons, nil
}

// nullSeasonNameFromModelはNULL許容のモデル季節名をsqlcのNullSeasonNameに
// 変換する。nilはNULL (季節名未定) になる。
func nullSeasonNameFromModel(name *model.SeasonName) query.NullSeasonName {
	if name == nil {
		return query.NullSeasonName{}
	}
	return query.NullSeasonName{SeasonName: query.SeasonName(*name), Valid: true}
}

// toAnimeSeasonModelはqueryの行をドメインモデルに変換する。
func toAnimeSeasonModel(row query.AnimeSeason) model.AnimeSeason {
	season := model.AnimeSeason{
		ID:        model.AnimeSeasonID(row.ID),
		AnimeID:   model.AnimeID(row.AnimeID),
		Year:      row.Year,
		IsPrimary: row.IsPrimary,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
	if row.Name.Valid {
		name := model.SeasonName(row.Name.SeasonName)
		season.Name = &name
	}
	return season
}

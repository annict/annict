package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// AnimeSeasonRepository handles data access for the anime_seasons table (the seasons an
// anime is listed in, e.g. 2024 spring).
//
// [Ja] AnimeSeasonRepository は anime_seasons テーブル (anime が掲載される季節。例: 2024
// spring) へのデータアクセスを担う。
type AnimeSeasonRepository struct {
	queries *query.Queries
}

// NewAnimeSeasonRepository constructs an AnimeSeasonRepository.
//
// [Ja] NewAnimeSeasonRepository は AnimeSeasonRepository を生成する。
func NewAnimeSeasonRepository(queries *query.Queries) *AnimeSeasonRepository {
	return &AnimeSeasonRepository{queries: queries}
}

// WithTx returns a new AnimeSeasonRepository bound to the given transaction.
//
// [Ja] WithTx はトランザクションを使用する新しい AnimeSeasonRepository を返す。
func (r *AnimeSeasonRepository) WithTx(tx *sql.Tx) *AnimeSeasonRepository {
	return &AnimeSeasonRepository{queries: r.queries.WithTx(tx)}
}

// CreateAnimeSeasonParams holds the attributes for creating a season. The id and
// timestamps are assigned by the database. Name is nil when the season name is
// undetermined (only the year is known). IsPrimary is passed explicitly because the
// works-sourced row must be is_primary=true while the column defaults to false. The
// natural key (anime_id, year, name) is enforced by a UNIQUE index (NULLS NOT
// DISTINCT), and a partial UNIQUE index keeps at most one is_primary row per anime.
//
// [Ja] CreateAnimeSeasonParams は季節作成時の属性を保持する。id とタイムスタンプは
// データベースが採番する。Name は季節名が未定 (年のみ判明) のとき nil。IsPrimary は
// works が source する行が is_primary=true でなければならない一方で列の既定値が false の
// ため、明示的に渡す。自然キー (anime_id, year, name) は UNIQUE インデックス (NULLS NOT
// DISTINCT) で守られ、部分 UNIQUE インデックスが anime ごとに is_primary 行を高々 1 つに保つ。
type CreateAnimeSeasonParams struct {
	AnimeID   model.AnimeID
	Year      int32
	Name      *model.SeasonName
	IsPrimary bool
}

// Create inserts a new season and returns the created row.
//
// [Ja] Create は新しい季節を挿入し、作成された行を返す。
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

// There is intentionally no Update: works source the year and name (both in the natural
// key) plus is_primary (fixed true for the works-managed slot), so a row is never
// mutated in place — a changed season is a delete of the old row plus a create of the
// new one (see SyncAnimeSeasonsUsecase).
//
// [Ja] Update は意図的に設けない。works は year と name (どちらも自然キー) に加え
// is_primary (works 管理スロットでは true 固定) を source するため、行をその場で書き換える
// ことがない。季節の変更は旧行の削除と新行の作成になる (SyncAnimeSeasonsUsecase を参照)。

// Delete removes the season with the given primary key.
//
// [Ja] Delete は指定主キーの季節を削除する。
func (r *AnimeSeasonRepository) Delete(ctx context.Context, id model.AnimeSeasonID) error {
	return r.queries.DeleteAnimeSeason(ctx, int64(id))
}

// ListByAnimeIDs loads the seasons for the given anime IDs, ordered by
// (anime_id, year, name). It is used by the phase 2 satellite reconciliation to
// batch-fetch the existing rows for a page of anime-resolved works in one query instead
// of N per-anime lookups. An empty input returns an empty slice without querying.
//
// [Ja] ListByAnimeIDs は指定 anime ID 群の季節を (anime_id, year, name) 順でロードする。
// フェーズ 2 の別表リコンシリエーションが、anime 解決済み works の 1 ページぶんの既存行を
// N 回の anime 単位ルックアップではなく 1 クエリで一括取得するために使う。空入力では
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

// nullSeasonNameFromModel converts a nullable model season name into the sqlc
// NullSeasonName: nil becomes NULL (an undetermined season name).
//
// [Ja] nullSeasonNameFromModel は NULL 許容のモデル季節名を sqlc の NullSeasonName に
// 変換する。nil は NULL (季節名未定) になる。
func nullSeasonNameFromModel(name *model.SeasonName) query.NullSeasonName {
	if name == nil {
		return query.NullSeasonName{}
	}
	return query.NullSeasonName{SeasonName: query.SeasonName(*name), Valid: true}
}

// toAnimeSeasonModel converts a query row into the domain model.
//
// [Ja] toAnimeSeasonModel は query の行をドメインモデルに変換する。
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

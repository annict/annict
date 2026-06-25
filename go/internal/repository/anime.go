package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// AnimeRepository handles data access for the animes table (layer 1: content
// identity).
//
// [Ja] AnimeRepository は animes テーブル (第 1 層: コンテンツ同一性) への
// データアクセスを担う。
type AnimeRepository struct {
	queries *query.Queries
}

// NewAnimeRepository constructs an AnimeRepository.
//
// [Ja] NewAnimeRepository は AnimeRepository を生成する。
func NewAnimeRepository(queries *query.Queries) *AnimeRepository {
	return &AnimeRepository{queries: queries}
}

// WithTx returns a new AnimeRepository bound to the given transaction.
//
// [Ja] WithTx はトランザクションを使用する新しい AnimeRepository を返す。
func (r *AnimeRepository) WithTx(tx *sql.Tx) *AnimeRepository {
	return &AnimeRepository{queries: r.queries.WithTx(tx)}
}

// CreateAnimeParams holds the content attributes for creating an anime. The id
// and timestamps are assigned by the database.
//
// [Ja] CreateAnimeParams はアニメ作成時の内容属性を保持する。id とタイムスタンプは
// データベースが採番する。
type CreateAnimeParams struct {
	Title            sql.NullString
	TitleKana        sql.NullString
	TitleRo          sql.NullString
	TitleEn          sql.NullString
	TitleAlter       sql.NullString
	TitleAlterRo     sql.NullString
	TitleAlterEn     sql.NullString
	TitleAlterOther  sql.NullString
	Media            model.AnimeMedia
	ReleaseStatus    model.ReleaseStatus
	Synopsis         sql.NullString
	SynopsisEn       sql.NullString
	SynopsisSource   sql.NullString
	SynopsisSourceEn sql.NullString
	Status           model.AnimeStatus
	ArchiveMessage   sql.NullString
}

// Create inserts a new anime and returns the created row.
//
// [Ja] Create は新しいアニメを挿入し、作成された行を返す。
func (r *AnimeRepository) Create(ctx context.Context, params CreateAnimeParams) (*model.Anime, error) {
	row, err := r.queries.CreateAnime(ctx, query.CreateAnimeParams{
		Title:            params.Title,
		TitleKana:        params.TitleKana,
		TitleRo:          params.TitleRo,
		TitleEn:          params.TitleEn,
		TitleAlter:       params.TitleAlter,
		TitleAlterRo:     params.TitleAlterRo,
		TitleAlterEn:     params.TitleAlterEn,
		TitleAlterOther:  params.TitleAlterOther,
		Media:            toQueryNullAnimeMedia(params.Media),
		ReleaseStatus:    toQueryNullReleaseStatus(params.ReleaseStatus),
		Synopsis:         params.Synopsis,
		SynopsisEn:       params.SynopsisEn,
		SynopsisSource:   params.SynopsisSource,
		SynopsisSourceEn: params.SynopsisSourceEn,
		Status:           toQueryAnimeStatus(params.Status),
		ArchiveMessage:   params.ArchiveMessage,
	})
	if err != nil {
		return nil, err
	}
	anime := toAnimeModel(row)
	return &anime, nil
}

// UpdateAnimeParams holds the content attributes for updating an anime,
// identified by ID.
//
// [Ja] UpdateAnimeParams は ID で特定したアニメの更新時の内容属性を保持する。
type UpdateAnimeParams struct {
	ID               model.AnimeID
	Title            sql.NullString
	TitleKana        sql.NullString
	TitleRo          sql.NullString
	TitleEn          sql.NullString
	TitleAlter       sql.NullString
	TitleAlterRo     sql.NullString
	TitleAlterEn     sql.NullString
	TitleAlterOther  sql.NullString
	Media            model.AnimeMedia
	ReleaseStatus    model.ReleaseStatus
	Synopsis         sql.NullString
	SynopsisEn       sql.NullString
	SynopsisSource   sql.NullString
	SynopsisSourceEn sql.NullString
	Status           model.AnimeStatus
	ArchiveMessage   sql.NullString
}

// Update overwrites an anime's content attributes.
//
// [Ja] Update はアニメの内容属性を上書きする。
func (r *AnimeRepository) Update(ctx context.Context, params UpdateAnimeParams) error {
	return r.queries.UpdateAnime(ctx, query.UpdateAnimeParams{
		ID:               int64(params.ID),
		Title:            params.Title,
		TitleKana:        params.TitleKana,
		TitleRo:          params.TitleRo,
		TitleEn:          params.TitleEn,
		TitleAlter:       params.TitleAlter,
		TitleAlterRo:     params.TitleAlterRo,
		TitleAlterEn:     params.TitleAlterEn,
		TitleAlterOther:  params.TitleAlterOther,
		Media:            toQueryNullAnimeMedia(params.Media),
		ReleaseStatus:    toQueryNullReleaseStatus(params.ReleaseStatus),
		Synopsis:         params.Synopsis,
		SynopsisEn:       params.SynopsisEn,
		SynopsisSource:   params.SynopsisSource,
		SynopsisSourceEn: params.SynopsisSourceEn,
		Status:           toQueryAnimeStatus(params.Status),
		ArchiveMessage:   params.ArchiveMessage,
	})
}

// GetByID looks up an anime by its ID. It returns (nil, nil) when no row
// matches, keeping sql.ErrNoRows from leaking out of the repository.
//
// [Ja] GetByID は ID でアニメを検索する。該当行が無い場合は (nil, nil) を返し、
// sql.ErrNoRows を Repository の外へ漏らさない。
func (r *AnimeRepository) GetByID(ctx context.Context, id model.AnimeID) (*model.Anime, error) {
	row, err := r.queries.GetAnimeByID(ctx, int64(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	anime := toAnimeModel(row)
	return &anime, nil
}

// ListByIDs loads the animes with the given IDs, ordered by id. It is used by the
// phase 2 reconciliation to batch-fetch the existing animes for a page of mapped
// works in one query instead of N per-row lookups. An empty input returns an empty
// slice without querying.
//
// [Ja] ListByIDs は指定 ID の animes を id 昇順でロードする。フェーズ 2 の
// リコンシリエーションが、マッピング済み works の 1 ページぶんの既存 anime を
// N 回の行単位ルックアップではなく 1 クエリで一括取得するために使う。
// 空入力ではクエリせず空スライスを返す。
func (r *AnimeRepository) ListByIDs(ctx context.Context, ids []model.AnimeID) ([]*model.Anime, error) {
	if len(ids) == 0 {
		return []*model.Anime{}, nil
	}

	rawIDs := make([]int64, len(ids))
	for i, id := range ids {
		rawIDs[i] = int64(id)
	}

	rows, err := r.queries.ListAnimesByIDs(ctx, rawIDs)
	if err != nil {
		return nil, err
	}

	animes := make([]*model.Anime, len(rows))
	for i, row := range rows {
		anime := toAnimeModel(row)
		animes[i] = &anime
	}
	return animes, nil
}

// toAnimeModel converts a query row into the domain model. Nullable enum
// columns map to the empty string when NULL.
//
// [Ja] toAnimeModel は query の行をドメインモデルに変換する。NULL 許容の enum
// カラムは NULL のとき空文字列に写像する。
func toAnimeModel(row query.Anime) model.Anime {
	anime := model.Anime{
		ID:               model.AnimeID(row.ID),
		Title:            row.Title,
		TitleKana:        row.TitleKana,
		TitleRo:          row.TitleRo,
		TitleEn:          row.TitleEn,
		TitleAlter:       row.TitleAlter,
		TitleAlterRo:     row.TitleAlterRo,
		TitleAlterEn:     row.TitleAlterEn,
		TitleAlterOther:  row.TitleAlterOther,
		Synopsis:         row.Synopsis,
		SynopsisEn:       row.SynopsisEn,
		SynopsisSource:   row.SynopsisSource,
		SynopsisSourceEn: row.SynopsisSourceEn,
		Status:           model.AnimeStatus(row.Status),
		ArchiveMessage:   row.ArchiveMessage,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}
	if row.Media.Valid {
		anime.Media = model.AnimeMedia(row.Media.AnimeMedia)
	}
	if row.ReleaseStatus.Valid {
		anime.ReleaseStatus = model.ReleaseStatus(row.ReleaseStatus.ReleaseStatus)
	}
	return anime
}

// toQueryNullAnimeMedia maps the domain medium to the sqlc nullable enum,
// treating the empty string as NULL.
//
// [Ja] toQueryNullAnimeMedia はドメインの媒体を sqlc の NULL 許容 enum に写像し、
// 空文字列を NULL として扱う。
func toQueryNullAnimeMedia(m model.AnimeMedia) query.NullAnimeMedia {
	if m == "" {
		return query.NullAnimeMedia{}
	}
	return query.NullAnimeMedia{AnimeMedia: query.AnimeMedia(m), Valid: true}
}

// toQueryNullReleaseStatus maps the domain release status to the sqlc nullable
// enum, treating the empty string as NULL.
//
// [Ja] toQueryNullReleaseStatus はドメインの公開ステータスを sqlc の NULL 許容
// enum に写像し、空文字列を NULL として扱う。
func toQueryNullReleaseStatus(s model.ReleaseStatus) query.NullReleaseStatus {
	if s == "" {
		return query.NullReleaseStatus{}
	}
	return query.NullReleaseStatus{ReleaseStatus: query.ReleaseStatus(s), Valid: true}
}

// toQueryAnimeStatus maps the domain status to the sqlc enum, defaulting the
// empty string to 'published' to mirror the column default.
//
// [Ja] toQueryAnimeStatus はドメインのステータスを sqlc の enum に写像し、
// カラムの既定値に合わせて空文字列を 'published' に既定する。
func toQueryAnimeStatus(s model.AnimeStatus) query.AnimeStatus {
	if s == "" {
		return query.AnimeStatusPublished
	}
	return query.AnimeStatus(s)
}

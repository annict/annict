package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// AnimeRepositoryはanimesテーブル (第1層: コンテンツ同一性) への
// データアクセスを担う。
type AnimeRepository struct {
	queries *query.Queries
}

// NewAnimeRepositoryはAnimeRepositoryを生成する。
func NewAnimeRepository(queries *query.Queries) *AnimeRepository {
	return &AnimeRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいAnimeRepositoryを返す。
func (r *AnimeRepository) WithTx(tx *sql.Tx) *AnimeRepository {
	return &AnimeRepository{queries: r.queries.WithTx(tx)}
}

// CreateAnimeParamsはアニメ作成時の内容属性を保持する。idとタイムスタンプは
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

// Createは新しいアニメを挿入し、作成された行を返す。
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

// UpdateAnimeParamsはIDで特定したアニメの更新時の内容属性を保持する。
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

// Updateはアニメの内容属性を上書きする。
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

// UpdateStatusはanimeのライフサイクル状態だけを更新し、内容属性をすべて保持する。
// 非公開の経路が古い行全体のスナップショットを競合した編集へ上書きしないために使う。
func (r *AnimeRepository) UpdateStatus(ctx context.Context, id model.AnimeID, status model.AnimeStatus) error {
	return r.queries.UpdateAnimeStatus(ctx, query.UpdateAnimeStatusParams{
		ID:     int64(id),
		Status: toQueryAnimeStatus(status),
	})
}

// GetByIDはIDでアニメを検索する。該当行が無い場合は (nil, nil) を返し、
// sql.ErrNoRowsをRepositoryの外へ漏らさない。
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

// ListByIDsは指定IDのanimesをid昇順でロードする。フェーズ2の
// リコンシリエーションが、マッピング済みworksの1ページぶんの既存animeを
// N回の行単位ルックアップではなく1クエリで一括取得するために使う。
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

// toAnimeModelはqueryの行をドメインモデルに変換する。NULL許容のenum
// カラムはNULLのとき空文字列に写像する。
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

// toQueryNullAnimeMediaはドメインの媒体をsqlcのNULL許容enumに写像し、
// 空文字列をNULLとして扱う。
func toQueryNullAnimeMedia(m model.AnimeMedia) query.NullAnimeMedia {
	if m == "" {
		return query.NullAnimeMedia{}
	}
	return query.NullAnimeMedia{AnimeMedia: query.AnimeMedia(m), Valid: true}
}

// toQueryNullReleaseStatusはドメインの公開ステータスをsqlcのNULL許容
// enumに写像し、空文字列をNULLとして扱う。
func toQueryNullReleaseStatus(s model.ReleaseStatus) query.NullReleaseStatus {
	if s == "" {
		return query.NullReleaseStatus{}
	}
	return query.NullReleaseStatus{ReleaseStatus: query.ReleaseStatus(s), Valid: true}
}

// toQueryAnimeStatusはドメインのステータスをsqlcのenumに写像し、
// カラムの既定値に合わせて空文字列を 'published' に既定する。
func toQueryAnimeStatus(s model.AnimeStatus) query.AnimeStatus {
	if s == "" {
		return query.AnimeStatusPublished
	}
	return query.AnimeStatus(s)
}

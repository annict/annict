package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// AnimeEventRepositoryはanime_eventsテーブル (animeのカレンダーイベント。例: 放送
// 期間) へのデータアクセスを担う。
type AnimeEventRepository struct {
	queries *query.Queries
}

// NewAnimeEventRepositoryはAnimeEventRepositoryを生成する。
func NewAnimeEventRepository(queries *query.Queries) *AnimeEventRepository {
	return &AnimeEventRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいAnimeEventRepositoryを返す。
func (r *AnimeEventRepository) WithTx(tx *sql.Tx) *AnimeEventRepository {
	return &AnimeEventRepository{queries: r.queries.WithTx(tx)}
}

// CreateAnimeEventParamsはイベント作成時の属性を保持する。id / title / title_en /
// description / description_en / sort_number / タイムスタンプはデータベースが採番する
// (テキスト列はNULL、sort_numberは0が既定値)。worksは放送期間のみをsourceするため、
// 自然キー (anime_id, kind) と日付だけを持つ。EndedOnはイベントの終了が未定のときnil。
// (anime_id, kind) がUNIQUEインデックスで守られる自然キー。
type CreateAnimeEventParams struct {
	AnimeID   model.AnimeID
	Kind      model.AnimeEventKind
	StartedOn time.Time
	EndedOn   *time.Time
}

// Createは新しいイベントを挿入し、作成された行を返す。
func (r *AnimeEventRepository) Create(ctx context.Context, params CreateAnimeEventParams) (*model.AnimeEvent, error) {
	row, err := r.queries.CreateAnimeEvent(ctx, query.CreateAnimeEventParams{
		AnimeID:   int64(params.AnimeID),
		Kind:      query.AnimeEventKind(params.Kind),
		StartedOn: params.StartedOn,
		EndedOn:   nullTimeFromPtr(params.EndedOn),
	})
	if err != nil {
		return nil, err
	}
	event := toAnimeEventModel(row)
	return &event, nil
}

// UpdateAnimeEventParamsは主キーで特定したイベントの更新時の属性を保持する。可変なのは
// started_on / ended_onのみで、自然キー (anime_id, kind) は固定、sourceしないtitle /
// description / sort_number列は保全するため、放送期間が変わったイベントは削除と再作成では
// なくその場で更新する。
type UpdateAnimeEventParams struct {
	ID        model.AnimeEventID
	StartedOn time.Time
	EndedOn   *time.Time
}

// Updateは指定行のstarted_on / ended_onを上書きする。
func (r *AnimeEventRepository) Update(ctx context.Context, params UpdateAnimeEventParams) error {
	return r.queries.UpdateAnimeEvent(ctx, query.UpdateAnimeEventParams{
		ID:        int64(params.ID),
		StartedOn: params.StartedOn,
		EndedOn:   nullTimeFromPtr(params.EndedOn),
	})
}

// Deleteは指定主キーのイベントを削除する。
func (r *AnimeEventRepository) Delete(ctx context.Context, id model.AnimeEventID) error {
	return r.queries.DeleteAnimeEvent(ctx, int64(id))
}

// ListByAnimeIDsは指定anime ID群のイベントを (anime_id, kind) 順でロードする。
// フェーズ2の別表リコンシリエーションが、anime解決済みworksの1ページぶんの既存行を
// N回のanime単位ルックアップではなく1クエリで一括取得するために使う。空入力では
// クエリせず空スライスを返す。
func (r *AnimeEventRepository) ListByAnimeIDs(ctx context.Context, animeIDs []model.AnimeID) ([]*model.AnimeEvent, error) {
	if len(animeIDs) == 0 {
		return []*model.AnimeEvent{}, nil
	}

	rawIDs := make([]int64, len(animeIDs))
	for i, id := range animeIDs {
		rawIDs[i] = int64(id)
	}

	rows, err := r.queries.ListAnimeEventsByAnimeIDs(ctx, rawIDs)
	if err != nil {
		return nil, err
	}

	events := make([]*model.AnimeEvent, len(rows))
	for i, row := range rows {
		event := toAnimeEventModel(row)
		events[i] = &event
	}
	return events, nil
}

// nullTimeFromPtrはNULL許容の日付ポインタをsqlcのNullTimeに変換する。nilは
// NULL (終了未定・不明) になる。
func nullTimeFromPtr(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

// toAnimeEventModelはqueryの行をドメインモデルに変換する。
func toAnimeEventModel(row query.AnimeEvent) model.AnimeEvent {
	event := model.AnimeEvent{
		ID:         model.AnimeEventID(row.ID),
		AnimeID:    model.AnimeID(row.AnimeID),
		Kind:       model.AnimeEventKind(row.Kind),
		StartedOn:  row.StartedOn,
		SortNumber: row.SortNumber,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}
	if row.EndedOn.Valid {
		endedOn := row.EndedOn.Time
		event.EndedOn = &endedOn
	}
	if row.Title.Valid {
		title := row.Title.String
		event.Title = &title
	}
	if row.TitleEn.Valid {
		titleEn := row.TitleEn.String
		event.TitleEn = &titleEn
	}
	if row.Description.Valid {
		description := row.Description.String
		event.Description = &description
	}
	if row.DescriptionEn.Valid {
		descriptionEn := row.DescriptionEn.String
		event.DescriptionEn = &descriptionEn
	}
	return event
}

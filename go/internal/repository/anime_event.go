package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// AnimeEventRepository handles data access for the anime_events table (an anime's
// calendar events, e.g. its broadcast period).
//
// [Ja] AnimeEventRepository は anime_events テーブル (anime のカレンダーイベント。例: 放送
// 期間) へのデータアクセスを担う。
type AnimeEventRepository struct {
	queries *query.Queries
}

// NewAnimeEventRepository constructs an AnimeEventRepository.
//
// [Ja] NewAnimeEventRepository は AnimeEventRepository を生成する。
func NewAnimeEventRepository(queries *query.Queries) *AnimeEventRepository {
	return &AnimeEventRepository{queries: queries}
}

// WithTx returns a new AnimeEventRepository bound to the given transaction.
//
// [Ja] WithTx はトランザクションを使用する新しい AnimeEventRepository を返す。
func (r *AnimeEventRepository) WithTx(tx *sql.Tx) *AnimeEventRepository {
	return &AnimeEventRepository{queries: r.queries.WithTx(tx)}
}

// CreateAnimeEventParams holds the attributes for creating an event. The id,
// title / title_en / description / description_en, sort_number and timestamps are
// assigned by the database (the text columns default to NULL, sort_number to 0); works
// source only the broadcast period, so this carries just the natural key
// (anime_id, kind) and the dates. EndedOn is nil when the event is open-ended.
// (anime_id, kind) is the natural key enforced by a UNIQUE index.
//
// [Ja] CreateAnimeEventParams はイベント作成時の属性を保持する。id / title / title_en /
// description / description_en / sort_number / タイムスタンプはデータベースが採番する
// (テキスト列は NULL、sort_number は 0 が既定値)。works は放送期間のみを source するため、
// 自然キー (anime_id, kind) と日付だけを持つ。EndedOn はイベントの終了が未定のとき nil。
// (anime_id, kind) が UNIQUE インデックスで守られる自然キー。
type CreateAnimeEventParams struct {
	AnimeID   model.AnimeID
	Kind      model.AnimeEventKind
	StartedOn time.Time
	EndedOn   *time.Time
}

// Create inserts a new event and returns the created row.
//
// [Ja] Create は新しいイベントを挿入し、作成された行を返す。
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

// UpdateAnimeEventParams holds the attributes for updating an event, identified by its
// primary key. Only started_on / ended_on are mutable; the natural key (anime_id, kind)
// is fixed and the non-sourced title / description / sort_number columns are preserved,
// so an event whose broadcast period changed is updated in place rather than deleted and
// re-created.
//
// [Ja] UpdateAnimeEventParams は主キーで特定したイベントの更新時の属性を保持する。可変なのは
// started_on / ended_on のみで、自然キー (anime_id, kind) は固定、source しない title /
// description / sort_number 列は保全するため、放送期間が変わったイベントは削除と再作成では
// なくその場で更新する。
type UpdateAnimeEventParams struct {
	ID        model.AnimeEventID
	StartedOn time.Time
	EndedOn   *time.Time
}

// Update overwrites the started_on / ended_on of the identified row.
//
// [Ja] Update は指定行の started_on / ended_on を上書きする。
func (r *AnimeEventRepository) Update(ctx context.Context, params UpdateAnimeEventParams) error {
	return r.queries.UpdateAnimeEvent(ctx, query.UpdateAnimeEventParams{
		ID:        int64(params.ID),
		StartedOn: params.StartedOn,
		EndedOn:   nullTimeFromPtr(params.EndedOn),
	})
}

// Delete removes the event with the given primary key.
//
// [Ja] Delete は指定主キーのイベントを削除する。
func (r *AnimeEventRepository) Delete(ctx context.Context, id model.AnimeEventID) error {
	return r.queries.DeleteAnimeEvent(ctx, int64(id))
}

// ListByAnimeIDs loads the events for the given anime IDs, ordered by (anime_id, kind).
// It is used by the phase 2 satellite reconciliation to batch-fetch the existing rows
// for a page of anime-resolved works in one query instead of N per-anime lookups. An
// empty input returns an empty slice without querying.
//
// [Ja] ListByAnimeIDs は指定 anime ID 群のイベントを (anime_id, kind) 順でロードする。
// フェーズ 2 の別表リコンシリエーションが、anime 解決済み works の 1 ページぶんの既存行を
// N 回の anime 単位ルックアップではなく 1 クエリで一括取得するために使う。空入力では
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

// nullTimeFromPtr converts a nullable date pointer into the sqlc NullTime: nil becomes
// NULL (an open-ended / unknown end date).
//
// [Ja] nullTimeFromPtr は NULL 許容の日付ポインタを sqlc の NullTime に変換する。nil は
// NULL (終了未定・不明) になる。
func nullTimeFromPtr(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

// toAnimeEventModel converts a query row into the domain model.
//
// [Ja] toAnimeEventModel は query の行をドメインモデルに変換する。
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

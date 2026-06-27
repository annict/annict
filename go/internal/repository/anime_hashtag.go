package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// AnimeHashtagRepository handles data access for the anime_hashtags table (an anime's
// hashtags, e.g. the tag shown / searched for the work on X).
//
// [Ja] AnimeHashtagRepository は anime_hashtags テーブル (anime のハッシュタグ。例: X で
// 作品に表示・検索されるタグ) へのデータアクセスを担う。
type AnimeHashtagRepository struct {
	queries *query.Queries
}

// NewAnimeHashtagRepository constructs an AnimeHashtagRepository.
//
// [Ja] NewAnimeHashtagRepository は AnimeHashtagRepository を生成する。
func NewAnimeHashtagRepository(queries *query.Queries) *AnimeHashtagRepository {
	return &AnimeHashtagRepository{queries: queries}
}

// WithTx returns a new AnimeHashtagRepository bound to the given transaction.
//
// [Ja] WithTx はトランザクションを使用する新しい AnimeHashtagRepository を返す。
func (r *AnimeHashtagRepository) WithTx(tx *sql.Tx) *AnimeHashtagRepository {
	return &AnimeHashtagRepository{queries: r.queries.WithTx(tx)}
}

// CreateAnimeHashtagParams holds the attributes for creating a hashtag. The id,
// sort_number and timestamps are assigned by the database (sort_number defaults to 0);
// works source only the tag value, so this carries just the natural key
// (anime_id, hashtag), which a UNIQUE index enforces.
//
// [Ja] CreateAnimeHashtagParams はハッシュタグ作成時の属性を保持する。id / sort_number /
// タイムスタンプはデータベースが採番する (sort_number は 0 が既定値)。works はタグ値のみを
// source するため、UNIQUE インデックスで守られる自然キー (anime_id, hashtag) だけを持つ。
type CreateAnimeHashtagParams struct {
	AnimeID model.AnimeID
	Hashtag string
}

// Create inserts a new hashtag and returns the created row.
//
// [Ja] Create は新しいハッシュタグを挿入し、作成された行を返す。
func (r *AnimeHashtagRepository) Create(ctx context.Context, params CreateAnimeHashtagParams) (*model.AnimeHashtag, error) {
	row, err := r.queries.CreateAnimeHashtag(ctx, query.CreateAnimeHashtagParams{
		AnimeID: int64(params.AnimeID),
		Hashtag: params.Hashtag,
	})
	if err != nil {
		return nil, err
	}
	hashtag := toAnimeHashtagModel(row)
	return &hashtag, nil
}

// There is intentionally no Update: the hashtag is both the entire works-sourced value
// and the natural key, so a row is never mutated in place — a changed tag is a delete of
// the old row plus a create of the new one (see SyncAnimeHashtagsUsecase).
//
// [Ja] Update は意図的に設けない。hashtag は works が source する値そのものであると同時に
// 自然キーでもあるため、行をその場で書き換えることがない。タグの変更は旧行の削除と新行の作成
// になる (SyncAnimeHashtagsUsecase を参照)。

// Delete removes the hashtag with the given primary key.
//
// [Ja] Delete は指定主キーのハッシュタグを削除する。
func (r *AnimeHashtagRepository) Delete(ctx context.Context, id model.AnimeHashtagID) error {
	return r.queries.DeleteAnimeHashtag(ctx, int64(id))
}

// ListByAnimeIDs loads the hashtags for the given anime IDs, ordered by
// (anime_id, sort_number, hashtag). It is used by the phase 2 satellite reconciliation
// to batch-fetch the existing rows for a page of anime-resolved works in one query
// instead of N per-anime lookups. An empty input returns an empty slice without querying.
//
// [Ja] ListByAnimeIDs は指定 anime ID 群のハッシュタグを (anime_id, sort_number, hashtag)
// 順でロードする。フェーズ 2 の別表リコンシリエーションが、anime 解決済み works の 1 ページ
// ぶんの既存行を N 回の anime 単位ルックアップではなく 1 クエリで一括取得するために使う。
// 空入力ではクエリせず空スライスを返す。
func (r *AnimeHashtagRepository) ListByAnimeIDs(ctx context.Context, animeIDs []model.AnimeID) ([]*model.AnimeHashtag, error) {
	if len(animeIDs) == 0 {
		return []*model.AnimeHashtag{}, nil
	}

	rawIDs := make([]int64, len(animeIDs))
	for i, id := range animeIDs {
		rawIDs[i] = int64(id)
	}

	rows, err := r.queries.ListAnimeHashtagsByAnimeIDs(ctx, rawIDs)
	if err != nil {
		return nil, err
	}

	hashtags := make([]*model.AnimeHashtag, len(rows))
	for i, row := range rows {
		hashtag := toAnimeHashtagModel(row)
		hashtags[i] = &hashtag
	}
	return hashtags, nil
}

// toAnimeHashtagModel converts a query row into the domain model.
//
// [Ja] toAnimeHashtagModel は query の行をドメインモデルに変換する。
func toAnimeHashtagModel(row query.AnimeHashtag) model.AnimeHashtag {
	return model.AnimeHashtag{
		ID:         model.AnimeHashtagID(row.ID),
		AnimeID:    model.AnimeID(row.AnimeID),
		Hashtag:    row.Hashtag,
		SortNumber: row.SortNumber,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}
}

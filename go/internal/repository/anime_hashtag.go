package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// AnimeHashtagRepositoryはanime_hashtagsテーブル (animeのハッシュタグ。例: Xで
// 作品に表示・検索されるタグ) へのデータアクセスを担う。
type AnimeHashtagRepository struct {
	queries *query.Queries
}

// NewAnimeHashtagRepositoryはAnimeHashtagRepositoryを生成する。
func NewAnimeHashtagRepository(queries *query.Queries) *AnimeHashtagRepository {
	return &AnimeHashtagRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいAnimeHashtagRepositoryを返す。
func (r *AnimeHashtagRepository) WithTx(tx *sql.Tx) *AnimeHashtagRepository {
	return &AnimeHashtagRepository{queries: r.queries.WithTx(tx)}
}

// CreateAnimeHashtagParamsはハッシュタグ作成時の属性を保持する。id / sort_number /
// タイムスタンプはデータベースが採番する (sort_numberは0が既定値)。worksはタグ値のみを
// sourceするため、UNIQUEインデックスで守られる自然キー (anime_id, hashtag) だけを持つ。
type CreateAnimeHashtagParams struct {
	AnimeID model.AnimeID
	Hashtag string
}

// Createは新しいハッシュタグを挿入し、作成された行を返す。
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

// Updateは意図的に設けない。hashtagはworksがsourceする値そのものであると同時に
// 自然キーでもあるため、行をその場で書き換えることがない。タグの変更は旧行の削除と新行の作成
// になる (SyncAnimeHashtagsUsecaseを参照)。

// Deleteは指定主キーのハッシュタグを削除する。
func (r *AnimeHashtagRepository) Delete(ctx context.Context, id model.AnimeHashtagID) error {
	return r.queries.DeleteAnimeHashtag(ctx, int64(id))
}

// ListByAnimeIDsは指定anime ID群のハッシュタグを (anime_id, sort_number, hashtag)
// 順でロードする。フェーズ2の別表リコンシリエーションが、anime解決済みworksの1ページ
// ぶんの既存行をN回のanime単位ルックアップではなく1クエリで一括取得するために使う。
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

// toAnimeHashtagModelはqueryの行をドメインモデルに変換する。
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

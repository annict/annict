package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// AnimeLinkRepositoryはanime_linksテーブル (公式サイトやWikipediaなどanimeの
// 外部リンク) へのデータアクセスを担う。
type AnimeLinkRepository struct {
	queries *query.Queries
}

// NewAnimeLinkRepositoryはAnimeLinkRepositoryを生成する。
func NewAnimeLinkRepository(queries *query.Queries) *AnimeLinkRepository {
	return &AnimeLinkRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいAnimeLinkRepositoryを返す。
func (r *AnimeLinkRepository) WithTx(tx *sql.Tx) *AnimeLinkRepository {
	return &AnimeLinkRepository{queries: r.queries.WithTx(tx)}
}

// CreateAnimeLinkParamsはリンク作成時の属性を保持する。id・label・sort_number・
// タイムスタンプはデータベースが採番する。(anime_id, kind, language) が別表同期の
// リコンサイル対象となる自然キー。worksはURLのみをsourceとするため、label / label_enは
// NULLのまま、sort_numberは0が既定値になる。
type CreateAnimeLinkParams struct {
	AnimeID  model.AnimeID
	Kind     model.AnimeLinkKind
	Language model.Language
	URL      string
}

// Createは新しいリンクを挿入し、作成された行を返す。
func (r *AnimeLinkRepository) Create(ctx context.Context, params CreateAnimeLinkParams) (*model.AnimeLink, error) {
	row, err := r.queries.CreateAnimeLink(ctx, query.CreateAnimeLinkParams{
		AnimeID:  int64(params.AnimeID),
		Kind:     query.AnimeLinkKind(params.Kind),
		Language: query.Language(params.Language),
		Url:      params.URL,
	})
	if err != nil {
		return nil, err
	}
	link := toAnimeLinkModel(row)
	return &link, nil
}

// UpdateAnimeLinkParamsは主キーで特定したリンクの更新時の属性を保持する。可変なのは
// urlのみで、自然キー (anime_id, kind, language) は固定のため、URLが変わったリンクは削除と
// 再作成ではなくその場で更新する。label / label_enはworks由来ではないため触らない。
type UpdateAnimeLinkParams struct {
	ID  model.AnimeLinkID
	URL string
}

// Updateは指定行のurlを上書きする。
func (r *AnimeLinkRepository) Update(ctx context.Context, params UpdateAnimeLinkParams) error {
	return r.queries.UpdateAnimeLink(ctx, query.UpdateAnimeLinkParams{
		ID:  int64(params.ID),
		Url: params.URL,
	})
}

// Deleteは指定主キーのリンクを削除する。
func (r *AnimeLinkRepository) Delete(ctx context.Context, id model.AnimeLinkID) error {
	return r.queries.DeleteAnimeLink(ctx, int64(id))
}

// ListByAnimeIDsは指定anime ID群のリンクを (anime_id, kind, language) 順でロードする。
// フェーズ2の別表リコンシリエーションが、anime解決済みworksの1ページぶんの既存行を
// N回のanime単位ルックアップではなく1クエリで一括取得するために使う。空入力では
// クエリせず空スライスを返す。
func (r *AnimeLinkRepository) ListByAnimeIDs(ctx context.Context, animeIDs []model.AnimeID) ([]*model.AnimeLink, error) {
	if len(animeIDs) == 0 {
		return []*model.AnimeLink{}, nil
	}

	rawIDs := make([]int64, len(animeIDs))
	for i, id := range animeIDs {
		rawIDs[i] = int64(id)
	}

	rows, err := r.queries.ListAnimeLinksByAnimeIDs(ctx, rawIDs)
	if err != nil {
		return nil, err
	}

	links := make([]*model.AnimeLink, len(rows))
	for i, row := range rows {
		link := toAnimeLinkModel(row)
		links[i] = &link
	}
	return links, nil
}

// toAnimeLinkModelはqueryの行をドメインモデルに変換する。NULL許容のlabelカラムは
// 値が無いときnilポインタになる。
func toAnimeLinkModel(row query.AnimeLink) model.AnimeLink {
	link := model.AnimeLink{
		ID:         model.AnimeLinkID(row.ID),
		AnimeID:    model.AnimeID(row.AnimeID),
		Kind:       model.AnimeLinkKind(row.Kind),
		Language:   model.Language(row.Language),
		URL:        row.Url,
		SortNumber: row.SortNumber,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}
	if row.Label.Valid {
		label := row.Label.String
		link.Label = &label
	}
	if row.LabelEn.Valid {
		labelEn := row.LabelEn.String
		link.LabelEn = &labelEn
	}
	return link
}

package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// AnimeLinkRepository handles data access for the anime_links table (an anime's
// external links such as its official site and Wikipedia page).
//
// [Ja] AnimeLinkRepository は anime_links テーブル (公式サイトや Wikipedia など anime の
// 外部リンク) へのデータアクセスを担う。
type AnimeLinkRepository struct {
	queries *query.Queries
}

// NewAnimeLinkRepository constructs an AnimeLinkRepository.
//
// [Ja] NewAnimeLinkRepository は AnimeLinkRepository を生成する。
func NewAnimeLinkRepository(queries *query.Queries) *AnimeLinkRepository {
	return &AnimeLinkRepository{queries: queries}
}

// WithTx returns a new AnimeLinkRepository bound to the given transaction.
//
// [Ja] WithTx はトランザクションを使用する新しい AnimeLinkRepository を返す。
func (r *AnimeLinkRepository) WithTx(tx *sql.Tx) *AnimeLinkRepository {
	return &AnimeLinkRepository{queries: r.queries.WithTx(tx)}
}

// CreateAnimeLinkParams holds the attributes for creating a link. The id, label,
// sort_number and timestamps are assigned by the database; (anime_id, kind, language)
// is the natural key the satellite sync reconciles on. Works only source the URL, so
// label / label_en stay NULL and sort_number defaults to 0.
//
// [Ja] CreateAnimeLinkParams はリンク作成時の属性を保持する。id・label・sort_number・
// タイムスタンプはデータベースが採番する。(anime_id, kind, language) が別表同期の
// リコンサイル対象となる自然キー。works は URL のみを source とするため、label / label_en は
// NULL のまま、sort_number は 0 が既定値になる。
type CreateAnimeLinkParams struct {
	AnimeID  model.AnimeID
	Kind     model.AnimeLinkKind
	Language model.Language
	URL      string
}

// Create inserts a new link and returns the created row.
//
// [Ja] Create は新しいリンクを挿入し、作成された行を返す。
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

// UpdateAnimeLinkParams holds the attributes for updating a link, identified by its
// primary key. Only url is mutable; the natural key (anime_id, kind, language) is
// fixed, so a link whose URL changed is updated in place rather than deleted and
// re-created. label / label_en are not works-sourced and are left untouched.
//
// [Ja] UpdateAnimeLinkParams は主キーで特定したリンクの更新時の属性を保持する。可変なのは
// url のみで、自然キー (anime_id, kind, language) は固定のため、URL が変わったリンクは削除と
// 再作成ではなくその場で更新する。label / label_en は works 由来ではないため触らない。
type UpdateAnimeLinkParams struct {
	ID  model.AnimeLinkID
	URL string
}

// Update overwrites the url of the identified row.
//
// [Ja] Update は指定行の url を上書きする。
func (r *AnimeLinkRepository) Update(ctx context.Context, params UpdateAnimeLinkParams) error {
	return r.queries.UpdateAnimeLink(ctx, query.UpdateAnimeLinkParams{
		ID:  int64(params.ID),
		Url: params.URL,
	})
}

// Delete removes the link with the given primary key.
//
// [Ja] Delete は指定主キーのリンクを削除する。
func (r *AnimeLinkRepository) Delete(ctx context.Context, id model.AnimeLinkID) error {
	return r.queries.DeleteAnimeLink(ctx, int64(id))
}

// ListByAnimeIDs loads the links for the given anime IDs, ordered by
// (anime_id, kind, language). It is used by the phase 2 satellite reconciliation to
// batch-fetch the existing rows for a page of anime-resolved works in one query
// instead of N per-anime lookups. An empty input returns an empty slice without
// querying.
//
// [Ja] ListByAnimeIDs は指定 anime ID 群のリンクを (anime_id, kind, language) 順でロードする。
// フェーズ 2 の別表リコンシリエーションが、anime 解決済み works の 1 ページぶんの既存行を
// N 回の anime 単位ルックアップではなく 1 クエリで一括取得するために使う。空入力では
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

// toAnimeLinkModel converts a query row into the domain model. The nullable label
// columns become nil pointers when absent.
//
// [Ja] toAnimeLinkModel は query の行をドメインモデルに変換する。NULL 許容の label カラムは
// 値が無いとき nil ポインタになる。
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

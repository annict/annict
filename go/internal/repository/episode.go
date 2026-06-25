package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// EpisodeRepository handles data access for the episodes table and related
// joins.
//
// [Ja] EpisodeRepository は episodes テーブルおよび関連 JOIN へのデータアクセスを担う。
type EpisodeRepository struct {
	queries *query.Queries
}

// NewEpisodeRepository constructs an EpisodeRepository.
//
// [Ja] NewEpisodeRepository は EpisodeRepository を生成する。
func NewEpisodeRepository(queries *query.Queries) *EpisodeRepository {
	return &EpisodeRepository{queries: queries}
}

// WithTx returns a new EpisodeRepository bound to the given transaction.
//
// [Ja] WithTx はトランザクションを使用する新しい EpisodeRepository を返す。
func (r *EpisodeRepository) WithTx(tx *sql.Tx) *EpisodeRepository {
	return &EpisodeRepository{queries: r.queries.WithTx(tx)}
}

// ListForAnimeSyncByIDs loads the episodes with the given IDs, projecting the
// columns the phase 2 reconciliation maps onto animes / anime_classifications
// (including the episodes.anime_id mapping column). The parent work's anime_id is
// resolved through a JOIN on episodes.work_id and surfaced as ParentAnimeID, so
// the episode sync never has to look the parent up row by row. Rows are ordered
// by id; missing IDs are silently skipped. An empty input returns an empty slice
// without querying.
//
// [Ja] ListForAnimeSyncByIDs は指定 ID の episodes を、フェーズ 2 のリコンシリエーションが
// animes / anime_classifications に写像するカラム (episodes.anime_id のマッピングカラムを
// 含む) を射影してロードする。親作品の anime_id は episodes.work_id の JOIN で解決して
// ParentAnimeID として返すため、エピソード同期が親を行単位で引く必要はない。行は id 昇順で、
// 存在しない ID は黙って除外される。空入力ではクエリせず空スライスを返す。
func (r *EpisodeRepository) ListForAnimeSyncByIDs(ctx context.Context, episodeIDs []model.EpisodeID) ([]*model.Episode, error) {
	if len(episodeIDs) == 0 {
		return []*model.Episode{}, nil
	}

	ids := make([]int64, len(episodeIDs))
	for i, id := range episodeIDs {
		ids[i] = int64(id)
	}

	rows, err := r.queries.ListEpisodesForAnimeSyncByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	episodes := make([]*model.Episode, len(rows))
	for i, row := range rows {
		episodes[i] = episodeFromAnimeSyncRow(row)
	}
	return episodes, nil
}

// UpdateAnimeID writes back the episodes.anime_id mapping column, marking the
// episode as synced to the given anime. updated_at is intentionally left
// untouched so the bookkeeping write is not mistaken for a content change on the
// source-of-truth row.
//
// [Ja] UpdateAnimeID は episodes.anime_id マッピングカラムを書き戻し、エピソードを
// 指定アニメへ同期済みとして印付ける。updated_at は意図的に触れず、正本側の行への記帳
// 書き込みが内容変更と取り違えられないようにする。
func (r *EpisodeRepository) UpdateAnimeID(ctx context.Context, episodeID model.EpisodeID, animeID model.AnimeID) error {
	return r.queries.UpdateEpisodeAnimeID(ctx, query.UpdateEpisodeAnimeIDParams{
		ID:      int64(episodeID),
		AnimeID: sql.NullInt64{Int64: int64(animeID), Valid: true},
	})
}

// episodeFromAnimeSyncRow converts an anime-sync query row into *model.Episode.
// The nullable columns (title / number / raw_number / archive_message / anime_id /
// parent_anime_id) are carried as pointers so the sync usecase can distinguish
// "absent" from a zero value, mirroring how workFromAnimeSyncRow handles works.
//
// [Ja] episodeFromAnimeSyncRow は anime 同期の query 行を *model.Episode に変換する。
// NULL 許容カラム (title / number / raw_number / archive_message / anime_id /
// parent_anime_id) はポインタで持ち、同期 UseCase が「未設定」とゼロ値を区別できる
// ようにする。workFromAnimeSyncRow が works を扱うのと同じ方針。
func episodeFromAnimeSyncRow(row query.ListEpisodesForAnimeSyncByIDsRow) *model.Episode {
	episode := &model.Episode{
		ID:         model.EpisodeID(row.ID),
		WorkID:     model.WorkID(row.WorkID),
		TitleRo:    row.TitleRo,
		TitleEn:    row.TitleEn,
		SortNumber: row.SortNumber,
		Status:     model.EpisodeStatus(row.Status),
	}
	if row.Title.Valid {
		title := row.Title.String
		episode.Title = &title
	}
	if row.Number.Valid {
		number := row.Number.String
		episode.Number = &number
	}
	if row.RawNumber.Valid {
		rawNumber := row.RawNumber.Float64
		episode.RawNumber = &rawNumber
	}
	if row.ArchiveMessage.Valid {
		archiveMessage := row.ArchiveMessage.String
		episode.ArchiveMessage = &archiveMessage
	}
	if row.AnimeID.Valid {
		animeID := model.AnimeID(row.AnimeID.Int64)
		episode.AnimeID = &animeID
	}
	if row.ParentAnimeID.Valid {
		parentAnimeID := model.AnimeID(row.ParentAnimeID.Int64)
		episode.ParentAnimeID = &parentAnimeID
	}
	return episode
}

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

// DBEpisodeListParams identifies one page of a work's episode list on the Annict
// DB screen. Page is 1-based.
//
// [Ja] DBEpisodeListParams は Annict DB 画面の、ある作品のエピソード一覧 1 ページ分を
// 指定する。Page は 1 始まり。
type DBEpisodeListParams struct {
	WorkID  model.WorkID
	Page    int32
	PerPage int32
}

// ListForDB loads one page of a work's episodes for the Annict DB screen, newest
// episode first (sort_number descending, matching the Rails screen) with the id as
// a tiebreaker so equal sort_numbers still paginate deterministically.
//
// Episodes remain the source of truth during the migration, so the screen reads
// them rather than the derived animes / anime_classifications rows, which only
// catch up with Rails-side changes after the hourly sync batch. Deleted episodes
// are excluded by deleted_at alone (the Rails `without_deleted` scope), and the
// remaining rows carry unpublished_at / deleted_at so callers can derive the
// display status through model.Episode.DerivedStatus.
//
// [Ja] ListForDB は Annict DB 画面向けに、ある作品のエピソードを 1 ページ分、新しい
// エピソードから順に (Rails 画面に合わせて sort_number 降順で) ロードする。sort_number が
// 同値でもページングが決定的になるよう id をタイブレーカにする。
//
// 移行期間中の正本は episodes 側であるため、画面も派生である animes /
// anime_classifications ではなく episodes を読む (派生側は Rails 経由の変更が毎時の
// 同期バッチ後にしか反映されない)。削除済みエピソードの除外は deleted_at のみで行い
// (Rails の `without_deleted` スコープと同じ)、残った行は unpublished_at / deleted_at を
// 持つため、呼び出し側は model.Episode.DerivedStatus で表示用の状態を導出できる。
func (r *EpisodeRepository) ListForDB(ctx context.Context, params DBEpisodeListParams) ([]*model.Episode, error) {
	// Widen before multiplying: callers accept any page number that fits in an int32, and at
	// 100 rows per page the int32 product wraps negative inside that range, which PostgreSQL
	// rejects as an OFFSET.
	//
	// [Ja] 乗算の前に幅を広げる。呼び出し側は int32 に収まるページ番号をすべて受け付けるが、
	// 1 ページ 100 件では int32 同士の積がその範囲内で負に折り返し、PostgreSQL がその OFFSET
	// を拒否するため。
	offset := int64(params.Page-1) * int64(params.PerPage)

	rows, err := r.queries.ListDBEpisodes(ctx, query.ListDBEpisodesParams{
		WorkID:     int64(params.WorkID),
		PerPage:    params.PerPage,
		PageOffset: offset,
	})
	if err != nil {
		return nil, err
	}

	episodes := make([]*model.Episode, len(rows))
	for i, row := range rows {
		episodes[i] = episodeFromDBListRow(row)
	}
	return episodes, nil
}

// CountForDB returns the total number of episodes the DB screen lists for the work,
// using the same filter as ListForDB so the pagination total matches the listed
// rows.
//
// [Ja] CountForDB は DB 画面がその作品について一覧するエピソードの総件数を返す。
// ページネーションの総数が一覧の行と一致するよう、ListForDB と同じ絞り込みを使う。
func (r *EpisodeRepository) CountForDB(ctx context.Context, workID model.WorkID) (int64, error) {
	return r.queries.CountDBEpisodes(ctx, int64(workID))
}

// episodeFromDBListRow converts an Annict DB list row into *model.Episode. The row
// is a partial load: the anime mapping columns, the dormant status column and
// archive_message are not selected and stay at their zero value.
//
// [Ja] episodeFromDBListRow は Annict DB 一覧の行を *model.Episode に変換する。行は
// 部分ロードで、anime マッピングカラム・休眠カラム status・archive_message は選択せず
// ゼロ値のまま残る。
func episodeFromDBListRow(row query.ListDBEpisodesRow) *model.Episode {
	episode := &model.Episode{
		ID:                  model.EpisodeID(row.ID),
		WorkID:              model.WorkID(row.WorkID),
		TitleRo:             row.TitleRo,
		TitleEn:             row.TitleEn,
		SortNumber:          row.SortNumber,
		EpisodeRecordsCount: row.EpisodeRecordsCount,
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
	if row.UnpublishedAt.Valid {
		unpublishedAt := row.UnpublishedAt.Time
		episode.UnpublishedAt = &unpublishedAt
	}
	if row.DeletedAt.Valid {
		deletedAt := row.DeletedAt.Time
		episode.DeletedAt = &deletedAt
	}
	return episode
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

// ListIDsAfter returns up to batchSize episode IDs whose id is greater than
// afterID, in ascending id order. It is the keyset-pagination primitive the phase 2
// batch job (task 2-4) uses to walk the whole episodes table page by page: pass
// afterID=0 for the first page, then the last returned id as the cursor for the
// next page, until an empty page signals the end. Keyset (id > cursor) is used over
// LIMIT/OFFSET because the batch scans a large table and OFFSET re-reads all skipped
// rows on every page, degrading to O(n^2) over a full scan.
//
// [Ja] ListIDsAfter は afterID より大きい episode ID を昇順で最大 batchSize 件返す。
// フェーズ 2 のバッチジョブ (タスク 2-4) が episodes テーブル全体をページ単位で走査する
// ための keyset ページネーションの基本操作で、最初のページは afterID=0 を渡し、以降は
// 直前に返った末尾の id をカーソルにして次ページを引き、空ページで終端を知る。
// LIMIT/OFFSET ではなく keyset (id > カーソル) を使うのは、バッチが大テーブルを走査する
// ため。OFFSET はページごとにスキップ分を読み直し、全件走査では O(n^2) に劣化する。
func (r *EpisodeRepository) ListIDsAfter(ctx context.Context, afterID model.EpisodeID, batchSize int) ([]model.EpisodeID, error) {
	rows, err := r.queries.ListEpisodeIDsAfter(ctx, query.ListEpisodeIDsAfterParams{
		AfterID: int64(afterID),
		// batchSize is a small bounded page size (default 1000), never near int32 max.
		//
		// [Ja] batchSize は小さく上限のあるページサイズ (既定 1000) で int32 上限には達しない。
		BatchSize: int32(batchSize), // #nosec G115
	})
	if err != nil {
		return nil, err
	}

	ids := make([]model.EpisodeID, len(rows))
	for i, id := range rows {
		ids[i] = model.EpisodeID(id)
	}
	return ids, nil
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
// The nullable columns (title / number / raw_number / archive_message /
// unpublished_at / deleted_at / anime_id / parent_anime_id) are carried as pointers
// so the sync usecase can distinguish "absent" from a zero value, mirroring how
// workFromAnimeSyncRow handles works.
//
// [Ja] episodeFromAnimeSyncRow は anime 同期の query 行を *model.Episode に変換する。
// NULL 許容カラム (title / number / raw_number / archive_message / unpublished_at /
// deleted_at / anime_id / parent_anime_id) はポインタで持ち、同期 UseCase が「未設定」と
// ゼロ値を区別できるようにする。workFromAnimeSyncRow が works を扱うのと同じ方針。
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
	if row.UnpublishedAt.Valid {
		unpublishedAt := row.UnpublishedAt.Time
		episode.UnpublishedAt = &unpublishedAt
	}
	if row.DeletedAt.Valid {
		deletedAt := row.DeletedAt.Time
		episode.DeletedAt = &deletedAt
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

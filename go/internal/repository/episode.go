package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// pgErrCodeLockNotAvailable is the SQLSTATE PostgreSQL raises when a NOWAIT locking clause finds
// the row already locked (55P03, lock_not_available).
//
// [Ja] pgErrCodeLockNotAvailable は、NOWAIT のロック句が既にロック済みの行に当たったときに
// PostgreSQL が返す SQLSTATE (55P03, lock_not_available)。
const pgErrCodeLockNotAvailable = "55P03"

// ErrEpisodeLockUnavailable reports that an episode row the update needs was held by another
// transaction, so the attempt was abandoned instead of waiting for it. The NOWAIT clause it
// comes from is what keeps Rails' episode -> work save order from deadlocking against the Go
// work -> episodes update order, and PostgreSQL aborts the transaction when it fires. Callers
// therefore retry the whole transaction rather than the failed statement. Translating the
// driver error here keeps the SQLSTATE inside the infrastructure layer, as Update already does
// for sql.ErrNoRows.
//
// [Ja] ErrEpisodeLockUnavailable は、更新に必要なエピソード行を他のトランザクションが保持して
// いたため、待たずに試行を中断したことを表す。由来する NOWAIT 句は、Rails の episode -> work の
// 保存順序が Go の work -> episodes の更新順序とデッドロックするのを防ぐためのものであり、
// 発火時に PostgreSQL はトランザクションを中断する。したがって呼び出し側は、失敗した
// ステートメントではなくトランザクション全体を再試行する。ここでドライバのエラーを翻訳するのは、
// Update が sql.ErrNoRows について既にそうしているのと同じく、SQLSTATE を Infrastructure 層に
// 閉じ込めるため。
var ErrEpisodeLockUnavailable = errors.New("エピソード行のロックを取得できませんでした")

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
// archive_message are not selected and stay at their zero value. The preceding
// episode's two numbers come from the query's neighbour derivation, so they are nil
// for the work's first episode.
//
// [Ja] episodeFromDBListRow は Annict DB 一覧の行を *model.Episode に変換する。行は
// 部分ロードで、anime マッピングカラム・休眠カラム status・archive_message は選択せず
// ゼロ値のまま残る。直前のエピソードの 2 系統の話数はクエリ側の隣接行の導出に由来し、
// 作品の最初のエピソードでは nil になる。
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
	if row.PrevNumber.Valid {
		prevNumber := row.PrevNumber.String
		episode.PrevNumber = &prevNumber
	}
	if row.PrevRawNumber.Valid {
		prevRawNumber := row.PrevRawNumber.Float64
		episode.PrevRawNumber = &prevRawNumber
	}
	return episode
}

// DBEpisodeEditTarget is what the Annict DB episode edit form loads: the episode being
// edited and its parent work. The work is a partial load carrying only what the page needs
// from it (the heading's title and the subnav's no_episodes), so it is kept beside the
// episode rather than folded into model.Episode, which holds only the episode's own columns.
//
// [Ja] DBEpisodeEditTarget は Annict DB のエピソード編集フォームが読み込むもの (編集対象の
// エピソードとその親作品) を表す。作品はページが必要とするカラム (見出しの title と
// サブナビの no_episodes) だけの部分ロードのため、エピソード自身のカラムだけを持つ
// model.Episode に畳み込まず、エピソードと並べて持つ。
type DBEpisodeEditTarget struct {
	Episode *model.Episode
	Work    *model.Work
}

// GetForEditByID loads the episode the Annict DB edit form edits, together with its parent
// work. Deleted episodes and episodes of deleted works are excluded by the query, so
// (nil, nil) means the id names no editable episode.
//
// [Ja] GetForEditByID は Annict DB の編集フォームが編集するエピソードを、その親作品と一緒に
// 読み込む。削除済みのエピソードと、削除済み作品のエピソードはクエリ側で除外するため、
// (nil, nil) は編集できるエピソードがその id に無いことを表す。
func (r *EpisodeRepository) GetForEditByID(ctx context.Context, id model.EpisodeID) (*DBEpisodeEditTarget, error) {
	row, err := r.queries.GetEpisodeForEditByID(ctx, int64(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	episode := &model.Episode{
		ID:         model.EpisodeID(row.ID),
		WorkID:     model.WorkID(row.WorkID),
		SortNumber: row.SortNumber,
		TitleEn:    row.TitleEn,
	}
	if row.Number.Valid {
		number := row.Number.String
		episode.Number = &number
	}
	if row.RawNumber.Valid {
		rawNumber := row.RawNumber.Float64
		episode.RawNumber = &rawNumber
	}
	if row.Title.Valid {
		title := row.Title.String
		episode.Title = &title
	}
	if row.UpdatedAt.Valid {
		updatedAt := row.UpdatedAt.Time
		episode.UpdatedAt = &updatedAt
	}

	return &DBEpisodeEditTarget{
		Episode: episode,
		Work: &model.Work{
			ID:         model.WorkID(row.WorkID),
			Title:      row.WorkTitle,
			NoEpisodes: row.WorkNoEpisodes,
		},
	}, nil
}

// GetForUpdateByID loads the columns the Annict DB episode update needs on top of the
// submitted values: title_ro and the state timestamps the animes dual-write maps but the form
// does not edit, plus the episode's own anime and its parent work's. Deleted episodes and
// episodes of deleted works are excluded by the query, as they are for the edit form, so
// (nil, nil) means the id names no updatable episode.
//
// [Ja] GetForUpdateByID は、Annict DB のエピソード更新が送信された値に加えて必要とするカラムを
// 読み込む。animes への両書きが写像するがフォームでは編集しない title_ro と状態のタイムスタンプ、
// およびエピソード自身の anime と親作品の anime。削除済みのエピソードと削除済み作品のエピソード
// は、編集フォームと同じくクエリ側で除外するため、(nil, nil) は更新できるエピソードがその id に
// 無いことを表す。
func (r *EpisodeRepository) GetForUpdateByID(ctx context.Context, id model.EpisodeID) (*model.Episode, error) {
	row, err := r.queries.GetEpisodeForUpdateByID(ctx, int64(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	episode := &model.Episode{
		ID:      model.EpisodeID(row.ID),
		WorkID:  model.WorkID(row.WorkID),
		TitleRo: row.TitleRo,
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

	return episode, nil
}

// UpdateEpisodeParams holds the attributes of one episode edit submit. UserID attributes the
// recorded change to its author, as it does on create.
//
// WorkID is the parent observed by the edit pre-read. Update locks the episode's current parent
// and requires it to match, so a row moved to another work in between is not written under the
// old ordering assumptions.
//
// Version is the updated_at the submit was made against: nil states that the row carried no
// updated_at when the form was opened, which is a version in its own right because the shared
// column is nullable.
//
// [Ja] UpdateEpisodeParams はエピソード編集の 1 回の送信の属性を保持する。UserID は作成時と同じく
// 記録される変更を作成者に帰属させる。
//
// WorkID は編集用の事前読み取りで観測した親作品。Update はエピソードの現在の親をロックして
// 一致を要求するため、その間に別作品へ移された行を古い並び順の前提で書かない。
//
// Version は送信が前提とする updated_at。nil は、フォームを開いた時点で行が updated_at を持って
// いなかったことを表す。共有カラムが NULL 許容であるため、これも 1 つの版として扱う。
type UpdateEpisodeParams struct {
	ID         model.EpisodeID
	WorkID     model.WorkID
	Number     *string
	RawNumber  *float64
	Title      *string
	TitleEn    string
	SortNumber int32
	Version    *time.Time
	UserID     model.UserID
}

// Update applies one episode edit submit, reporting false when no row matched: either the
// episode is gone or its updated_at has moved on since the form was opened. The caller turns
// that into a conflict rather than retrying, so a submit made against a stale read never
// overwrites the write that happened in between.
//
// The parent work is locked in a statement of its own before UpdateDBEpisode, giving that later
// statement a fresh READ COMMITTED snapshot after any wait. The neighbours UpdateDBEpisode goes
// on to write or reference are then listed and locked in ascending id order with NOWAIT, which bounds the
// rows one edit locks instead of letting them grow with the work's episode count. A lock miss
// comes back as ErrEpisodeLockUnavailable, and the use case rolls the aborted transaction back
// and retries it whole.
//
// [Ja] Update はエピソード編集の 1 回の送信を適用し、どの行も一致しなかった場合に false を返す
// (エピソードが失われたか、フォームを開いてから updated_at が進んだか)。呼び出し側はこれを再試行
// せず競合として扱うため、古い読み取りに対する送信が、その間に入った書き込みを上書きすることは
// ない。
//
// 親作品は UpdateDBEpisode より前の独立した文でロックし、待機があった場合も後段の文が READ
// COMMITTED の新しいスナップショットを得るようにする。続いて UpdateDBEpisode が書くか参照する
// 隣接行を列挙し、id 昇順で NOWAIT ロックする。これにより 1 回の編集がロックする行数が、作品の
// エピソード数に比例して増えずに抑えられる。ロック取得失敗は ErrEpisodeLockUnavailable として
// 返り、UseCase が中断されたトランザクション全体を rollback して再試行する。
func (r *EpisodeRepository) Update(ctx context.Context, params UpdateEpisodeParams) (bool, error) {
	lockedWorkID, err := r.queries.LockWorkForEpisodeUpdateByID(ctx, int64(params.ID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	if model.WorkID(lockedWorkID) != params.WorkID {
		return false, nil
	}

	neighbourIDs, err := r.queries.ListEpisodeIDsForEpisodeUpdateByID(ctx, query.ListEpisodeIDsForEpisodeUpdateByIDParams{
		ID:         int64(params.ID),
		WorkID:     int64(params.WorkID),
		SortNumber: params.SortNumber,
	})
	if err != nil {
		return false, err
	}
	if err := r.queries.LockEpisodesForEpisodeUpdateByIDs(ctx, neighbourIDs); err != nil {
		if isLockNotAvailable(err) {
			return false, fmt.Errorf("%w: %w", ErrEpisodeLockUnavailable, err)
		}
		return false, err
	}

	_, err = r.queries.UpdateDBEpisode(ctx, query.UpdateDBEpisodeParams{
		ID:         int64(params.ID),
		WorkID:     int64(params.WorkID),
		Number:     nullStringFromPtr(params.Number),
		RawNumber:  nullFloat64FromPtr(params.RawNumber),
		Title:      nullStringFromPtr(params.Title),
		TitleEn:    params.TitleEn,
		SortNumber: params.SortNumber,
		Version:    nullTimeFromPtr(params.Version),
		UserID:     int64(params.UserID),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// isLockNotAvailable reports whether err is PostgreSQL's NOWAIT lock miss.
//
// [Ja] isLockNotAvailable は err が PostgreSQL の NOWAIT のロック取得失敗かどうかを返す。
func isLockNotAvailable(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == pgErrCodeLockNotAvailable
}

// DBEpisodeArchiveTarget is the episode whose archiving is being confirmed, together with the
// parent work its page heading and subnav describe.
//
// [Ja] DBEpisodeArchiveTarget は非公開を確認する対象のエピソードと、そのページの見出し・
// サブナビが示す親作品。
type DBEpisodeArchiveTarget struct {
	Episode *model.Episode
	Work    *model.Work
}

// GetForArchiveByID loads the episode the Annict DB archive-confirmation page names, together
// with its parent work. Deleted episodes and episodes of deleted works are excluded by the
// query, so (nil, nil) means the id names no episode the page can be shown for. The returned
// episode carries its state timestamps, so the caller decides through
// model.Episode.DerivedStatus whether the episode is in a state that can be archived.
//
// [Ja] GetForArchiveByID は Annict DB の非公開確認ページが名指しするエピソードを、その親作品と
// 一緒に読み込む。削除済みのエピソードと、削除済み作品のエピソードはクエリ側で除外するため、
// (nil, nil) はページを出せるエピソードがその id に無いことを表す。返すエピソードは状態の
// タイムスタンプを持つため、非公開にできる状態かどうかは呼び出し側が
// model.Episode.DerivedStatus で判断する。
func (r *EpisodeRepository) GetForArchiveByID(ctx context.Context, id model.EpisodeID) (*DBEpisodeArchiveTarget, error) {
	row, err := r.queries.GetEpisodeForArchiveByID(ctx, int64(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	episode := &model.Episode{
		ID:     model.EpisodeID(row.ID),
		WorkID: model.WorkID(row.WorkID),
	}
	if row.Number.Valid {
		number := row.Number.String
		episode.Number = &number
	}
	if row.Title.Valid {
		title := row.Title.String
		episode.Title = &title
	}
	if row.UnpublishedAt.Valid {
		unpublishedAt := row.UnpublishedAt.Time
		episode.UnpublishedAt = &unpublishedAt
	}
	if row.DeletedAt.Valid {
		deletedAt := row.DeletedAt.Time
		episode.DeletedAt = &deletedAt
	}
	return &DBEpisodeArchiveTarget{
		Episode: episode,
		Work: &model.Work{
			ID:         model.WorkID(row.WorkID),
			Title:      row.WorkTitle,
			NoEpisodes: row.WorkNoEpisodes,
		},
	}, nil
}

// ArchiveEpisodeParams identifies the episode one archive submit unpublishes. WorkID is the
// parent the confirmation page was built from; Archive requires the episode to still belong to
// it, so the counter decrement lands on the work that was counted.
//
// [Ja] ArchiveEpisodeParams は 1 回の非公開の送信が非公開にするエピソードを指定する。WorkID は
// 確認ページが前提とした親作品で、Archive はエピソードが今もそこに属していることを要求する。
// カウンターの減算を、数えていた作品に当てるため。
type ArchiveEpisodeParams struct {
	ID     model.EpisodeID
	WorkID model.WorkID
}

// ArchiveEpisodeResult reports the anime mapping on the episode row that was actually
// archived. AnimeID is nil for an episode not yet mapped to the reference model.
//
// [Ja] ArchiveEpisodeResult は、実際に非公開にしたエピソード行が持つ anime の写像を報告する。
// 参照モデルへ未マッピングのエピソードでは AnimeID は nil。
type ArchiveEpisodeResult struct {
	AnimeID *model.AnimeID
}

// Archive unpublishes an episode, returning nil when no row matched: the episode is gone, it
// was archived by someone else since the confirmation page was opened, or it no longer belongs
// to the work that page named. The caller turns that into the same not-found response the page
// itself gives, rather than reporting a write that did not happen. A successful result carries
// the anime mapping returned by the updated row, not the pre-transaction projection.
//
// [Ja] Archive はエピソードを非公開にし、どの行も一致しなかった場合に nil を返す (エピソード
// が失われた、確認ページを開いてから他者が非公開にした、またはそのページが名指しした作品に
// もう属していない)。呼び出し側はこれを、確認ページ自身が返すのと同じ not found の応答に変換し、
// 起きなかった書き込みを報告しない。成功時の結果は、トランザクション前の射影ではなく、更新した
// 行が返した anime の写像を運ぶ。
func (r *EpisodeRepository) Archive(ctx context.Context, params ArchiveEpisodeParams) (*ArchiveEpisodeResult, error) {
	row, err := r.queries.ArchiveDBEpisode(ctx, query.ArchiveDBEpisodeParams{
		ID:     int64(params.ID),
		WorkID: int64(params.WorkID),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	result := &ArchiveEpisodeResult{}
	if row.AnimeID.Valid {
		animeID := model.AnimeID(row.AnimeID.Int64)
		result.AnimeID = &animeID
	}

	return result, nil
}

// CreateEpisodeParams holds the attributes for creating an episode. The columns an editor
// does not fill in (title_ro / title_en / the counter caches / the state timestamps) keep
// their column defaults, so a created episode starts out published.
//
// AnimeID is the episodes.anime_id mapping column, set when the episode's anime was created
// alongside it and left nil while the parent work is not mapped yet (the sync creates the
// anime later). PrevEpisodeID names the episode that precedes this one at insert time.
//
// [Ja] CreateEpisodeParams はエピソード作成時の属性を保持する。編集者が入力しないカラム
// (title_ro / title_en / カウンターキャッシュ / 状態のタイムスタンプ) はカラムの既定値のまま
// にするため、作成されたエピソードは公開状態で始まる。
//
// AnimeID は episodes.anime_id のマッピングカラムで、エピソードの anime を併せて作成した
// ときに入り、親作品が未マッピングのあいだは nil のままになる (anime は後続の同期が作る)。
// PrevEpisodeID は挿入時点でこのエピソードの直前に来るエピソードを指す。
type CreateEpisodeParams struct {
	WorkID        model.WorkID
	Number        *string
	RawNumber     *float64
	Title         *string
	SortNumber    int32
	PrevEpisodeID *model.EpisodeID
	AnimeID       *model.AnimeID
	UserID        model.UserID
}

// Create inserts a new episode and returns its ID.
//
// [Ja] Create は新しいエピソードを挿入し、その ID を返す。
func (r *EpisodeRepository) Create(ctx context.Context, params CreateEpisodeParams) (model.EpisodeID, error) {
	id, err := r.queries.CreateEpisode(ctx, query.CreateEpisodeParams{
		WorkID:        int64(params.WorkID),
		Number:        nullStringFromPtr(params.Number),
		RawNumber:     nullFloat64FromPtr(params.RawNumber),
		Title:         nullStringFromPtr(params.Title),
		SortNumber:    params.SortNumber,
		PrevEpisodeID: nullInt64FromEpisodeID(params.PrevEpisodeID),
		AnimeID:       nullInt64FromAnimeID(params.AnimeID),
		UserID:        int64(params.UserID),
	})
	if err != nil {
		return 0, err
	}

	return model.EpisodeID(id), nil
}

// nullStringFromPtr maps an optional string onto the driver's nullable string, so a column
// an editor left empty is written as NULL instead of as an empty string.
//
// [Ja] nullStringFromPtr は任意の文字列をドライバの NULL 許容文字列に写像する。編集者が空の
// まま送ったカラムを、空文字列ではなく NULL として書くため。
func nullStringFromPtr(value *string) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *value, Valid: true}
}

// nullFloat64FromPtr maps an optional float onto the driver's nullable float.
//
// [Ja] nullFloat64FromPtr は任意の float をドライバの NULL 許容 float に写像する。
func nullFloat64FromPtr(value *float64) sql.NullFloat64 {
	if value == nil {
		return sql.NullFloat64{}
	}
	return sql.NullFloat64{Float64: *value, Valid: true}
}

// nullInt64FromEpisodeID maps an optional episode ID onto the driver's nullable integer.
//
// [Ja] nullInt64FromEpisodeID は任意のエピソード ID をドライバの NULL 許容整数に写像する。
func nullInt64FromEpisodeID(id *model.EpisodeID) sql.NullInt64 {
	if id == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(*id), Valid: true}
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

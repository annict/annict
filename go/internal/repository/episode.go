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

// pgErrCodeLockNotAvailableは、NOWAITのロック句が既にロック済みの行に当たったときに
// PostgreSQLが返すSQLSTATE (55P03, lock_not_available)。
const pgErrCodeLockNotAvailable = "55P03"

// ErrEpisodeLockUnavailableは、更新に必要なエピソード行を他のトランザクションが保持して
// いたため、待たずに試行を中断したことを表す。由来するNOWAIT句は、Railsのepisode -> workの
// 保存順序がGoのwork -> episodesの更新順序とデッドロックするのを防ぐためのものであり、
// 発火時にPostgreSQLはトランザクションを中断する。したがって呼び出し側は、失敗した
// ステートメントではなくトランザクション全体を再試行する。ここでドライバのエラーを翻訳するのは、
// Updateがsql.ErrNoRowsについて既にそうしているのと同じく、SQLSTATEをInfrastructure層に
// 閉じ込めるため。
var ErrEpisodeLockUnavailable = errors.New("エピソード行のロックを取得できませんでした")

// EpisodeRepositoryはepisodesテーブルおよび関連JOINへのデータアクセスを担う。
type EpisodeRepository struct {
	queries *query.Queries
}

// NewEpisodeRepositoryはEpisodeRepositoryを生成する。
func NewEpisodeRepository(queries *query.Queries) *EpisodeRepository {
	return &EpisodeRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいEpisodeRepositoryを返す。
func (r *EpisodeRepository) WithTx(tx *sql.Tx) *EpisodeRepository {
	return &EpisodeRepository{queries: r.queries.WithTx(tx)}
}

// DBEpisodeListParamsはAnnict DB画面の、ある作品のエピソード一覧1ページ分を
// 指定する。Pageは1始まり。
type DBEpisodeListParams struct {
	WorkID  model.WorkID
	Page    int32
	PerPage int32
}

// ListForDBはAnnict DB画面向けに、ある作品のエピソードを1ページ分、新しい
// エピソードから順に (Rails画面に合わせてsort_number降順で) ロードする。sort_numberが
// 同値でもページングが決定的になるようidをタイブレーカにする。
//
// 移行期間中の正本はepisodes側であるため、画面も派生であるanimes /
// anime_classificationsではなくepisodesを読む (派生側はRails経由の変更が毎時の
// 同期バッチ後にしか反映されない)。削除済みエピソードの除外はdeleted_atのみで行い
// (Railsの `without_deleted` スコープと同じ)、残った行はunpublished_at / deleted_atを
// 持つため、呼び出し側はmodel.Episode.DerivedStatusで表示用の状態を導出できる。
func (r *EpisodeRepository) ListForDB(ctx context.Context, params DBEpisodeListParams) ([]*model.Episode, error) {
	// 乗算の前に幅を広げる。呼び出し側はint32に収まるページ番号をすべて受け付けるが、
	// 1ページ100件ではint32同士の積がその範囲内で負に折り返し、PostgreSQLがそのOFFSET
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

// CountForDBはDB画面がその作品について一覧するエピソードの総件数を返す。
// ページネーションの総数が一覧の行と一致するよう、ListForDBと同じ絞り込みを使う。
func (r *EpisodeRepository) CountForDB(ctx context.Context, workID model.WorkID) (int64, error) {
	return r.queries.CountDBEpisodes(ctx, int64(workID))
}

// episodeFromDBListRowはAnnict DB一覧の行を *model.Episodeに変換する。行は
// 部分ロードで、animeマッピングカラムは選択せずゼロ値のまま残る。直前のエピソードの
// 2系統の話数はクエリ側の隣接行の導出に由来し、作品の最初のエピソードではnilになる。
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

// DBEpisodeEditTargetはAnnict DBのエピソード編集フォームが読み込むもの (編集対象の
// エピソードとその親作品) を表す。作品はページが必要とするカラム (見出しのtitleと
// サブナビのno_episodes) だけの部分ロードのため、エピソード自身のカラムだけを持つ
// model.Episodeに畳み込まず、エピソードと並べて持つ。
type DBEpisodeEditTarget struct {
	Episode *model.Episode
	Work    *model.Work
}

// GetForEditByIDはAnnict DBの編集フォームが編集するエピソードを、その親作品と一緒に
// 読み込む。削除済みのエピソードと、削除済み作品のエピソードはクエリ側で除外するため、
// (nil, nil) は編集できるエピソードがそのidに無いことを表す。
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

// GetForUpdateByIDは、Annict DBのエピソード更新が送信された値に加えて必要とするカラムを
// 読み込む。animesへの両書きが写像するがフォームでは編集しないtitle_roと状態のタイムスタンプ、
// およびエピソード自身のanimeと親作品のanime。削除済みのエピソードと削除済み作品のエピソード
// は、編集フォームと同じくクエリ側で除外するため、(nil, nil) は更新できるエピソードがそのidに
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

// UpdateEpisodeParamsはエピソード編集の1回の送信の属性を保持する。UserIDは作成時と同じく
// 記録される変更を作成者に帰属させる。
//
// WorkIDは編集用の事前読み取りで観測した親作品。Updateはエピソードの現在の親をロックして
// 一致を要求するため、その間に別作品へ移された行を古い並び順の前提で書かない。
//
// Versionは送信が前提とするupdated_at。nilは、フォームを開いた時点で行がupdated_atを持って
// いなかったことを表す。共有カラムがNULL許容であるため、これも1つの版として扱う。
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

// Updateはエピソード編集の1回の送信を適用し、どの行も一致しなかった場合にfalseを返す
// (エピソードが失われたか、フォームを開いてからupdated_atが進んだか)。呼び出し側はこれを再試行
// せず競合として扱うため、古い読み取りに対する送信が、その間に入った書き込みを上書きすることは
// ない。
//
// 親作品はUpdateDBEpisodeより前の独立した文でロックし、待機があった場合も後段の文がREAD
// COMMITTEDの新しいスナップショットを得るようにする。続いてUpdateDBEpisodeが書くか参照する
// 隣接行を列挙し、id昇順でNOWAITロックする。これにより1回の編集がロックする行数が、作品の
// エピソード数に比例して増えずに抑えられる。ロック取得失敗はErrEpisodeLockUnavailableとして
// 返り、UseCaseが中断されたトランザクション全体をrollbackして再試行する。
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

// isLockNotAvailableはerrがPostgreSQLのNOWAITのロック取得失敗かどうかを返す。
func isLockNotAvailable(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == pgErrCodeLockNotAvailable
}

// DBEpisodeArchiveTargetは非公開を確認する対象のエピソードと、そのページの見出し・
// サブナビが示す親作品。
type DBEpisodeArchiveTarget struct {
	Episode *model.Episode
	Work    *model.Work
}

// GetForArchiveByIDはAnnict DBの非公開エンドポイント群が対象とするエピソードを、その
// 親作品と一緒に読み込む。削除済みのエピソードと、削除済み作品のエピソードはクエリ側で除外する
// ため、(nil, nil) はそれらのエンドポイントが操作できるエピソードがそのidに無いことを表す。
// 返すエピソードは状態で絞らずタイムスタンプを持つため、どの状態を受け付けるかは呼び出し側が
// model.Episode.DerivedStatusで判断する (確認ページと非公開の送信は公開中のエピソード、
// 再公開の送信は非公開のエピソード)。
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

// ArchiveEpisodeParamsは1回の非公開の送信が非公開にするエピソードを指定する。WorkIDは
// 確認ページが前提とした親作品で、Archiveはエピソードが今もそこに属していることを要求する。
// カウンターの減算を、数えていた作品に当てるため。
type ArchiveEpisodeParams struct {
	ID     model.EpisodeID
	WorkID model.WorkID
}

// ArchiveEpisodeResultは、実際に非公開にしたエピソード行が持つanimeの写像を報告する。
// 参照モデルへ未マッピングのエピソードではAnimeIDはnil。
type ArchiveEpisodeResult struct {
	AnimeID *model.AnimeID
}

// Archiveはエピソードを非公開にし、どの行も一致しなかった場合にnilを返す (エピソード
// が失われた、親作品が削除された、確認ページを開いてから他者が非公開にした、またはそのページが
// 名指しした作品にもう属していない)。呼び出し側はこれを、確認ページ自身が返すのと同じnot found
// の応答に変換し、起きなかった書き込みを報告しない。成功時の結果は、トランザクション前の射影では
// なく、更新した行が返したanimeの写像を運ぶ。
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

// UnarchiveEpisodeParamsは1回の再公開の送信が公開に戻すエピソードを指定する。WorkIDは
// 送信元の一覧が名指しした親作品で、Unarchiveはエピソードが今もそこに属していることを要求する。
// カウンターの加算を、その行を数えていなかった作品に当てるため。
type UnarchiveEpisodeParams struct {
	ID     model.EpisodeID
	WorkID model.WorkID
}

// UnarchiveEpisodeResultは、実際に再公開したエピソード行が持つanimeの写像を報告する。
// 参照モデルへ未マッピングのエピソードではAnimeIDはnil。
type UnarchiveEpisodeResult struct {
	AnimeID *model.AnimeID
}

// Unarchiveは非公開のエピソードを再公開し、どの行も一致しなかった場合にnilを返す
// (エピソードが失われた、親作品が削除された、一覧を開いてから他者が再公開した、またはその一覧が
// 名指しした作品にもう属していない)。呼び出し側はこれを、一覧が表示できないエピソードに対して
// 返すのと同じnot foundの応答に変換し、起きなかった書き込みを報告しない。成功時の結果は、
// トランザクション前の射影ではなく、更新した行が返したanimeの写像を運ぶ。
func (r *EpisodeRepository) Unarchive(ctx context.Context, params UnarchiveEpisodeParams) (*UnarchiveEpisodeResult, error) {
	row, err := r.queries.UnarchiveDBEpisode(ctx, query.UnarchiveDBEpisodeParams{
		ID:     int64(params.ID),
		WorkID: int64(params.WorkID),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	result := &UnarchiveEpisodeResult{}
	if row.AnimeID.Valid {
		animeID := model.AnimeID(row.AnimeID.Int64)
		result.AnimeID = &animeID
	}

	return result, nil
}

// GetForDeleteByIDはAnnict DBの削除エンドポイントが対象とするエピソードを読み込む。
// 削除済みのエピソードと、削除済み作品のエピソードはクエリ側で除外するため、(nil, nil) はその
// エンドポイントが操作できるエピソードがそのidに無いことを表す。公開中のエピソードも非公開の
// エピソードも返す。削除はどちらも受け付けるため、GetForArchiveByIDと違い呼び出し側が判断する
// 状態のタイムスタンプは運ばない。返すエピソードは、削除を束縛し、成功時に着地する親作品を持つ。
func (r *EpisodeRepository) GetForDeleteByID(ctx context.Context, id model.EpisodeID) (*model.Episode, error) {
	row, err := r.queries.GetEpisodeForDeleteByID(ctx, int64(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &model.Episode{
		ID:     model.EpisodeID(row.ID),
		WorkID: model.WorkID(row.WorkID),
	}, nil
}

// DeleteEpisodeParamsは1回の削除の送信がソフトデリートするエピソードを指定する。WorkID
// は送信元の一覧が名指しした親作品で、Deleteはエピソードが今もそこに属していることを要求する。
// カウンターの減算を、その行を数えていた作品に当てるため。
type DeleteEpisodeParams struct {
	ID     model.EpisodeID
	WorkID model.WorkID
}

// DeleteEpisodeResultは、実際に削除したエピソード行が持つanimeの写像を報告する。
// 参照モデルへ未マッピングのエピソードではAnimeIDはnil。
type DeleteEpisodeResult struct {
	AnimeID *model.AnimeID
}

// Deleteはエピソードをソフトデリートし、どの行も一致しなかった場合にnilを返す
// (エピソードが失われた、親作品が削除された、一覧を開いてから他者が削除した、またはその一覧が
// 名指しした作品にもう属していない)。呼び出し側はこれを、一覧が表示できないエピソードに対して
// 返すのと同じnot foundの応答に変換し、起きなかった書き込みを報告しない。成功時の結果は、
// トランザクション前の射影ではなく、更新した行が返したanimeの写像を運ぶ。
func (r *EpisodeRepository) Delete(ctx context.Context, params DeleteEpisodeParams) (*DeleteEpisodeResult, error) {
	row, err := r.queries.DeleteDBEpisode(ctx, query.DeleteDBEpisodeParams{
		ID:     int64(params.ID),
		WorkID: int64(params.WorkID),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	result := &DeleteEpisodeResult{}
	if row.AnimeID.Valid {
		animeID := model.AnimeID(row.AnimeID.Int64)
		result.AnimeID = &animeID
	}

	return result, nil
}

// CreateEpisodeParamsはエピソード作成時の属性を保持する。編集者が入力しないカラム
// (title_ro / title_en / カウンターキャッシュ / 状態のタイムスタンプ) はカラムの既定値のまま
// にするため、作成されたエピソードは公開状態で始まる。
//
// AnimeIDはepisodes.anime_idのマッピングカラムで、エピソードのanimeを併せて作成した
// ときに入り、親作品が未マッピングのあいだはnilのままになる (animeは後続の同期が作る)。
// PrevEpisodeIDは挿入時点でこのエピソードの直前に来るエピソードを指す。
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

// Createは新しいエピソードを挿入し、そのIDを返す。
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

// nullStringFromPtrは任意の文字列をドライバのNULL許容文字列に写像する。編集者が空の
// まま送ったカラムを、空文字列ではなくNULLとして書くため。
func nullStringFromPtr(value *string) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *value, Valid: true}
}

// nullFloat64FromPtrは任意のfloatをドライバのNULL許容floatに写像する。
func nullFloat64FromPtr(value *float64) sql.NullFloat64 {
	if value == nil {
		return sql.NullFloat64{}
	}
	return sql.NullFloat64{Float64: *value, Valid: true}
}

// nullInt64FromEpisodeIDは任意のエピソードIDをドライバのNULL許容整数に写像する。
func nullInt64FromEpisodeID(id *model.EpisodeID) sql.NullInt64 {
	if id == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(*id), Valid: true}
}

// ListForAnimeSyncByIDsは指定IDのepisodesを、フェーズ2のリコンシリエーションが
// animes / anime_classificationsに写像するカラム (episodes.anime_idのマッピングカラムを
// 含む) を射影してロードする。親作品のanime_idはepisodes.work_idのJOINで解決して
// ParentAnimeIDとして返すため、エピソード同期が親を行単位で引く必要はない。行はid昇順で、
// 存在しないIDは黙って除外される。空入力ではクエリせず空スライスを返す。
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

// ListIDsAfterはafterIDより大きいepisode IDを昇順で最大batchSize件返す。
// フェーズ2のバッチジョブ (タスク2-4) がepisodesテーブル全体をページ単位で走査する
// ためのkeysetページネーションの基本操作で、最初のページはafterID=0を渡し、以降は
// 直前に返った末尾のidをカーソルにして次ページを引き、空ページで終端を知る。
// LIMIT/OFFSETではなくkeyset (id > カーソル) を使うのは、バッチが大テーブルを走査する
// ため。OFFSETはページごとにスキップ分を読み直し、全件走査ではO(n^2) に劣化する。
func (r *EpisodeRepository) ListIDsAfter(ctx context.Context, afterID model.EpisodeID, batchSize int) ([]model.EpisodeID, error) {
	rows, err := r.queries.ListEpisodeIDsAfter(ctx, query.ListEpisodeIDsAfterParams{
		AfterID: int64(afterID),
		// batchSizeは小さく上限のあるページサイズ (既定1000) でint32上限には達しない。
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

// UpdateAnimeIDはepisodes.anime_idマッピングカラムを書き戻し、エピソードを
// 指定アニメへ同期済みとして印付ける。updated_atは意図的に触れず、正本側の行への記帳
// 書き込みが内容変更と取り違えられないようにする。
func (r *EpisodeRepository) UpdateAnimeID(ctx context.Context, episodeID model.EpisodeID, animeID model.AnimeID) error {
	return r.queries.UpdateEpisodeAnimeID(ctx, query.UpdateEpisodeAnimeIDParams{
		ID:      int64(episodeID),
		AnimeID: sql.NullInt64{Int64: int64(animeID), Valid: true},
	})
}

// episodeFromAnimeSyncRowはanime同期のquery行を *model.Episodeに変換する。
// NULL許容カラム (title / number / raw_number / unpublished_at / deleted_at /
// anime_id / parent_anime_id) はポインタで持ち、同期UseCaseが「未設定」とゼロ値を
// 区別できるようにする。workFromAnimeSyncRowがworksを扱うのと同じ方針。
func episodeFromAnimeSyncRow(row query.ListEpisodesForAnimeSyncByIDsRow) *model.Episode {
	episode := &model.Episode{
		ID:         model.EpisodeID(row.ID),
		WorkID:     model.WorkID(row.WorkID),
		TitleRo:    row.TitleRo,
		TitleEn:    row.TitleEn,
		SortNumber: row.SortNumber,
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

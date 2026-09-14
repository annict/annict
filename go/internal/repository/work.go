package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// WorkRepositoryはworksテーブルおよび関連JOINへのデータアクセスを担う。
type WorkRepository struct {
	queries *query.Queries
}

func NewWorkRepository(queries *query.Queries) *WorkRepository {
	return &WorkRepository{queries: queries}
}

func (r *WorkRepository) GetByID(ctx context.Context, id model.WorkID) (*model.Work, error) {
	row, err := r.queries.GetWorkByID(ctx, int64(id))
	if err != nil {
		return nil, err
	}
	return workFromGetByIDRow(row), nil
}

func workFromGetByIDRow(row query.GetWorkByIDRow) *model.Work {
	work := &model.Work{
		ID:                  model.WorkID(row.ID),
		Title:               row.Title,
		TitleEn:             row.TitleEn,
		RecommendedImageURL: row.RecommendedImageUrl,
		WatchersCount:       row.WatchersCount,
	}
	if row.TitleKana != "" {
		titleKana := row.TitleKana
		work.TitleKana = &titleKana
	}
	applyNullableWorkFields(work, row.SeasonYear, row.SeasonName, row.CreatedAt)
	return work
}

// GetForStateChangeByIDはAnnict DB管理画面の作品の状態変更 (非公開・公開・削除) の
// 確認画面が必要とするカラムを読み込む: 表示するタイトルと、呼び出し側が現在の状態を導出し、
// その画面が操作できない作品を弾けるようにするための作品状態のsource (unpublished_at /
// deleted_at)。該当するworkが無い場合は (nil, nil) を返す。
func (r *WorkRepository) GetForStateChangeByID(ctx context.Context, id model.WorkID) (*model.Work, error) {
	row, err := r.queries.GetWorkForStateChangeByID(ctx, int64(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	work := &model.Work{
		ID:    model.WorkID(row.ID),
		Title: row.Title,
	}
	if row.UnpublishedAt.Valid {
		unpublishedAt := row.UnpublishedAt.Time
		work.UnpublishedAt = &unpublishedAt
	}
	if row.DeletedAt.Valid {
		deletedAt := row.DeletedAt.Time
		work.DeletedAt = &deletedAt
	}
	return work, nil
}

// GetForEditByIDはAnnict DB管理画面の編集フォームが各フィールドを初期表示
// するために必要なworksの全カラムを読み込む。該当するworkが無い場合は
// (nil, nil) を返す。
func (r *WorkRepository) GetForEditByID(ctx context.Context, id model.WorkID) (*model.Work, error) {
	row, err := r.queries.GetWorkForEditByID(ctx, int64(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return workFromGetForEditByIDRow(row), nil
}

func workFromGetForEditByIDRow(row query.GetWorkForEditByIDRow) *model.Work {
	work := &model.Work{
		ID:                    model.WorkID(row.ID),
		Title:                 row.Title,
		TitleAlter:            row.TitleAlter,
		TitleEn:               row.TitleEn,
		TitleAlterEn:          row.TitleAlterEn,
		Media:                 row.Media,
		OfficialSiteURL:       row.OfficialSiteUrl,
		OfficialSiteURLEn:     row.OfficialSiteUrlEn,
		WikipediaURL:          row.WikipediaUrl,
		WikipediaURLEn:        row.WikipediaUrlEn,
		Synopsis:              row.Synopsis,
		SynopsisSource:        row.SynopsisSource,
		SynopsisEn:            row.SynopsisEn,
		SynopsisSourceEn:      row.SynopsisSourceEn,
		StartEpisodeRawNumber: row.StartEpisodeRawNumber,
		NoEpisodes:            row.NoEpisodes,
	}
	if row.TitleKana != "" {
		titleKana := row.TitleKana
		work.TitleKana = &titleKana
	}
	applyNullableWorkFields(work, row.SeasonYear, row.SeasonName, sql.NullTime{})
	if row.StartedOn.Valid {
		startedOn := row.StartedOn.Time
		work.StartedOn = &startedOn
	}
	if row.EndedOn.Valid {
		endedOn := row.EndedOn.Time
		work.EndedOn = &endedOn
	}
	if row.TwitterUsername.Valid {
		twitterUsername := row.TwitterUsername.String
		work.TwitterUsername = &twitterUsername
	}
	if row.TwitterHashtag.Valid {
		twitterHashtag := row.TwitterHashtag.String
		work.TwitterHashtag = &twitterHashtag
	}
	if row.ScTid.Valid {
		scTid := row.ScTid.Int32
		work.ScTid = &scTid
	}
	if row.MalAnimeID.Valid {
		malAnimeID := row.MalAnimeID.Int32
		work.MalAnimeID = &malAnimeID
	}
	if row.ManualEpisodesCount.Valid {
		manualEpisodesCount := row.ManualEpisodesCount.Int32
		work.ManualEpisodesCount = &manualEpisodesCount
	}
	if row.NumberFormatID.Valid {
		numberFormatID := model.NumberFormatID(row.NumberFormatID.Int64)
		work.NumberFormatID = &numberFormatID
	}
	if row.UpdatedAt.Valid {
		updatedAt := row.UpdatedAt.Time
		work.UpdatedAt = &updatedAt
	}
	return work
}

// DBEpisodeListWorkはAnnict DBのエピソード一覧ページの親作品を表す。作品そのものと、
// ページの自動生成の案内が報告する2つの値を持つ。これらの値はworksのカラムではなく
// 作品のエピソード・スロットの集計であるため、model.Workに畳み込まず作品と並べて持つ。
type DBEpisodeListWork struct {
	Work *model.Work
	// PublishedEpisodeCountは作品のエピソードのうち、非公開でも削除済みでもないもの
	// (Railsのonly_keptスコープ) の件数。一覧自体の総件数とは異なる (一覧は非公開の
	// エピソードも表示するため総件数に含める)。
	PublishedEpisodeCount int64
	// MaxGeneratableEpisodeNumberは作品が持つ有効なスロットの最大numberで、しょぼい
	// カレンダー由来の自動生成が到達する話数を表す。そのようなスロットが無い作品では0。
	MaxGeneratableEpisodeNumber int64
}

// GetForEpisodeListByIDはAnnict DBのエピソード一覧が親作品から必要とするもの
// (ページ見出しに使うtitle、共有の作品サブナビが使うno_episodes、自動生成の案内が報告する
// 集計値) を読み込む。deleted_atが入った作品はクエリ側で除外する (エピソード一覧が使うRails
// のWork.without_deleted.findと同じ)。そのため (nil, nil) は一覧を出せる作品がそのidに
// 無いことを表す。
func (r *WorkRepository) GetForEpisodeListByID(ctx context.Context, id model.WorkID) (*DBEpisodeListWork, error) {
	row, err := r.queries.GetWorkForEpisodeListByID(ctx, int64(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	work := &model.Work{
		ID:         model.WorkID(row.ID),
		Title:      row.Title,
		NoEpisodes: row.NoEpisodes,
	}
	if row.ManualEpisodesCount.Valid {
		manualEpisodesCount := row.ManualEpisodesCount.Int32
		work.ManualEpisodesCount = &manualEpisodesCount
	}

	return &DBEpisodeListWork{
		Work:                        work,
		PublishedEpisodeCount:       row.PublishedEpisodeCount,
		MaxGeneratableEpisodeNumber: row.MaxGeneratableEpisodeNumber,
	}, nil
}

// DBEpisodeFormWorkはAnnict DBエピソードフォームが描画する親作品と、Rails互換の
// 手動作成状態を保持する。
type DBEpisodeFormWork struct {
	Work                *model.Work
	ManualCreationState model.ManualEpisodeCreationState
}

// GetForEpisodeFormByIDはAnnict DBのエピソードフォームが親作品から必要とするもの
// (ページ見出しに使うtitle、共有の作品サブナビが使うno_episodes、手動作成を制限する理由)
// を読み込む。削除済みの作品はクエリ側で除外するため、(nil, nil) はフォームを出せる作品が
// そのidに無いことを表す。
func (r *WorkRepository) GetForEpisodeFormByID(ctx context.Context, id model.WorkID) (*DBEpisodeFormWork, error) {
	row, err := r.queries.GetWorkForEpisodeFormByID(ctx, int64(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &DBEpisodeFormWork{
		Work: &model.Work{
			ID:         model.WorkID(row.ID),
			Title:      row.Title,
			NoEpisodes: row.NoEpisodes,
		},
		ManualCreationState: model.ManualEpisodeCreationState{
			EpisodesFilled: row.EpisodesFilled.Valid && row.EpisodesFilled.Bool,
			SlotsExist:     row.SlotsExist,
		},
	}, nil
}

// DBEpisodeCreateWorkはAnnict DBのエピソード一括作成の親作品を表す。作品そのもの
// (idとマッピング先のanime) と、新規行の採番の起点になる値を持つ。起点の値はworksの
// カラムではなく作品の既存エピソードの集計であるため、model.Workに畳み込まず作品と並べて
// 持つ。
type DBEpisodeCreateWork struct {
	Work                *model.Work
	ManualCreationState model.ManualEpisodeCreationState
	// EpisodeCountは作品のエピソードを、非公開のものも削除済みのものも含めて数える
	// (採番の元にしたRailsのフォームと同じ)。
	EpisodeCount int64
	// LatestEpisodeはsort_numberが最大のエピソードで、最初に作る行が直前の
	// エピソードとして名指しする。作品がエピソードを持たないあいだはnil。
	LatestEpisode *DBEpisodeSortAnchor
}

// ExistsForEpisodeCreateByIDは指定された未削除作品が存在するかを返す。作成ユースケースは
// 入力行をパースする前にこの軽量な確認を行い、書き込みトランザクション内で同じ作品をロック
// して再確認する。
func (r *WorkRepository) ExistsForEpisodeCreateByID(ctx context.Context, id model.WorkID) (bool, error) {
	return r.queries.ExistsWorkForEpisodeCreateByID(ctx, int64(id))
}

// LockForEpisodeCreateByIDは現在のトランザクションで対象作品をロックする。このロックで
// 同じ作品への一括作成を直列化する。falseは作品が存在しないか削除済みであることを表す。
func (r *WorkRepository) LockForEpisodeCreateByID(ctx context.Context, id model.WorkID) (bool, error) {
	_, err := r.queries.LockWorkForEpisodeCreateByID(ctx, int64(id))
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// DBEpisodeSortAnchorは一括作成の採番が必要とする分だけに絞ったエピソード。id (次の
// エピソードのprev_episode_idとして保存する) とsort_number (新規作成したエピソードが最大
// の座を引き継ぐかの判定に使う) を持つ。
type DBEpisodeSortAnchor struct {
	ID         model.EpisodeID
	SortNumber int32
}

// GetForEpisodeCreateByIDはエピソード一括作成の親作品を、採番の起点と併せて読み込む。
// 削除済みの作品はクエリ側で除外する (createアクションが使うRailsの
// Work.without_deleted.findと同じ)。そのため (nil, nil) はエピソードを作成できる作品がその
// idに無いことを表す。
func (r *WorkRepository) GetForEpisodeCreateByID(ctx context.Context, id model.WorkID) (*DBEpisodeCreateWork, error) {
	row, err := r.queries.GetWorkForEpisodeCreateByID(ctx, int64(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	work := &model.Work{ID: model.WorkID(row.ID)}
	if row.AnimeID.Valid {
		animeID := model.AnimeID(row.AnimeID.Int64)
		work.AnimeID = &animeID
	}

	result := &DBEpisodeCreateWork{
		Work:         work,
		EpisodeCount: row.EpisodeCount,
		ManualCreationState: model.ManualEpisodeCreationState{
			EpisodesFilled: row.EpisodesFilled.Valid && row.EpisodesFilled.Bool,
			SlotsExist:     row.SlotsExist,
		},
	}
	if row.LatestEpisodeID != 0 {
		result.LatestEpisode = &DBEpisodeSortAnchor{
			ID:         model.EpisodeID(row.LatestEpisodeID),
			SortNumber: row.LatestSortNumber,
		}
	}

	return result, nil
}

// IncrementEpisodesCountは公開エピソードを作成した後、Railsのカウンターキャッシュと
// touchの副作用を適用する。呼び出し元は既に作品をロックしている。
func (r *WorkRepository) IncrementEpisodesCount(ctx context.Context, workID model.WorkID, createdCount int32) error {
	affected, err := r.queries.IncrementWorkEpisodesCount(ctx, query.IncrementWorkEpisodesCountParams{
		CreatedCount: createdCount,
		WorkID:       int64(workID),
	})
	if err != nil {
		return err
	}
	if affected != 1 {
		return errors.New("作品のエピソード件数を更新できませんでした")
	}
	return nil
}

// GetPopularは人気作品を返す。戻り値の各 *model.Workは呼び出しごとに新規生成されるため、
// 呼び出し側 (主にUseCase) がCasts / Staffsなどの関連エンティティを後付けで
// 代入する用法を許容している。Repositoryでキャッシュやプール再利用を導入する
// 場合はこの前提を見直すこと。
func (r *WorkRepository) GetPopular(ctx context.Context) ([]*model.Work, error) {
	rows, err := r.queries.GetPopularWorks(ctx)
	if err != nil {
		return nil, err
	}

	works := make([]*model.Work, len(rows))
	for i, row := range rows {
		works[i] = workFromPopularRow(row)
	}
	return works, nil
}

func workFromPopularRow(row query.GetPopularWorksRow) *model.Work {
	work := &model.Work{
		ID:                  model.WorkID(row.ID),
		Title:               row.Title,
		TitleEn:             row.TitleEn,
		RecommendedImageURL: row.RecommendedImageUrl,
		WatchersCount:       row.WatchersCount,
	}
	applyImageData(work, row.ImageData)
	applyNullableWorkFields(work, row.SeasonYear, row.SeasonName, row.CreatedAt)
	return work
}

// sqlc生成型のnullableカラムを *model.Workにマッピングするヘルパー。
// SeasonYear / SeasonName / CreatedAtは複数のrow型で共通するため、
// 呼び出し元ごとに揺れないよう変換ロジックを1箇所に集約している。
// CreatedAtをロードしない呼び出し元はsql.NullTime{} を渡すことでスキップできる。
func applyNullableWorkFields(work *model.Work, seasonYear, seasonName sql.NullInt32, createdAt sql.NullTime) {
	if seasonYear.Valid {
		v := seasonYear.Int32
		work.SeasonYear = &v
	}
	if seasonName.Valid {
		v := seasonName.Int32
		work.SeasonName = &v
	}
	if createdAt.Valid {
		work.CreatedAt = createdAt.Time
	}
}

// work_images.image_dataカラムを *model.Workにマッピングするヘルパー。
// LEFT JOINでwork_images行が一致しない場合はValid=falseとなり、
// ImageDataは空文字列のままになる。
func applyImageData(work *model.Work, imageData sql.NullString) {
	if imageData.Valid {
		work.ImageData = imageData.String
	}
}

func (r *WorkRepository) WithTx(tx *sql.Tx) *WorkRepository {
	return &WorkRepository{queries: r.queries.WithTx(tx)}
}

type DBWorkListParams struct {
	FilterNoEpisodes bool
	FilterNoImage    bool
	FilterNoSeason   bool
	FilterNoSlots    bool
	SeasonYear       *int32
	SeasonName       *int32
	// SeasonYears / SeasonNamesはリリース時期の複数選択フィルタで照合する
	// (年, 季節) ペアを表す並列配列。空スライスならフィルタは無効。両スライスは同じ
	// 長さで、SeasonYearsのi番目がSeasonNamesのi番目と対になる。
	SeasonYears []int32
	SeasonNames []int32
	Page        int32
	PerPage     int32
}

func (r *WorkRepository) ListForDB(ctx context.Context, params DBWorkListParams) ([]*model.Work, error) {
	// 乗算の前に幅を広げる。呼び出し側はint32に収まるページ番号をすべて受け付けるが、
	// 1ページ100件ではint32同士の積がその範囲内で負に折り返し、PostgreSQLがそのOFFSET
	// を拒否するため。
	offset := int64(params.Page-1) * int64(params.PerPage)

	rows, err := r.queries.ListDBWorks(ctx, query.ListDBWorksParams{
		FilterNoEpisodes: sql.NullBool{Bool: params.FilterNoEpisodes, Valid: params.FilterNoEpisodes},
		FilterNoImage:    sql.NullBool{Bool: params.FilterNoImage, Valid: params.FilterNoImage},
		FilterNoSeason:   sql.NullBool{Bool: params.FilterNoSeason, Valid: params.FilterNoSeason},
		FilterNoSlots:    sql.NullBool{Bool: params.FilterNoSlots, Valid: params.FilterNoSlots},
		SeasonYear:       nullInt32FromPtr(params.SeasonYear),
		SeasonName:       nullInt32FromPtr(params.SeasonName),
		SeasonYears:      params.SeasonYears,
		SeasonNames:      params.SeasonNames,
		PerPage:          params.PerPage,
		PageOffset:       offset,
	})
	if err != nil {
		return nil, err
	}

	works := make([]*model.Work, len(rows))
	for i, row := range rows {
		work := &model.Work{
			ID:            model.WorkID(row.ID),
			Title:         row.Title,
			TitleEn:       row.TitleEn,
			Media:         row.Media,
			WatchersCount: row.WatchersCount,
		}
		if row.TitleKana != "" {
			titleKana := row.TitleKana
			work.TitleKana = &titleKana
		}
		if row.ScTid.Valid {
			scTid := row.ScTid.Int32
			work.ScTid = &scTid
		}
		if row.MalAnimeID.Valid {
			malAnimeID := row.MalAnimeID.Int32
			work.MalAnimeID = &malAnimeID
		}
		if row.UnpublishedAt.Valid {
			unpublishedAt := row.UnpublishedAt.Time
			work.UnpublishedAt = &unpublishedAt
		}
		if row.DeletedAt.Valid {
			deletedAt := row.DeletedAt.Time
			work.DeletedAt = &deletedAt
		}
		applyImageData(work, row.ImageData)
		applyNullableWorkFields(work, row.SeasonYear, row.SeasonName, sql.NullTime{})
		works[i] = work
	}
	return works, nil
}

func (r *WorkRepository) CountForDB(ctx context.Context, params DBWorkListParams) (int64, error) {
	return r.queries.CountDBWorks(ctx, query.CountDBWorksParams{
		FilterNoEpisodes: sql.NullBool{Bool: params.FilterNoEpisodes, Valid: params.FilterNoEpisodes},
		FilterNoImage:    sql.NullBool{Bool: params.FilterNoImage, Valid: params.FilterNoImage},
		FilterNoSeason:   sql.NullBool{Bool: params.FilterNoSeason, Valid: params.FilterNoSeason},
		FilterNoSlots:    sql.NullBool{Bool: params.FilterNoSlots, Valid: params.FilterNoSlots},
		SeasonYear:       nullInt32FromPtr(params.SeasonYear),
		SeasonName:       nullInt32FromPtr(params.SeasonName),
		SeasonYears:      params.SeasonYears,
		SeasonNames:      params.SeasonNames,
	})
}

// ExistsKeptByTitleはtitleを使っている生存中のworkがあるかを返す。「生存中」は
// RailsのWorkの一意性スコープ (only_kept) と同じくdeleted_atとunpublished_atが
// ともにNULLの行を指し、非公開・削除済みの作品はタイトルを塞がない。excludeIDは編集中の
// workを指し、タイトルを変えない更新が自分自身と衝突しないようにする。作成時はnilを渡す。
func (r *WorkRepository) ExistsKeptByTitle(ctx context.Context, title string, excludeID *model.WorkID) (bool, error) {
	params := query.ExistsKeptWorkByTitleParams{Title: title}
	if excludeID != nil {
		params.ExcludeID = sql.NullInt64{Int64: int64(*excludeID), Valid: true}
	}
	return r.queries.ExistsKeptWorkByTitle(ctx, params)
}

type CreateWorkParams struct {
	Title                 string
	TitleKana             string
	TitleAlter            string
	TitleEn               string
	TitleAlterEn          string
	Media                 int32
	SeasonYear            sql.NullInt32
	SeasonName            sql.NullInt32
	StartedOn             sql.NullTime
	EndedOn               sql.NullTime
	OfficialSiteURL       string
	OfficialSiteURLEn     string
	WikipediaURL          string
	WikipediaURLEn        string
	TwitterUsername       sql.NullString
	TwitterHashtag        sql.NullString
	ScTid                 sql.NullInt32
	MalAnimeID            sql.NullInt32
	Synopsis              string
	SynopsisSource        string
	SynopsisEn            string
	SynopsisSourceEn      string
	ManualEpisodesCount   sql.NullInt32
	StartEpisodeRawNumber float64
	NumberFormatID        sql.NullInt64
	NoEpisodes            bool
}

func (r *WorkRepository) Create(ctx context.Context, params CreateWorkParams) (model.WorkID, error) {
	id, err := r.queries.CreateWork(ctx, query.CreateWorkParams{
		Title:                 params.Title,
		TitleKana:             params.TitleKana,
		TitleAlter:            params.TitleAlter,
		TitleEn:               params.TitleEn,
		TitleAlterEn:          params.TitleAlterEn,
		Media:                 params.Media,
		SeasonYear:            params.SeasonYear,
		SeasonName:            params.SeasonName,
		StartedOn:             params.StartedOn,
		EndedOn:               params.EndedOn,
		OfficialSiteUrl:       params.OfficialSiteURL,
		OfficialSiteUrlEn:     params.OfficialSiteURLEn,
		WikipediaUrl:          params.WikipediaURL,
		WikipediaUrlEn:        params.WikipediaURLEn,
		TwitterUsername:       params.TwitterUsername,
		TwitterHashtag:        params.TwitterHashtag,
		ScTid:                 params.ScTid,
		MalAnimeID:            params.MalAnimeID,
		Synopsis:              params.Synopsis,
		SynopsisSource:        params.SynopsisSource,
		SynopsisEn:            params.SynopsisEn,
		SynopsisSourceEn:      params.SynopsisSourceEn,
		ManualEpisodesCount:   params.ManualEpisodesCount,
		StartEpisodeRawNumber: params.StartEpisodeRawNumber,
		NumberFormatID:        params.NumberFormatID,
		NoEpisodes:            params.NoEpisodes,
	})
	if err != nil {
		return 0, err
	}
	return model.WorkID(id), nil
}

// UpdateWorkParamsはAnnict DB管理画面の作品編集フォームで編集可能なworks
// カラムをIDで特定して保持する。CreateWorkParams (作成フォームが書くカラムそのもの) を
// 埋め込み、対象IDを足すことで、作成と更新の両書き込み経路が1つのフィールド集合を
// 共有する。status / anime_idと派生カウンターは意図的に触れない (status変更はアーカイブ/
// 削除フローの管轄で、anime_idは同期が持つマッピングカラム)。
type UpdateWorkParams struct {
	ID model.WorkID
	// Versionは送信が前提とするupdated_at。nilは、フォームを開いた時点で行が
	// updated_atを持っていなかったことを表す。共有カラムがNULL許容であるため、これも1つの
	// 版として照合する。
	Version *time.Time
	CreateWorkParams
}

// Updateは指定IDのworkの編集可能カラムを上書きしてupdated_atを更新し、どの行も
// 一致しなかった場合にfalseを返す (workが失われたか、編集フォームを開いてからupdated_atが
// 進んだか)。呼び出し側はこれを再試行せず競合として扱うため、古い読み取りに対する送信が、その間に
// 入った書き込みを上書きすることはない。
func (r *WorkRepository) Update(ctx context.Context, params UpdateWorkParams) (bool, error) {
	rows, err := r.queries.UpdateWork(ctx, query.UpdateWorkParams{
		ID:                    int64(params.ID),
		Version:               nullTimeFromPtr(params.Version),
		Title:                 params.Title,
		TitleKana:             params.TitleKana,
		TitleAlter:            params.TitleAlter,
		TitleEn:               params.TitleEn,
		TitleAlterEn:          params.TitleAlterEn,
		Media:                 params.Media,
		SeasonYear:            params.SeasonYear,
		SeasonName:            params.SeasonName,
		StartedOn:             params.StartedOn,
		EndedOn:               params.EndedOn,
		OfficialSiteUrl:       params.OfficialSiteURL,
		OfficialSiteUrlEn:     params.OfficialSiteURLEn,
		WikipediaUrl:          params.WikipediaURL,
		WikipediaUrlEn:        params.WikipediaURLEn,
		TwitterUsername:       params.TwitterUsername,
		TwitterHashtag:        params.TwitterHashtag,
		ScTid:                 params.ScTid,
		MalAnimeID:            params.MalAnimeID,
		Synopsis:              params.Synopsis,
		SynopsisSource:        params.SynopsisSource,
		SynopsisEn:            params.SynopsisEn,
		SynopsisSourceEn:      params.SynopsisSourceEn,
		ManualEpisodesCount:   params.ManualEpisodesCount,
		StartEpisodeRawNumber: params.StartEpisodeRawNumber,
		NumberFormatID:        params.NumberFormatID,
		NoEpisodes:            params.NoEpisodes,
	})
	if err != nil {
		return false, err
	}

	return rows > 0, nil
}

// ListForAnimeSyncByIDsは指定IDのworksを、フェーズ2のリコンシリエーションが
// animes / anime_classificationsに写像するカラム (works.anime_idのマッピングカラムを
// 含む) を射影してロードする。行はid昇順で、存在しないIDは黙って除外される。
// 空入力ではクエリせず空スライスを返す。
func (r *WorkRepository) ListForAnimeSyncByIDs(ctx context.Context, workIDs []model.WorkID) ([]*model.Work, error) {
	if len(workIDs) == 0 {
		return []*model.Work{}, nil
	}

	ids := make([]int64, len(workIDs))
	for i, id := range workIDs {
		ids[i] = int64(id)
	}

	rows, err := r.queries.ListWorksForAnimeSyncByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	works := make([]*model.Work, len(rows))
	for i, row := range rows {
		works[i] = workFromAnimeSyncRow(row)
	}
	return works, nil
}

// ListForSatelliteSyncByIDsは指定IDのworksを、フェーズ2のリコンシリエーションが
// 別表 (外部ID / リンク / 公式アカウント / ハッシュタグ / 季節 / イベント) に写像する
// カラムとworks.anime_idのマッピングカラムを射影してロードする。行はid昇順で、存在
// しないIDは黙って除外される。空入力ではクエリせず空スライスを返す。
func (r *WorkRepository) ListForSatelliteSyncByIDs(ctx context.Context, workIDs []model.WorkID) ([]*model.Work, error) {
	if len(workIDs) == 0 {
		return []*model.Work{}, nil
	}

	ids := make([]int64, len(workIDs))
	for i, id := range workIDs {
		ids[i] = int64(id)
	}

	rows, err := r.queries.ListWorksForSatelliteSyncByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	works := make([]*model.Work, len(rows))
	for i, row := range rows {
		works[i] = workFromSatelliteSyncRow(row)
	}
	return works, nil
}

// ListIDsAfterはafterIDより大きいwork IDを昇順で最大batchSize件返す。
// フェーズ2のバッチジョブ (タスク2-4) がworksテーブル全体をページ単位で走査する
// ためのkeysetページネーションの基本操作で、最初のページはafterID=0を渡し、以降は
// 直前に返った末尾のidをカーソルにして次ページを引き、空ページで終端を知る。
// LIMIT/OFFSETではなくkeyset (id > カーソル) を使うのは、バッチが大テーブルを走査する
// ため。OFFSETはページごとにスキップ分を読み直し、全件走査ではO(n^2) に劣化する。
func (r *WorkRepository) ListIDsAfter(ctx context.Context, afterID model.WorkID, batchSize int) ([]model.WorkID, error) {
	rows, err := r.queries.ListWorkIDsAfter(ctx, query.ListWorkIDsAfterParams{
		AfterID: int64(afterID),
		// batchSizeは小さく上限のあるページサイズ (既定1000) でint32上限には達しない。
		BatchSize: int32(batchSize), // #nosec G115
	})
	if err != nil {
		return nil, err
	}

	ids := make([]model.WorkID, len(rows))
	for i, id := range rows {
		ids[i] = model.WorkID(id)
	}
	return ids, nil
}

// UpdateAnimeIDはworks.anime_idマッピングカラムを書き戻し、作品を指定アニメへ
// 同期済みとして印付ける。updated_atは意図的に触れず、正本側の行への記帳書き込みが
// 内容変更と取り違えられないようにする。
func (r *WorkRepository) UpdateAnimeID(ctx context.Context, workID model.WorkID, animeID model.AnimeID) error {
	return r.queries.UpdateWorkAnimeID(ctx, query.UpdateWorkAnimeIDParams{
		ID:      int64(workID),
		AnimeID: sql.NullInt64{Int64: int64(animeID), Valid: true},
	})
}

// UpdateUnpublishedAtはworks.unpublished_at状態カラムを書き込み、updated_atを
// 更新する。非nilの時刻を渡すと作品を非公開にし (Unpublishable#unpublish)、nilを渡すと
// クリアして再公開する (Unpublishable#publish)。UpdateAnimeIDと異なりupdated_atを更新
// するのは、公開状態の変更が記帳ではなく実質的な内容変更であるため。
func (r *WorkRepository) UpdateUnpublishedAt(ctx context.Context, id model.WorkID, unpublishedAt *time.Time) error {
	var nullTime sql.NullTime
	if unpublishedAt != nil {
		nullTime = sql.NullTime{Time: *unpublishedAt, Valid: true}
	}
	return r.queries.UpdateWorkUnpublishedAt(ctx, query.UpdateWorkUnpublishedAtParams{
		ID:            int64(id),
		UnpublishedAt: nullTime,
	})
}

// UpdateDeletedAtはworks.deleted_at状態カラムを書き込み、updated_atを更新する。
// 非nilの時刻を渡すと作品をソフトデリートし (SoftDeletable#destroy)、nilを渡すと復元する。
// UpdateUnpublishedAtと同じくupdated_atを更新するのは、削除状態の変更が記帳ではなく実質的な
// 内容変更であるため。
func (r *WorkRepository) UpdateDeletedAt(ctx context.Context, id model.WorkID, deletedAt *time.Time) error {
	var nullTime sql.NullTime
	if deletedAt != nil {
		nullTime = sql.NullTime{Time: *deletedAt, Valid: true}
	}
	return r.queries.UpdateWorkDeletedAt(ctx, query.UpdateWorkDeletedAtParams{
		ID:        int64(id),
		DeletedAt: nullTime,
	})
}

// workFromAnimeSyncRowはanime同期のquery行を *model.Workに変換する。
// worksのテキストカラムはNOT NULL DEFAULT ” なのでここでは空文字列のまま保持し、
// animesが「未設定」をNULLで表す都合に合わせて後段 (同期UseCase) でNULLに写像する。
func workFromAnimeSyncRow(row query.ListWorksForAnimeSyncByIDsRow) *model.Work {
	work := &model.Work{
		ID:                    model.WorkID(row.ID),
		Title:                 row.Title,
		TitleRo:               row.TitleRo,
		TitleEn:               row.TitleEn,
		TitleAlter:            row.TitleAlter,
		TitleAlterEn:          row.TitleAlterEn,
		Media:                 row.Media,
		Synopsis:              row.Synopsis,
		SynopsisEn:            row.SynopsisEn,
		SynopsisSource:        row.SynopsisSource,
		SynopsisSourceEn:      row.SynopsisSourceEn,
		NoEpisodes:            row.NoEpisodes,
		StartEpisodeRawNumber: row.StartEpisodeRawNumber,
	}
	if row.TitleKana != "" {
		titleKana := row.TitleKana
		work.TitleKana = &titleKana
	}
	if row.UnpublishedAt.Valid {
		unpublishedAt := row.UnpublishedAt.Time
		work.UnpublishedAt = &unpublishedAt
	}
	if row.DeletedAt.Valid {
		deletedAt := row.DeletedAt.Time
		work.DeletedAt = &deletedAt
	}
	if row.ManualEpisodesCount.Valid {
		manualEpisodesCount := row.ManualEpisodesCount.Int32
		work.ManualEpisodesCount = &manualEpisodesCount
	}
	if row.NumberFormatID.Valid {
		numberFormatID := model.NumberFormatID(row.NumberFormatID.Int64)
		work.NumberFormatID = &numberFormatID
	}
	if row.AnimeID.Valid {
		animeID := model.AnimeID(row.AnimeID.Int64)
		work.AnimeID = &animeID
	}
	return work
}

// workFromSatelliteSyncRowは別表同期のquery行を *model.Workに変換する。別表が
// sourceとするカラム (とid / anime_id) だけを射影し、残りの *model.Workフィールドは
// ゼロ値のまま。NOT NULL DEFAULT ” のurl列はここでは空文字列のまま保持し、後段
// (リコンサイルヘルパー) で「行なし」に写像する。workFromAnimeSyncRowが空→NULLの写像を
// 同期UseCaseに委ねるのと同じ扱い。
func workFromSatelliteSyncRow(row query.ListWorksForSatelliteSyncByIDsRow) *model.Work {
	work := &model.Work{
		ID:                model.WorkID(row.ID),
		OfficialSiteURL:   row.OfficialSiteUrl,
		OfficialSiteURLEn: row.OfficialSiteUrlEn,
		WikipediaURL:      row.WikipediaUrl,
		WikipediaURLEn:    row.WikipediaUrlEn,
	}
	if row.AnimeID.Valid {
		animeID := model.AnimeID(row.AnimeID.Int64)
		work.AnimeID = &animeID
	}
	if row.ScTid.Valid {
		scTid := row.ScTid.Int32
		work.ScTid = &scTid
	}
	if row.MalAnimeID.Valid {
		malAnimeID := row.MalAnimeID.Int32
		work.MalAnimeID = &malAnimeID
	}
	if row.TwitterUsername.Valid {
		twitterUsername := row.TwitterUsername.String
		work.TwitterUsername = &twitterUsername
	}
	if row.TwitterHashtag.Valid {
		twitterHashtag := row.TwitterHashtag.String
		work.TwitterHashtag = &twitterHashtag
	}
	if row.SeasonYear.Valid {
		seasonYear := row.SeasonYear.Int32
		work.SeasonYear = &seasonYear
	}
	if row.SeasonName.Valid {
		seasonName := row.SeasonName.Int32
		work.SeasonName = &seasonName
	}
	if row.StartedOn.Valid {
		startedOn := row.StartedOn.Time
		work.StartedOn = &startedOn
	}
	if row.EndedOn.Valid {
		endedOn := row.EndedOn.Time
		work.EndedOn = &endedOn
	}
	return work
}

func nullInt32FromPtr(v *int32) sql.NullInt32 {
	if v == nil {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: *v, Valid: true}
}

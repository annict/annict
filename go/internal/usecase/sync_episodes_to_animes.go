package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// SyncEpisodesToAnimesUsecaseは指定されたepisodesの集合を、animes (第1層) と
// anime_classifications (第2層) の参照モデルへリコンサイルする。episodesに対する
// フェーズ2同期の中核で、各episodeをanime + kind='episode' の分類に写像し、未マッピングの
// episodeにはanimeを新規作成して (episodes.anime_idを書き戻し)、マッピング済みのうち
// 写像内容が実際に異なる行だけを更新する。
//
// episodeの分類は親作品のanime_idを必要とする (CHECK制約によりepisodeでは
// parent_anime_idがNOT NULL)。ローダーがepisodes.work_idのJOINで解決し、親作品が
// 未同期のepisodeは失敗させずに後続の実行へ繰り延べる。これによりバッチジョブ (タスク2-4)
// はworks同期をepisodes同期より先に流すだけでよい。
//
// 処理対象のepisode IDはページ単位で呼び出し側 (フェーズ2のバッチジョブ、タスク2-4) が
// 渡す。本UseCase自体はepisodesテーブル全体をスキャンしない。
type SyncEpisodesToAnimesUsecase struct {
	db                      *sql.DB
	episodeRepo             *repository.EpisodeRepository
	animeRepo               *repository.AnimeRepository
	animeClassificationRepo *repository.AnimeClassificationRepository
}

// NewSyncEpisodesToAnimesUsecaseはSyncEpisodesToAnimesUsecaseを生成する。
func NewSyncEpisodesToAnimesUsecase(
	db *sql.DB,
	episodeRepo *repository.EpisodeRepository,
	animeRepo *repository.AnimeRepository,
	animeClassificationRepo *repository.AnimeClassificationRepository,
) *SyncEpisodesToAnimesUsecase {
	return &SyncEpisodesToAnimesUsecase{
		db:                      db,
		episodeRepo:             episodeRepo,
		animeRepo:               animeRepo,
		animeClassificationRepo: animeClassificationRepo,
	}
}

// SyncEpisodesToAnimesInputは今回のリコンサイル対象のepisode IDを保持する
// (通常はバッチジョブが駆動する全件スキャンの1ページ分)。
type SyncEpisodesToAnimesInput struct {
	EpisodeIDs []model.EpisodeID
}

// SyncEpisodesToAnimesResultはリコンサイル結果の件数を報告する。Created /
// Updatedの合計が、正本切り替え判定が依拠する差分検出メトリクスになる。
// SkippedNoParentは親作品が未同期のため繰り延べたepisodeを数える。非ゼロは通常、
// works同期が親に追いついていないだけで、それらのepisodeは後続の実行で取り込まれる。
type SyncEpisodesToAnimesResult struct {
	Processed       int
	Created         int
	Updated         int
	Unchanged       int
	SkippedNoParent int
}

// Executeは入力episodesをanimes / anime_classificationsへリコンサイルする。
//
// 書き込みUseCaseのルールに従い、すべての取得 (解決済みの親anime_idを伴うepisodesと
// 既存のanimes / 分類) はトランザクション前に行い、applyPlanのトランザクション内は
// 永続化のみを行う。
func (uc *SyncEpisodesToAnimesUsecase) Execute(ctx context.Context, input SyncEpisodesToAnimesInput) (*SyncEpisodesToAnimesResult, error) {
	episodes, err := uc.episodeRepo.ListForAnimeSyncByIDs(ctx, input.EpisodeIDs)
	if err != nil {
		return nil, fmt.Errorf("同期対象episodesの取得に失敗: %w", err)
	}
	if len(episodes) == 0 {
		return &SyncEpisodesToAnimesResult{}, nil
	}

	// 既にマッピング済みのepisodesについて既存のanimes / 分類を1クエリずつで
	// 一括取得し、差分検出時の行単位N回ルックアップを避ける。
	mappedAnimeIDs := collectMappedAnimeIDsFromEpisodes(episodes)

	existingAnimes, err := uc.animeRepo.ListByIDs(ctx, mappedAnimeIDs)
	if err != nil {
		return nil, fmt.Errorf("既存animesの取得に失敗: %w", err)
	}
	existingClassifications, err := uc.animeClassificationRepo.ListByAnimeIDs(ctx, mappedAnimeIDs)
	if err != nil {
		return nil, fmt.Errorf("既存anime_classificationsの取得に失敗: %w", err)
	}

	plan := planEpisodeAnimeSync(episodes, indexAnimesByID(existingAnimes), indexClassificationsByAnimeID(existingClassifications))

	return uc.applyPlan(ctx, plan)
}

// collectMappedAnimeIDsFromEpisodesは既にマッピング済みのepisodesのanime IDを返す。
func collectMappedAnimeIDsFromEpisodes(episodes []*model.Episode) []model.AnimeID {
	ids := make([]model.AnimeID, 0, len(episodes))
	for _, e := range episodes {
		if e.AnimeID != nil {
			ids = append(ids, *e.AnimeID)
		}
	}
	return ids
}

// episodeAnimeCreateは未マッピングepisodeに対する作成の単位。animeを挿入し、
// そのkind='episode' 分類を挿入し (anime_idは適用時に挿入されたanimeのIDで埋める)、
// episodes.anime_idを書き戻す。
type episodeAnimeCreate struct {
	episodeID      model.EpisodeID
	anime          repository.CreateAnimeParams
	classification repository.CreateAnimeClassificationParams
}

// episodeAnimeUpdateは内容がずれたマッピング済みepisodeに対する更新の単位。
// 既に同期済みの部分のポインタはnil。classificationCreateはanimeは存在するが
// 分類行が欠落している場合だけセットされる。
type episodeAnimeUpdate struct {
	animeID              model.AnimeID
	anime                *repository.UpdateAnimeParams
	classificationUpdate *repository.UpdateAnimeClassificationParams
	classificationCreate *repository.CreateAnimeClassificationParams
}

type episodeAnimeSyncPlan struct {
	creates         []episodeAnimeCreate
	updates         []episodeAnimeUpdate
	unchanged       int
	skippedNoParent int
	processed       int
}

// planEpisodeAnimeSyncはI/Oを行わずに、episodeごとに作成 / 更新 / 据え置き /
// 繰り延べを判断する。親作品が未同期 (解決したParentAnimeIDがnil) のepisodeは、
// episode分類がNOT NULLのparent_anime_idを要するため繰り延べる。episodeが未マッピングの
// とき、またはマッピング済みでもanime行が欠落しているとき (不整合なマッピングの自己修復) に
// 作成する。
func planEpisodeAnimeSync(
	episodes []*model.Episode,
	animeByID map[model.AnimeID]*model.Anime,
	classificationByAnimeID map[model.AnimeID]*model.AnimeClassification,
) episodeAnimeSyncPlan {
	plan := episodeAnimeSyncPlan{processed: len(episodes)}

	for _, e := range episodes {
		if e.ParentAnimeID == nil {
			plan.skippedNoParent++
			continue
		}

		var existingAnime *model.Anime
		if e.AnimeID != nil {
			existingAnime = animeByID[*e.AnimeID]
		}

		if existingAnime == nil {
			plan.creates = append(plan.creates, episodeAnimeCreate{
				episodeID:      e.ID,
				anime:          animeCreateParamsFromEpisode(e),
				classification: classificationCreateParamsFromEpisode(e, 0),
			})
			continue
		}

		animeID := existingAnime.ID

		var animeUpdate *repository.UpdateAnimeParams
		desiredAnime := animeUpdateParamsFromEpisode(e, existingAnime)
		if animeChanged(existingAnime, desiredAnime) {
			animeUpdate = &desiredAnime
		}

		var classificationUpdate *repository.UpdateAnimeClassificationParams
		var classificationCreate *repository.CreateAnimeClassificationParams
		if existingClassification := classificationByAnimeID[animeID]; existingClassification == nil {
			create := classificationCreateParamsFromEpisode(e, animeID)
			classificationCreate = &create
		} else {
			desiredClassification := classificationUpdateParamsFromEpisode(e, animeID)
			if episodeClassificationChanged(existingClassification, desiredClassification) {
				classificationUpdate = &desiredClassification
			}
		}

		if animeUpdate == nil && classificationUpdate == nil && classificationCreate == nil {
			plan.unchanged++
			continue
		}

		plan.updates = append(plan.updates, episodeAnimeUpdate{
			animeID:              animeID,
			anime:                animeUpdate,
			classificationUpdate: classificationUpdate,
			classificationCreate: classificationCreate,
		})
	}

	return plan
}

// applyPlanは計画を1トランザクションで永続化する。各作成はanimeを挿入し、
// 続けてその分類を挿入されたanimeのIDで挿入し、episodes.anime_idを書き戻すため、
// episodeが中途半端なanimeにマッピングされたまま残ることはない。
func (uc *SyncEpisodesToAnimesUsecase) applyPlan(ctx context.Context, plan episodeAnimeSyncPlan) (*SyncEpisodesToAnimesResult, error) {
	result := &SyncEpisodesToAnimesResult{
		Processed:       plan.processed,
		Unchanged:       plan.unchanged,
		SkippedNoParent: plan.skippedNoParent,
	}
	if len(plan.creates) == 0 && len(plan.updates) == 0 {
		return result, nil
	}

	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("トランザクションの開始に失敗: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	animeRepo := uc.animeRepo.WithTx(tx)
	classificationRepo := uc.animeClassificationRepo.WithTx(tx)
	episodeRepo := uc.episodeRepo.WithTx(tx)

	for _, c := range plan.creates {
		anime, err := animeRepo.Create(ctx, c.anime)
		if err != nil {
			return nil, fmt.Errorf("animeの作成に失敗 (episode_id=%d): %w", c.episodeID, err)
		}

		classificationParams := c.classification
		classificationParams.AnimeID = anime.ID
		if _, err := classificationRepo.Create(ctx, classificationParams); err != nil {
			return nil, fmt.Errorf("anime_classificationの作成に失敗 (episode_id=%d): %w", c.episodeID, err)
		}

		if err := episodeRepo.UpdateAnimeID(ctx, c.episodeID, anime.ID); err != nil {
			return nil, fmt.Errorf("episodes.anime_idの書き戻しに失敗 (episode_id=%d): %w", c.episodeID, err)
		}
		result.Created++
	}

	for _, u := range plan.updates {
		if u.anime != nil {
			if err := animeRepo.Update(ctx, *u.anime); err != nil {
				return nil, fmt.Errorf("animeの更新に失敗 (anime_id=%d): %w", u.animeID, err)
			}
		}
		if u.classificationCreate != nil {
			if _, err := classificationRepo.Create(ctx, *u.classificationCreate); err != nil {
				return nil, fmt.Errorf("anime_classificationの作成に失敗 (anime_id=%d): %w", u.animeID, err)
			}
		}
		if u.classificationUpdate != nil {
			if err := classificationRepo.UpdateByAnimeID(ctx, *u.classificationUpdate); err != nil {
				return nil, fmt.Errorf("anime_classificationの更新に失敗 (anime_id=%d): %w", u.animeID, err)
			}
		}
		result.Updated++
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("トランザクションのコミットに失敗: %w", err)
	}

	return result, nil
}

// animeCreateParamsFromEpisodeはepisodeを新規挿入用の第1層anime属性に写像する。
// episodeが源泉とするのはtitle / title_ro / title_en / statusのみで、statusはepisodeの
// unpublished_at / deleted_atタイムスタンプから導出する。他のカラム (title_kana、
// title_alter系、media、release_status、synopsis系、archive_message) はゼロ値 (NULL) の
// まま残す。archive_messageはanimes専用でepisodeからは源泉としない。
func animeCreateParamsFromEpisode(e *model.Episode) repository.CreateAnimeParams {
	return repository.CreateAnimeParams{
		Title:          nullStringFromStringPtr(e.Title),
		TitleRo:        nullStringFromNonEmpty(e.TitleRo),
		TitleEn:        nullStringFromNonEmpty(e.TitleEn),
		Status:         animeStatusFromEpisodeStatus(e.DerivedStatus()),
		ArchiveMessage: sql.NullString{},
	}
}

// animeUpdateParamsFromEpisodeはepisodeを更新用の第1層anime属性に写像する。
// statusはepisodeのunpublished_at / deleted_atタイムスタンプから導出する。episodeが
// 源泉としないカラム (archive_messageなど) は既存行から引き継ぎ、同期が他所 (編集者や別
// ローダー) で設定された値を上書きしないようにする。archive_messageはanimes専用。
func animeUpdateParamsFromEpisode(e *model.Episode, existing *model.Anime) repository.UpdateAnimeParams {
	return repository.UpdateAnimeParams{
		ID:               existing.ID,
		Title:            nullStringFromStringPtr(e.Title),
		TitleKana:        existing.TitleKana,
		TitleRo:          nullStringFromNonEmpty(e.TitleRo),
		TitleEn:          nullStringFromNonEmpty(e.TitleEn),
		TitleAlter:       existing.TitleAlter,
		TitleAlterRo:     existing.TitleAlterRo,
		TitleAlterEn:     existing.TitleAlterEn,
		TitleAlterOther:  existing.TitleAlterOther,
		Media:            existing.Media,
		ReleaseStatus:    existing.ReleaseStatus,
		Synopsis:         existing.Synopsis,
		SynopsisEn:       existing.SynopsisEn,
		SynopsisSource:   existing.SynopsisSource,
		SynopsisSourceEn: existing.SynopsisSourceEn,
		Status:           animeStatusFromEpisodeStatus(e.DerivedStatus()),
		ArchiveMessage:   existing.ArchiveMessage,
	}
}

// classificationCreateParamsFromEpisodeはepisodeをそのkind='episode' 分類に
// 写像する。parent_anime_id (親作品のanime) とsort_numberはepisodeではNOT NULLが要求され、
// work専用の生成設定 (number_format_id / episode_start_number / expected_episodes_count) と
// standaloneはテーブルのCHECK制約を満たすためゼロ値のまま残す。
func classificationCreateParamsFromEpisode(e *model.Episode, animeID model.AnimeID) repository.CreateAnimeClassificationParams {
	return repository.CreateAnimeClassificationParams{
		AnimeID:       animeID,
		Kind:          model.AnimeClassificationKindEpisode,
		ParentAnimeID: e.ParentAnimeID,
		Number:        numericStringFromFloatPtr(e.RawNumber),
		NumberText:    nullStringFromStringPtr(e.Number),
		SortNumber:    sql.NullInt32{Int32: e.SortNumber, Valid: true},
		Standalone:    false,
	}
}

func classificationUpdateParamsFromEpisode(e *model.Episode, animeID model.AnimeID) repository.UpdateAnimeClassificationParams {
	return repository.UpdateAnimeClassificationParams{
		AnimeID:       animeID,
		Kind:          model.AnimeClassificationKindEpisode,
		ParentAnimeID: e.ParentAnimeID,
		Number:        numericStringFromFloatPtr(e.RawNumber),
		NumberText:    nullStringFromStringPtr(e.Number),
		SortNumber:    sql.NullInt32{Int32: e.SortNumber, Valid: true},
		Standalone:    false,
	}
}

// episodeClassificationChangedは既存分類がepisode由来の目標状態とepisode関連
// フィールドで異なるかを返す。
func episodeClassificationChanged(existing *model.AnimeClassification, desired repository.UpdateAnimeClassificationParams) bool {
	equal := existing.Kind == desired.Kind &&
		animeIDPtrEqual(existing.ParentAnimeID, desired.ParentAnimeID) &&
		nullStringEqual(existing.Number, desired.Number) &&
		nullStringEqual(existing.NumberText, desired.NumberText) &&
		nullInt32Equal(existing.SortNumber, desired.SortNumber) &&
		existing.Standalone == desired.Standalone
	return !equal
}

// animeStatusFromEpisodeStatusはepisodeの導出ライフサイクル状態をanimeのstatus
// enumに写像する。animeStatusFromWorkStatusと同じ純粋なenumアダプタで、timestampsから
// statusへの優先順位はmodel.Episode.DerivedStatusに集約されているため、animesベースで
// status = archived / deletedを書いても次回のリコンシリエーションでpublishedに戻され
// ない。mergedはanime専用でepisodeからは生成されない。
func animeStatusFromEpisodeStatus(s model.EpisodeStatus) model.AnimeStatus {
	switch s {
	case model.EpisodeStatusDeleted:
		return model.AnimeStatusDeleted
	case model.EpisodeStatusArchived:
		return model.AnimeStatusArchived
	default:
		return model.AnimeStatusPublished
	}
}

// numericStringFromFloatPtrは任意のfloatをNUMERICカラム用の正準テキストに描画する。
// 正確にラウンドトリップする最短表現を使う。nilポインタはNULLに写像する
// (episodes.raw_numberはNULL許容)。
func numericStringFromFloatPtr(p *float64) sql.NullString {
	if p == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: strconv.FormatFloat(*p, 'f', -1, 64), Valid: true}
}

// animeIDPtrEqualは2つの任意のアニメID外部キーが等しいかを返す。両方nilは
// 等しいとみなす。
func animeIDPtrEqual(a, b *model.AnimeID) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

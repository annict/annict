package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// RailsのWork#media enum値。animesスキーマは `web` を `ona` に改名する
// (ADR 0008)。他の値は意味を保つ。0 (`other`) はキャッチオール。
const (
	workMediaOther = 0
	workMediaTV    = 1
	workMediaOVA   = 2
	workMediaMovie = 3
	workMediaONA   = 4
)

// SyncWorksToAnimesUsecaseは指定されたworksの集合を、animes (第1層) と
// anime_classifications (第2層) の参照モデルへリコンサイルする。worksに対する
// フェーズ2同期の中核で、各workをanime + kind='work' の分類に写像し、未マッピングの
// workにはanimeを新規作成して (works.anime_idを書き戻し)、マッピング済みのうち
// 写像内容が実際に異なる行だけを更新する。
//
// 処理対象のwork IDはページ単位で呼び出し側 (フェーズ2のバッチジョブ、タスク2-4) が
// 渡す。本UseCase自体はworksテーブル全体をスキャンしない。
type SyncWorksToAnimesUsecase struct {
	db                      *sql.DB
	workRepo                *repository.WorkRepository
	animeRepo               *repository.AnimeRepository
	animeClassificationRepo *repository.AnimeClassificationRepository
}

// NewSyncWorksToAnimesUsecaseはSyncWorksToAnimesUsecaseを生成する。
func NewSyncWorksToAnimesUsecase(
	db *sql.DB,
	workRepo *repository.WorkRepository,
	animeRepo *repository.AnimeRepository,
	animeClassificationRepo *repository.AnimeClassificationRepository,
) *SyncWorksToAnimesUsecase {
	return &SyncWorksToAnimesUsecase{
		db:                      db,
		workRepo:                workRepo,
		animeRepo:               animeRepo,
		animeClassificationRepo: animeClassificationRepo,
	}
}

// SyncWorksToAnimesInputは今回のリコンサイル対象のwork IDを保持する
// (通常はバッチジョブが駆動する全件スキャンの1ページ分)。
type SyncWorksToAnimesInput struct {
	WorkIDs []model.WorkID
}

// SyncWorksToAnimesResultはリコンサイル結果の件数を報告する。Created /
// Updatedの合計が、正本切り替え判定が依拠する差分検出メトリクスになる
// (両方が0の実行は、処理したworksが既にanimesと一致していることを意味する)。
type SyncWorksToAnimesResult struct {
	Processed int
	Created   int
	Updated   int
	Unchanged int
}

// Executeは入力worksをanimes / anime_classificationsへリコンサイルする。
//
// 書き込みUseCaseのルールに従い、すべての取得 (worksと既存のanimes / 分類) は
// トランザクション前に行い、applyPlanのトランザクション内は永続化のみを行う。
func (uc *SyncWorksToAnimesUsecase) Execute(ctx context.Context, input SyncWorksToAnimesInput) (*SyncWorksToAnimesResult, error) {
	works, err := uc.workRepo.ListForAnimeSyncByIDs(ctx, input.WorkIDs)
	if err != nil {
		return nil, fmt.Errorf("同期対象worksの取得に失敗: %w", err)
	}
	if len(works) == 0 {
		return &SyncWorksToAnimesResult{}, nil
	}

	// 既にマッピング済みのworksについて既存のanimes / 分類を1クエリずつで
	// 一括取得し、差分検出時の行単位N回ルックアップを避ける。
	mappedAnimeIDs := collectMappedAnimeIDs(works)

	existingAnimes, err := uc.animeRepo.ListByIDs(ctx, mappedAnimeIDs)
	if err != nil {
		return nil, fmt.Errorf("既存animesの取得に失敗: %w", err)
	}
	existingClassifications, err := uc.animeClassificationRepo.ListByAnimeIDs(ctx, mappedAnimeIDs)
	if err != nil {
		return nil, fmt.Errorf("既存anime_classificationsの取得に失敗: %w", err)
	}

	plan := planWorkAnimeSync(works, indexAnimesByID(existingAnimes), indexClassificationsByAnimeID(existingClassifications))

	return uc.applyPlan(ctx, plan)
}

// collectMappedAnimeIDsは既にマッピング済みのworksのanime IDを返す。
func collectMappedAnimeIDs(works []*model.Work) []model.AnimeID {
	ids := make([]model.AnimeID, 0, len(works))
	for _, w := range works {
		if w.AnimeID != nil {
			ids = append(ids, *w.AnimeID)
		}
	}
	return ids
}

func indexAnimesByID(animes []*model.Anime) map[model.AnimeID]*model.Anime {
	byID := make(map[model.AnimeID]*model.Anime, len(animes))
	for _, a := range animes {
		byID[a.ID] = a
	}
	return byID
}

func indexClassificationsByAnimeID(classifications []*model.AnimeClassification) map[model.AnimeID]*model.AnimeClassification {
	byAnimeID := make(map[model.AnimeID]*model.AnimeClassification, len(classifications))
	for _, c := range classifications {
		byAnimeID[c.AnimeID] = c
	}
	return byAnimeID
}

// workAnimeCreateは未マッピングworkに対する作成の単位。animeを挿入し、
// そのkind='work' 分類を挿入し (anime_idは適用時に挿入されたanimeのIDで埋める)、
// works.anime_idを書き戻す。
type workAnimeCreate struct {
	workID         model.WorkID
	anime          repository.CreateAnimeParams
	classification repository.CreateAnimeClassificationParams
}

// workAnimeUpdateは内容がずれたマッピング済みworkに対する更新の単位。
// 既に同期済みの部分のポインタはnil。classificationCreateはanimeは存在するが
// 分類行が欠落している場合だけセットされる。
type workAnimeUpdate struct {
	animeID              model.AnimeID
	anime                *repository.UpdateAnimeParams
	classificationUpdate *repository.UpdateAnimeClassificationParams
	classificationCreate *repository.CreateAnimeClassificationParams
}

type workAnimeSyncPlan struct {
	creates   []workAnimeCreate
	updates   []workAnimeUpdate
	unchanged int
	processed int
}

// planWorkAnimeSyncはI/Oを行わずに、workごとに作成 / 更新 / 据え置きを判断する。
// workが未マッピングのとき、またはマッピング済みでもanime行が欠落しているとき
// (不整合なマッピングの自己修復) に作成する。
func planWorkAnimeSync(
	works []*model.Work,
	animeByID map[model.AnimeID]*model.Anime,
	classificationByAnimeID map[model.AnimeID]*model.AnimeClassification,
) workAnimeSyncPlan {
	plan := workAnimeSyncPlan{processed: len(works)}

	for _, w := range works {
		var existingAnime *model.Anime
		if w.AnimeID != nil {
			existingAnime = animeByID[*w.AnimeID]
		}

		if existingAnime == nil {
			plan.creates = append(plan.creates, workAnimeCreate{
				workID:         w.ID,
				anime:          animeCreateParamsFromWork(w),
				classification: classificationCreateParamsFromWork(w, 0),
			})
			continue
		}

		animeID := existingAnime.ID

		var animeUpdate *repository.UpdateAnimeParams
		desiredAnime := animeUpdateParamsFromWork(w, existingAnime)
		if animeChanged(existingAnime, desiredAnime) {
			animeUpdate = &desiredAnime
		}

		var classificationUpdate *repository.UpdateAnimeClassificationParams
		var classificationCreate *repository.CreateAnimeClassificationParams
		if existingClassification := classificationByAnimeID[animeID]; existingClassification == nil {
			create := classificationCreateParamsFromWork(w, animeID)
			classificationCreate = &create
		} else {
			desiredClassification := classificationUpdateParamsFromWork(w, animeID)
			if classificationChanged(existingClassification, desiredClassification) {
				classificationUpdate = &desiredClassification
			}
		}

		if animeUpdate == nil && classificationUpdate == nil && classificationCreate == nil {
			plan.unchanged++
			continue
		}

		plan.updates = append(plan.updates, workAnimeUpdate{
			animeID:              animeID,
			anime:                animeUpdate,
			classificationUpdate: classificationUpdate,
			classificationCreate: classificationCreate,
		})
	}

	return plan
}

// applyPlanは計画を1トランザクションで永続化する。各作成はanimeを挿入し、
// 続けてその分類を挿入されたanimeのIDで挿入し、works.anime_idを書き戻すため、
// workが中途半端なanimeにマッピングされたまま残ることはない。
func (uc *SyncWorksToAnimesUsecase) applyPlan(ctx context.Context, plan workAnimeSyncPlan) (*SyncWorksToAnimesResult, error) {
	result := &SyncWorksToAnimesResult{Processed: plan.processed, Unchanged: plan.unchanged}
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
	workRepo := uc.workRepo.WithTx(tx)

	for _, c := range plan.creates {
		anime, err := animeRepo.Create(ctx, c.anime)
		if err != nil {
			return nil, fmt.Errorf("animeの作成に失敗 (work_id=%d): %w", c.workID, err)
		}

		classificationParams := c.classification
		classificationParams.AnimeID = anime.ID
		if _, err := classificationRepo.Create(ctx, classificationParams); err != nil {
			return nil, fmt.Errorf("anime_classificationの作成に失敗 (work_id=%d): %w", c.workID, err)
		}

		if err := workRepo.UpdateAnimeID(ctx, c.workID, anime.ID); err != nil {
			return nil, fmt.Errorf("works.anime_idの書き戻しに失敗 (work_id=%d): %w", c.workID, err)
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

// animeCreateParamsFromWorkはworkを新規挿入用の第1層anime属性に写像する。
// statusはworkのunpublished_at / deleted_atタイムスタンプから導出する。animesが
// worksから取り込まないカラム (title_alter_ro / title_alter_other / release_status /
// archive_message) はゼロ値 (NULL) のまま残す。
// archive_messageはanimes専用でここでは源泉としない。
func animeCreateParamsFromWork(w *model.Work) repository.CreateAnimeParams {
	return repository.CreateAnimeParams{
		Title:            nullStringFromNonEmpty(w.Title),
		TitleKana:        nullStringFromStringPtr(w.TitleKana),
		TitleRo:          nullStringFromNonEmpty(w.TitleRo),
		TitleEn:          nullStringFromNonEmpty(w.TitleEn),
		TitleAlter:       nullStringFromNonEmpty(w.TitleAlter),
		TitleAlterEn:     nullStringFromNonEmpty(w.TitleAlterEn),
		Media:            mediaToAnimeMedia(w.Media),
		Synopsis:         nullStringFromNonEmpty(w.Synopsis),
		SynopsisEn:       nullStringFromNonEmpty(w.SynopsisEn),
		SynopsisSource:   nullStringFromNonEmpty(w.SynopsisSource),
		SynopsisSourceEn: nullStringFromNonEmpty(w.SynopsisSourceEn),
		Status:           animeStatusFromWorkStatus(w.DerivedStatus()),
		ArchiveMessage:   sql.NullString{},
	}
}

// animeUpdateParamsFromWorkはworkを更新用の第1層anime属性に写像する。
// statusはworkのunpublished_at / deleted_atタイムスタンプから導出する。animesが
// worksから取り込まないカラム (release_status / archive_messageなど) は既存行から
// 引き継ぎ、同期が編集者の設定値を上書きしないようにする。archive_messageはanimes専用。
func animeUpdateParamsFromWork(w *model.Work, existing *model.Anime) repository.UpdateAnimeParams {
	return repository.UpdateAnimeParams{
		ID:               existing.ID,
		Title:            nullStringFromNonEmpty(w.Title),
		TitleKana:        nullStringFromStringPtr(w.TitleKana),
		TitleRo:          nullStringFromNonEmpty(w.TitleRo),
		TitleEn:          nullStringFromNonEmpty(w.TitleEn),
		TitleAlter:       nullStringFromNonEmpty(w.TitleAlter),
		TitleAlterEn:     nullStringFromNonEmpty(w.TitleAlterEn),
		TitleAlterRo:     existing.TitleAlterRo,
		TitleAlterOther:  existing.TitleAlterOther,
		ReleaseStatus:    existing.ReleaseStatus,
		Media:            mediaToAnimeMedia(w.Media),
		Synopsis:         nullStringFromNonEmpty(w.Synopsis),
		SynopsisEn:       nullStringFromNonEmpty(w.SynopsisEn),
		SynopsisSource:   nullStringFromNonEmpty(w.SynopsisSource),
		SynopsisSourceEn: nullStringFromNonEmpty(w.SynopsisSourceEn),
		Status:           animeStatusFromWorkStatus(w.DerivedStatus()),
		ArchiveMessage:   existing.ArchiveMessage,
	}
}

// animeChangedは既存animeがwork由来の目標状態と異なるかを返す。引き継いだ
// カラムは構成上等しく比較されるため、更新を生むのはworks由来の属性だけ。
func animeChanged(existing *model.Anime, desired repository.UpdateAnimeParams) bool {
	equal := nullStringEqual(existing.Title, desired.Title) &&
		nullStringEqual(existing.TitleKana, desired.TitleKana) &&
		nullStringEqual(existing.TitleRo, desired.TitleRo) &&
		nullStringEqual(existing.TitleEn, desired.TitleEn) &&
		nullStringEqual(existing.TitleAlter, desired.TitleAlter) &&
		nullStringEqual(existing.TitleAlterRo, desired.TitleAlterRo) &&
		nullStringEqual(existing.TitleAlterEn, desired.TitleAlterEn) &&
		nullStringEqual(existing.TitleAlterOther, desired.TitleAlterOther) &&
		existing.Media == desired.Media &&
		existing.ReleaseStatus == desired.ReleaseStatus &&
		nullStringEqual(existing.Synopsis, desired.Synopsis) &&
		nullStringEqual(existing.SynopsisEn, desired.SynopsisEn) &&
		nullStringEqual(existing.SynopsisSource, desired.SynopsisSource) &&
		nullStringEqual(existing.SynopsisSourceEn, desired.SynopsisSourceEn) &&
		existing.Status == desired.Status &&
		nullStringEqual(existing.ArchiveMessage, desired.ArchiveMessage)
	return !equal
}

// classificationCreateParamsFromWorkはworkをそのkind='work' 分類に写像する。
// episode専用フィールドはテーブルのCHECK制約を満たすためNULLのまま、no_episodesは
// 同じ極性でstandaloneに引き継ぐ。
func classificationCreateParamsFromWork(w *model.Work, animeID model.AnimeID) repository.CreateAnimeClassificationParams {
	return repository.CreateAnimeClassificationParams{
		AnimeID:               animeID,
		Kind:                  model.AnimeClassificationKindWork,
		Standalone:            w.NoEpisodes,
		NumberFormatID:        w.NumberFormatID,
		EpisodeStartNumber:    numericStringFromFloat(w.StartEpisodeRawNumber),
		ExpectedEpisodesCount: nullInt32FromInt32Ptr(w.ManualEpisodesCount),
	}
}

func classificationUpdateParamsFromWork(w *model.Work, animeID model.AnimeID) repository.UpdateAnimeClassificationParams {
	return repository.UpdateAnimeClassificationParams{
		AnimeID:               animeID,
		Kind:                  model.AnimeClassificationKindWork,
		Standalone:            w.NoEpisodes,
		NumberFormatID:        w.NumberFormatID,
		EpisodeStartNumber:    numericStringFromFloat(w.StartEpisodeRawNumber),
		ExpectedEpisodesCount: nullInt32FromInt32Ptr(w.ManualEpisodesCount),
	}
}

// classificationChangedは既存分類がwork由来の目標状態とwork関連フィールドで
// 異なるかを返す。
func classificationChanged(existing *model.AnimeClassification, desired repository.UpdateAnimeClassificationParams) bool {
	equal := existing.Kind == desired.Kind &&
		existing.Standalone == desired.Standalone &&
		numberFormatIDEqual(existing.NumberFormatID, desired.NumberFormatID) &&
		nullStringEqual(existing.EpisodeStartNumber, desired.EpisodeStartNumber) &&
		nullInt32Equal(existing.ExpectedEpisodesCount, desired.ExpectedEpisodesCount)
	return !equal
}

// mediaToAnimeMediaはRailsのWork#media整数をanime_media enumに写像する。
// works.mediaはNOT NULLかつRails enumに制約されるため、キャッチオール
// (0 = other) は想定外の値も吸収する。
func mediaToAnimeMedia(media int32) model.AnimeMedia {
	switch media {
	case workMediaTV:
		return model.AnimeMediaTV
	case workMediaOVA:
		return model.AnimeMediaOVA
	case workMediaMovie:
		return model.AnimeMediaMovie
	case workMediaONA:
		return model.AnimeMediaONA
	case workMediaOther:
		return model.AnimeMediaOther
	default:
		return model.AnimeMediaOther
	}
}

// animeStatusFromWorkStatusはworkの導出ライフサイクル状態をanimeのstatus
// enumに写像する。mediaToAnimeMediaと同じ純粋なenumアダプタで、timestampsから
// statusへの優先順位はmodel.Work.DerivedStatusに集約されているため、animesベースで
// status = archived / deletedを書いても次回のリコンシリエーションでpublishedに戻され
// ない。mergedはanime専用でworkからは生成されない。
func animeStatusFromWorkStatus(s model.WorkStatus) model.AnimeStatus {
	switch s {
	case model.WorkStatusDeleted:
		return model.AnimeStatusDeleted
	case model.WorkStatusArchived:
		return model.AnimeStatusArchived
	default:
		return model.AnimeStatusPublished
	}
}

// numericStringFromFloatはfloatをNUMERICカラム用の正準テキストに描画する。
// 正確にラウンドトリップする最短表現を使う。works.start_episode_raw_numberは
// NOT NULLなので結果は常にvalid。
func numericStringFromFloat(f float64) sql.NullString {
	return sql.NullString{String: strconv.FormatFloat(f, 'f', -1, 64), Valid: true}
}

// nullStringFromNonEmptyは空文字列のときNULLに写像し、worksの
// NOT NULL DEFAULT ” カラムに対してanimesが「未設定」をNULLで表す扱いに合わせる。
func nullStringFromNonEmpty(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func nullStringFromStringPtr(p *string) sql.NullString {
	if p == nil || *p == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: *p, Valid: true}
}

func nullInt32FromInt32Ptr(p *int32) sql.NullInt32 {
	if p == nil {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: *p, Valid: true}
}

func nullStringEqual(a, b sql.NullString) bool {
	return a.Valid == b.Valid && a.String == b.String
}

func nullInt32Equal(a, b sql.NullInt32) bool {
	return a.Valid == b.Valid && a.Int32 == b.Int32
}

func numberFormatIDEqual(a, b *model.NumberFormatID) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

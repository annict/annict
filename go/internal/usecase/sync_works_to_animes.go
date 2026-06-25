package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// Rails Work#media enum values. The animes schema renames `web` to `ona`
// (ADR 0008); the other values keep their meaning. 0 (`other`) is the catch-all.
//
// [Ja] Rails の Work#media enum 値。animes スキーマは `web` を `ona` に改名する
// (ADR 0008)。他の値は意味を保つ。0 (`other`) はキャッチオール。
const (
	workMediaOther = 0
	workMediaTV    = 1
	workMediaOVA   = 2
	workMediaMovie = 3
	workMediaONA   = 4
)

// SyncWorksToAnimesUsecase reconciles a given set of works into the animes (layer 1)
// and anime_classifications (layer 2) reference model. It is the phase 2 sync core
// for works: it maps each work onto an anime + a kind='work' classification, creates
// the anime for unmapped works (writing works.anime_id back), and updates only the
// rows whose mapped content actually differs.
//
// The caller (the phase 2 batch job, task 2-4) supplies the work IDs to process,
// page by page; this usecase does not scan the whole works table itself.
//
// [Ja] SyncWorksToAnimesUsecase は指定された works の集合を、animes (第 1 層) と
// anime_classifications (第 2 層) の参照モデルへリコンサイルする。works に対する
// フェーズ 2 同期の中核で、各 work を anime + kind='work' の分類に写像し、未マッピングの
// work には anime を新規作成して (works.anime_id を書き戻し)、マッピング済みのうち
// 写像内容が実際に異なる行だけを更新する。
//
// 処理対象の work ID はページ単位で呼び出し側 (フェーズ 2 のバッチジョブ、タスク 2-4) が
// 渡す。本 UseCase 自体は works テーブル全体をスキャンしない。
type SyncWorksToAnimesUsecase struct {
	db                      *sql.DB
	workRepo                *repository.WorkRepository
	animeRepo               *repository.AnimeRepository
	animeClassificationRepo *repository.AnimeClassificationRepository
}

// NewSyncWorksToAnimesUsecase constructs a SyncWorksToAnimesUsecase.
//
// [Ja] NewSyncWorksToAnimesUsecase は SyncWorksToAnimesUsecase を生成する。
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

// SyncWorksToAnimesInput carries the work IDs to reconcile in this run (typically one
// page of the full-table scan driven by the batch job).
//
// [Ja] SyncWorksToAnimesInput は今回のリコンサイル対象の work ID を保持する
// (通常はバッチジョブが駆動する全件スキャンの 1 ページ分)。
type SyncWorksToAnimesInput struct {
	WorkIDs []model.WorkID
}

// SyncWorksToAnimesResult reports the reconciliation outcome counts. Created /
// Updated together form the diff-detection metric the cutover decision depends on
// (a run that reports both as zero means the processed works already match animes).
//
// [Ja] SyncWorksToAnimesResult はリコンサイル結果の件数を報告する。Created /
// Updated の合計が、正本切り替え判定が依拠する差分検出メトリクスになる
// (両方が 0 の実行は、処理した works が既に animes と一致していることを意味する)。
type SyncWorksToAnimesResult struct {
	Processed int
	Created   int
	Updated   int
	Unchanged int
}

// Execute reconciles the input works into animes / anime_classifications.
//
// Following the write-usecase rule, every read (the works and the existing animes /
// classifications) happens before the transaction; the transaction in applyPlan
// performs persistence only.
//
// [Ja] Execute は入力 works を animes / anime_classifications へリコンサイルする。
//
// 書き込み UseCase のルールに従い、すべての取得 (works と既存の animes / 分類) は
// トランザクション前に行い、applyPlan のトランザクション内は永続化のみを行う。
func (uc *SyncWorksToAnimesUsecase) Execute(ctx context.Context, input SyncWorksToAnimesInput) (*SyncWorksToAnimesResult, error) {
	works, err := uc.workRepo.ListForAnimeSyncByIDs(ctx, input.WorkIDs)
	if err != nil {
		return nil, fmt.Errorf("同期対象 works の取得に失敗: %w", err)
	}
	if len(works) == 0 {
		return &SyncWorksToAnimesResult{}, nil
	}

	// Batch-fetch the existing animes / classifications for the already-mapped works
	// in one query each, avoiding N per-row lookups during diff detection.
	//
	// [Ja] 既にマッピング済みの works について既存の animes / 分類を 1 クエリずつで
	// 一括取得し、差分検出時の行単位 N 回ルックアップを避ける。
	mappedAnimeIDs := collectMappedAnimeIDs(works)

	existingAnimes, err := uc.animeRepo.ListByIDs(ctx, mappedAnimeIDs)
	if err != nil {
		return nil, fmt.Errorf("既存 animes の取得に失敗: %w", err)
	}
	existingClassifications, err := uc.animeClassificationRepo.ListByAnimeIDs(ctx, mappedAnimeIDs)
	if err != nil {
		return nil, fmt.Errorf("既存 anime_classifications の取得に失敗: %w", err)
	}

	plan := planWorkAnimeSync(works, indexAnimesByID(existingAnimes), indexClassificationsByAnimeID(existingClassifications))

	return uc.applyPlan(ctx, plan)
}

// collectMappedAnimeIDs returns the anime IDs of the works that are already mapped.
//
// [Ja] collectMappedAnimeIDs は既にマッピング済みの works の anime ID を返す。
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

// workAnimeCreate is a unit of work for an unmapped work: insert the anime, insert
// its kind='work' classification (anime_id is filled with the inserted anime's ID at
// apply time), and write works.anime_id back.
//
// [Ja] workAnimeCreate は未マッピング work に対する作成の単位。anime を挿入し、
// その kind='work' 分類を挿入し (anime_id は適用時に挿入された anime の ID で埋める)、
// works.anime_id を書き戻す。
type workAnimeCreate struct {
	workID         model.WorkID
	anime          repository.CreateAnimeParams
	classification repository.CreateAnimeClassificationParams
}

// workAnimeUpdate is a unit of work for a mapped work whose content drifted. Each
// pointer is nil when that part is already in sync; classificationCreate is set only
// when the anime exists but its classification row is missing.
//
// [Ja] workAnimeUpdate は内容がずれたマッピング済み work に対する更新の単位。
// 既に同期済みの部分のポインタは nil。classificationCreate は anime は存在するが
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

// planWorkAnimeSync decides, per work, whether to create / update / leave untouched,
// without performing any I/O. A work is created when it is unmapped, or when it is
// mapped but its anime row is missing (self-healing for an inconsistent mapping).
//
// [Ja] planWorkAnimeSync は I/O を行わずに、work ごとに作成 / 更新 / 据え置きを判断する。
// work が未マッピングのとき、またはマッピング済みでも anime 行が欠落しているとき
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

// applyPlan persists the plan in a single transaction. Each create inserts the anime,
// then its classification with the inserted anime's ID, then writes works.anime_id
// back, so a work is never left mapped to a half-built anime.
//
// [Ja] applyPlan は計画を 1 トランザクションで永続化する。各作成は anime を挿入し、
// 続けてその分類を挿入された anime の ID で挿入し、works.anime_id を書き戻すため、
// work が中途半端な anime にマッピングされたまま残ることはない。
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
			return nil, fmt.Errorf("anime の作成に失敗 (work_id=%d): %w", c.workID, err)
		}

		classificationParams := c.classification
		classificationParams.AnimeID = anime.ID
		if _, err := classificationRepo.Create(ctx, classificationParams); err != nil {
			return nil, fmt.Errorf("anime_classification の作成に失敗 (work_id=%d): %w", c.workID, err)
		}

		if err := workRepo.UpdateAnimeID(ctx, c.workID, anime.ID); err != nil {
			return nil, fmt.Errorf("works.anime_id の書き戻しに失敗 (work_id=%d): %w", c.workID, err)
		}
		result.Created++
	}

	for _, u := range plan.updates {
		if u.anime != nil {
			if err := animeRepo.Update(ctx, *u.anime); err != nil {
				return nil, fmt.Errorf("anime の更新に失敗 (anime_id=%d): %w", u.animeID, err)
			}
		}
		if u.classificationCreate != nil {
			if _, err := classificationRepo.Create(ctx, *u.classificationCreate); err != nil {
				return nil, fmt.Errorf("anime_classification の作成に失敗 (anime_id=%d): %w", u.animeID, err)
			}
		}
		if u.classificationUpdate != nil {
			if err := classificationRepo.UpdateByAnimeID(ctx, *u.classificationUpdate); err != nil {
				return nil, fmt.Errorf("anime_classification の更新に失敗 (anime_id=%d): %w", u.animeID, err)
			}
		}
		result.Updated++
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("トランザクションのコミットに失敗: %w", err)
	}

	return result, nil
}

// animeCreateParamsFromWork maps a work onto the layer-1 anime attributes for a fresh
// insert. The columns animes does not source from works (title_alter_ro /
// title_alter_other / release_status) stay at their zero value (NULL).
//
// [Ja] animeCreateParamsFromWork は work を新規挿入用の第 1 層 anime 属性に写像する。
// animes が works から取り込まないカラム (title_alter_ro / title_alter_other /
// release_status) はゼロ値 (NULL) のまま残す。
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
		Status:           workStatusToAnimeStatus(w.Status),
		ArchiveMessage:   nullStringFromStringPtr(w.ArchiveMessage),
	}
}

// animeUpdateParamsFromWork maps a work onto the layer-1 anime attributes for an
// update. The columns animes does not source from works are carried over from the
// existing row so the sync never clobbers editor-set values (release_status etc.).
//
// [Ja] animeUpdateParamsFromWork は work を更新用の第 1 層 anime 属性に写像する。
// animes が works から取り込まないカラムは既存行から引き継ぎ、同期が編集者の設定値
// (release_status など) を上書きしないようにする。
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
		Status:           workStatusToAnimeStatus(w.Status),
		ArchiveMessage:   nullStringFromStringPtr(w.ArchiveMessage),
	}
}

// animeChanged reports whether the existing anime differs from the work-derived
// desired state. The preserved columns compare equal by construction, so only the
// work-sourced attributes can trigger an update.
//
// [Ja] animeChanged は既存 anime が work 由来の目標状態と異なるかを返す。引き継いだ
// カラムは構成上等しく比較されるため、更新を生むのは works 由来の属性だけ。
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

// classificationCreateParamsFromWork maps a work onto its kind='work' classification.
// The episode-only fields stay NULL to satisfy the table's CHECK constraints, and
// no_episodes carries over to standalone with the same polarity.
//
// [Ja] classificationCreateParamsFromWork は work をその kind='work' 分類に写像する。
// episode 専用フィールドはテーブルの CHECK 制約を満たすため NULL のまま、no_episodes は
// 同じ極性で standalone に引き継ぐ。
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

// classificationChanged reports whether the existing classification differs from the
// work-derived desired state for the work-relevant fields.
//
// [Ja] classificationChanged は既存分類が work 由来の目標状態と work 関連フィールドで
// 異なるかを返す。
func classificationChanged(existing *model.AnimeClassification, desired repository.UpdateAnimeClassificationParams) bool {
	equal := existing.Kind == desired.Kind &&
		existing.Standalone == desired.Standalone &&
		numberFormatIDEqual(existing.NumberFormatID, desired.NumberFormatID) &&
		nullStringEqual(existing.EpisodeStartNumber, desired.EpisodeStartNumber) &&
		nullInt32Equal(existing.ExpectedEpisodesCount, desired.ExpectedEpisodesCount)
	return !equal
}

// mediaToAnimeMedia maps the Rails Work#media integer onto the anime_media enum. The
// catch-all (0 = other) also absorbs any unexpected value, since works.media is
// NOT NULL and constrained to the Rails enum.
//
// [Ja] mediaToAnimeMedia は Rails の Work#media 整数を anime_media enum に写像する。
// works.media は NOT NULL かつ Rails enum に制約されるため、キャッチオール
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

// workStatusToAnimeStatus maps work_status onto anime_status. The three work values
// are a subset of anime_status; merged is anime-only and never produced here.
//
// [Ja] workStatusToAnimeStatus は work_status を anime_status に写像する。work の 3 値は
// anime_status の部分集合で、merged は anime 専用でありここでは生成されない。
func workStatusToAnimeStatus(status model.WorkStatus) model.AnimeStatus {
	switch status {
	case model.WorkStatusPublished:
		return model.AnimeStatusPublished
	case model.WorkStatusArchived:
		return model.AnimeStatusArchived
	case model.WorkStatusDeleted:
		return model.AnimeStatusDeleted
	default:
		return model.AnimeStatusPublished
	}
}

// numericStringFromFloat renders a float as the canonical text for a NUMERIC column,
// using the shortest representation that round-trips exactly. works.start_episode_raw_number
// is NOT NULL, so the result is always valid.
//
// [Ja] numericStringFromFloat は float を NUMERIC カラム用の正準テキストに描画する。
// 正確にラウンドトリップする最短表現を使う。works.start_episode_raw_number は
// NOT NULL なので結果は常に valid。
func numericStringFromFloat(f float64) sql.NullString {
	return sql.NullString{String: strconv.FormatFloat(f, 'f', -1, 64), Valid: true}
}

// nullStringFromNonEmpty maps a value to NULL when it is the empty string, matching
// how animes uses NULL for "absent" against works' NOT NULL DEFAULT ” columns.
//
// [Ja] nullStringFromNonEmpty は空文字列のとき NULL に写像し、works の
// NOT NULL DEFAULT ” カラムに対して animes が「未設定」を NULL で表す扱いに合わせる。
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

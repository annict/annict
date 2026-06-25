package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// SyncEpisodesToAnimesUsecase reconciles a given set of episodes into the animes
// (layer 1) and anime_classifications (layer 2) reference model. It is the phase 2
// sync core for episodes: it maps each episode onto an anime + a kind='episode'
// classification, creates the anime for unmapped episodes (writing episodes.anime_id
// back), and updates only the rows whose mapped content actually differs.
//
// An episode's classification needs the parent work's anime_id (parent_anime_id is
// NOT NULL for an episode per the CHECK constraint). The loader resolves it through
// the episodes.work_id JOIN; an episode whose parent work is not yet synced is
// deferred to a later run rather than failing, so the batch job (task 2-4) only has
// to run the works sync before the episodes sync.
//
// The caller (the phase 2 batch job, task 2-4) supplies the episode IDs to process,
// page by page; this usecase does not scan the whole episodes table itself.
//
// [Ja] SyncEpisodesToAnimesUsecase は指定された episodes の集合を、animes (第 1 層) と
// anime_classifications (第 2 層) の参照モデルへリコンサイルする。episodes に対する
// フェーズ 2 同期の中核で、各 episode を anime + kind='episode' の分類に写像し、未マッピングの
// episode には anime を新規作成して (episodes.anime_id を書き戻し)、マッピング済みのうち
// 写像内容が実際に異なる行だけを更新する。
//
// episode の分類は親作品の anime_id を必要とする (CHECK 制約により episode では
// parent_anime_id が NOT NULL)。ローダーが episodes.work_id の JOIN で解決し、親作品が
// 未同期の episode は失敗させずに後続の実行へ繰り延べる。これによりバッチジョブ (タスク 2-4)
// は works 同期を episodes 同期より先に流すだけでよい。
//
// 処理対象の episode ID はページ単位で呼び出し側 (フェーズ 2 のバッチジョブ、タスク 2-4) が
// 渡す。本 UseCase 自体は episodes テーブル全体をスキャンしない。
type SyncEpisodesToAnimesUsecase struct {
	db                      *sql.DB
	episodeRepo             *repository.EpisodeRepository
	animeRepo               *repository.AnimeRepository
	animeClassificationRepo *repository.AnimeClassificationRepository
}

// NewSyncEpisodesToAnimesUsecase constructs a SyncEpisodesToAnimesUsecase.
//
// [Ja] NewSyncEpisodesToAnimesUsecase は SyncEpisodesToAnimesUsecase を生成する。
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

// SyncEpisodesToAnimesInput carries the episode IDs to reconcile in this run
// (typically one page of the full-table scan driven by the batch job).
//
// [Ja] SyncEpisodesToAnimesInput は今回のリコンサイル対象の episode ID を保持する
// (通常はバッチジョブが駆動する全件スキャンの 1 ページ分)。
type SyncEpisodesToAnimesInput struct {
	EpisodeIDs []model.EpisodeID
}

// SyncEpisodesToAnimesResult reports the reconciliation outcome counts. Created /
// Updated together form the diff-detection metric the cutover decision depends on.
// SkippedNoParent counts episodes deferred because their parent work is not yet
// synced; a non-zero value usually just means the works sync has not caught up to a
// parent, and those episodes are picked up on a later run.
//
// [Ja] SyncEpisodesToAnimesResult はリコンサイル結果の件数を報告する。Created /
// Updated の合計が、正本切り替え判定が依拠する差分検出メトリクスになる。
// SkippedNoParent は親作品が未同期のため繰り延べた episode を数える。非ゼロは通常、
// works 同期が親に追いついていないだけで、それらの episode は後続の実行で取り込まれる。
type SyncEpisodesToAnimesResult struct {
	Processed       int
	Created         int
	Updated         int
	Unchanged       int
	SkippedNoParent int
}

// Execute reconciles the input episodes into animes / anime_classifications.
//
// Following the write-usecase rule, every read (the episodes with their resolved
// parent anime_id and the existing animes / classifications) happens before the
// transaction; the transaction in applyPlan performs persistence only.
//
// [Ja] Execute は入力 episodes を animes / anime_classifications へリコンサイルする。
//
// 書き込み UseCase のルールに従い、すべての取得 (解決済みの親 anime_id を伴う episodes と
// 既存の animes / 分類) はトランザクション前に行い、applyPlan のトランザクション内は
// 永続化のみを行う。
func (uc *SyncEpisodesToAnimesUsecase) Execute(ctx context.Context, input SyncEpisodesToAnimesInput) (*SyncEpisodesToAnimesResult, error) {
	episodes, err := uc.episodeRepo.ListForAnimeSyncByIDs(ctx, input.EpisodeIDs)
	if err != nil {
		return nil, fmt.Errorf("同期対象 episodes の取得に失敗: %w", err)
	}
	if len(episodes) == 0 {
		return &SyncEpisodesToAnimesResult{}, nil
	}

	// Batch-fetch the existing animes / classifications for the already-mapped
	// episodes in one query each, avoiding N per-row lookups during diff detection.
	//
	// [Ja] 既にマッピング済みの episodes について既存の animes / 分類を 1 クエリずつで
	// 一括取得し、差分検出時の行単位 N 回ルックアップを避ける。
	mappedAnimeIDs := collectMappedAnimeIDsFromEpisodes(episodes)

	existingAnimes, err := uc.animeRepo.ListByIDs(ctx, mappedAnimeIDs)
	if err != nil {
		return nil, fmt.Errorf("既存 animes の取得に失敗: %w", err)
	}
	existingClassifications, err := uc.animeClassificationRepo.ListByAnimeIDs(ctx, mappedAnimeIDs)
	if err != nil {
		return nil, fmt.Errorf("既存 anime_classifications の取得に失敗: %w", err)
	}

	plan := planEpisodeAnimeSync(episodes, indexAnimesByID(existingAnimes), indexClassificationsByAnimeID(existingClassifications))

	return uc.applyPlan(ctx, plan)
}

// collectMappedAnimeIDsFromEpisodes returns the anime IDs of the episodes that are
// already mapped.
//
// [Ja] collectMappedAnimeIDsFromEpisodes は既にマッピング済みの episodes の anime ID を返す。
func collectMappedAnimeIDsFromEpisodes(episodes []*model.Episode) []model.AnimeID {
	ids := make([]model.AnimeID, 0, len(episodes))
	for _, e := range episodes {
		if e.AnimeID != nil {
			ids = append(ids, *e.AnimeID)
		}
	}
	return ids
}

// episodeAnimeCreate is a unit of work for an unmapped episode: insert the anime,
// insert its kind='episode' classification (anime_id is filled with the inserted
// anime's ID at apply time), and write episodes.anime_id back.
//
// [Ja] episodeAnimeCreate は未マッピング episode に対する作成の単位。anime を挿入し、
// その kind='episode' 分類を挿入し (anime_id は適用時に挿入された anime の ID で埋める)、
// episodes.anime_id を書き戻す。
type episodeAnimeCreate struct {
	episodeID      model.EpisodeID
	anime          repository.CreateAnimeParams
	classification repository.CreateAnimeClassificationParams
}

// episodeAnimeUpdate is a unit of work for a mapped episode whose content drifted.
// Each pointer is nil when that part is already in sync; classificationCreate is set
// only when the anime exists but its classification row is missing.
//
// [Ja] episodeAnimeUpdate は内容がずれたマッピング済み episode に対する更新の単位。
// 既に同期済みの部分のポインタは nil。classificationCreate は anime は存在するが
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

// planEpisodeAnimeSync decides, per episode, whether to create / update / leave
// untouched / defer, without performing any I/O. An episode is deferred when its
// parent work is not yet synced (resolved ParentAnimeID is nil), because the episode
// classification requires a non-NULL parent_anime_id. An episode is created when it
// is unmapped, or when it is mapped but its anime row is missing (self-healing for
// an inconsistent mapping).
//
// [Ja] planEpisodeAnimeSync は I/O を行わずに、episode ごとに作成 / 更新 / 据え置き /
// 繰り延べを判断する。親作品が未同期 (解決した ParentAnimeID が nil) の episode は、
// episode 分類が NOT NULL の parent_anime_id を要するため繰り延べる。episode が未マッピングの
// とき、またはマッピング済みでも anime 行が欠落しているとき (不整合なマッピングの自己修復) に
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

// applyPlan persists the plan in a single transaction. Each create inserts the
// anime, then its classification with the inserted anime's ID, then writes
// episodes.anime_id back, so an episode is never left mapped to a half-built anime.
//
// [Ja] applyPlan は計画を 1 トランザクションで永続化する。各作成は anime を挿入し、
// 続けてその分類を挿入された anime の ID で挿入し、episodes.anime_id を書き戻すため、
// episode が中途半端な anime にマッピングされたまま残ることはない。
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
			return nil, fmt.Errorf("anime の作成に失敗 (episode_id=%d): %w", c.episodeID, err)
		}

		classificationParams := c.classification
		classificationParams.AnimeID = anime.ID
		if _, err := classificationRepo.Create(ctx, classificationParams); err != nil {
			return nil, fmt.Errorf("anime_classification の作成に失敗 (episode_id=%d): %w", c.episodeID, err)
		}

		if err := episodeRepo.UpdateAnimeID(ctx, c.episodeID, anime.ID); err != nil {
			return nil, fmt.Errorf("episodes.anime_id の書き戻しに失敗 (episode_id=%d): %w", c.episodeID, err)
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

// animeCreateParamsFromEpisode maps an episode onto the layer-1 anime attributes
// for a fresh insert. Episodes only source title / title_ro / title_en / status /
// archive_message; every other column (title_kana, the title_alter family, media,
// release_status, synopsis family) stays at its zero value (NULL).
//
// [Ja] animeCreateParamsFromEpisode は episode を新規挿入用の第 1 層 anime 属性に写像する。
// episode が源泉とするのは title / title_ro / title_en / status / archive_message のみで、
// 他のカラム (title_kana、title_alter 系、media、release_status、synopsis 系) はゼロ値
// (NULL) のまま残す。
func animeCreateParamsFromEpisode(e *model.Episode) repository.CreateAnimeParams {
	return repository.CreateAnimeParams{
		Title:          nullStringFromStringPtr(e.Title),
		TitleRo:        nullStringFromNonEmpty(e.TitleRo),
		TitleEn:        nullStringFromNonEmpty(e.TitleEn),
		Status:         episodeStatusToAnimeStatus(e.Status),
		ArchiveMessage: nullStringFromStringPtr(e.ArchiveMessage),
	}
}

// animeUpdateParamsFromEpisode maps an episode onto the layer-1 anime attributes for
// an update. The columns episodes do not source are carried over from the existing
// row so the sync never clobbers values set elsewhere (an editor, or another loader).
//
// [Ja] animeUpdateParamsFromEpisode は episode を更新用の第 1 層 anime 属性に写像する。
// episode が源泉としないカラムは既存行から引き継ぎ、同期が他所 (編集者や別ローダー) で
// 設定された値を上書きしないようにする。
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
		Status:           episodeStatusToAnimeStatus(e.Status),
		ArchiveMessage:   nullStringFromStringPtr(e.ArchiveMessage),
	}
}

// classificationCreateParamsFromEpisode maps an episode onto its kind='episode'
// classification. parent_anime_id (the parent work's anime) and sort_number are
// required to be non-NULL for an episode, while the work-only generation settings
// (number_format_id / episode_start_number / expected_episodes_count) and standalone
// stay at their zero value to satisfy the table's CHECK constraints.
//
// [Ja] classificationCreateParamsFromEpisode は episode をその kind='episode' 分類に
// 写像する。parent_anime_id (親作品の anime) と sort_number は episode では NOT NULL が要求され、
// work 専用の生成設定 (number_format_id / episode_start_number / expected_episodes_count) と
// standalone はテーブルの CHECK 制約を満たすためゼロ値のまま残す。
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

// episodeClassificationChanged reports whether the existing classification differs
// from the episode-derived desired state for the episode-relevant fields.
//
// [Ja] episodeClassificationChanged は既存分類が episode 由来の目標状態と episode 関連
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

// episodeStatusToAnimeStatus maps episode_status onto anime_status. The three
// episode values are a subset of anime_status; merged is anime-only and never
// produced here.
//
// [Ja] episodeStatusToAnimeStatus は episode_status を anime_status に写像する。
// episode の 3 値は anime_status の部分集合で、merged は anime 専用でありここでは生成されない。
func episodeStatusToAnimeStatus(status model.EpisodeStatus) model.AnimeStatus {
	switch status {
	case model.EpisodeStatusPublished:
		return model.AnimeStatusPublished
	case model.EpisodeStatusArchived:
		return model.AnimeStatusArchived
	case model.EpisodeStatusDeleted:
		return model.AnimeStatusDeleted
	default:
		return model.AnimeStatusPublished
	}
}

// numericStringFromFloatPtr renders an optional float as the canonical text for a
// NUMERIC column, using the shortest representation that round-trips exactly. A nil
// pointer maps to NULL (episodes.raw_number is nullable).
//
// [Ja] numericStringFromFloatPtr は任意の float を NUMERIC カラム用の正準テキストに描画する。
// 正確にラウンドトリップする最短表現を使う。nil ポインタは NULL に写像する
// (episodes.raw_number は NULL 許容)。
func numericStringFromFloatPtr(p *float64) sql.NullString {
	if p == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: strconv.FormatFloat(*p, 'f', -1, 64), Valid: true}
}

// animeIDPtrEqual reports whether two optional anime ID FKs are equal, treating two
// nils as equal.
//
// [Ja] animeIDPtrEqual は 2 つの任意のアニメ ID 外部キーが等しいかを返す。両方 nil は
// 等しいとみなす。
func animeIDPtrEqual(a, b *model.AnimeID) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

package usecase

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// workSourcedSeasonIsPrimary is the is_primary value works assign to the single season
// they source. Unlike anime_links / anime_official_accounts there is no kind / service
// discriminator column, so the works-managed key space is this slot: the reconcile
// limits deletes to is_primary rows, leaving a secondary season an editor might add
// directly (which gets is_primary=false) untouched (see reconcileSatellite). A partial
// UNIQUE index keeps at most one is_primary row per anime, which is why a season change
// must delete the old primary row before creating the new one (see applyPlan).
//
// [Ja] workSourcedSeasonIsPrimary は works が source する単一の季節に割り当てる is_primary
// 値。anime_links / anime_official_accounts と違い kind / service の判別列が無いため、works
// 管理下のキー空間はこのスロットになる。リコンサイルは削除を is_primary の行に限定し、編集者が
// 直接足しうる副次シーズン (is_primary=false を持つ) を壊さない (reconcileSatellite を参照)。
// 部分 UNIQUE インデックスが anime ごとに is_primary 行を高々 1 つに保つため、季節の変更は
// 新しい主行を作成する前に旧主行を削除しなければならない (applyPlan を参照)。
const workSourcedSeasonIsPrimary = true

// animeSeasonKey is the natural key reconcileSatellite matches desired rows (derived
// from works) against existing rows on: (anime_id, year, name). It includes the
// anime_id because a reconciler diffs a whole page of works at once, so a key without
// it would collide across works. name uses the empty SeasonName ("") to stand for a
// NULL name (season name undetermined); the enum values are never empty, so "" is
// unambiguous and mirrors the (anime_id, year, name) UNIQUE index built NULLS NOT
// DISTINCT (a NULL name collapses to a single value).
//
// [Ja] animeSeasonKey は reconcileSatellite が (works から導出した) あるべき行を既存行と
// 突合する自然キー (anime_id, year, name)。リコンサイラは works のページ全体を一括で突合する
// ため、anime_id を含めないと works をまたいでキーが衝突する。name は空の SeasonName ("") を
// NULL の name (季節名未定) を表すために使う。enum 値は決して空にならないため "" は曖昧でなく、
// NULLS NOT DISTINCT で張った (anime_id, year, name) UNIQUE インデックス (NULL の name は
// 単一の値に畳まれる) と対応する。
type animeSeasonKey struct {
	animeID model.AnimeID
	year    int32
	name    model.SeasonName
}

// SyncAnimeSeasonsUsecase is the phase 2 satellite reconciler for the anime_seasons
// table. Registered into SyncWorkSatellitesUsecase, it maps each anime-resolved work's
// season_year / season_name columns onto a season row of its anime, then creates /
// deletes the rows so they match. It is a self-contained write usecase mirroring
// SyncWorksToAnimesUsecase: it reads the existing rows, plans the diff via
// reconcileSatellite, and applies the plan in its own transaction. The Reconcile method
// (not Execute) satisfies the satelliteReconciler interface.
//
// Like SyncAnimeHashtagsUsecase there is no update path: works source the year and name
// (both in the natural key) plus is_primary (fixed true for the works-managed slot), so
// a changed season is a delete (old) plus a create (new), never an in-place update.
//
// [Ja] SyncAnimeSeasonsUsecase は anime_seasons テーブルに対するフェーズ 2 の別表
// リコンサイラ。SyncWorkSatellitesUsecase に登録され、anime 解決済みの各 work の season_year
// / season_name カラムをその anime の季節行に写像し、一致するよう行を作成 / 削除する。
// SyncWorksToAnimesUsecase を写した自己完結の書き込み UseCase で、既存行を読み、
// reconcileSatellite で差分を計画し、自前のトランザクションで適用する。Execute ではなく
// Reconcile メソッドが satelliteReconciler インターフェースを満たす。
//
// SyncAnimeHashtagsUsecase と同じく更新パスを持たない。works は year と name (どちらも
// 自然キー) に加え is_primary (works 管理スロットでは true 固定) を source するため、季節の
// 変更はその場の更新ではなく削除 (old) + 作成 (new) になる。
type SyncAnimeSeasonsUsecase struct {
	db   *sql.DB
	repo *repository.AnimeSeasonRepository
}

// NewSyncAnimeSeasonsUsecase constructs a SyncAnimeSeasonsUsecase.
//
// [Ja] NewSyncAnimeSeasonsUsecase は SyncAnimeSeasonsUsecase を生成する。
func NewSyncAnimeSeasonsUsecase(db *sql.DB, repo *repository.AnimeSeasonRepository) *SyncAnimeSeasonsUsecase {
	return &SyncAnimeSeasonsUsecase{db: db, repo: repo}
}

// Reconcile reconciles the anime_seasons rows for the given anime-resolved works.
// Following the write-usecase rule, the desired rows are derived and the existing rows
// are read before applyPlan opens a transaction; the transaction performs persistence
// only.
//
// [Ja] Reconcile は指定された anime 解決済み works について anime_seasons 行をリコンサイル
// する。書き込み UseCase のルールに従い、あるべき行の導出と既存行の取得は applyPlan が
// トランザクションを開くより前に行い、トランザクション内は永続化のみを行う。
func (uc *SyncAnimeSeasonsUsecase) Reconcile(ctx context.Context, works []*model.Work) (satelliteReconcileCounts, error) {
	desired := desiredAnimeSeasons(works)

	existing, err := uc.repo.ListByAnimeIDs(ctx, collectMappedAnimeIDs(works))
	if err != nil {
		return satelliteReconcileCounts{}, fmt.Errorf("既存 anime_seasons の取得に失敗: %w", err)
	}

	plan := reconcileSatellite(
		desired,
		existing,
		func(d repository.CreateAnimeSeasonParams) animeSeasonKey {
			return animeSeasonKey{animeID: d.AnimeID, year: d.Year, name: seasonNameKey(d.Name)}
		},
		func(e *model.AnimeSeason) animeSeasonKey {
			return animeSeasonKey{animeID: e.AnimeID, year: e.Year, name: seasonNameKey(e.Name)}
		},
		func(e *model.AnimeSeason) bool { return e.IsPrimary },
		// year, name and is_primary are all either part of the natural key or fixed for
		// the works-managed slot, so a key match means the existing row already equals
		// the desired one — there is nothing to update. A changed season is a different
		// key (delete old + create new), so changed is always false and the plan never
		// carries updates.
		//
		// [Ja] year / name / is_primary はいずれも自然キーの一部か works 管理スロットで固定
		// のため、キーが一致する = 既存行が既にあるべき行と等しい、ということで更新するものが
		// ない。季節の変更は別のキー (旧行の削除 + 新行の作成) になるため、changed は常に false
		// で、計画は更新を持たない。
		func(repository.CreateAnimeSeasonParams, *model.AnimeSeason) bool { return false },
	)

	return uc.applyPlan(ctx, plan)
}

// desiredAnimeSeasons derives the season rows a batch of works should have: season_year
// + season_name -> one is_primary season per work. A row is created only when
// season_year is present (a NULL year yields no row, so a work without one contributes
// nothing and any existing works-managed row for it is later deleted). season_name maps
// the legacy integer (1=winter, 2=spring, 3=summer, 4=autumn) onto the enum with autumn
// folded to fall; a NULL or out-of-range integer leaves the name nil (season name
// undetermined), still yielding a year-only row.
//
// [Ja] desiredAnimeSeasons は works のバッチが持つべき季節行を導出する。season_year +
// season_name -> work ごとに 1 つの is_primary な季節。行は season_year が存在するときだけ
// 作る (NULL の year は行を作らないため、それを持たない work は何も寄与せず、既存の works
// 管理下の行があれば後段で削除される)。season_name は旧 integer (1=winter, 2=spring,
// 3=summer, 4=autumn) を enum に写像し autumn は fall に寄せる。NULL や範囲外の integer は
// name を nil (季節名未定) のままにし、年のみの行を作る。
func desiredAnimeSeasons(works []*model.Work) []repository.CreateAnimeSeasonParams {
	desired := make([]repository.CreateAnimeSeasonParams, 0, len(works))
	for _, w := range works {
		if w.AnimeID == nil || w.SeasonYear == nil {
			continue
		}
		desired = append(desired, repository.CreateAnimeSeasonParams{
			AnimeID:   *w.AnimeID,
			Year:      *w.SeasonYear,
			Name:      seasonNameFromWorkInt(w.SeasonName),
			IsPrimary: workSourcedSeasonIsPrimary,
		})
	}
	return desired
}

// seasonNameFromWorkInt maps the legacy works.season_name integer to the SeasonName
// enum: 1=winter, 2=spring, 3=summer, 4=autumn (folded to fall). A NULL (nil) or
// out-of-range integer returns nil, meaning the season name is undetermined.
//
// [Ja] seasonNameFromWorkInt は旧 works.season_name の integer を SeasonName enum に写像
// する: 1=winter, 2=spring, 3=summer, 4=autumn (fall に寄せる)。NULL (nil) や範囲外の
// integer は nil を返し、季節名が未定であることを表す。
func seasonNameFromWorkInt(n *int32) *model.SeasonName {
	if n == nil {
		return nil
	}
	var name model.SeasonName
	switch *n {
	case 1:
		name = model.SeasonNameWinter
	case 2:
		name = model.SeasonNameSpring
	case 3:
		name = model.SeasonNameSummer
	case 4:
		name = model.SeasonNameFall
	default:
		return nil
	}
	return &name
}

// seasonNameKey reduces a nullable season name to the comparable key value, mapping nil
// (a NULL name) to the empty SeasonName "". Enum values are never empty, so "" cannot
// collide with a real name.
//
// [Ja] seasonNameKey は NULL 許容の季節名を比較可能なキー値に落とし込み、nil (NULL の name)
// を空の SeasonName "" に写像する。enum 値は決して空にならないため "" が実在の name と衝突
// することはない。
func seasonNameKey(name *model.SeasonName) model.SeasonName {
	if name == nil {
		return ""
	}
	return *name
}

// applyPlan persists the reconcile plan in a single transaction and returns the
// per-table counts. Seasons never produce updates (see Reconcile), so only deletes and
// creates are applied, and deletes run first: the partial UNIQUE index on (anime_id)
// WHERE is_primary forbids two is_primary rows for an anime even within a transaction,
// so a season change (delete old primary + create new primary) must remove the old row
// before inserting the new one. When there is nothing to write it returns early without
// opening a transaction, so an already-synced page costs no write.
//
// [Ja] applyPlan はリコンサイル計画を 1 トランザクションで永続化し、テーブルごとの件数を
// 返す。季節は更新を生まない (Reconcile を参照) ため削除と作成のみを適用し、削除を先に走らせる。
// (anime_id) WHERE is_primary の部分 UNIQUE インデックスはトランザクション内であっても anime
// あたり 2 つの is_primary 行を許さないため、季節の変更 (旧主行の削除 + 新主行の作成) は新行を
// 挿入する前に旧行を削除しなければならない。書き込むものが無ければトランザクションを開かずに
// 早期 return するため、既に同期済みのページは書き込みコストがかからない。
func (uc *SyncAnimeSeasonsUsecase) applyPlan(ctx context.Context, plan satelliteReconcilePlan[repository.CreateAnimeSeasonParams, *model.AnimeSeason]) (satelliteReconcileCounts, error) {
	counts := satelliteReconcileCounts{Unchanged: plan.unchanged}
	if len(plan.creates) == 0 && len(plan.deletes) == 0 {
		return counts, nil
	}

	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return satelliteReconcileCounts{}, fmt.Errorf("トランザクションの開始に失敗: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	repo := uc.repo.WithTx(tx)

	for _, existing := range plan.deletes {
		if err := repo.Delete(ctx, existing.ID); err != nil {
			return satelliteReconcileCounts{}, fmt.Errorf("anime_season の削除に失敗 (id=%d): %w", existing.ID, err)
		}
		counts.Deleted++
	}

	for _, create := range plan.creates {
		if _, err := repo.Create(ctx, create); err != nil {
			return satelliteReconcileCounts{}, fmt.Errorf("anime_season の作成に失敗 (anime_id=%d, year=%d): %w", create.AnimeID, create.Year, err)
		}
		counts.Created++
	}

	if err := tx.Commit(); err != nil {
		return satelliteReconcileCounts{}, fmt.Errorf("トランザクションのコミットに失敗: %w", err)
	}

	return counts, nil
}

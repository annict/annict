package usecase

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// workSourcedSeasonIsPrimaryはworksがsourceする単一の季節に割り当てるis_primary
// 値。anime_links / anime_official_accountsと違いkind / serviceの判別列が無いため、works
// 管理下のキー空間はこのスロットになる。リコンサイルは削除をis_primaryの行に限定し、編集者が
// 直接足しうる副次シーズン (is_primary=falseを持つ) を壊さない (reconcileSatelliteを参照)。
// 部分UNIQUEインデックスがanimeごとにis_primary行を高々1つに保つため、季節の変更は
// 新しい主行を作成する前に旧主行を削除しなければならない (applyPlanを参照)。
const workSourcedSeasonIsPrimary = true

// animeSeasonKeyはreconcileSatelliteが (worksから導出した) あるべき行を既存行と
// 突合する自然キー (anime_id, year, name)。リコンサイラはworksのページ全体を一括で突合する
// ため、anime_idを含めないとworksをまたいでキーが衝突する。nameは空のSeasonName ("") を
// NULLのname (季節名未定) を表すために使う。enum値は決して空にならないため "" は曖昧でなく、
// NULLS NOT DISTINCTで張った (anime_id, year, name) UNIQUEインデックス (NULLのnameは
// 単一の値に畳まれる) と対応する。
type animeSeasonKey struct {
	animeID model.AnimeID
	year    int32
	name    model.SeasonName
}

// SyncAnimeSeasonsUsecaseはanime_seasonsテーブルに対するフェーズ2の別表
// リコンサイラ。SyncWorkSatellitesUsecaseに登録され、anime解決済みの各workのseason_year
// / season_nameカラムをそのanimeの季節行に写像し、一致するよう行を作成 / 削除する。
// SyncWorksToAnimesUsecaseを写した自己完結の書き込みUseCaseで、既存行を読み、
// reconcileSatelliteで差分を計画し、自前のトランザクションで適用する。Executeではなく
// ReconcileメソッドがsatelliteReconcilerインターフェースを満たす。
//
// SyncAnimeHashtagsUsecaseと同じく更新パスを持たない。worksはyearとname (どちらも
// 自然キー) に加えis_primary (works管理スロットではtrue固定) をsourceするため、季節の
// 変更はその場の更新ではなく削除 (old) + 作成 (new) になる。
type SyncAnimeSeasonsUsecase struct {
	db   *sql.DB
	repo *repository.AnimeSeasonRepository
}

// NewSyncAnimeSeasonsUsecaseはSyncAnimeSeasonsUsecaseを生成する。
func NewSyncAnimeSeasonsUsecase(db *sql.DB, repo *repository.AnimeSeasonRepository) *SyncAnimeSeasonsUsecase {
	return &SyncAnimeSeasonsUsecase{db: db, repo: repo}
}

// Reconcileは指定されたanime解決済みworksについてanime_seasons行をリコンサイル
// する。書き込みUseCaseのルールに従い、あるべき行の導出と既存行の取得はapplyPlanが
// トランザクションを開くより前に行い、トランザクション内は永続化のみを行う。
func (uc *SyncAnimeSeasonsUsecase) Reconcile(ctx context.Context, works []*model.Work) (satelliteReconcileCounts, error) {
	existing, err := uc.repo.ListByAnimeIDs(ctx, collectMappedAnimeIDs(works))
	if err != nil {
		return satelliteReconcileCounts{}, fmt.Errorf("既存anime_seasonsの取得に失敗: %w", err)
	}

	return uc.applyPlan(ctx, planAnimeSeasons(works, existing))
}

// planAnimeSeasonsは指定されたanime解決済みworksのanime_seasons行について、
// 既存行に対するリコンサイル計画を組み立てる。フェーズ2のバッチリコンサイラ (Reconcile) と
// フェーズ3の作品 作成 / 更新 の両書き (planWorkSatellites) で共有し、あるべき行の導出・
// 自然キー・削除限定・変更検出を両経路で単一の正本に保つ。
func planAnimeSeasons(works []*model.Work, existing []*model.AnimeSeason) satelliteReconcilePlan[repository.CreateAnimeSeasonParams, *model.AnimeSeason] {
	return reconcileSatellite(
		desiredAnimeSeasons(works),
		existing,
		func(d repository.CreateAnimeSeasonParams) animeSeasonKey {
			return animeSeasonKey{animeID: d.AnimeID, year: d.Year, name: seasonNameKey(d.Name)}
		},
		func(e *model.AnimeSeason) animeSeasonKey {
			return animeSeasonKey{animeID: e.AnimeID, year: e.Year, name: seasonNameKey(e.Name)}
		},
		func(e *model.AnimeSeason) bool { return e.IsPrimary },
		// year / name / is_primaryはいずれも自然キーの一部かworks管理スロットで固定
		// のため、キーが一致する = 既存行が既にあるべき行と等しい、ということで更新するものが
		// ない。季節の変更は別のキー (旧行の削除 + 新行の作成) になるため、changedは常にfalse
		// で、計画は更新を持たない。
		func(repository.CreateAnimeSeasonParams, *model.AnimeSeason) bool { return false },
	)
}

// desiredAnimeSeasonsはworksのバッチが持つべき季節行を導出する。season_year +
// season_name -> workごとに1つのis_primaryな季節。行はseason_yearが存在するときだけ
// 作る (NULLのyearは行を作らないため、それを持たないworkは何も寄与せず、既存のworks
// 管理下の行があれば後段で削除される)。season_nameは旧integer (1=winter, 2=spring,
// 3=summer, 4=autumn) をenumに写像しautumnはfallに寄せる。NULLや範囲外のintegerは
// nameをnil (季節名未定) のままにし、年のみの行を作る。
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

// seasonNameFromWorkIntは旧works.season_nameのintegerをSeasonName enumに写像
// する: 1=winter, 2=spring, 3=summer, 4=autumn (fallに寄せる)。NULL (nil) や範囲外の
// integerはnilを返し、季節名が未定であることを表す。
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

// seasonNameKeyはNULL許容の季節名を比較可能なキー値に落とし込み、nil (NULLのname)
// を空のSeasonName "" に写像する。enum値は決して空にならないため "" が実在のnameと衝突
// することはない。
func seasonNameKey(name *model.SeasonName) model.SeasonName {
	if name == nil {
		return ""
	}
	return *name
}

// applyPlanはリコンサイル計画を1トランザクションで永続化し、テーブルごとの件数を
// 返す。季節は更新を生まない (Reconcileを参照) ため削除と作成のみを適用し、削除を先に走らせる。
// (anime_id) WHERE is_primaryの部分UNIQUEインデックスはトランザクション内であってもanime
// あたり2つのis_primary行を許さないため、季節の変更 (旧主行の削除 + 新主行の作成) は新行を
// 挿入する前に旧行を削除しなければならない。書き込むものが無ければトランザクションを開かずに
// 早期returnするため、既に同期済みのページは書き込みコストがかからない。
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
			return satelliteReconcileCounts{}, fmt.Errorf("anime_seasonの削除に失敗 (id=%d): %w", existing.ID, err)
		}
		counts.Deleted++
	}

	for _, create := range plan.creates {
		if _, err := repo.Create(ctx, create); err != nil {
			return satelliteReconcileCounts{}, fmt.Errorf("anime_seasonの作成に失敗 (anime_id=%d, year=%d): %w", create.AnimeID, create.Year, err)
		}
		counts.Created++
	}

	if err := tx.Commit(); err != nil {
		return satelliteReconcileCounts{}, fmt.Errorf("トランザクションのコミットに失敗: %w", err)
	}

	return counts, nil
}

package usecase

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// workSourcedHashtagSortNumberはworksがsourceする単一ハッシュタグに割り当てる
// sort_number。他の別表と違いanime_hashtagsにはkind / serviceの判別列が無いため、works
// 管理下のキー空間はこのスロットになる。リコンサイルは削除をsort_number 0の行に限定し、
// 編集者が直接足しうる2つ目のハッシュタグ (非ゼロのsort_numberを持つ) を壊さない
// (reconcileSatelliteを参照)。
const workSourcedHashtagSortNumber int32 = 0

// animeHashtagKeyはreconcileSatelliteが (worksから導出した) あるべき行を既存行と
// 突合する自然キー (anime_id, hashtag)。リコンサイラはworksのページ全体を一括で突合する
// ため、anime_idを含めないとworksをまたいでキーが衝突する。hashtagを (anime_idのスロット
// だけでなく) 含めることで、1つのanimeが複数のハッシュタグ行 (編集者追加の2つ目のタグ) を
// 持っても差分マップ内でキーが衝突しない。
type animeHashtagKey struct {
	animeID model.AnimeID
	hashtag string
}

// SyncAnimeHashtagsUsecaseはanime_hashtagsテーブルに対するフェーズ2の別表
// リコンサイラ。SyncWorkSatellitesUsecaseに登録され、anime解決済みの各workの
// twitter_hashtagカラムをそのanimeのハッシュタグ行に写像し、一致するよう行を作成 / 削除
// する。SyncWorksToAnimesUsecaseを写した自己完結の書き込みUseCaseで、既存行を読み、
// reconcileSatelliteで差分を計画し、自前のトランザクションで適用する。Executeではなく
// ReconcileメソッドがsatelliteReconcilerインターフェースを満たす。
//
// 兄弟のリコンサイラと違い更新パスを持たない。hashtagはworksがsourceする値そのもので
// あると同時に自然キーでもあるため、タグの変更はその場の更新ではなく削除 (old) + 作成 (new)
// になる。
type SyncAnimeHashtagsUsecase struct {
	db   *sql.DB
	repo *repository.AnimeHashtagRepository
}

// NewSyncAnimeHashtagsUsecaseはSyncAnimeHashtagsUsecaseを生成する。
func NewSyncAnimeHashtagsUsecase(db *sql.DB, repo *repository.AnimeHashtagRepository) *SyncAnimeHashtagsUsecase {
	return &SyncAnimeHashtagsUsecase{db: db, repo: repo}
}

// Reconcileは指定されたanime解決済みworksについてanime_hashtags行をリコンサイル
// する。書き込みUseCaseのルールに従い、あるべき行の導出と既存行の取得はapplyPlanが
// トランザクションを開くより前に行い、トランザクション内は永続化のみを行う。
func (uc *SyncAnimeHashtagsUsecase) Reconcile(ctx context.Context, works []*model.Work) (satelliteReconcileCounts, error) {
	existing, err := uc.repo.ListByAnimeIDs(ctx, collectMappedAnimeIDs(works))
	if err != nil {
		return satelliteReconcileCounts{}, fmt.Errorf("既存anime_hashtagsの取得に失敗: %w", err)
	}

	return uc.applyPlan(ctx, planAnimeHashtags(works, existing))
}

// planAnimeHashtagsは指定されたanime解決済みworksのanime_hashtags行について、
// 既存行に対するリコンサイル計画を組み立てる。フェーズ2のバッチリコンサイラ (Reconcile) と
// フェーズ3の作品 作成 / 更新 の両書き (planWorkSatellites) で共有し、あるべき行の導出・
// 自然キー・削除限定・変更検出を両経路で単一の正本に保つ。
func planAnimeHashtags(works []*model.Work, existing []*model.AnimeHashtag) satelliteReconcilePlan[repository.CreateAnimeHashtagParams, *model.AnimeHashtag] {
	return reconcileSatellite(
		desiredAnimeHashtags(works),
		existing,
		func(d repository.CreateAnimeHashtagParams) animeHashtagKey {
			return animeHashtagKey{animeID: d.AnimeID, hashtag: d.Hashtag}
		},
		func(e *model.AnimeHashtag) animeHashtagKey {
			return animeHashtagKey{animeID: e.AnimeID, hashtag: e.Hashtag}
		},
		func(e *model.AnimeHashtag) bool { return e.SortNumber == workSourcedHashtagSortNumber },
		// hashtagはworksがsourceする値そのものであると同時に自然キーでもあるため、
		// キーが一致する = 既存行が既にあるべき行と等しい、ということで更新するものがない。
		// タグの変更は別のキー (旧行の削除 + 新行の作成) になるため、changedは常にfalseで、
		// 計画は更新を持たない。
		func(repository.CreateAnimeHashtagParams, *model.AnimeHashtag) bool { return false },
	)
}

// desiredAnimeHashtagsはworksのバッチが持つべきハッシュタグ行を導出する。
// twitter_hashtag -> ハッシュタグ。タグはそのまま使う (Railsは先頭の '#' を付けずに保持して
// おり、anime_hashtagsが持つ素の文字列の形と一致する)。twitter_hashtagがNULLまたは空の
// ときは行を作らないため、それを持たないworkは何も寄与せず、既存のworks管理下の行があれば
// 後段で削除される。
func desiredAnimeHashtags(works []*model.Work) []repository.CreateAnimeHashtagParams {
	desired := make([]repository.CreateAnimeHashtagParams, 0, len(works))
	for _, w := range works {
		if w.AnimeID == nil {
			continue
		}
		if hashtag := hashtagFromStringPtr(w.TwitterHashtag); hashtag != "" {
			desired = append(desired, repository.CreateAnimeHashtagParams{
				AnimeID: *w.AnimeID,
				Hashtag: hashtag,
			})
		}
	}
	return desired
}

// hashtagFromStringPtrはworksのNULL許容ハッシュタグ列を読み、NULL (nil) と空文字列
// の両方を "" (行なし) に写像する。Railsは欠損のハッシュタグをNULLまたは "" で持つため、
// ここではどちらも欠損として扱う。
func hashtagFromStringPtr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// applyPlanはリコンサイル計画を1トランザクションで永続化し、テーブルごとの件数を
// 返す。ハッシュタグは更新を生まない (Reconcileを参照) ため作成と削除のみを適用する。
// 書き込むものが無ければトランザクションを開かずに早期returnするため、既に同期済みの
// ページは書き込みコストがかからない。
func (uc *SyncAnimeHashtagsUsecase) applyPlan(ctx context.Context, plan satelliteReconcilePlan[repository.CreateAnimeHashtagParams, *model.AnimeHashtag]) (satelliteReconcileCounts, error) {
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

	for _, create := range plan.creates {
		if _, err := repo.Create(ctx, create); err != nil {
			return satelliteReconcileCounts{}, fmt.Errorf("anime_hashtagの作成に失敗 (anime_id=%d, hashtag=%s): %w", create.AnimeID, create.Hashtag, err)
		}
		counts.Created++
	}

	for _, existing := range plan.deletes {
		if err := repo.Delete(ctx, existing.ID); err != nil {
			return satelliteReconcileCounts{}, fmt.Errorf("anime_hashtagの削除に失敗 (id=%d): %w", existing.ID, err)
		}
		counts.Deleted++
	}

	if err := tx.Commit(); err != nil {
		return satelliteReconcileCounts{}, fmt.Errorf("トランザクションのコミットに失敗: %w", err)
	}

	return counts, nil
}

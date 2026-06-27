package usecase

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// workSourcedHashtagSortNumber is the sort_number works assign to the single hashtag
// they source. Unlike the other satellite tables, anime_hashtags has no kind / service
// discriminator column, so the works-managed key space is this slot: the reconcile
// limits deletes to sort_number 0 rows, leaving a second hashtag an editor might add
// directly (which gets a non-zero sort_number) untouched (see reconcileSatellite).
//
// [Ja] workSourcedHashtagSortNumber は works が source する単一ハッシュタグに割り当てる
// sort_number。他の別表と違い anime_hashtags には kind / service の判別列が無いため、works
// 管理下のキー空間はこのスロットになる。リコンサイルは削除を sort_number 0 の行に限定し、
// 編集者が直接足しうる 2 つ目のハッシュタグ (非ゼロの sort_number を持つ) を壊さない
// (reconcileSatellite を参照)。
const workSourcedHashtagSortNumber int32 = 0

// animeHashtagKey is the natural key reconcileSatellite matches desired rows (derived
// from works) against existing rows on: (anime_id, hashtag). It includes the anime_id
// because a reconciler diffs a whole page of works at once, so a key without it would
// collide across works. Including the hashtag (not just the anime_id slot) is also what
// keeps multiple hashtag rows per anime — an editor-added second tag — from colliding in
// the diff map.
//
// [Ja] animeHashtagKey は reconcileSatellite が (works から導出した) あるべき行を既存行と
// 突合する自然キー (anime_id, hashtag)。リコンサイラは works のページ全体を一括で突合する
// ため、anime_id を含めないと works をまたいでキーが衝突する。hashtag を (anime_id のスロット
// だけでなく) 含めることで、1 つの anime が複数のハッシュタグ行 (編集者追加の 2 つ目のタグ) を
// 持っても差分マップ内でキーが衝突しない。
type animeHashtagKey struct {
	animeID model.AnimeID
	hashtag string
}

// SyncAnimeHashtagsUsecase is the phase 2 satellite reconciler for the anime_hashtags
// table. Registered into SyncWorkSatellitesUsecase, it maps each anime-resolved work's
// twitter_hashtag column onto a hashtag row of its anime, then creates / deletes the rows
// so they match. It is a self-contained write usecase mirroring SyncWorksToAnimesUsecase:
// it reads the existing rows, plans the diff via reconcileSatellite, and applies the plan
// in its own transaction. The Reconcile method (not Execute) satisfies the
// satelliteReconciler interface.
//
// Unlike its sibling reconcilers there is no update path: the hashtag is the whole
// works-sourced value and also the natural key, so a changed tag is a delete (old) plus a
// create (new), never an in-place update.
//
// [Ja] SyncAnimeHashtagsUsecase は anime_hashtags テーブルに対するフェーズ 2 の別表
// リコンサイラ。SyncWorkSatellitesUsecase に登録され、anime 解決済みの各 work の
// twitter_hashtag カラムをその anime のハッシュタグ行に写像し、一致するよう行を作成 / 削除
// する。SyncWorksToAnimesUsecase を写した自己完結の書き込み UseCase で、既存行を読み、
// reconcileSatellite で差分を計画し、自前のトランザクションで適用する。Execute ではなく
// Reconcile メソッドが satelliteReconciler インターフェースを満たす。
//
// 兄弟のリコンサイラと違い更新パスを持たない。hashtag は works が source する値そのもので
// あると同時に自然キーでもあるため、タグの変更はその場の更新ではなく削除 (old) + 作成 (new)
// になる。
type SyncAnimeHashtagsUsecase struct {
	db   *sql.DB
	repo *repository.AnimeHashtagRepository
}

// NewSyncAnimeHashtagsUsecase constructs a SyncAnimeHashtagsUsecase.
//
// [Ja] NewSyncAnimeHashtagsUsecase は SyncAnimeHashtagsUsecase を生成する。
func NewSyncAnimeHashtagsUsecase(db *sql.DB, repo *repository.AnimeHashtagRepository) *SyncAnimeHashtagsUsecase {
	return &SyncAnimeHashtagsUsecase{db: db, repo: repo}
}

// Reconcile reconciles the anime_hashtags rows for the given anime-resolved works.
// Following the write-usecase rule, the desired rows are derived and the existing rows
// are read before applyPlan opens a transaction; the transaction performs persistence
// only.
//
// [Ja] Reconcile は指定された anime 解決済み works について anime_hashtags 行をリコンサイル
// する。書き込み UseCase のルールに従い、あるべき行の導出と既存行の取得は applyPlan が
// トランザクションを開くより前に行い、トランザクション内は永続化のみを行う。
func (uc *SyncAnimeHashtagsUsecase) Reconcile(ctx context.Context, works []*model.Work) (satelliteReconcileCounts, error) {
	desired := desiredAnimeHashtags(works)

	existing, err := uc.repo.ListByAnimeIDs(ctx, collectMappedAnimeIDs(works))
	if err != nil {
		return satelliteReconcileCounts{}, fmt.Errorf("既存 anime_hashtags の取得に失敗: %w", err)
	}

	plan := reconcileSatellite(
		desired,
		existing,
		func(d repository.CreateAnimeHashtagParams) animeHashtagKey {
			return animeHashtagKey{animeID: d.AnimeID, hashtag: d.Hashtag}
		},
		func(e *model.AnimeHashtag) animeHashtagKey {
			return animeHashtagKey{animeID: e.AnimeID, hashtag: e.Hashtag}
		},
		func(e *model.AnimeHashtag) bool { return e.SortNumber == workSourcedHashtagSortNumber },
		// The hashtag is the entire works-sourced value and also the natural key, so a
		// key match means the existing row already equals the desired one — there is
		// nothing to update. A changed tag is a different key (delete old + create new),
		// so changed is always false and the plan never carries updates.
		//
		// [Ja] hashtag は works が source する値そのものであると同時に自然キーでもあるため、
		// キーが一致する = 既存行が既にあるべき行と等しい、ということで更新するものがない。
		// タグの変更は別のキー (旧行の削除 + 新行の作成) になるため、changed は常に false で、
		// 計画は更新を持たない。
		func(repository.CreateAnimeHashtagParams, *model.AnimeHashtag) bool { return false },
	)

	return uc.applyPlan(ctx, plan)
}

// desiredAnimeHashtags derives the hashtag rows a batch of works should have:
// twitter_hashtag -> a hashtag. The tag is taken verbatim (Rails already stores it
// without a leading '#', matching the bare-string form anime_hashtags holds). A NULL or
// empty twitter_hashtag yields no row, so a work without one contributes nothing and any
// existing works-managed row for it is later deleted.
//
// [Ja] desiredAnimeHashtags は works のバッチが持つべきハッシュタグ行を導出する。
// twitter_hashtag -> ハッシュタグ。タグはそのまま使う (Rails は先頭の '#' を付けずに保持して
// おり、anime_hashtags が持つ素の文字列の形と一致する)。twitter_hashtag が NULL または空の
// ときは行を作らないため、それを持たない work は何も寄与せず、既存の works 管理下の行があれば
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

// hashtagFromStringPtr reads a works nullable hashtag column, mapping both NULL (nil) and
// the empty string to "" (no row). Rails stores an absent hashtag as NULL or "", so both
// are treated as missing here.
//
// [Ja] hashtagFromStringPtr は works の NULL 許容ハッシュタグ列を読み、NULL (nil) と空文字列
// の両方を "" (行なし) に写像する。Rails は欠損のハッシュタグを NULL または "" で持つため、
// ここではどちらも欠損として扱う。
func hashtagFromStringPtr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// applyPlan persists the reconcile plan in a single transaction and returns the
// per-table counts. Hashtags never produce updates (see Reconcile), so only creates and
// deletes are applied; when there is nothing to write it returns early without opening a
// transaction, so an already-synced page costs no write.
//
// [Ja] applyPlan はリコンサイル計画を 1 トランザクションで永続化し、テーブルごとの件数を
// 返す。ハッシュタグは更新を生まない (Reconcile を参照) ため作成と削除のみを適用する。
// 書き込むものが無ければトランザクションを開かずに早期 return するため、既に同期済みの
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
			return satelliteReconcileCounts{}, fmt.Errorf("anime_hashtag の作成に失敗 (anime_id=%d, hashtag=%s): %w", create.AnimeID, create.Hashtag, err)
		}
		counts.Created++
	}

	for _, existing := range plan.deletes {
		if err := repo.Delete(ctx, existing.ID); err != nil {
			return satelliteReconcileCounts{}, fmt.Errorf("anime_hashtag の削除に失敗 (id=%d): %w", existing.ID, err)
		}
		counts.Deleted++
	}

	if err := tx.Commit(); err != nil {
		return satelliteReconcileCounts{}, fmt.Errorf("トランザクションのコミットに失敗: %w", err)
	}

	return counts, nil
}

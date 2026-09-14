package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/model"
)

// satelliteWorkLoaderは別表ソース列を射影したworksをロードする。実体は
// *repository.WorkRepository。DBに触れず、テスト自身のworksを返すフェイクローダーで
// オーケストレーションをテストできるよう狭いインターフェースにしている (ローダーSQL自体は
// リポジトリテストでカバーする)。
type satelliteWorkLoader interface {
	ListForSatelliteSyncByIDs(ctx context.Context, workIDs []model.WorkID) ([]*model.Work, error)
}

// satelliteReconcilerはanime解決済みのworksのバッチに対して1つの別表
// (例: anime_external_ids) をリコンサイルし、テーブルごとの件数を返す。タスク2-8以降が
// テーブルごとに1つずつ登録する。各リコンサイラは既存行を読み、reconcileSatelliteで差分を
// 計画し、自前のトランザクションで作成 / 更新 / 削除を適用する自己完結した書き込みUseCase
// (SyncWorksToAnimesUsecaseを写したもの)。登録された各リコンサイラは既存行を読み差分を
// 永続化する。1つも登録されていなければ、本パスはworksをロードしanime_id未解決で
// 繰り延べた件数を数えるだけになる。
//
// Reconcileに渡されるworksスライスは登録済みの全リコンサイラで共有される同一の
// ものなので、リコンサイラはこれをread-onlyとして扱う。スライスや *model.Work
// ポインタ経由で到達できるフィールドを変更すると、同じページで後に走るリコンサイラに
// その変更が漏れる。
type satelliteReconciler interface {
	Reconcile(ctx context.Context, works []*model.Work) (satelliteReconcileCounts, error)
}

// satelliteReconcileCountsは1つの別表のリコンサイル結果。リコンサイラの戻り値と、
// リコンサイラ横断で積算する集計の双方に使う。
type satelliteReconcileCounts struct {
	Created   int
	Updated   int
	Deleted   int
	Unchanged int
}

// SyncWorkSatellitesUsecaseはフェーズ2の第3パス。worksとepisodesをanimes /
// anime_classificationsにリコンサイルした後、works (episodesではなく) がsourceとする
// 6つの別表をリコンサイルする。works同期 (タスク2-2) に織り込まず独立したパスとして走らせ、
// works同期の中核 (写像・差分・anime_id書き戻し) を小さく保ち、テーブルごとのリコンサイラを
// 1つずつ足せるようにする。
//
// anime_idが解決済みのworkだけをリコンサイルし、未マッピングのworkは繰り延べる
// (SkippedNoAnimeに数える)。worksパスがanime_idを書き戻した後の実行で取り込まれる。
// リコンサイルは冪等なので繰り延べは安全。
type SyncWorkSatellitesUsecase struct {
	workLoader  satelliteWorkLoader
	reconcilers []satelliteReconciler
}

// NewSyncWorkSatellitesUsecaseはSyncWorkSatellitesUsecaseを生成する。リコンサイラは
// 可変長で渡し、タスク2-8以降が別表ごとに1つずつ登録できるようにする。0個でも有効で、
// その場合は本パスがworksのロードと未マッピングworkの繰り延べだけを行う。
func NewSyncWorkSatellitesUsecase(
	workLoader satelliteWorkLoader,
	reconcilers ...satelliteReconciler,
) *SyncWorkSatellitesUsecase {
	return &SyncWorkSatellitesUsecase{
		workLoader:  workLoader,
		reconcilers: reconcilers,
	}
}

// SyncWorkSatellitesInputは今回のリコンサイル対象のwork IDを保持する
// (通常はバッチジョブが駆動する全件スキャンの1ページ分)。
type SyncWorkSatellitesInput struct {
	WorkIDs []model.WorkID
}

// SyncWorkSatellitesResultは全別表で集計したリコンサイル結果の件数を報告する。
// Created / Updated / Deletedの合計が、正本切り替え判定が依拠する差分検出メトリクス。
// SkippedNoAnimeはanime_id未解決で繰り延べたworkの件数。
type SyncWorkSatellitesResult struct {
	Processed      int
	Created        int
	Updated        int
	Deleted        int
	Unchanged      int
	SkippedNoAnime int
}

// Executeは入力worksについて別表をリコンサイルする。書き込みUseCaseのルールに従い、
// worksはどのリコンサイラがトランザクションを開くより前にロードする。登録された各リコンサイラ
// が既存行を読んで差分を永続化する。
func (uc *SyncWorkSatellitesUsecase) Execute(ctx context.Context, input SyncWorkSatellitesInput) (*SyncWorkSatellitesResult, error) {
	works, err := uc.workLoader.ListForSatelliteSyncByIDs(ctx, input.WorkIDs)
	if err != nil {
		return nil, fmt.Errorf("別表同期対象worksの取得に失敗: %w", err)
	}

	resolved := worksWithAnime(works)
	result := &SyncWorkSatellitesResult{
		Processed:      len(works),
		SkippedNoAnime: len(works) - len(resolved),
	}
	if len(resolved) == 0 {
		return result, nil
	}

	for _, reconciler := range uc.reconcilers {
		counts, err := reconciler.Reconcile(ctx, resolved)
		if err != nil {
			return nil, fmt.Errorf("別表のリコンサイルに失敗: %w", err)
		}
		result.Created += counts.Created
		result.Updated += counts.Updated
		result.Deleted += counts.Deleted
		result.Unchanged += counts.Unchanged
	}

	return result, nil
}

// worksWithAnimeはanime_idが解決済みのworksを返す。anime未解決 (anime_idがnil)
// のworkは別表行をまだ紐付けられないため、後続の実行へ繰り延べる。
func worksWithAnime(works []*model.Work) []*model.Work {
	resolved := make([]*model.Work, 0, len(works))
	for _, w := range works {
		if w.AnimeID != nil {
			resolved = append(resolved, w)
		}
	}
	return resolved
}

package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// episodeTitleFillWindowは、しょぼいカレンダーからタイトルを入れる対象にするエピソードの、
// 作成日時を実行時刻からさかのぼる幅 (Railsと同じ7日)。
const episodeTitleFillWindow = 7 * 24 * time.Hour

// SyobocalSubTitleLookupは、しょぼいカレンダーから作品のサブタイトルを引く。結果は
// TID → 話数 → サブタイトルの形で、サブタイトルの無い作品は含まない。
type SyobocalSubTitleLookup interface {
	LookupSubTitles(ctx context.Context, tids []int32) (map[int32]map[int32]string, error)
}

// FillEpisodeTitlesUsecaseは、タイトルが空のエピソードに、しょぼいカレンダーのサブタイトルを
// 入れる。エピソードのタイトルとanimeのタイトルを同じトランザクションで両書きする。
type FillEpisodeTitlesUsecase struct {
	db             *sql.DB
	episodeRepo    *repository.EpisodeRepository
	animeRepo      *repository.AnimeRepository
	subTitleLookup SyobocalSubTitleLookup
}

// NewFillEpisodeTitlesUsecaseはFillEpisodeTitlesUsecaseを生成する。
func NewFillEpisodeTitlesUsecase(
	db *sql.DB,
	episodeRepo *repository.EpisodeRepository,
	animeRepo *repository.AnimeRepository,
	subTitleLookup SyobocalSubTitleLookup,
) *FillEpisodeTitlesUsecase {
	return &FillEpisodeTitlesUsecase{
		db:             db,
		episodeRepo:    episodeRepo,
		animeRepo:      animeRepo,
		subTitleLookup: subTitleLookup,
	}
}

// FillEpisodeTitlesInputは実行時刻を持つ。対象にするエピソードの作成日時の下限は、この
// 時刻からさかのぼって取る。
type FillEpisodeTitlesInput struct {
	Now time.Time
}

// FillEpisodeTitlesOutputは1回の実行の結果を、報告に使える形で返す。
type FillEpisodeTitlesOutput struct {
	FilledCount    int
	FailedEpisodes []FailedEpisode
}

// FailedEpisodeは、タイトルの更新がエラーで失敗したエピソードを表す。
type FailedEpisode struct {
	EpisodeID model.EpisodeID
	Err       error
}

// Executeは対象のエピソードの作品のサブタイトルをまとめて引き、エピソードごとに1トランザクション
// でタイトルを入れる。しょぼいカレンダーへの問い合わせはトランザクションの外で行う。対象の
// 読み込みとしょぼいカレンダーへの問い合わせの失敗はエラーとして返し、エピソードごとの失敗は
// 結果に積んで次へ進む。
func (uc *FillEpisodeTitlesUsecase) Execute(ctx context.Context, input FillEpisodeTitlesInput) (*FillEpisodeTitlesOutput, error) {
	targets, err := uc.episodeRepo.ListForTitleFill(ctx, input.Now.Add(-episodeTitleFillWindow))
	if err != nil {
		return nil, fmt.Errorf("タイトルを入れるエピソードの取得に失敗: %w", err)
	}

	output := &FillEpisodeTitlesOutput{}
	if len(targets) == 0 {
		return output, nil
	}

	subTitles, err := uc.subTitleLookup.LookupSubTitles(ctx, scTidsOf(targets))
	if err != nil {
		return nil, fmt.Errorf("しょぼいカレンダーからのサブタイトルの取得に失敗: %w", err)
	}

	for _, target := range targets {
		title := subTitles[target.ScTid][int32(*target.Episode.RawNumber)]
		if title == "" {
			continue
		}

		filled, err := uc.fillTitle(ctx, target, title, input.Now.Add(-episodeTitleFillWindow))
		if err != nil {
			output.FailedEpisodes = append(output.FailedEpisodes, FailedEpisode{EpisodeID: target.Episode.ID, Err: err})
			continue
		}
		if filled {
			output.FilledCount++
		}
	}

	return output, nil
}

// fillTitleは現在も問い合わせ時の作品・話数・TIDと対象条件に一致するエピソードにタイトルを
// 入れ、更新した行が現在対応するanimeへ両書きする。条件が変わっていれば書かずにfalseを返す。
// Railsのupdate_columnと同じく、DB活動の記録と作品のtouchは行わない。
func (uc *FillEpisodeTitlesUsecase) fillTitle(ctx context.Context, target repository.DBEpisodeTitleFillTarget, title string, createdFrom time.Time) (bool, error) {
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("トランザクションの開始に失敗しました: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	episode, err := uc.episodeRepo.WithTx(tx).UpdateEmptyTitle(ctx, target, title, createdFrom)
	if err != nil {
		return false, fmt.Errorf("エピソードのタイトルの更新に失敗しました: %w", err)
	}
	if episode == nil {
		return false, nil
	}

	if episode.AnimeID != nil {
		if err := uc.animeRepo.WithTx(tx).UpdateTitle(ctx, *episode.AnimeID, title); err != nil {
			return false, fmt.Errorf("animeのタイトルの更新に失敗しました: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("トランザクションのコミットに失敗しました: %w", err)
	}
	return true, nil
}

// scTidsOfは対象のエピソードの作品のTIDを、重複を除いて昇順に返す。
func scTidsOf(targets []repository.DBEpisodeTitleFillTarget) []int32 {
	seen := make(map[int32]bool, len(targets))
	tids := make([]int32, 0, len(targets))
	for _, target := range targets {
		if seen[target.ScTid] {
			continue
		}
		seen[target.ScTid] = true
		tids = append(tids, target.ScTid)
	}
	sort.Slice(tids, func(i, j int) bool { return tids[i] < tids[j] })
	return tids
}

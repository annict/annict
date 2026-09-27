package usecase

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// WorkSatelliteReposは作品 作成 / 更新UseCaseがworks / animes /
// anime_classificationsと並んで両書きする6つの別表リポジトリを束ねる。1つの値で渡すことで
// CreateWorkUsecase / UpdateWorkUsecaseのコンストラクタに6個の引数が増えるのを防ぐ。WithTx
// はトランザクションに束ねたコピーを返し、両書きがworks / animeの書き込みと同じ
// トランザクションに参加できるようにする。
type WorkSatelliteRepos struct {
	ExternalID      *repository.AnimeExternalIDRepository
	Link            *repository.AnimeLinkRepository
	OfficialAccount *repository.AnimeOfficialAccountRepository
	Hashtag         *repository.AnimeHashtagRepository
	Season          *repository.AnimeSeasonRepository
	Event           *repository.AnimeEventRepository
}

// WithTxは各リポジトリを指定トランザクションに束ねた束のコピーを返す。
func (r WorkSatelliteRepos) WithTx(tx *sql.Tx) WorkSatelliteRepos {
	return WorkSatelliteRepos{
		ExternalID:      r.ExternalID.WithTx(tx),
		Link:            r.Link.WithTx(tx),
		OfficialAccount: r.OfficialAccount.WithTx(tx),
		Hashtag:         r.Hashtag.WithTx(tx),
		Season:          r.Season.WithTx(tx),
		Event:           r.Event.WithTx(tx),
	}
}

// workSatelliteExistingは1つのanimeの既存別表行を束ねる。planWorkSatellitesが突合
// できるようトランザクションの前に読み込む (トランザクション内は永続化のみとする書き込み
// UseCaseのルールに従う)。作成経路ではanimeが新規でまだ行を持たないため、ゼロ値
// (全スライスnil) のままにする。
type workSatelliteExisting struct {
	externalIDs      []*model.AnimeExternalID
	links            []*model.AnimeLink
	officialAccounts []*model.AnimeOfficialAccount
	hashtags         []*model.AnimeHashtag
	seasons          []*model.AnimeSeason
	events           []*model.AnimeEvent
}

// workSatellitePlansは1つのworkについてのテーブルごとのリコンサイル計画を束ねる。
// planWorkSatellitesが生成し、applyWorkSatellitePlansがトランザクション内で適用する。
type workSatellitePlans struct {
	externalIDs      satelliteReconcilePlan[repository.CreateAnimeExternalIDParams, *model.AnimeExternalID]
	links            satelliteReconcilePlan[repository.CreateAnimeLinkParams, *model.AnimeLink]
	officialAccounts satelliteReconcilePlan[repository.CreateAnimeOfficialAccountParams, *model.AnimeOfficialAccount]
	hashtags         satelliteReconcilePlan[repository.CreateAnimeHashtagParams, *model.AnimeHashtag]
	seasons          satelliteReconcilePlan[repository.CreateAnimeSeasonParams, *model.AnimeSeason]
	events           satelliteReconcilePlan[repository.CreateAnimeEventParams, *model.AnimeEvent]
}

// readWorkSatelliteExistingは指定animeの既存別表行を6テーブルすべてから読み込む。
// 更新の両書きは、計画をトランザクション外で組み立てられるよう (書き込みUseCaseのルール)
// トランザクションを開く前に呼ぶ。作成経路は新規animeに行が無いためスキップする。
func readWorkSatelliteExisting(ctx context.Context, repos WorkSatelliteRepos, animeID model.AnimeID) (workSatelliteExisting, error) {
	animeIDs := []model.AnimeID{animeID}

	externalIDs, err := repos.ExternalID.ListByAnimeIDs(ctx, animeIDs)
	if err != nil {
		return workSatelliteExisting{}, fmt.Errorf("既存anime_external_idsの取得に失敗: %w", err)
	}
	links, err := repos.Link.ListByAnimeIDs(ctx, animeIDs)
	if err != nil {
		return workSatelliteExisting{}, fmt.Errorf("既存anime_linksの取得に失敗: %w", err)
	}
	officialAccounts, err := repos.OfficialAccount.ListByAnimeIDs(ctx, animeIDs)
	if err != nil {
		return workSatelliteExisting{}, fmt.Errorf("既存anime_official_accountsの取得に失敗: %w", err)
	}
	hashtags, err := repos.Hashtag.ListByAnimeIDs(ctx, animeIDs)
	if err != nil {
		return workSatelliteExisting{}, fmt.Errorf("既存anime_hashtagsの取得に失敗: %w", err)
	}
	seasons, err := repos.Season.ListByAnimeIDs(ctx, animeIDs)
	if err != nil {
		return workSatelliteExisting{}, fmt.Errorf("既存anime_seasonsの取得に失敗: %w", err)
	}
	events, err := repos.Event.ListByAnimeIDs(ctx, animeIDs)
	if err != nil {
		return workSatelliteExisting{}, fmt.Errorf("既存anime_eventsの取得に失敗: %w", err)
	}

	return workSatelliteExisting{
		externalIDs:      externalIDs,
		links:            links,
		officialAccounts: officialAccounts,
		hashtags:         hashtags,
		seasons:          seasons,
		events:           events,
	}, nil
}

// planWorkSatellitesは1つのworkが持つべき別表行を既存行と突合する。フェーズ2の
// バッチ同期と同じテーブルごとのplan* ヘルパーを再利用して写像を単一の正本に保つ (写像が
// ドリフトすると不変条件テストの後続同期がUnchanged以外を返して顕在化する)。workはAnimeID
// を既に持っている必要がある (作成は作りたてのanime、更新はマッピング済みanime)。AnimeIDが
// nilのworkは行を導出しない。作成では既存行が空のため、全計画が作成のみになる。
func planWorkSatellites(work *model.Work, existing workSatelliteExisting) workSatellitePlans {
	works := []*model.Work{work}
	return workSatellitePlans{
		externalIDs:      planAnimeExternalIDs(works, existing.externalIDs),
		links:            planAnimeLinks(works, existing.links),
		officialAccounts: planAnimeOfficialAccounts(works, existing.officialAccounts),
		hashtags:         planAnimeHashtags(works, existing.hashtags),
		seasons:          planAnimeSeasons(works, existing.seasons),
		events:           planAnimeEvents(works, existing.events),
	}
}

// applyWorkSatellitePlansは6つの別表リコンサイル計画を、渡されたtx束ねリポジトリ
// 経由で永続化する。呼び出し元が作成 / 更新トランザクションに束ねたrepos
// (WorkSatelliteRepos.WithTx) を渡すため、別表の書き込みはworks / animeの書き込みと原子的に
// コミットされる。
func applyWorkSatellitePlans(ctx context.Context, repos WorkSatelliteRepos, plans workSatellitePlans) error {
	if err := applySatellitePlan(ctx, plans.externalIDs,
		func(ctx context.Context, d repository.CreateAnimeExternalIDParams) error {
			_, err := repos.ExternalID.Create(ctx, d)
			return err
		},
		func(ctx context.Context, d repository.CreateAnimeExternalIDParams, e *model.AnimeExternalID) error {
			return repos.ExternalID.Update(ctx, repository.UpdateAnimeExternalIDParams{ID: e.ID, ExternalID: d.ExternalID})
		},
		func(ctx context.Context, e *model.AnimeExternalID) error {
			return repos.ExternalID.Delete(ctx, e.ID)
		},
	); err != nil {
		return fmt.Errorf("anime_external_idsの両書きに失敗: %w", err)
	}

	if err := applySatellitePlan(ctx, plans.links,
		func(ctx context.Context, d repository.CreateAnimeLinkParams) error {
			_, err := repos.Link.Create(ctx, d)
			return err
		},
		func(ctx context.Context, d repository.CreateAnimeLinkParams, e *model.AnimeLink) error {
			return repos.Link.Update(ctx, repository.UpdateAnimeLinkParams{ID: e.ID, URL: d.URL})
		},
		func(ctx context.Context, e *model.AnimeLink) error {
			return repos.Link.Delete(ctx, e.ID)
		},
	); err != nil {
		return fmt.Errorf("anime_linksの両書きに失敗: %w", err)
	}

	if err := applySatellitePlan(ctx, plans.officialAccounts,
		func(ctx context.Context, d repository.CreateAnimeOfficialAccountParams) error {
			_, err := repos.OfficialAccount.Create(ctx, d)
			return err
		},
		func(ctx context.Context, d repository.CreateAnimeOfficialAccountParams, e *model.AnimeOfficialAccount) error {
			return repos.OfficialAccount.Update(ctx, repository.UpdateAnimeOfficialAccountParams{ID: e.ID, Account: d.Account})
		},
		func(ctx context.Context, e *model.AnimeOfficialAccount) error {
			return repos.OfficialAccount.Delete(ctx, e.ID)
		},
	); err != nil {
		return fmt.Errorf("anime_official_accountsの両書きに失敗: %w", err)
	}

	// anime_hashtags / anime_seasonsは更新パスを持たない (値の変更は削除 + 作成。
	// それぞれのplan* ヘルパーを参照) ためupdate funcにnilを渡す。万一planが更新を
	// 持った場合、applySatellitePlanはnilによるpanicではなくエラーを返す。
	if err := applySatellitePlan(ctx, plans.hashtags,
		func(ctx context.Context, d repository.CreateAnimeHashtagParams) error {
			_, err := repos.Hashtag.Create(ctx, d)
			return err
		},
		nil,
		func(ctx context.Context, e *model.AnimeHashtag) error {
			return repos.Hashtag.Delete(ctx, e.ID)
		},
	); err != nil {
		return fmt.Errorf("anime_hashtagsの両書きに失敗: %w", err)
	}

	if err := applySatellitePlan(ctx, plans.seasons,
		func(ctx context.Context, d repository.CreateAnimeSeasonParams) error {
			_, err := repos.Season.Create(ctx, d)
			return err
		},
		nil,
		func(ctx context.Context, e *model.AnimeSeason) error {
			return repos.Season.Delete(ctx, e.ID)
		},
	); err != nil {
		return fmt.Errorf("anime_seasonsの両書きに失敗: %w", err)
	}

	if err := applySatellitePlan(ctx, plans.events,
		func(ctx context.Context, d repository.CreateAnimeEventParams) error {
			_, err := repos.Event.Create(ctx, d)
			return err
		},
		func(ctx context.Context, d repository.CreateAnimeEventParams, e *model.AnimeEvent) error {
			return repos.Event.Update(ctx, repository.UpdateAnimeEventParams{ID: e.ID, StartedOn: d.StartedOn, EndedOn: d.EndedOn})
		},
		func(ctx context.Context, e *model.AnimeEvent) error {
			return repos.Event.Delete(ctx, e.ID)
		},
	); err != nil {
		return fmt.Errorf("anime_eventsの両書きに失敗: %w", err)
	}

	return nil
}

// applySatellitePlanは単一の別表のリコンサイル計画を、呼び出し元の作成 / 更新 / 削除
// 関数 (それぞれtx束ねリポジトリを閉じ込める) 経由で永続化する。削除を作成より先に走らせる
// ことで、自然キーの変更 (旧削除 + 新作成) が、同一スロットに2行を許さない部分UNIQUE
// インデックス (anime_seasons) の下でも安全になる。削除先行は他テーブルでも無害で、削除される
// キーが作成されるキーになることはない。update funcは更新を持たないテーブル (anime_hashtags /
// anime_seasons) ではnilで、万一更新が紛れ込んでもnil panicではなくエラーになる。
func applySatellitePlan[D any, E any](
	ctx context.Context,
	plan satelliteReconcilePlan[D, E],
	create func(context.Context, D) error,
	update func(context.Context, D, E) error,
	del func(context.Context, E) error,
) error {
	for _, existing := range plan.deletes {
		if err := del(ctx, existing); err != nil {
			return err
		}
	}
	for _, desired := range plan.creates {
		if err := create(ctx, desired); err != nil {
			return err
		}
	}
	for _, u := range plan.updates {
		if update == nil {
			return fmt.Errorf("この別表は更新をサポートしないが更新計画が生成された")
		}
		if err := update(ctx, u.desired, u.existing); err != nil {
			return err
		}
	}
	return nil
}

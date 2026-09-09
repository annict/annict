package usecase

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// WorkSatelliteRepos bundles the six satellite-table repositories the work create /
// update usecases dual-write to alongside works / animes / anime_classifications. Passing
// them as one value keeps the CreateWorkUsecase / UpdateWorkUsecase constructors from
// growing six extra parameters; WithTx returns a tx-bound copy so the dual-write joins the
// same transaction as the works / anime writes.
//
// [Ja] WorkSatelliteRepos は作品 作成 / 更新 UseCase が works / animes /
// anime_classifications と並んで両書きする 6 つの別表リポジトリを束ねる。1 つの値で渡すことで
// CreateWorkUsecase / UpdateWorkUsecase のコンストラクタに 6 個の引数が増えるのを防ぐ。WithTx
// はトランザクションに束ねたコピーを返し、両書きが works / anime の書き込みと同じ
// トランザクションに参加できるようにする。
type WorkSatelliteRepos struct {
	ExternalID      *repository.AnimeExternalIDRepository
	Link            *repository.AnimeLinkRepository
	OfficialAccount *repository.AnimeOfficialAccountRepository
	Hashtag         *repository.AnimeHashtagRepository
	Season          *repository.AnimeSeasonRepository
	Event           *repository.AnimeEventRepository
}

// WithTx returns a copy of the bundle with every repository bound to the given transaction.
//
// [Ja] WithTx は各リポジトリを指定トランザクションに束ねた束のコピーを返す。
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

// workSatelliteExisting bundles one anime's existing satellite rows, read before the
// transaction so planWorkSatellites can diff against them (following the write-usecase
// rule that the transaction body performs persistence only). It is left zero-valued (all
// nil slices) for the create path, where the anime is brand new and owns no rows yet.
//
// [Ja] workSatelliteExisting は 1 つの anime の既存別表行を束ねる。planWorkSatellites が突合
// できるようトランザクションの前に読み込む (トランザクション内は永続化のみとする書き込み
// UseCase のルールに従う)。作成経路では anime が新規でまだ行を持たないため、ゼロ値
// (全スライス nil) のままにする。
type workSatelliteExisting struct {
	externalIDs      []*model.AnimeExternalID
	links            []*model.AnimeLink
	officialAccounts []*model.AnimeOfficialAccount
	hashtags         []*model.AnimeHashtag
	seasons          []*model.AnimeSeason
	events           []*model.AnimeEvent
}

// workSatellitePlans bundles the per-table reconcile plans for one work, produced by
// planWorkSatellites and applied by applyWorkSatellitePlans inside the transaction.
//
// [Ja] workSatellitePlans は 1 つの work についてのテーブルごとのリコンサイル計画を束ねる。
// planWorkSatellites が生成し、applyWorkSatellitePlans がトランザクション内で適用する。
type workSatellitePlans struct {
	externalIDs      satelliteReconcilePlan[repository.CreateAnimeExternalIDParams, *model.AnimeExternalID]
	links            satelliteReconcilePlan[repository.CreateAnimeLinkParams, *model.AnimeLink]
	officialAccounts satelliteReconcilePlan[repository.CreateAnimeOfficialAccountParams, *model.AnimeOfficialAccount]
	hashtags         satelliteReconcilePlan[repository.CreateAnimeHashtagParams, *model.AnimeHashtag]
	seasons          satelliteReconcilePlan[repository.CreateAnimeSeasonParams, *model.AnimeSeason]
	events           satelliteReconcilePlan[repository.CreateAnimeEventParams, *model.AnimeEvent]
}

// readWorkSatelliteExisting loads the existing satellite rows of the given anime from all
// six tables. The update dual-write calls it before opening its transaction so the plan is
// computed outside the transaction (the write-usecase rule); the create path skips it
// because a brand-new anime has no rows.
//
// [Ja] readWorkSatelliteExisting は指定 anime の既存別表行を 6 テーブルすべてから読み込む。
// 更新の両書きは、計画をトランザクション外で組み立てられるよう (書き込み UseCase のルール)
// トランザクションを開く前に呼ぶ。作成経路は新規 anime に行が無いためスキップする。
func readWorkSatelliteExisting(ctx context.Context, repos WorkSatelliteRepos, animeID model.AnimeID) (workSatelliteExisting, error) {
	animeIDs := []model.AnimeID{animeID}

	externalIDs, err := repos.ExternalID.ListByAnimeIDs(ctx, animeIDs)
	if err != nil {
		return workSatelliteExisting{}, fmt.Errorf("既存 anime_external_ids の取得に失敗: %w", err)
	}
	links, err := repos.Link.ListByAnimeIDs(ctx, animeIDs)
	if err != nil {
		return workSatelliteExisting{}, fmt.Errorf("既存 anime_links の取得に失敗: %w", err)
	}
	officialAccounts, err := repos.OfficialAccount.ListByAnimeIDs(ctx, animeIDs)
	if err != nil {
		return workSatelliteExisting{}, fmt.Errorf("既存 anime_official_accounts の取得に失敗: %w", err)
	}
	hashtags, err := repos.Hashtag.ListByAnimeIDs(ctx, animeIDs)
	if err != nil {
		return workSatelliteExisting{}, fmt.Errorf("既存 anime_hashtags の取得に失敗: %w", err)
	}
	seasons, err := repos.Season.ListByAnimeIDs(ctx, animeIDs)
	if err != nil {
		return workSatelliteExisting{}, fmt.Errorf("既存 anime_seasons の取得に失敗: %w", err)
	}
	events, err := repos.Event.ListByAnimeIDs(ctx, animeIDs)
	if err != nil {
		return workSatelliteExisting{}, fmt.Errorf("既存 anime_events の取得に失敗: %w", err)
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

// planWorkSatellites reconciles the satellite rows a single work should own against the
// existing rows, reusing the same per-table plan* helpers the phase 2 batch sync uses so
// the mapping stays single-sourced (a spurious diff would surface as the invariant test's
// follow-up sync reporting a non-Unchanged result). The work must already carry its
// AnimeID (the just-created anime for create, the mapped anime for update); a work with a
// nil AnimeID derives no rows. For create the existing rows are empty, so every plan is
// all-creates.
//
// [Ja] planWorkSatellites は 1 つの work が持つべき別表行を既存行と突合する。フェーズ 2 の
// バッチ同期と同じテーブルごとの plan* ヘルパーを再利用して写像を単一の正本に保つ (写像が
// ドリフトすると不変条件テストの後続同期が Unchanged 以外を返して顕在化する)。work は AnimeID
// を既に持っている必要がある (作成は作りたての anime、更新はマッピング済み anime)。AnimeID が
// nil の work は行を導出しない。作成では既存行が空のため、全計画が作成のみになる。
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

// applyWorkSatellitePlans persists all six satellite reconcile plans through the given
// tx-bound repositories. The caller passes repos already bound to the create / update
// transaction (WorkSatelliteRepos.WithTx), so the satellite writes commit atomically with
// the works / anime writes.
//
// [Ja] applyWorkSatellitePlans は 6 つの別表リコンサイル計画を、渡された tx 束ねリポジトリ
// 経由で永続化する。呼び出し元が作成 / 更新トランザクションに束ねた repos
// (WorkSatelliteRepos.WithTx) を渡すため、別表の書き込みは works / anime の書き込みと原子的に
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
		return fmt.Errorf("anime_external_ids の両書きに失敗: %w", err)
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
		return fmt.Errorf("anime_links の両書きに失敗: %w", err)
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
		return fmt.Errorf("anime_official_accounts の両書きに失敗: %w", err)
	}

	// anime_hashtags / anime_seasons have no update path (a changed value is a delete +
	// create, see their plan* helpers), so a nil update func is passed; applySatellitePlan
	// then errors if a plan ever unexpectedly carries updates instead of nil-panicking.
	//
	// [Ja] anime_hashtags / anime_seasons は更新パスを持たない (値の変更は削除 + 作成。
	// それぞれの plan* ヘルパーを参照) ため update func に nil を渡す。万一 plan が更新を
	// 持った場合、applySatellitePlan は nil による panic ではなくエラーを返す。
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
		return fmt.Errorf("anime_hashtags の両書きに失敗: %w", err)
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
		return fmt.Errorf("anime_seasons の両書きに失敗: %w", err)
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
		return fmt.Errorf("anime_events の両書きに失敗: %w", err)
	}

	return nil
}

// applySatellitePlan persists one satellite table's reconcile plan through the caller's
// create / update / delete functions (each closing over a tx-bound repository). Deletes run
// before creates so a natural-key change (delete-old + create-new) is safe even under a
// partial UNIQUE index that forbids two rows in the same slot (anime_seasons); deletes-first
// is harmless for the other tables because a deleted key is never also a created key. The
// update func is nil for tables whose plan never carries updates (anime_hashtags /
// anime_seasons); an update slipping through then errors rather than nil-panicking.
//
// [Ja] applySatellitePlan は単一の別表のリコンサイル計画を、呼び出し元の作成 / 更新 / 削除
// 関数 (それぞれ tx 束ねリポジトリを閉じ込める) 経由で永続化する。削除を作成より先に走らせる
// ことで、自然キーの変更 (旧削除 + 新作成) が、同一スロットに 2 行を許さない部分 UNIQUE
// インデックス (anime_seasons) の下でも安全になる。削除先行は他テーブルでも無害で、削除される
// キーが作成されるキーになることはない。update func は更新を持たないテーブル (anime_hashtags /
// anime_seasons) では nil で、万一更新が紛れ込んでも nil panic ではなくエラーになる。
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

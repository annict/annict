package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// APIAnimeは、REST API V2で返す作品かエピソード1件分の情報
type APIAnime struct {
	Anime            *model.Anime
	Classification   *model.AnimeClassification
	Seasons          []*model.AnimeSeason
	Events           []*model.AnimeEvent
	Links            []*model.AnimeLink
	ExternalIDs      []*model.AnimeExternalID
	OfficialAccounts []*model.AnimeOfficialAccount
	Hashtags         []*model.AnimeHashtag
	LegacyWorkID     *model.WorkID
	LegacyEpisodeID  *model.EpisodeID

	// EpisodesCountは、作品の公開中のエピソードの数。作品のときだけ使う
	EpisodesCount int64

	// Work・PrevEpisode・NextEpisodeは、エピソードの親の作品と、同じ作品の公開中の前後のエピソード。
	// エピソードのときだけ使う
	Work        *model.Anime
	PrevEpisode *model.Anime
	NextEpisode *model.Anime
}

// apiAnimeLoaderは、作品・エピソードの関連データを、要素ごとにではなくIDの一覧でまとめて引く
type apiAnimeLoader struct {
	animeRepo               *repository.AnimeRepository
	animeClassificationRepo *repository.AnimeClassificationRepository
	animeLegacyIDRepo       *repository.AnimeLegacyIDRepository
	satelliteRepos          WorkSatelliteRepos
}

// loadは、animesの並びのまま、それぞれの関連データを組み立てて返す
func (l *apiAnimeLoader) load(ctx context.Context, animes []*model.Anime) ([]*APIAnime, error) {
	ids := make([]model.AnimeID, len(animes))
	for i, anime := range animes {
		ids[i] = anime.ID
	}

	classifications, err := l.animeClassificationRepo.ListByAnimeIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("分類の取得に失敗: %w", err)
	}
	classificationByID := make(map[model.AnimeID]*model.AnimeClassification, len(classifications))
	for _, c := range classifications {
		classificationByID[c.AnimeID] = c
	}

	results := make([]*APIAnime, len(animes))
	var workIDs, parentIDs []model.AnimeID
	for i, anime := range animes {
		c, ok := classificationByID[anime.ID]
		if !ok {
			return nil, fmt.Errorf("anime %d の分類がありません", anime.ID)
		}
		results[i] = &APIAnime{Anime: anime, Classification: c}
		if c.ParentAnimeID != nil {
			parentIDs = append(parentIDs, *c.ParentAnimeID)
		} else {
			workIDs = append(workIDs, anime.ID)
		}
	}

	if err := l.loadSatellites(ctx, ids, results); err != nil {
		return nil, err
	}
	if err := l.loadLegacyIDs(ctx, ids, results); err != nil {
		return nil, err
	}

	episodesCounts, err := l.animeClassificationRepo.CountPublishedEpisodesByParentAnimeIDs(ctx, workIDs)
	if err != nil {
		return nil, fmt.Errorf("エピソードの数の取得に失敗: %w", err)
	}
	parents, err := l.animeRepo.ListByIDs(ctx, parentIDs)
	if err != nil {
		return nil, fmt.Errorf("親の作品の取得に失敗: %w", err)
	}
	parentByID := make(map[model.AnimeID]*model.Anime, len(parents))
	for _, p := range parents {
		parentByID[p.ID] = p
	}

	var episodes []*APIAnime
	for _, r := range results {
		parentID := r.Classification.ParentAnimeID
		if parentID == nil {
			r.EpisodesCount = episodesCounts[r.Anime.ID]
			continue
		}
		parent, ok := parentByID[*parentID]
		if !ok {
			return nil, fmt.Errorf("エピソード %d の親の作品 %d がありません", r.Anime.ID, *parentID)
		}
		// 親の非公開化・削除は子のstatusを変えないため、親の状態でも子の可視性を判定する
		if parent.Status != model.AnimeStatusPublished {
			return nil, &model.AppError{
				Code:     model.AppErrCodeResourceNotFound,
				UserMsg:  i18n.T(ctx, "error_not_found_message"),
				Metadata: map[string]string{"anime_id": r.Anime.ID.String()},
			}
		}
		r.Work = parent
		episodes = append(episodes, r)
	}
	if err := l.loadAdjacentEpisodes(ctx, parentIDs, episodes); err != nil {
		return nil, err
	}

	return results, nil
}

// loadSatellitesは、季節・予定・リンク・外部のID・公式アカウント・ハッシュタグを組み立てる
func (l *apiAnimeLoader) loadSatellites(ctx context.Context, ids []model.AnimeID, results []*APIAnime) error {
	seasons, err := l.satelliteRepos.Season.ListByAnimeIDs(ctx, ids)
	if err != nil {
		return fmt.Errorf("季節の取得に失敗: %w", err)
	}
	events, err := l.satelliteRepos.Event.ListByAnimeIDs(ctx, ids)
	if err != nil {
		return fmt.Errorf("予定の取得に失敗: %w", err)
	}
	links, err := l.satelliteRepos.Link.ListByAnimeIDs(ctx, ids)
	if err != nil {
		return fmt.Errorf("リンクの取得に失敗: %w", err)
	}
	externalIDs, err := l.satelliteRepos.ExternalID.ListByAnimeIDs(ctx, ids)
	if err != nil {
		return fmt.Errorf("外部のIDの取得に失敗: %w", err)
	}
	officialAccounts, err := l.satelliteRepos.OfficialAccount.ListByAnimeIDs(ctx, ids)
	if err != nil {
		return fmt.Errorf("公式アカウントの取得に失敗: %w", err)
	}
	hashtags, err := l.satelliteRepos.Hashtag.ListByAnimeIDs(ctx, ids)
	if err != nil {
		return fmt.Errorf("ハッシュタグの取得に失敗: %w", err)
	}

	seasonsByID := groupByAnimeID(seasons, func(s *model.AnimeSeason) model.AnimeID { return s.AnimeID })
	eventsByID := groupByAnimeID(events, func(e *model.AnimeEvent) model.AnimeID { return e.AnimeID })
	linksByID := groupByAnimeID(links, func(l *model.AnimeLink) model.AnimeID { return l.AnimeID })
	externalIDsByID := groupByAnimeID(externalIDs, func(e *model.AnimeExternalID) model.AnimeID { return e.AnimeID })
	officialAccountsByID := groupByAnimeID(officialAccounts, func(a *model.AnimeOfficialAccount) model.AnimeID { return a.AnimeID })
	hashtagsByID := groupByAnimeID(hashtags, func(h *model.AnimeHashtag) model.AnimeID { return h.AnimeID })

	for _, r := range results {
		id := r.Anime.ID
		r.Seasons = seasonsByID[id]
		r.Events = eventsByID[id]
		r.Links = linksByID[id]
		r.ExternalIDs = externalIDsByID[id]
		r.OfficialAccounts = officialAccountsByID[id]
		r.Hashtags = hashtagsByID[id]
	}
	return nil
}

// loadLegacyIDsは、REST API V1・GraphQL APIでの作品・エピソードのIDを組み立てる
func (l *apiAnimeLoader) loadLegacyIDs(ctx context.Context, ids []model.AnimeID, results []*APIAnime) error {
	legacyIDs, err := l.animeLegacyIDRepo.ListByAnimeIDs(ctx, ids)
	if err != nil {
		return fmt.Errorf("旧IDの取得に失敗: %w", err)
	}
	legacyIDsByID := groupByAnimeID(legacyIDs, func(l *model.AnimeLegacyID) model.AnimeID { return l.AnimeID })

	for _, r := range results {
		for _, legacyID := range legacyIDsByID[r.Anime.ID] {
			if legacyID.WorkID != nil {
				r.LegacyWorkID = legacyID.WorkID
			}
			if legacyID.EpisodeID != nil {
				r.LegacyEpisodeID = legacyID.EpisodeID
			}
		}
	}
	return nil
}

// loadAdjacentEpisodesは、エピソードと同じ作品の公開中の前後のエピソードを組み立てる。
// エピソードの数によらず、前後のIDとそのanimeをそれぞれ1回でまとめて引く
func (l *apiAnimeLoader) loadAdjacentEpisodes(ctx context.Context, parentIDs []model.AnimeID, episodes []*APIAnime) error {
	if len(episodes) == 0 {
		return nil
	}

	ids := make([]model.AnimeID, len(episodes))
	for i, e := range episodes {
		ids[i] = e.Anime.ID
	}
	neighbors, err := l.animeClassificationRepo.ListPublishedEpisodeNeighbors(ctx, parentIDs, ids)
	if err != nil {
		return fmt.Errorf("前後のエピソードの取得に失敗: %w", err)
	}

	neighborsByID := make(map[model.AnimeID]*model.EpisodeNeighbors, len(neighbors))
	var neighborIDs []model.AnimeID
	for _, n := range neighbors {
		neighborsByID[n.AnimeID] = n
		if n.PrevAnimeID != nil {
			neighborIDs = append(neighborIDs, *n.PrevAnimeID)
		}
		if n.NextAnimeID != nil {
			neighborIDs = append(neighborIDs, *n.NextAnimeID)
		}
	}
	neighborAnimes, err := l.animeRepo.ListByIDs(ctx, neighborIDs)
	if err != nil {
		return fmt.Errorf("前後のエピソードのanimeの取得に失敗: %w", err)
	}
	animeByID := make(map[model.AnimeID]*model.Anime, len(neighborAnimes))
	for _, a := range neighborAnimes {
		animeByID[a.ID] = a
	}

	for _, e := range episodes {
		n, ok := neighborsByID[e.Anime.ID]
		if !ok {
			// 公開中のエピソードだけを並べるため、返せるエピソードなら必ず前後を求められる
			return fmt.Errorf("エピソード %d の前後を求められません", e.Anime.ID)
		}
		if n.PrevAnimeID != nil {
			e.PrevEpisode = animeByID[*n.PrevAnimeID]
		}
		if n.NextAnimeID != nil {
			e.NextEpisode = animeByID[*n.NextAnimeID]
		}
	}
	return nil
}

// groupByAnimeIDは、要素をanime IDごとに、元の並びのまま分ける
func groupByAnimeID[T any](items []*T, animeID func(*T) model.AnimeID) map[model.AnimeID][]*T {
	grouped := make(map[model.AnimeID][]*T)
	for _, item := range items {
		id := animeID(item)
		grouped[id] = append(grouped[id], item)
	}
	return grouped
}

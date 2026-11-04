package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/validator"
)

// GetAPILegacyIDsUsecaseは、REST API V2で旧作品・旧エピソードのIDから `anime_id` を引くユースケース
type GetAPILegacyIDsUsecase struct {
	validator               *validator.APILegacyIDIndexValidator
	animeRepo               *repository.AnimeRepository
	animeClassificationRepo *repository.AnimeClassificationRepository
	animeRedirectRepo       *repository.AnimeRedirectRepository
	animeLegacyIDRepo       *repository.AnimeLegacyIDRepository
}

func NewGetAPILegacyIDsUsecase(
	validator *validator.APILegacyIDIndexValidator,
	animeRepo *repository.AnimeRepository,
	animeClassificationRepo *repository.AnimeClassificationRepository,
	animeRedirectRepo *repository.AnimeRedirectRepository,
	animeLegacyIDRepo *repository.AnimeLegacyIDRepository,
) *GetAPILegacyIDsUsecase {
	return &GetAPILegacyIDsUsecase{
		validator:               validator,
		animeRepo:               animeRepo,
		animeClassificationRepo: animeClassificationRepo,
		animeRedirectRepo:       animeRedirectRepo,
		animeLegacyIDRepo:       animeLegacyIDRepo,
	}
}

// GetAPILegacyIDsInputは、`anime_id` を引く旧IDの一覧
type GetAPILegacyIDsInput struct {
	WorkIDs    []model.WorkID
	EpisodeIDs []model.EpisodeID
}

// APILegacyWorkIDは、旧作品のIDと、いま返せるanimeのID (返せなければnil) の組
type APILegacyWorkID struct {
	WorkID  model.WorkID
	AnimeID *model.AnimeID
}

// APILegacyEpisodeIDは、旧エピソードのIDと、いま返せるanimeのID (返せなければnil) の組
type APILegacyEpisodeID struct {
	EpisodeID model.EpisodeID
	AnimeID   *model.AnimeID
}

// GetAPILegacyIDsOutputは、旧IDごとの結果。要求の順に、同じ旧IDは1件にまとめて並べる
type GetAPILegacyIDsOutput struct {
	Works    []APILegacyWorkID
	Episodes []APILegacyEpisodeID
}

// Executeは、旧IDごとに、getAnimeで取得できるanimeのIDを返す。
// 統合したanimeは統合先のIDを返し、写した先の無い旧IDと、返せないanime (非公開・削除済みと、親の作品が公開中でないエピソード) はnilにする
func (uc *GetAPILegacyIDsUsecase) Execute(ctx context.Context, input GetAPILegacyIDsInput) (*GetAPILegacyIDsOutput, error) {
	if err := uc.validator.Validate(ctx, validator.APILegacyIDIndexValidatorInput{
		WorkIDs:    input.WorkIDs,
		EpisodeIDs: input.EpisodeIDs,
	}); err != nil {
		return nil, err
	}

	workIDs := uniqueIDs(input.WorkIDs)
	episodeIDs := uniqueIDs(input.EpisodeIDs)
	workLegacyIDs, err := uc.animeLegacyIDRepo.ListByWorkIDs(ctx, workIDs)
	if err != nil {
		return nil, fmt.Errorf("旧作品のanimeのIDの取得に失敗: %w", err)
	}
	episodeLegacyIDs, err := uc.animeLegacyIDRepo.ListByEpisodeIDs(ctx, episodeIDs)
	if err != nil {
		return nil, fmt.Errorf("旧エピソードのanimeのIDの取得に失敗: %w", err)
	}

	animeIDByWorkID := make(map[model.WorkID]model.AnimeID, len(workLegacyIDs))
	animeIDByEpisodeID := make(map[model.EpisodeID]model.AnimeID, len(episodeLegacyIDs))
	animeIDs := make([]model.AnimeID, 0, len(workLegacyIDs)+len(episodeLegacyIDs))
	for _, l := range workLegacyIDs {
		animeIDByWorkID[*l.WorkID] = l.AnimeID
		animeIDs = append(animeIDs, l.AnimeID)
	}
	for _, l := range episodeLegacyIDs {
		animeIDByEpisodeID[*l.EpisodeID] = l.AnimeID
		animeIDs = append(animeIDs, l.AnimeID)
	}

	visibleIDs, err := uc.resolveVisibleAnimeIDs(ctx, animeIDs)
	if err != nil {
		return nil, err
	}

	output := &GetAPILegacyIDsOutput{
		Works:    make([]APILegacyWorkID, len(workIDs)),
		Episodes: make([]APILegacyEpisodeID, len(episodeIDs)),
	}
	for i, id := range workIDs {
		output.Works[i] = APILegacyWorkID{WorkID: id, AnimeID: lookupVisibleAnimeID(animeIDByWorkID, visibleIDs, id)}
	}
	for i, id := range episodeIDs {
		output.Episodes[i] = APILegacyEpisodeID{EpisodeID: id, AnimeID: lookupVisibleAnimeID(animeIDByEpisodeID, visibleIDs, id)}
	}
	return output, nil
}

// resolveVisibleAnimeIDsは、animeのIDから、getAnimeで取得できるanimeのIDへのmapを返す。
// 統合したanimeは統合先を指し、取得できないanimeはmapに含めない。
// getAnimeと同じく、エピソードは親の作品が公開中のときだけ取得できる
func (uc *GetAPILegacyIDsUsecase) resolveVisibleAnimeIDs(ctx context.Context, ids []model.AnimeID) (map[model.AnimeID]model.AnimeID, error) {
	animes, err := uc.animeRepo.ListByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("animeの取得に失敗: %w", err)
	}

	resolved := make(map[model.AnimeID]model.AnimeID, len(animes))
	var mergedIDs []model.AnimeID
	for _, a := range animes {
		switch a.Status {
		case model.AnimeStatusPublished:
			resolved[a.ID] = a.ID
		case model.AnimeStatusMerged:
			mergedIDs = append(mergedIDs, a.ID)
		}
	}

	// 統合の記録は常に最終の統合先を指すため、1回たどれば足りる
	canonicalIDs, err := uc.animeRedirectRepo.ListCanonicalAnimeIDs(ctx, mergedIDs)
	if err != nil {
		return nil, fmt.Errorf("統合先の取得に失敗: %w", err)
	}
	if err := uc.resolveCanonicalAnimeIDs(ctx, canonicalIDs, resolved); err != nil {
		return nil, err
	}

	hiddenIDs, err := uc.listEpisodeIDsOfHiddenWorks(ctx, resolved)
	if err != nil {
		return nil, err
	}
	for id, visibleID := range resolved {
		if hiddenIDs[visibleID] {
			delete(resolved, id)
		}
	}
	return resolved, nil
}

// resolveCanonicalAnimeIDsは、統合先が公開中の統合したanimeを、統合先を指すようresolvedに加える
func (uc *GetAPILegacyIDsUsecase) resolveCanonicalAnimeIDs(ctx context.Context, canonicalIDs map[model.AnimeID]model.AnimeID, resolved map[model.AnimeID]model.AnimeID) error {
	ids := make([]model.AnimeID, 0, len(canonicalIDs))
	for _, id := range canonicalIDs {
		ids = append(ids, id)
	}
	canonicals, err := uc.animeRepo.ListByIDs(ctx, ids)
	if err != nil {
		return fmt.Errorf("統合先のanimeの取得に失敗: %w", err)
	}

	published := make(map[model.AnimeID]bool, len(canonicals))
	for _, a := range canonicals {
		published[a.ID] = a.Status == model.AnimeStatusPublished
	}
	for oldID, canonicalID := range canonicalIDs {
		if published[canonicalID] {
			resolved[oldID] = canonicalID
		}
	}
	return nil
}

// listEpisodeIDsOfHiddenWorksは、resolvedの行き先のうち、親の作品が公開中でないエピソードのIDを返す。
// 親の非公開化・削除は子のstatusを変えないため、親の状態で確かめる
func (uc *GetAPILegacyIDsUsecase) listEpisodeIDsOfHiddenWorks(ctx context.Context, resolved map[model.AnimeID]model.AnimeID) (map[model.AnimeID]bool, error) {
	ids := make([]model.AnimeID, 0, len(resolved))
	for _, id := range resolved {
		ids = append(ids, id)
	}
	classifications, err := uc.animeClassificationRepo.ListByAnimeIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("分類の取得に失敗: %w", err)
	}
	if len(classifications) != len(uniqueIDs(ids)) {
		// 公開中のanimeには分類がある (作品・エピソードの保存が両書きする)。欠けているのはデータの不整合
		return nil, fmt.Errorf("分類の無いanimeがあります (anime %d件・分類 %d件)", len(uniqueIDs(ids)), len(classifications))
	}

	parentIDByID := make(map[model.AnimeID]model.AnimeID)
	var parentIDs []model.AnimeID
	for _, c := range classifications {
		if c.ParentAnimeID != nil {
			parentIDByID[c.AnimeID] = *c.ParentAnimeID
			parentIDs = append(parentIDs, *c.ParentAnimeID)
		}
	}
	parents, err := uc.animeRepo.ListByIDs(ctx, parentIDs)
	if err != nil {
		return nil, fmt.Errorf("親の作品の取得に失敗: %w", err)
	}
	publishedParents := make(map[model.AnimeID]bool, len(parents))
	for _, p := range parents {
		publishedParents[p.ID] = p.Status == model.AnimeStatusPublished
	}

	hidden := make(map[model.AnimeID]bool)
	for id, parentID := range parentIDByID {
		if !publishedParents[parentID] {
			hidden[id] = true
		}
	}
	return hidden, nil
}

// lookupVisibleAnimeIDは、旧IDから写した先のanimeをたどり、取得できるanimeのIDを返す。取得できなければnilを返す
func lookupVisibleAnimeID[K comparable](animeIDByLegacyID map[K]model.AnimeID, visibleIDs map[model.AnimeID]model.AnimeID, legacyID K) *model.AnimeID {
	animeID, ok := animeIDByLegacyID[legacyID]
	if !ok {
		return nil
	}
	visibleID, ok := visibleIDs[animeID]
	if !ok {
		return nil
	}
	return &visibleID
}

// uniqueIDsは、並びを保ったまま重複を取り除く
func uniqueIDs[T comparable](ids []T) []T {
	seen := make(map[T]bool, len(ids))
	unique := make([]T, 0, len(ids))
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			unique = append(unique, id)
		}
	}
	return unique
}

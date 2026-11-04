// Package animeはREST API V2の作品とエピソードのハンドラーを提供する
package anime

import (
	"encoding/json"
	"fmt"
	"strconv"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/annict/annict/go/internal/apigen"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/usecase"
)

// Handlerは作品とエピソードのAPIを担当するハンドラー
type Handler struct {
	getAPIAnimeUC *usecase.GetAPIAnimeUsecase
}

// NewHandlerは新しいHandlerを作成する
func NewHandler(getAPIAnimeUC *usecase.GetAPIAnimeUsecase) *Handler {
	return &Handler{getAPIAnimeUC: getAPIAnimeUC}
}

// workAnimeBodyとepisodeAnimeBodyは、共通の情報に `kind` と固有の情報を足したAPIの応答の形。
// 生成コードの `apigen.WorkAnime`・`apigen.EpisodeAnime` は共通の情報を展開した別々の型のため、
// 共通の情報を1か所で組み立てられるよう、`apigen.AnimeAttributes` を埋め込んでJSONで同じ形にする
type workAnimeBody struct {
	apigen.AnimeAttributes
	Kind apigen.WorkAnimeKind  `json:"kind"`
	Work apigen.WorkAttributes `json:"work"`
}

type episodeAnimeBody struct {
	apigen.AnimeAttributes
	Kind    apigen.EpisodeAnimeKind  `json:"kind"`
	Episode apigen.EpisodeAttributes `json:"episode"`
}

// ToAPIAnimeは作品かエピソードをAPIの応答の形に変換する
func ToAPIAnime(a *usecase.APIAnime) (apigen.Anime, error) {
	attrs := toAPIAnimeAttributes(a)

	var body any
	switch a.Classification.Kind {
	case model.AnimeClassificationKindWork:
		body = workAnimeBody{
			AnimeAttributes: attrs,
			Kind:            apigen.Work,
			Work: apigen.WorkAttributes{
				Standalone:            a.Classification.Standalone,
				ExpectedEpisodesCount: nullInt32Ptr(a.Classification.ExpectedEpisodesCount.Int32, a.Classification.ExpectedEpisodesCount.Valid),
				EpisodesCount:         a.EpisodesCount,
			},
		}
	case model.AnimeClassificationKindEpisode:
		episode, err := toAPIEpisodeAttributes(a)
		if err != nil {
			return apigen.Anime{}, err
		}
		body = episodeAnimeBody{AnimeAttributes: attrs, Kind: apigen.Episode, Episode: episode}
	default:
		return apigen.Anime{}, fmt.Errorf("anime %d の分類 %q をAPIの形にできません", a.Anime.ID, a.Classification.Kind)
	}

	b, err := json.Marshal(body)
	if err != nil {
		return apigen.Anime{}, fmt.Errorf("anime %d をJSONにできません: %w", a.Anime.ID, err)
	}
	var anime apigen.Anime
	if err := anime.UnmarshalJSON(b); err != nil {
		return apigen.Anime{}, fmt.Errorf("anime %d の応答を組み立てられません: %w", a.Anime.ID, err)
	}
	return anime, nil
}

// toAPIAnimeAttributesは作品とエピソードに共通の情報を組み立てる。
// 値の無い配列も `null` ではなく空の配列で返すよう、長さ0のスライスを作る
func toAPIAnimeAttributes(a *usecase.APIAnime) apigen.AnimeAttributes {
	anime := a.Anime
	attrs := apigen.AnimeAttributes{
		Id:               int64(anime.ID),
		Title:            nullStringPtr(anime.Title.String, anime.Title.Valid),
		TitleKana:        nullStringPtr(anime.TitleKana.String, anime.TitleKana.Valid),
		TitleRo:          nullStringPtr(anime.TitleRo.String, anime.TitleRo.Valid),
		TitleEn:          nullStringPtr(anime.TitleEn.String, anime.TitleEn.Valid),
		TitleAlter:       nullStringPtr(anime.TitleAlter.String, anime.TitleAlter.Valid),
		TitleAlterRo:     nullStringPtr(anime.TitleAlterRo.String, anime.TitleAlterRo.Valid),
		TitleAlterEn:     nullStringPtr(anime.TitleAlterEn.String, anime.TitleAlterEn.Valid),
		TitleAlterOther:  nullStringPtr(anime.TitleAlterOther.String, anime.TitleAlterOther.Valid),
		Synopsis:         nullStringPtr(anime.Synopsis.String, anime.Synopsis.Valid),
		SynopsisEn:       nullStringPtr(anime.SynopsisEn.String, anime.SynopsisEn.Valid),
		SynopsisSource:   nullStringPtr(anime.SynopsisSource.String, anime.SynopsisSource.Valid),
		SynopsisSourceEn: nullStringPtr(anime.SynopsisSourceEn.String, anime.SynopsisSourceEn.Valid),
		Seasons:          make([]apigen.AnimeSeason, 0, len(a.Seasons)),
		Events:           make([]apigen.AnimeEvent, 0, len(a.Events)),
		Links:            make([]apigen.AnimeLink, 0, len(a.Links)),
		ExternalIds:      make([]apigen.AnimeExternalId, 0, len(a.ExternalIDs)),
		OfficialAccounts: make([]apigen.AnimeOfficialAccount, 0, len(a.OfficialAccounts)),
		Hashtags:         make([]apigen.AnimeHashtag, 0, len(a.Hashtags)),
	}
	if anime.Media != "" {
		media := apigen.AnimeMedia(anime.Media)
		attrs.Media = &media
	}
	if anime.ReleaseStatus != "" {
		status := apigen.AnimeReleaseStatus(anime.ReleaseStatus)
		attrs.ReleaseStatus = &status
	}
	if a.LegacyWorkID != nil {
		id := int64(*a.LegacyWorkID)
		attrs.LegacyWorkId = &id
	}
	if a.LegacyEpisodeID != nil {
		id := int64(*a.LegacyEpisodeID)
		attrs.LegacyEpisodeId = &id
	}

	for _, s := range a.Seasons {
		season := apigen.AnimeSeason{Year: s.Year, IsPrimary: s.IsPrimary}
		if s.Name != nil {
			name := apigen.SeasonName(*s.Name)
			season.Name = &name
		}
		attrs.Seasons = append(attrs.Seasons, season)
	}
	for _, e := range a.Events {
		event := apigen.AnimeEvent{
			Kind:          apigen.AnimeEventKind(e.Kind),
			StartedOn:     openapi_types.Date{Time: e.StartedOn},
			Title:         e.Title,
			TitleEn:       e.TitleEn,
			Description:   e.Description,
			DescriptionEn: e.DescriptionEn,
		}
		if e.EndedOn != nil {
			event.EndedOn = &openapi_types.Date{Time: *e.EndedOn}
		}
		attrs.Events = append(attrs.Events, event)
	}
	for _, l := range a.Links {
		attrs.Links = append(attrs.Links, apigen.AnimeLink{
			Kind:     apigen.AnimeLinkKind(l.Kind),
			Language: apigen.AnimeLinkLanguage(l.Language),
			Url:      l.URL,
			Label:    l.Label,
			LabelEn:  l.LabelEn,
		})
	}
	for _, e := range a.ExternalIDs {
		attrs.ExternalIds = append(attrs.ExternalIds, apigen.AnimeExternalId{
			Service:    apigen.AnimeExternalIdService(e.Service),
			ExternalId: e.ExternalID,
		})
	}
	for _, o := range a.OfficialAccounts {
		attrs.OfficialAccounts = append(attrs.OfficialAccounts, apigen.AnimeOfficialAccount{
			Service: apigen.AnimeOfficialAccountService(o.Service),
			Account: o.Account,
			Label:   o.Label,
			LabelEn: o.LabelEn,
		})
	}
	for _, h := range a.Hashtags {
		attrs.Hashtags = append(attrs.Hashtags, apigen.AnimeHashtag{Hashtag: h.Hashtag})
	}
	return attrs
}

// toAPIEpisodeAttributesはエピソードに固有の情報を組み立てる
func toAPIEpisodeAttributes(a *usecase.APIAnime) (apigen.EpisodeAttributes, error) {
	c := a.Classification
	if a.Work == nil {
		return apigen.EpisodeAttributes{}, fmt.Errorf("エピソード %d の親の作品がありません", a.Anime.ID)
	}

	episode := apigen.EpisodeAttributes{
		NumberText:  nullStringPtr(c.NumberText.String, c.NumberText.Valid),
		SortNumber:  c.SortNumber.Int32,
		Work:        toAPIAnimeReference(a.Work),
		PrevEpisode: toAPIAnimeReferencePtr(a.PrevEpisode),
		NextEpisode: toAPIAnimeReferencePtr(a.NextEpisode),
	}
	if c.Number.Valid {
		// 話数は小数 (例 `3.5`) を正確に保つため文字列で持っている
		number, err := strconv.ParseFloat(c.Number.String, 64)
		if err != nil {
			return apigen.EpisodeAttributes{}, fmt.Errorf("エピソード %d の話数 %q を数値にできません: %w", a.Anime.ID, c.Number.String, err)
		}
		episode.Number = &number
	}
	return episode, nil
}

func toAPIAnimeReference(anime *model.Anime) apigen.AnimeReference {
	return apigen.AnimeReference{
		Id:    int64(anime.ID),
		Title: nullStringPtr(anime.Title.String, anime.Title.Valid),
	}
}

func toAPIAnimeReferencePtr(anime *model.Anime) *apigen.AnimeReference {
	if anime == nil {
		return nil
	}
	ref := toAPIAnimeReference(anime)
	return &ref
}

func nullStringPtr(s string, valid bool) *string {
	if !valid {
		return nil
	}
	return &s
}

func nullInt32Ptr(n int32, valid bool) *int32 {
	if !valid {
		return nil
	}
	return &n
}

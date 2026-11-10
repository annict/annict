package usecase

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// animeSnapshotFromRowsは、事前に読んだanime・分類・別表の行から、書き込む前の作品・エピソードの
// スナップショットを組み立てる。分類の行が無いときは分類の列を空のままにする。
func animeSnapshotFromRows(anime *model.Anime, classification *model.AnimeClassification, satellites workSatelliteExisting) model.AnimeSnapshot {
	s := model.AnimeSnapshot{}
	setAnimeSnapshotColumns(&s, repository.UpdateAnimeParams{
		Title:            anime.Title,
		TitleKana:        anime.TitleKana,
		TitleRo:          anime.TitleRo,
		TitleEn:          anime.TitleEn,
		TitleAlter:       anime.TitleAlter,
		TitleAlterRo:     anime.TitleAlterRo,
		TitleAlterEn:     anime.TitleAlterEn,
		TitleAlterOther:  anime.TitleAlterOther,
		Media:            anime.Media,
		ReleaseStatus:    anime.ReleaseStatus,
		Synopsis:         anime.Synopsis,
		SynopsisEn:       anime.SynopsisEn,
		SynopsisSource:   anime.SynopsisSource,
		SynopsisSourceEn: anime.SynopsisSourceEn,
		Status:           anime.Status,
		ArchiveMessage:   anime.ArchiveMessage,
	})
	if classification != nil {
		setAnimeSnapshotClassification(&s, repository.UpdateAnimeClassificationParams{
			Kind:                  classification.Kind,
			ParentAnimeID:         classification.ParentAnimeID,
			Number:                classification.Number,
			NumberText:            classification.NumberText,
			SortNumber:            classification.SortNumber,
			Standalone:            classification.Standalone,
			NumberFormatID:        classification.NumberFormatID,
			EpisodeStartNumber:    classification.EpisodeStartNumber,
			ExpectedEpisodesCount: classification.ExpectedEpisodesCount,
		})
	}
	setAnimeSnapshotSatellites(&s, satellites, workSatellitePlans{})
	return s
}

// readAnimeSnapshotは、animeの分類と別表の行を読み、書き込む前のスナップショットを組み立てる。
// animeの列だけを書き換える書き込み (非公開・再公開・削除) が、トランザクションの前に呼ぶ。
func readAnimeSnapshot(ctx context.Context, classificationRepo *repository.AnimeClassificationRepository, satelliteRepos WorkSatelliteRepos, anime *model.Anime) (model.AnimeSnapshot, error) {
	classification, err := classificationRepo.GetByAnimeID(ctx, anime.ID)
	if err != nil {
		return model.AnimeSnapshot{}, fmt.Errorf("anime_classificationの取得に失敗: %w", err)
	}
	satellites, err := readWorkSatelliteExisting(ctx, satelliteRepos, anime.ID)
	if err != nil {
		return model.AnimeSnapshot{}, err
	}
	return animeSnapshotFromRows(anime, classification, satellites), nil
}

// animeSnapshotFromCreateParamsは、作成するanime・分類・別表の値から作成の版のスナップショットを組み立てる。
// 別表は既存の行が無いため、計画の作成だけが行になる。
func animeSnapshotFromCreateParams(anime repository.CreateAnimeParams, classification repository.CreateAnimeClassificationParams, satellites workSatellitePlans) model.AnimeSnapshot {
	s := model.AnimeSnapshot{}
	setAnimeSnapshotColumns(&s, repository.UpdateAnimeParams{
		Title:            anime.Title,
		TitleKana:        anime.TitleKana,
		TitleRo:          anime.TitleRo,
		TitleEn:          anime.TitleEn,
		TitleAlter:       anime.TitleAlter,
		TitleAlterRo:     anime.TitleAlterRo,
		TitleAlterEn:     anime.TitleAlterEn,
		TitleAlterOther:  anime.TitleAlterOther,
		Media:            anime.Media,
		ReleaseStatus:    anime.ReleaseStatus,
		Synopsis:         anime.Synopsis,
		SynopsisEn:       anime.SynopsisEn,
		SynopsisSource:   anime.SynopsisSource,
		SynopsisSourceEn: anime.SynopsisSourceEn,
		Status:           anime.Status,
		ArchiveMessage:   anime.ArchiveMessage,
	})
	setAnimeSnapshotClassification(&s, repository.UpdateAnimeClassificationParams(classification))
	setAnimeSnapshotSatellites(&s, workSatelliteExisting{}, satellites)
	return s
}

// setAnimeSnapshotColumnsは、animesに書く値をスナップショットのanimesの列に写す。IDは使わない。
// media・release_statusの空文字列はNULLの列を読んだ値のため、未設定として持つ。
func setAnimeSnapshotColumns(s *model.AnimeSnapshot, p repository.UpdateAnimeParams) {
	s.Title = stringPtrFromNullString(p.Title)
	s.TitleKana = stringPtrFromNullString(p.TitleKana)
	s.TitleRo = stringPtrFromNullString(p.TitleRo)
	s.TitleEn = stringPtrFromNullString(p.TitleEn)
	s.TitleAlter = stringPtrFromNullString(p.TitleAlter)
	s.TitleAlterRo = stringPtrFromNullString(p.TitleAlterRo)
	s.TitleAlterEn = stringPtrFromNullString(p.TitleAlterEn)
	s.TitleAlterOther = stringPtrFromNullString(p.TitleAlterOther)
	s.Media = nonEmptyPtr(p.Media)
	s.ReleaseStatus = nonEmptyPtr(p.ReleaseStatus)
	s.Synopsis = stringPtrFromNullString(p.Synopsis)
	s.SynopsisEn = stringPtrFromNullString(p.SynopsisEn)
	s.SynopsisSource = stringPtrFromNullString(p.SynopsisSource)
	s.SynopsisSourceEn = stringPtrFromNullString(p.SynopsisSourceEn)
	s.Status = p.Status
	s.ArchiveMessage = stringPtrFromNullString(p.ArchiveMessage)
}

// setAnimeSnapshotClassificationは、anime_classificationsに書く値をスナップショットの分類の列に写す。
func setAnimeSnapshotClassification(s *model.AnimeSnapshot, p repository.UpdateAnimeClassificationParams) {
	s.Kind = p.Kind
	s.ParentAnimeID = p.ParentAnimeID
	s.Number = stringPtrFromNullString(p.Number)
	s.NumberText = stringPtrFromNullString(p.NumberText)
	s.SortNumber = int32PtrFromNullInt32(p.SortNumber)
	s.Standalone = p.Standalone
	s.NumberFormatID = p.NumberFormatID
	s.EpisodeStartNumber = stringPtrFromNullString(p.EpisodeStartNumber)
	s.ExpectedEpisodesCount = int32PtrFromNullInt32(p.ExpectedEpisodesCount)
}

// setAnimeSnapshotSatellitesは、別表の既存の行にリコンサイル計画を当てた結果 (書き込んだ後の別表の行) を
// スナップショットに写す。計画の削除した行を除き、更新した行は計画が書き換える列だけを差し替え、作成した行は
// 作成のクエリが書かない列をテーブルの既定値 (ラベルは空・並び順は0) で持つ。計画がゼロ値なら既存の行を
// そのまま写す。
func setAnimeSnapshotSatellites(s *model.AnimeSnapshot, existing workSatelliteExisting, plans workSatellitePlans) {
	s.ExternalIDs = nil
	externalIDUpdates := updatesByExisting(plans.externalIDs)
	for _, e := range keptRows(existing.externalIDs, plans.externalIDs.deletes) {
		row := model.AnimeSnapshotExternalID{Service: e.Service, ExternalID: e.ExternalID}
		if d, ok := externalIDUpdates[e]; ok {
			row.ExternalID = d.ExternalID
		}
		s.ExternalIDs = append(s.ExternalIDs, row)
	}
	for _, d := range plans.externalIDs.creates {
		s.ExternalIDs = append(s.ExternalIDs, model.AnimeSnapshotExternalID{Service: d.Service, ExternalID: d.ExternalID})
	}

	s.Links = nil
	linkUpdates := updatesByExisting(plans.links)
	for _, e := range keptRows(existing.links, plans.links.deletes) {
		row := model.AnimeSnapshotLink{Kind: e.Kind, Language: e.Language, URL: e.URL, Label: e.Label, LabelEn: e.LabelEn, SortNumber: e.SortNumber}
		if d, ok := linkUpdates[e]; ok {
			row.URL = d.URL
		}
		s.Links = append(s.Links, row)
	}
	for _, d := range plans.links.creates {
		s.Links = append(s.Links, model.AnimeSnapshotLink{Kind: d.Kind, Language: d.Language, URL: d.URL})
	}

	s.OfficialAccounts = nil
	accountUpdates := updatesByExisting(plans.officialAccounts)
	for _, e := range keptRows(existing.officialAccounts, plans.officialAccounts.deletes) {
		row := model.AnimeSnapshotOfficialAccount{Service: e.Service, Account: e.Account, Label: e.Label, LabelEn: e.LabelEn, SortNumber: e.SortNumber}
		if d, ok := accountUpdates[e]; ok {
			row.Account = d.Account
		}
		s.OfficialAccounts = append(s.OfficialAccounts, row)
	}
	for _, d := range plans.officialAccounts.creates {
		s.OfficialAccounts = append(s.OfficialAccounts, model.AnimeSnapshotOfficialAccount{Service: d.Service, Account: d.Account})
	}

	s.Hashtags = nil
	for _, e := range keptRows(existing.hashtags, plans.hashtags.deletes) {
		s.Hashtags = append(s.Hashtags, model.AnimeSnapshotHashtag{Hashtag: e.Hashtag, SortNumber: e.SortNumber})
	}
	for _, d := range plans.hashtags.creates {
		s.Hashtags = append(s.Hashtags, model.AnimeSnapshotHashtag{Hashtag: d.Hashtag})
	}

	s.Seasons = nil
	for _, e := range keptRows(existing.seasons, plans.seasons.deletes) {
		s.Seasons = append(s.Seasons, model.AnimeSnapshotSeason{Year: e.Year, Name: e.Name, IsPrimary: e.IsPrimary})
	}
	for _, d := range plans.seasons.creates {
		s.Seasons = append(s.Seasons, model.AnimeSnapshotSeason{Year: d.Year, Name: d.Name, IsPrimary: d.IsPrimary})
	}

	s.Events = nil
	eventUpdates := updatesByExisting(plans.events)
	for _, e := range keptRows(existing.events, plans.events.deletes) {
		row := model.AnimeSnapshotEvent{
			Kind:          e.Kind,
			StartedOn:     e.StartedOn,
			EndedOn:       e.EndedOn,
			Title:         e.Title,
			TitleEn:       e.TitleEn,
			Description:   e.Description,
			DescriptionEn: e.DescriptionEn,
			SortNumber:    e.SortNumber,
		}
		if d, ok := eventUpdates[e]; ok {
			row.StartedOn = d.StartedOn
			row.EndedOn = d.EndedOn
		}
		s.Events = append(s.Events, row)
	}
	for _, d := range plans.events.creates {
		s.Events = append(s.Events, model.AnimeSnapshotEvent{Kind: d.Kind, StartedOn: d.StartedOn, EndedOn: d.EndedOn})
	}
}

// keptRowsは既存の行のうち、計画が削除しない行を元の順で返す。
func keptRows[E comparable](existing []E, deletes []E) []E {
	deleted := make(map[E]bool, len(deletes))
	for _, e := range deletes {
		deleted[e] = true
	}
	kept := make([]E, 0, len(existing))
	for _, e := range existing {
		if !deleted[e] {
			kept = append(kept, e)
		}
	}
	return kept
}

// updatesByExistingは計画の更新を、更新する既存の行から引けるようにする。
func updatesByExisting[D any, E comparable](plan satelliteReconcilePlan[D, E]) map[E]D {
	updates := make(map[E]D, len(plan.updates))
	for _, u := range plan.updates {
		updates[u.existing] = u.desired
	}
	return updates
}

func stringPtrFromNullString(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}
	return &v.String
}

func int32PtrFromNullInt32(v sql.NullInt32) *int32 {
	if !v.Valid {
		return nil
	}
	return &v.Int32
}

func nonEmptyPtr[T ~string](v T) *T {
	if v == "" {
		return nil
	}
	return &v
}

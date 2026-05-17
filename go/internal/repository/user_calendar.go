// Package repository はデータアクセス層を提供します
package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// UserCalendarRepository はカレンダーデータの取得を担当します
type UserCalendarRepository struct {
	queries *query.Queries
}

// NewUserCalendarRepository はUserCalendarRepositoryを作成します
func NewUserCalendarRepository(queries *query.Queries) *UserCalendarRepository {
	return &UserCalendarRepository{queries: queries}
}

// GetByUsername はユーザー名からカレンダーデータを取得します
// 読み取り専用の処理のためUsecaseは使用せず、Repositoryで完結します
func (r *UserCalendarRepository) GetByUsername(ctx context.Context, username string, now time.Time) (*model.UserCalendar, error) {
	// 1. ユーザー情報を取得
	userInfo, err := r.queries.GetUserCalendarInfo(ctx, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	// 2. ユーザーIDを取得するためにユーザー情報を再取得
	user, err := r.queries.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	// 3. ライブラリエントリからprogram_idを取得
	libraryEntries, err := r.queries.GetLibraryEntryProgramIDs(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	// 4. program_idと視聴済みエピソードIDを収集
	var programIDs []int64
	watchedEpisodeIDs := make(map[model.EpisodeID]bool)
	for _, entry := range libraryEntries {
		if entry.ProgramID.Valid {
			programIDs = append(programIDs, entry.ProgramID.Int64)
		}
		for _, epID := range entry.WatchedEpisodeIds {
			watchedEpisodeIDs[model.EpisodeID(epID)] = true
		}
	}

	// 5. 放送枠を取得
	var slots []model.CalendarSlot
	if len(programIDs) > 0 {
		// 現在時刻から7日後までを対象
		startedAtTo := now.AddDate(0, 0, 7)
		slotsRows, err := r.queries.GetCalendarSlots(ctx, query.GetCalendarSlotsParams{
			ProgramIds:    programIDs,
			StartedAtFrom: now,
			StartedAtTo:   startedAtTo,
		})
		if err != nil {
			return nil, err
		}

		// 視聴済みエピソードを除外してModelに変換
		for _, row := range slotsRows {
			if row.EpisodeID.Valid && !watchedEpisodeIDs[model.EpisodeID(row.EpisodeID.Int64)] {
				slot := model.CalendarSlot{
					ID:            model.SlotID(row.ID),
					StartedAt:     row.StartedAt,
					WorkID:        model.WorkID(row.WorkID),
					WorkTitle:     row.WorkTitle,
					WorkTitleEn:   row.WorkTitleEn,
					EpisodeNumber: row.EpisodeNumber,
					ChannelName:   row.ChannelName,
				}
				if row.EpisodeID.Valid {
					slot.EpisodeID = model.EpisodeID(row.EpisodeID.Int64)
				}
				if row.EpisodeTitle.Valid {
					slot.EpisodeTitle = row.EpisodeTitle.String
				}
				slots = append(slots, slot)
			}
		}
	}

	// 6. 作品（放送開始日）を取得
	worksRows, err := r.queries.GetCalendarWorks(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	works := make([]model.CalendarWork, 0, len(worksRows))
	for _, row := range worksRows {
		if row.StartedOn.Valid {
			works = append(works, model.CalendarWork{
				ID:        model.WorkID(row.ID),
				Title:     row.Title,
				TitleEn:   row.TitleEn,
				StartedOn: row.StartedOn.Time,
			})
		}
	}

	return &model.UserCalendar{
		Username: userInfo.Username,
		TimeZone: userInfo.TimeZone,
		Locale:   userInfo.Locale,
		Slots:    slots,
		Works:    works,
	}, nil
}

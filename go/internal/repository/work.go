package repository

import (
	"context"

	"github.com/annict/annict/internal/model"
	"github.com/annict/annict/internal/query"
)

// WorkRepository はWork関連のデータアクセスを担当します
type WorkRepository struct {
	queries *query.Queries
}

// NewWorkRepository はWorkRepositoryを作成します
func NewWorkRepository(queries *query.Queries) *WorkRepository {
	return &WorkRepository{queries: queries}
}

// GetByID は作品IDで作品を取得します
func (r *WorkRepository) GetByID(ctx context.Context, id int64) (query.GetWorkByIDRow, error) {
	return r.queries.GetWorkByID(ctx, id)
}

// GetPopularWorksWithDetails は人気作品をキャスト・スタッフ情報と共に取得します
func (r *WorkRepository) GetPopularWorksWithDetails(ctx context.Context) ([]model.WorkWithDetails, error) {
	// 1. クエリ実行
	worksRows, err := r.queries.GetPopularWorks(ctx)
	if err != nil {
		return nil, err
	}

	if len(worksRows) == 0 {
		return []model.WorkWithDetails{}, nil
	}

	// 2. query.GetPopularWorksRow → model.Work に変換
	works := make([]model.Work, len(worksRows))
	workIDs := make([]int64, len(worksRows))
	for i, row := range worksRows {
		works[i] = r.workFromPopularRow(row)
		workIDs[i] = row.ID
	}

	// 3. キャストとスタッフを取得
	castsRows, err := r.queries.GetCastsByWorkIDs(ctx, workIDs)
	if err != nil {
		return nil, err
	}

	staffsRows, err := r.queries.GetStaffsByWorkIDs(ctx, workIDs)
	if err != nil {
		return nil, err
	}

	// 4. query結果をmodelに変換
	casts := r.castsFromRows(castsRows)
	staffs := r.staffsFromRows(staffsRows)

	// 5. 組み合わせる
	return r.combineWorkData(works, casts, staffs), nil
}

// workFromPopularRow は query.GetPopularWorksRow を model.Work に変換します
func (r *WorkRepository) workFromPopularRow(row query.GetPopularWorksRow) model.Work {
	work := model.Work{
		ID:                  row.ID,
		Title:               row.Title,
		TitleEn:             row.TitleEn,
		RecommendedImageURL: row.RecommendedImageUrl,
		WatchersCount:       row.WatchersCount,
	}

	// ImageDataを設定（NULLの場合は空文字列）
	if row.ImageData.Valid {
		work.ImageData = row.ImageData.String
	}

	// SeasonYearを設定
	if row.SeasonYear.Valid {
		work.SeasonYear = &row.SeasonYear.Int32
	}

	// SeasonNameを設定
	if row.SeasonName.Valid {
		work.SeasonName = &row.SeasonName.Int32
	}

	// CreatedAtを設定
	if row.CreatedAt.Valid {
		work.CreatedAt = row.CreatedAt.Time
	}

	return work
}

// castsFromRows は query結果を model.Cast に変換します
func (r *WorkRepository) castsFromRows(rows []query.GetCastsByWorkIDsRow) []model.Cast {
	casts := make([]model.Cast, len(rows))
	for i, row := range rows {
		casts[i] = model.Cast{
			ID:     row.ID,
			WorkID: row.WorkID,
			Name:   row.Name,
			NameEn: row.NameEn,
		}

		// CharacterNameを設定
		if row.CharacterName.Valid {
			casts[i].CharacterName = row.CharacterName.String
		}

		// CharacterNameEnを設定
		if row.CharacterNameEn.Valid {
			casts[i].CharacterNameEn = row.CharacterNameEn.String
		}

		// PersonNameを設定
		if row.PersonName.Valid {
			casts[i].PersonName = row.PersonName.String
		}

		// PersonNameEnを設定
		if row.PersonNameEn.Valid {
			casts[i].PersonNameEn = row.PersonNameEn.String
		}
	}
	return casts
}

// staffsFromRows は query結果を model.Staff に変換します
func (r *WorkRepository) staffsFromRows(rows []query.GetStaffsByWorkIDsRow) []model.Staff {
	staffs := make([]model.Staff, len(rows))
	for i, row := range rows {
		staffs[i] = model.Staff{
			ID:          row.ID,
			WorkID:      row.WorkID,
			Name:        row.Name,
			NameEn:      row.NameEn,
			Role:        row.Role,
			RoleOtherEn: row.RoleOtherEn,
		}

		// RoleOtherを設定
		if row.RoleOther.Valid {
			staffs[i].RoleOther = row.RoleOther.String
		}
	}
	return staffs
}

// combineWorkData は作品データとキャスト・スタッフデータを組み合わせます
func (r *WorkRepository) combineWorkData(
	works []model.Work,
	casts []model.Cast,
	staffs []model.Staff,
) []model.WorkWithDetails {
	// キャストとスタッフをwork_idでマッピング
	castsMap := make(map[int64][]model.Cast)
	for _, cast := range casts {
		castsMap[cast.WorkID] = append(castsMap[cast.WorkID], cast)
	}

	staffsMap := make(map[int64][]model.Staff)
	for _, staff := range staffs {
		staffsMap[staff.WorkID] = append(staffsMap[staff.WorkID], staff)
	}

	// WorkWithDetailsのスライスを作成
	result := make([]model.WorkWithDetails, len(works))
	for i, work := range works {
		result[i] = model.WorkWithDetails{
			Work:   work,
			Casts:  castsMap[work.ID],
			Staffs: staffsMap[work.ID],
		}
	}

	return result
}

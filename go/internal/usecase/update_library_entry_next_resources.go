package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/annict/annict/go/internal/repository"
)

// libraryEntryNextResourcesBatchSizeは1回のUPDATE文が読むライブラリの上限。件数に依らず
// 1文あたりのロックの保持とWALの生成を一定に抑えるために区切る。
const libraryEntryNextResourcesBatchSize = 1000

// libraryEntryNextResourcesWindowDaysは、更新の対象にするライブラリの更新日時を、実行日から
// さかのぼる日数 (Railsと同じ3日)。
const libraryEntryNextResourcesWindowDays = 3

// UpdateLibraryEntryNextResourcesUsecaseは、「見てる」のライブラリの次のエピソード・次の
// 放送枠を、自動生成で増えたエピソードと放送枠に合わせて更新する。
type UpdateLibraryEntryNextResourcesUsecase struct {
	libraryEntryRepo *repository.LibraryEntryRepository
}

// NewUpdateLibraryEntryNextResourcesUsecaseはUpdateLibraryEntryNextResourcesUsecaseを生成する。
func NewUpdateLibraryEntryNextResourcesUsecase(libraryEntryRepo *repository.LibraryEntryRepository) *UpdateLibraryEntryNextResourcesUsecase {
	return &UpdateLibraryEntryNextResourcesUsecase{libraryEntryRepo: libraryEntryRepo}
}

// UpdateLibraryEntryNextResourcesInputは実行時刻を持つ。対象にするライブラリの更新日時の下限は
// この時刻から求め、書き換えた行のupdated_atにもこの時刻を入れる。
type UpdateLibraryEntryNextResourcesInput struct {
	Now time.Time
}

// UpdateLibraryEntryNextResourcesOutputは書き換えたライブラリの件数を返す。
type UpdateLibraryEntryNextResourcesOutput struct {
	UpdatedCount int
}

// Executeは対象のライブラリをidの順にページに区切り、ページごとに1つのUPDATE文で更新する。
// ページの更新に失敗したら、そこで止めてエラーを返す。書き換えた行は条件を満たさなくなり、
// 残りは翌日の実行で読み直すため、途中までの更新は戻さない。
func (uc *UpdateLibraryEntryNextResourcesUsecase) Execute(ctx context.Context, input UpdateLibraryEntryNextResourcesInput) (*UpdateLibraryEntryNextResourcesOutput, error) {
	updatedFrom := libraryEntryNextResourcesUpdatedFrom(input.Now)

	output := &UpdateLibraryEntryNextResourcesOutput{}
	var afterID int64
	for {
		lastID, updatedCount, err := uc.libraryEntryRepo.UpdateNextResources(ctx, afterID, updatedFrom, input.Now, libraryEntryNextResourcesBatchSize)
		if err != nil {
			return nil, fmt.Errorf("ライブラリの次のエピソード・次の放送枠の更新に失敗 (id %dより後): %w", afterID, err)
		}
		if lastID == 0 {
			return output, nil
		}

		output.UpdatedCount += int(updatedCount)
		afterID = lastID
	}
}

// libraryEntryNextResourcesUpdatedFromは、対象にするライブラリの更新日時の下限として、実行日
// (UTC) のlibraryEntryNextResourcesWindowDays日前の0時 (UTC) を返す。Railsの
// `after(Date.today - 3, field: "library_entries.updated_at")` に合わせ、時刻ではなく日付で区切る。
func libraryEntryNextResourcesUpdatedFrom(now time.Time) time.Time {
	year, month, day := now.UTC().Date()
	return time.Date(year, month, day-libraryEntryNextResourcesWindowDays, 0, 0, 0, 0, time.UTC)
}

package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// SlotRepositoryはslotsテーブルおよび関連JOINへのデータアクセスを担う。
type SlotRepository struct {
	queries *query.Queries
}

// NewSlotRepositoryはSlotRepositoryを生成する。
func NewSlotRepository(queries *query.Queries) *SlotRepository {
	return &SlotRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいSlotRepositoryを返す。
func (r *SlotRepository) WithTx(tx *sql.Tx) *SlotRepository {
	return &SlotRepository{queries: r.queries.WithTx(tx)}
}

// ListForEpisodeGenerationは、開始日時がstartedFrom以上startedBefore未満の放送枠のうち、
// 自動生成がエピソードを紐づける対象を開始日時の順に返す。各放送枠は番組と、話数の計算で
// 回数から引く不定期放送の数 (PrecedingIrregularCount) を持つ。
func (r *SlotRepository) ListForEpisodeGeneration(ctx context.Context, startedFrom, startedBefore time.Time) ([]*model.Slot, error) {
	rows, err := r.queries.ListSlotsForEpisodeGeneration(ctx, query.ListSlotsForEpisodeGenerationParams{
		StartedFrom:   startedFrom,
		StartedBefore: startedBefore,
	})
	if err != nil {
		return nil, err
	}

	slots := make([]*model.Slot, len(rows))
	for i, row := range rows {
		// クエリは番組と回数を持つ放送枠に絞っているため、どちらもNULLではない。
		slots[i] = &model.Slot{
			ID:     model.SlotID(row.ID),
			WorkID: model.WorkID(row.WorkID),
			Program: &model.Program{
				ID:                              model.ProgramID(row.ProgramID.Int64),
				MinimumEpisodeGeneratableNumber: row.MinimumEpisodeGeneratableNumber,
			},
			Number:                  row.Number.Int32,
			StartedAt:               row.StartedAt,
			PrecedingIrregularCount: row.PrecedingIrregularCount,
		}
	}
	return slots, nil
}

// LinkEpisodeは、エピソードが紐づいていない未削除の放送枠にエピソードを紐づける。falseは、
// 放送枠がすでに紐づいているか削除されていて、何も書かなかったことを表す。
func (r *SlotRepository) LinkEpisode(ctx context.Context, slotID model.SlotID, episodeID model.EpisodeID) (bool, error) {
	affected, err := r.queries.LinkSlotEpisode(ctx, query.LinkSlotEpisodeParams{
		EpisodeID: sql.NullInt64{Int64: int64(episodeID), Valid: true},
		ID:        int64(slotID),
	})
	if err != nil {
		return false, err
	}
	return affected == 1, nil
}

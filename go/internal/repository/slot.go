package repository

import (
	"context"
	"database/sql"
	"errors"
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
// 自動生成がエピソードを紐づける対象のIDと所属作品を開始日時の順に返す。
// 話数の計算に使う値は、作品ロック後にGetForEpisodeGenerationByIDで読み直す。
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
		slots[i] = &model.Slot{ID: model.SlotID(row.ID), WorkID: model.WorkID(row.WorkID)}
	}
	return slots, nil
}

// GetForEpisodeGenerationByIDは、作品ロック後に対象の放送枠をロックして現在の値を返す。
// WithTxで使い、同じトランザクション内でLinkEpisodeを呼ぶ。対象外・紐づき済み・所属変更で
// 見つからないときと、ほかのトランザクションが放送枠をロックしているときはnilを返す。
func (r *SlotRepository) GetForEpisodeGenerationByID(ctx context.Context, id model.SlotID, workID model.WorkID, startedFrom, startedBefore time.Time) (*model.Slot, error) {
	row, err := r.queries.GetSlotForEpisodeGenerationByID(ctx, query.GetSlotForEpisodeGenerationByIDParams{
		ID:            int64(id),
		WorkID:        int64(workID),
		StartedFrom:   startedFrom,
		StartedBefore: startedBefore,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &model.Slot{
		ID:     model.SlotID(row.ID),
		WorkID: model.WorkID(row.WorkID),
		Program: &model.Program{
			ID:                              model.ProgramID(row.ProgramID.Int64),
			MinimumEpisodeGeneratableNumber: row.MinimumEpisodeGeneratableNumber,
		},
		Number:                  row.Number.Int32,
		StartedAt:               row.StartedAt,
		PrecedingIrregularCount: row.PrecedingIrregularCount,
	}, nil
}

// LinkEpisodeはGetForEpisodeGenerationByIDでロックした放送枠へエピソードを紐づける。
// falseは、照合に使った属性や対象条件が一致せず、何も書かなかったことを表す。
func (r *SlotRepository) LinkEpisode(ctx context.Context, slot *model.Slot, episodeID model.EpisodeID) (bool, error) {
	affected, err := r.queries.LinkSlotEpisode(ctx, query.LinkSlotEpisodeParams{
		EpisodeID: sql.NullInt64{Int64: int64(episodeID), Valid: true},
		ID:        int64(slot.ID),
		WorkID:    int64(slot.WorkID),
		ProgramID: sql.NullInt64{Int64: int64(slot.Program.ID), Valid: true},
		Number:    sql.NullInt32{Int32: slot.Number, Valid: true},
		StartedAt: slot.StartedAt,
	})
	if err != nil {
		return false, err
	}
	return affected == 1, nil
}

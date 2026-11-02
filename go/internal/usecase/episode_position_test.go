package usecase

import (
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

func episodeSibling(id model.EpisodeID, rawNumber *float64, sortNumber int32) repository.DBEpisodeGenerationSibling {
	return repository.DBEpisodeGenerationSibling{ID: id, RawNumber: rawNumber, SortNumber: sortNumber, Published: true}
}

func rawNumberPtr(rawNumber float64) *float64 {
	return &rawNumber
}

func episodeIDValue(id *model.EpisodeID) any {
	if id == nil {
		return nil
	}
	return *id
}

func TestGeneratedEpisodePosition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		siblings      []repository.DBEpisodeGenerationSibling
		maxSortNumber int32
		rawNumber     float64
		wantOK        bool
		wantSort      int32
		wantPrev      any
		wantNext      any
	}{
		{
			name:          "エピソードの無い作品では先頭に置く",
			maxSortNumber: 0,
			rawNumber:     1,
			wantOK:        true,
			wantSort:      100,
		},
		{
			name: "大きい話数が無ければ全エピソードの最大値の先に足す",
			siblings: []repository.DBEpisodeGenerationSibling{
				episodeSibling(1, rawNumberPtr(1), 100),
				episodeSibling(2, rawNumberPtr(2), 200),
			},
			// 削除済みのエピソードが最大値を持つ。
			maxSortNumber: 900,
			rawNumber:     3,
			wantOK:        true,
			wantSort:      1000,
			wantPrev:      model.EpisodeID(2),
		},
		{
			name: "話数の無いエピソードは大きいと見なさず、その後ろに足す",
			siblings: []repository.DBEpisodeGenerationSibling{
				episodeSibling(1, rawNumberPtr(1), 100),
				episodeSibling(2, nil, 200),
			},
			maxSortNumber: 200,
			rawNumber:     2,
			wantOK:        true,
			wantSort:      300,
			wantPrev:      model.EpisodeID(2),
		},
		{
			name: "大きい話数があればその直前の中間に入れる",
			siblings: []repository.DBEpisodeGenerationSibling{
				episodeSibling(1, rawNumberPtr(1), 100),
				episodeSibling(3, rawNumberPtr(3), 300),
				episodeSibling(4, rawNumberPtr(4), 400),
			},
			maxSortNumber: 400,
			rawNumber:     2,
			wantOK:        true,
			wantSort:      200,
			wantPrev:      model.EpisodeID(1),
			wantNext:      model.EpisodeID(3),
		},
		{
			name: "先頭より前に入れるときは1ステップ前に置く",
			siblings: []repository.DBEpisodeGenerationSibling{
				episodeSibling(13, rawNumberPtr(13), 1300),
			},
			maxSortNumber: 1300,
			rawNumber:     12,
			wantOK:        true,
			wantSort:      1200,
			wantNext:      model.EpisodeID(13),
		},
		{
			name: "前後の間が空いていなければ置けない",
			siblings: []repository.DBEpisodeGenerationSibling{
				episodeSibling(1, rawNumberPtr(1), 100),
				episodeSibling(3, rawNumberPtr(3), 101),
			},
			maxSortNumber: 101,
			rawNumber:     2,
			wantOK:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, ok := generatedEpisodePosition(tt.siblings, tt.maxSortNumber, tt.rawNumber)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v、期待値 = %v", ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if got.SortNumber != tt.wantSort {
				t.Errorf("SortNumber = %d、期待値 = %d", got.SortNumber, tt.wantSort)
			}
			if episodeIDValue(got.PrevEpisodeID) != tt.wantPrev {
				t.Errorf("PrevEpisodeID = %v、期待値 = %v", episodeIDValue(got.PrevEpisodeID), tt.wantPrev)
			}
			if episodeIDValue(got.NextEpisodeID) != tt.wantNext {
				t.Errorf("NextEpisodeID = %v、期待値 = %v", episodeIDValue(got.NextEpisodeID), tt.wantNext)
			}
		})
	}
}

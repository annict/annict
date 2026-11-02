package usecase

import (
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// episodePositionは新しいエピソードを置く並びの位置を表す。
type episodePosition struct {
	SortNumber    int32
	PrevEpisodeID *model.EpisodeID
	// NextEpisodeIDは並びで直後に来る未削除のエピソードで、新しいエピソードを前の
	// エピソードとして名乗るよう張り替える。末尾に足すときはnil。
	NextEpisodeID *model.EpisodeID
}

// appendedEpisodePositionは作品の末尾に足すエピソードの位置を返す。sort_numberは作品の
// 全エピソード (削除済み・非公開を含む) の最大値から1ステップ先にする。削除済みの行が持つ値と
// ぶつけないためと、開始話数が1でない作品で既存のエピソードより前に並ばないようにするため。
func appendedEpisodePosition(maxSortNumber int32, prevEpisodeID *model.EpisodeID) episodePosition {
	return episodePosition{
		SortNumber:    maxSortNumber + episodeSortNumberStep,
		PrevEpisodeID: prevEpisodeID,
	}
}

// generatedEpisodePositionは、自動生成するrawNumber話のエピソードを置く位置を返す。
// siblingsは作品の未削除のエピソードをsort_number順に並べたもの。
//
// rawNumberより大きい話数のエピソードがあれば、並びで最初のものの直前に入れ、無ければ末尾に
// 足す。話数の無いエピソード (編集者が置いた総集編など) はどの話数より大きいとも見なさない
// ため、末尾にあればその後ろに足すことになり、編集者が置いた位置を自動生成が動かさない。
// 直前に入れるときに前後のsort_numberの間が空いていなければ、falseを返して作らせない。
func generatedEpisodePosition(siblings []repository.DBEpisodeGenerationSibling, maxSortNumber int32, rawNumber float64) (episodePosition, bool) {
	nextIndex := -1
	for i, sibling := range siblings {
		if sibling.RawNumber != nil && *sibling.RawNumber > rawNumber {
			nextIndex = i
			break
		}
	}

	if nextIndex < 0 {
		var prevEpisodeID *model.EpisodeID
		if len(siblings) > 0 {
			prevEpisodeID = &siblings[len(siblings)-1].ID
		}
		return appendedEpisodePosition(maxSortNumber, prevEpisodeID), true
	}

	next := siblings[nextIndex]
	position := episodePosition{NextEpisodeID: &next.ID}

	if nextIndex == 0 {
		// 先頭に入れるときは前のエピソードが無いため、末尾に足すときと同じ間隔だけ前に置く。
		position.SortNumber = next.SortNumber - episodeSortNumberStep
		return position, true
	}

	prev := siblings[nextIndex-1]
	gap := int64(next.SortNumber) - int64(prev.SortNumber)
	if gap < 2 {
		return episodePosition{}, false
	}
	position.SortNumber = int32(int64(prev.SortNumber) + gap/2) // #nosec G115 -- 2つのint32の間の値
	position.PrevEpisodeID = &prev.ID
	return position, true
}

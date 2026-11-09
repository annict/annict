package repository

// AnimeIDReconcilePageは、関連テーブルのanime_idを照合した1ページの結果。
type AnimeIDReconcilePage struct {
	// LastIDはページの最後の行のID。次のページはこれより大きいIDから読む。ページが空なら0。
	LastID int64
	// Updatedは、写し元と食い違っていて書き換えた行の件数。
	Updated int
}

func newAnimeIDReconcilePage(lastID, updatedCount int64) AnimeIDReconcilePage {
	return AnimeIDReconcilePage{LastID: lastID, Updated: int(updatedCount)}
}

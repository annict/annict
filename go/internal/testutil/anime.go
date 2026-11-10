package testutil

import (
	"database/sql"
	"testing"

	"github.com/annict/annict/go/internal/model"
)

// CreateTestAnimeはテスト用の公開中の作品・エピソード (animesの行) を作成し、IDを返す。分類の行は作らない。
func CreateTestAnime(t *testing.T, tx *sql.Tx) model.AnimeID {
	t.Helper()
	return model.AnimeID(createTestRow(t, tx, "作品・エピソード",
		`INSERT INTO animes (status) VALUES ('published') RETURNING id`,
	))
}

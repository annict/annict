package seed

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/annict/annict/internal/testutil"
)

func TestCreateEpisodeUsecase_ExecuteBatchWithTx(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	ctx := context.Background()

	// テスト用の作品を作成
	workUsecase := NewCreateWorkUsecase(db)
	works, err := workUsecase.ExecuteBatchWithTx(ctx, tx, []CreateWorkParams{
		{Title: "テストアニメ1", TitleKana: "", Media: "tv", OfficialSiteURL: ""},
		{Title: "テストアニメ2", TitleKana: "", Media: "tv", OfficialSiteURL: ""},
	}, nil)
	if err != nil {
		t.Fatalf("作品作成エラー: %v", err)
	}

	if len(works) != 2 {
		t.Fatalf("作品数が不正: got %d, want 2", len(works))
	}

	workID1 := works[0].WorkID
	workID2 := works[1].WorkID

	// エピソード生成Usecaseを作成
	uc := NewCreateEpisodeUsecase(db)

	// テスト用のエピソードパラメータを作成
	// 作品1: 3話、作品2: 2話
	episodes := []CreateEpisodeParams{
		{WorkID: workID1, Number: "1", Title: "第1話", SortNumber: 1, PrevEpisodeID: nil},
		{WorkID: workID1, Number: "2", Title: "第2話", SortNumber: 2, PrevEpisodeID: nil},
		{WorkID: workID1, Number: "3", Title: "第3話", SortNumber: 3, PrevEpisodeID: nil},
		{WorkID: workID2, Number: "1", Title: "第1話", SortNumber: 1, PrevEpisodeID: nil},
		{WorkID: workID2, Number: "2", Title: "第2話", SortNumber: 2, PrevEpisodeID: nil},
	}

	// エピソードをバッチで作成
	results, err := uc.ExecuteBatchWithTx(ctx, tx, episodes, nil)
	if err != nil {
		t.Fatalf("エピソード作成エラー: %v", err)
	}

	// 結果を検証
	if len(results) != 5 {
		t.Fatalf("エピソード数が不正: got %d, want 5", len(results))
	}

	// エピソードIDが正しく設定されているか確認
	for i, result := range results {
		if result.EpisodeID == 0 {
			t.Errorf("エピソード%dのIDが0です", i+1)
		}
	}

	// データベースからエピソードを取得して検証
	query := `SELECT id, work_id, number, title, sort_number, prev_episode_id FROM episodes ORDER BY work_id, sort_number`
	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		t.Fatalf("エピソード取得エラー: %v", err)
	}
	defer rows.Close()

	var episodesInDB []struct {
		ID            int64
		WorkID        int64
		Number        string
		Title         string
		SortNumber    int32
		PrevEpisodeID *int64
	}

	for rows.Next() {
		var ep struct {
			ID            int64
			WorkID        int64
			Number        string
			Title         string
			SortNumber    int32
			PrevEpisodeID *int64
		}
		if err := rows.Scan(&ep.ID, &ep.WorkID, &ep.Number, &ep.Title, &ep.SortNumber, &ep.PrevEpisodeID); err != nil {
			t.Fatalf("エピソードスキャンエラー: %v", err)
		}
		episodesInDB = append(episodesInDB, ep)
	}

	if len(episodesInDB) != 5 {
		t.Fatalf("DBのエピソード数が不正: got %d, want 5", len(episodesInDB))
	}

	// 作品1のエピソード連鎖を検証
	// エピソード1: prev_episode_id = nil
	if episodesInDB[0].PrevEpisodeID != nil {
		t.Errorf("作品1エピソード1のprev_episode_idがnilではありません: %v", *episodesInDB[0].PrevEpisodeID)
	}

	// エピソード2: prev_episode_id = エピソード1のID
	if episodesInDB[1].PrevEpisodeID == nil {
		t.Errorf("作品1エピソード2のprev_episode_idがnilです")
	} else if *episodesInDB[1].PrevEpisodeID != episodesInDB[0].ID {
		t.Errorf("作品1エピソード2のprev_episode_idが不正: got %d, want %d", *episodesInDB[1].PrevEpisodeID, episodesInDB[0].ID)
	}

	// エピソード3: prev_episode_id = エピソード2のID
	if episodesInDB[2].PrevEpisodeID == nil {
		t.Errorf("作品1エピソード3のprev_episode_idがnilです")
	} else if *episodesInDB[2].PrevEpisodeID != episodesInDB[1].ID {
		t.Errorf("作品1エピソード3のprev_episode_idが不正: got %d, want %d", *episodesInDB[2].PrevEpisodeID, episodesInDB[1].ID)
	}

	// 作品2のエピソード連鎖を検証
	// エピソード1: prev_episode_id = nil
	if episodesInDB[3].PrevEpisodeID != nil {
		t.Errorf("作品2エピソード1のprev_episode_idがnilではありません: %v", *episodesInDB[3].PrevEpisodeID)
	}

	// エピソード2: prev_episode_id = エピソード1のID
	if episodesInDB[4].PrevEpisodeID == nil {
		t.Errorf("作品2エピソード2のprev_episode_idがnilです")
	} else if *episodesInDB[4].PrevEpisodeID != episodesInDB[3].ID {
		t.Errorf("作品2エピソード2のprev_episode_idが不正: got %d, want %d", *episodesInDB[4].PrevEpisodeID, episodesInDB[3].ID)
	}
}

func TestGenerateEpisodeParamsForWork(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	workID := int64(123)
	episodeCount := 12

	episodes := GenerateEpisodeParamsForWork(r, workID, episodeCount)

	// エピソード数を検証
	if len(episodes) != episodeCount {
		t.Errorf("エピソード数が不正: got %d, want %d", len(episodes), episodeCount)
	}

	// 各エピソードを検証
	for i, ep := range episodes {
		expectedNumber := i + 1

		// WorkIDを検証
		if ep.WorkID != workID {
			t.Errorf("エピソード%dのWorkIDが不正: got %d, want %d", i+1, ep.WorkID, workID)
		}

		// Numberを検証
		expectedNumberStr := fmt.Sprintf("第%d話", expectedNumber)
		if ep.Number != expectedNumberStr {
			t.Errorf("エピソード%dのNumberが不正: got %s, want %s", i+1, ep.Number, expectedNumberStr)
		}

		// SortNumberを検証
		if ep.SortNumber != int32(expectedNumber) {
			t.Errorf("エピソード%dのSortNumberが不正: got %d, want %d", i+1, ep.SortNumber, expectedNumber)
		}

		// Titleが空でないことを検証
		if ep.Title == "" {
			t.Errorf("エピソード%dのTitleが空です", i+1)
		}

		// PrevEpisodeIDがnilであることを検証（ExecuteBatch内で設定される）
		if ep.PrevEpisodeID != nil {
			t.Errorf("エピソード%dのPrevEpisodeIDがnilではありません: %v", i+1, *ep.PrevEpisodeID)
		}
	}
}

func TestCreateEpisodeUsecase_SingleEpisode(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	ctx := context.Background()

	// テスト用の作品を作成
	workUsecase := NewCreateWorkUsecase(db)
	works, err := workUsecase.ExecuteBatchWithTx(ctx, tx, []CreateWorkParams{
		{Title: "テストアニメ", TitleKana: "", Media: "tv", OfficialSiteURL: ""},
	}, nil)
	if err != nil {
		t.Fatalf("作品作成エラー: %v", err)
	}

	workID := works[0].WorkID

	// エピソード生成Usecaseを作成
	uc := NewCreateEpisodeUsecase(db)

	// 単一のエピソードを作成
	episodes := []CreateEpisodeParams{
		{WorkID: workID, Number: "1", Title: "第1話", SortNumber: 1, PrevEpisodeID: nil},
	}

	results, err := uc.ExecuteBatchWithTx(ctx, tx, episodes, nil)
	if err != nil {
		t.Fatalf("エピソード作成エラー: %v", err)
	}

	// 結果を検証
	if len(results) != 1 {
		t.Fatalf("エピソード数が不正: got %d, want 1", len(results))
	}

	if results[0].EpisodeID == 0 {
		t.Errorf("エピソードIDが0です")
	}

	// データベースからエピソードを取得して検証
	var ep struct {
		ID            int64
		WorkID        int64
		Number        string
		Title         string
		SortNumber    int32
		PrevEpisodeID *int64
	}

	query := `SELECT id, work_id, number, title, sort_number, prev_episode_id FROM episodes WHERE id = $1`
	err = tx.QueryRowContext(ctx, query, results[0].EpisodeID).Scan(
		&ep.ID, &ep.WorkID, &ep.Number, &ep.Title, &ep.SortNumber, &ep.PrevEpisodeID,
	)
	if err != nil {
		t.Fatalf("エピソード取得エラー: %v", err)
	}

	// 検証
	if ep.WorkID != workID {
		t.Errorf("WorkIDが不正: got %d, want %d", ep.WorkID, workID)
	}
	if ep.Number != "1" {
		t.Errorf("Numberが不正: got %s, want 1", ep.Number)
	}
	if ep.Title != "第1話" {
		t.Errorf("Titleが不正: got %s, want 第1話", ep.Title)
	}
	if ep.SortNumber != 1 {
		t.Errorf("SortNumberが不正: got %d, want 1", ep.SortNumber)
	}
	if ep.PrevEpisodeID != nil {
		t.Errorf("PrevEpisodeIDがnilではありません: %v", *ep.PrevEpisodeID)
	}
}

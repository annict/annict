package repository_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/annict/annict/internal/query"
	"github.com/annict/annict/internal/repository"
	"github.com/annict/annict/internal/testutil"
)

// TestGetPopularWorksWithDetails_Success は人気作品をキャスト・スタッフ情報と共に取得できることをテスト
func TestGetPopularWorksWithDetails_Success(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewWorkRepository(queries)

	// テスト用作品を作成（watchers_count=100）
	workID1 := testutil.NewWorkBuilder(t, tx).
		WithTitle("人気アニメ1").
		WithSeason(2024, testutil.SeasonSpring).
		Build()

	// 作品画像を作成
	testutil.NewWorkImageBuilder(t, tx, workID1).Build()

	// キャストを作成
	createTestCast(t, tx, workID1, "キャラクター1", "声優1")
	createTestCast(t, tx, workID1, "キャラクター2", "声優2")

	// スタッフを作成
	createTestStaff(t, tx, workID1, "監督", "director")
	createTestStaff(t, tx, workID1, "脚本", "series_composition")

	// 2つ目の作品を作成（watchers_count=100）
	workID2 := testutil.NewWorkBuilder(t, tx).
		WithTitle("人気アニメ2").
		WithSeason(2024, testutil.SeasonSummer).
		Build()

	// 作品画像を作成
	testutil.NewWorkImageBuilder(t, tx, workID2).Build()

	// キャストを作成（1件のみ）
	createTestCast(t, tx, workID2, "キャラクター3", "声優3")

	// 人気作品をキャスト・スタッフ情報と共に取得
	results, err := repo.GetPopularWorksWithDetails(context.Background())
	if err != nil {
		t.Fatalf("人気作品の取得に失敗: %v", err)
	}

	// 結果を検証
	if len(results) < 2 {
		t.Fatalf("期待する作品数が返されませんでした: got %d, want at least 2", len(results))
	}

	// 1つ目の作品を検証
	found1 := false
	for _, result := range results {
		if result.Work.ID == workID1 {
			found1 = true
			if result.Work.Title != "人気アニメ1" {
				t.Errorf("作品タイトルが一致しません: got %s, want %s", result.Work.Title, "人気アニメ1")
			}
			if result.Work.WatchersCount != 100 {
				t.Errorf("watchers_countが一致しません: got %d, want %d", result.Work.WatchersCount, 100)
			}
			if len(result.Casts) != 2 {
				t.Errorf("キャスト数が一致しません: got %d, want %d", len(result.Casts), 2)
			}
			if len(result.Staffs) != 2 {
				t.Errorf("スタッフ数が一致しません: got %d, want %d", len(result.Staffs), 2)
			}
			break
		}
	}
	if !found1 {
		t.Error("作品1が結果に含まれていません")
	}

	// 2つ目の作品を検証
	found2 := false
	for _, result := range results {
		if result.Work.ID == workID2 {
			found2 = true
			if result.Work.Title != "人気アニメ2" {
				t.Errorf("作品タイトルが一致しません: got %s, want %s", result.Work.Title, "人気アニメ2")
			}
			if len(result.Casts) != 1 {
				t.Errorf("キャスト数が一致しません: got %d, want %d", len(result.Casts), 1)
			}
			if len(result.Staffs) != 0 {
				t.Errorf("スタッフ数が一致しません: got %d, want %d", len(result.Staffs), 0)
			}
			break
		}
	}
	if !found2 {
		t.Error("作品2が結果に含まれていません")
	}
}

// TestGetPopularWorksWithDetails_EmptyResult は作品が存在しない場合に空の配列が返されることをテスト
func TestGetPopularWorksWithDetails_EmptyResult(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewWorkRepository(queries)

	// 人気作品を取得（作品が存在しないため空の配列が返される）
	results, err := repo.GetPopularWorksWithDetails(context.Background())
	if err != nil {
		t.Fatalf("人気作品の取得に失敗: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("空の配列が返されるべきですが、%d件の作品が返されました", len(results))
	}
}

// TestGetPopularWorksWithDetails_NoCastsOrStaffs はキャスト・スタッフがない作品も正しく取得できることをテスト
func TestGetPopularWorksWithDetails_NoCastsOrStaffs(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewWorkRepository(queries)

	// キャスト・スタッフがない作品を作成
	workID := testutil.NewWorkBuilder(t, tx).
		WithTitle("キャストなしアニメ").
		WithSeason(2024, testutil.SeasonWinter).
		Build()

	// 作品画像を作成
	testutil.NewWorkImageBuilder(t, tx, workID).Build()

	// 人気作品を取得
	results, err := repo.GetPopularWorksWithDetails(context.Background())
	if err != nil {
		t.Fatalf("人気作品の取得に失敗: %v", err)
	}

	// 結果を検証
	if len(results) == 0 {
		t.Fatal("作品が返されませんでした")
	}

	found := false
	for _, result := range results {
		if result.Work.ID == workID {
			found = true
			if result.Work.Title != "キャストなしアニメ" {
				t.Errorf("作品タイトルが一致しません: got %s, want %s", result.Work.Title, "キャストなしアニメ")
			}
			if len(result.Casts) != 0 {
				t.Errorf("キャスト配列は空であるべきですが、%d件が返されました", len(result.Casts))
			}
			if len(result.Staffs) != 0 {
				t.Errorf("スタッフ配列は空であるべきですが、%d件が返されました", len(result.Staffs))
			}
			break
		}
	}
	if !found {
		t.Error("作品が結果に含まれていません")
	}
}

// createTestCast はテスト用のキャストデータを作成します
func createTestCast(t *testing.T, tx *sql.Tx, workID int64, characterName, personName string) {
	t.Helper()

	// キャラクターを作成
	var characterID int64
	err := tx.QueryRow(`
		INSERT INTO characters (name, name_en, name_kana, series_id, created_at, updated_at)
		VALUES ($1, $2, '', NULL, NOW(), NOW())
		RETURNING id
	`, characterName, characterName).Scan(&characterID)
	if err != nil {
		t.Fatalf("キャラクターの作成に失敗: %v", err)
	}

	// 人物を作成
	var personID int64
	err = tx.QueryRow(`
		INSERT INTO people (name, name_en, name_kana, created_at, updated_at)
		VALUES ($1, $2, '', NOW(), NOW())
		RETURNING id
	`, personName, personName).Scan(&personID)
	if err != nil {
		t.Fatalf("人物の作成に失敗: %v", err)
	}

	// キャストを作成
	_, err = tx.Exec(`
		INSERT INTO casts (work_id, character_id, person_id, name, name_en, sort_number, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	`, workID, characterID, personID, characterName+" as "+personName, characterName+" as "+personName, 1)
	if err != nil {
		t.Fatalf("キャストの作成に失敗: %v", err)
	}
}

// createTestStaff はテスト用のスタッフデータを作成します
func createTestStaff(t *testing.T, tx *sql.Tx, workID int64, name, role string) {
	t.Helper()

	// 人物を作成（resource_idとresource_typeに使用）
	var personID int64
	err := tx.QueryRow(`
		INSERT INTO people (name, name_en, name_kana, created_at, updated_at)
		VALUES ($1, $2, '', NOW(), NOW())
		RETURNING id
	`, name, name).Scan(&personID)
	if err != nil {
		t.Fatalf("人物の作成に失敗: %v", err)
	}

	// スタッフを作成
	_, err = tx.Exec(`
		INSERT INTO staffs (work_id, name, name_en, role, role_other, role_other_en, resource_id, resource_type, sort_number, created_at, updated_at)
		VALUES ($1, $2, $3, $4, '', '', $5, $6, 1, NOW(), NOW())
	`, workID, name, name, role, personID, "Person")
	if err != nil {
		t.Fatalf("スタッフの作成に失敗: %v", err)
	}
}

package repository_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
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

// TestWorkRepository_ListForDB はDB管理画面用の作品一覧取得をテスト
func TestWorkRepository_ListForDB(t *testing.T) {
	t.Parallel()

	t.Run("正常系: フィルタなしで作品一覧を取得できる", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTestDB(t)
		ctx := context.Background()

		testutil.NewWorkBuilder(t, tx).WithTitle("作品A").Build()
		testutil.NewWorkBuilder(t, tx).WithTitle("作品B").Build()

		repo := repository.NewWorkRepository(query.New(db)).WithTx(tx)
		items, err := repo.ListForDB(ctx, repository.DBWorkListParams{
			Page:    1,
			PerPage: 100,
		})
		if err != nil {
			t.Fatalf("ListForDB() error = %v", err)
		}

		if len(items) < 2 {
			t.Fatalf("ListForDB() got %d items, want >= 2", len(items))
		}

		// ID降順で取得されることを確認（作品Bが先）
		if items[0].Title != "作品B" {
			t.Errorf("items[0].Title = %q, want %q", items[0].Title, "作品B")
		}
		if items[1].Title != "作品A" {
			t.Errorf("items[1].Title = %q, want %q", items[1].Title, "作品A")
		}
	})

	t.Run("正常系: 削除済み作品は除外される", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTestDB(t)
		ctx := context.Background()

		testutil.NewWorkBuilder(t, tx).WithTitle("公開作品").WithStatus("published").Build()
		testutil.NewWorkBuilder(t, tx).WithTitle("削除作品").WithStatus("deleted").Build()

		repo := repository.NewWorkRepository(query.New(db)).WithTx(tx)
		items, err := repo.ListForDB(ctx, repository.DBWorkListParams{
			Page:    1,
			PerPage: 100,
		})
		if err != nil {
			t.Fatalf("ListForDB() error = %v", err)
		}

		for _, item := range items {
			if item.Title == "削除作品" {
				t.Error("deleted work should not be returned")
			}
		}
	})

	t.Run("正常系: エピソード未登録フィルタ", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTestDB(t)
		ctx := context.Background()

		// エピソードありの作品
		workWithEp := testutil.NewWorkBuilder(t, tx).WithTitle("エピソードあり").Build()
		testutil.NewEpisodeBuilder(t, tx, workWithEp).Build()

		// エピソードなしの作品（no_episodes=false）
		testutil.NewWorkBuilder(t, tx).WithTitle("エピソードなし").Build()

		// no_episodes=trueの作品（エピソード不要とマーク済み）
		testutil.NewWorkBuilder(t, tx).WithTitle("エピソード不要").WithNoEpisodes(true).Build()

		repo := repository.NewWorkRepository(query.New(db)).WithTx(tx)
		items, err := repo.ListForDB(ctx, repository.DBWorkListParams{
			FilterNoEpisodes: true,
			Page:             1,
			PerPage:          100,
		})
		if err != nil {
			t.Fatalf("ListForDB() error = %v", err)
		}

		found := false
		for _, item := range items {
			if item.Title == "エピソードあり" {
				t.Error("work with episodes should not be returned")
			}
			if item.Title == "エピソード不要" {
				t.Error("work with no_episodes=true should not be returned")
			}
			if item.Title == "エピソードなし" {
				found = true
			}
		}
		if !found {
			t.Error("work without episodes should be returned")
		}
	})

	t.Run("正常系: 画像未設定フィルタ", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTestDB(t)
		ctx := context.Background()

		testutil.CreateTestWorkWithImage(t, tx, "画像あり")
		testutil.NewWorkBuilder(t, tx).WithTitle("画像なし").Build()

		repo := repository.NewWorkRepository(query.New(db)).WithTx(tx)
		items, err := repo.ListForDB(ctx, repository.DBWorkListParams{
			FilterNoImage: true,
			Page:          1,
			PerPage:       100,
		})
		if err != nil {
			t.Fatalf("ListForDB() error = %v", err)
		}

		for _, item := range items {
			if item.Title == "画像あり" {
				t.Error("work with image should not be returned")
			}
		}
	})

	t.Run("正常系: シーズン未設定フィルタ", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTestDB(t)
		ctx := context.Background()

		testutil.NewWorkBuilder(t, tx).WithTitle("シーズンあり").WithSeason(2024, testutil.SeasonSpring).Build()
		testutil.NewWorkBuilder(t, tx).WithTitle("シーズンなし").WithNoSeason().Build()

		repo := repository.NewWorkRepository(query.New(db)).WithTx(tx)
		items, err := repo.ListForDB(ctx, repository.DBWorkListParams{
			FilterNoSeason: true,
			Page:           1,
			PerPage:        100,
		})
		if err != nil {
			t.Fatalf("ListForDB() error = %v", err)
		}

		for _, item := range items {
			if item.Title == "シーズンあり" {
				t.Error("work with season should not be returned")
			}
		}

		found := false
		for _, item := range items {
			if item.Title == "シーズンなし" {
				found = true
			}
		}
		if !found {
			t.Error("work without season should be returned")
		}
	})

	t.Run("正常系: シーズン指定フィルタ", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTestDB(t)
		ctx := context.Background()

		testutil.NewWorkBuilder(t, tx).WithTitle("2024春").WithSeason(2024, testutil.SeasonSpring).Build()
		testutil.NewWorkBuilder(t, tx).WithTitle("2024夏").WithSeason(2024, testutil.SeasonSummer).Build()

		year := int32(2024)
		season := int32(testutil.SeasonSpring)
		repo := repository.NewWorkRepository(query.New(db)).WithTx(tx)
		items, err := repo.ListForDB(ctx, repository.DBWorkListParams{
			SeasonYear: &year,
			SeasonName: &season,
			Page:       1,
			PerPage:    100,
		})
		if err != nil {
			t.Fatalf("ListForDB() error = %v", err)
		}

		for _, item := range items {
			if item.Title == "2024夏" {
				t.Error("work in different season should not be returned")
			}
		}

		found := false
		for _, item := range items {
			if item.Title == "2024春" {
				found = true
			}
		}
		if !found {
			t.Error("work in specified season should be returned")
		}
	})

	t.Run("正常系: ページネーション", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTestDB(t)
		ctx := context.Background()

		testutil.NewWorkBuilder(t, tx).WithTitle("作品1").Build()
		testutil.NewWorkBuilder(t, tx).WithTitle("作品2").Build()
		testutil.NewWorkBuilder(t, tx).WithTitle("作品3").Build()

		repo := repository.NewWorkRepository(query.New(db)).WithTx(tx)

		// 1ページ目（2件ずつ）
		page1, err := repo.ListForDB(ctx, repository.DBWorkListParams{Page: 1, PerPage: 2})
		if err != nil {
			t.Fatalf("ListForDB() page1 error = %v", err)
		}
		if len(page1) != 2 {
			t.Fatalf("page1 got %d items, want 2", len(page1))
		}

		// 2ページ目
		page2, err := repo.ListForDB(ctx, repository.DBWorkListParams{Page: 2, PerPage: 2})
		if err != nil {
			t.Fatalf("ListForDB() page2 error = %v", err)
		}
		if len(page2) != 1 {
			t.Fatalf("page2 got %d items, want 1", len(page2))
		}

		// ページ1とページ2で重複がないことを確認
		if page1[0].ID == page2[0].ID || page1[1].ID == page2[0].ID {
			t.Error("pages should not have overlapping items")
		}
	})
}

// TestWorkRepository_CountForDB はDB管理画面用の作品数取得をテスト
func TestWorkRepository_CountForDB(t *testing.T) {
	t.Parallel()

	t.Run("正常系: フィルタなしで作品数を取得できる", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTestDB(t)
		ctx := context.Background()

		testutil.NewWorkBuilder(t, tx).WithTitle("作品A").Build()
		testutil.NewWorkBuilder(t, tx).WithTitle("作品B").Build()
		testutil.NewWorkBuilder(t, tx).WithTitle("削除済み").WithStatus("deleted").Build()

		repo := repository.NewWorkRepository(query.New(db)).WithTx(tx)
		count, err := repo.CountForDB(ctx, repository.DBWorkListParams{})
		if err != nil {
			t.Fatalf("CountForDB() error = %v", err)
		}

		if count < 2 {
			t.Errorf("CountForDB() = %d, want >= 2", count)
		}
	})

	t.Run("正常系: フィルタ適用時のカウント", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTestDB(t)
		ctx := context.Background()

		testutil.NewWorkBuilder(t, tx).WithTitle("シーズンあり").WithSeason(2024, testutil.SeasonSpring).Build()
		testutil.NewWorkBuilder(t, tx).WithTitle("シーズンなし").WithNoSeason().Build()

		repo := repository.NewWorkRepository(query.New(db)).WithTx(tx)
		count, err := repo.CountForDB(ctx, repository.DBWorkListParams{
			FilterNoSeason: true,
		})
		if err != nil {
			t.Fatalf("CountForDB() error = %v", err)
		}

		if count != 1 {
			t.Errorf("CountForDB() = %d, want 1", count)
		}
	})
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

package repository_test

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// TestWorkRepository_GetPopular は人気作品の一覧取得をテスト
func TestWorkRepository_GetPopular(t *testing.T) {
	t.Parallel()

	t.Run("正常系: 人気作品の一覧を取得できる", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)
		repo := repository.NewWorkRepository(queries)

		workID := testutil.NewWorkBuilder(t, tx).
			WithTitle("人気アニメ").
			WithSeason(2024, testutil.SeasonSpring).
			Build()
		testutil.NewWorkImageBuilder(t, tx, workID).Build()

		works, err := repo.GetPopular(context.Background())
		if err != nil {
			t.Fatalf("GetPopular() error = %v", err)
		}

		found := false
		for _, w := range works {
			if w.ID == workID {
				found = true
				if w.Title != "人気アニメ" {
					t.Errorf("Title = %q, want %q", w.Title, "人気アニメ")
				}
				if w.WatchersCount != 100 {
					t.Errorf("WatchersCount = %d, want 100", w.WatchersCount)
				}
			}
		}
		if !found {
			t.Errorf("作成した作品 (ID=%d) が結果に含まれていません", workID)
		}
	})

	t.Run("正常系: 作品が存在しない場合は空のスライスを返す", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)
		repo := repository.NewWorkRepository(queries)

		works, err := repo.GetPopular(context.Background())
		if err != nil {
			t.Fatalf("GetPopular() error = %v", err)
		}
		if len(works) != 0 {
			t.Errorf("len(works) = %d, want 0", len(works))
		}
	})

	t.Run("正常系: watchers_count の降順で取得される", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)
		repo := repository.NewWorkRepository(queries)

		// 同一トランザクション内で 3 件投入し、当該作品集合内の相対順序を検証する
		// （他テストとの並列実行下でもトランザクション分離により独立性が保たれる）
		idLow := testutil.NewWorkBuilder(t, tx).WithTitle("少").WithWatchersCount(10).Build()
		idHigh := testutil.NewWorkBuilder(t, tx).WithTitle("多").WithWatchersCount(1000).Build()
		idMid := testutil.NewWorkBuilder(t, tx).WithTitle("中").WithWatchersCount(500).Build()

		works, err := repo.GetPopular(context.Background())
		if err != nil {
			t.Fatalf("GetPopular() error = %v", err)
		}

		// 当該テストで投入した 3 作品のみを抽出して順序を検証する
		var ordered []model.WorkID
		for _, w := range works {
			if w.ID == idLow || w.ID == idHigh || w.ID == idMid {
				ordered = append(ordered, w.ID)
			}
		}
		if len(ordered) != 3 {
			t.Fatalf("投入した 3 作品が結果に揃わない: got=%v", ordered)
		}

		want := []model.WorkID{idHigh, idMid, idLow}
		for i, id := range want {
			if ordered[i] != id {
				t.Errorf("ordered[%d] = %v, want %v (watchers_count 降順)", i, ordered[i], id)
			}
		}
	})
}

// TestWorkRepository_GetByID は作品IDで作品を取得し *model.Work に変換できることをテスト
func TestWorkRepository_GetByID(t *testing.T) {
	t.Parallel()

	t.Run("正常系: シーズン情報ありの作品を取得できる", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)
		repo := repository.NewWorkRepository(queries)

		workID := testutil.NewWorkBuilder(t, tx).
			WithTitle("作品X").
			WithSeason(2024, testutil.SeasonSpring).
			Build()

		work, err := repo.GetByID(context.Background(), workID)
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}
		if work == nil {
			t.Fatal("work should not be nil")
		}
		if work.ID != workID {
			t.Errorf("ID = %v, want %v", work.ID, workID)
		}
		if work.Title != "作品X" {
			t.Errorf("Title = %q, want %q", work.Title, "作品X")
		}
		if work.WatchersCount != 100 {
			t.Errorf("WatchersCount = %d, want 100", work.WatchersCount)
		}
		if work.SeasonYear == nil || *work.SeasonYear != 2024 {
			t.Errorf("SeasonYear = %v, want 2024", work.SeasonYear)
		}
		if work.SeasonName == nil || *work.SeasonName != testutil.SeasonSpring {
			t.Errorf("SeasonName = %v, want %d", work.SeasonName, testutil.SeasonSpring)
		}
		// title_kana は WorkBuilder が空文字で投入するため、Model 側は nil になる
		if work.TitleKana != nil {
			t.Errorf("TitleKana = %v, want nil (empty string row should map to nil)", work.TitleKana)
		}
		if work.CreatedAt.IsZero() {
			t.Error("CreatedAt should not be zero")
		}
	})

	t.Run("正常系: シーズン情報なしの作品はポインタが nil になる", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)
		repo := repository.NewWorkRepository(queries)

		workID := testutil.NewWorkBuilder(t, tx).
			WithTitle("シーズンなし作品").
			WithNoSeason().
			Build()

		work, err := repo.GetByID(context.Background(), workID)
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}
		if work == nil {
			t.Fatal("work should not be nil")
		}
		if work.SeasonYear != nil {
			t.Errorf("SeasonYear = %v, want nil", work.SeasonYear)
		}
		if work.SeasonName != nil {
			t.Errorf("SeasonName = %v, want nil", work.SeasonName)
		}
	})

	t.Run("異常系: 存在しないIDを指定するとエラーを返す", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)
		repo := repository.NewWorkRepository(queries)

		_, err := repo.GetByID(context.Background(), 999999999)
		if err == nil {
			t.Fatal("expected error for non-existent ID, got nil")
		}
	})
}

// TestWorkRepository_ListForDB はDB管理画面用の作品一覧取得をテスト
func TestWorkRepository_ListForDB(t *testing.T) {
	t.Parallel()

	t.Run("正常系: フィルタなしで作品一覧を取得できる", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTx(t)
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
		db, tx := testutil.SetupTx(t)
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
		db, tx := testutil.SetupTx(t)
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
		db, tx := testutil.SetupTx(t)
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
		db, tx := testutil.SetupTx(t)
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
		db, tx := testutil.SetupTx(t)
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
		db, tx := testutil.SetupTx(t)
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
		db, tx := testutil.SetupTx(t)
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
		db, tx := testutil.SetupTx(t)
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

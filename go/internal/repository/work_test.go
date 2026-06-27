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

// TestWorkRepository_ListIDsAfter verifies keyset pagination. Other tests commit
// works to the shared test DB, so this test cannot assert an exact page content;
// instead it asserts the keyset invariants that hold regardless of foreign rows:
// the first id strictly greater than the cursor, ascending order, the limit, and
// strict cursor advancement. id1-1 as a cursor pins the first row to id1 because
// no row can have an id in the open interval (id1-1, id1).
//
// [Ja] TestWorkRepository_ListIDsAfter は keyset ページネーションを検証する。他テストが
// 共有テスト DB に works をコミットするため、ページ内容の厳密一致は検証できない。代わりに
// 他行の有無に依らず成立する keyset の不変条件 (カーソルより厳密に大きい最初の id・昇順・
// LIMIT・カーソルの厳密前進) を検証する。カーソルに id1-1 を使うと、開区間 (id1-1, id1) に
// id を持つ行は存在しえないため、最初の行が id1 に固定される。
func TestWorkRepository_ListIDsAfter(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewWorkRepository(query.New(db).WithTx(tx))
	ctx := context.Background()

	// Three works in ascending id order (the sequence is monotonic). The middle
	// row is unnamed; it only needs to exist so the first page (limit 2) is full
	// and id3 stays ahead of the second-page cursor.
	//
	// [Ja] id 昇順の 3 件 (シーケンスは単調増加)。中間の行は名前を付けない。最初のページ
	// (limit 2) が満杯になり、id3 が 2 ページ目のカーソルより先に残るために存在させるだけ。
	id1 := testutil.NewWorkBuilder(t, tx).Build()
	testutil.NewWorkBuilder(t, tx).Build()
	id3 := testutil.NewWorkBuilder(t, tx).Build()

	t.Run("カーソル直後の id を LIMIT どおり 1 件返す", func(t *testing.T) {
		got, err := repo.ListIDsAfter(ctx, id1-1, 1)
		if err != nil {
			t.Fatalf("ListIDsAfter() error = %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("len = %d, want 1", len(got))
		}
		if got[0] != id1 {
			t.Errorf("got[0] = %d, want %d", got[0], id1)
		}
	})

	t.Run("昇順かつカーソルより大きい id だけを LIMIT 件数まで返す", func(t *testing.T) {
		got, err := repo.ListIDsAfter(ctx, id1-1, 2)
		if err != nil {
			t.Fatalf("ListIDsAfter() error = %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("len = %d, want 2", len(got))
		}
		if got[0] != id1 {
			t.Errorf("got[0] = %d, want %d", got[0], id1)
		}
		if got[0] >= got[1] {
			t.Errorf("not ascending: %v", got)
		}
	})

	t.Run("カーソルを進めると重複なく前進する", func(t *testing.T) {
		page1, err := repo.ListIDsAfter(ctx, id1-1, 2)
		if err != nil {
			t.Fatalf("ListIDsAfter() error = %v", err)
		}
		cursor := page1[len(page1)-1]

		page2, err := repo.ListIDsAfter(ctx, cursor, 2)
		if err != nil {
			t.Fatalf("ListIDsAfter() error = %v", err)
		}
		// id3 is still ahead of the cursor (the cursor is at most id2), so the next
		// page is non-empty and every id is strictly greater than the cursor.
		//
		// [Ja] id3 はまだカーソルの先にある (カーソルは高々 id2) ため、次ページは空でなく、
		// すべての id がカーソルより厳密に大きい。
		if len(page2) == 0 {
			t.Fatal("page2 is empty, want at least id3")
		}
		if page2[0] <= cursor {
			t.Errorf("page2[0] = %d, want > cursor %d", page2[0], cursor)
		}
	})

	t.Run("全 id より大きいカーソルでは空を返す", func(t *testing.T) {
		got, err := repo.ListIDsAfter(ctx, id3+1_000_000_000, 10)
		if err != nil {
			t.Fatalf("ListIDsAfter() error = %v", err)
		}
		if len(got) != 0 {
			t.Errorf("len = %d, want 0", len(got))
		}
	})
}

func TestWorkRepository_ListForSatelliteSyncByIDs(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewWorkRepository(queries)
	animeRepo := repository.NewAnimeRepository(queries)
	ctx := context.Background()

	animeID := createTestAnime(t, animeRepo, "別表同期アニメ")

	// A fully populated work mapped to an anime, plus one with every nullable satellite
	// column left empty and no anime_id, to check the projection round-trips both. The
	// builder seeds the base row; the UPDATE sets the satellite columns deterministically.
	//
	// [Ja] anime に紐づくフル設定の work と、NULL 許容の別表列をすべて空にし anime_id も
	// 持たない work を用意し、射影が双方を往復できるか検証する。ビルダーが土台の行を作り、
	// UPDATE で別表列を決定論的に設定する。
	full := testutil.NewWorkBuilder(t, tx).WithTitle("フル設定").Build()
	if _, err := tx.Exec(`
		UPDATE works SET
			anime_id = $2,
			sc_tid = 123,
			mal_anime_id = 456,
			official_site_url = 'https://example.dev/ja',
			official_site_url_en = 'https://example.dev/en',
			wikipedia_url = 'https://ja.wikipedia.example.dev',
			wikipedia_url_en = 'https://en.wikipedia.example.dev',
			twitter_username = 'anime_official',
			twitter_hashtag = 'anime',
			season_year = 2026,
			season_name = 2,
			started_on = '2026-01-05',
			ended_on = '2026-03-30'
		WHERE id = $1
	`, int64(full), int64(animeID)); err != nil {
		t.Fatalf("works の別表列の設定に失敗: %v", err)
	}

	empty := testutil.NewWorkBuilder(t, tx).WithTitle("別表列なし").WithNoSeason().Build()
	if _, err := tx.Exec(`
		UPDATE works SET
			anime_id = NULL,
			sc_tid = NULL,
			mal_anime_id = NULL,
			official_site_url = '',
			official_site_url_en = '',
			wikipedia_url = '',
			wikipedia_url_en = '',
			twitter_username = NULL,
			twitter_hashtag = NULL,
			season_year = NULL,
			season_name = NULL,
			started_on = NULL,
			ended_on = NULL
		WHERE id = $1
	`, int64(empty)); err != nil {
		t.Fatalf("works の別表列のクリアに失敗: %v", err)
	}

	t.Run("射影と anime_id 解決を id 昇順で返す", func(t *testing.T) {
		// Input order is reversed to confirm the loader orders by id, not by input.
		//
		// [Ja] 入力順を逆にして、ローダーが入力順でなく id 昇順で返すことを確認する。
		works, err := repo.ListForSatelliteSyncByIDs(ctx, []model.WorkID{empty, full})
		if err != nil {
			t.Fatalf("ListForSatelliteSyncByIDs() error = %v", err)
		}
		if len(works) != 2 {
			t.Fatalf("len = %d, want 2", len(works))
		}

		got := works[0]
		if got.ID != full {
			t.Fatalf("works[0].ID = %d, want %d (id 昇順)", got.ID, full)
		}
		if got.AnimeID == nil || *got.AnimeID != animeID {
			t.Errorf("AnimeID = %v, want %d", got.AnimeID, animeID)
		}
		if got.ScTid == nil || *got.ScTid != 123 {
			t.Errorf("ScTid = %v, want 123", got.ScTid)
		}
		if got.MalAnimeID == nil || *got.MalAnimeID != 456 {
			t.Errorf("MalAnimeID = %v, want 456", got.MalAnimeID)
		}
		if got.OfficialSiteURL != "https://example.dev/ja" {
			t.Errorf("OfficialSiteURL = %q", got.OfficialSiteURL)
		}
		if got.OfficialSiteURLEn != "https://example.dev/en" {
			t.Errorf("OfficialSiteURLEn = %q", got.OfficialSiteURLEn)
		}
		if got.WikipediaURL != "https://ja.wikipedia.example.dev" {
			t.Errorf("WikipediaURL = %q", got.WikipediaURL)
		}
		if got.WikipediaURLEn != "https://en.wikipedia.example.dev" {
			t.Errorf("WikipediaURLEn = %q", got.WikipediaURLEn)
		}
		if got.TwitterUsername == nil || *got.TwitterUsername != "anime_official" {
			t.Errorf("TwitterUsername = %v, want anime_official", got.TwitterUsername)
		}
		if got.TwitterHashtag == nil || *got.TwitterHashtag != "anime" {
			t.Errorf("TwitterHashtag = %v, want anime", got.TwitterHashtag)
		}
		if got.SeasonYear == nil || *got.SeasonYear != 2026 {
			t.Errorf("SeasonYear = %v, want 2026", got.SeasonYear)
		}
		if got.SeasonName == nil || *got.SeasonName != 2 {
			t.Errorf("SeasonName = %v, want 2", got.SeasonName)
		}
		if got.StartedOn == nil || got.StartedOn.Year() != 2026 || got.StartedOn.Month() != 1 || got.StartedOn.Day() != 5 {
			t.Errorf("StartedOn = %v, want 2026-01-05", got.StartedOn)
		}
		if got.EndedOn == nil || got.EndedOn.Year() != 2026 || got.EndedOn.Month() != 3 || got.EndedOn.Day() != 30 {
			t.Errorf("EndedOn = %v, want 2026-03-30", got.EndedOn)
		}
	})

	t.Run("NULL / 空のソース列は nil・空文字列で返る", func(t *testing.T) {
		works, err := repo.ListForSatelliteSyncByIDs(ctx, []model.WorkID{empty})
		if err != nil {
			t.Fatalf("ListForSatelliteSyncByIDs() error = %v", err)
		}
		if len(works) != 1 {
			t.Fatalf("len = %d, want 1", len(works))
		}

		got := works[0]
		if got.AnimeID != nil {
			t.Errorf("AnimeID = %v, want nil", got.AnimeID)
		}
		if got.ScTid != nil || got.MalAnimeID != nil {
			t.Errorf("ScTid / MalAnimeID = %v / %v, want nil / nil", got.ScTid, got.MalAnimeID)
		}
		if got.TwitterUsername != nil || got.TwitterHashtag != nil {
			t.Errorf("TwitterUsername / TwitterHashtag = %v / %v, want nil / nil", got.TwitterUsername, got.TwitterHashtag)
		}
		if got.SeasonYear != nil || got.SeasonName != nil {
			t.Errorf("SeasonYear / SeasonName = %v / %v, want nil / nil", got.SeasonYear, got.SeasonName)
		}
		if got.StartedOn != nil || got.EndedOn != nil {
			t.Errorf("StartedOn / EndedOn = %v / %v, want nil / nil", got.StartedOn, got.EndedOn)
		}
		// NOT NULL DEFAULT '' url columns keep the empty string (mapped to "no row" later).
		//
		// [Ja] NOT NULL DEFAULT '' の url 列は空文字列のまま (後段で「行なし」に写像)。
		if got.OfficialSiteURL != "" || got.WikipediaURL != "" || got.OfficialSiteURLEn != "" || got.WikipediaURLEn != "" {
			t.Errorf("url 列が空文字列でない: %q / %q / %q / %q", got.OfficialSiteURL, got.WikipediaURL, got.OfficialSiteURLEn, got.WikipediaURLEn)
		}
	})

	t.Run("空入力ではクエリせず空スライスを返す", func(t *testing.T) {
		works, err := repo.ListForSatelliteSyncByIDs(ctx, nil)
		if err != nil {
			t.Fatalf("ListForSatelliteSyncByIDs() error = %v", err)
		}
		if len(works) != 0 {
			t.Errorf("len = %d, want 0", len(works))
		}
	})
}

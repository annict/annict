package repository_test

import (
	"context"
	"database/sql"
	"math"
	"testing"
	"time"

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

	t.Run("正常系: タイトル 3 種とメディアがマッピングされる", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTx(t)
		ctx := context.Background()

		workID := testutil.NewWorkBuilder(t, tx).
			WithTitle("メディア確認作品").
			WithTitleKana("めでぃあかくにんさくひん").
			WithTitleEn("Media Check Work").
			WithMedia(2).
			Build()

		repo := repository.NewWorkRepository(query.New(db)).WithTx(tx)
		items, err := repo.ListForDB(ctx, repository.DBWorkListParams{
			Page:    1,
			PerPage: 100,
		})
		if err != nil {
			t.Fatalf("ListForDB() error = %v", err)
		}

		var target *model.Work
		for _, item := range items {
			if item.ID == workID {
				target = item
				break
			}
		}
		if target == nil {
			t.Fatalf("ListForDB() did not return the built work (id=%d)", int64(workID))
		}
		if target.TitleKana == nil || *target.TitleKana != "めでぃあかくにんさくひん" {
			t.Errorf("TitleKana = %v, want %q", target.TitleKana, "めでぃあかくにんさくひん")
		}
		if target.TitleEn != "Media Check Work" {
			t.Errorf("TitleEn = %q, want %q", target.TitleEn, "Media Check Work")
		}
		if target.Media != 2 {
			t.Errorf("Media = %d, want 2", target.Media)
		}
	})

	t.Run("正常系: sc_tid / mal_anime_id がマッピングされる", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTx(t)
		ctx := context.Background()

		withIDs := testutil.NewWorkBuilder(t, tx).
			WithTitle("外部サービスあり").
			WithScTid(3524).
			WithMalAnimeID(20).
			Build()
		withoutIDs := testutil.NewWorkBuilder(t, tx).WithTitle("外部サービスなし").Build()

		repo := repository.NewWorkRepository(query.New(db)).WithTx(tx)
		items, err := repo.ListForDB(ctx, repository.DBWorkListParams{
			Page:    1,
			PerPage: 100,
		})
		if err != nil {
			t.Fatalf("ListForDB() error = %v", err)
		}

		byID := make(map[model.WorkID]*model.Work, len(items))
		for _, item := range items {
			byID[item.ID] = item
		}

		got := byID[withIDs]
		if got == nil {
			t.Fatalf("ListForDB() did not return the built work (id=%d)", int64(withIDs))
		}
		if got.ScTid == nil || *got.ScTid != 3524 {
			t.Errorf("ScTid = %v, want 3524", got.ScTid)
		}
		if got.MalAnimeID == nil || *got.MalAnimeID != 20 {
			t.Errorf("MalAnimeID = %v, want 20", got.MalAnimeID)
		}

		empty := byID[withoutIDs]
		if empty == nil {
			t.Fatalf("ListForDB() did not return the built work (id=%d)", int64(withoutIDs))
		}
		if empty.ScTid != nil || empty.MalAnimeID != nil {
			t.Errorf("ScTid / MalAnimeID = %v / %v, want nil / nil", empty.ScTid, empty.MalAnimeID)
		}
	})

	t.Run("正常系: title_kana が空文字列なら TitleKana は nil になる", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTx(t)
		ctx := context.Background()

		workID := testutil.NewWorkBuilder(t, tx).WithTitle("ふりがななし作品").Build()

		repo := repository.NewWorkRepository(query.New(db)).WithTx(tx)
		items, err := repo.ListForDB(ctx, repository.DBWorkListParams{
			Page:    1,
			PerPage: 100,
		})
		if err != nil {
			t.Fatalf("ListForDB() error = %v", err)
		}

		var target *model.Work
		for _, item := range items {
			if item.ID == workID {
				target = item
				break
			}
		}
		if target == nil {
			t.Fatalf("ListForDB() did not return the built work (id=%d)", int64(workID))
		}
		if target.TitleKana != nil {
			t.Errorf("TitleKana = %v, want nil", *target.TitleKana)
		}
	})

	t.Run("正常系: 削除済み (deleted_at) 作品は除外され、非公開 (unpublished_at) 作品は残る", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTx(t)
		ctx := context.Background()

		archivedAt := time.Now()
		publishedID := testutil.NewWorkBuilder(t, tx).WithTitle("公開作品").Build()
		archivedID := testutil.NewWorkBuilder(t, tx).WithTitle("非公開作品").WithUnpublishedAt(archivedAt).Build()
		testutil.NewWorkBuilder(t, tx).WithTitle("削除作品").WithDeletedAt(time.Now()).Build()

		repo := repository.NewWorkRepository(query.New(db)).WithTx(tx)
		items, err := repo.ListForDB(ctx, repository.DBWorkListParams{
			Page:    1,
			PerPage: 100,
		})
		if err != nil {
			t.Fatalf("ListForDB() error = %v", err)
		}

		byID := make(map[model.WorkID]*model.Work, len(items))
		for _, item := range items {
			if item.Title == "削除作品" {
				t.Error("deleted work should not be returned")
			}
			byID[item.ID] = item
		}

		// The published work is returned with both state timestamps NULL.
		//
		// [Ja] 公開作品は両方の状態タイムスタンプが NULL で返る。
		published := byID[publishedID]
		if published == nil {
			t.Fatalf("ListForDB() did not return the published work (id=%d)", int64(publishedID))
		}
		if published.UnpublishedAt != nil || published.DeletedAt != nil {
			t.Errorf("published work state = {UnpublishedAt: %v, DeletedAt: %v}, want both nil", published.UnpublishedAt, published.DeletedAt)
		}

		// The archived work stays in the list (Rails without_deleted) with unpublished_at mapped.
		//
		// [Ja] 非公開作品は一覧に残り (Rails without_deleted)、unpublished_at がマッピングされる。
		archived := byID[archivedID]
		if archived == nil {
			t.Fatalf("ListForDB() did not return the archived work (id=%d)", int64(archivedID))
		}
		if archived.UnpublishedAt == nil {
			t.Error("archived work UnpublishedAt should be populated, got nil")
		}
		if archived.DeletedAt != nil {
			t.Errorf("archived work DeletedAt = %v, want nil", archived.DeletedAt)
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

		// A work whose only episode is archived (unpublished_at) counts as having no
		// episodes, matching the Rails scope Work.with_no_episodes. The dormant
		// episodes.status of such a row stays 'published', so reading it would hide
		// the work from this filter.
		//
		// [Ja] 唯一のエピソードが非公開 (unpublished_at) の作品は、Rails の
		// Work.with_no_episodes と同じくエピソードが無い作品として扱う。この行の休眠
		// episodes.status は 'published' のままなので、それを読むと本フィルタから漏れる。
		workUnpublishedEp := testutil.NewWorkBuilder(t, tx).WithTitle("非公開エピソードのみ").Build()
		testutil.NewEpisodeBuilder(t, tx, workUnpublishedEp).WithUnpublishedAt(time.Now()).Build()

		// A work whose only episode is soft-deleted (deleted_at) also counts as having no episodes.
		//
		// [Ja] 唯一のエピソードが削除済み (deleted_at) の作品も、エピソードが無い作品として扱う。
		workDeletedEp := testutil.NewWorkBuilder(t, tx).WithTitle("削除エピソードのみ").Build()
		testutil.NewEpisodeBuilder(t, tx, workDeletedEp).WithDeletedAt(time.Now()).Build()

		// A live episode whose dormant status alone says 'deleted' still counts as an
		// episode here, since the filter reads the timestamps and not episodes.status.
		//
		// [Ja] 休眠 status だけが 'deleted' の生きたエピソードも、ここではエピソードとして
		// 数える (絞り込みが読むのは timestamps であり episodes.status ではないため)。
		workDormantStatusEp := testutil.NewWorkBuilder(t, tx).WithTitle("休眠 status だけ削除").Build()
		dormantStatusEpID := testutil.NewEpisodeBuilder(t, tx, workDormantStatusEp).Build()
		if _, err := tx.Exec(`UPDATE episodes SET status = 'deleted' WHERE id = $1`, int64(dormantStatusEpID)); err != nil {
			t.Fatalf("休眠 status の更新に失敗: %v", err)
		}

		repo := repository.NewWorkRepository(query.New(db)).WithTx(tx)
		items, err := repo.ListForDB(ctx, repository.DBWorkListParams{
			FilterNoEpisodes: true,
			Page:             1,
			PerPage:          100,
		})
		if err != nil {
			t.Fatalf("ListForDB() error = %v", err)
		}

		titles := make(map[string]bool, len(items))
		for _, item := range items {
			titles[item.Title] = true
		}

		for _, title := range []string{"エピソードなし", "非公開エピソードのみ", "削除エピソードのみ"} {
			if !titles[title] {
				t.Errorf("work %q should be returned", title)
			}
		}
		for _, title := range []string{"エピソードあり", "エピソード不要", "休眠 status だけ削除"} {
			if titles[title] {
				t.Errorf("work %q should not be returned", title)
			}
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

	t.Run("正常系: 放送予定未登録フィルタ", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTx(t)
		ctx := context.Background()

		channelID := testutil.NewChannelBuilder(t, tx).Build()

		// A work with an active slot (deleted_at / unpublished_at NULL) is excluded by the filter.
		//
		// [Ja] 有効な放送枠 (deleted_at / unpublished_at が NULL) を持つ作品はフィルタで除外される。
		workWithSlot := testutil.NewWorkBuilder(t, tx).WithTitle("放送予定あり").Build()
		if _, err := tx.Exec(
			`INSERT INTO slots (work_id, channel_id, started_at, created_at, updated_at) VALUES ($1, $2, NOW(), NOW(), NOW())`,
			int64(workWithSlot), channelID,
		); err != nil {
			t.Fatalf("放送枠の投入に失敗: %v", err)
		}

		// A work whose only slot is soft-deleted counts as having no slots.
		//
		// [Ja] 唯一の放送枠がソフト削除済みの作品は「放送予定なし」とみなされる。
		workWithDeletedSlot := testutil.NewWorkBuilder(t, tx).WithTitle("放送予定削除済み").Build()
		if _, err := tx.Exec(
			`INSERT INTO slots (work_id, channel_id, started_at, deleted_at, created_at, updated_at) VALUES ($1, $2, NOW(), NOW(), NOW(), NOW())`,
			int64(workWithDeletedSlot), channelID,
		); err != nil {
			t.Fatalf("削除済み放送枠の投入に失敗: %v", err)
		}

		// A work whose only slot is unpublished (unpublished_at set) also counts as having no slots.
		//
		// [Ja] 唯一の放送枠が非公開 (unpublished_at がセット) の作品も「放送予定なし」とみなされる。
		workWithUnpublishedSlot := testutil.NewWorkBuilder(t, tx).WithTitle("放送予定非公開").Build()
		if _, err := tx.Exec(
			`INSERT INTO slots (work_id, channel_id, started_at, unpublished_at, created_at, updated_at) VALUES ($1, $2, NOW(), NOW(), NOW(), NOW())`,
			int64(workWithUnpublishedSlot), channelID,
		); err != nil {
			t.Fatalf("非公開放送枠の投入に失敗: %v", err)
		}

		workWithoutSlot := testutil.NewWorkBuilder(t, tx).WithTitle("放送予定なし").Build()

		repo := repository.NewWorkRepository(query.New(db)).WithTx(tx)
		items, err := repo.ListForDB(ctx, repository.DBWorkListParams{
			FilterNoSlots: true,
			Page:          1,
			PerPage:       100,
		})
		if err != nil {
			t.Fatalf("ListForDB() error = %v", err)
		}

		byID := make(map[model.WorkID]bool, len(items))
		for _, item := range items {
			byID[item.ID] = true
		}
		if byID[workWithSlot] {
			t.Error("放送予定ありの作品は除外されるべき")
		}
		if !byID[workWithDeletedSlot] {
			t.Error("削除済み放送枠しか持たない作品は含まれるべき")
		}
		if !byID[workWithUnpublishedSlot] {
			t.Error("非公開放送枠しか持たない作品は含まれるべき")
		}
		if !byID[workWithoutSlot] {
			t.Error("放送予定なしの作品は含まれるべき")
		}
	})

	t.Run("正常系: リリース時期の複数選択フィルタ", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTx(t)
		ctx := context.Background()

		spring2024 := testutil.NewWorkBuilder(t, tx).WithTitle("2024春").WithSeason(2024, testutil.SeasonSpring).Build()
		summer2024 := testutil.NewWorkBuilder(t, tx).WithTitle("2024夏").WithSeason(2024, testutil.SeasonSummer).Build()
		winter2023 := testutil.NewWorkBuilder(t, tx).WithTitle("2023冬").WithSeason(2023, testutil.SeasonWinter).Build()
		noSeason := testutil.NewWorkBuilder(t, tx).WithTitle("シーズンなし").WithNoSeason().Build()

		repo := repository.NewWorkRepository(query.New(db)).WithTx(tx)
		// Select two disjoint (year, season) pairs; only works matching one of them stay.
		//
		// [Ja] 交わらない 2 つの (年, 季節) ペアを選択し、いずれかに一致する作品だけが残る。
		items, err := repo.ListForDB(ctx, repository.DBWorkListParams{
			SeasonYears: []int32{2024, 2023},
			SeasonNames: []int32{testutil.SeasonSpring, testutil.SeasonWinter},
			Page:        1,
			PerPage:     100,
		})
		if err != nil {
			t.Fatalf("ListForDB() error = %v", err)
		}

		byID := make(map[model.WorkID]bool, len(items))
		for _, item := range items {
			byID[item.ID] = true
		}
		if !byID[spring2024] {
			t.Error("2024春は選択したペアに一致するので含まれるべき")
		}
		if !byID[winter2023] {
			t.Error("2023冬は選択したペアに一致するので含まれるべき")
		}
		if byID[summer2024] {
			t.Error("2024夏は選択したペアに一致しないので除外されるべき")
		}
		if byID[noSeason] {
			t.Error("シーズンなしの作品は除外されるべき")
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

	t.Run("境界値: 最大ページ番号でも OFFSET がオーバーフローしない", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTx(t)
		ctx := context.Background()

		repo := repository.NewWorkRepository(query.New(db)).WithTx(tx)
		got, err := repo.ListForDB(ctx, repository.DBWorkListParams{
			Page:    math.MaxInt32,
			PerPage: 100,
		})
		if err != nil {
			t.Fatalf("ListForDB() error = %v", err)
		}
		if len(got) != 0 {
			t.Errorf("len(got) = %d, want 0", len(got))
		}
	})
}

// TestWorkRepository_CountForDB はDB管理画面用の作品数取得をテスト
func TestWorkRepository_CountForDB(t *testing.T) {
	t.Parallel()

	t.Run("正常系: 削除済み (deleted_at) 作品はカウントから除外される", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTx(t)
		ctx := context.Background()

		// A distinctive season isolates the count from works that GetTestDB-based
		// tests commit to the shared DB, letting us assert an exact count instead of
		// a loose lower bound.
		//
		// [Ja] GetTestDB を使うテストが共有 DB にコミットする作品からカウントを隔離する
		// ため固有のシーズンを使い、緩い下限ではなく正確な件数をアサートできるようにする。
		const isolatedYear = 1901
		testutil.NewWorkBuilder(t, tx).WithTitle("公開作品A").WithSeason(isolatedYear, testutil.SeasonSpring).Build()
		testutil.NewWorkBuilder(t, tx).WithTitle("公開作品B").WithSeason(isolatedYear, testutil.SeasonSpring).Build()
		testutil.NewWorkBuilder(t, tx).WithTitle("削除作品").WithSeason(isolatedYear, testutil.SeasonSpring).WithDeletedAt(time.Now()).Build()

		repo := repository.NewWorkRepository(query.New(db)).WithTx(tx)
		count, err := repo.CountForDB(ctx, repository.DBWorkListParams{
			SeasonYears: []int32{isolatedYear},
			SeasonNames: []int32{testutil.SeasonSpring},
		})
		if err != nil {
			t.Fatalf("CountForDB() error = %v", err)
		}

		// Only the two published works count; the deleted (deleted_at) work is excluded.
		//
		// [Ja] 公開作品 2 件だけがカウントされ、削除済み (deleted_at) 作品は除外される。
		if count != 2 {
			t.Errorf("CountForDB() = %d, want 2", count)
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

	t.Run("正常系: エピソード未登録フィルタのカウントも旧層の timestamps で決まる", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTx(t)
		ctx := context.Background()

		// The count must use the same filter as the list. A distinctive season isolates
		// it from works that GetTestDB-based tests commit to the shared DB.
		//
		// [Ja] 件数は一覧と同じ絞り込みを使う。GetTestDB を使うテストが共有 DB にコミット
		// する作品から隔離するため固有のシーズンを使う。
		const isolatedYear = 1902
		newWork := func(title string) model.WorkID {
			return testutil.NewWorkBuilder(t, tx).WithTitle(title).WithSeason(isolatedYear, testutil.SeasonSpring).Build()
		}

		newWork("エピソードなし")
		testutil.NewEpisodeBuilder(t, tx, newWork("非公開エピソードのみ")).WithUnpublishedAt(time.Now()).Build()
		testutil.NewEpisodeBuilder(t, tx, newWork("削除エピソードのみ")).WithDeletedAt(time.Now()).Build()
		testutil.NewEpisodeBuilder(t, tx, newWork("エピソードあり")).Build()

		repo := repository.NewWorkRepository(query.New(db)).WithTx(tx)
		count, err := repo.CountForDB(ctx, repository.DBWorkListParams{
			FilterNoEpisodes: true,
			SeasonYears:      []int32{isolatedYear},
			SeasonNames:      []int32{testutil.SeasonSpring},
		})
		if err != nil {
			t.Fatalf("CountForDB() error = %v", err)
		}

		if count != 3 {
			t.Errorf("CountForDB() = %d, want 3 (公開エピソードを持つ作品だけを除外)", count)
		}
	})

	t.Run("正常系: リリース時期の複数選択がカウントにも適用される", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTx(t)
		ctx := context.Background()

		testutil.NewWorkBuilder(t, tx).WithTitle("2024春").WithSeason(2024, testutil.SeasonSpring).Build()
		testutil.NewWorkBuilder(t, tx).WithTitle("2024夏").WithSeason(2024, testutil.SeasonSummer).Build()

		repo := repository.NewWorkRepository(query.New(db)).WithTx(tx)
		count, err := repo.CountForDB(ctx, repository.DBWorkListParams{
			SeasonYears: []int32{2024},
			SeasonNames: []int32{testutil.SeasonSpring},
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

// TestWorkRepository_GetForEditByID verifies the edit-form loader reads every
// column and returns (nil, nil) for a nonexistent ID.
//
// [Ja] TestWorkRepository_GetForEditByID は編集フォーム用ローダーが全カラムを読み込み、
// 存在しないIDでは (nil, nil) を返すことを検証する。
func TestWorkRepository_GetForEditByID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewWorkRepository(queries)
	ctx := context.Background()

	workID := testutil.NewWorkBuilder(t, tx).
		WithTitle("編集ロードテスト").
		WithSeason(2024, testutil.SeasonAutumn).
		Build()
	if _, err := tx.Exec(`
		UPDATE works SET
			title_kana = 'へんしゅうろーどてすと',
			title_alter = '別名',
			title_en = 'Edit Load Test',
			title_alter_en = 'Alt EN',
			media = 1,
			official_site_url = 'https://example.dev/ja',
			official_site_url_en = 'https://example.dev/en',
			wikipedia_url = 'https://ja.wikipedia.example.dev',
			wikipedia_url_en = 'https://en.wikipedia.example.dev',
			twitter_username = 'handle',
			twitter_hashtag = 'tag',
			sc_tid = 123,
			mal_anime_id = 456,
			synopsis = 'あらすじ',
			synopsis_source = '出典',
			synopsis_en = 'Synopsis',
			synopsis_source_en = 'Source',
			manual_episodes_count = 12,
			start_episode_raw_number = 2.5,
			no_episodes = true,
			started_on = '2024-10-01',
			ended_on = '2024-12-20'
		WHERE id = $1
	`, int64(workID)); err != nil {
		t.Fatalf("works のフィールド設定に失敗: %v", err)
	}

	t.Run("全フィールドを読み込む", func(t *testing.T) {
		work, err := repo.GetForEditByID(ctx, workID)
		if err != nil {
			t.Fatalf("GetForEditByID() error = %v", err)
		}
		if work == nil {
			t.Fatal("work should not be nil")
		}
		if work.ID != workID {
			t.Errorf("ID = %d, want %d", work.ID, workID)
		}
		if work.Title != "編集ロードテスト" {
			t.Errorf("Title = %q", work.Title)
		}
		if work.TitleKana == nil || *work.TitleKana != "へんしゅうろーどてすと" {
			t.Errorf("TitleKana = %v", work.TitleKana)
		}
		if work.TitleAlter != "別名" {
			t.Errorf("TitleAlter = %q", work.TitleAlter)
		}
		if work.TitleEn != "Edit Load Test" {
			t.Errorf("TitleEn = %q", work.TitleEn)
		}
		if work.TitleAlterEn != "Alt EN" {
			t.Errorf("TitleAlterEn = %q", work.TitleAlterEn)
		}
		if work.Media != 1 {
			t.Errorf("Media = %d, want 1", work.Media)
		}
		if work.OfficialSiteURL != "https://example.dev/ja" {
			t.Errorf("OfficialSiteURL = %q", work.OfficialSiteURL)
		}
		if work.OfficialSiteURLEn != "https://example.dev/en" {
			t.Errorf("OfficialSiteURLEn = %q", work.OfficialSiteURLEn)
		}
		if work.WikipediaURL != "https://ja.wikipedia.example.dev" {
			t.Errorf("WikipediaURL = %q", work.WikipediaURL)
		}
		if work.WikipediaURLEn != "https://en.wikipedia.example.dev" {
			t.Errorf("WikipediaURLEn = %q", work.WikipediaURLEn)
		}
		if work.TwitterUsername == nil || *work.TwitterUsername != "handle" {
			t.Errorf("TwitterUsername = %v", work.TwitterUsername)
		}
		if work.TwitterHashtag == nil || *work.TwitterHashtag != "tag" {
			t.Errorf("TwitterHashtag = %v", work.TwitterHashtag)
		}
		if work.ScTid == nil || *work.ScTid != 123 {
			t.Errorf("ScTid = %v", work.ScTid)
		}
		if work.MalAnimeID == nil || *work.MalAnimeID != 456 {
			t.Errorf("MalAnimeID = %v", work.MalAnimeID)
		}
		if work.Synopsis != "あらすじ" {
			t.Errorf("Synopsis = %q", work.Synopsis)
		}
		if work.SynopsisSource != "出典" {
			t.Errorf("SynopsisSource = %q", work.SynopsisSource)
		}
		if work.SynopsisEn != "Synopsis" {
			t.Errorf("SynopsisEn = %q", work.SynopsisEn)
		}
		if work.SynopsisSourceEn != "Source" {
			t.Errorf("SynopsisSourceEn = %q", work.SynopsisSourceEn)
		}
		if work.ManualEpisodesCount == nil || *work.ManualEpisodesCount != 12 {
			t.Errorf("ManualEpisodesCount = %v", work.ManualEpisodesCount)
		}
		if work.StartEpisodeRawNumber != 2.5 {
			t.Errorf("StartEpisodeRawNumber = %v, want 2.5", work.StartEpisodeRawNumber)
		}
		if !work.NoEpisodes {
			t.Error("NoEpisodes = false, want true")
		}
		if work.SeasonYear == nil || *work.SeasonYear != 2024 {
			t.Errorf("SeasonYear = %v, want 2024", work.SeasonYear)
		}
		if work.SeasonName == nil || *work.SeasonName != testutil.SeasonAutumn {
			t.Errorf("SeasonName = %v, want %d", work.SeasonName, testutil.SeasonAutumn)
		}
		if work.StartedOn == nil || work.StartedOn.Year() != 2024 || work.StartedOn.Month() != 10 || work.StartedOn.Day() != 1 {
			t.Errorf("StartedOn = %v, want 2024-10-01", work.StartedOn)
		}
		if work.EndedOn == nil || work.EndedOn.Year() != 2024 || work.EndedOn.Month() != 12 || work.EndedOn.Day() != 20 {
			t.Errorf("EndedOn = %v, want 2024-12-20", work.EndedOn)
		}
		if work.NumberFormatID != nil {
			t.Errorf("NumberFormatID = %v, want nil", work.NumberFormatID)
		}
	})

	t.Run("存在しないIDは (nil, nil)", func(t *testing.T) {
		work, err := repo.GetForEditByID(ctx, model.WorkID(999999999))
		if err != nil {
			t.Fatalf("GetForEditByID() error = %v", err)
		}
		if work != nil {
			t.Errorf("work = %v, want nil", work)
		}
	})
}

// TestWorkRepository_GetForArchiveByID はアーカイブ確認画面用の作品取得をテスト
func TestWorkRepository_GetForArchiveByID(t *testing.T) {
	t.Parallel()

	t.Run("公開中の作品を取得できる", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTx(t)
		repo := repository.NewWorkRepository(query.New(db).WithTx(tx))

		workID := testutil.NewWorkBuilder(t, tx).WithTitle("アーカイブ確認作品").Build()

		work, err := repo.GetForArchiveByID(context.Background(), workID)
		if err != nil {
			t.Fatalf("GetForArchiveByID() error = %v", err)
		}
		if work == nil {
			t.Fatal("work should not be nil")
		}
		if work.ID != workID {
			t.Errorf("work.ID = %d, want %d", work.ID, workID)
		}
		if work.Title != "アーカイブ確認作品" {
			t.Errorf("work.Title = %q, want %q", work.Title, "アーカイブ確認作品")
		}
		if work.UnpublishedAt != nil {
			t.Errorf("work.UnpublishedAt = %v, want nil", work.UnpublishedAt)
		}
		if work.DeletedAt != nil {
			t.Errorf("work.DeletedAt = %v, want nil", work.DeletedAt)
		}
	})

	t.Run("非公開・削除の状態を読み取れる", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTx(t)
		repo := repository.NewWorkRepository(query.New(db).WithTx(tx))

		now := time.Now()
		archivedID := testutil.NewWorkBuilder(t, tx).WithTitle("非公開作品").WithUnpublishedAt(now).Build()
		deletedID := testutil.NewWorkBuilder(t, tx).WithTitle("削除作品").WithDeletedAt(now).Build()

		archived, err := repo.GetForArchiveByID(context.Background(), archivedID)
		if err != nil {
			t.Fatalf("GetForArchiveByID() error = %v", err)
		}
		if archived == nil || archived.UnpublishedAt == nil {
			t.Fatalf("archived work should carry UnpublishedAt, got %+v", archived)
		}
		if archived.DerivedStatus() != model.WorkStatusArchived {
			t.Errorf("archived.DerivedStatus() = %q, want archived", archived.DerivedStatus())
		}

		deleted, err := repo.GetForArchiveByID(context.Background(), deletedID)
		if err != nil {
			t.Fatalf("GetForArchiveByID() error = %v", err)
		}
		if deleted == nil || deleted.DeletedAt == nil {
			t.Fatalf("deleted work should carry DeletedAt, got %+v", deleted)
		}
		if deleted.DerivedStatus() != model.WorkStatusDeleted {
			t.Errorf("deleted.DerivedStatus() = %q, want deleted", deleted.DerivedStatus())
		}
	})

	t.Run("存在しないIDは (nil, nil)", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTx(t)
		repo := repository.NewWorkRepository(query.New(db).WithTx(tx))

		work, err := repo.GetForArchiveByID(context.Background(), model.WorkID(999999999))
		if err != nil {
			t.Fatalf("GetForArchiveByID() error = %v", err)
		}
		if work != nil {
			t.Errorf("work = %v, want nil", work)
		}
	})
}

func TestWorkRepository_GetForEpisodeListByID(t *testing.T) {
	t.Parallel()

	t.Run("見出しとサブナビに必要なカラムを取得できる", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTx(t)
		repo := repository.NewWorkRepository(query.New(db).WithTx(tx))

		workID := testutil.NewWorkBuilder(t, tx).
			WithTitle("エピソード一覧の親作品").
			WithNoEpisodes(true).
			WithManualEpisodesCount(12).
			Build()

		listWork, err := repo.GetForEpisodeListByID(context.Background(), workID)
		if err != nil {
			t.Fatalf("GetForEpisodeListByID() error = %v", err)
		}
		if listWork == nil {
			t.Fatal("listWork should not be nil")
		}
		if listWork.Work.ID != workID {
			t.Errorf("Work.ID = %d, want %d", listWork.Work.ID, workID)
		}
		if listWork.Work.Title != "エピソード一覧の親作品" {
			t.Errorf("Work.Title = %q, want %q", listWork.Work.Title, "エピソード一覧の親作品")
		}
		if !listWork.Work.NoEpisodes {
			t.Error("Work.NoEpisodes = false, want true")
		}
		if listWork.Work.ManualEpisodesCount == nil || *listWork.Work.ManualEpisodesCount != 12 {
			t.Errorf("Work.ManualEpisodesCount = %v, want 12", listWork.Work.ManualEpisodesCount)
		}
	})

	// The auto-generation notice reports the published episode count and how far the
	// auto-generation could number them, so both values must exclude the episodes and slots
	// the Rails only_kept scope drops and must ignore other works' rows.
	//
	// [Ja] 自動生成の案内は公開中のエピソード数と自動生成が到達する話数を報告するため、
	// 両方の値は Rails の only_kept スコープが落とすエピソード・スロットを除外し、
	// 他作品の行を数えないこと。
	t.Run("自動生成の案内は有効なエピソードとスロットだけを集計する", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTx(t)
		repo := repository.NewWorkRepository(query.New(db).WithTx(tx))

		workID := testutil.NewWorkBuilder(t, tx).WithTitle("件数集計の親作品").Build()
		testutil.NewEpisodeBuilder(t, tx, workID).WithNumber("第1話").Build()
		testutil.NewEpisodeBuilder(t, tx, workID).WithNumber("第2話").Build()
		testutil.NewEpisodeBuilder(t, tx, workID).WithNumber("第3話").WithUnpublishedAt(time.Now()).Build()
		testutil.NewEpisodeBuilder(t, tx, workID).WithNumber("第4話").WithDeletedAt(time.Now()).Build()

		channelID := testutil.NewChannelBuilder(t, tx).Build()
		newSlot := func() *testutil.SlotBuilder {
			return testutil.NewSlotBuilder(t, tx).WithWorkID(workID).WithChannelID(channelID)
		}
		newSlot().WithNumber(5).Build()
		newSlot().WithNumber(9).WithUnpublishedAt(time.Now()).Build()
		newSlot().WithNumber(12).WithDeletedAt(time.Now()).Build()

		otherWorkID := testutil.NewWorkBuilder(t, tx).WithTitle("別の作品").Build()
		testutil.NewEpisodeBuilder(t, tx, otherWorkID).WithNumber("第1話").Build()
		testutil.NewSlotBuilder(t, tx).WithWorkID(otherWorkID).WithChannelID(channelID).WithNumber(24).Build()

		listWork, err := repo.GetForEpisodeListByID(context.Background(), workID)
		if err != nil {
			t.Fatalf("GetForEpisodeListByID() error = %v", err)
		}
		if listWork == nil {
			t.Fatal("listWork should not be nil")
		}
		if listWork.PublishedEpisodeCount != 2 {
			t.Errorf("PublishedEpisodeCount = %d, want 2", listWork.PublishedEpisodeCount)
		}
		if listWork.MaxGeneratableEpisodeNumber != 5 {
			t.Errorf("MaxGeneratableEpisodeNumber = %d, want 5", listWork.MaxGeneratableEpisodeNumber)
		}
	})

	// slots.number is nullable, and MAX skips those rows. The Rails notice instead reads the
	// first slot ordered by number descending, which PostgreSQL sorts NULLs first for, so it
	// reports 0 for the same work. Keeping the highest reachable number is the deliberate
	// divergence the query documents, so pin it here.
	//
	// [Ja] slots.number は NULL を許容し、MAX はその行を飛ばす。一方 Rails の案内は number
	// 降順の先頭スロットを読むため、PostgreSQL の NULLS FIRST により同じ作品で 0 を報告する。
	// 到達できる最大話数を保つのはクエリのコメントに記した意図的な差異なので、ここで固定する。
	t.Run("話数未設定のスロットは生成可能話数を 0 に落とさない", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTx(t)
		repo := repository.NewWorkRepository(query.New(db).WithTx(tx))

		workID := testutil.NewWorkBuilder(t, tx).WithTitle("話数未設定スロットを持つ作品").Build()

		channelID := testutil.NewChannelBuilder(t, tx).Build()
		testutil.NewSlotBuilder(t, tx).WithWorkID(workID).WithChannelID(channelID).WithNumber(5).Build()
		testutil.NewSlotBuilder(t, tx).WithWorkID(workID).WithChannelID(channelID).Build()

		listWork, err := repo.GetForEpisodeListByID(context.Background(), workID)
		if err != nil {
			t.Fatalf("GetForEpisodeListByID() error = %v", err)
		}
		if listWork == nil {
			t.Fatal("listWork should not be nil")
		}
		if listWork.MaxGeneratableEpisodeNumber != 5 {
			t.Errorf("MaxGeneratableEpisodeNumber = %d, want 5", listWork.MaxGeneratableEpisodeNumber)
		}
	})

	// A work with no slot at all has nothing to auto-generate from, which the notice reports
	// as zero rather than leaving the maximum episode number absent.
	//
	// [Ja] スロットを 1 つも持たない作品には自動生成の元が無い。案内は最大話数を
	// 欠落させず 0 として報告する。
	t.Run("スロットが無い作品の生成可能話数は 0", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTx(t)
		repo := repository.NewWorkRepository(query.New(db).WithTx(tx))

		workID := testutil.NewWorkBuilder(t, tx).WithTitle("スロットなしの作品").Build()

		listWork, err := repo.GetForEpisodeListByID(context.Background(), workID)
		if err != nil {
			t.Fatalf("GetForEpisodeListByID() error = %v", err)
		}
		if listWork == nil {
			t.Fatal("listWork should not be nil")
		}
		if listWork.MaxGeneratableEpisodeNumber != 0 {
			t.Errorf("MaxGeneratableEpisodeNumber = %d, want 0", listWork.MaxGeneratableEpisodeNumber)
		}
		if listWork.Work.ManualEpisodesCount != nil {
			t.Errorf("Work.ManualEpisodesCount = %v, want nil", listWork.Work.ManualEpisodesCount)
		}
	})

	t.Run("削除済みの作品は (nil, nil)", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTx(t)
		repo := repository.NewWorkRepository(query.New(db).WithTx(tx))

		workID := testutil.NewWorkBuilder(t, tx).
			WithTitle("削除済みの親作品").
			WithDeletedAt(time.Now()).
			Build()

		listWork, err := repo.GetForEpisodeListByID(context.Background(), workID)
		if err != nil {
			t.Fatalf("GetForEpisodeListByID() error = %v", err)
		}
		if listWork != nil {
			t.Errorf("listWork = %v, want nil", listWork)
		}
	})

	t.Run("存在しないIDは (nil, nil)", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTx(t)
		repo := repository.NewWorkRepository(query.New(db).WithTx(tx))

		listWork, err := repo.GetForEpisodeListByID(context.Background(), model.WorkID(999999999))
		if err != nil {
			t.Fatalf("GetForEpisodeListByID() error = %v", err)
		}
		if listWork != nil {
			t.Errorf("listWork = %v, want nil", listWork)
		}
	})
}

func TestWorkRepository_GetForEpisodeFormByID_ManualCreationState(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewWorkRepository(query.New(db).WithTx(tx))
	ctx := context.Background()

	filledWorkID := testutil.NewWorkBuilder(t, tx).
		WithTitle("予定話数到達").
		WithManualEpisodesCount(2).
		Build()
	testutil.NewEpisodeBuilder(t, tx, filledWorkID).WithNumber("第1話").Build()
	testutil.NewEpisodeBuilder(t, tx, filledWorkID).WithNumber("第2話").Build()
	testutil.NewEpisodeBuilder(t, tx, filledWorkID).WithNumber("非公開").WithUnpublishedAt(time.Now()).Build()
	testutil.NewEpisodeBuilder(t, tx, filledWorkID).WithNumber("削除済み").WithDeletedAt(time.Now()).Build()

	filledWork, err := repo.GetForEpisodeFormByID(ctx, filledWorkID)
	if err != nil {
		t.Fatalf("予定話数到達 GetForEpisodeFormByID() error = %v", err)
	}
	if filledWork == nil || !filledWork.ManualCreationState.EpisodesFilled {
		t.Fatalf("予定話数到達の ManualCreationState = %+v, want EpisodesFilled", filledWork)
	}
	if filledWork.ManualCreationState.Allowed() {
		t.Error("予定話数到達作品の手動作成が許可されています")
	}

	slotWorkID := testutil.NewWorkBuilder(t, tx).WithTitle("放送枠あり").Build()
	channelID := testutil.NewChannelBuilder(t, tx).Build()
	testutil.NewSlotBuilder(t, tx).
		WithWorkID(slotWorkID).
		WithChannelID(channelID).
		WithStartedAt(time.Now()).
		Build()

	slotWork, err := repo.GetForEpisodeFormByID(ctx, slotWorkID)
	if err != nil {
		t.Fatalf("放送枠あり GetForEpisodeFormByID() error = %v", err)
	}
	if slotWork == nil || !slotWork.ManualCreationState.SlotsExist {
		t.Fatalf("放送枠ありの ManualCreationState = %+v, want SlotsExist", slotWork)
	}
	if slotWork.ManualCreationState.Allowed() {
		t.Error("放送枠あり作品の手動作成が許可されています")
	}

	plainWorkID := testutil.NewWorkBuilder(t, tx).WithTitle("制限なし").Build()
	plainWork, err := repo.GetForEpisodeFormByID(ctx, plainWorkID)
	if err != nil {
		t.Fatalf("制限なし GetForEpisodeFormByID() error = %v", err)
	}
	if plainWork == nil || !plainWork.ManualCreationState.Allowed() {
		t.Fatalf("制限なしの ManualCreationState = %+v, want Allowed", plainWork)
	}
}

// TestWorkRepository_GetForEpisodeCreateByID covers what an episode bulk create reads about
// its parent work: the anchors the new rows are numbered from, and the state that refuses the
// submit. The anchors aggregate every episode of the work, including the unpublished and the
// deleted ones, so a work whose episodes were archived does not hand out sort_numbers that
// are already taken.
//
// The manual-creation state is asserted here as well as on the form query because the two
// queries carry the same predicates separately. This is the copy that refuses a POST, so a
// drift on this side alone would disable the form while letting a direct submit through.
//
// [Ja] TestWorkRepository_GetForEpisodeCreateByID はエピソード一括作成が親作品から読み取る
// もの (新規行の採番の起点と、送信を却下する状態) を検証する。起点は非公開・削除済みを含む
// 作品のすべてのエピソードを集計するため、エピソードを非公開にした作品で既に使われている
// sort_number を振り直さない。
//
// 手動作成の状態をフォーム用クエリと併せてここでも検証するのは、2 つのクエリが同じ述語を
// 別々に持っているため。POST を却下するのはこちらの写しであり、この側だけがずれると、
// フォームは無効化されるのに直接の送信は通ってしまう。
func TestWorkRepository_GetForEpisodeCreateByID(t *testing.T) {
	t.Parallel()

	t.Run("採番の起点に非公開・削除済みのエピソードも数える", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTx(t)
		repo := repository.NewWorkRepository(query.New(db).WithTx(tx))

		workID := testutil.NewWorkBuilder(t, tx).WithTitle("一括作成の親作品").Build()
		insertEpisodeWithSortNumber(t, tx, workID, "第1話", 100)
		latestID := insertEpisodeWithSortNumber(t, tx, workID, "第2話", 200)
		testutil.NewEpisodeBuilder(t, tx, workID).WithNumber("第3話").WithUnpublishedAt(time.Now()).Build()
		testutil.NewEpisodeBuilder(t, tx, workID).WithNumber("第4話").WithDeletedAt(time.Now()).Build()

		otherWorkID := testutil.NewWorkBuilder(t, tx).WithTitle("別の作品").Build()
		insertEpisodeWithSortNumber(t, tx, otherWorkID, "第1話", 9999)

		createWork, err := repo.GetForEpisodeCreateByID(context.Background(), workID)
		if err != nil {
			t.Fatalf("GetForEpisodeCreateByID() error = %v", err)
		}
		if createWork == nil {
			t.Fatal("createWork should not be nil")
		}
		if createWork.Work.ID != workID {
			t.Errorf("Work.ID = %d, want %d", createWork.Work.ID, workID)
		}
		if createWork.EpisodeCount != 4 {
			t.Errorf("EpisodeCount = %d, want 4", createWork.EpisodeCount)
		}
		// The builder fixes sort_number at 0, so the two rows inserted with explicit values
		// are the ones that decide the latest episode.
		//
		// [Ja] ビルダーは sort_number を 0 で固定するため、最新のエピソードを決めるのは
		// 明示的な値で挿入した 2 行になる。
		if createWork.LatestEpisode == nil {
			t.Fatal("LatestEpisode should not be nil")
		}
		if createWork.LatestEpisode.ID != latestID {
			t.Errorf("LatestEpisode.ID = %d, want %d", createWork.LatestEpisode.ID, latestID)
		}
		if createWork.LatestEpisode.SortNumber != 200 {
			t.Errorf("LatestEpisode.SortNumber = %d, want 200", createWork.LatestEpisode.SortNumber)
		}
	})

	t.Run("エピソードが無い作品には起点が無い", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTx(t)
		repo := repository.NewWorkRepository(query.New(db).WithTx(tx))

		workID := testutil.NewWorkBuilder(t, tx).WithTitle("エピソードなしの親作品").Build()

		createWork, err := repo.GetForEpisodeCreateByID(context.Background(), workID)
		if err != nil {
			t.Fatalf("GetForEpisodeCreateByID() error = %v", err)
		}
		if createWork == nil {
			t.Fatal("createWork should not be nil")
		}
		if createWork.EpisodeCount != 0 {
			t.Errorf("EpisodeCount = %d, want 0", createWork.EpisodeCount)
		}
		if createWork.LatestEpisode != nil {
			t.Errorf("LatestEpisode = %+v, want nil", createWork.LatestEpisode)
		}
		// An unmapped work reports no anime, which is what makes the create skip the
		// reference-model writes.
		//
		// [Ja] 未マッピングの作品は anime を持たないと報告する。作成が参照モデルへの書き込みを
		// 飛ばすのはこの値による。
		if createWork.Work.AnimeID != nil {
			t.Errorf("Work.AnimeID = %v, want nil", createWork.Work.AnimeID)
		}
	})

	t.Run("マッピング済みの作品は anime_id を返す", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTx(t)
		repo := repository.NewWorkRepository(query.New(db).WithTx(tx))

		animeID := insertEpisodeSyncParentAnime(t, tx)
		workID := insertEpisodeSyncWork(t, tx, sql.NullInt64{Int64: int64(animeID), Valid: true})

		createWork, err := repo.GetForEpisodeCreateByID(context.Background(), workID)
		if err != nil {
			t.Fatalf("GetForEpisodeCreateByID() error = %v", err)
		}
		if createWork == nil {
			t.Fatal("createWork should not be nil")
		}
		if createWork.Work.AnimeID == nil || *createWork.Work.AnimeID != animeID {
			t.Errorf("Work.AnimeID = %v, want %d", createWork.Work.AnimeID, int64(animeID))
		}
	})

	t.Run("手動作成の制限を判定する", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTx(t)
		repo := repository.NewWorkRepository(query.New(db).WithTx(tx))
		ctx := context.Background()

		filledWorkID := testutil.NewWorkBuilder(t, tx).
			WithTitle("作成側: 予定話数到達").
			WithManualEpisodesCount(2).
			Build()
		testutil.NewEpisodeBuilder(t, tx, filledWorkID).WithNumber("第1話").Build()
		testutil.NewEpisodeBuilder(t, tx, filledWorkID).WithNumber("第2話").Build()

		// The predicate counts kept episodes only, so a work stays creatable while its
		// unpublished and deleted rows are what would fill the expected count.
		//
		// [Ja] 述語は公開中のエピソードだけを数えるため、予定話数を満たすのが非公開・削除済み
		// の行である作品は作成可能なままになる。
		keptOnlyWorkID := testutil.NewWorkBuilder(t, tx).
			WithTitle("作成側: 公開中だけでは予定話数に届かない").
			WithManualEpisodesCount(3).
			Build()
		testutil.NewEpisodeBuilder(t, tx, keptOnlyWorkID).WithNumber("第1話").Build()
		testutil.NewEpisodeBuilder(t, tx, keptOnlyWorkID).WithNumber("第2話").Build()
		testutil.NewEpisodeBuilder(t, tx, keptOnlyWorkID).WithNumber("第3話").WithUnpublishedAt(time.Now()).Build()
		testutil.NewEpisodeBuilder(t, tx, keptOnlyWorkID).WithNumber("第4話").WithDeletedAt(time.Now()).Build()

		slotWorkID := testutil.NewWorkBuilder(t, tx).WithTitle("作成側: 放送枠あり").Build()
		channelID := testutil.NewChannelBuilder(t, tx).Build()
		testutil.NewSlotBuilder(t, tx).
			WithWorkID(slotWorkID).
			WithChannelID(channelID).
			WithStartedAt(time.Now()).
			Build()

		plainWorkID := testutil.NewWorkBuilder(t, tx).WithTitle("作成側: 制限なし").Build()

		tests := []struct {
			name            string
			workID          model.WorkID
			wantRestriction model.ManualEpisodeCreationRestriction
		}{
			{name: "予定話数到達", workID: filledWorkID, wantRestriction: model.ManualEpisodeCreationEpisodesFilled},
			{name: "公開中だけでは予定話数に届かない", workID: keptOnlyWorkID, wantRestriction: model.ManualEpisodeCreationAllowed},
			{name: "放送枠あり", workID: slotWorkID, wantRestriction: model.ManualEpisodeCreationSlotsExist},
			{name: "制限なし", workID: plainWorkID, wantRestriction: model.ManualEpisodeCreationAllowed},
		}

		// The cases share one transaction, so they run in sequence rather than in parallel.
		//
		// [Ja] 各ケースは 1 つのトランザクションを共有するため、並行ではなく順に実行する。
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				createWork, err := repo.GetForEpisodeCreateByID(ctx, tt.workID)
				if err != nil {
					t.Fatalf("GetForEpisodeCreateByID() error = %v", err)
				}
				if createWork == nil {
					t.Fatal("createWork should not be nil")
				}
				if got := createWork.ManualCreationState.Restriction(); got != tt.wantRestriction {
					t.Errorf("Restriction() = %q, want %q", got, tt.wantRestriction)
				}
			})
		}
	})

	t.Run("削除済みの作品は (nil, nil)", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTx(t)
		repo := repository.NewWorkRepository(query.New(db).WithTx(tx))

		workID := testutil.NewWorkBuilder(t, tx).
			WithTitle("削除済みの一括作成対象").
			WithDeletedAt(time.Now()).
			Build()

		createWork, err := repo.GetForEpisodeCreateByID(context.Background(), workID)
		if err != nil {
			t.Fatalf("GetForEpisodeCreateByID() error = %v", err)
		}
		if createWork != nil {
			t.Errorf("createWork = %v, want nil", createWork)
		}
	})
}

// insertEpisodeWithSortNumber creates an episode with an explicit sort_number. The shared
// EpisodeBuilder fixes it at 0, and the create anchors depend on the ordering it defines.
//
// [Ja] insertEpisodeWithSortNumber は sort_number を明示してエピソードを作成する。共有の
// EpisodeBuilder はこれを 0 で固定するが、作成の起点は sort_number が定める順序に依存する
// ため。
func insertEpisodeWithSortNumber(t *testing.T, tx *sql.Tx, workID model.WorkID, number string, sortNumber int32) model.EpisodeID {
	t.Helper()

	var id int64
	if err := tx.QueryRow(`
		INSERT INTO episodes (work_id, number, sort_number, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id
	`, int64(workID), number, sortNumber).Scan(&id); err != nil {
		t.Fatalf("エピソードの作成に失敗: %v", err)
	}
	return model.EpisodeID(id)
}

// TestWorkRepository_UpdateUnpublishedAt は unpublished_at の設定・クリアをテスト
func TestWorkRepository_UpdateUnpublishedAt(t *testing.T) {
	t.Parallel()

	t.Run("時刻を設定すると非公開になる", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTx(t)
		repo := repository.NewWorkRepository(query.New(db).WithTx(tx))
		ctx := context.Background()

		workID := testutil.NewWorkBuilder(t, tx).WithTitle("非公開化対象").Build()

		now := time.Now()
		if err := repo.UpdateUnpublishedAt(ctx, workID, &now); err != nil {
			t.Fatalf("UpdateUnpublishedAt() error = %v", err)
		}

		work, err := repo.GetForArchiveByID(ctx, workID)
		if err != nil || work == nil {
			t.Fatalf("GetForArchiveByID() work=%v err=%v", work, err)
		}
		if work.UnpublishedAt == nil {
			t.Error("work.UnpublishedAt should be set, got nil")
		}
	})

	t.Run("nil を渡すと再公開される", func(t *testing.T) {
		t.Parallel()
		db, tx := testutil.SetupTx(t)
		repo := repository.NewWorkRepository(query.New(db).WithTx(tx))
		ctx := context.Background()

		workID := testutil.NewWorkBuilder(t, tx).WithTitle("再公開対象").WithUnpublishedAt(time.Now()).Build()

		if err := repo.UpdateUnpublishedAt(ctx, workID, nil); err != nil {
			t.Fatalf("UpdateUnpublishedAt() error = %v", err)
		}

		work, err := repo.GetForArchiveByID(ctx, workID)
		if err != nil || work == nil {
			t.Fatalf("GetForArchiveByID() work=%v err=%v", work, err)
		}
		if work.UnpublishedAt != nil {
			t.Errorf("work.UnpublishedAt = %v, want nil after clearing", work.UnpublishedAt)
		}
	})
}

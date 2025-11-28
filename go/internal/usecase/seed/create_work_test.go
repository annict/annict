package seed

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"testing"

	"github.com/annict/annict/internal/seed"
	"github.com/annict/annict/internal/testutil"
)

// TestCreateWorkUsecase_ExecuteBatch はExecuteBatchメソッドのテスト
func TestCreateWorkUsecase_ExecuteBatch(t *testing.T) {
	// テストDBをセットアップ（トランザクションは各サブテストで作成）
	db, _ := testutil.SetupTestDB(t)

	// Usecaseを作成
	uc := NewCreateWorkUsecase(db)

	// SeasonNameのヘルパー変数
	seasonSpring := seed.SeasonSpring
	seasonWinter := seed.SeasonWinter
	seasonYear2024 := int32(2024)
	seasonYear2023 := int32(2023)

	// テストケース
	tests := []struct {
		name      string
		works     []CreateWorkParams
		wantCount int
		wantErr   bool
	}{
		{
			name: "正常系: 3つの作品を作成",
			works: []CreateWorkParams{
				{
					Title:           "魔法の冒険",
					TitleKana:       "まほうのぼうけん",
					Media:           seed.MediaTV,
					OfficialSiteURL: "https://example.com/work1",
					SeasonYear:      &seasonYear2024,
					SeasonName:      &seasonSpring,
				},
				{
					Title:           "伝説の戦士",
					TitleKana:       "でんせつのせんし",
					Media:           seed.MediaOVA,
					OfficialSiteURL: "https://example.com/work2",
					SeasonYear:      &seasonYear2023,
					SeasonName:      &seasonWinter,
				},
				{
					Title:           "不思議な学園",
					TitleKana:       "ふしぎながくえん",
					Media:           seed.MediaMovie,
					OfficialSiteURL: "",
					SeasonYear:      &seasonYear2024,
					SeasonName:      &seasonSpring,
				},
			},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name: "正常系: 1つの作品を作成",
			works: []CreateWorkParams{
				{
					Title:           "シングル作品",
					TitleKana:       "",
					Media:           seed.MediaWeb,
					OfficialSiteURL: "",
					SeasonYear:      &seasonYear2024,
					SeasonName:      &seasonSpring,
				},
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "正常系: SeasonYearとSeasonNameがnil",
			works: []CreateWorkParams{
				{
					Title:           "シーズン未定作品",
					TitleKana:       "",
					Media:           seed.MediaTV,
					OfficialSiteURL: "",
					SeasonYear:      nil,
					SeasonName:      nil,
				},
			},
			wantCount: 1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 各サブテストで新しいトランザクションを作成
			_, tx := testutil.SetupTestDB(t)
			defer tx.Rollback()

			ctx := context.Background()

			// ExecuteBatchWithTxを実行（テスト用トランザクションを使用）
			results, err := uc.ExecuteBatchWithTx(ctx, tx, tt.works, nil)

			// エラーチェック
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteBatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// 結果の数をチェック
			if len(results) != tt.wantCount {
				t.Errorf("ExecuteBatch() got %d results, want %d", len(results), tt.wantCount)
				return
			}

			// 各作品が正しく作成されたか検証
			for i, result := range results {
				// WorkIDが取得できていることを確認
				if result.WorkID == 0 {
					t.Errorf("result[%d].WorkID is 0", i)
				}

				// worksテーブルにレコードが作成されたか確認
				var title, titleKana string
				var media, seasonName sql.NullInt32
				var seasonYear sql.NullInt32
				var officialSiteURL string
				err := tx.QueryRow(`
					SELECT title, title_kana, media, official_site_url, season_year, season_name
					FROM works WHERE id = $1
				`, result.WorkID).Scan(&title, &titleKana, &media, &officialSiteURL, &seasonYear, &seasonName)
				if err != nil {
					t.Errorf("作品レコードの取得に失敗: %v", err)
					continue
				}

				// タイトルが正しいか確認
				if title != tt.works[i].Title {
					t.Errorf("title = %v, want %v", title, tt.works[i].Title)
				}

				// title_kanaが正しいか確認
				if titleKana != tt.works[i].TitleKana {
					t.Errorf("title_kana = %v, want %v", titleKana, tt.works[i].TitleKana)
				}

				// MediaTypeが正しく変換されているか確認
				var expectedMedia int32
				switch tt.works[i].Media {
				case seed.MediaTV:
					expectedMedia = 0
				case seed.MediaOVA:
					expectedMedia = 1
				case seed.MediaMovie:
					expectedMedia = 2
				case seed.MediaWeb:
					expectedMedia = 3
				}
				if media.Valid && media.Int32 != expectedMedia {
					t.Errorf("media = %v, want %v", media.Int32, expectedMedia)
				}

				// official_site_urlが正しいか確認
				if officialSiteURL != tt.works[i].OfficialSiteURL {
					t.Errorf("official_site_url = %v, want %v", officialSiteURL, tt.works[i].OfficialSiteURL)
				}

				// season_yearが正しいか確認
				if tt.works[i].SeasonYear != nil {
					if !seasonYear.Valid {
						t.Errorf("season_year should be valid")
					} else if seasonYear.Int32 != *tt.works[i].SeasonYear {
						t.Errorf("season_year = %v, want %v", seasonYear.Int32, *tt.works[i].SeasonYear)
					}
				} else {
					if seasonYear.Valid {
						t.Errorf("season_year should be NULL")
					}
				}

				// season_nameが正しいか確認（enum値）
				if tt.works[i].SeasonName != nil {
					if !seasonName.Valid {
						t.Errorf("season_name should be valid")
					} else {
						var expectedSeasonName int32
						switch *tt.works[i].SeasonName {
						case seed.SeasonWinter:
							expectedSeasonName = 1
						case seed.SeasonSpring:
							expectedSeasonName = 2
						case seed.SeasonSummer:
							expectedSeasonName = 3
						case seed.SeasonAutumn:
							expectedSeasonName = 4
						}
						if seasonName.Int32 != expectedSeasonName {
							t.Errorf("season_name = %v, want %v", seasonName.Int32, expectedSeasonName)
						}
					}
				} else {
					if seasonName.Valid {
						t.Errorf("season_name should be NULL")
					}
				}
			}
		})
	}
}

// TestCreateWorkUsecase_MediaTypeConversion はMediaType変換のテスト
func TestCreateWorkUsecase_MediaTypeConversion(t *testing.T) {
	// テストDBとトランザクションをセットアップ
	db, tx := testutil.SetupTestDB(t)
	defer tx.Rollback()

	// Usecaseを作成
	uc := NewCreateWorkUsecase(db)

	ctx := context.Background()

	seasonYear := int32(2024)
	seasonSpring := seed.SeasonSpring

	// 各MediaTypeを作成
	mediaTypes := []struct {
		mediaType seed.MediaType
		expected  int32
	}{
		{seed.MediaTV, 0},
		{seed.MediaOVA, 1},
		{seed.MediaMovie, 2},
		{seed.MediaWeb, 3},
	}

	for _, mt := range mediaTypes {
		works := []CreateWorkParams{
			{
				Title:           fmt.Sprintf("作品 %s", mt.mediaType),
				TitleKana:       "",
				Media:           mt.mediaType,
				OfficialSiteURL: "",
				SeasonYear:      &seasonYear,
				SeasonName:      &seasonSpring,
			},
		}

		results, err := uc.ExecuteBatchWithTx(ctx, tx, works, nil)
		if err != nil {
			t.Fatalf("ExecuteBatch() error = %v", err)
		}

		if len(results) != 1 {
			t.Fatalf("ExecuteBatch() returned %d results, want 1", len(results))
		}

		// DBからmediaを取得
		var media int32
		err = tx.QueryRow("SELECT media FROM works WHERE id = $1", results[0].WorkID).Scan(&media)
		if err != nil {
			t.Fatalf("mediaの取得に失敗: %v", err)
		}

		// 期待値と比較
		if media != mt.expected {
			t.Errorf("MediaType %s: media = %v, want %v", mt.mediaType, media, mt.expected)
		}
	}
}

// TestCreateWorkUsecase_LargeBatch は大量の作品作成のテスト
func TestCreateWorkUsecase_LargeBatch(t *testing.T) {
	// -short フラグが指定されている場合はスキップ（CI用）
	if testing.Short() {
		t.Skip("長時間テストのため -short フラグでスキップします")
	}

	// テストDBとトランザクションをセットアップ
	db, tx := testutil.SetupTestDB(t)
	defer tx.Rollback()

	// Usecaseを作成
	uc := NewCreateWorkUsecase(db)

	ctx := context.Background()

	// 2500の作品を作成（バッチサイズ1000を超えるケース）
	workCount := 2500
	works := make([]CreateWorkParams, workCount)
	seasonYear := int32(2024)
	seasonSpring := seed.SeasonSpring

	for i := 0; i < workCount; i++ {
		works[i] = CreateWorkParams{
			Title:           fmt.Sprintf("作品 %d", i+1),
			TitleKana:       "",
			Media:           seed.MediaTV,
			OfficialSiteURL: "",
			SeasonYear:      &seasonYear,
			SeasonName:      &seasonSpring,
		}
	}

	// ExecuteBatchWithTxを実行（テスト用トランザクションを使用）
	results, err := uc.ExecuteBatchWithTx(ctx, tx, works, nil)
	if err != nil {
		t.Fatalf("ExecuteBatch() error = %v", err)
	}

	// 結果の数をチェック
	if len(results) != workCount {
		t.Errorf("ExecuteBatch() got %d results, want %d", len(results), workCount)
	}

	// 最初と最後の作品を検証
	if len(results) > 0 {
		if results[0].WorkID == 0 {
			t.Error("results[0].WorkID is 0")
		}
		if results[len(results)-1].WorkID == 0 {
			t.Error("results[last].WorkID is 0")
		}
	}
}

// TestGenerateRandomWorkParams はランダム作品パラメータ生成のテスト
func TestGenerateRandomWorkParams(t *testing.T) {
	// 固定シードでランダム生成器を初期化
	r := rand.New(rand.NewSource(12345))

	// ランダムパラメータを生成
	params := GenerateRandomWorkParams(r)

	// タイトルが空でないことを確認
	if params.Title == "" {
		t.Error("Title should not be empty")
	}

	// MediaTypeが有効な値であることを確認
	validMedia := false
	for _, media := range seed.AllMediaTypes {
		if params.Media == media {
			validMedia = true
			break
		}
	}
	if !validMedia {
		t.Errorf("Invalid Media: %v", params.Media)
	}

	// SeasonYearが2020〜2025の範囲であることを確認
	if params.SeasonYear != nil {
		if *params.SeasonYear < 2020 || *params.SeasonYear > 2025 {
			t.Errorf("SeasonYear out of range: %v", *params.SeasonYear)
		}
	}

	// SeasonNameが有効な値であることを確認
	if params.SeasonName != nil {
		validSeason := false
		for _, season := range seed.AllSeasons {
			if *params.SeasonName == season {
				validSeason = true
				break
			}
		}
		if !validSeason {
			t.Errorf("Invalid SeasonName: %v", *params.SeasonName)
		}
	}
}

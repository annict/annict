package seed

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	"github.com/annict/annict/internal/seed"
	"github.com/schollz/progressbar/v3"
)

// CreateWorkParams 作品作成のパラメータ
type CreateWorkParams struct {
	Title           string
	TitleKana       string
	Media           seed.MediaType
	OfficialSiteURL string
	SeasonYear      *int32
	SeasonName      *seed.SeasonName
}

// CreateWorkResult 作品作成の結果
type CreateWorkResult struct {
	WorkID int64
}

// CreateWorkUsecase 作品生成Usecase（シード専用、バルクインサート対応）
type CreateWorkUsecase struct {
	db *sql.DB
}

// NewCreateWorkUsecase 新しいCreateWorkUsecaseを作成
func NewCreateWorkUsecase(db *sql.DB) *CreateWorkUsecase {
	return &CreateWorkUsecase{
		db: db,
	}
}

// ExecuteBatch 複数の作品をバッチで作成します
// 1000件ごとにコミットしてパフォーマンスを最適化します
func (uc *CreateWorkUsecase) ExecuteBatch(ctx context.Context, works []CreateWorkParams, progressBar *progressbar.ProgressBar) ([]CreateWorkResult, error) {
	return uc.executeBatchWithTx(ctx, nil, works, progressBar)
}

// ExecuteBatchWithTx 複数の作品をバッチで作成します（テスト用：既存トランザクションを使用）
// txがnilの場合は内部でトランザクションを作成します
func (uc *CreateWorkUsecase) ExecuteBatchWithTx(ctx context.Context, tx *sql.Tx, works []CreateWorkParams, progressBar *progressbar.ProgressBar) ([]CreateWorkResult, error) {
	return uc.executeBatchWithTx(ctx, tx, works, progressBar)
}

// executeBatchWithTx 内部実装：トランザクションを受け取るか新規作成する
func (uc *CreateWorkUsecase) executeBatchWithTx(ctx context.Context, existingTx *sql.Tx, works []CreateWorkParams, progressBar *progressbar.ProgressBar) ([]CreateWorkResult, error) {
	results := make([]CreateWorkResult, 0, len(works))

	// 既存トランザクションがある場合は、バッチサイズを無視して全件処理
	if existingTx != nil {
		// マルチ行INSERTのチャンクサイズ（100件ずつ）
		multiInsertChunkSize := 100
		for i := 0; i < len(works); i += multiInsertChunkSize {
			end := i + multiInsertChunkSize
			if end > len(works) {
				end = len(works)
			}
			chunk := works[i:end]

			// マルチ行INSERTで作成
			chunkResults, err := uc.createMultipleWorks(ctx, existingTx, chunk)
			if err != nil {
				return nil, fmt.Errorf("作品マルチ行INSERT エラー: %w", err)
			}
			results = append(results, chunkResults...)

			// 進捗表示を更新
			if progressBar != nil {
				progressBar.Add(len(chunk))
			}
		}
		return results, nil
	}

	// 既存トランザクションがない場合は、5000件ごとにコミット
	commitBatchSize := 5000
	multiInsertChunkSize := 500

	for i := 0; i < len(works); i += commitBatchSize {
		end := i + commitBatchSize
		if end > len(works) {
			end = len(works)
		}
		batch := works[i:end]

		// トランザクション開始
		tx, err := uc.db.BeginTx(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("トランザクション開始エラー: %w", err)
		}
		defer tx.Rollback()

		// バッチ内の作品を100件ずつマルチ行INSERTで作成
		for j := 0; j < len(batch); j += multiInsertChunkSize {
			chunkEnd := j + multiInsertChunkSize
			if chunkEnd > len(batch) {
				chunkEnd = len(batch)
			}
			chunk := batch[j:chunkEnd]

			// マルチ行INSERTで作成
			chunkResults, err := uc.createMultipleWorks(ctx, tx, chunk)
			if err != nil {
				return nil, fmt.Errorf("作品マルチ行INSERT エラー: %w", err)
			}
			results = append(results, chunkResults...)

			// 進捗表示を更新
			if progressBar != nil {
				progressBar.Add(len(chunk))
			}
		}

		// コミット
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("トランザクションコミットエラー: %w", err)
		}
	}

	return results, nil
}

// createMultipleWorks 複数の作品をマルチ行INSERTで作成します（トランザクション内）
func (uc *CreateWorkUsecase) createMultipleWorks(ctx context.Context, tx *sql.Tx, worksList []CreateWorkParams) ([]CreateWorkResult, error) {
	if len(worksList) == 0 {
		return []CreateWorkResult{}, nil
	}

	// マルチ行INSERT用のクエリを構築
	queryBuilder := `INSERT INTO works (
		title, title_kana, media, official_site_url,
		wikipedia_url, season_year, season_name,
		watchers_count, episodes_count,
		created_at, updated_at
	) VALUES `

	// VALUES句とパラメータを構築
	values := []interface{}{}
	now := time.Now()

	for i, params := range worksList {
		if i > 0 {
			queryBuilder += ", "
		}

		// プレースホルダーの開始位置（各行は11個のパラメータ）
		offset := i * 11
		queryBuilder += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			offset+1, offset+2, offset+3, offset+4, offset+5, offset+6, offset+7, offset+8, offset+9, offset+10, offset+11)

		// MediaTypeをinteger値に変換（Rails enum互換）
		var mediaInt int
		switch params.Media {
		case seed.MediaTV:
			mediaInt = 0
		case seed.MediaOVA:
			mediaInt = 1
		case seed.MediaMovie:
			mediaInt = 2
		case seed.MediaWeb:
			mediaInt = 3
		default:
			mediaInt = 0 // デフォルトはTV
		}

		// SeasonNameをinteger値に変換（Rails enum互換）
		var seasonNameInt interface{}
		if params.SeasonName != nil {
			switch *params.SeasonName {
			case seed.SeasonWinter:
				seasonNameInt = 1
			case seed.SeasonSpring:
				seasonNameInt = 2
			case seed.SeasonSummer:
				seasonNameInt = 3
			case seed.SeasonAutumn:
				seasonNameInt = 4
			default:
				seasonNameInt = nil
			}
		} else {
			seasonNameInt = nil
		}

		// パラメータを追加
		values = append(values,
			params.Title,
			params.TitleKana,
			mediaInt,
			params.OfficialSiteURL,
			"", // wikipedia_url (デフォルト空文字)
			params.SeasonYear,
			seasonNameInt,
			0,   // watchers_count (デフォルト0)
			0,   // episodes_count (デフォルト0)
			now, // created_at
			now, // updated_at
		)
	}

	queryBuilder += " RETURNING id"

	// マルチ行INSERTを実行
	rows, err := tx.QueryContext(ctx, queryBuilder, values...)
	if err != nil {
		return nil, fmt.Errorf("worksテーブルへのマルチ行INSERT エラー: %w", err)
	}
	defer rows.Close()

	// 挿入されたIDを取得
	results := make([]CreateWorkResult, 0, len(worksList))
	for rows.Next() {
		var workID int64
		if err := rows.Scan(&workID); err != nil {
			return nil, fmt.Errorf("RETURNING id のスキャンエラー: %w", err)
		}
		results = append(results, CreateWorkResult{WorkID: workID})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("行の反復処理エラー: %w", err)
	}

	return results, nil
}

// createSingleWork 単一の作品を作成します（トランザクション内）
// 注意: この関数は後方互換性のために残していますが、
// パフォーマンスのためにcreateMultipleWorksの使用を推奨します
//
//lint:ignore U1000 後方互換性のために保持
func (uc *CreateWorkUsecase) createSingleWork(ctx context.Context, tx *sql.Tx, params CreateWorkParams) (*CreateWorkResult, error) {
	// デフォルト値の設定
	titleKana := params.TitleKana
	if titleKana == "" {
		titleKana = ""
	}

	officialSiteURL := params.OfficialSiteURL
	if officialSiteURL == "" {
		officialSiteURL = ""
	}

	// 作品を作成
	workID, err := uc.createWork(ctx, tx, params)
	if err != nil {
		return nil, fmt.Errorf("作品レコード作成エラー: %w", err)
	}

	return &CreateWorkResult{
		WorkID: workID,
	}, nil
}

// createWork worksテーブルにレコードを作成します
//
//lint:ignore U1000 後方互換性のために保持
func (uc *CreateWorkUsecase) createWork(ctx context.Context, tx *sql.Tx, params CreateWorkParams) (int64, error) {
	query := `
		INSERT INTO works (
			title, title_kana, media, official_site_url,
			wikipedia_url, season_year, season_name,
			watchers_count, episodes_count,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7,
			$8, $9,
			$10, $11
		) RETURNING id
	`

	// MediaTypeをinteger値に変換（Rails enum互換）
	var mediaInt int
	switch params.Media {
	case seed.MediaTV:
		mediaInt = 0
	case seed.MediaOVA:
		mediaInt = 1
	case seed.MediaMovie:
		mediaInt = 2
	case seed.MediaWeb:
		mediaInt = 3
	default:
		mediaInt = 0 // デフォルトはTV
	}

	// SeasonNameをinteger値に変換（Rails enum互換）
	var seasonNameInt interface{}
	if params.SeasonName != nil {
		switch *params.SeasonName {
		case seed.SeasonWinter:
			seasonNameInt = 1
		case seed.SeasonSpring:
			seasonNameInt = 2
		case seed.SeasonSummer:
			seasonNameInt = 3
		case seed.SeasonAutumn:
			seasonNameInt = 4
		default:
			seasonNameInt = nil
		}
	} else {
		seasonNameInt = nil
	}

	var workID int64
	err := tx.QueryRowContext(
		ctx,
		query,
		params.Title,
		params.TitleKana,
		mediaInt,
		params.OfficialSiteURL,
		"", // wikipedia_url (デフォルト空文字)
		params.SeasonYear,
		seasonNameInt,
		0,          // watchers_count (デフォルト0)
		0,          // episodes_count (デフォルト0)
		time.Now(), // created_at
		time.Now(), // updated_at
	).Scan(&workID)

	if err != nil {
		return 0, fmt.Errorf("worksテーブルへの挿入エラー: %w", err)
	}

	return workID, nil
}

// GenerateRandomWorkParams はランダムな作品パラメータを生成します
// シードデータ生成時に使用するヘルパー関数
func GenerateRandomWorkParams(r *rand.Rand) CreateWorkParams {
	title := seed.GenerateAnimeTitle(r)
	seasonYear := seed.GenerateSeasonYear(r)
	seasonName := seed.GenerateSeasonName(r)
	media := seed.GenerateMediaType(r, true) // 加重ランダム（TVアニメの出現率を高く）

	return CreateWorkParams{
		Title:           title,
		TitleKana:       "",                              // 空文字でOK（NOT NULL制約対応）
		Media:           media,                           // TV, OVA, Movie, Web
		OfficialSiteURL: "",                              // 空文字でOK（NOT NULL制約対応）
		SeasonYear:      &seasonYear,                     // 2020〜2025
		SeasonName:      (*seed.SeasonName)(&seasonName), // spring, summer, autumn, winter
	}
}

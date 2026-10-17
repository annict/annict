package usecase

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// newGetDBWorksUsecaseはテスト用トランザクション上にUseCaseを組み立てる。本
// UseCaseは読み取りのみで自らトランザクションを開かないためSetupTxを使う。
func newGetDBWorksUsecase(t *testing.T) (*GetDBWorksUsecase, *sql.Tx) {
	t.Helper()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)

	return NewGetDBWorksUsecase(repository.NewWorkRepository(queries)), tx
}

// 同じパッケージには作品をコミットするテストがあり、SetupTxのトランザクションからも
// それらの作品が見える。件数と並び順を厳密に検証するテストは、他のテストが使わない
// 年の組でリリース時期を絞り、自身が作成した作品だけを一覧の対象にする。

// assertDBWorkIDsは一覧の作品IDが期待した順序と件数で並んでいることを検証する。
func assertDBWorkIDs(t *testing.T, works []*model.Work, wantIDs []model.WorkID) {
	t.Helper()

	if len(works) != len(wantIDs) {
		t.Fatalf("len(Works) = %d、期待値 = %d", len(works), len(wantIDs))
	}
	for i, want := range wantIDs {
		if works[i].ID != want {
			t.Errorf("Works[%d].ID = %d、期待値 = %d", i, works[i].ID, want)
		}
	}
}

// TestGetDBWorksUsecase_Execute_ReturnsWorksInDescendingIDOrderは、作品を新しく
// 登録された順 (idの降順) に返し、総件数も返すことを検証する。
func TestGetDBWorksUsecase_Execute_ReturnsWorksInDescendingIDOrder(t *testing.T) {
	t.Parallel()

	uc, tx := newGetDBWorksUsecase(t)

	firstID := testutil.NewWorkBuilder(t, tx).WithTitle("作品1").WithSeason(1901, testutil.SeasonSpring).Build()
	secondID := testutil.NewWorkBuilder(t, tx).WithTitle("作品2").WithSeason(1901, testutil.SeasonSpring).Build()
	thirdID := testutil.NewWorkBuilder(t, tx).WithTitle("作品3").WithSeason(1901, testutil.SeasonSpring).Build()

	output, err := uc.Execute(context.Background(), GetDBWorksInput{
		SeasonYears: []int32{1901},
		SeasonNames: []int32{testutil.SeasonSpring},
		Page:        1,
		PerPage:     100,
	})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	assertDBWorkIDs(t, output.Works, []model.WorkID{thirdID, secondID, firstID})
	if output.TotalCount != 3 {
		t.Errorf("TotalCount = %d、期待値 = 3", output.TotalCount)
	}
}

// TestGetDBWorksUsecase_Execute_FiltersWorksは、各絞り込みのチェックボックスに
// 対応する入力が一覧を絞ることを検証する。
// リリース時期未登録の絞り込みは年の組と併用できず、他のテストがコミットした作品も
// 一覧に混ざるため、作成した作品の有無だけを検証する。
func TestGetDBWorksUsecase_Execute_FiltersWorks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input GetDBWorksInput
		// setupは絞り込みに一致する作品と一致しない作品を1件ずつ作成し、そのIDを返す。
		setup func(t *testing.T, tx *sql.Tx) (matched, unmatched model.WorkID)
	}{
		{
			name:  "エピソード未登録",
			input: GetDBWorksInput{FilterNoEpisodes: true},
			setup: func(t *testing.T, tx *sql.Tx) (model.WorkID, model.WorkID) {
				withEpisode := testutil.NewWorkBuilder(t, tx).WithTitle("エピソードあり").Build()
				testutil.NewEpisodeBuilder(t, tx, withEpisode).Build()
				return testutil.NewWorkBuilder(t, tx).WithTitle("エピソードなし").Build(), withEpisode
			},
		},
		{
			name:  "画像未設定",
			input: GetDBWorksInput{FilterNoImage: true},
			setup: func(t *testing.T, tx *sql.Tx) (model.WorkID, model.WorkID) {
				withImage, _ := testutil.CreateTestWorkWithImage(t, tx, "画像あり")
				return testutil.NewWorkBuilder(t, tx).WithTitle("画像なし").Build(), withImage
			},
		},
		{
			name:  "リリース時期未登録",
			input: GetDBWorksInput{FilterNoSeason: true},
			setup: func(t *testing.T, tx *sql.Tx) (model.WorkID, model.WorkID) {
				withSeason := testutil.NewWorkBuilder(t, tx).WithTitle("リリース時期あり").WithSeason(2024, testutil.SeasonSpring).Build()
				return testutil.NewWorkBuilder(t, tx).WithTitle("リリース時期なし").WithNoSeason().Build(), withSeason
			},
		},
		{
			name:  "放送予定未登録",
			input: GetDBWorksInput{FilterNoSlots: true},
			setup: func(t *testing.T, tx *sql.Tx) (model.WorkID, model.WorkID) {
				withSlot := testutil.NewWorkBuilder(t, tx).WithTitle("放送予定あり").Build()
				channelID := testutil.NewChannelBuilder(t, tx).Build()
				testutil.NewSlotBuilder(t, tx).WithWorkID(withSlot).WithChannelID(channelID).Build()
				return testutil.NewWorkBuilder(t, tx).WithTitle("放送予定なし").Build(), withSlot
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc, tx := newGetDBWorksUsecase(t)
			matched, unmatched := tt.setup(t, tx)

			input := tt.input
			input.Page = 1
			input.PerPage = 100
			output, err := uc.Execute(context.Background(), input)
			if err != nil {
				t.Fatalf("Execute()のエラー = %v", err)
			}

			byID := make(map[model.WorkID]bool, len(output.Works))
			for _, work := range output.Works {
				byID[work.ID] = true
			}
			if !byID[matched] {
				t.Errorf("絞り込みに一致する作品 (id=%d) が返されなかった", matched)
			}
			if byID[unmatched] {
				t.Errorf("絞り込みに一致しない作品 (id=%d) が返された", unmatched)
			}
		})
	}
}

// TestGetDBWorksUsecase_Execute_FiltersBySeasonPairsは、SeasonYears / SeasonNamesの
// (年, 季節) の組のいずれかに一致する作品だけを返すことを検証する。
// 年だけ、季節だけが一致する作品は組に一致しないため除外される。
func TestGetDBWorksUsecase_Execute_FiltersBySeasonPairs(t *testing.T) {
	t.Parallel()

	uc, tx := newGetDBWorksUsecase(t)

	winter1902 := testutil.NewWorkBuilder(t, tx).WithTitle("1902冬").WithSeason(1902, testutil.SeasonWinter).Build()
	testutil.NewWorkBuilder(t, tx).WithTitle("1902春").WithSeason(1902, testutil.SeasonSpring).Build()
	spring1903 := testutil.NewWorkBuilder(t, tx).WithTitle("1903春").WithSeason(1903, testutil.SeasonSpring).Build()
	testutil.NewWorkBuilder(t, tx).WithTitle("1903冬").WithSeason(1903, testutil.SeasonWinter).Build()
	testutil.NewWorkBuilder(t, tx).WithTitle("リリース時期なし").WithNoSeason().Build()

	output, err := uc.Execute(context.Background(), GetDBWorksInput{
		SeasonYears: []int32{1903, 1902},
		SeasonNames: []int32{testutil.SeasonSpring, testutil.SeasonWinter},
		Page:        1,
		PerPage:     100,
	})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	assertDBWorkIDs(t, output.Works, []model.WorkID{spring1903, winter1902})
	if output.TotalCount != 2 {
		t.Errorf("TotalCount = %d、期待値 = 2", output.TotalCount)
	}
}

// TestGetDBWorksUsecase_Execute_FiltersBySeasonNoneは、SeasonNamesの0が「季節未登録」
// を表し、その年でseason_nameがNULLの作品に一致することを検証する。季節ありの組と
// 同時に指定すると、いずれかの組に一致する作品を返す。
func TestGetDBWorksUsecase_Execute_FiltersBySeasonNone(t *testing.T) {
	t.Parallel()

	uc, tx := newGetDBWorksUsecase(t)

	yearOnly1902 := testutil.NewWorkBuilder(t, tx).WithTitle("1902季節未登録").WithSeasonYearOnly(1902).Build()
	testutil.NewWorkBuilder(t, tx).WithTitle("1902冬").WithSeason(1902, testutil.SeasonWinter).Build()
	spring1903 := testutil.NewWorkBuilder(t, tx).WithTitle("1903春").WithSeason(1903, testutil.SeasonSpring).Build()
	testutil.NewWorkBuilder(t, tx).WithTitle("1903季節未登録").WithSeasonYearOnly(1903).Build()
	testutil.NewWorkBuilder(t, tx).WithTitle("リリース時期なし").WithNoSeason().Build()

	output, err := uc.Execute(context.Background(), GetDBWorksInput{
		SeasonYears: []int32{1903, 1902},
		SeasonNames: []int32{testutil.SeasonSpring, 0},
		Page:        1,
		PerPage:     100,
	})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	assertDBWorkIDs(t, output.Works, []model.WorkID{spring1903, yearOnly1902})
	if output.TotalCount != 2 {
		t.Errorf("TotalCount = %d、期待値 = 2", output.TotalCount)
	}
}

// TestGetDBWorksUsecase_Execute_PaginatesWorksはPerPage / Pageが作品を1ページ分だけ
// 切り出す一方、TotalCountは絞り込み後の一覧対象すべてを報告し続けることを検証する。
// これによりページネーションが残りのページを描画できる。
func TestGetDBWorksUsecase_Execute_PaginatesWorks(t *testing.T) {
	t.Parallel()

	uc, tx := newGetDBWorksUsecase(t)

	firstID := testutil.NewWorkBuilder(t, tx).WithTitle("作品1").WithSeason(1904, testutil.SeasonSpring).Build()
	testutil.NewWorkBuilder(t, tx).WithTitle("作品2").WithSeason(1904, testutil.SeasonSummer).Build()
	secondID := testutil.NewWorkBuilder(t, tx).WithTitle("作品3").WithSeason(1904, testutil.SeasonSpring).Build()
	thirdID := testutil.NewWorkBuilder(t, tx).WithTitle("作品4").WithSeason(1904, testutil.SeasonSpring).Build()

	tests := []struct {
		name    string
		page    int32
		wantIDs []model.WorkID
	}{
		{name: "1ページ目", page: 1, wantIDs: []model.WorkID{thirdID, secondID}},
		{name: "2ページ目", page: 2, wantIDs: []model.WorkID{firstID}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := uc.Execute(context.Background(), GetDBWorksInput{
				SeasonYears: []int32{1904},
				SeasonNames: []int32{testutil.SeasonSpring},
				Page:        tt.page,
				PerPage:     2,
			})
			if err != nil {
				t.Fatalf("Execute()のエラー = %v", err)
			}

			assertDBWorkIDs(t, output.Works, tt.wantIDs)
			if output.TotalCount != 3 {
				t.Errorf("TotalCount = %d、期待値 = 3 (絞り込み後の総数)", output.TotalCount)
			}
		})
	}
}

// TestGetDBWorksUsecase_Execute_ReturnsListErrorは一覧の取得に失敗した場合、
// 出力を返さず、呼び出し元が元のエラーを判別できることを検証する。
func TestGetDBWorksUsecase_Execute_ReturnsListError(t *testing.T) {
	t.Parallel()

	uc, _ := newGetDBWorksUsecase(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	output, err := uc.Execute(ctx, GetDBWorksInput{Page: 1, PerPage: 100})
	if output != nil {
		t.Errorf("出力 = %v、期待値 = nil", output)
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("エラー = %v、期待値 = %v", err, context.Canceled)
	}
}

// countErrorTxは一覧取得を通常どおり実行し、総数取得のクエリだけをキャンセルする。
// 一覧取得はQueryContext、総数取得はQueryRowContextを使うため、QueryRowContextだけを
// 失敗させる。listCalledは一覧取得が実行されたことを記録し、テストが総数取得の分岐で
// 失敗したことを確かめられるようにする。
type countErrorTx struct {
	*sql.Tx
	listCalled bool
}

func (tx *countErrorTx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	tx.listCalled = true
	return tx.Tx.QueryContext(ctx, query, args...)
}

func (tx *countErrorTx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	canceledCtx, cancel := context.WithCancel(ctx)
	cancel()
	return tx.Tx.QueryRowContext(canceledCtx, query, args...)
}

// TestGetDBWorksUsecase_Execute_ReturnsCountErrorは一覧取得後に総数の取得に失敗した
// 場合、出力を返さず、呼び出し元が元のエラーを判別できることを検証する。
func TestGetDBWorksUsecase_Execute_ReturnsCountError(t *testing.T) {
	t.Parallel()

	_, tx := testutil.SetupTx(t)
	countTx := &countErrorTx{Tx: tx}
	uc := NewGetDBWorksUsecase(repository.NewWorkRepository(query.New(countTx)))

	output, err := uc.Execute(context.Background(), GetDBWorksInput{Page: 1, PerPage: 100})
	if !countTx.listCalled {
		t.Fatal("一覧取得が実行されなかった (総数取得より前の分岐で失敗している)")
	}
	if output != nil {
		t.Errorf("出力 = %v、期待値 = nil", output)
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("エラー = %v、期待値 = %v", err, context.Canceled)
	}
}

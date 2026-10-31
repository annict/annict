package usecase

import (
	"context"
	"database/sql"
	"sync/atomic"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// newCreateEpisodesFromSlotsUsecaseは共有テストDBに対して自動生成UseCaseを組み立てる。
// 本UseCaseは放送枠ごとに自前のトランザクションを開くため、テストはGetTestDBを使い、
// コミット済みの行がUseCaseの内側トランザクションから見えるようにする。
func newCreateEpisodesFromSlotsUsecase(db *sql.DB) *CreateEpisodesFromSlotsUsecase {
	queries := query.New(db)
	return NewCreateEpisodesFromSlotsUsecase(
		db,
		repository.NewSlotRepository(queries),
		repository.NewWorkRepository(queries),
		repository.NewEpisodeRepository(queries),
		repository.NewAnimeRepository(queries),
		repository.NewAnimeClassificationRepository(queries),
	)
}

var generationNowCounter atomic.Int64

// generationNowはテストごとに60日ずつ離れた実行時刻を返す。自動生成は実行時刻の前後7日の
// 放送枠をすべて拾うため、共有DBにコミットしたほかのテストの放送枠を、別のテストの実行が
// 拾わないようにする。
func generationNow() time.Time {
	n := generationNowCounter.Add(1)
	return time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(n) * 60 * 24 * time.Hour)
}

// generationProgramは自動生成のテストが放送枠を置く番組。
type generationProgram struct {
	t         *testing.T
	db        *sql.DB
	workID    model.WorkID
	channelID int64
	programID int64
}

// insertGenerationProgramは作品の番組を共有DBにコミットし、テスト終了時に番組とその
// 放送枠を消す。作品の後始末より先に走るよう、作品を作った後に呼ぶ。
func insertGenerationProgram(t *testing.T, db *sql.DB, workID model.WorkID, minimumEpisodeGeneratableNumber int32) *generationProgram {
	t.Helper()

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("番組作成トランザクションの開始に失敗: %v", err)
	}
	channelGroupID := testutil.NewChannelGroupBuilder(t, tx).Build()
	channelID := testutil.NewChannelBuilder(t, tx).WithChannelGroupID(channelGroupID).Build()
	programID := testutil.NewProgramBuilder(t, tx).
		WithChannelID(channelID).
		WithWorkID(workID).
		WithMinimumEpisodeGeneratableNumber(minimumEpisodeGeneratableNumber).
		Build()
	if err := tx.Commit(); err != nil {
		t.Fatalf("番組作成トランザクションのコミットに失敗: %v", err)
	}

	t.Cleanup(func() {
		statements := []struct {
			sql string
			arg int64
		}{
			{`DELETE FROM slots WHERE program_id = $1`, programID},
			{`DELETE FROM programs WHERE id = $1`, programID},
			{`DELETE FROM channels WHERE id = $1`, channelID},
			{`DELETE FROM channel_groups WHERE id = $1`, channelGroupID},
		}
		for _, statement := range statements {
			if _, err := db.Exec(statement.sql, statement.arg); err != nil {
				t.Logf("番組の後始末に失敗 (%s): %v", statement.sql, err)
			}
		}
	})

	return &generationProgram{t: t, db: db, workID: workID, channelID: channelID, programID: programID}
}

// insertSlotは番組の放送枠を共有DBにコミットする。configureでビルダーの設定を足す。
func (p *generationProgram) insertSlot(number int32, startedAt time.Time, configure func(*testutil.SlotBuilder)) model.SlotID {
	p.t.Helper()

	tx, err := p.db.Begin()
	if err != nil {
		p.t.Fatalf("放送枠作成トランザクションの開始に失敗: %v", err)
	}
	builder := testutil.NewSlotBuilder(p.t, tx).
		WithWorkID(p.workID).
		WithChannelID(p.channelID).
		WithProgramID(p.programID).
		WithNumber(number).
		WithStartedAt(startedAt)
	if configure != nil {
		configure(builder)
	}
	id := builder.Build()
	if err := tx.Commit(); err != nil {
		p.t.Fatalf("放送枠作成トランザクションのコミットに失敗: %v", err)
	}
	return id
}

// generationEpisodeInputは、自動生成のテストで作品に置く既存のエピソード。
func generationEpisodeInput(workID model.WorkID, rawNumber float64, sortNumber int32) syncEpisodeInput {
	in := defaultSyncEpisodeInput(workID)
	in.rawNumber = sql.NullFloat64{Float64: rawNumber, Valid: true}
	in.sortNumber = sortNumber
	return in
}

func readSlotEpisodeID(t *testing.T, db *sql.DB, slotID model.SlotID) sql.NullInt64 {
	t.Helper()

	var episodeID sql.NullInt64
	if err := db.QueryRow(`SELECT episode_id FROM slots WHERE id = $1`, int64(slotID)).Scan(&episodeID); err != nil {
		t.Fatalf("放送枠の読み込みに失敗: %v", err)
	}
	return episodeID
}

func countWorkEpisodes(t *testing.T, db *sql.DB, workID model.WorkID) int {
	t.Helper()

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM episodes WHERE work_id = $1`, int64(workID)).Scan(&count); err != nil {
		t.Fatalf("エピソード件数の読み込みに失敗: %v", err)
	}
	return count
}

// TestCreateEpisodesFromSlotsUsecase_Execute_CreatesEpisodeは、紐づけるエピソードが無い
// 放送枠に、animeと分類を両書きしたエピソードを作って紐づけることを検証する。対象の範囲外の
// 放送枠と、すでに紐づいた放送枠には触れない。
func TestCreateEpisodesFromSlotsUsecase_Execute_CreatesEpisode(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	ctx := context.Background()
	now := generationNow()

	workID, parentAnimeID := insertMappedCreateTargetWork(t, db)
	firstID := insertSyncEpisode(t, db, generationEpisodeInput(workID, 1, 100))
	program := insertGenerationProgram(t, db, workID, 1)
	program.insertSlot(1, now.Add(-24*time.Hour), func(b *testutil.SlotBuilder) { b.WithEpisodeID(firstID) })
	targetID := program.insertSlot(2, now.Add(6*24*time.Hour), nil)
	outsideID := program.insertSlot(3, now.Add(8*24*time.Hour), nil)

	output, err := newCreateEpisodesFromSlotsUsecase(db).Execute(ctx, CreateEpisodesFromSlotsInput{Now: now})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	if len(output.CreatedEpisodes) != 1 || output.LinkedSlotCount != 0 || len(output.SkippedSlots) != 0 || len(output.FailedSlots) != 0 {
		t.Fatalf("output = %+v、期待値 = 作成1件のみ", output)
	}
	created := output.CreatedEpisodes[0]
	if created.SlotID != targetID || created.WorkID != workID || created.Number != "第2話" || created.WorkTitle == "" {
		t.Errorf("CreatedEpisodes[0] = %+v、期待値 = 放送枠%d・作品%d・第2話", created, targetID, workID)
	}

	if got := readSlotEpisodeID(t, db, targetID); got.Int64 != int64(created.EpisodeID) {
		t.Errorf("対象の放送枠のepisode_id = %+v、期待値 = %d", got, created.EpisodeID)
	}
	if got := readSlotEpisodeID(t, db, outsideID); got.Valid {
		t.Errorf("範囲外の放送枠のepisode_id = %+v、期待値 = NULL", got)
	}

	episode := readCreatedEpisode(t, db, created.EpisodeID)
	if episode.rawNumber.Float64 != 2 || episode.sortNumber != 200 || episode.prevEpisodeID.Int64 != int64(firstID) || episode.title.Valid {
		t.Errorf("エピソード = %+v、期待値 = 2話・sort_number 200・前のエピソード%d・タイトル無し", episode, firstID)
	}
	if !episode.animeID.Valid {
		t.Fatal("episodes.anime_idが書かれていません")
	}

	classification, err := repository.NewAnimeClassificationRepository(query.New(db)).GetByAnimeID(ctx, model.AnimeID(episode.animeID.Int64))
	if err != nil || classification == nil {
		t.Fatalf("GetByAnimeID()のclassification = %v、エラー = %v", classification, err)
	}
	if classification.ParentAnimeID == nil || *classification.ParentAnimeID != parentAnimeID || classification.NumberText.String != "第2話" {
		t.Errorf("classification = %+v、期待値 = 親anime%dの下の「第2話」", classification, parentAnimeID)
	}

	var (
		episodesCount int32
		workTouched   bool
		activityCount int
	)
	if err := db.QueryRow(`SELECT episodes_count, updated_at >= created_at FROM works WHERE id = $1`, int64(workID)).Scan(&episodesCount, &workTouched); err != nil {
		t.Fatalf("作品の読み込みに失敗: %v", err)
	}
	if episodesCount != 1 || !workTouched {
		t.Errorf("作品 = (episodes_count %d, touch %v)、期待値 = (1, true)", episodesCount, workTouched)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM db_activities WHERE root_resource_type = 'Work' AND root_resource_id = $1`, int64(workID)).Scan(&activityCount); err != nil {
		t.Fatalf("DB活動の読み込みに失敗: %v", err)
	}
	if activityCount != 0 {
		t.Errorf("DB活動 = %d件、期待値 = 0", activityCount)
	}

	// 2回目の実行では対象の放送枠が紐づき済みのため、何もしない。
	again, err := newCreateEpisodesFromSlotsUsecase(db).Execute(ctx, CreateEpisodesFromSlotsInput{Now: now})
	if err != nil {
		t.Fatalf("2回目のExecute()のエラー = %v", err)
	}
	if len(again.CreatedEpisodes) != 0 || again.LinkedSlotCount != 0 {
		t.Errorf("2回目のoutput = %+v、期待値 = 何もしない", again)
	}
}

// TestCreateEpisodesFromSlotsUsecase_Execute_LinksExistingEpisodeは、計算した話数の公開中の
// エピソードがあればそれに紐づけ、非公開のエピソードには紐づけないことを検証する。
func TestCreateEpisodesFromSlotsUsecase_Execute_LinksExistingEpisode(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	ctx := context.Background()
	now := generationNow()

	workID := insertCreateTargetWork(t, db, sql.NullInt64{})
	secondID := insertSyncEpisode(t, db, generationEpisodeInput(workID, 2, 200))
	archived := generationEpisodeInput(workID, 3, 300)
	archived.unpublishedAt = sql.NullTime{Time: now, Valid: true}
	insertSyncEpisode(t, db, archived)
	program := insertGenerationProgram(t, db, workID, 1)
	linkedSlotID := program.insertSlot(2, now.Add(time.Hour), nil)
	createdSlotID := program.insertSlot(3, now.Add(2*time.Hour), nil)

	output, err := newCreateEpisodesFromSlotsUsecase(db).Execute(ctx, CreateEpisodesFromSlotsInput{Now: now})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	if output.LinkedSlotCount != 1 || len(output.CreatedEpisodes) != 1 {
		t.Fatalf("output = %+v、期待値 = 紐づけ1件・作成1件", output)
	}
	if got := readSlotEpisodeID(t, db, linkedSlotID); got.Int64 != int64(secondID) {
		t.Errorf("回数2の放送枠のepisode_id = %+v、期待値 = %d", got, secondID)
	}
	if got := readSlotEpisodeID(t, db, createdSlotID); got.Int64 != int64(output.CreatedEpisodes[0].EpisodeID) {
		t.Errorf("回数3の放送枠のepisode_id = %+v、期待値 = %d", got, output.CreatedEpisodes[0].EpisodeID)
	}
	// animeに対応付いていない作品では、episodesの行だけを書く。
	if episode := readCreatedEpisode(t, db, output.CreatedEpisodes[0].EpisodeID); episode.animeID.Valid {
		t.Errorf("episodes.anime_id = %+v、期待値 = NULL", episode.animeID)
	}
}

// TestCreateEpisodesFromSlotsUsecase_Execute_SkipsWhenCandidatesLinkedは、計算した話数の
// エピソードが同じ番組の別の放送枠に紐づいているとき、例外にせず飛ばして報告することを検証する
// (途中の話数から放送する番組で、編集者が紐づけを付け替えた場合)。
func TestCreateEpisodesFromSlotsUsecase_Execute_SkipsWhenCandidatesLinked(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	ctx := context.Background()
	now := generationNow()

	workID := insertCreateTargetWork(t, db, sql.NullInt64{})
	thirdID := insertSyncEpisode(t, db, generationEpisodeInput(workID, 3, 300))
	fourthID := insertSyncEpisode(t, db, generationEpisodeInput(workID, 4, 400))
	program := insertGenerationProgram(t, db, workID, 3)
	program.insertSlot(1, now.Add(-2*24*time.Hour), func(b *testutil.SlotBuilder) { b.WithEpisodeID(thirdID) })
	program.insertSlot(2, now.Add(-24*time.Hour), func(b *testutil.SlotBuilder) { b.WithEpisodeID(fourthID) })
	skippedID := program.insertSlot(4, now.Add(24*time.Hour), nil)

	output, err := newCreateEpisodesFromSlotsUsecase(db).Execute(ctx, CreateEpisodesFromSlotsInput{Now: now})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	if len(output.SkippedSlots) != 1 || len(output.CreatedEpisodes) != 0 || len(output.FailedSlots) != 0 {
		t.Fatalf("output = %+v、期待値 = 飛ばした放送枠1件のみ", output)
	}
	skipped := output.SkippedSlots[0]
	if skipped.SlotID != skippedID || skipped.WorkID != workID || skipped.Reason != SlotSkipReasonCandidatesLinked || skipped.WorkTitle == "" {
		t.Errorf("SkippedSlots[0] = %+v、期待値 = 放送枠%d・候補が紐づき済み", skipped, skippedID)
	}
	if got := readSlotEpisodeID(t, db, skippedID); got.Valid {
		t.Errorf("飛ばした放送枠のepisode_id = %+v、期待値 = NULL", got)
	}
	if got := countWorkEpisodes(t, db, workID); got != 2 {
		t.Errorf("作品のエピソード = %d件、期待値 = 2", got)
	}
}

// TestCreateEpisodesFromSlotsUsecase_Execute_CountsPrecedingIrregularSlotsは、話数の計算で
// 前にある不定期放送だけを引くことと、予定エピソード数との比較を作品内の順番で行うことを
// 検証する。
func TestCreateEpisodesFromSlotsUsecase_Execute_CountsPrecedingIrregularSlots(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	ctx := context.Background()
	now := generationNow()

	workID := insertCreateTargetWork(t, db, sql.NullInt64{})
	if _, err := db.Exec(`UPDATE works SET manual_episodes_count = 2 WHERE id = $1`, int64(workID)); err != nil {
		t.Fatalf("予定エピソード数の設定に失敗: %v", err)
	}
	firstID := insertSyncEpisode(t, db, generationEpisodeInput(workID, 1, 100))
	program := insertGenerationProgram(t, db, workID, 2)
	irregular := func(b *testutil.SlotBuilder) { b.WithIrregular() }
	program.insertSlot(1, now.Add(-3*24*time.Hour), func(b *testutil.SlotBuilder) { b.WithEpisodeID(firstID) })
	program.insertSlot(2, now.Add(-2*24*time.Hour), irregular)
	// 回数3は前の不定期放送1件を引いて作品内の2番目 (第2話) になる。後ろの不定期放送は数えない。
	targetID := program.insertSlot(3, now.Add(24*time.Hour), nil)
	// 回数4は作品内の3番目で、予定エピソード数 (2) を超えるため黙って飛ばす。
	exceededID := program.insertSlot(4, now.Add(2*24*time.Hour), nil)
	program.insertSlot(5, now.Add(3*24*time.Hour), irregular)

	output, err := newCreateEpisodesFromSlotsUsecase(db).Execute(ctx, CreateEpisodesFromSlotsInput{Now: now})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	if len(output.CreatedEpisodes) != 1 || len(output.SkippedSlots) != 0 || len(output.FailedSlots) != 0 {
		t.Fatalf("output = %+v、期待値 = 作成1件のみ", output)
	}
	created := output.CreatedEpisodes[0]
	if created.SlotID != targetID || created.Number != "第2話" {
		t.Errorf("CreatedEpisodes[0] = %+v、期待値 = 放送枠%d・第2話", created, targetID)
	}
	if got := readSlotEpisodeID(t, db, exceededID); got.Valid {
		t.Errorf("予定エピソード数を超える放送枠のepisode_id = %+v、期待値 = NULL", got)
	}
}

// TestCreateEpisodesFromSlotsUsecase_Execute_InsertsBeforeLargerEpisodeは、より大きい話数の
// エピソードがあればその直前に入れ、直後のエピソードの前のエピソードを付け替えることを検証する。
func TestCreateEpisodesFromSlotsUsecase_Execute_InsertsBeforeLargerEpisode(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	ctx := context.Background()
	now := generationNow()

	workID := insertCreateTargetWork(t, db, sql.NullInt64{})
	firstID := insertSyncEpisode(t, db, generationEpisodeInput(workID, 1, 100))
	thirdID := insertSyncEpisode(t, db, generationEpisodeInput(workID, 3, 300))
	if _, err := db.Exec(`UPDATE episodes SET prev_episode_id = $2, updated_at = '2001-01-01' WHERE id = $1`, int64(thirdID), int64(firstID)); err != nil {
		t.Fatalf("第3話の前のエピソードの設定に失敗: %v", err)
	}
	program := insertGenerationProgram(t, db, workID, 1)
	slotID := program.insertSlot(2, now.Add(time.Hour), nil)

	output, err := newCreateEpisodesFromSlotsUsecase(db).Execute(ctx, CreateEpisodesFromSlotsInput{Now: now})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	if len(output.CreatedEpisodes) != 1 || output.CreatedEpisodes[0].SlotID != slotID {
		t.Fatalf("output = %+v、期待値 = 放送枠%dの作成1件", output, slotID)
	}
	createdID := output.CreatedEpisodes[0].EpisodeID
	episode := readCreatedEpisode(t, db, createdID)
	if episode.sortNumber != 200 || episode.prevEpisodeID.Int64 != int64(firstID) {
		t.Errorf("エピソード = %+v、期待値 = sort_number 200・前のエピソード%d", episode, firstID)
	}

	var (
		prevEpisodeID sql.NullInt64
		updatedAt     time.Time
	)
	if err := db.QueryRow(`SELECT prev_episode_id, updated_at FROM episodes WHERE id = $1`, int64(thirdID)).Scan(&prevEpisodeID, &updatedAt); err != nil {
		t.Fatalf("第3話の読み込みに失敗: %v", err)
	}
	if prevEpisodeID.Int64 != int64(createdID) {
		t.Errorf("第3話のprev_episode_id = %+v、期待値 = %d", prevEpisodeID, createdID)
	}
	if updatedAt.Year() != 2001 {
		t.Errorf("第3話のupdated_at = %v、期待値 = 2001年のまま", updatedAt)
	}
}

// TestCreateEpisodesFromSlotsUsecase_Execute_SkipsWhenNoSortNumberGapは、入れる位置の前後の
// sort_numberの間が空いていなければ、作らずに飛ばして報告することを検証する。
func TestCreateEpisodesFromSlotsUsecase_Execute_SkipsWhenNoSortNumberGap(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	ctx := context.Background()
	now := generationNow()

	workID := insertCreateTargetWork(t, db, sql.NullInt64{})
	insertSyncEpisode(t, db, generationEpisodeInput(workID, 1, 100))
	insertSyncEpisode(t, db, generationEpisodeInput(workID, 3, 101))
	program := insertGenerationProgram(t, db, workID, 1)
	slotID := program.insertSlot(2, now.Add(time.Hour), nil)

	output, err := newCreateEpisodesFromSlotsUsecase(db).Execute(ctx, CreateEpisodesFromSlotsInput{Now: now})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	if len(output.SkippedSlots) != 1 || len(output.CreatedEpisodes) != 0 {
		t.Fatalf("output = %+v、期待値 = 飛ばした放送枠1件のみ", output)
	}
	if skipped := output.SkippedSlots[0]; skipped.SlotID != slotID || skipped.Reason != SlotSkipReasonNoSortNumberGap {
		t.Errorf("SkippedSlots[0] = %+v、期待値 = 放送枠%d・間が空いていない", skipped, slotID)
	}
	if got := countWorkEpisodes(t, db, workID); got != 2 {
		t.Errorf("作品のエピソード = %d件、期待値 = 2", got)
	}
}

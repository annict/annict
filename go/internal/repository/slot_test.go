package repository_test

import (
	"context"
	"database/sql"
	"slices"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// TestSlotRepository_ListForEpisodeGenerationは、自動生成の対象の放送枠の絞り込みと、
// 話数の計算に使う不定期放送の数を検証する。開始日時の下限は無く、ほかのテストがコミットした
// 放送枠も一覧に入るため、結果は自分の作品の放送枠に絞ってから確かめる。
func TestSlotRepository_ListForEpisodeGeneration(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	ctx := context.Background()

	before := time.Date(2301, 1, 15, 0, 0, 0, 0, time.UTC)
	inRange := before.Add(-13 * 24 * time.Hour)

	workID := testutil.NewWorkBuilder(t, tx).Build()
	channelID := testutil.NewChannelBuilder(t, tx).Build()
	programID := testutil.NewProgramBuilder(t, tx).
		WithChannelID(channelID).
		WithWorkID(workID).
		WithMinimumEpisodeGeneratableNumber(2).
		Build()
	newSlot := func() *testutil.SlotBuilder {
		return testutil.NewSlotBuilder(t, tx).
			WithWorkID(workID).
			WithChannelID(channelID).
			WithProgramID(programID).
			WithStartedAt(inRange)
	}

	// 不定期放送の数に入るのは、回数が番組の下限 (2) 以上で、対象の放送枠より小さいものだけ。
	newSlot().WithNumber(1).WithIrregular().Build()
	newSlot().WithNumber(2).WithIrregular().Build()
	target := newSlot().WithNumber(4).WithStartedAt(inRange.Add(time.Hour)).Build()
	newSlot().WithNumber(5).WithIrregular().Build()
	earlier := newSlot().WithNumber(3).Build()

	episodeID := testutil.NewEpisodeBuilder(t, tx, workID).Build()
	newSlot().WithNumber(6).WithEpisodeID(episodeID).Build()
	newSlot().WithNumber(7).WithRebroadcast().Build()
	newSlot().WithNumber(8).WithDeletedAt(time.Now()).Build()
	newSlot().WithNumber(9).WithUnpublishedAt(time.Now()).Build()
	newSlot().WithNumber(10).WithStartedAt(before).Build()
	// 開始日時の下限は無く、ずっと前に始まった放送枠も対象になる。
	past := newSlot().WithNumber(11).WithStartedAt(time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)).Build()
	newSlot().Build()

	archivedWorkID := testutil.NewWorkBuilder(t, tx).WithUnpublishedAt(time.Now()).Build()
	archivedProgramID := testutil.NewProgramBuilder(t, tx).WithChannelID(channelID).WithWorkID(archivedWorkID).Build()
	testutil.NewSlotBuilder(t, tx).
		WithWorkID(archivedWorkID).
		WithChannelID(channelID).
		WithProgramID(archivedProgramID).
		WithNumber(1).
		WithStartedAt(inRange).
		Build()

	repo := repository.NewSlotRepository(query.New(db)).WithTx(tx)
	slots, err := repo.ListForEpisodeGeneration(ctx, before)
	if err != nil {
		t.Fatalf("ListForEpisodeGeneration()のエラー = %v", err)
	}

	got := filterSlotIDsByWork(slots, workID, archivedWorkID)
	want := []model.SlotID{past, earlier, target}
	if !slices.Equal(got, want) {
		t.Errorf("放送枠の並び = %v、期待値 = %v", got, want)
	}

	slot, err := repo.GetForEpisodeGenerationByID(ctx, target, workID, before)
	if err != nil || slot == nil {
		t.Fatalf("GetForEpisodeGenerationByID() = %+v、エラー = %v", slot, err)
	}
	if slot.WorkID != workID || slot.Number != 4 || !slot.StartedAt.Equal(inRange.Add(time.Hour)) {
		t.Errorf("放送枠 = %+v、期待値 = 作品%d・回数4", slot, workID)
	}
	if slot.Program == nil || slot.Program.ID != model.ProgramID(programID) || slot.Program.MinimumEpisodeGeneratableNumber != 2 {
		t.Errorf("Program = %+v、期待値 = 番組%d・下限2", slot.Program, programID)
	}
	if slot.PrecedingIrregularCount != 1 {
		t.Errorf("PrecedingIrregularCount = %d、期待値 = 1", slot.PrecedingIrregularCount)
	}
	if pastSlot, err := repo.GetForEpisodeGenerationByID(ctx, past, workID, before); err != nil || pastSlot == nil {
		t.Errorf("ずっと前に始まった放送枠のGetForEpisodeGenerationByID() = %+v、エラー = %v、期待値 = 放送枠", pastSlot, err)
	}
}

// TestSlotRepository_ListForEpisodeGeneration_ManualEpisodesCountは、予定エピソード数を超える話に
// あたる放送枠と、公開中のエピソード数が予定エピソード数に達した作品で紐づけの候補が無い放送枠を、
// 一覧に載せないことを検証する。
func TestSlotRepository_ListForEpisodeGeneration_ManualEpisodesCount(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	ctx := context.Background()

	before := time.Date(2302, 1, 15, 0, 0, 0, 0, time.UTC)
	startedAt := before.Add(-24 * time.Hour)
	channelID := testutil.NewChannelBuilder(t, tx).Build()
	newProgram := func(workID model.WorkID) func(number int32, configure func(*testutil.SlotBuilder)) model.SlotID {
		programID := testutil.NewProgramBuilder(t, tx).WithChannelID(channelID).WithWorkID(workID).Build()
		return func(number int32, configure func(*testutil.SlotBuilder)) model.SlotID {
			builder := testutil.NewSlotBuilder(t, tx).
				WithWorkID(workID).
				WithChannelID(channelID).
				WithProgramID(programID).
				WithNumber(number).
				WithStartedAt(startedAt.Add(time.Duration(number) * time.Minute))
			if configure != nil {
				configure(builder)
			}
			return builder.Build()
		}
	}

	// 公開中のエピソードが無く、予定エピソード数 (2) までの話にあたる放送枠だけが載る。作品内の
	// 順番は不定期放送を除いて数えるため、回数3が2番目、回数4が3番目になる。
	notFilledWorkID := testutil.NewWorkBuilder(t, tx).WithManualEpisodesCount(2).Build()
	notFilledSlot := newProgram(notFilledWorkID)
	firstSlot := notFilledSlot(1, nil)
	notFilledSlot(2, func(b *testutil.SlotBuilder) { b.WithIrregular() })
	secondSlot := notFilledSlot(3, nil)
	notFilledSlot(4, nil)

	// 公開中のエピソード数が予定エピソード数に達した作品では、計算した話数と等しい公開中の
	// エピソードがある放送枠だけが載る。話数は開始話数 (11) を起点に数える。
	filledWorkID := testutil.NewWorkBuilder(t, tx).WithManualEpisodesCount(3).Build()
	if _, err := tx.Exec(`UPDATE works SET start_episode_raw_number = 11 WHERE id = $1`, int64(filledWorkID)); err != nil {
		t.Fatalf("開始話数の設定に失敗: %v", err)
	}
	testutil.NewEpisodeBuilder(t, tx, filledWorkID).WithRawNumber(11).Build()
	testutil.NewEpisodeBuilder(t, tx, filledWorkID).WithRawNumber(15).Build()
	testutil.NewEpisodeBuilder(t, tx, filledWorkID).WithRawNumber(16).Build()
	testutil.NewEpisodeBuilder(t, tx, filledWorkID).WithRawNumber(12).WithUnpublishedAt(time.Now()).Build()
	testutil.NewEpisodeBuilder(t, tx, filledWorkID).WithRawNumber(13).WithDeletedAt(time.Now()).Build()
	filledSlot := newProgram(filledWorkID)
	candidateSlot := filledSlot(1, nil)
	filledSlot(2, nil)
	filledSlot(3, nil)

	// 非公開・削除済みのエピソードは公開中のエピソード数に含めない。
	partlyPublishedWorkID := testutil.NewWorkBuilder(t, tx).WithManualEpisodesCount(2).Build()
	testutil.NewEpisodeBuilder(t, tx, partlyPublishedWorkID).WithRawNumber(1).Build()
	testutil.NewEpisodeBuilder(t, tx, partlyPublishedWorkID).WithRawNumber(2).WithUnpublishedAt(time.Now()).Build()
	testutil.NewEpisodeBuilder(t, tx, partlyPublishedWorkID).WithRawNumber(3).WithDeletedAt(time.Now()).Build()
	partlyPublishedSlot := newProgram(partlyPublishedWorkID)(2, nil)

	repo := repository.NewSlotRepository(query.New(db)).WithTx(tx)
	slots, err := repo.ListForEpisodeGeneration(ctx, before)
	if err != nil {
		t.Fatalf("ListForEpisodeGeneration()のエラー = %v", err)
	}

	got := filterSlotIDsByWork(slots, notFilledWorkID, filledWorkID, partlyPublishedWorkID)
	want := []model.SlotID{firstSlot, candidateSlot, partlyPublishedSlot, secondSlot}
	if !slices.Equal(got, want) {
		t.Errorf("放送枠 = %v、期待値 = %v", got, want)
	}
}

// filterSlotIDsByWorkは、一覧のうち指定した作品の放送枠のIDを、一覧の順のまま返す。
func filterSlotIDsByWork(slots []*model.Slot, workIDs ...model.WorkID) []model.SlotID {
	var ids []model.SlotID
	for _, slot := range slots {
		if slices.Contains(workIDs, slot.WorkID) {
			ids = append(ids, slot.ID)
		}
	}
	return ids
}

// TestSlotRepository_LinkEpisodeは、紐づいていない放送枠にだけエピソードを紐づけることを
// 検証する。
func TestSlotRepository_LinkEpisode(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	ctx := context.Background()

	workID := testutil.NewWorkBuilder(t, tx).Build()
	channelID := testutil.NewChannelBuilder(t, tx).Build()
	programID := testutil.NewProgramBuilder(t, tx).WithWorkID(workID).WithChannelID(channelID).Build()
	now := time.Now().Truncate(time.Microsecond)
	slotID := testutil.NewSlotBuilder(t, tx).WithWorkID(workID).WithChannelID(channelID).WithProgramID(programID).WithNumber(1).WithStartedAt(now).Build()
	deletedSlotID := testutil.NewSlotBuilder(t, tx).WithWorkID(workID).WithChannelID(channelID).WithProgramID(programID).WithNumber(2).WithStartedAt(now).WithDeletedAt(now).Build()
	firstID := testutil.NewEpisodeBuilder(t, tx, workID).Build()
	secondID := testutil.NewEpisodeBuilder(t, tx, workID).Build()

	var firstAnimeID int64
	if err := tx.QueryRow(`INSERT INTO animes (status) VALUES ('published') RETURNING id`).Scan(&firstAnimeID); err != nil {
		t.Fatalf("animesの挿入に失敗: %v", err)
	}
	if _, err := tx.Exec(`UPDATE episodes SET anime_id = $1 WHERE id = $2`, firstAnimeID, int64(firstID)); err != nil {
		t.Fatalf("episodes.anime_idの設定に失敗: %v", err)
	}

	repo := repository.NewSlotRepository(query.New(db)).WithTx(tx)
	slot, err := repo.GetForEpisodeGenerationByID(ctx, slotID, workID, now.Add(time.Hour))
	if err != nil || slot == nil {
		t.Fatalf("GetForEpisodeGenerationByID() = %+v、エラー = %v", slot, err)
	}

	linked, err := repo.LinkEpisode(ctx, slot, firstID)
	if err != nil {
		t.Fatalf("LinkEpisode()のエラー = %v", err)
	}
	if !linked {
		t.Error("紐づいていない放送枠に紐づけられなかった")
	}

	linked, err = repo.LinkEpisode(ctx, slot, secondID)
	if err != nil {
		t.Fatalf("LinkEpisode()のエラー = %v", err)
	}
	if linked {
		t.Error("すでに紐づいた放送枠を書き換えた")
	}

	deletedSlot := *slot
	deletedSlot.ID = deletedSlotID
	deletedSlot.Number = 2
	linked, err = repo.LinkEpisode(ctx, &deletedSlot, secondID)
	if err != nil {
		t.Fatalf("LinkEpisode()のエラー = %v", err)
	}
	if linked {
		t.Error("削除済みの放送枠に紐づけた")
	}

	var episodeID int64
	var episodeAnimeID sql.NullInt64
	if err := tx.QueryRow(`SELECT episode_id, episode_anime_id FROM slots WHERE id = $1`, int64(slotID)).Scan(&episodeID, &episodeAnimeID); err != nil {
		t.Fatalf("放送枠の読み込みに失敗: %v", err)
	}
	if episodeID != int64(firstID) {
		t.Errorf("slots.episode_id = %d、期待値 = %d", episodeID, firstID)
	}
	if !episodeAnimeID.Valid || episodeAnimeID.Int64 != firstAnimeID {
		t.Errorf("slots.episode_anime_id = %+v、期待値 = %d (紐づけたエピソードのanime_id)", episodeAnimeID, firstAnimeID)
	}
	var deletedEpisodeAnimeID sql.NullInt64
	if err := tx.QueryRow(`SELECT episode_anime_id FROM slots WHERE id = $1`, int64(deletedSlotID)).Scan(&deletedEpisodeAnimeID); err != nil {
		t.Fatalf("放送枠の読み込みに失敗: %v", err)
	}
	if deletedEpisodeAnimeID.Valid {
		t.Errorf("紐づけなかった放送枠のepisode_anime_id = %+v、期待値 = NULL", deletedEpisodeAnimeID)
	}
}

// TestSlotRepository_GenerationRejectsChangedSlotは、読み直しと紐づけの両方で対象条件と
// 所属を検査する。紐づけには照合で読んだ番組・回数・開始時刻との一致も必要。
func TestSlotRepository_GenerationRejectsChangedSlot(t *testing.T) {
	t.Parallel()

	for _, change := range []struct {
		name   string
		update string
	}{
		{"非公開", `UPDATE slots SET unpublished_at = NOW() WHERE id = $1`},
		{"削除", `UPDATE slots SET deleted_at = NOW() WHERE id = $1`},
		{"再放送", `UPDATE slots SET rebroadcast = true WHERE id = $1`},
		{"不定期放送", `UPDATE slots SET irregular = true WHERE id = $1`},
		{"番組なし", `UPDATE slots SET program_id = NULL WHERE id = $1`},
		{"回数なし", `UPDATE slots SET number = NULL WHERE id = $1`},
		{"回数変更", `UPDATE slots SET number = 2 WHERE id = $1`},
		{"範囲外への日時変更", `UPDATE slots SET started_at = started_at + INTERVAL '30 days' WHERE id = $1`},
		{"過去への日時変更", `UPDATE slots SET started_at = started_at - INTERVAL '30 days' WHERE id = $1`},
	} {
		t.Run(change.name, func(t *testing.T) {
			t.Parallel()
			db, tx := testutil.SetupTx(t)
			ctx := context.Background()
			workID := testutil.NewWorkBuilder(t, tx).Build()
			otherWorkID := testutil.NewWorkBuilder(t, tx).Build()
			channelID := testutil.NewChannelBuilder(t, tx).Build()
			programID := testutil.NewProgramBuilder(t, tx).WithWorkID(workID).WithChannelID(channelID).Build()
			now := time.Now().Truncate(time.Microsecond)
			slotID := testutil.NewSlotBuilder(t, tx).WithWorkID(workID).WithChannelID(channelID).WithProgramID(programID).WithNumber(1).WithStartedAt(now).Build()
			episodeID := testutil.NewEpisodeBuilder(t, tx, workID).Build()
			repo := repository.NewSlotRepository(query.New(db)).WithTx(tx)
			slot, err := repo.GetForEpisodeGenerationByID(ctx, slotID, workID, now.Add(time.Hour))
			if err != nil || slot == nil {
				t.Fatalf("放送枠の取得 = %+v、エラー = %v", slot, err)
			}
			if got, err := repo.GetForEpisodeGenerationByID(ctx, slotID, otherWorkID, now.Add(time.Hour)); err != nil || got != nil {
				t.Fatalf("所属の違う放送枠の取得 = %+v、エラー = %v、期待値 = nil", got, err)
			}
			if _, err := tx.Exec(change.update, int64(slotID)); err != nil {
				t.Fatalf("放送枠の変更に失敗: %v", err)
			}
			if linked, err := repo.LinkEpisode(ctx, slot, episodeID); err != nil || linked {
				t.Errorf("変更後のLinkEpisode() = %v、エラー = %v、期待値 = false", linked, err)
			}
			got, err := repo.GetForEpisodeGenerationByID(ctx, slotID, workID, now.Add(time.Hour))
			if err != nil {
				t.Fatalf("変更後のGetForEpisodeGenerationByID()のエラー = %v", err)
			}
			switch {
			case change.name == "回数変更":
				if got == nil || got.Number != 2 {
					t.Errorf("変更後の放送枠 = %+v、期待値 = 回数2", got)
				}
			case change.name == "過去への日時変更":
				// 開始日時の下限は無いため、過去へ動かした放送枠も新しい開始日時で読み直せる。
				if got == nil || !got.StartedAt.Equal(now.Add(-30*24*time.Hour)) {
					t.Errorf("変更後の放送枠 = %+v、期待値 = 30日前の開始日時", got)
				}
			case got != nil:
				t.Errorf("変更後の放送枠 = %+v、期待値 = nil", got)
			}
		})
	}
}

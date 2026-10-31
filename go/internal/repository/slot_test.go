package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// TestSlotRepository_ListForEpisodeGenerationは、自動生成の対象の放送枠の絞り込みと、
// 話数の計算に使う不定期放送の数を検証する。開始日時の範囲は、ほかのテストがコミットした
// 放送枠と重ならない遠い未来に取る。
func TestSlotRepository_ListForEpisodeGeneration(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	ctx := context.Background()

	from := time.Date(2301, 1, 1, 0, 0, 0, 0, time.UTC)
	before := from.Add(14 * 24 * time.Hour)
	inRange := from.Add(24 * time.Hour)

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
	slots, err := repo.ListForEpisodeGeneration(ctx, from, before)
	if err != nil {
		t.Fatalf("ListForEpisodeGeneration()のエラー = %v", err)
	}

	if len(slots) != 2 {
		t.Fatalf("len(slots) = %d、期待値 = 2", len(slots))
	}
	if slots[0].ID != earlier || slots[1].ID != target {
		t.Errorf("放送枠の並び = [%d, %d]、期待値 = [%d, %d]", slots[0].ID, slots[1].ID, earlier, target)
	}

	if slots[1].WorkID != workID {
		t.Errorf("WorkID = %d、期待値 = %d", slots[1].WorkID, workID)
	}
	got, err := repo.GetForEpisodeGenerationByID(ctx, target, workID, from, before)
	if err != nil || got == nil {
		t.Fatalf("GetForEpisodeGenerationByID() = %+v、エラー = %v", got, err)
	}
	if got.WorkID != workID || got.Number != 4 || !got.StartedAt.Equal(inRange.Add(time.Hour)) {
		t.Errorf("放送枠 = %+v、期待値 = 作品%d・回数4", got, workID)
	}
	if got.Program == nil || got.Program.ID != model.ProgramID(programID) || got.Program.MinimumEpisodeGeneratableNumber != 2 {
		t.Errorf("Program = %+v、期待値 = 番組%d・下限2", got.Program, programID)
	}
	if got.PrecedingIrregularCount != 1 {
		t.Errorf("PrecedingIrregularCount = %d、期待値 = 1", got.PrecedingIrregularCount)
	}
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

	repo := repository.NewSlotRepository(query.New(db)).WithTx(tx)
	slot, err := repo.GetForEpisodeGenerationByID(ctx, slotID, workID, now.Add(-time.Hour), now.Add(time.Hour))
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
	if err := tx.QueryRow(`SELECT episode_id FROM slots WHERE id = $1`, int64(slotID)).Scan(&episodeID); err != nil {
		t.Fatalf("放送枠の読み込みに失敗: %v", err)
	}
	if episodeID != int64(firstID) {
		t.Errorf("slots.episode_id = %d、期待値 = %d", episodeID, firstID)
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
			slot, err := repo.GetForEpisodeGenerationByID(ctx, slotID, workID, now.Add(-time.Hour), now.Add(time.Hour))
			if err != nil || slot == nil {
				t.Fatalf("放送枠の取得 = %+v、エラー = %v", slot, err)
			}
			if got, err := repo.GetForEpisodeGenerationByID(ctx, slotID, otherWorkID, now.Add(-time.Hour), now.Add(time.Hour)); err != nil || got != nil {
				t.Fatalf("所属の違う放送枠の取得 = %+v、エラー = %v、期待値 = nil", got, err)
			}
			if _, err := tx.Exec(change.update, int64(slotID)); err != nil {
				t.Fatalf("放送枠の変更に失敗: %v", err)
			}
			if linked, err := repo.LinkEpisode(ctx, slot, episodeID); err != nil || linked {
				t.Errorf("変更後のLinkEpisode() = %v、エラー = %v、期待値 = false", linked, err)
			}
			got, err := repo.GetForEpisodeGenerationByID(ctx, slotID, workID, now.Add(-time.Hour), now.Add(time.Hour))
			if err != nil {
				t.Fatal(err)
			}
			if change.name == "回数変更" {
				if got == nil || got.Number != 2 {
					t.Errorf("変更後の放送枠 = %+v、期待値 = 回数2", got)
				}
			} else if got != nil {
				t.Errorf("変更後の放送枠 = %+v、期待値 = nil", got)
			}
		})
	}
}

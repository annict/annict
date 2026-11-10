package usecase

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// personSnapshotFromPersonは、事前に読んだ人物 (PersonRepository.GetForRevisionByIDの結果) から、書き込む前の
// スナップショットを組み立てる。
func personSnapshotFromPerson(p *model.Person) model.PersonSnapshot {
	return model.PersonSnapshot{
		Name:              p.Name,
		NameKana:          p.NameKana,
		NameEn:            p.NameEn,
		Nickname:          p.Nickname,
		NicknameEn:        p.NicknameEn,
		Gender:            p.Gender,
		BloodType:         p.BloodType,
		PrefectureID:      p.PrefectureID,
		Birthday:          p.Birthday,
		Height:            p.Height,
		URL:               p.URL,
		URLEn:             p.URLEn,
		WikipediaURL:      p.WikipediaURL,
		WikipediaURLEn:    p.WikipediaURLEn,
		TwitterUsername:   p.TwitterUsername,
		TwitterUsernameEn: p.TwitterUsernameEn,
		Status:            p.Status,
		ArchiveMessage:    p.ArchiveMessage,
	}
}

// personSnapshotFromCreateParamsは、作成する人物の値から作成の版のスナップショットを組み立てる。作成のクエリが
// 書かない列は、peopleの既定値 (英語の列は空文字列・NULL許容の列はNULL・状態は公開中) で持つ。
func personSnapshotFromCreateParams(params repository.CreatePersonParams) model.PersonSnapshot {
	return model.PersonSnapshot{
		Name:     params.Name,
		NameKana: params.NameKana,
		Status:   model.PersonStatusPublished,
	}
}

// setPersonSnapshotColumnsは、人物の更新が書く値をスナップショットに写す。状態と非公開の理由は更新が書かないため
// 書き込む前の値のまま残す。
func setPersonSnapshotColumns(s *model.PersonSnapshot, params repository.UpdatePersonParams) {
	s.Name = params.Name
	s.NameKana = params.NameKana
	s.NameEn = params.NameEn
	s.Nickname = params.Nickname
	s.NicknameEn = params.NicknameEn
	s.Gender = params.Gender
	s.BloodType = params.BloodType
	s.PrefectureID = params.PrefectureID
	s.Birthday = params.Birthday
	s.Height = params.Height
	s.URL = params.URL
	s.URLEn = params.URLEn
	s.WikipediaURL = params.WikipediaURL
	s.WikipediaURLEn = params.WikipediaURLEn
	s.TwitterUsername = params.TwitterUsername
	s.TwitterUsernameEn = params.TwitterUsernameEn
}

// personWriteは、人物の書き込みをトランザクションの中で行う関数。リポジトリはトランザクションに束ねたものを受け取り、
// 行を書いたかを返す。
type personWrite func(personRepo *repository.PersonRepository) (bool, error)

// writePersonWithRecordは、人物の書き込みと編集履歴の記録を1つのトランザクションで行う。書き込みが行を
// 書かなかったときは何も記録せずにロールバックし、falseを返す。recordPlanがnil (内容が変わらない) のときは
// 書き込みだけを行う。
func writePersonWithRecord(
	ctx context.Context,
	db *sql.DB,
	personRepo *repository.PersonRepository,
	dbEditRecordRepos DBEditRecordRepos,
	recordPlan *dbEditRecordPlan,
	write personWrite,
) (bool, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("トランザクションの開始に失敗しました: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	written, err := write(personRepo.WithTx(tx))
	if err != nil || !written {
		return false, err
	}
	if err := writeDBEditRecord(ctx, dbEditRecordRepos.WithTx(tx), recordPlan); err != nil {
		return false, err
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("トランザクションのコミットに失敗しました: %w", err)
	}
	return true, nil
}

// planPersonStatusRecordは、人物の状態だけを変える書き込み (非公開・再公開・削除・掲載停止・掲載再開) の記録の
// 計画を立てる。書き込んだ後の状態は、書き込む前の状態の状態の列を差し替えたもの。
func planPersonStatusRecord(
	ctx context.Context,
	dbEditRecordRepos DBEditRecordRepos,
	user *model.User,
	action model.DBEditAction,
	before *model.Person,
	status model.PersonStatus,
) (*dbEditRecordPlan, error) {
	beforeSnapshot := personSnapshotFromPerson(before)
	after := beforeSnapshot
	after.Status = status
	plan, err := planDBEditRecord(ctx, dbEditRecordRepos, DBEditRecordInput{
		Actor:  model.DBEditActorUser,
		UserID: &user.ID,
		Action: action,
		People: []PersonRevisionInput{{PersonID: before.ID, Before: &beforeSnapshot, After: after}},
	})
	if err != nil {
		return nil, fmt.Errorf("編集履歴の記録の計画に失敗しました: %w", err)
	}
	return plan, nil
}

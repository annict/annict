package usecase

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// DBEditRecordReposはAnnict DBの編集履歴を記録するリポジトリを束ねる。各書き込みUseCaseは
// この束を1つの値で受け取り、planDBEditRecordとwriteDBEditRecordで記録する。WithTxは
// トランザクションに束ねたコピーを返し、記録をリソースの書き込みと同じトランザクションに参加させる。
type DBEditRecordRepos struct {
	DBEdit         *repository.DBEditRepository
	AnimeRevision  *repository.AnimeRevisionRepository
	PersonRevision *repository.PersonRevisionRepository
}

// WithTxは各リポジトリを指定トランザクションに束ねた束のコピーを返す。
func (r DBEditRecordRepos) WithTx(tx *sql.Tx) DBEditRecordRepos {
	return DBEditRecordRepos{
		DBEdit:         r.DBEdit.WithTx(tx),
		AnimeRevision:  r.AnimeRevision.WithTx(tx),
		PersonRevision: r.PersonRevision.WithTx(tx),
	}
}

// DBEditRecordInputは1回の保存 (操作) として記録する編集。
type DBEditRecordInput struct {
	Actor model.DBEditActor
	// UserIDは編集者。ActorがDBEditActorUserのときは必須で、DBEditActorSystemのときは空にする。
	UserID *model.UserID
	Action model.DBEditAction
	Animes []AnimeRevisionInput
	People []PersonRevisionInput
}

// AnimeRevisionInputは1つの作品・エピソードの、書き込む前と後の状態。Beforeは書き込みの前に読んだ状態で、
// 作成ではnilにする。Afterはトランザクションの前に、書き込む値と事前に読んだ行から組み立てる。
type AnimeRevisionInput struct {
	AnimeID model.AnimeID
	Before  *model.AnimeSnapshot
	After   model.AnimeSnapshot
}

// PersonRevisionInputは1つの人物の、書き込む前と後の状態。BeforeとAfterの意味はAnimeRevisionInputと同じ。
type PersonRevisionInput struct {
	PersonID model.PersonID
	Before   *model.PersonSnapshot
	After    model.PersonSnapshot
}

// dbEditRecordPlanはトランザクションの中で書く操作と版。内容の変わらないリソースは除いてある。
type dbEditRecordPlan struct {
	params repository.CreateDBEditParams
	animes []plannedAnimeRevision
	people []plannedPersonRevision
}

// plannedAnimeRevisionは1つの作品・エピソードについて書く版。initialがnilでなければ、版の前に起点の版を書く。
type plannedAnimeRevision struct {
	animeID model.AnimeID
	initial *model.AnimeSnapshot
	after   model.AnimeSnapshot
}

// plannedPersonRevisionは1つの人物について書く版。initialの意味はplannedAnimeRevisionと同じ。
type plannedPersonRevision struct {
	personID model.PersonID
	initial  *model.PersonSnapshot
	after    model.PersonSnapshot
}

// planDBEditRecordは、トランザクションの前に記録の計画を立てる。書き込む前と後の状態が同じリソースは除き、
// すべて同じときはnilを返す (操作も版も作らない)。版をまだ持たない既存のリソースには、書き込む前の状態を
// 起点の版として足す。版の有無はここで調べ、トランザクションの中では読まない。
func planDBEditRecord(ctx context.Context, repos DBEditRecordRepos, input DBEditRecordInput) (*dbEditRecordPlan, error) {
	if (input.Actor == model.DBEditActorUser) != (input.UserID != nil) {
		return nil, fmt.Errorf("編集履歴の主体%qと編集者の有無が合いません", input.Actor)
	}

	plan := &dbEditRecordPlan{
		params: repository.CreateDBEditParams{Actor: input.Actor, UserID: input.UserID, Action: input.Action},
	}

	var existingAnimeIDs []model.AnimeID
	for _, a := range input.Animes {
		if a.Before != nil && model.DiffAnimeSnapshots(a.Before, a.After).IsEmpty() {
			continue
		}
		plan.animes = append(plan.animes, plannedAnimeRevision{animeID: a.AnimeID, initial: a.Before, after: a.After})
		if a.Before != nil {
			existingAnimeIDs = append(existingAnimeIDs, a.AnimeID)
		}
	}
	var existingPersonIDs []model.PersonID
	for _, p := range input.People {
		if p.Before != nil && model.DiffPersonSnapshots(p.Before, p.After).IsEmpty() {
			continue
		}
		plan.people = append(plan.people, plannedPersonRevision{personID: p.PersonID, initial: p.Before, after: p.After})
		if p.Before != nil {
			existingPersonIDs = append(existingPersonIDs, p.PersonID)
		}
	}
	if len(plan.animes) == 0 && len(plan.people) == 0 {
		return nil, nil
	}

	// 版をすでに持つリソースは、1つ前の版が比べる相手になるため、起点の版を足さない。
	animesWithRevisions, err := repos.AnimeRevision.ListAnimeIDsWithRevisions(ctx, existingAnimeIDs)
	if err != nil {
		return nil, fmt.Errorf("作品・エピソードの版の有無の取得に失敗: %w", err)
	}
	for i := range plan.animes {
		if animesWithRevisions[plan.animes[i].animeID] {
			plan.animes[i].initial = nil
		}
	}
	peopleWithRevisions, err := repos.PersonRevision.ListPersonIDsWithRevisions(ctx, existingPersonIDs)
	if err != nil {
		return nil, fmt.Errorf("人物の版の有無の取得に失敗: %w", err)
	}
	for i := range plan.people {
		if peopleWithRevisions[plan.people[i].personID] {
			plan.people[i].initial = nil
		}
	}

	return plan, nil
}

// writeDBEditRecordは、planDBEditRecordが立てた計画のとおりに操作と版を書く。reposは呼び出し元の
// トランザクションに束ねたものを渡す。planがnilのときは何も書かない。起点の版は、同じリソースの版より
// 小さいIDになるよう先に書く。
func writeDBEditRecord(ctx context.Context, repos DBEditRecordRepos, plan *dbEditRecordPlan) error {
	if plan == nil {
		return nil
	}

	edit, err := repos.DBEdit.Create(ctx, plan.params)
	if err != nil {
		return fmt.Errorf("編集履歴の操作の作成に失敗: %w", err)
	}

	for _, a := range plan.animes {
		if a.initial != nil {
			if err := repos.AnimeRevision.CreateInitial(ctx, a.animeID, *a.initial); err != nil {
				return fmt.Errorf("作品・エピソード%sの起点の版の作成に失敗: %w", a.animeID, err)
			}
		}
		if err := repos.AnimeRevision.Create(ctx, edit.ID, a.animeID, a.after); err != nil {
			return fmt.Errorf("作品・エピソード%sの版の作成に失敗: %w", a.animeID, err)
		}
	}
	for _, p := range plan.people {
		if p.initial != nil {
			if err := repos.PersonRevision.CreateInitial(ctx, p.personID, *p.initial); err != nil {
				return fmt.Errorf("人物%sの起点の版の作成に失敗: %w", p.personID, err)
			}
		}
		if err := repos.PersonRevision.Create(ctx, edit.ID, p.personID, p.after); err != nil {
			return fmt.Errorf("人物%sの版の作成に失敗: %w", p.personID, err)
		}
	}

	return nil
}

// setCreatedAnimeIDは、作成する作品・エピソードの版に、トランザクションの中で作ったanimeのIDを入れる。
// 作成ではanimeのIDが挿入するまで決まらないため、計画は仮のIDで立て、挿入した直後にこれで差し替える。
// indexはDBEditRecordInput.Animesでの並び順で、作成の版は内容が変わらないとして除かれることがないため
// 計画でも同じ位置にある。
func (p *dbEditRecordPlan) setCreatedAnimeID(index int, animeID model.AnimeID) {
	p.animes[index].animeID = animeID
}

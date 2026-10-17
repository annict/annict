// Package db_personはAnnict DB管理画面の人物関連機能を提供するHTTPハンドラーを定義する。
package db_person

import (
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/usecase"
)

// HandlerはAnnict DB管理画面の人物関連HTTPハンドラーが共有する依存をまとめる。
type Handler struct {
	cfg               *config.Config
	sessionManager    *session.Manager
	flashMgr          *session.FlashManager
	getDBPeopleUC     *usecase.GetDBPeopleUsecase
	getDBPersonNewUC  *usecase.GetDBPersonNewUsecase
	createPeopleUC    *usecase.CreatePeopleUsecase
	getDBPersonEditUC *usecase.GetDBPersonEditUsecase
	updatePersonUC    *usecase.UpdatePersonUsecase
	deletePersonUC    *usecase.DeletePersonUsecase
}

func NewHandler(
	cfg *config.Config,
	sessionManager *session.Manager,
	flashMgr *session.FlashManager,
	getDBPeopleUC *usecase.GetDBPeopleUsecase,
	getDBPersonNewUC *usecase.GetDBPersonNewUsecase,
	createPeopleUC *usecase.CreatePeopleUsecase,
	getDBPersonEditUC *usecase.GetDBPersonEditUsecase,
	updatePersonUC *usecase.UpdatePersonUsecase,
	deletePersonUC *usecase.DeletePersonUsecase,
) *Handler {
	return &Handler{
		cfg:               cfg,
		sessionManager:    sessionManager,
		flashMgr:          flashMgr,
		getDBPeopleUC:     getDBPeopleUC,
		getDBPersonNewUC:  getDBPersonNewUC,
		createPeopleUC:    createPeopleUC,
		getDBPersonEditUC: getDBPersonEditUC,
		updatePersonUC:    updatePersonUC,
		deletePersonUC:    deletePersonUC,
	}
}

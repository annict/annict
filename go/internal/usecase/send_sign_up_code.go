package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/annict/annict/go/internal/auth"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/worker"
)

// SendSignUpCodeUsecase は新規登録確認コードを生成・送信するユースケースです
type SendSignUpCodeUsecase struct {
	db             *sql.DB
	signUpCodeRepo *repository.SignUpCodeRepository
	riverClient    *worker.Client
}

// NewSendSignUpCodeUsecase は新しいSendSignUpCodeUsecaseを作成します
func NewSendSignUpCodeUsecase(db *sql.DB, signUpCodeRepo *repository.SignUpCodeRepository, riverClient *worker.Client) *SendSignUpCodeUsecase {
	return &SendSignUpCodeUsecase{
		db:             db,
		signUpCodeRepo: signUpCodeRepo,
		riverClient:    riverClient,
	}
}

// SendSignUpCodeResult はコード送信の結果を表します
type SendSignUpCodeResult struct {
	Code  string // 平文コード（テスト用）
	Email string // メールアドレス
}

// Execute は新規登録確認コードを生成し、メール送信ジョブをエンキューします
func (uc *SendSignUpCodeUsecase) Execute(ctx context.Context, email string, locale string) (*SendSignUpCodeResult, error) {
	// トランザクション開始
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("トランザクション開始に失敗: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	signUpCodeRepoTx := uc.signUpCodeRepo.WithTx(tx)

	// 既存の未使用コードを無効化
	if err := signUpCodeRepoTx.InvalidateByEmail(ctx, email); err != nil {
		return nil, fmt.Errorf("古いコードの無効化に失敗: %w", err)
	}

	// 新しい6桁コードを生成
	code, err := auth.GenerateVerificationCode()
	if err != nil {
		return nil, fmt.Errorf("コード生成に失敗: %w", err)
	}

	// コードをbcryptでハッシュ化
	codeDigest, err := auth.HashCode(code)
	if err != nil {
		return nil, fmt.Errorf("コードのハッシュ化に失敗: %w", err)
	}

	// コードをデータベースに保存（有効期限: 15分）
	if err := signUpCodeRepoTx.Create(ctx, repository.SignUpCodeCreateParams{
		Email:      email,
		CodeDigest: codeDigest,
		ExpiresAt:  time.Now().Add(15 * time.Minute),
	}); err != nil {
		return nil, fmt.Errorf("新規登録確認コードの作成に失敗: %w", err)
	}

	// トランザクションコミット
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("トランザクションのコミットに失敗: %w", err)
	}

	slog.InfoContext(ctx, "新規登録確認コードを作成しました",
		"email", email,
	)

	// メール送信ジョブをエンキュー
	if uc.riverClient != nil {
		args, err := worker.BuildSignUpCodeEmail(ctx, email, code, locale)
		if err != nil {
			slog.ErrorContext(ctx, "新規登録確認コードメールの構築に失敗しました",
				"email", email,
				"error", err,
			)
		} else {
			_, err = uc.riverClient.Client().Insert(ctx, *args, nil)
			if err != nil {
				slog.ErrorContext(ctx, "新規登録確認コード送信ジョブのエンキューに失敗しました",
					"email", email,
					"error", err,
				)
			} else {
				slog.InfoContext(ctx, "新規登録確認コード送信ジョブをエンキューしました",
					"email", email,
				)
			}
		}
	} else {
		slog.WarnContext(ctx, "River クライアントが設定されていないため、メール送信ジョブをエンキューできませんでした",
			"email", email,
		)
	}

	return &SendSignUpCodeResult{
		Code:  code,
		Email: email,
	}, nil
}

package usecase

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	annictStripe "github.com/annict/annict/go/internal/stripe"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/validator"
)

// TestCreateCheckoutSessionUsecase_Execute_PropagatesNonNotFoundError verifies that
// the duplicate-subscription check propagates a non-not-found retrieval error instead
// of swallowing it and continuing the checkout. It is a regression test for the
// previous `if err == nil` guard that ignored every GetByID error.
//
// [Ja] 重複サブスクリプションチェックが、not-found 以外の取得エラーを握り潰して checkout を
// 続行せず、そのまま伝播させることをテストする。全 GetByID エラーを無視していた以前の
// `if err == nil` ガードに対する回帰テスト。
func TestCreateCheckoutSessionUsecase_Execute_PropagatesNonNotFoundError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := testutil.GetTestDB()

	// Begin and immediately roll back a transaction so that subsequent queries fail
	// with sql.ErrTxDone (a non-not-found error) rather than returning sql.ErrNoRows.
	// This simulates a genuine retrieval error without depending on connection state.
	//
	// [Ja] トランザクションを開始して即座にロールバックし、以降のクエリが sql.ErrNoRows
	// ではなく sql.ErrTxDone (not-found 以外のエラー) で失敗するようにする。接続状態に
	// 依存せず、本物の取得エラーを再現できる。
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("トランザクションの開始に失敗: %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("トランザクションのロールバックに失敗: %v", err)
	}

	queries := query.New(db).WithTx(tx)
	stripeSubscriberRepo := repository.NewStripeSubscriberRepository(queries)
	cfg := &config.Config{Env: "test", Domain: "test.annict.com"}
	stripeCfg := &annictStripe.Config{
		PriceMonthlyID: "price_monthly_test",
		PriceYearlyID:  "price_yearly_test",
	}
	v := validator.NewSupportersCheckoutCreateValidator()

	// nil Stripe client: the retrieval error must surface before any Stripe call.
	// [Ja] Stripe クライアントは nil。取得エラーは Stripe 呼び出しより前に表面化するはず。
	uc := NewCreateCheckoutSessionUsecase(cfg, stripeSubscriberRepo, stripeCfg, nil, v)

	subscriberID := model.StripeSubscriberID(99999)
	user := &model.User{
		ID:                 model.UserID(1),
		StripeSubscriberID: &subscriberID,
	}

	_, err = uc.Execute(ctx, CreateCheckoutSessionInput{
		User:   user,
		Plan:   "monthly",
		Locale: "ja",
	})
	if err == nil {
		t.Fatal("not-found 以外の取得エラー時はエラーが伝播するべきですが、nil が返りました")
	}

	// The propagated error must wrap the underlying retrieval error, not a misleading
	// downstream error (e.g. missing Stripe client).
	//
	// [Ja] 伝播するエラーは下流の誤解を招くエラー (例: Stripe クライアント未設定) ではなく、
	// 元の取得エラーをラップしているべき。
	if !errors.Is(err, sql.ErrTxDone) {
		t.Errorf("取得エラーがそのまま伝播するべきですが、別のエラーが返りました: %v", err)
	}

	// It must not be treated as a duplicate (Conflict): not-found / errors are distinct
	// from an active subscription.
	//
	// [Ja] 重複 (Conflict) として扱われてはならない。not-found / エラーはアクティブな
	// サブスクリプションとは区別される。
	if appErr := model.AsAppError(err); appErr != nil && appErr.Code == model.AppErrCodeConflict {
		t.Errorf("取得エラーを重複 (Conflict) として扱うべきではありません: %v", err)
	}
}

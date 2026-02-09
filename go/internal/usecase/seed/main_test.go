package seed

import (
	"os"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/annict/annict/go/internal/auth"
)

func TestMain(m *testing.M) {
	// シードデータ生成ではbcryptを多数回呼び出すため、テスト用に低コストに設定
	auth.SetBcryptCostForTest(bcrypt.MinCost)
	os.Exit(m.Run())
}

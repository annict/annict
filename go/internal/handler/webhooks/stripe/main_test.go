package stripe

import (
	"os"
	"testing"

	"github.com/annict/annict/go/internal/testutil"
)

func TestMain(m *testing.M) {
	os.Exit(testutil.SetupTestMain(m))
}

package pgrepo

import (
	"os"
	"testing"

	"github.com/squ1ky/pr-manager/internal/testutils/testdb"
)

func TestMain(m *testing.M) {
	code := m.Run()
	testdb.Teardown()
	os.Exit(code)
}

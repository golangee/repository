package mem

import (
	"github.com/golangee/repository/internal/test"
	"testing"
)

func TestRepository(t *testing.T) {
	var a test.CrudTestRepository[test.A, string]
	a = NewRepository[test.A, string]()
	test.Test(t, test.CreateTestSet1(), a)

	var a2 test.CrudTestRepository[test.B, test.A]
	a2 = NewRepository[test.B, test.A]()
	test.Test(t, test.CreateTestSet2(), a2)

	var a3 test.CrudTestRepository[*test.B, int]
	a3 = NewRepository[*test.B, int]()
	test.Test(t, test.CreateTestSet3(), a3)
}

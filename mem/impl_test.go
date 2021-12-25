package mem

import (
	"github.com/golangee/repository"
	"github.com/golangee/repository/internal/test"
	"testing"
)

type MyEntity struct {
	Stuff string
}

func TestRepository(t *testing.T) {
	var a test.CrudTestRepository[test.A, string]
	a = NewRepository[test.A, string]()
	test.Test(t, test.CreateTestSet1(), a)
}

func TestRepository_Count(t *testing.T) {

	var repo repository.CrudRepository[*MyEntity, string]
	repo = NewRepository[*MyEntity, string]()
	repo.Count()
	entity := &MyEntity{Stuff: "abc"}
	repo.Save("abc", entity)
	t.Log(repo.Count())
	repo.DeleteByID("123")
	t.Log(repo.Count())
	t.Log(repo.FindByID("abc"))
	repo.FindAll(func(id string, e *MyEntity) error {
		t.Log("->", e)
		return nil
	})

	var duckType interface {
		DeleteByID(id string) error
		Count() (int64, error)
		FindByID(id string) (*MyEntity, error)
	}

	duckType = repo
	duckType.DeleteByID("abc")
	t.Log(duckType.Count())

}

func TestRepository_Count2(t *testing.T) {
	var repo repository.CrudRepository[MyEntity, string]
	repo = NewRepository[MyEntity, string]()
	repo.Count()
	entity := MyEntity{Stuff: "abc"}
	repo.Save("abc", entity)
	t.Log(repo.Count())
	repo.DeleteByID("123")
	t.Log(repo.Count())
	t.Log(repo.FindByID("abc"))
	repo.FindAll(func(id string, e MyEntity) error {
		t.Log("->", e)
		return nil
	})
	repo.DeleteByID("abc")
	t.Log(repo.Count())

}

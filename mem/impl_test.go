package mem

import (
	"github.com/golangee/repository"
	"testing"
)

type MyEntity struct {
	ID string
}

func (e MyEntity) GetID() string {
	return e.ID
}

func TestRepository_Count(t *testing.T) {
	var repo repository.CrudRepository[*MyEntity, string]
	repo = NewRepository[*MyEntity, string]()
	repo.Count()
	entity := &MyEntity{ID: "abc"}
	repo.Save(entity)
	t.Log(repo.Count())
	repo.DeleteByID("123")
	t.Log(repo.Count())
	t.Log(repo.FindByID("abc"))
	repo.FindAll(func(e *MyEntity) error {
		t.Log("->", e)
		return nil
	})
	repo.DeleteByID("abc")
	t.Log(repo.Count())

}

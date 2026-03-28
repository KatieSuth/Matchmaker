package store

import (
	"fmt"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

type PostgresStore struct {
	q *db.Queries
}

type Store interface {
	doSomething()
	/*
	   List() []model.Item
	   Get(id int) (model.Item, error)
	   Create(req model.CreateItemRequest) model.Item
	   Update(id int, req model.UpdateItemRequest) (model.Item, error)
	   Delete(id int) error
	*/
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{q: db.New(pool)}
}

func (s *PostgresStore) doSomething() {
	fmt.Print("something")
}

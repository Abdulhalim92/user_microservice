package repository

import (
	"github.com/jackc/pgx/v5"
	"user/internal/logging"
	"user/internal/model"
)

type User interface {
	CreateUser(u *model.User) (int, error)
	GetUser(u *model.User) (*model.User, error)
	ExistsUser(userName string) (bool, error)
}

type Repository struct {
	User
}

func NewRepository(db *pgx.Conn, log *logging.Logger) *Repository {
	return &Repository{
		User: NewUserRepository(db, log),
	}
}

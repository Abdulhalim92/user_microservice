package repository

import (
	"context"
	"github.com/jackc/pgx/v5"
	"user/internal/logging"
	"user/internal/model"
)

type UserRepository struct {
	DbConn *pgx.Conn
	Logger *logging.Logger
}

func NewUserRepository(db *pgx.Conn, log *logging.Logger) *UserRepository {
	return &UserRepository{
		DbConn: db,
		Logger: log,
	}
}

func (r *UserRepository) CreateUser(u *model.User) (int, error) {

	var id int

	sqlQuery := "INSERT INTO users (name, password) VALUES ($1, $2) RETURNING id"

	err := r.DbConn.QueryRow(context.Background(), sqlQuery, u.Name, u.Password).Scan(&id)
	if err != nil {
		r.Logger.Error(err)
		return 0, err
	}

	return id, nil
}

func (r *UserRepository) GetUser(u *model.User) (*model.User, error) {

	var user *model.User

	sqlQuery := "SELECT * FROM users WHERE name = $1 and password = $2"

	err := r.DbConn.QueryRow(context.Background(), sqlQuery, u.Name, u.Password).Scan(&user)
	if err != nil {
		r.Logger.Error(err)
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) ExistsUser(userName string) (bool, error) {

	var id int

	sqlQuery := "SELECT id FROM users WHERE name = $1"

	err := r.DbConn.QueryRow(context.Background(), sqlQuery, userName).Scan(&id)
	if err != nil {
		r.Logger.Error(err)
		return true, err
	}
	if id > 0 {
		r.Logger.Println("such user exists")
		return true, nil
	}

	return false, nil
}

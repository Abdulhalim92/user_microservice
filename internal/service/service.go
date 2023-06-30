package service

import (
	"github.com/go-redis/redis"
	"user/internal/logging"
	"user/internal/model"
	"user/internal/repository"
)

type User interface {
	CreateUser(u *model.User) (int, error)
	GetUser(u *model.User) (int, error)
	SignOut(userID int) error
	GenerateHash(password string) (string, error)
	CompareHashPassword(passFromDb, passFromUser string) error
	ExistsUser(userName string) (bool, error)
}

type Token interface {
	CreateToken(userID int) (*model.TokenDetails, error)
	CreateAuth(userID int, td *model.TokenDetails) error
	DeleteAuth(giveUuid string) (int64, error)
}

type Service struct {
	User
	Token
}

func NewService(rep *repository.Repository, log *logging.Logger, redis *redis.Client) *Service {
	return &Service{
		User:  NewUserService(rep, log),
		Token: NewTokenService(rep, log, redis),
	}
}

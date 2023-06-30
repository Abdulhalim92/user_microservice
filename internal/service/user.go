package service

import (
	"golang.org/x/crypto/bcrypt"
	"user/internal/logging"
	"user/internal/model"
	"user/internal/repository"
)

type UserService struct {
	rep    *repository.Repository
	logger *logging.Logger
}

func NewUserService(rep *repository.Repository, log *logging.Logger) *UserService {
	return &UserService{
		rep:    rep,
		logger: log,
	}
}

func (s *UserService) CreateUser(u *model.User) (int, error) {

	hash, err := s.GenerateHash(u.Password)
	if err != nil {
		s.logger.Error(err)
		return 0, err
	}

	u.Password = hash

	userID, err := s.rep.CreateUser(u)
	if err != nil {
		s.logger.Error(err)
		return 0, err
	}

	return userID, nil
}

func (s *UserService) GetUser(u *model.User) (int, error) {

	user, err := s.rep.GetUser(u)
	if err != nil {
		s.logger.Error(err)
		return 0, err
	}

	err = s.CompareHashPassword(user.Password, u.Password)
	if err != nil {
		s.logger.Error(err)
		return 0, err
	}

	return u.ID, nil
}

func (s *UserService) SignOut(userID int) error {
	return nil // TODO
}

func (s *UserService) ExistsUser(userName string) (bool, error) {
	return s.rep.ExistsUser(userName)
}

func (s *UserService) GenerateHash(password string) (string, error) {

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Errorf("cannot generating password with error: %s", err.Error())
		return "", err
	}

	return string(bytes), nil
}

func (s *UserService) CompareHashPassword(passFromDb, passFromUser string) error {

	err := bcrypt.CompareHashAndPassword([]byte(passFromDb), []byte(passFromUser))
	if err != nil {
		s.logger.Errorf("mismatched password: %s", passFromUser)
		return err
	}

	return nil
}

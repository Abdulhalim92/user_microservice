package service

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis"
	"github.com/twinj/uuid"
	"os"
	"time"
	"user/internal/logging"
	"user/internal/model"
	"user/internal/repository"
)

type TokenService struct {
	rep    *repository.Repository
	logger *logging.Logger
	redis  *redis.Client
}

func NewTokenService(rep *repository.Repository, log *logging.Logger, redis *redis.Client) *TokenService {
	return &TokenService{
		rep:    rep,
		logger: log,
		redis:  redis,
	}
}

func (s *TokenService) CreateToken(userID int) (*model.TokenDetails, error) {

	var td model.TokenDetails

	td.AtExpires = time.Now().Add(15 * time.Minute).Unix()
	td.AccessUuid = uuid.NewV4().String()

	td.RtExpires = time.Now().Add(time.Hour * 24 * 7).Unix()
	td.RefreshUuid = uuid.NewV4().String()

	// Creating Access Token
	err := os.Setenv("ACCESS_SECRET", "secret")
	if err != nil {
		return nil, err
	}

	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["access_uuid"] = td.AccessUuid
	atClaims["user_id"] = userID
	atClaims["exp"] = td.AtExpires
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString([]byte(os.Getenv("ACCESS_SECRET")))
	if err != nil {
		return nil, err
	}

	// Creating Refresh Token
	err = os.Setenv("REFRESH_SECRET", "secret")
	if err != nil {
		return nil, err
	}
	rtClaims := jwt.MapClaims{}
	rtClaims["refresh_uuid"] = td.RefreshUuid
	rtClaims["user_id"] = userID
	rtClaims["exp"] = td.RtExpires
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = rt.SignedString([]byte(os.Getenv("ACCESS_SECRET")))
	if err != nil {
		return nil, err
	}

	return &td, nil

}

func (s *TokenService) CreateAuth(userID int, td *model.TokenDetails) error {

	at := time.Unix(td.AtExpires, 0) // converting Unix to UTC(to Time object)
	rt := time.Unix(td.RtExpires, 0)
	now := time.Now()

	errAccess := s.redis.Set(td.AccessUuid, userID, at.Sub(now)).Err()
	if errAccess != nil {
		s.logger.Error(errAccess)
		return errAccess
	}
	errRefresh := s.redis.Set(td.RefreshUuid, userID, rt.Sub(now)).Err()
	if errRefresh != nil {
		s.logger.Error(errRefresh)
		return errRefresh
	}

	return nil
}

func (s *TokenService) DeleteAuth(giveUuid string) (int64, error) {
	deleted, err := s.redis.Del(giveUuid).Result()
	if err != nil {
		s.logger.Error(err)
		return 0, err
	}

	return deleted, nil
}

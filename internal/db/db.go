package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"user/config"
	"user/internal/logging"
)

func InitDb(cfg config.DbCfg) (*pgx.Conn, error) {

	log := logging.GetLogger()

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DbName, cfg.Sslmode)

	dbConn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		log.Fatalf("cannot to connect to database: %v", err)
		return nil, err
	}

	return dbConn, nil
}

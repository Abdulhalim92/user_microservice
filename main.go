package main

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/nats-io/nats.go"
	"net"
	"os"
	"user/config"
	"user/internal/db"
	"user/internal/handler"
	"user/internal/logging"
	"user/internal/redis"
	"user/internal/repository"
	"user/internal/service"
)

func main() {
	log := logging.GetLogger()

	cfg := config.GetConfig()

	nc, err := nats.Connect(net.JoinHostPort(cfg.BrokerCfg.Host, cfg.BrokerCfg.Port), nats.Name("user service"))
	if err != nil {
		log.Fatal(err)
	}

	redisClient, err := redis.InitRedis(cfg.RedisCfg)
	if err != nil {
		log.Println(err)
	}

	pgxConn, err := db.InitDb(cfg.DbCfg)
	if err != nil {
		log.Println(err)
	}
	defer pgxConn.Close(context.Background())

	migrate(pgxConn)

	newRepository := repository.NewRepository(pgxConn, log)

	newService := service.NewService(newRepository, log, redisClient)

	newHandler := handler.NewHandler(nc, log, newService)
	newHandler.Init()

}

func migrate(pgxConn *pgx.Conn) {

	var exists bool

	migrateBytes, err := os.ReadFile("./pkg/schemas/inits.sql")
	if err != nil {
		logging.GetLogger().Fatal(err)
	}

	err = pgxConn.QueryRow(context.Background(), "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = $1)", "users").Scan(&exists)
	if err != nil {
		logging.GetLogger().Fatal(err)
	}
	if !exists {
		_, err := pgxConn.Exec(context.Background(), string(migrateBytes))
		if err != nil {
			logging.GetLogger().Fatal(err)
		}
	}
}

package main

import (
	"github.com/nats-io/nats.go"
	"net"
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

	newRepository := repository.NewRepository(pgxConn, log)

	newService := service.NewService(newRepository, log, redisClient)

	newHandler := handler.NewHandler(nc, log, newService)
	newHandler.Init()

}

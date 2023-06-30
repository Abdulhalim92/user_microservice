package redis

import (
	"github.com/go-redis/redis"
	"net"
	"user/config"
	"user/internal/logging"
)

func InitRedis(cfg config.RedisCfg) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr: net.JoinHostPort(cfg.Host, cfg.Port),
	})
	_, err := client.Ping().Result()
	if err != nil {
		logging.GetLogger().Fatal(err)
		return nil, err
	}

	return client, nil
}

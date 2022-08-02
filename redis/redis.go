package redis

import (
	"os"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

const Nil = redis.Nil

// channels: name[string]: {
// 	Url: string,
// 	IsUsed: boolean,
//  UserId: string,
// }
// locations: name[string]: {
// 	Longitude: double,
// 	Latitude:  double,
// 	Timestamp: time,
// }

func New() (*redis.Client, error) {
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	client := redis.NewClient(&redis.Options{
		Addr: redisHost + ":" + redisPort,
	})
	if err := client.Ping().Err(); err != nil {
		return nil, errors.Wrapf(err, "failed to ping redis server")
	}
	return client, nil
}

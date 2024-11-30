package initializer

import (
	"context"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client
var RCTX = context.Background()

func RedisInit() {
	database := 0
	database, err := strconv.Atoi(os.Getenv("REDIS_DATABASE"))
	if err != nil {
		database = 0
	}
	RDB = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),
		Password: os.Getenv("REDIS_PASSWORD"), // no password set
		DB:       database,                    // use default DB
		Protocol: 3,                           // specify 2 for RESP 2 or 3 for RESP 3
	})
}

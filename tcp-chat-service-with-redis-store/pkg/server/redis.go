package server

import (
	"errors"
	"log"

	re "github.com/go-redis/redis"
	"github.com/shameerb/tcp-chat-redis/pkg/common"
)

type redis struct {
	client *re.Client
	pubsub *re.PubSub
}

func initRedis(redisAddr string) (*redis, error) {
	log.Println("Initializing redis client")
	client := re.NewClient(&re.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
	})
	_, err := client.Ping().Result()
	if err != nil {
		log.Printf("Error connecting to redis: %s", err.Error())
		return nil, err
	}
	pubsub := client.Subscribe(common.CHANNEL)
	return &redis{client: client, pubsub: pubsub}, nil
}

func (r *redis) exists(key string) bool {
	v := r.client.Exists(key)
	return v.Val() == int64(1)
}

func (r *redis) set(key, value string) bool {
	bc := r.client.SetNX(key, value, 0)
	return bc.Val()
}

func (r *redis) delete(key string) error {
	v := r.client.Del(key)
	if v.Val() != int64(1) {
		return errors.New("user didn't exist")
	}
	return nil
}

func (r *redis) publish(msg string) error {
	return r.client.Publish(common.CHANNEL, msg).Err()
}

func (r *redis) search(key string) ([]string, error) {
	// keys is resource intensive. Use scan instead.
	// run a for loop with pagination using scan. 		keys, nextCursor, err := client.Scan(context.Background(), cursor, match, int64(count)).Result()
	keys, err := r.client.Keys(key).Result()
	if err != nil {
		return nil, err
	}

	return keys, nil

}

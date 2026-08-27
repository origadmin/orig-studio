package pubsub

import (
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/redis/go-redis/v9"
)

func NewRedisStreams(redisAddr string, logger watermill.LoggerAdapter) *PubSub {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	subscriber, err := redisstream.NewSubscriber(
		redisstream.SubscriberConfig{
			Client:        rdb,
			Consumer:      "transcode-service",
			ConsumerGroup: "transcode-workers",
			OldestId:      "$",
			BlockTime:     5 * time.Second,
			ClaimInterval: 60 * time.Second,
			MaxIdleTime:   30 * time.Minute,
		},
		logger,
	)
	if err != nil {
		panic(err)
	}

	publisher, err := redisstream.NewPublisher(
		redisstream.PublisherConfig{
			Client:        rdb,
			DefaultMaxlen: 10000,
		},
		logger,
	)
	if err != nil {
		panic(err)
	}

	return &PubSub{
		Pub: publisher,
		Sub: subscriber,
	}
}

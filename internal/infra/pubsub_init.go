package infra

import (
	"os"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/origadmin/runtime/log"

	"origadmin/application/origstudio/internal/infra/pubsub"
)

func NewPubSub(logger log.Logger) *pubsub.PubSub {
	wmLogger := watermill.NewStdLogger(true, true)
	if redisAddr := os.Getenv("WATERMILL_REDIS_ADDR"); redisAddr != "" {
		log.Infof("Watermill: using Redis Streams pub/sub (addr=%s)", redisAddr)
		return pubsub.NewRedisStreams(redisAddr, wmLogger)
	}
	log.Infof("Watermill: using GoChannel (in-process) pub/sub")
	return pubsub.NewGoChannel(64, wmLogger)
}

func NewPublisher(ps *pubsub.PubSub) message.Publisher {
	return ps.Pub
}

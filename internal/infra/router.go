/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package infra

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/origadmin/runtime/log"

	"origadmin/application/origstudio/internal/infra/pubsub"
	mediabiz "origadmin/application/origstudio/internal/features/media/biz"
)

// NewRouter creates a new Watermill router.
func NewRouter(
	transcodeHandler *mediabiz.TranscodeHandler,
	ps *pubsub.PubSub,
	logger log.Logger,
) (*message.Router, error) {
	wmLogger := watermill.NewStdLogger(true, true)
	router, err := message.NewRouter(message.RouterConfig{}, wmLogger)
	if err != nil {
		return nil, err
	}
	router.AddHandler(
		"media_transcode",
		pubsub.MediaEncodeRequestTopic,
		ps.Sub,
		"",  // no output topic (handler publishes directly)
		nil, // no output publisher needed
		func(msg *message.Message) ([]*message.Message, error) {
			return nil, transcodeHandler.Handle(msg)
		},
	)
	return router, nil
}

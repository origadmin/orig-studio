/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package infra

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/origadmin/runtime/log"

	"origadmin/application/origstudio/internal/infra/pubsub"
)

// NewPubSub creates a new pubsub instance.
func NewPubSub(logger log.Logger) *pubsub.PubSub {
	wmLogger := watermill.NewStdLogger(true, true)
	return pubsub.NewGoChannel(64, wmLogger)
}

// NewPublisher creates a new message publisher.
func NewPublisher(ps *pubsub.PubSub) message.Publisher {
	return ps.Pub
}

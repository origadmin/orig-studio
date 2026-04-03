/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package pubsub provides pub/sub topic definitions and factory functions.
package pubsub

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
)

// User event topics
const (
	UserRoleAssignedTopic = "user.role.assigned"
	UserCreatedTopic      = "user.created"
	UserDeletedTopic      = "user.deleted"
)

// Media encoding event topics
const (
	MediaEncodeRequestTopic   = "media.encode.request"   // triggers transcoding pipeline
	MediaEncodeProgressTopic  = "media.encode.progress"  // progress updates per profile
	MediaEncodeCompletedTopic = "media.encode.completed" // all profiles done for a media
)

// EventType constants used in message payloads for SSE consumers.
const (
	EventTypeEncodeProgress  = "encode.progress"
	EventTypeEncodeCompleted = "encode.completed"
	EventTypeEncodeFailed    = "encode.failed"
)

// NewMessage creates a new Watermill message with a generated UUID.
func NewMessage(payload []byte) *message.Message {
	return message.NewMessage(watermill.NewUUID(), payload)
}

// PubSub wraps a Watermill publisher and subscriber.
type PubSub struct {
	Pub message.Publisher
	Sub message.Subscriber
}

// NewGoChannel creates an in-process publisher/subscriber using GoChannel.
// Zero external dependencies — messages are delivered within the same process.
func NewGoChannel(bufferSize int64, logger watermill.LoggerAdapter) *PubSub {
	pubSub := gochannel.NewGoChannel(
		gochannel.Config{
			OutputChannelBuffer: bufferSize,
		},
		logger,
	)
	return &PubSub{
		Pub: pubSub,
		Sub: pubSub,
	}
}

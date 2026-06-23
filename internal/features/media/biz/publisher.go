// Copyright (c) 2024 OrigAdmin. All rights reserved.

package biz

// EventPublisher defines the interface for publishing async events.
// This abstracts the underlying message queue (e.g., watermill) to avoid
// direct coupling between shared code and EE-only infrastructure.
type EventPublisher interface {
	// Publish sends a payload to a topic. Returns nil on success.
	Publish(topic string, payload []byte) error
}

// MediaEventPublisher defines the interface for publishing media encoding events.
// Used by transcode workers to report progress/completion.
type MediaEventPublisher interface {
	Publish(mediaID string, event *EncodingEvent)
}

// MediaEncodeRequestTopic is the topic name for media encoding requests.
const MediaEncodeRequestTopic = "media.encode.request"

// noopPublisher is a default no-op implementation used when no publisher is available.
type noopPublisher struct{}

func (n *noopPublisher) Publish(topic string, payload []byte) error {
	return nil
}

// NewNoopPublisher returns a publisher that discards all messages.
// Useful for tests and CE builds that don't have async processing.
func NewNoopPublisher() EventPublisher {
	return &noopPublisher{}
}

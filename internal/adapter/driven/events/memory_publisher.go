package events

import (
	"context"
	"log"

	"spin-transactions/internal/application"
	domainevents "spin-transactions/internal/domain/events"
)

var _ application.EventPublisher = (*MemoryPublisher)(nil)

// MemoryPublisher is an in-process driven adapter. It does not imply a message bus.
type MemoryPublisher struct{}

func NewMemoryPublisher() *MemoryPublisher { return &MemoryPublisher{} }

func (p *MemoryPublisher) Publish(_ context.Context, event domainevents.DomainEvent) error {
	log.Printf("domain event: %s at %s", event.Name(), event.OccurredAt().Format("2006-01-02T15:04:05Z07:00"))
	return nil
}

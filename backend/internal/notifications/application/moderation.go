package application

import "context"

// ModerationMessage is a snapshot of a submission, independent of later decisions.
type ModerationMessage struct {
	ID                                  int64
	Kind, EntityID, Label, Detail, Path string
}
type ModerationSender interface {
	Send(context.Context, ModerationMessage) error
}
type ModerationQueue interface {
	DeliverModeration(context.Context, ModerationSender) (bool, error)
}
type ModerationNotifier struct {
	queue  ModerationQueue
	sender ModerationSender
}

func NewModerationNotifier(queue ModerationQueue, sender ModerationSender) *ModerationNotifier {
	return &ModerationNotifier{queue: queue, sender: sender}
}
func (n *ModerationNotifier) Deliver(ctx context.Context) (bool, error) {
	return n.queue.DeliverModeration(ctx, n.sender)
}

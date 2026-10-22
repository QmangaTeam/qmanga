package application

import "context"

// DigestSender is compatible with Identity's configured mail transport.
type DigestSender interface {
	Send(context.Context, string, string, string) error
}
type DigestQueue interface {
	DeliverDigest(context.Context, DigestSender, string) (bool, error)
}
type Digest struct {
	queue  DigestQueue
	sender DigestSender
	origin string
}

func NewDigest(queue DigestQueue, sender DigestSender, origin string) *Digest {
	return &Digest{queue, sender, origin}
}
func (d *Digest) Deliver(ctx context.Context) (bool, error) {
	return d.queue.DeliverDigest(ctx, d.sender, d.origin)
}

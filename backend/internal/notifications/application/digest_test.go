package application

import (
	"context"
	"errors"
	"testing"
)

type digestQueueProbe struct {
	sender DigestSender
	origin string
	err    error
}

func (p *digestQueueProbe) DeliverDigest(_ context.Context, s DigestSender, o string) (bool, error) {
	p.sender = s
	p.origin = o
	return true, p.err
}

type mailProbe struct{}

func (mailProbe) Send(context.Context, string, string, string) error { return nil }
func TestDigestUsesInjectedTransportAndPropagatesFailure(t *testing.T) {
	expected := errors.New("mail unavailable")
	queue := &digestQueueProbe{err: expected}
	sender := mailProbe{}
	worked, err := NewDigest(queue, sender, "https://qmanga.test").Deliver(context.Background())
	if !worked || !errors.Is(err, expected) || queue.sender != sender || queue.origin != "https://qmanga.test" {
		t.Fatal("digest composition lost port result")
	}
}

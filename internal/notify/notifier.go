// Package notify sends formatted digests to Telegram and other channels.
package notify

import "context"

// Notifier delivers a rendered message to an external sink.
type Notifier interface {
	Notify(ctx context.Context, message string) error
}

// StubNotifier is a phase-0 placeholder that performs no network I/O.
type StubNotifier struct{}

// Notify implements Notifier.
func (StubNotifier) Notify(ctx context.Context, _ string) error {
	return ctx.Err()
}

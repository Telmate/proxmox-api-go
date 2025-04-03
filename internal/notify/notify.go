package notify

import "sync"

// Channel is a structure that mimics a context for cancellation.
type Channel struct {
	doneCh chan struct{}
	once   sync.Once
}

// New initializes and returns a new Channel.
func New() *Channel {
	return &Channel{
		doneCh: make(chan struct{}),
	}
}

// Done returns a channel that gets closed when Close() is called.
func (c *Channel) Done() <-chan struct{} {
	return c.doneCh
}

// Close safely closes the channel, ensuring it is closed only once.
func (c *Channel) Close() {
	c.once.Do(func() {
		close(c.doneCh)
	})
}

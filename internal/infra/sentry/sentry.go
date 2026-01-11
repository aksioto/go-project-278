package sentry

import (
	"time"

	"github.com/getsentry/sentry-go"
)

type Client struct {
	active bool
	flush  func()
}

func Init(opts sentry.ClientOptions, flushTimeout time.Duration) (*Client, error) {
	if opts.Dsn == "" {
		return &Client{active: false}, nil
	}

	if err := sentry.Init(opts); err != nil {
		return nil, err
	}

	return &Client{
		active: true,
		flush: func() {
			sentry.Flush(flushTimeout)
		},
	}, nil
}

func (c *Client) Close() {
	if c.active && c.flush != nil {
		c.flush()
	}
}

func (c *Client) Active() bool {
	return c.active
}

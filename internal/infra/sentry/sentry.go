package sentry

import (
	"time"

	"github.com/getsentry/sentry-go"
)

var sensitiveHeaders = []string{
	"Authorization",
	"Cookie",
	"Set-Cookie",
	"X-Auth-Token",
	"X-Api-Key",
}

type Client struct {
	enabled      bool
	flushTimeout time.Duration
}

func Init(opts sentry.ClientOptions, flushTimeout time.Duration) (*Client, error) {
	if opts.Dsn == "" {
		return &Client{enabled: false}, nil
	}

	if opts.BeforeSend == nil {
		opts.BeforeSend = sanitizeEvent
	}

	if err := sentry.Init(opts); err != nil {
		return nil, err
	}

	return &Client{
		enabled:      true,
		flushTimeout: flushTimeout,
	}, nil
}

func sanitizeEvent(event *sentry.Event, _ *sentry.EventHint) *sentry.Event {
	if event.Request != nil && event.Request.Headers != nil {
		for _, header := range sensitiveHeaders {
			if _, exists := event.Request.Headers[header]; exists {
				event.Request.Headers[header] = "[FILTERED]"
			}
		}
	}
	return event
}

func (c *Client) Close() {
	if c.enabled {
		sentry.Flush(c.flushTimeout)
	}
}

func (c *Client) Active() bool {
	return c.enabled
}

// Package client implements the Translaas SDK HTTP client matching .NET TranslaasClient.
package client

import (
	"context"
	"time"
)

const (
	// DefaultTimeout is applied when Options.Timeout is zero.
	DefaultTimeout = 30 * time.Second

	sdkTranslationsPrefix = "sdk/v1/translations"
)

// Client is the consumer-facing HTTP client boundary.
type Client interface {
	GetEntry(ctx context.Context, group, entry, lang string, opts ...GetEntryOption) (string, error)
}

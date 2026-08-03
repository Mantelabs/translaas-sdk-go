package cachefile

import (
	"context"
	"errors"
	"net"

	"github.com/Mantelabs/translaas-sdk-go/models"
)

func isNetworkOrAPIError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) {
		return false
	}

	var apiErr *models.APIError
	if errors.As(err, &apiErr) {
		return true
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	var netErr net.Error
	return errors.As(err, &netErr)
}

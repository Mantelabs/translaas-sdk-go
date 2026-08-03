package client

import (
	"context"
	"strings"

	"github.com/Mantelabs/translaas-sdk-go/models"
)

// NewWithResolvedProject constructs a Client, resolving DefaultProjectID from ValidateAPIKey
// when it is not configured and the API key is scoped to a single project.
func NewWithResolvedProject(ctx context.Context, opts Options, optFns ...Option) (Client, error) {
	if strings.TrimSpace(opts.DefaultProjectID) != "" {
		return New(opts, optFns...)
	}

	cli, err := New(opts, optFns...)
	if err != nil {
		return nil, err
	}

	validate, err := cli.ValidateAPIKey(ctx)
	if err != nil {
		return nil, err
	}

	projectID := models.ReadJSONULID(validate.ProjectID)
	if projectID == "" {
		return cli, nil
	}

	opts.DefaultProjectID = projectID
	return New(opts, optFns...)
}

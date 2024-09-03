package client

import (
	"context"
)

type ServiceImageUpdater interface {
	UpdateImage(ctx context.Context, serviceId, imageURL string) error
}

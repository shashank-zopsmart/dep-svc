package client

import "gofr.dev/pkg/gofr"

type ServiceImageUpdater interface {
	UpdateImage(ctx *gofr.Context, serviceId, imageURL string, serviceCreds any) error
}

package kops

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	gofrSvc "gofr.dev/pkg/gofr/service"
)

type client struct {
	kopsSvc gofrSvc.HTTP
}

func New(svc gofrSvc.HTTP) *client {
	return &client{kopsSvc: svc}
}

type imageUpdate struct {
	Image         string      `json:"image"`
	DeploymentKey interface{} `json:"deploymentKey"`
}

var errService = errors.New("non OK status code received")

func (c *client) UpdateImage(ctx context.Context, serviceId, imageURL string, serviceCreds any) error {
	api := fmt.Sprintf("service/%s/image", serviceId)

	payload, _ := json.Marshal(imageUpdate{Image: imageURL, DeploymentKey: serviceCreds})

	resp, err := c.kopsSvc.Put(ctx, api, nil, payload)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errService
	}

	return nil
}

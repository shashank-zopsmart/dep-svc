package kops

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"gofr.dev/pkg/gofr"
	gofrSvc "gofr.dev/pkg/gofr/service"
)

type client struct {
	kopsSvc gofrSvc.HTTP
}

func New(svc gofrSvc.HTTP) *client {
	return &client{kopsSvc: svc}
}

type errorResp struct {
	Errors any `json:"errors"`
}

type imageUpdate struct {
	Image         string      `json:"image"`
	DeploymentKey interface{} `json:"deploymentKey"`
}

var errService = errors.New("non OK status code received")

func (c *client) UpdateImage(ctx *gofr.Context, serviceId, imageURL string, serviceCreds any) error {
	api := fmt.Sprintf("cli/service/%s/image", serviceId)

	payload, _ := json.Marshal(imageUpdate{Image: imageURL, DeploymentKey: serviceCreds})

	resp, err := c.kopsSvc.Put(ctx, api, nil, payload)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var er errorResp

		b, _ := io.ReadAll(resp.Body)
		_ = json.Unmarshal(b, &er)

		ctx.Logger.Errorf("error response from kops api, err : %v", er)

		return errService
	}

	return nil
}

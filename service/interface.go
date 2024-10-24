package service

import (
	"gofr.dev/pkg/gofr"

	"kops-deploy-service/models"
)

type ImageUploader interface {
	UploadToArtifactory(ctx *gofr.Context, img *models.ImageDetails) (string, error)
}

type ImageDeployer interface {
	DeployImage(ctx *gofr.Context, serviceID, imageURL string, serviceCreds any) error
}

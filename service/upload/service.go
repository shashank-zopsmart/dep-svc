package upload

import (
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/google/uuid"
	"gofr.dev/pkg/gofr"
	"os"
	"strings"

	"kops-deploy-service/models"
)

type service struct {
}

func New() *service {
	return &service{}
}

func (s *service) UploadToArtifactory(ctx *gofr.Context, img *models.ImageDetails) (string, error) {
	dir := getUniqueDir()
	defer os.RemoveAll(dir)

	err := img.Data.CreateLocalCopies(dir)
	if err != nil {
		return "", err
	}

	path, err := pushImage(ctx, img, dir)
	if err != nil {
		return "", err
	}

	ctx.Logger.Infof("successfully pushed image %v to artifact registry", img.Name)

	return path, nil
}

func pushImage(ctx *gofr.Context, img *models.ImageDetails, path string) (string, error) {
	var (
		imagePath string
		auth      *authn.Basic
	)

	switch strings.ToUpper(img.CloudProvider) {
	case GCP:
		googleReg, err := NewGCR(img)
		if err != nil {
			return "", err
		}

		imagePath = googleReg.getImagePath(img)
		auth, err = googleReg.getAuth(ctx)
		if err != nil {
			return "", err
		}
	case AWS:
		awsReg, err := NewECR(img)
		if err != nil {
			return "", err
		}

		imagePath = awsReg.getImagePath(img)
		auth, err = awsReg.getAuth(ctx, &img.ServiceDetails)
		if err != nil {
			return "", err
		}
	}

	ref, err := name.ParseReference(imagePath)
	if err != nil {
		return "", err
	}

	imgTar, err := tarball.ImageFromPath(path+"/temp/"+img.Name+img.Tag+".tar", nil)
	if err != nil {
		return "", err
	}

	// Push the image to the specified registry
	err = remote.Write(ref, imgTar, remote.WithAuth(auth))
	if err != nil {
		return "", err
	}

	return imagePath, nil
}

func getUniqueDir() string {
	dirName, _ := uuid.NewUUID()
	return dirName.String()
}

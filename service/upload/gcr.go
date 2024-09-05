package upload

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/go-containerregistry/pkg/authn"
	"gofr.dev/pkg/gofr"
	"golang.org/x/oauth2/google"

	"kops-deploy-service/models"
)

const GCP = "GCP"

var errIncorrectCloud = errors.New("could not fetch proper credential details for service")

type gcr struct {
	creds *models.GCPCreds
}

func NewGCR(creds *models.ImageDetails) (*gcr, error) {
	var gcpCreds models.GCPCreds

	credString, ok := creds.ServiceCreds.(string)
	if !ok {
		return nil, errIncorrectCloud
	}

	err := json.Unmarshal([]byte(credString), &gcpCreds)
	if err != nil {
		return nil, errIncorrectCloud
	}

	creds.ServiceCreds = gcpCreds

	return &gcr{creds: &gcpCreds}, nil
}

func (g *gcr) getImagePath(img *models.ImageDetails) string {
	return fmt.Sprintf("%s-docker.pkg.dev/%s/kops-dev/%s/%s:%s", img.Region, g.creds.ProjectID, img.Repository, img.Name, img.Tag)
}

func (g *gcr) getAuth(ctx *gofr.Context) (*authn.Basic, error) {
	b, _ := json.Marshal(g.creds)

	creds, err := google.CredentialsFromJSON(ctx, b, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return nil, err
	}

	token, err := creds.TokenSource.Token()
	if err != nil {
		return nil, err
	}

	return &authn.Basic{
		Username: "oauth2accesstoken",
		Password: token.AccessToken,
	}, nil
}

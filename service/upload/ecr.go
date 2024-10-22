package upload

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	ecrSvc "github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/google/go-containerregistry/pkg/authn"
	"gofr.dev/pkg/gofr"

	"kops-deploy-service/models"
)

const AWS = "AWS"

type ecr struct {
	creds *models.AWSCreds
}

func NewECR(img *models.ImageDetails) (*ecr, error) {
	var awsCreds models.AWSCreds

	credString, ok := img.ServiceCreds.(string)
	if !ok {
		return nil, errIncorrectCloud
	}

	err := json.Unmarshal([]byte(credString), &awsCreds)
	if err != nil {
		return nil, errIncorrectCloud
	}

	img.ServiceCreds = awsCreds

	return &ecr{creds: &awsCreds}, nil
}

func (e *ecr) getImagePath(img *models.ImageDetails) string {
	return fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/%s:%s-%s:%s",
		img.ServiceDetails.AccountID, img.ServiceDetails.Region, img.ServiceDetails.Repository, img.ServiceDetails.ServiceName, img.Name, img.Tag)
}

func (e *ecr) getAuth(ctx *gofr.Context, sd *models.ServiceDetails) (*authn.Basic, error) {
	// Set up the custom credentials provider
	customCredentials := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(e.creds.AccessKey, e.creds.AccessSecret, ""))

	// Load the AWS config with the custom credentials provider
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(sd.Region),
		config.WithCredentialsProvider(customCredentials),
	)
	if err != nil {
		return nil, err
	}

	// Create an ECR client
	svc := ecrSvc.NewFromConfig(cfg)

	// Get the ECR authorization token
	tokenOutput, err := svc.GetAuthorizationToken(ctx, &ecrSvc.GetAuthorizationTokenInput{})
	if err != nil {
		return nil, err
	}

	if len(tokenOutput.AuthorizationData) == 0 {
		return nil, err
	}

	// Decode the authorization token
	authToken := *tokenOutput.AuthorizationData[0].AuthorizationToken

	decodedToken, err := base64.StdEncoding.DecodeString(authToken)
	if err != nil {
		return nil, err
	}

	tokenParts := strings.SplitN(string(decodedToken), ":", 2)
	if len(tokenParts) != 2 {
		return nil, err
	}

	return &authn.Basic{
		Username: tokenParts[0],
		Password: tokenParts[1],
	}, nil
}

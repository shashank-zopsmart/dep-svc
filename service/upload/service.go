package upload

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	pkgPath "path"
	"strings"

	"gofr.dev/pkg/gofr"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/google/uuid"

	"kops-deploy-service/models"
)

type service struct {
	docker *client.Client
}

func New(docker *client.Client) *service {
	return &service{docker: docker}
}

func (s *service) UploadToArtifactory(ctx *gofr.Context, img *models.ImageDetails) (string, error) {
	dir := getUniqueDir()
	defer os.RemoveAll(dir)

	err := img.Data.CreateLocalCopies(dir)
	if err != nil {
		return "", err
	}

	repoPath := pkgPath.Join(dir, img.ModuleName)

	err = buildProject(repoPath, img.Lang, ctx.Logger)
	if err != nil {
		return "", err
	}

	err = s.buildImage(ctx, img, repoPath)
	if err != nil {
		return "", err
	}

	err = s.saveImage(ctx, img, repoPath)
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

type buildOutput struct {
	Stream string `json:"stream"`
	Status string `json:"status"`
	Error  string `json:"error"`
}

func (s *service) buildImage(ctx *gofr.Context, img *models.ImageDetails, path string) error {
	curWd, err := os.Getwd()
	if err != nil {
		return err
	}

	defer os.Chdir(curWd)

	os.Chdir(path)

	buildContext, err := archive.TarWithOptions(".", &archive.TarOptions{})
	if err != nil {
		ctx.Errorf("unable to generate the build context for current project, error : %v", err)

		return err
	}

	options := types.ImageBuildOptions{
		SuppressOutput: false,
		Remove:         true,
		Tags:           []string{img.Name + ":" + img.Tag},
		Dockerfile:     "/Dockerfile",
	}

	imageBuildResponse, err := s.docker.ImageBuild(ctx, buildContext, options)
	if err != nil {
		ctx.Errorf("unable to buield docker image, error : %v", err)

		return err
	}

	defer imageBuildResponse.Body.Close()

	// Decode and print formatted output
	decoder := json.NewDecoder(imageBuildResponse.Body)

	for {
		var output buildOutput
		if er := decoder.Decode(&output); er == io.EOF {
			break
		} else if er != nil {
			ctx.Error(er)
		}

		if output.Stream != "" && output.Stream != `\n'` {
			ctx.Debug(output.Stream)
		}

		if output.Status != "" {
			ctx.Info(output.Status)
		}

		if output.Error != "" {
			ctx.Debug("Error: %s\n", output.Error)
		}
	}

	return nil
}

func (s *service) saveImage(ctx *gofr.Context, img *models.ImageDetails, baseDir string) error {
	tarFileName := pkgPath.Join(baseDir, img.Name+":"+img.Tag+".tar")

	tarFile, err := os.Create(tarFileName)
	if err != nil {
		ctx.Errorf("unable to create the image tar file, error : %v", err)
		return err
	}
	defer tarFile.Close()

	reader, err := s.docker.ImageSave(ctx, []string{img.Name + ":" + img.Tag})
	if err != nil {
		ctx.Errorf("uanble to create save image, error : %v", err)
		return err
	}
	defer reader.Close()

	// Write the image data to the tar file
	_, err = io.Copy(tarFile, reader)
	if err != nil {
		ctx.Logger.Errorf("unable to save image to file, error : %v", err)
		return err
	}

	return nil
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

	filepath := pkgPath.Join(path, img.ModuleName, fmt.Sprintf("%s:%s.tar", img.Name, img.Tag))

	imgTar, err := tarball.ImageFromPath(filepath, nil)
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

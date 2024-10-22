package main

import (
	"github.com/docker/docker/client"
	"gofr.dev/pkg/gofr"
	gofrSvc "gofr.dev/pkg/gofr/service"
	"kops-deploy-service/client/kops"
	"kops-deploy-service/handler/deploy"
	depSvc "kops-deploy-service/service/deploy"
	"kops-deploy-service/service/upload"
)

func main() {
	//os.Setenv("DOCKER_HOST", "unix:///Users/raramuri/.colima/default/docker.sock")
	app := gofr.New()

	kopsClient := gofrSvc.NewHTTPService(app.Config.Get("KOPS_SERVICE_URL"), app.Logger(), app.Metrics(),
		&gofrSvc.DefaultHeaders{
			Headers: map[string]string{
				"cli-api-key": app.Config.Get("KOPS_API_KEY"),
			},
		})

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		app.Logger().Errorf("failed to initialise docker client: %v", err)

		return
	}

	kopsSvc := kops.New(kopsClient)
	uploadSvc := upload.New(dockerClient)
	deploySvc := depSvc.New(kopsSvc)

	deployHndlr := deploy.New(uploadSvc, deploySvc)

	app.POST("/deploy", deployHndlr.UploadImage)

	app.Run()
}

package main

import (
	"gofr.dev/pkg/gofr"
	gofrSvc "gofr.dev/pkg/gofr/service"

	"kops-deploy-service/client/kops"
	"kops-deploy-service/handler/deploy"
	depSvc "kops-deploy-service/service/deploy"
	"kops-deploy-service/service/upload"
)

func main() {
	app := gofr.New()

	kopsClient := gofrSvc.NewHTTPService(app.Config.Get("KOPS_SERVICE_URL"), app.Logger(), app.Metrics(),
		&gofrSvc.DefaultHeaders{
			Headers: map[string]string{
				"cli-api-key": app.Config.Get("KOPS_API_KEY"),
			},
		})

	kopsSvc := kops.New(kopsClient)
	uploadSvc := upload.New()
	deploySvc := depSvc.New(kopsSvc)

	deployHndlr := deploy.New(uploadSvc, deploySvc)

	app.POST("/", deployHndlr.UploadImage)

	app.Run()
}

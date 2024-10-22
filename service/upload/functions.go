package upload

import (
	"fmt"
	"os"
	"os/exec"

	"gofr.dev/pkg/gofr/logging"
)

const (
	golang = "golang"
	java   = "java"
	js     = "js"
)

func buildProject(path, lang string, logger logging.Logger) error {
	err := Build(path, lang, logger)
	if err != nil {
		logger.Errorf("error while building the project binary, please check the project code!")

		return err
	}

	return nil
}

// Build executes the build command for the project specific to language.
func Build(path, lang string, logger logging.Logger) error {
	switch lang {
	case golang:
		return buildGolang(path, logger)
	case js:
		// TODO: necessary steps for javascript build
		break
	case java:
		// TODO: necessary steps for building java projects
		break
	}

	return nil
}

func buildGolang(path string, logger logging.Logger) error {
	fmt.Println("Creating binary for the project")

	curWd, err := os.Getwd()
	if err != nil {
		return err
	}

	err = os.Chdir(path)
	if err != nil {
		return err
	}

	defer os.Chdir(curWd)

	output, err := exec.Command("sh", "-c", "CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .").CombinedOutput()
	if err != nil {
		logger.Error("error occurred while creating binary!", string(output))

		return err
	}

	logger.Info("Binary created successfully")

	return nil
}

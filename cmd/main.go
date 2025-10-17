package main

import (
	"os"

	"github.com/spring-financial-group/jx3-openapi-generation/cmd/app"
	"github.com/spring-financial-group/mqube-go-common/pkg/logger"
)

func main() {
	logger.InitCLILogger()

	if err := app.Run(nil); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

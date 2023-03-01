package main

import (
	"os"

	"spring-financial-group/jx3-openapi-generation/cmd/app"
)

func main() {
	if err := app.Run(nil); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

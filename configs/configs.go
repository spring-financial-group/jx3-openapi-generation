// Package configs embeds the openapitools config files into the binary
package configs

import "embed"

//go:embed *.json
var FS embed.FS

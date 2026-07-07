// Package templates embeds the packaging template files into the binary
package templates

import "embed"

//go:embed angular csharp java javascript typescript
var FS embed.FS

// Package templates embeds the packaging template files into the binary
package templates

import "embed"

//go:embed all:angular all:csharp all:java all:javascript all:typescript
var FS embed.FS

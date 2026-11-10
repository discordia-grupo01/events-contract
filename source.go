package eventscontract

import "embed"

//go:embed */events.go envelope/meta.go
var Source embed.FS

//go:build !dev

package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var distFS embed.FS

var DistFS fs.FS = distFS

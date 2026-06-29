package ui

import (
	"embed"
	"io/fs"
)

//go:embed views
var Views embed.FS

//go:embed static
var Static embed.FS

func ViewsFS() fs.FS {
	sub, err := fs.Sub(Views, "views")
	if err != nil {
		panic(err)
	}
	return sub
}

func StaticFS() fs.FS {
	sub, err := fs.Sub(Static, "static")
	if err != nil {
		panic(err)
	}
	return sub
}

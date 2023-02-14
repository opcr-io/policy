package templates

import "embed"

//go:embed github/* all:policy-template/* gitlab/*
var staticAssets embed.FS

// Static embedded FS service openapi.json file.
func Assets() embed.FS {
	return staticAssets
}

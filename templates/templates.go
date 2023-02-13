package templates

import "embed"

//go:embed github/* policy-template/* gitlab/*
var staticAssets embed.FS

// Static embedded FS service openapi.json file.
func Assets() embed.FS {
	return staticAssets
}

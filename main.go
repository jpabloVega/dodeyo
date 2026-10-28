package main

import (
	"embed"

	"github.com/jpabloVega/dodeyo/cmd"
)

var (

	//go:embed assets/*.ttf assets/*.json
	rootAssetsFS embed.FS
)

func main() {
	cmd.AssetsFS = rootAssetsFS

	cmd.Execute()
}

package poseidon

import (
	_ "embed"
	"encoding/json"
	"html/template"
	"io/fs"
	"net/http"
)

//go:embed templates/index.go.html
var indexTemplateBytes []byte

type SSRAssets struct {
	JS  []string
	CSS []string
}

type SSRProvider interface {
	HandleRequest(r *http.Request) SSRPayload
}

type SSRPayload struct {
	Lang          string
	Title         string
	Description   string
	Assets        SSRAssets
	HeaderContent template.HTML
}

type viteManifest map[string]viteManifestEntry

type viteManifestEntry struct {
	File string   `json:"file"`
	CSS  []string `json:"css"`
}

func SSRAssetsFromViteManifest(file fs.File, entryPoint string) SSRAssets {
	payload := viteManifest{}

	if err := json.NewDecoder(file).Decode(&payload); err != nil {
		panic(err)
	}

	return SSRAssets{
		JS: []string{
			payload[entryPoint].File,
		},
		CSS: payload[entryPoint].CSS,
	}
}

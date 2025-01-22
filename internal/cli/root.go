package cli

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/lunagic/poseidon/poseidon"
	"github.com/lunagic/environment-go/environment"
	"github.com/spf13/cobra"
)

type Config struct {
	Host         string `env:"HOST"`
	Port         int    `env:"PORT"`
	Root         string `env:"POSEIDON_ROOT"`
	SPAMode      bool   `env:"POSEIDON_SPA_MODE"`
	Index        string `env:"POSEIDON_INDEX"`
	NotFoundFile string `env:"POSEIDON_NOT_FOUND_FILE"`
	CachePolicy  bool   `env:"POSEIDON_CACHE_POLICY"`
	GZIP         bool   `env:"POSEIDON_GZIP"`
}

func (config *Config) ListenAddress() string {
	return fmt.Sprintf("%s:%d", config.Host, config.Port)
}

func (config *Config) AddFlags(cmd *cobra.Command) {
	cmd.Flags().IntVarP(
		&config.Port,
		"port",
		"p",
		config.Port,
		"the port to run on",
	)

	cmd.Flags().StringVar(
		&config.Host,
		"host",
		config.Host,
		"the host to run on",
	)

	cmd.Flags().StringVarP(
		&config.Root,
		"root",
		"r",
		config.Root,
		"the root directory to serve",
	)

	cmd.Flags().BoolVar(
		&config.SPAMode,
		"spa",
		config.SPAMode,
		"enables SPA mode (serve index on 404)",
	)

	cmd.Flags().BoolVar(
		&config.GZIP,
		"gzip",
		config.GZIP,
		"enables gzip compression",
	)

	cmd.Flags().BoolVar(
		&config.CachePolicy,
		"cache-policy",
		config.CachePolicy,
		"enables cache policy headers",
	)

	cmd.Flags().StringVar(
		&config.NotFoundFile,
		"not-found-file",
		config.NotFoundFile,
		"file to serve with the 404 status code",
	)

	cmd.Flags().StringVar(
		&config.Index,
		"index",
		config.Index,
		"the index file to use",
	)
}

func Cmd() *cobra.Command {
	// Default configs
	config := &Config{
		Host:         "127.0.0.1",
		Port:         3000,
		NotFoundFile: "404.html",
		Index:        "index.html",
		Root:         ".",
		GZIP:         true,
		CachePolicy:  true,
	}
	if err := environment.New().Decode(config); err != nil {
		panic(err)
	}

	root := &cobra.Command{
		Use: "poseidon",
		RunE: func(cmd *cobra.Command, args []string) error {
			fileSystem := os.DirFS(config.Root)

			configFuncs := []poseidon.ConfigFunc{}

			if config.CachePolicy {
				configFuncs = append(configFuncs, poseidon.WithCachePolicy())
			}

			if config.Index != "" {
				configFuncs = append(configFuncs, poseidon.WithCustomIndex(config.Index))
			}

			if config.NotFoundFile != "" {
				configFuncs = append(configFuncs, poseidon.WithCustomNotFoundFile(config.NotFoundFile))
			}

			if config.SPAMode {
				configFuncs = append(configFuncs, poseidon.WithSPA())
			}

			if config.GZIP {
				configFuncs = append(configFuncs, poseidon.WithGZipCompression())
			}

			service, err := poseidon.New(fileSystem, configFuncs...)
			if err != nil {
				return err
			}

			server := &http.Server{
				Addr: config.ListenAddress(),
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					log.Printf("Request: %s", r.URL.Path)
					service.ServeHTTP(w, r)
				}),
			}

			log.Printf("Listing on http://%s", server.Addr)
			return server.ListenAndServe()
		},
	}

	config.AddFlags(root)

	return root
}

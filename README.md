# poseidon

Static File Server

## Installation

```shell
go install github.com/aaronellington/poseidon@latest
```

## Command Line Usage

```shell
poseidon --help
```

## Configuration

| Flag               | Default        | Env Var                   | Description                                          |
| ------------------ | -------------- | ------------------------- | ---------------------------------------------------- |
| `--host`           | `"127.0.0.1"`  | `HOST`                    | the host to run on                                   |
| `--port`           | `3000`         | `PORT`                    | the port to run on                                   |
| `--root`           | `"."`          | `POSEIDON_ROOT`           | the root directory to serve files from               |
| `--index`          | `"index.html"` | `POSEIDON_INDEX`          | the default file to be served in a directory         |
| `--not-found-file` | `"404.html"`   | `POSEIDON_NOT_FOUND_FILE` | the file that gets served in a "not found" situation |
| `--cache-policy`   | `true`         | `POSEIDON_CACHE_POLICY`   | enables caching headers to be set                    |
| `--gzip`           | `true`         | `POSEIDON_GZIP`           | enables gzip compression                             |
| `--spa`            | `false`        | `POSEIDON_SPA_MODE`       | serves the index in a "not found" situation          |

## Command Line Example

```shell
poseidon --not-found-file=404/index.html
```

## Docker Example

```Dockerfile
FROM node:20 AS builder
WORKDIR /workspace
COPY . .
RUN npm install
RUN npm run build

FROM ghcr.io/aaronellington/poseidon:latest
ENV POSEIDON_NOT_FOUND_FILE=404/index.html
COPY --from=builder /workspace/out .
```

## Library Example

https://pkg.go.dev/github.com/aaronellington/poseidon/poseidon

```shell
go get github.com/aaronellington/poseidon@latest
```

```go
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/aaronellington/poseidon/poseidon"
)

func main() {
	service, err := poseidon.New(
		os.DirFS("."),
		poseidon.WithCachePolicy(),
		poseidon.WithCustomNotFoundFile("404/index.html"),
	)
	if err != nil {
		log.Fatal(err)
	}

	server := &http.Server{
		Addr:    "127.0.0.1:3000",
		Handler: service,
	}

	log.Printf("Listing on http://%s", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
```

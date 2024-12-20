package poseidon

import (
	"io/fs"
	"net/http"
	"slices"
	"strings"
)

func New(
	fileSystem fs.FS,
	configFuncs ...ConfigFunc,
) (*Service, error) {
	service := &Service{
		fileSystem:      fileSystem,
		index:           "index.html",
		notFoundHandler: http.NotFoundHandler(),
		middlewares:     []Middleware{},
	}

	for _, configFunc := range configFuncs {
		if err := configFunc(service); err != nil {
			return nil, err
		}
	}

	slices.Reverse(service.middlewares)

	return service, nil
}

type Middleware func(next http.Handler) http.Handler

type Service struct {
	fileSystem      fs.FS
	index           string
	notFoundHandler http.Handler
	middlewares     []Middleware
}

func (service *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		service.internalServeHTTP(w, r)
	})

	for _, middleware := range service.middlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

func (service *Service) internalServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	path = strings.TrimPrefix(path, "/")
	if strings.HasSuffix(path, "/") || path == "" {
		path += service.index
	}

	file, err := service.fileSystem.Open(path)
	if err != nil {
		doNotCache(w)
		service.notFoundHandler.ServeHTTP(w, r)
		return
	}

	// If directory, add trailing slash and retry
	if info, err := file.Stat(); err != nil {
		panic(err)
	} else if info.IsDir() {
		r.URL.Path += "/"
		service.ServeHTTP(w, r)
		return
	}

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		writeFile(w, file)
	})

	handler.ServeHTTP(w, r)
}

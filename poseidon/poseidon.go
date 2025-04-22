package poseidon

import (
	"io/fs"
	"net/http"
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
		middlewares:     Middlewares{},
	}

	for _, configFunc := range configFuncs {
		if err := configFunc(service); err != nil {
			return nil, err
		}
	}

	service.handler = service.middlewares.Apply(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		service.internalServeHTTP(w, r)
	}))

	return service, nil
}

type Service struct {
	fileSystem      fs.FS
	index           string
	notFoundHandler http.Handler
	middlewares     Middlewares
	handler         http.Handler
}

func (service *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	service.handler.ServeHTTP(w, r)
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
	} else if info.IsDir() && !strings.HasSuffix(r.URL.Path, "/") {
		doNotCache(w)
		http.Redirect(w, r, r.URL.Path+"/", http.StatusTemporaryRedirect)
		return
	}

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeFile(w, file, http.StatusOK)
	})

	handler.ServeHTTP(w, r)
}

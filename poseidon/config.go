package poseidon

import (
	"io"
	"net/http"
	"strings"
)

type ConfigFunc func(service *Service) error

func WithMiddleware(middleware Middleware) ConfigFunc {
	return func(service *Service) error {
		service.middlewares = append(service.middlewares, middleware)

		return nil
	}
}

func WithCachePolicy() ConfigFunc {
	checkers := []func(path string) bool{
		// Next.js
		func(path string) bool {
			return strings.HasPrefix(path, "/_next/")
		},
	}

	return WithMiddleware(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, checker := range checkers {
				if checker(r.URL.Path) {
					w.Header().Set("Cache-Control", "public,max-age=31536000,immutable")
					next.ServeHTTP(w, r)
					return
				}
			}

			// Don't cache anything by default
			w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
			next.ServeHTTP(w, r)
		})
	})
}

func WithCustomIndex(index string) ConfigFunc {
	return func(service *Service) error {
		service.index = index

		return nil
	}
}

func WithCustomNotFoundHandler(handler http.Handler) ConfigFunc {
	return func(service *Service) error {
		service.notFoundHandler = handler

		return nil
	}
}

func WithCustomNotFoundFile(filePath string) ConfigFunc {
	return ConfigFunc(func(service *Service) error {
		return WithCustomNotFoundHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Open the custom not found file
			file, err := service.fileSystem.Open(filePath)
			if err != nil {
				http.NotFound(w, r)
				return
			}

			w.WriteHeader(http.StatusNotFound)
			io.Copy(w, file)
			file.Close()
		}))(service)
	})
}

func WithSPA() ConfigFunc {
	return func(service *Service) error {
		service.notFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Prevent infinite recursion
			if r.URL.Path == service.index {
				http.NotFoundHandler().ServeHTTP(w, r)
				return
			}

			// Change the path to the index and retry
			r.URL.Path = service.index
			service.ServeHTTP(w, r)
		})

		return nil
	}
}

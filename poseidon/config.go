package poseidon

import (
	"compress/gzip"
	"html/template"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
)

type ConfigFunc func(service *Service) error

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func WithGZipCompression() ConfigFunc {
	return WithMiddleware(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip if browser does not support gzip
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("Content-Encoding", "gzip")
			gzipWriter := gzip.NewWriter(w)
			defer func() {
				_ = gzipWriter.Close()
			}()
			gzipResponseWriter := gzipResponseWriter{Writer: gzipWriter, ResponseWriter: w}

			next.ServeHTTP(gzipResponseWriter, r)
		})
	})
}

func WithMiddleware(middleware Middleware) ConfigFunc {
	return func(service *Service) error {
		service.middlewares = append(service.middlewares, middleware)

		return nil
	}
}

func WithCachePolicy(checkers ...func(string) bool) ConfigFunc {
	return WithMiddleware(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, checker := range checkers {
				if checker(r.URL.Path) {
					cacheForever(w)
					next.ServeHTTP(w, r)
					return
				}
			}

			// Don't cache anything by default
			doNotCache(w)
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
		service.notFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.Header.Get("accept"), "text/html") {
				http.NotFound(w, r)
				return
			}

			handler.ServeHTTP(w, r)
		})

		return nil
	}
}

func WithCustomNotFoundFile(filePath string) ConfigFunc {
	return ConfigFunc(func(service *Service) error {
		return WithCustomNotFoundHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Open the custom not found file
			file, err := service.fileSystem.Open(filePath)
			if err != nil {
				doNotCache(w)
				http.NotFound(w, r)
				return
			}
			defer func() {
				_ = file.Close()
			}()

			doNotCache(w)
			writeFile(w, file, http.StatusNotFound)
		}))(service)
	})
}

func WithClientSideRouting() ConfigFunc {
	return func(service *Service) error {
		return WithCustomNotFoundHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Prevent infinite recursion
			if r.URL.Path == service.index {
				http.NotFound(w, r)
				return
			}

			// Change the path to the index and retry
			r.URL.Path = service.index
			service.ServeHTTP(w, r)
		}))(service)
	}
}

func WithClientSideRoutingAndServerSideRendering(provider SSRProvider) ConfigFunc {
	return func(service *Service) error {
		indexTemplate, err := template.New("index").Parse(string(indexTemplateBytes))
		if err != nil {
			return err
		}

		return WithCustomNotFoundHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := indexTemplate.Execute(w, provider.HandleRequest(r)); err != nil {
				_, _ = w.Write([]byte(err.Error()))
			}
		}))(service)
	}
}

func doNotCache(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}

func cacheForever(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "public,max-age=31536000,immutable")
	w.Header().Del("Pragma")
	w.Header().Del("Expires")
}

func writeFile(w http.ResponseWriter, file fs.File, status int) {
	stat, _ := file.Stat()
	w.Header().Set("content-type", mime.TypeByExtension(filepath.Ext(stat.Name())))
	w.WriteHeader(status)
	_, _ = io.Copy(w, file)
	_ = file.Close()
}

package poseidon

import (
	"net/http"
	"slices"
)

type Middlewares []Middleware

func (middlewares Middlewares) Apply(next http.Handler) http.Handler {
	slices.Reverse(middlewares)

	for _, middleware := range middlewares {
		next = middleware(next)
	}

	return next
}

type Middleware func(next http.Handler) http.Handler

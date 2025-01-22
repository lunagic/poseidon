package poseidon_test

import (
	"log"
	"net/http"
	"os"

	"github.com/lunagic/poseidon/poseidon"
)

func ExampleNew() {
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

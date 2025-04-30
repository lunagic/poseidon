package poseidon

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
)

func RespondJSON(w http.ResponseWriter, status int, payload any) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(payloadBytes)
}

func RespondXML(w http.ResponseWriter, status int, payload any) {
	payloadBytes, err := xml.MarshalIndent(payload, "", "    ")
	if err != nil {
		panic(err)
	}
	w.Header().Set("content-type", "application/xml")
	w.WriteHeader(status)
	_, _ = w.Write([]byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n"))
	_, _ = w.Write(payloadBytes)
}

package api_server

import (
	"encoding/json"
	"log"
	"net/http"
)

func RespondWithJSON(request *http.Request, writer http.ResponseWriter, code int, payload interface{}) {
	log.Printf("%s [%d]\n", request.URL.Path, code)
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(code)
	json.NewEncoder(writer).Encode(payload)
}
func RespondWithError(request *http.Request, writer http.ResponseWriter, code int, message string) {
	RespondWithJSON(request, writer, code, map[string]string{"error": message})
}

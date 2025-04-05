package utils

import (
	"log"
	"net/http"
)

// WriteErrorWithCannotWriteResponse Write error message to the log and sets the status code to 500,
// if w.Write return err.
func WriteErrorWithCannotWriteResponse(w http.ResponseWriter, err error) {
	log.Printf("cannot write response: %v", err)
	w.WriteHeader(http.StatusInternalServerError)
}

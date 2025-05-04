package utils

import (
	"go.uber.org/zap"
	"net/http"
)

// WriteErrorWithCannotWriteResponse Write error message to the log and sets the status code to 500,
// if w.Write return err.
func WriteErrorWithCannotWriteResponse(w http.ResponseWriter, err error) {
	zap.L().Error("cannot write response: %v", zap.Error(err))
	w.WriteHeader(http.StatusInternalServerError)
}

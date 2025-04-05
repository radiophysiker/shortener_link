package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi"

	"github.com/radiophysiker/shortener_link/internal/usecases"
	"github.com/radiophysiker/shortener_link/internal/utils"
)

func (h *URLHandler) GetFullURL(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "id")
	fullURL, err := h.URLUseCase.GetFullURL(shortURL)
	if err != nil {
		if errors.Is(err, usecases.ErrEmptyShortURL) {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte("short url is empty"))
			if err != nil {
				utils.WriteErrorWithCannotWriteResponse(w, err)
			}
			return
		}
		if errors.Is(err, usecases.ErrURLNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_, err := w.Write([]byte("url is not found for " + shortURL))
			if err != nil {
				utils.WriteErrorWithCannotWriteResponse(w, err)
			}
			return
		}
		log.Printf("cannot get full URL: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Location", fullURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

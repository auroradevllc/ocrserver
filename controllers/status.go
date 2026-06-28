package controllers

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/otiai10/gosseract/v2"
)

const version = "0.3.0"

type tesseractStatus struct {
	Version   string   `json:"version"`
	Languages []string `json:"languages"`
}

type statusResponse struct {
	Success   bool            `json:"success"`
	Version   string          `json:"version"`
	Tesseract tesseractStatus `json:"tesseract"`
}

// Status ...
func Status(w http.ResponseWriter, r *http.Request) {
	langs, err := gosseract.GetAvailableLanguages()

	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, errorResponse{
			Error: err.Error(),
		})
		return
	}

	render.JSON(w, r, statusResponse{
		Success: true,
		Version: version,
		Tesseract: tesseractStatus{
			Version:   clientVersion,
			Languages: langs,
		},
	})
}

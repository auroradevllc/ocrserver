package controllers

import (
	"net/http"

	"github.com/go-chi/render"
)

const version = "0.5.0"

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
	render.JSON(w, r, statusResponse{
		Success: true,
		Version: version,
		Tesseract: tesseractStatus{
			Version:   clientVersion,
			Languages: langs,
		},
	})
}

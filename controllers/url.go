package controllers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/render"
	"github.com/otiai10/gosseract/v2"
)

type urlBody struct {
	URL       string `json:"url" validate:"required,url"`
	Trim      string `json:"trim"`
	Languages string `json:"languages"`
	Whitelist string `json:"whitelist"`
}

// URL ...
func URL(w http.ResponseWriter, r *http.Request) {
	var body urlBody

	err := json.NewDecoder(r.Body).Decode(&body)

	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, errorResponse{
			Error: err.Error(),
		})
		return
	}

	if err := v.Struct(body); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, errorResponse{
			Error: err.Error(),
		})
		return
	}

	tempfile, err := os.CreateTemp("", "ocrserver"+"-")

	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, &errorResponse{Error: err.Error()})
		return
	}

	defer func() {
		_ = tempfile.Close()
		_ = os.Remove(tempfile.Name())
	}()

	// Get url to local file
	res, err := http.Get(body.URL)

	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, &errorResponse{Error: err.Error()})
		return
	}

	defer res.Body.Close()

	s := sha256.New()

	mw := io.MultiWriter(tempfile, s)

	_, err = io.Copy(mw, res.Body)

	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, &errorResponse{Error: err.Error()})
		return
	}

	h := s.Sum(nil)

	w.Header().Set("X-File-Hash", hex.EncodeToString(h))

	client := gosseract.NewClient()
	defer client.Close()

	client.Languages = []string{"eng"}

	if body.Languages != "" {
		client.Languages = strings.Split(body.Languages, ",")
	}

	if err := client.SetImage(tempfile.Name()); err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, errorResponse{
			Error: fmt.Sprintf("failed to set tesseract image: %v", err),
		})

		return
	}

	if body.Whitelist != "" {
		client.SetWhitelist(body.Whitelist)
	}

	text, err := client.Text()

	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, err)
		return
	}

	render.JSON(w, r, ocrResponse{
		Success: true,
		Result:  strings.Trim(text, body.Trim),
		Version: version,
	})
}

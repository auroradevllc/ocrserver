package controllers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/anthonynsimon/bild/effect"
	"github.com/go-chi/render"
	"github.com/otiai10/gosseract/v2"
)

type urlBody struct {
	URL              string `json:"url" validate:"required,url"`
	Trim             string `json:"trim"`
	Languages        string `json:"languages"`
	Whitelist        string `json:"whitelist"`
	ConvertGrayscale bool   `json:"convertGrayscale"`
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

	if body.ConvertGrayscale {
		// Read the first 512 bytes to use for the content type, file extensions lie
		h := make([]byte, 512)

		if _, err := tempfile.ReadAt(h, 0); err != nil {
			serveError(w, r, err)
			return
		}

		t := http.DetectContentType(h)

		var img image.Image

		// Reset file position to read
		if _, err := tempfile.Seek(0, io.SeekStart); err != nil {
			serveError(w, r, err)
			return
		}

		switch t {
		case "image/png":
			img, err = png.Decode(tempfile)
		case "image/jpeg", "image/jpg":
			img, err = jpeg.Decode(tempfile)
		}

		result := effect.GrayscaleWithWeights(img, 0.2126, 0.7152, 0.0722)

		if err := tempfile.Truncate(0); err != nil {
			serveError(w, r, err)
			return
		}

		// Re-encode the image
		switch t {
		case "image/png":
			err = png.Encode(tempfile, result)
		case "image/jpeg", "image/jpg":
			err = jpeg.Encode(tempfile, result, nil)
		}

		if err != nil {
			serveError(w, r, err)
			return
		}
	}

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
		serveError(w, r, err)
		return
	}

	render.JSON(w, r, ocrResponse{
		Success: true,
		Result:  strings.Trim(text, body.Trim),
		Version: version,
	})
}

func serveError(w http.ResponseWriter, r *http.Request, err error) {
	render.Status(r, http.StatusInternalServerError)
	render.JSON(w, r, err)
}

package controllers

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"net/http"
	"regexp"
	"strings"

	"github.com/anthonynsimon/bild/effect"
	"github.com/go-chi/render"
	"github.com/otiai10/gosseract/v2"
)

type base64Body struct {
	Base64           string `json:"base64" validate:"required"`
	Trim             string `json:"trim"`
	Languages        string `json:"languages"`
	Whitelist        string `json:"whitelist"`
	ConvertGrayscale bool   `json:"convertGrayscale"`
}

// Base64 ...
func Base64(w http.ResponseWriter, r *http.Request) {
	var body base64Body

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

	body.Base64 = regexp.MustCompile("data:image\\/png;base64,").ReplaceAllString(body.Base64, "")

	b, err := base64.StdEncoding.DecodeString(body.Base64)

	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, errorResponse{
			Error: err.Error(),
		})
		return
	}

	s := sha256.New()
	s.Write(b)
	h := s.Sum(nil)

	w.Header().Set("X-File-Hash", hex.EncodeToString(h))

	if body.ConvertGrayscale {
		t := http.DetectContentType(b)

		var img image.Image

		switch t {
		case "image/png":
			img, err = png.Decode(bytes.NewReader(b))
		case "image/jpeg", "image/jpg":
			img, err = jpeg.Decode(bytes.NewReader(b))
		}

		result := effect.GrayscaleWithWeights(img, 0.2126, 0.7152, 0.0722)

		var buf bytes.Buffer

		switch t {
		case "image/png":
			err = png.Encode(&buf, result)
		case "image/jpeg", "image/jpg":
			err = jpeg.Encode(&buf, result, nil)
		}

		b = buf.Bytes()
	}

	client := gosseract.NewClient()
	defer client.Close()

	client.Languages = []string{"eng"}

	if body.Languages != "" {
		client.Languages = strings.Split(body.Languages, ",")
	}

	if err := client.SetImageFromBytes(b); err != nil {
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

package controllers

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-chi/render"
	"github.com/otiai10/gosseract/v2"
	"gocv.io/x/gocv"
)

type base64Body struct {
	Base64           string                 `json:"base64" validate:"required"`
	Trim             string                 `json:"trim"`
	Languages        string                 `json:"languages"`
	Whitelist        string                 `json:"whitelist"`
	ConvertGrayscale bool                   `json:"convertGrayscale"`
	Deskew           bool                   `json:"deskew"`
	PageSegMode      *gosseract.PageSegMode `json:"pageSegMode"`
}

// Base64 ...
func Base64(w http.ResponseWriter, r *http.Request) {
	var body base64Body

	err := json.NewDecoder(r.Body).Decode(&body)

	if err != nil {
		serveError(w, r, err)
		return
	}

	if err := v.Struct(body); err != nil {
		serveError(w, r, err)
		return
	}

	body.Base64 = regexp.MustCompile("data:image\\/png;base64,").ReplaceAllString(body.Base64, "")

	b, err := base64.StdEncoding.DecodeString(body.Base64)

	if err != nil {
		serveError(w, r, err)
		return
	}

	s := sha256.New()
	s.Write(b)
	h := s.Sum(nil)

	w.Header().Set("X-File-Hash", hex.EncodeToString(h))

	if body.Deskew || body.ConvertGrayscale {
		img, err := gocv.IMDecode(b, gocv.IMReadColor)
		defer img.Close()

		if err != nil {
			serveError(w, r, err)
			return
		}

		var outputImg gocv.Mat

		if body.Deskew {
			deskewed, err := deskewMaterial(img)
			defer deskewed.Close()

			if err != nil {
				serveError(w, r, err)
				return
			}

			outputImg = *deskewed
		} else if body.ConvertGrayscale {
			outputImg = gocv.NewMat()
			defer outputImg.Close()

			if err := gocv.CvtColor(img, &outputImg, gocv.ColorBGRToGray); err != nil {
				serveError(w, r, err)
				return
			}
		} else {
			serveError(w, r, fmt.Errorf("invalid image operation"))
			return
		}

		t := http.DetectContentType(b)

		var ext gocv.FileExt

		switch t {
		case "image/png":
			ext = gocv.PNGFileExt
		case "image/jpg", "image/jpeg":
			ext = gocv.JPEGFileExt
		}

		buf, err := gocv.IMEncode(ext, outputImg)
		defer buf.Close()

		if err != nil {
			serveError(w, r, err)
			return
		}

		b = buf.GetBytes()
	}

	client := gosseract.NewClient()
	defer client.Close()

	client.Languages = []string{"eng"}

	if body.Languages != "" {
		client.Languages = strings.Split(body.Languages, ",")
	}

	if err := client.SetImageFromBytes(b); err != nil {
		serveError(w, r, fmt.Errorf("failed to set tesseract image: %v", err))
		return
	}

	if body.Whitelist != "" {
		client.SetWhitelist(body.Whitelist)
	}

	if body.PageSegMode != nil {
		client.SetPageSegMode(*body.PageSegMode)
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

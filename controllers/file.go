package controllers

import (
	"crypto/sha256"
	"encoding/hex"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/anthonynsimon/bild/effect"
	"github.com/go-chi/render"
	"github.com/otiai10/gosseract/v2"
)

var (
	imgexp = regexp.MustCompile("^image")
)

// FileUpload ...
func FileUpload(w http.ResponseWriter, r *http.Request) {
	// Get uploaded file
	r.ParseMultipartForm(32 << 20)

	// upload, h, err := r.FormFile("file")
	upload, _, err := r.FormFile("file")

	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, &errorResponse{Error: err.Error()})
		return
	}

	defer upload.Close()

	// Create physical file
	tempfile, err := os.CreateTemp("", "ocrserver"+"-")

	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, &errorResponse{Error: err.Error()})
		return
	}

	defer func() {
		tempfile.Close()
		_ = os.Remove(tempfile.Name())
	}()

	s := sha256.New()

	mw := io.MultiWriter(tempfile, s)

	// Make uploaded physical
	if _, err = io.Copy(mw, upload); err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, &errorResponse{Error: err.Error()})
		return
	}

	h := s.Sum(nil)

	w.Header().Set("X-File-Hash", hex.EncodeToString(h))

	if val := r.FormValue("convertGrayscale"); val == "true" {
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

	if err := client.SetImage(tempfile.Name()); err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, &errorResponse{Error: err.Error()})
		return
	}

	client.Languages = []string{"eng"}

	if langs := r.FormValue("languages"); langs != "" {
		client.Languages = strings.Split(langs, ",")
	}

	if whitelist := r.FormValue("whitelist"); whitelist != "" {
		client.SetWhitelist(whitelist)
	}

	var out string
	switch r.FormValue("format") {
	case "hocr":
		out, err = client.HOCRText()
	default:
		out, err = client.Text()
	}
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, &errorResponse{Error: err.Error()})
		return
	}

	render.JSON(w, r, ocrResponse{
		Success: true,
		Result:  strings.Trim(out, r.FormValue("trim")),
		Version: version,
	})
}

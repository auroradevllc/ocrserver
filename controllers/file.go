package controllers

import (
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

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
		os.Remove(tempfile.Name())
	}()

	// Make uploaded physical
	if _, err = io.Copy(tempfile, upload); err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, &errorResponse{Error: err.Error()})
		return
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

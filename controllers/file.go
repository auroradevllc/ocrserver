package controllers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
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
	upload, head, err := r.FormFile("file")

	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, &errorResponse{Error: err.Error()})
		return
	}

	defer upload.Close()

	ext := strings.ToLower(path.Ext(head.Filename))

	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		serveError(w, r, fmt.Errorf("%s is not a valid image", head.Filename))
		return
	}

	// Create physical file
	tempfile, err := os.CreateTemp("", "ocrserver-*"+ext)

	if err != nil {
		serveError(w, r, err)
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
		serveError(w, r, err)
		return
	}

	h := s.Sum(nil)

	w.Header().Set("X-File-Hash", hex.EncodeToString(h))

	if val := r.FormValue("deskew"); val == "true" {
		if err := deskewImageFile(tempfile); err != nil {
			serveError(w, r, err)
			return
		}
	} else if val := r.FormValue("convertGrayscale"); val == "true" {
		// Note: This won't do anything with the above so we do an else if
		if err := grayscaleImageFile(tempfile); err != nil {
			serveError(w, r, err)
			return
		}
	}

	client := gosseract.NewClient()
	defer client.Close()

	if err := client.SetImage(tempfile.Name()); err != nil {
		serveError(w, r, err)
		return
	}

	client.Languages = []string{"eng"}

	if langs := r.FormValue("languages"); langs != "" {
		client.Languages = strings.Split(langs, ",")
	}

	if whitelist := r.FormValue("whitelist"); whitelist != "" {
		client.SetWhitelist(whitelist)
	}

	if pageSegMode := r.FormValue("pageSegMode"); pageSegMode != "" {
		pageSegModeInt, err := strconv.Atoi(pageSegMode)

		if err != nil {
			serveError(w, r, err)
			return
		}

		client.SetPageSegMode(gosseract.PageSegMode(pageSegModeInt))
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

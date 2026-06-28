package controllers

import (
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"

	"github.com/anthonynsimon/bild/effect"
)

func grayscaleImageFile(file *os.File) error {
	h := make([]byte, 512)

	if _, err := file.ReadAt(h, 0); err != nil {
		return err
	}

	t := http.DetectContentType(h)

	var img image.Image

	// Reset file position to read
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return err
	}

	var err error

	switch t {
	case "image/png":
		img, err = png.Decode(file)
	case "image/jpeg", "image/jpg":
		img, err = jpeg.Decode(file)
	}

	if err != nil {
		return err
	}

	if img != nil {
		result := effect.GrayscaleWithWeights(img, 0.2126, 0.7152, 0.0722)

		// Seek to position 0
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			return err
		}

		// Truncate file
		if err := file.Truncate(0); err != nil {
			return err
		}

		// Re-encode the image
		switch t {
		case "image/png":
			err = png.Encode(file, result)
		case "image/jpeg", "image/jpg":
			err = jpeg.Encode(file, result, nil)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

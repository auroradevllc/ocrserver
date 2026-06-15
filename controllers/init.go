package controllers

import (
	"github.com/go-playground/validator/v10"
	"github.com/otiai10/gosseract/v2"
)

var (
	v             = validator.New()
	clientVersion string
)

func init() {
	client := gosseract.NewClient()
	defer client.Close()

	clientVersion = client.Version()
}

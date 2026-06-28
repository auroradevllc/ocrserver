package controllers

import (
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/otiai10/gosseract/v2"
)

var (
	v             = validator.New()
	clientVersion string
	langs         []string
)

func init() {
	client := gosseract.NewClient()
	defer client.Close()

	clientVersion = client.Version()

	var err error

	langs, err = gosseract.GetAvailableLanguages()

	if err != nil {
		log.Fatalln("Unable to get available languages:", err)
	}
}

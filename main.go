package main

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/auroradevllc/ocrserver/controllers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

//go:embed app/assets/*
var assetFs embed.FS

//go:embed app/views/index.html
var indexTpl []byte

func main() {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)

	// API
	r.Get("/status", controllers.Status)
	r.Post("/base64", controllers.Base64)
	r.Post("/file", controllers.FileUpload)
	r.Post("/url", controllers.URL)

	// Sample Page
	h := controllers.NewIndexHandler(
		template.Must(template.New("index").Parse(string(indexTpl))),
	)

	r.Handle("/", h)

	// Assets for sample page
	fs, err := fs.Sub(assetFs, "app/assets")

	if err != nil {
		log.Fatal(err)
	}

	r.Handle("/assets/*", http.StripPrefix("/assets/",
		http.FileServer(http.FS(fs)),
	))

	port := os.Getenv("PORT")

	if port == "" {
		log.Fatalln("Required env `PORT` is not specified.")
	}

	log.Printf("listening on port %s", port)

	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Println(err)
	}
}

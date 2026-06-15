package controllers

import (
	"html/template"
	"net/http"
)

type IndexHandler struct {
	tpl *template.Template
}

func NewIndexHandler(tpl *template.Template) *IndexHandler {
	return &IndexHandler{
		tpl: tpl,
	}
}

// Index ...
func (h *IndexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	_ = h.tpl.Execute(w, map[string]any{
		"AppName": "ocrserver",
	})
}

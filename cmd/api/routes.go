package main

import (
	"net/http"

	"itposao/internal/app"
)

func routes(application *app.Application) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", application.HandleHealth)
	return mux
}

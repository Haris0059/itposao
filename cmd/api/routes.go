package main

import (
	"net/http"

	"itposao/internal/app"
)

func routes(application *app.Application) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /v1/health", application.HandleHealth)
	mux.HandleFunc("GET /v1/jobs", application.HandleJobs)
	mux.HandleFunc("GET /v1/companies", application.HandleCompanies)
	mux.HandleFunc("GET /v1/countries", application.HandleCountries)
	mux.Handle("/", http.FileServer(http.Dir("web")))

	return mux
}

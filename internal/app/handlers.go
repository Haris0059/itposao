package app

import (
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (app *Application) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

func (app *Application) HandleJobs(w http.ResponseWriter, r *http.Request) {
	if app.Jobs == nil {
		http.Error(w, "database not configured", http.StatusServiceUnavailable)
		return
	}

	jobs, err := app.Jobs.Latest(r.Context(), 50)
	if err != nil {
		app.Logger.Error("list jobs failed", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, jobs)
}

func (app *Application) HandleCompanies(w http.ResponseWriter, r *http.Request) {
	if app.Companies == nil {
		http.Error(w, "database not configured", http.StatusServiceUnavailable)
		return
	}

	companies, err := app.Companies.List(r.Context())
	if err != nil {
		app.Logger.Error("list companies failed", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, companies)
}

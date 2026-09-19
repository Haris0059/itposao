package app

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"itposao/internal/models"
)

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func (app *Application) HandleHealth(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("ok"))
}

func (app *Application) HandleJobs(w http.ResponseWriter, r *http.Request) {
	if app.Jobs == nil {
		http.Error(w, "database not configured", http.StatusServiceUnavailable)
		return
	}

	jobs, err := app.Jobs.Latest(r.Context(), 500)
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

func (app *Application) HandleCountries(w http.ResponseWriter, r *http.Request) {
	if app.Countries == nil {
		http.Error(w, "database not configured", http.StatusServiceUnavailable)
		return
	}

	countries, err := app.Countries.List(r.Context())
	if err != nil {
		app.Logger.Error("list countries failed", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, countries)
}

func (app *Application) HandleJob(w http.ResponseWriter, r *http.Request) {
	if app.Jobs == nil {
		http.Error(w, "database not configured", http.StatusServiceUnavailable)
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	job, err := app.Jobs.Get(r.Context(), id)
	if errors.Is(err, models.ErrNoRecord) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		app.Logger.Error("get job failed", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, job)
}

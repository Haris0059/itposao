// Package app holds the Application, the DI container for shared
// dependencies (logger, config, etc.) that HTTP handlers need.
package app

import (
	"log/slog"

	"itposao/internal/models"
)

type Application struct {
	Logger *slog.Logger
	Jobs   *models.JobModel
}

func New(logger *slog.Logger, jobs *models.JobModel) *Application {
	return &Application{Logger: logger, Jobs: jobs}
}

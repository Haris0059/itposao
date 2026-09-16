package app

import "net/http"

func (app *Application) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

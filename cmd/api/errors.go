package main

import (
	"net/http"
)

func (app *application) internalServerResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("internal server error", "method", r.Method, "path", r.URL.Path)

	writeJSONError(w, http.StatusInternalServerError, err.Error())
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnf("bad request error", "method", r.Method, "path", r.URL.Path, "error", err)

	writeJSONError(w, http.StatusBadRequest, err.Error())
}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("not found error", "method", r.Method, "path", r.URL.Path)

	writeJSONError(w, http.StatusNotFound, "not found")
}

// func (app *application) conflictResponse(w http.ResponseWriter, r *http.Request, err error) {
// 	app.logger.Errorw("conflict error", "method", r.Method, "path", r.URL.Path, "error", err)

// 	writeJSONError(w, http.StatusConflict, err.Error())
// }

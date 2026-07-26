package main

import (
	"net/http"

	"go.uber.org/zap"
)

func (app *application) badRequestError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("bad request",
		zap.Error(err),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.Int("status", http.StatusBadRequest),
	)
	if writeErr := WriteJSONError(w, http.StatusBadRequest, err.Error()); writeErr != nil {
		app.logger.Errorw("failed to write error response", zap.Error(writeErr))
	}
}
func (app *application) notFoundError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("not found",
		zap.Error(err),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.Int("status", http.StatusNotFound),
	)
	if writeErr := WriteJSONError(w, http.StatusNotFound, "not found"); writeErr != nil {
		app.logger.Errorw("failed to write error response", zap.Error(writeErr))
	}
}

func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("internal server error",
		zap.Error(err),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.Int("status", http.StatusInternalServerError),
	)
	if writeErr := WriteJSONError(w, http.StatusInternalServerError, "the server encountered a problem"); writeErr != nil {
		app.logger.Errorw("failed to write error response", zap.Error(writeErr))
	}
}
func (app *application) conflictError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("conflict",
		zap.Error(err),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.Int("status", http.StatusConflict),
	)
	if writeErr := WriteJSONError(w, http.StatusConflict, "conflict"); writeErr != nil {
		app.logger.Errorw("failed to write error response", zap.Error(writeErr))
	}
}
func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("conflict",
		zap.Error(err),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.Int("status", http.StatusNotFound),
	)
	if writeErr := WriteJSONError(w, http.StatusNotFound, "conflict"); writeErr != nil {
		app.logger.Errorw("failed to write error response", zap.Error(writeErr))
	}
}
func (app *application) unauthorizedErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("unauthorized error",
		zap.Error(err),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.Int("status", http.StatusUnauthorized),
	)
	if writeErr := WriteJSONError(w, http.StatusUnauthorized, "unauthorized"); writeErr != nil {
		app.logger.Errorw("failed to write error response", zap.Error(writeErr))
	}
}
func (app *application) unauthorizedBasicErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("unauthorized basic error",
		zap.Error(err),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.Int("status", http.StatusUnauthorized),
	)
	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
	if writeErr := WriteJSONError(w, http.StatusUnauthorized, "unauthorized"); writeErr != nil {
		app.logger.Errorw("failed to write error response", zap.Error(writeErr))
	}
}
func (app *application) forbiddenResponse(w http.ResponseWriter, r *http.Request) {
	app.logger.Errorw("unauthorized basic error",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.Int("status", http.StatusForbidden),
	)
	if writeErr := WriteJSONError(w, http.StatusForbidden, "forbidden"); writeErr != nil {
		app.logger.Errorw("forbidden", zap.Error(writeErr))
	}
}
func (app *application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request, retryAfter string) {
	app.logger.Errorw("TooManyRequests error",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.Int("status", http.StatusTooManyRequests),
	)
	w.Header().Set("Retry-After", retryAfter)
	if writeErr := WriteJSONError(w, http.StatusTooManyRequests, "TooManyRequests"); writeErr != nil {
		app.logger.Errorw("TooManyRequests", zap.Error(writeErr))
	}
}

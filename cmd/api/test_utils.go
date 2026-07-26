package main

import (
	"net/http"
	"net/http/httptest"
	"social/internal/auth"
	"social/internal/ratelimiter"
	"social/internal/store"
	"social/internal/store/cache"
	"testing"

	"go.uber.org/zap"
)

func newTestApplication(t *testing.T, cfg config) *application {
	t.Helper()

	// logger := zap.Must(zap.NewDevelopment()).Sugar()
	logger := zap.NewNop().Sugar()
	// zapconfig := zap.NewDevelopmentConfig()
	// zapconfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	// logger, _ := zapconfig.Build()
	mockStore := store.NewMockStorage()
	mockCacheStore := cache.NewMockCacheStorage()
	testAuth := &auth.TestAuthenticator{}

	ratelimiter := ratelimiter.NewFixedWindowLimiter(
		cfg.ratelimiter.RequestPerTimeFrame,
		cfg.ratelimiter.TimeFrame,
	)

	return &application{
		logger:        logger,
		store:         mockStore,
		cacheStorage:  mockCacheStore,
		authenticator: testAuth,
		config:        cfg,
		rateLimiter:   ratelimiter,
	}
}
func executeRequest(req *http.Request, mux http.Handler) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	return rr
}
func checkResponseCode(t *testing.T, expect, response int) {
	if expect != response {
		t.Errorf("expected the response code %d. but got %d", expect, response)
	}
}

package main

import (
	"context"
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"social/internal/auth"
	"social/internal/env"
	"social/internal/mailer"
	"social/internal/ratelimiter"
	"social/internal/store"
	"social/internal/store/cache"
	"syscall"
	"time"

	"social/docs"

	"github.com/arl/statsviz"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

type application struct {
	config        config
	store         store.Storage
	cacheStorage  cache.Storage
	logger        *zap.SugaredLogger
	mailer        mailer.Client
	authenticator auth.Authenticator
	rateLimiter   ratelimiter.Limiter
}
type config struct {
	addr        string
	apiURL      string
	db          dbConfig
	env         string
	mail        mailConfig
	frontendURL string
	auth        authConfig
	redisCfg    redisConfig
	ratelimiter ratelimiter.Config
}
type redisConfig struct {
	addr    string
	pw      string
	db      int
	enabled bool
}
type authConfig struct {
	basic basicConfig
	token tokenConfig
}
type tokenConfig struct {
	secret string
	exp    time.Duration
	iss    string
}
type basicConfig struct {
	user string
	pass string
}
type mailConfig struct {
	sendGrid  sendGridConfig
	fromEmail string
	exp       time.Duration
}
type sendGridConfig struct {
	apiKey string
}
type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

func (app *application) mount() http.Handler {
	statsvizSrv, err := statsviz.NewServer(statsviz.Root("/v1/debug/statsviz"))
	if err != nil {
		app.logger.Errorw("failed to create statsviz server", "error", err)
	}
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedOrigins:   []string{env.GetString("CORS_ALLOWED_ORIGIN", "http://*")},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	r.Use(middleware.ClientIPFromXFF())
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(app.RateLimiterMiddleware)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Route("/v1", func(r chi.Router) {

		// r.With(app.BasicAuthMiddleware()).Get("/health", app.healthCheckHandler)
		r.Get("/health", app.healthCheckHandler)
		r.With(app.BasicAuthMiddleware()).Get("/debug/vars", expvar.Handler().ServeHTTP)
		if statsvizSrv != nil {
			r.Get("/debug/statsviz", func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "/v1/debug/statsviz/", http.StatusFound)
			})
			r.Get("/debug/statsviz/*", statsvizSrv.Index())
			r.Get("/debug/statsviz/ws", statsvizSrv.Ws())
		}

		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(fmt.Sprintf("http://%s/v1/swagger/doc.json", app.config.apiURL))))
		r.Route("/posts", func(r chi.Router) {
			r.Use(app.AuthTokenMiddleware)
			r.Post("/", app.createPostHandler)

			r.Route("/{id}", func(r chi.Router) {
				r.Use(app.postsContextMiddleware)
				r.Get("/", app.getPostHandler)
				r.Patch("/", app.CheckPostOwnership("moderator", app.updatePostHandler))
				r.Delete("/", app.CheckPostOwnership("admin", app.deletePostHandler))
			})

		})
		r.Route("/users", func(r chi.Router) {
			r.Put("/activate/{token}", app.activateUserHandler)
			r.Route("/{id}", func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware)

				r.Get("/", app.getUserHandler)
				r.Patch("/follow", app.followUserHandler)
				r.Patch("/unfollow", app.unfollowUserHandler)
				r.Group(func(r chi.Router) {
					r.Use(app.AuthTokenMiddleware)
					r.Get("/feed", app.getUserFeedHandler)
				})
			})
		})
		r.Route("/authentication", func(r chi.Router) {
			r.Post("/user", app.registerUserHandler)
			r.Post("/token", app.createTokenHandler)
		})
	})

	return r
}
func (app *application) run(mux http.Handler) error {
	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.BasePath = "/v1"
	docs.SwaggerInfo.Host = app.config.apiURL

	srv := http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		IdleTimeout:  time.Minute,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
	}
	shutdown := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)

		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		app.logger.Infow("signal caught", "signal", s.String())

		shutdown <- srv.Shutdown(ctx)
	}()

	app.logger.Infow("server start listen on", "addr", srv.Addr, "env", app.config.env)
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	err = <-shutdown
	if err != nil {
		return err
	}
	app.logger.Infow("server has stopped", "addr", app.config.addr, "env", app.config.env)
	return nil
}

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	
	"github.com/thiagomes07/organiQ/backend/config"
	"github.com/thiagomes07/organiQ/backend/internal/infra/database"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Setup logger
	setupLogger(cfg)
	
	log.Info().
		Str("environment", cfg.Environment).
		Str("version", "1.0.0").
		Msg("Starting organiQ Backend")

	// Initialize database connection
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer database.Close(db)
	
	log.Info().Msg("Database connection established")

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		log.Fatal().Err(err).Msg("Failed to run migrations")
	}
	
	log.Info().Msg("Database migrations completed")

	// TODO: Initialize storage (MinIO/S3)
	// TODO: Initialize queue (LocalStack/SQS)
	// TODO: Initialize repositories
	// TODO: Initialize use cases
	// TODO: Initialize handlers
	// TODO: Start workers

	// Setup HTTP router
	router := setupRouter(cfg)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in goroutine
	serverErrors := make(chan error, 1)
	go func() {
		log.Info().
			Str("address", server.Addr).
			Msg("Server listening")
		serverErrors <- server.ListenAndServe()
	}()

	// Wait for interrupt signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Fatal().Err(err).Msg("Server error")
	case sig := <-shutdown:
		log.Info().
			Str("signal", sig.String()).
			Msg("Shutdown signal received")

		// Graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Error().Err(err).Msg("Graceful shutdown failed")
			if err := server.Close(); err != nil {
				log.Fatal().Err(err).Msg("Could not stop server")
			}
		}
		
		log.Info().Msg("Server stopped gracefully")
	}
}

func setupLogger(cfg *config.Config) {
	// Set log level
	level, err := zerolog.ParseLevel(cfg.Logger.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Set output format
	if cfg.Logger.Format == "console" {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		})
	} else {
		// JSON format (default)
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	}

	// Add context fields
	log.Logger = log.With().
		Str("service", "organiq-api").
		Str("environment", cfg.Environment).
		Logger()
}

func setupRouter(cfg *config.Config) *chi.Mux {
	router := chi.NewRouter()

	// Middleware stack
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Compress(5))
	
	// Timeout middleware
	router.Use(middleware.Timeout(60 * time.Second))

	// CORS middleware
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: cfg.CORS.AllowCredentials,
		MaxAge:           cfg.CORS.MaxAge,
	}))

	// Request logger middleware
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			
			log.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("remote_addr", r.RemoteAddr).
				Int("status", ww.Status()).
				Int("bytes", ww.BytesWritten()).
				Dur("duration", time.Since(start)).
				Msg("HTTP Request")
		})
	})

	// Health check endpoint (no auth required)
	router.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"status": "healthy",
			"timestamp": "` + time.Now().Format(time.RFC3339) + `",
			"version": "1.0.0",
			"dependencies": {
				"database": "healthy",
				"storage": "pending",
				"queue": "pending",
				"ai": "pending"
			}
		}`))
	})

	// API routes (will be implemented later)
	router.Route("/api", func(r chi.Router) {
		// Auth routes
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", notImplementedHandler)
			r.Post("/login", notImplementedHandler)
			r.Post("/refresh", notImplementedHandler)
			r.Post("/logout", notImplementedHandler)
			r.Get("/me", notImplementedHandler)
		})

		// Plans routes
		r.Get("/plans", notImplementedHandler)

		// Protected routes (require authentication)
		r.Group(func(r chi.Router) {
			// TODO: Add auth middleware
			// r.Use(authMiddleware)

			// Payment routes
			r.Route("/payments", func(r chi.Router) {
				r.Post("/create-checkout", notImplementedHandler)
				r.Get("/status/{sessionId}", notImplementedHandler)
				r.Post("/create-portal-session", notImplementedHandler)
			})

			// Wizard routes
			r.Route("/wizard", func(r chi.Router) {
				r.Post("/business", notImplementedHandler)
				r.Post("/competitors", notImplementedHandler)
				r.Post("/integrations", notImplementedHandler)
				r.Post("/generate-ideas", notImplementedHandler)
				r.Get("/ideas-status/{jobId}", notImplementedHandler)
				r.Post("/publish", notImplementedHandler)
			})

			// Articles routes
			r.Route("/articles", func(r chi.Router) {
				r.Get("/", notImplementedHandler)
				r.Get("/{id}", notImplementedHandler)
				r.Post("/{id}/republish", notImplementedHandler)
			})

			// Account routes
			r.Route("/account", func(r chi.Router) {
				r.Get("/", notImplementedHandler)
				r.Patch("/profile", notImplementedHandler)
				r.Patch("/integrations", notImplementedHandler)
				r.Get("/plan", notImplementedHandler)
			})
		})
	})

	// Webhook routes (special auth)
	router.Post("/api/payments/webhook", notImplementedHandler)

	// 404 handler
	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"not_found","message":"Endpoint not found"}`))
	})

	return router
}

func notImplementedHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error":"not_implemented","message":"This endpoint is not yet implemented"}`))
}
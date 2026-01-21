// cmd/api/main.go - TRECHO COMPLETO COM WORKERS

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
	"github.com/rs/zerolog/log"

	"organiq/config"
	domainrepo "organiq/internal/domain/repository"
	"organiq/internal/handler"
	"organiq/internal/infra/ai"
	"organiq/internal/infra/database"
	"organiq/internal/infra/queue"
	"organiq/internal/infra/repository/postgres"
	"organiq/internal/infra/storage"
	authMiddleware "organiq/internal/middleware"
	accountUC "organiq/internal/usecase/account"
	"organiq/internal/usecase/article"
	"organiq/internal/usecase/auth"
	"organiq/internal/usecase/wizard"
	"organiq/internal/util"
	"organiq/internal/worker"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	loggerCleanup, err := util.InitLogger(cfg, util.LoggerOptions{
		Service:     "organiq-api",
		IncludeHook: true,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	if loggerCleanup != nil {
		defer loggerCleanup()
	}

	log.Info().
		Str("environment", cfg.Environment).
		Str("version", "1.0.0").
		Msg("Starting organiQ Backend")

	// Initialize database
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

	// ============================================
	// REPOSITORIES (criados antes do QueueService para suportar MockQueue)
	// ============================================

	repositories := &struct {
		User         domainrepo.UserRepository
		RefreshToken domainrepo.RefreshTokenRepository
		Plan         domainrepo.PlanRepository
		Business     domainrepo.BusinessRepository
		Integration  domainrepo.IntegrationRepository
		ArticleJob   domainrepo.ArticleJobRepository
		ArticleIdea  domainrepo.ArticleIdeaRepository
		Article      domainrepo.ArticleRepository
		Payment      domainrepo.PaymentRepository
	}{
		User:         postgres.NewUserRepository(db),
		RefreshToken: postgres.NewRefreshTokenRepository(db),
		Plan:         postgres.NewPlanRepository(db),
		Business:     postgres.NewBusinessRepository(db),
		Integration:  postgres.NewIntegrationRepository(db),
		ArticleJob:   postgres.NewArticleJobRepository(db),
		ArticleIdea:  postgres.NewArticleIdeaRepository(db),
		Article:      postgres.NewArticleRepository(db),
		Payment:      postgres.NewPaymentRepository(db),
	}

	log.Info().Msg("All repositories initialized")

	// ============================================
	// INICIALIZAR SERVIÇOS INFRAESTRUTURA
	// ============================================

	// Crypto service
	cryptoService := util.NewCryptoService(
		cfg.Auth.PasswordPepper,
		cfg.Auth.JWTSecret,
	)

	// Storage service
	storageService, err := storage.NewStorageService(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize storage service")
	}
	log.Info().Msg("Storage service initialized")

	// Queue service - agora com suporte a MockQueue
	// Se MOCK_AI_GENERATION=true, usa MockQueue que simula processamento com delay de 30s
	queueService, err := queue.NewQueueServiceWithMock(cfg, queue.MockQueueDependencies{
		ArticleJobRepo:  repositories.ArticleJob,
		ArticleIdeaRepo: repositories.ArticleIdea,
		ArticleRepo:     repositories.Article,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize queue service")
	}

	if queue.IsMockModeEnabled() {
		log.Warn().Msg("🚧 MODO MOCK ATIVO - Geração de ideias será simulada")
	}
	log.Info().Msg("Queue service initialized")

	// AI agent client
	agentClient := ai.NewAgentClient(cfg)
	log.Info().Msg("AI agent client initialized")

	// ============================================
	// USE CASES: AUTH
	// ============================================

	registerUC := auth.NewRegisterUserUseCase(
		repositories.User,
		repositories.Plan,
		repositories.RefreshToken,
		cryptoService,
	)

	loginUC := auth.NewLoginUserUseCase(
		repositories.User,
		repositories.RefreshToken,
		cryptoService,
	)

	refreshUC := auth.NewRefreshAccessTokenUseCase(
		repositories.User,
		repositories.RefreshToken,
		cryptoService,
	)

	logoutUC := auth.NewLogoutUserUseCase(
		repositories.RefreshToken,
		cryptoService,
	)

	getMeUC := auth.NewGetMeUseCase(
		repositories.User,
	)

	// ============================================
	// USE CASES: WIZARD
	// ============================================

	saveBusinessUC := wizard.NewSaveBusinessUseCase(
		repositories.Business,
		repositories.User,
		storageService,
	)

	saveCompetitorsUC := wizard.NewSaveCompetitorsUseCase(
		repositories.Business,
		repositories.User,
	)

	saveIntegrationsUC := wizard.NewSaveIntegrationsUseCase(
		repositories.Integration,
		repositories.User,
		cryptoService,
	)

	generateIdeasUC := wizard.NewGenerateIdeasUseCase(
		repositories.User,
		repositories.Plan,
		repositories.Business,
		repositories.ArticleJob,
		repositories.ArticleIdea,
		queueService,
	)

	getIdeasStatusUC := wizard.NewGetIdeasStatusUseCase(
		repositories.User,
		repositories.ArticleJob,
		repositories.ArticleIdea,
	)

	getWizardDataUC := wizard.NewGetWizardDataUseCase(
		repositories.User,
		repositories.Plan,
		repositories.Business,
		repositories.Integration,
		repositories.ArticleIdea,
	)

	// ============================================
	// USE CASES: ARTICLE
	// ============================================

	listArticlesUC := article.NewListArticlesUseCase(
		repositories.Article,
	)

	getArticleUC := article.NewGetArticleUseCase(
		repositories.Article,
	)

	republishArticleUC := article.NewRepublishArticleUseCase(
		repositories.Article,
		repositories.ArticleIdea,
		queueService,
	)

	publishArticlesUC := wizard.NewPublishArticlesUseCase(
		repositories.User,
		repositories.Plan,
		repositories.ArticleIdea,
		repositories.Article,
		repositories.ArticleJob,
		queueService,
	)

	getPublishStatusUC := wizard.NewGetPublishStatusUseCase(
		repositories.User,
		repositories.ArticleJob,
		repositories.Article,
	)

	// ============================================
	// USE CASES: ACCOUNT
	// ============================================

	getAccountUC := accountUC.NewGetAccountUseCase(
		repositories.User,
		repositories.Plan,
		repositories.Integration,
	)

	updateProfileUC := accountUC.NewUpdateProfileUseCase(
		repositories.User,
	)

	updateIntegrationsUC := accountUC.NewUpdateIntegrationsUseCase(
		repositories.Integration,
		cryptoService,
	)

	updatePasswordUC := accountUC.NewUpdatePasswordUseCase(
		repositories.User,
		cryptoService,
	)

	getPlanUC := accountUC.NewGetPlanUseCase(
		repositories.User,
		repositories.Plan,
	)

	log.Info().Msg("All use cases initialized")

	// ============================================
	// HANDLERS
	// ============================================

	authHandler := handler.NewAuthHandler(
		registerUC,
		loginUC,
		refreshUC,
		logoutUC,
		getMeUC,
		repositories.Plan,
	)

	planHandler := handler.NewPlanHandler(repositories.Plan)

	wizardHandler := handler.NewWizardHandler(
		saveBusinessUC,
		saveCompetitorsUC,
		saveIntegrationsUC,
		generateIdeasUC,
		getIdeasStatusUC,
		publishArticlesUC,
		getPublishStatusUC,
		getWizardDataUC,
	)

	// ============================================
	// PAYMENT HANDLER
	// ============================================

	paymentHandler := handler.NewPaymentHandler(
		repositories.User,
		repositories.Plan,
		repositories.Payment,
		cryptoService,
		cfg.Payment.StripeSecretKey,
		"", // stripePubKey (se necessário)
		cfg.Payment.StripeWebhookSecret,
		cfg.Payment.MercadoPagoAccessToken,
		cfg.Payment.MercadoPagoWebhookSecret,
	)

	log.Info().Msg("PaymentHandler inicializado")

	// ============================================
	// ARTICLE HANDLER
	// ============================================

	articleHandler := handler.NewArticleHandler(
		listArticlesUC,
		getArticleUC,
		republishArticleUC,
	)

	log.Info().Msg("ArticleHandler inicializado")

	// ============================================
	// ACCOUNT HANDLER
	// ============================================

	accountHandler := handler.NewAccountHandler(
		getAccountUC,
		updateProfileUC,
		updateIntegrationsUC,
		updatePasswordUC,
		getPlanUC,
	)

	log.Info().Msg("AccountHandler inicializado")

	healthHandler := handler.NewHealthHandler(
		db,
		storageService,
		queueService,
		[]string{
			cfg.Queue.ArticleGenerationQueue,
			cfg.Queue.ArticlePublishQueue,
		},
	)

	log.Info().Msg("All handlers initialized")

	// ============================================
	// INICIALIZAR WORKER POOL
	// ============================================

	workerPoolSize := cfg.Worker.PoolSize
	if workerPoolSize <= 0 {
		workerPoolSize = 5 // Default
	}

	workerPool := worker.NewWorkerPool(
		workerPoolSize,
		queueService,
		repositories.ArticleJob,
		repositories.ArticleIdea,
		repositories.Article,
		repositories.Business,
		repositories.Integration,
		agentClient,
		cryptoService,
		cfg.Worker.PollInterval,
		cfg.Worker.MaxRetries,
	)

	log.Info().Int("pool_size", workerPoolSize).Msg("Worker pool created")

	// ============================================
	// SETUP HTTP ROUTER
	// ============================================

	// Rate limiters por endpoint conforme spec 3.5
	rateLimiters := &RateLimiters{}
	if cfg.RateLimit.Enabled {
		// Global: 1000 req/1min por IP
		rateLimiters.Global = authMiddleware.NewRateLimiter(1000, time.Minute)
		// Auth: 30 req/15min por IP (Ajustado para permitir mais tentativas legítimas)
		rateLimiters.Auth = authMiddleware.NewRateLimiter(30, 15*time.Minute)
		// Wizard generate/publish: 30 req/1h por User (Ajustado para permitir iteração na criação)
		rateLimiters.Wizard = authMiddleware.NewRateLimiter(30, time.Hour)
		// Articles: 100 req/1min por User
		rateLimiters.Articles = authMiddleware.NewRateLimiter(100, time.Minute)
	}

	router := setupRouter(cfg, authHandler, planHandler, wizardHandler, paymentHandler, articleHandler, accountHandler, healthHandler, cryptoService, rateLimiters)

	// ============================================
	// CREATE HTTP SERVER
	// ============================================

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	log.Info().Msg("HTTP server created")

	// ============================================
	// START WORKER POOL
	// ============================================

	workerCtx, workerCancel := context.WithCancel(context.Background())
	workerShutdownDone := workerPool.Start(workerCtx)

	log.Info().Msg("Worker pool started")

	// ============================================
	// START HTTP SERVER
	// ============================================

	serverErrors := make(chan error, 1)
	go func() {
		log.Info().
			Str("address", server.Addr).
			Msg("HTTP Server listening")
		serverErrors <- server.ListenAndServe()
	}()

	// ============================================
	// GRACEFUL SHUTDOWN LOGIC
	// ============================================

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server error")
		}

	case sig := <-shutdown:
		log.Info().
			Str("signal", sig.String()).
			Msg("Shutdown signal received - starting graceful shutdown")

		// ============================================
		// FASE 1: Stop HTTP server (sem aceitar novos requests)
		// ============================================

		log.Info().Msg("Stopping HTTP server...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Error().Err(err).Msg("HTTP server shutdown error")
			if err := server.Close(); err != nil {
				log.Fatal().Err(err).Msg("Could not close server")
			}
		}

		log.Info().Msg("HTTP server stopped")

		// ============================================
		// FASE 2: Stop worker pool (graceful shutdown com timeout)
		// ============================================

		log.Info().Msg("Stopping worker pool...")

		// Cancelar contexto dos workers
		workerCancel()

		// Aguardar shutdown completo com timeout
		workerShutdownCtx, workerShutdownCancel := context.WithTimeout(
			context.Background(),
			cfg.Server.ShutdownTimeout,
		)
		defer workerShutdownCancel()

		select {
		case <-workerShutdownDone:
			log.Info().Msg("Worker pool stopped gracefully")

		case <-workerShutdownCtx.Done():
			log.Warn().
				Dur("timeout", cfg.Server.ShutdownTimeout).
				Msg("Worker pool shutdown timeout - forcing exit")
		}

		log.Info().Msg("Server and workers stopped successfully")
	}
}

// ============================================
// HELPER FUNCTIONS
// ============================================

// RateLimiters agrupa os rate limiters por endpoint conforme spec 3.5
type RateLimiters struct {
	Global   *authMiddleware.RateLimiter // 1000 req/1min por IP
	Auth     *authMiddleware.RateLimiter // 5 req/15min por IP
	Wizard   *authMiddleware.RateLimiter // 10 req/1h por User
	Articles *authMiddleware.RateLimiter // 100 req/1min por User
}

func setupRouter(
	cfg *config.Config,
	authHandler *handler.AuthHandler,
	planHandler *handler.PlanHandler,
	wizardHandler *handler.WizardHandler,
	paymentHandler *handler.PaymentHandler,
	articleHandler *handler.ArticleHandler,
	accountHandler *handler.AccountHandler,
	healthHandler *handler.HealthHandler,
	cryptoService *util.CryptoService,
	rateLimiters *RateLimiters,
) *chi.Mux {
	router := chi.NewRouter()

	// Middlewares globais
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(authMiddleware.RecoveryMiddleware())
	router.Use(authMiddleware.LoggerMiddleware())
	router.Use(middleware.Compress(5))
	router.Use(middleware.Timeout(60 * time.Second))

	// Rate limit global: 1000 req/1min por IP
	if rateLimiters != nil && rateLimiters.Global != nil {
		router.Use(authMiddleware.RateLimitMiddleware(rateLimiters.Global, authMiddleware.IPIdentifier))
	}

	// Security Headers Middleware (spec 3.6)
	router.Use(securityHeadersMiddleware(cfg.Environment == "production"))

	// CORS
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link", "X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset"},
		AllowCredentials: cfg.CORS.AllowCredentials,
		MaxAge:           cfg.CORS.MaxAge,
	}))

	// Health check
	router.Get("/api/health", healthHandler.Check)

	// API routes
	router.Route("/api", func(r chi.Router) {
		// Public plans endpoint
		r.Get("/plans", planHandler.ListPlans)

		// Auth (rotas públicas com rate limit específico: 5 req/15min por IP)
		r.Route("/auth", func(r chi.Router) {
			if rateLimiters != nil && rateLimiters.Auth != nil {
				r.Use(authMiddleware.RateLimitMiddleware(rateLimiters.Auth, authMiddleware.IPIdentifier))
			}
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.Refresh)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.AuthMiddleware(cryptoService))

			// Auth routes que requerem autenticação (spec 4.2)
			r.Get("/auth/me", authHandler.GetMe)
			r.Post("/auth/logout", authHandler.Logout)

			// Wizard routes com rate limit específico: 10 req/1h por User
			r.Route("/wizard", func(r chi.Router) {
				r.Get("/data", wizardHandler.GetWizardData)
				r.Post("/business", wizardHandler.SaveBusiness)
				r.Post("/competitors", wizardHandler.SaveCompetitors)
				r.Post("/integrations", wizardHandler.SaveIntegrations)

				// Rate limit específico para generate-ideas e publish
				r.Group(func(r chi.Router) {
					if rateLimiters != nil && rateLimiters.Wizard != nil {
						r.Use(authMiddleware.RateLimitMiddleware(rateLimiters.Wizard, authMiddleware.UserIdentifier))
					}
					r.Post("/generate-ideas", wizardHandler.GenerateIdeas)
					r.Post("/publish", wizardHandler.PublishArticles)
				})

				r.Get("/ideas-status/{jobId}", wizardHandler.GetIdeasStatus)
				r.Get("/publish-status/{jobId}", wizardHandler.GetPublishStatus)
			})

			// Payment routes (protegidas) - spec 4.4
			r.Route("/payments", func(r chi.Router) {
				r.Post("/create-checkout", paymentHandler.CreateCheckout)
				r.Get("/status/{sessionId}", paymentHandler.GetStatus)
				r.Post("/create-portal-session", paymentHandler.CreatePortalSession)
				r.Post("/confirm-free-plan", paymentHandler.ConfirmFreePlan)
			})

			// Article routes (protegidas) com rate limit: 100 req/1min por User - spec 4.6
			r.Route("/articles", func(r chi.Router) {
				if rateLimiters != nil && rateLimiters.Articles != nil {
					r.Use(authMiddleware.RateLimitMiddleware(rateLimiters.Articles, authMiddleware.UserIdentifier))
				}
				r.Get("/", articleHandler.ListArticles)
				r.Get("/{id}", articleHandler.GetArticle)
				r.Post("/{id}/republish", articleHandler.RepublishArticle)
			})

			// Account routes (protegidas) - spec 4.7
			r.Route("/account", func(r chi.Router) {
				r.Get("/", accountHandler.GetAccount)
				r.Patch("/profile", accountHandler.UpdateProfile)
				r.Patch("/integrations", accountHandler.UpdateIntegrations)
				r.Patch("/password", accountHandler.UpdatePassword)
				r.Get("/plan", accountHandler.GetPlan)
			})
		})
	})

	// Public webhook routes (SEM auth middleware)
	router.Post("/webhooks/payments", paymentHandler.Webhook)

	return router
}

// securityHeadersMiddleware adiciona headers de segurança conforme spec 3.6
func securityHeadersMiddleware(isProduction bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Headers de segurança obrigatórios (spec 3.6)
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
			w.Header().Set("Content-Security-Policy", "default-src 'self'")

			// HSTS apenas em produção (requer HTTPS)
			if isProduction {
				w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
			}

			next.ServeHTTP(w, r)
		})
	}
}

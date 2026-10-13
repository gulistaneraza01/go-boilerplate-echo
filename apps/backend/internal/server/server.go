package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gulistaneraza01/go-boilerplate-echo/internal/config"
	"github.com/gulistaneraza01/go-boilerplate-echo/internal/database"
	"github.com/gulistaneraza01/go-boilerplate-echo/internal/lib/job"
	loggerPkg "github.com/gulistaneraza01/go-boilerplate-echo/internal/logger"
	"github.com/newrelic/go-agent/v3/integrations/nrredis-v9"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type Server struct {
	Config        *config.Config
	Logger        *zerolog.Logger
	LoggerService *loggerPkg.LoggerService
	DB            *database.Database
	RedisJob      *redis.Client
	RedisCache    *redis.Client
	httpServer    *http.Server
	Job           *job.JobService
}

func New(cfg *config.Config, logger *zerolog.Logger, loggerService *loggerPkg.LoggerService) (*Server, error) {
	db, err := database.New(cfg, logger, loggerService)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	redisJob, err := newRedisClient("redis job", cfg.RedisJob.Address, logger, loggerService)
	if err != nil {
		return nil, err
	}

	redisCache, err := newRedisClient("redis cache", cfg.RedisCache.Address, logger, loggerService)
	if err != nil {
		return nil, err
	}

	// job service
	jobService := job.NewJobService(logger, redisJob.Options())
	jobService.InitHandlers(cfg, logger)

	// Start job server
	if err := jobService.Start(); err != nil {
		return nil, err
	}

	server := &Server{
		Config:        cfg,
		Logger:        logger,
		LoggerService: loggerService,
		DB:            db,
		RedisJob:      redisJob,
		RedisCache:    redisCache,
		Job:           jobService,
	}

	// Start metrics collection
	// Runtime metrics are automatically collected by New Relic Go agent

	return server, nil
}

func (s *Server) SetupHTTPServer(handler http.Handler) {
	s.httpServer = &http.Server{
		Addr:         ":" + s.Config.Server.Port,
		Handler:      handler,
		ReadTimeout:  time.Duration(s.Config.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(s.Config.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(s.Config.Server.IdleTimeout) * time.Second,
	}
}

func (s *Server) Start() error {
	if s.httpServer == nil {
		return errors.New("HTTP server not initialized")
	}

	s.Logger.Info().
		Str("port", s.Config.Server.Port).
		Str("env", s.Config.Primary.Env).
		Msg("starting server")

	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	// Stop workers before closing the connections they depend on
	if s.Job != nil {
		s.Job.Stop()
	}

	if err := s.DB.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	if err := s.RedisJob.Close(); err != nil {
		return fmt.Errorf("failed to close redis job connection: %w", err)
	}

	if err := s.RedisCache.Close(); err != nil {
		return fmt.Errorf("failed to close redis cache connection: %w", err)
	}

	return nil
}

func newRedisClient(name, address string, logger *zerolog.Logger, loggerService *loggerPkg.LoggerService) (*redis.Client, error) {
	opts, err := redis.ParseURL(address)
	if err != nil {
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			err = urlErr.Err
		}
		return nil, fmt.Errorf("invalid %s address: %w", name, err)
	}

	client := redis.NewClient(opts)

	if loggerService != nil && loggerService.GetApplication() != nil {
		client.AddHook(nrredis.NewHook(client.Options()))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		logger.Error().Err(err).Str("client", name).Msg("Failed to connect to Redis, continuing without it")
	}

	return client, nil
}

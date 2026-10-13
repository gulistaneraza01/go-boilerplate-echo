package job

import (
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type JobService struct {
	Client *asynq.Client
	server *asynq.Server
	logger *zerolog.Logger
}

// redisConnOpt lets asynq build its client from the full parsed Redis URL options.
type redisConnOpt struct{ opts *redis.Options }

func (r redisConnOpt) MakeRedisClient() interface{} {
	opts := *r.opts
	return redis.NewClient(&opts)
}

func NewJobService(logger *zerolog.Logger, redisOpts *redis.Options) *JobService {
	clientOpt := redisConnOpt{opts: redisOpts}

	client := asynq.NewClient(clientOpt)

	server := asynq.NewServer(
		clientOpt,
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6, // Higher priority queue for important emails
				"default":  3, // Default priority for most emails
				"low":      1, // Lower priority for non-urgent emails
			},
		},
	)

	return &JobService{
		Client: client,
		server: server,
		logger: logger,
	}
}

func (j *JobService) Start() error {
	// Register task handlers
	mux := asynq.NewServeMux()
	mux.HandleFunc(TaskWelcome, j.handleWelcomeEmailTask)

	j.logger.Info().Msg("Starting background job server")
	if err := j.server.Start(mux); err != nil {
		return err
	}

	return nil
}

func (j *JobService) Stop() {
	j.logger.Info().Msg("Stopping background job server")
	j.server.Shutdown()
	j.Client.Close()
}

package service

import (
	"context"

	"github.com/gulistaneraza01/go-boilerplate-echo/internal/lib/aws"
	"github.com/gulistaneraza01/go-boilerplate-echo/internal/lib/job"
	"github.com/gulistaneraza01/go-boilerplate-echo/internal/repository"
	"github.com/gulistaneraza01/go-boilerplate-echo/internal/server"
)

type Services struct {
	Auth *AuthService
	Job  *job.JobService
	AWS  *aws.AWS
}

func NewServices(s *server.Server, repos *repository.Repositories) (*Services, error) {
	authService := NewAuthService(s)

	// Pass awsClient into services that need it, e.g. NewTodoService(s, repos.Todo, awsClient)
	awsClient, err := aws.NewAWS(context.Background(), s.Config.AWS)
	if err != nil {
		return nil, err
	}

	return &Services{
		Job:  s.Job,
		Auth: authService,
		AWS:  awsClient,
	}, nil
}

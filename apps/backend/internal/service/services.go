package service

import (
	"github.com/gulistaneraza01/go-boilerplate-echo/internal/lib/job"
	"github.com/gulistaneraza01/go-boilerplate-echo/internal/repository"
	"github.com/gulistaneraza01/go-boilerplate-echo/internal/server"
)

type Services struct {
	Auth *AuthService
	Job  *job.JobService
}

func NewServices(s *server.Server, repos *repository.Repositories) (*Services, error) {
	authService := NewAuthService(s)

	return &Services{
		Job:  s.Job,
		Auth: authService,
	}, nil
}

package repository

import "github.com/gulistaneraza01/go-boilerplate-echo/internal/server"

type Repositories struct{}

func NewRepositories(s *server.Server) *Repositories {
	return &Repositories{}
}

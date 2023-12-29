package api

import (
	"testing"

	"github.com/toduluz/savingsquadsbackend/internal/data"
)

func newTestApplication(t *testing.T) *Application {
	t.Helper()

	var cfg Config
	cfg.Jwt.Secret = "secret"

	return &Application{
		Config: cfg,
		Models: data.NewMockModels(),
	}
}

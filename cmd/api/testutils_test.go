package main

import (
	"io"
	"testing"

	"github.com/toduluz/savingsquadsbackend/internal/data"
	"github.com/toduluz/savingsquadsbackend/internal/jsonlog"
)

func newTestApplication(t *testing.T) *application {

	jwtSecret := "secert"

	var cfg config
	cfg.jwt.secret = jwtSecret

	return &application{
		config: cfg,
		logger: jsonlog.NewLogger(io.Discard, jsonlog.LevelOff),
		models: data.NewMockModels(),
	}
}

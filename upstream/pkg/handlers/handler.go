package handlers

import "log/slog"

// Handler holds fields shared by all HTTP handlers (e.g. logger).
type Handler struct {
	logger *slog.Logger
}

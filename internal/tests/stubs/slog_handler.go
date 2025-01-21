package stubs

import (
	"context"
	"log/slog"
)

type Handler struct{ slog.Handler }

func NewHandler() slog.Handler { return &Handler{} }

func (h *Handler) Enabled(_ context.Context, _ slog.Level) bool  { return false }
func (h *Handler) Handle(_ context.Context, _ slog.Record) error { return nil }
func (h *Handler) WithAttrs(_ []slog.Attr) slog.Handler          { return h }
func (h *Handler) WithGroup(_ string) slog.Handler               { return h }

package logs

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
)

type prefixHandler struct {
	handler slog.Handler // Используем полноценный обработчик
	prefix  string       // Префикс, добавляемый ко всем записям
}

func (ph *prefixHandler) Enabled(ctx context.Context, level slog.Level) bool {
	//TODO implement me
	return ph.handler.Enabled(ctx, level)

}

func (ph *prefixHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	//TODO implement me

	return ph.handler.WithAttrs(attrs)
}

func (ph *prefixHandler) WithGroup(name string) slog.Handler {
	//TODO implement me
	return ph.handler.WithGroup(name)

}

func (ph *prefixHandler) Handle(conn context.Context, r slog.Record) error {
	r.Message = ph.prefix + r.Message
	return ph.handler.Handle(conn, r)
}

func NewPrefixHandler(w io.Writer, handlerOpts *slog.HandlerOptions, prefix string) slog.Handler {
	baseHandler := slog.NewTextHandler(w, handlerOpts) // Или другой подходящий обработчик
	hand := &prefixHandler{
		handler: baseHandler,
		prefix:  fmt.Sprintf("[%s]: ", prefix),
	}
	return hand
}

func LogInit() *slog.Logger {
	w := os.Stdout
	opts := &slog.HandlerOptions{
		AddSource: false, // Не показывать источник (файл и строку)
	}

	hand := NewPrefixHandler(w, opts, "ObsSystem")
	logger := slog.New(hand)

	return logger
}

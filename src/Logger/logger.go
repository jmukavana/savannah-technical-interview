package Logger

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New() *zap.Logger {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, _ := cfg.Build()
	return logger
}

func ChiMiddleware(log *zap.Logger) func(next http.Handler) http.Handler {
	sugar := log.Sugar()
	return middleware.RequestLogger(&chiZapLogger{sugar})
}

type chiZapLogger struct{ sugar *zap.SugaredLogger }

func (c *chiZapLogger) NewLogEntry(r *http.Request) middleware.LogEntry {
	c.sugar.Infow("request", "method", r.Method, "path", r.URL.Path)
	return &chiZapLogEntry{c.sugar}
}

type chiZapLogEntry struct{ sugar *zap.SugaredLogger }

func (l *chiZapLogEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	l.sugar.Infow("response", "status", status, "bytes", bytes, "elapsed", elapsed)
}
func (l *chiZapLogEntry) Panic(v interface{}, stack []byte) {
	l.sugar.Errorw("panic", "value", v)
}

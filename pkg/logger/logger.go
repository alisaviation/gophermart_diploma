package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Log        *zap.Logger        = zap.NewNop()
	SugaredLog *zap.SugaredLogger = Log.Sugar()
)

func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	zl, err := cfg.Build()
	if err != nil {
		return err
	}

	Log = zl
	SugaredLog = Log.Sugar()
	return nil
}

func Sync() {
	_ = Log.Sync()
}

func RequestResponseLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(ww, r)

		duration := time.Since(start)

		Log.Info("HTTP request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("query", r.URL.RawQuery),
			zap.Int("status", ww.statusCode),
			zap.Int("size", ww.size),
			zap.Duration("duration", duration),
			zap.String("request_id", ww.Header().Get("X-Request-ID")),
		)

		if Log.Core().Enabled(zapcore.DebugLevel) {
			Log.Debug("HTTP request details",
				zap.Any("headers", sanitizeHeaders(r.Header)),
				zap.Any("response_headers", sanitizeHeaders(ww.Header())),
			)
		}
	})
}

func sanitizeHeaders(headers http.Header) map[string]string {
	safeHeaders := make(map[string]string)
	for k, v := range headers {
		if len(v) == 0 {
			continue
		}

		switch k {
		case "Authorization", "Cookie", "Set-Cookie":
			safeHeaders[k] = "***REDACTED***"
		default:
			safeHeaders[k] = v[0]
		}
	}
	return safeHeaders
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
	headers    http.Header
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}
func (rw *responseWriter) Header() http.Header {
	if rw.headers == nil {
		rw.headers = make(http.Header)
	}
	return rw.ResponseWriter.Header()
}

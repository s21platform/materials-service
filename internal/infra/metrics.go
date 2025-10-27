package infra

import (
	"context"
	"net/http"
	"strings"
	"time"

	"google.golang.org/grpc"

	"github.com/s21platform/metrics-lib/pkg"

	"github.com/s21platform/materials-service/internal/config"
)

func MetricsInterceptorGRPC(metrics *pkg.Metrics) func(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		startTime := time.Now()
		method := strings.Trim(strings.ReplaceAll(info.FullMethod, "/", "_"), "_")
		metrics.Increment(method)

		ctx = context.WithValue(ctx, config.KeyMetrics, metrics)
		resp, err := handler(ctx, req)

		if err != nil {
			metrics.Increment(method + "_error")
		}

		metrics.Duration(time.Since(startTime).Milliseconds(), method)

		return resp, err
	}
}

func MetricsInterceptorHTTP(next http.Handler, metrics *pkg.Metrics) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		metricName := strings.Trim(strings.ReplaceAll(r.Method+"_"+r.URL.Path, "/", "_"), "_")
		metricName = strings.ReplaceAll(metricName, "-", "_")

		metrics.Increment(metricName)

		ctx := context.WithValue(r.Context(), config.KeyMetrics, metrics)

		statusCodeRecorder := &statusCodeRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(statusCodeRecorder, r.WithContext(ctx))

		if statusCodeRecorder.statusCode >= http.StatusBadRequest {
			metrics.Increment(metricName + "_error")
		}

		metrics.Duration(time.Since(startTime).Milliseconds(), metricName)
	})
}

type statusCodeRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rec *statusCodeRecorder) WriteHeader(code int) {
	rec.statusCode = code
	rec.ResponseWriter.WriteHeader(code)
}

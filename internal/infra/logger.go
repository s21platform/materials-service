package infra

import (
	"context"
	"net/http"

	"google.golang.org/grpc"

	logger_lib "github.com/s21platform/logger-lib"
)

func LoggerGRPC(logger *logger_lib.Logger) func(context.Context, interface{}, *grpc.UnaryServerInfo, grpc.UnaryHandler) (interface{}, error) {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx = logger_lib.NewContext(ctx, logger)
		ctx = logger_lib.WithField(ctx, "method", info.FullMethod)
		return handler(ctx, req)
	}
}

func LoggerHTTP(next http.Handler, logger *logger_lib.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := logger_lib.NewContext(r.Context(), logger)
		ctx = logger_lib.WithField(ctx, "path", r.URL.Path)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
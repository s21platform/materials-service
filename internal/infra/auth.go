package infra

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/s21platform/materials-service/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func AuthInterceptorGRPC(
	ctx context.Context,
	req interface{},
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no info in metadata")
	}

	userIDs, ok := md["uuid"]
	if !ok || len(userIDs) != 1 {
		return nil, status.Errorf(codes.Unauthenticated, "no uuid or more than one in metadata")
	}

	ctx = context.WithValue(ctx, config.KeyUUID, userIDs[0])

	return handler(ctx, req)
}

func AuthInterceptorHTTP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		userID = strings.TrimSpace(userID)

		if userID == "" {
			writeErrorResponse(w, "missing or empty X-User-ID header", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), config.KeyUUID, userID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := map[string]string{
		"error": message,
	}

	_ = json.NewEncoder(w).Encode(errorResponse)
}

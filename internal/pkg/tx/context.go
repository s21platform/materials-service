package tx

import (
	"context"

	"google.golang.org/grpc"
)

type key string

const KeyTx = key("tx")

func WithDBRepoContext(ctx context.Context, repo DBRepo) context.Context {
	return context.WithValue(ctx, KeyTx, Tx{DbRepo: repo})
}

func TxMiddleWire(db DBRepo) func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (
		interface{}, error) {
		return handler(WithDBRepoContext(ctx, db), req)
	}
}

func fromContext(ctx context.Context) Tx {
	v, ok := ctx.Value(KeyTx).(Tx)
	if !ok {
		panic("no Tx found in context")
	}
	return v
}

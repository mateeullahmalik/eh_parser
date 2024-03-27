package ethereum

import "context"

type Client interface {
	GetBlockCount(ctx context.Context) (int32, error)
	GetTransaction(ctx context.Context, txID string) (GetTransactionResult, error)
	GetTransactionsByAddress(ctx context.Context, address string, block int32) (GetTransactionsResult, error)
}

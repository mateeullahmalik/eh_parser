package ethereum

import "context"

type Client interface {
	GetLatestBlockNumber(ctx context.Context) (int32, error)
	GetBlockTransactions(ctx context.Context, block int32) (TransactionResults, error)
}

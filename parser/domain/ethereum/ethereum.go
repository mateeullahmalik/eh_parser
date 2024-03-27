package ethereum

import (
	"context"

	"github.com/mateeullahmalik/eh_parser/parser/domain"
)

type EthClient interface {
	GetBlockCount(ctx context.Context) (int32, error)
	GetTransactionsByAddress(ctx context.Context, address string, block int32) (domain.Transactions, error)
}

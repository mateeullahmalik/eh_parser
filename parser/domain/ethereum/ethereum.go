package ethereum

import (
	"context"

	"github.com/mateeullahmalik/eh_parser/parser/domain"
)

type EthClient interface {
	GetBlockCount(ctx context.Context) (int32, error)
	GetTransactionsWithAddressesFilter(ctx context.Context, block int32, addresses ...string) (domain.Transactions, error)
}

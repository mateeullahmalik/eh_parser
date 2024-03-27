package ethereum

import (
	"context"

	"github.com/mateeullahmalik/eh_parser/ethereum"
	"github.com/mateeullahmalik/eh_parser/parser/domain"
)

type EthereumBlockchain struct {
	client ethereum.Client
}

func NewEthereumBlockchain(client ethereum.Client) *EthereumBlockchain {
	return &EthereumBlockchain{
		client: client,
	}
}

func (e *EthereumBlockchain) GetBlockCount(ctx context.Context) (int32, error) {
	return e.client.GetBlockCount(ctx)
}

func (e *EthereumBlockchain) GetTransactionsByAddress(ctx context.Context, address string, block int32) (txns domain.Transactions, err error) {
	transactions, err := e.client.GetTransactionsByAddress(ctx, address, block)
	if err != nil {
		return txns, err
	}

	txns = make(domain.Transactions, 0, len(transactions))
	for _, tx := range transactions {
		txns = append(txns, domain.Transaction{
			TxID:   tx.TxID,
			Amount: tx.Amount,
			From:   tx.From,
			To:     tx.To,
			Fee:    tx.Fee,
			Block:  tx.Block,
		})
	}

	return txns, nil
}
